package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

func (a *API) registerAuthRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	// Mêmes limites que login/register : google (idToken) et 2FA (code à 6 chiffres brute-forçable).
	r.Group(func(ar chi.Router) {
		ar.Use(rateLimit)
		ar.Post("/auth/google", a.googleLogin)
		ar.Post("/auth/2fa/verify", a.verify2FA)
	})

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		// Pas de gate CGU ici : se déconnecter doit rester possible.
		pr.Post("/auth/logout", a.logout)
		pr.Get("/auth/2fa/status", a.twoFactorStatus)
		pr.Post("/auth/2fa/setup", a.twoFactorSetup)
		pr.Post("/auth/2fa/confirm", a.twoFactorConfirm)
		pr.Post("/auth/2fa/disable", a.twoFactorDisable)
	})
}

// logout revokes the caller's issued tokens by bumping token_version.
//
// JWTs are stateless, so this is what makes a stolen refresh token useless
// before its 30-day expiry. It applies to every session of the account: a
// logout on the web also ends the mobile session.
func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if err := a.store.BumpTokenVersion(r.Context(), id.UserID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"loggedOut": true})
}

type googleLoginReq struct {
	IDToken    string `json:"idToken"`
	Audience   string `json:"audience,omitempty"` // "pro" (default, Nuxt) | "client" (Flutter pets)
	InviteCode string `json:"inviteCode,omitempty"`
	// CommercialUserID — optional nearby commercial pick when no invite code (client audience).
	CommercialUserID string `json:"commercialUserId,omitempty"`
	Consent          bool   `json:"consent,omitempty"` // requis pour create-if-absent audience=client (RGPD)
}

func (a *API) googleLogin(w http.ResponseWriter, r *http.Request) {
	if a.cfg.GoogleOAuthClientID == "" {
		writeErr(w, r, http.StatusNotImplemented, "not_configured", "not_configured")
		return
	}
	var req googleLoginReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.IDToken == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "id_token_required")
		return
	}

	payload, err := validateGoogleIDToken(r.Context(), req.IDToken, a.cfg.GoogleOAuthClientID)
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_google_token")
		return
	}
	var gClaims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := payload.Claims(&gClaims); err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_google_token")
		return
	}
	email := strings.TrimSpace(strings.ToLower(gClaims.Email))
	name := gClaims.Name
	if email == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "google_email_missing")
		return
	}
	if !gClaims.EmailVerified {
		writeErr(w, r, http.StatusForbidden, "email_not_verified", "google_email_not_verified")
		return
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	u, err := a.resolveGoogleUser(r, email, name, payload.Subject, req.Audience, req.Consent)
	if err != nil {
		a.writeGoogleAuthError(w, r, err)
		return
	}
	if u.IsWalkinPlaceholder || store.IsWalkinPlaceholderEmail(u.Email) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if normalizeGoogleAudience(req.Audience) == "client" {
		inviteStatus := a.tryClaimInvite(r, u.ID, req.InviteCode)
		if store.NormalizeInviteCode(req.InviteCode) == "" {
			a.tryLinkCommercialReferral(r, u.ID, req.CommercialUserID)
		}
		a.issueLoginResponseWithExtra(w, r, u, map[string]any{"inviteStatus": inviteStatus})
		return
	}
	a.issueLoginResponse(w, r, u)
}

func normalizeGoogleAudience(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "client":
		return "client"
	default:
		return "pro"
	}
}

func roleMatchesGoogleAudience(role kernel.Role, audience string) bool {
	switch audience {
	case "client":
		return role == kernel.RoleClient
	default:
		return role == kernel.RoleVet || role == kernel.RoleAdmin || kernel.IsSalesForce(role)
	}
}

func (a *API) linkOrMatchGoogle(ctx context.Context, u store.User, googleSub string) (store.User, error) {
	if u.GoogleSub == "" {
		if err := a.store.LinkGoogleAccount(ctx, u.ID, googleSub); err != nil {
			return store.User{}, err
		}
		return a.store.GetUserByID(ctx, u.ID)
	}
	if u.GoogleSub != googleSub {
		return store.User{}, errGoogleAccountMismatch
	}
	return u, nil
}

func (a *API) resolveGoogleUser(r *http.Request, email, fullName, googleSub, audienceRaw string, consent bool) (store.User, error) {
	ctx := r.Context()
	audience := normalizeGoogleAudience(audienceRaw)

	if u, err := a.store.GetUserByGoogleSub(ctx, googleSub); err == nil {
		if !roleMatchesGoogleAudience(u.Role, audience) {
			if audience == "client" {
				return store.User{}, errGoogleClientOnly
			}
			return store.User{}, errGoogleProOnly
		}
		return u, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	u, err := a.store.GetUserByEmail(ctx, email)
	if err == nil {
		if !roleMatchesGoogleAudience(u.Role, audience) {
			if audience == "client" {
				return store.User{}, errGoogleClientOnly
			}
			return store.User{}, errGoogleProOnly
		}
		return a.linkOrMatchGoogle(ctx, u, googleSub)
	}
	if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	// Unknown email: Pro can auto-register a vet; clients can create-if-absent via Google.
	if audience == "client" {
		if !consent {
			return store.User{}, errGoogleConsentRequired
		}
		locale := localeOf(r)
		return a.store.RegisterGoogleClient(ctx, store.RegisterGoogleClientInput{
			Email: email, FullName: fullName, GoogleSub: googleSub, PreferredLocale: locale,
			TermsAccepted: true,
		})
	}

	practiceName := fmt.Sprintf("Cabinet %s", strings.Split(fullName, " ")[0])
	locale := localeOf(r)
	return a.store.RegisterGoogleVet(ctx, store.RegisterGoogleVetInput{
		Email: email, FullName: fullName, GoogleSub: googleSub, PracticeName: practiceName,
		PreferredLocale: locale, AutoReplyDefault: t(r, "defaults.auto_reply_unavailable", nil),
		TermsAccepted: consent,
	})
}

var (
	errGoogleProOnly         = errors.New("google pro only")
	errGoogleClientOnly      = errors.New("google client only")
	errGoogleAccountMismatch = errors.New("google account mismatch")
	errGoogleConsentRequired = errors.New("google consent required")
)

func (a *API) writeGoogleAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errGoogleProOnly):
		writeErr(w, r, http.StatusForbidden, "google_pro_only", "google_pro_only")
	case errors.Is(err, errGoogleClientOnly):
		writeErr(w, r, http.StatusForbidden, "google_client_only", "google_client_only")
	case errors.Is(err, errGoogleAccountMismatch):
		writeErr(w, r, http.StatusConflict, "google_account_mismatch", "google_account_mismatch")
	case errors.Is(err, errGoogleConsentRequired):
		writeErr(w, r, http.StatusBadRequest, "consent_required", "consent_required")
	case errors.Is(err, store.ErrConflict):
		writeErr(w, r, http.StatusConflict, "conflict", "google_account_mismatch")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}

func (a *API) issueLoginResponse(w http.ResponseWriter, r *http.Request, u store.User) {
	a.issueLoginResponseWithExtra(w, r, u, nil)
}

func (a *API) issueLoginResponseWithExtra(w http.ResponseWriter, r *http.Request, u store.User, extra map[string]any) {
	if (u.Role == kernel.RoleVet || u.Role == kernel.RoleClient || u.Role == kernel.RoleCarePro ||
		u.Role == kernel.RoleVetAssistant || u.Role == kernel.RoleSecretary) && u.EmailVerifiedAt == nil {
		writeErr(w, r, http.StatusForbidden, "email_not_verified", "email_not_verified")
		return
	}
	_ = a.store.EnsureUserProfiles(r.Context(), u.ID)
	active, _ := a.store.GetActiveProfile(r.Context(), u.ID)
	// Reload user in case switch sync needed — use active profile role for tokens.
	if active.ID != "" {
		u2, err := a.store.GetUserByID(r.Context(), u.ID)
		if err == nil {
			u = u2
		}
	}
	profileID := active.ID
	if u.TOTPEnabled {
		mfa, err := a.tokens.IssueMFA(u.ID, u.Email, u.Role, u.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if len(extra) == 0 {
			httpx.WriteData(w, http.StatusOK, mfa)
			return
		}
		out := map[string]any{
			"requires2FA": mfa.Requires2FA,
			"mfaToken":    mfa.MFAToken,
			"expiresIn":   mfa.ExpiresIn,
		}
		maps.Copy(out, extra)
		httpx.WriteData(w, http.StatusOK, out)
		return
	}
	pair, err := a.tokens.IssueProfile(u.ID, u.Email, u.Role, u.PracticeID, profileID, u.TokenVersion)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.store.TouchLastLogin(r.Context(), u.ID)
	if len(extra) == 0 {
		httpx.WriteData(w, http.StatusOK, pair)
		return
	}
	out := map[string]any{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
	}
	maps.Copy(out, extra)
	httpx.WriteData(w, http.StatusOK, out)
}

// totpReplayGuard — anti-replay : un code TOTP accepté ne peut pas être rejoué
// dans sa fenêtre de validité (~90 s avec skew ±1). En mémoire, par instance.
type totpReplayGuard struct {
	mu   sync.Mutex
	last map[string]totpUse
}

type totpUse struct {
	code string
	at   time.Time
}

var totpGuard = totpReplayGuard{last: make(map[string]totpUse)}

// checkAndRemember retourne false si le même code vient d'être consommé pour cet utilisateur.
func (g *totpReplayGuard) checkAndRemember(userID, code string) bool {
	const replayWindow = 2 * time.Minute
	now := time.Now()
	g.mu.Lock()
	defer g.mu.Unlock()
	if use, ok := g.last[userID]; ok && use.code == code && now.Sub(use.at) < replayWindow {
		return false
	}
	// GC opportuniste.
	for k, use := range g.last {
		if now.Sub(use.at) >= replayWindow {
			delete(g.last, k)
		}
	}
	g.last[userID] = totpUse{code: code, at: now}
	return true
}

type verify2FAReq struct {
	MFAToken string `json:"mfaToken"`
	Code     string `json:"code"`
}

func (a *API) verify2FA(w http.ResponseWriter, r *http.Request) {
	var req verify2FAReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.MFAToken == "" || req.Code == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "mfa_fields_required")
		return
	}
	id, err := a.tokens.ParseMFA(req.MFAToken)
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "mfa_session_expired")
		return
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || !enabled || secret == "" {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "2fa_not_enabled")
		return
	}
	if !totp.Validate(req.Code, secret) || !totpGuard.checkAndRemember(id.UserID, req.Code) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_2fa_code")
		return
	}
	// Reload user: MFA token may carry a stale role after profile switch.
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	profileID := ""
	if active, err := a.store.GetActiveProfile(r.Context(), u.ID); err == nil {
		profileID = active.ID
	}
	pair, err := a.tokens.IssueProfile(u.ID, u.Email, u.Role, u.PracticeID, profileID, u.TokenVersion)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.store.TouchLastLogin(r.Context(), u.ID)
	httpx.WriteData(w, http.StatusOK, pair)
}

func (a *API) twoFactorStatus(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	_, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": enabled})
}

func (a *API) twoFactorSetup(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	_, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if enabled {
		writeErr(w, r, http.StatusConflict, "conflict", "2fa_already_enabled")
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "petsFollow Pro",
		AccountName: id.Email,
		SecretSize:  20,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := a.store.SetTOTPSecret(r.Context(), id.UserID, key.Secret()); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"secret":        key.Secret(),
		"otpauthUrl":    key.URL(),
		"qrCodeDataUrl": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	})
}

type twoFactorCodeReq struct {
	Code string `json:"code"`
}

func (a *API) twoFactorConfirm(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req twoFactorCodeReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Code == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "code_required")
		return
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || secret == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "setup_2fa_first")
		return
	}
	if enabled {
		writeErr(w, r, http.StatusConflict, "conflict", "2fa_already_enabled")
		return
	}
	if !totp.Validate(req.Code, secret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_2fa_code")
		return
	}
	if err := a.store.EnableTOTP(r.Context(), id.UserID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": true})
}

type twoFactorDisableReq struct {
	Code     string `json:"code"`
	Password string `json:"password"`
}

func (a *API) twoFactorDisable(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req twoFactorDisableReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Code == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "code_required")
		return
	}
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if u.PasswordHash != "" {
		if req.Password == "" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "wrong_password")
			return
		}
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || !enabled {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "2fa_not_enabled")
		return
	}
	if !totp.Validate(req.Code, secret) || !totpGuard.checkAndRemember(id.UserID, req.Code) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_2fa_code")
		return
	}
	if err := a.store.DisableTOTP(r.Context(), id.UserID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": false})
}

func validateGoogleIDToken(ctx context.Context, rawToken, clientID string) (*oidc.IDToken, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return verifier.Verify(ctx, rawToken)
}
