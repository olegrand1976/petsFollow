package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
	"github.com/coreos/go-oidc/v3/oidc"
)

func (a *API) registerAuthRoutes(r chi.Router) {
	r.Post("/auth/google", a.googleLogin)
	r.Post("/auth/2fa/verify", a.verify2FA)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/auth/2fa/status", a.twoFactorStatus)
		pr.Post("/auth/2fa/setup", a.twoFactorSetup)
		pr.Post("/auth/2fa/confirm", a.twoFactorConfirm)
		pr.Post("/auth/2fa/disable", a.twoFactorDisable)
	})
}

type googleLoginReq struct {
	IDToken string `json:"idToken"`
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
	email := gClaims.Email
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

	u, err := a.resolveGoogleUser(r, email, name, payload.Subject)
	if err != nil {
		a.writeGoogleAuthError(w, r, err)
		return
	}
	a.issueLoginResponse(w, r, u)
}

func (a *API) resolveGoogleUser(r *http.Request, email, fullName, googleSub string) (store.User, error) {
	ctx := r.Context()
	if u, err := a.store.GetUserByGoogleSub(ctx, googleSub); err == nil {
		return u, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	u, err := a.store.GetUserByEmail(ctx, email)
	if err == nil {
		switch u.Role {
		case kernel.RoleClient:
			return store.User{}, errGoogleProOnly
		case kernel.RoleAdmin:
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
		case kernel.RoleVet:
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
		default:
			return store.User{}, errGoogleProOnly
		}
	}
	if !errors.Is(err, store.ErrNotFound) {
		return store.User{}, err
	}

	practiceName := fmt.Sprintf("Cabinet %s", strings.Split(fullName, " ")[0])
	locale := localeOf(r)
	return a.store.RegisterGoogleVet(ctx, store.RegisterGoogleVetInput{
		Email: email, FullName: fullName, GoogleSub: googleSub, PracticeName: practiceName,
		PreferredLocale: locale, AutoReplyDefault: t(r, "defaults.auto_reply_unavailable", nil),
	})
}

var (
	errGoogleProOnly           = errors.New("google pro only")
	errGoogleAccountMismatch   = errors.New("google account mismatch")
)

func (a *API) writeGoogleAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errGoogleProOnly):
		writeErr(w, r, http.StatusForbidden, "forbidden", "google_pro_only")
	case errors.Is(err, errGoogleAccountMismatch):
		writeErr(w, r, http.StatusConflict, "conflict", "google_account_mismatch")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}

func (a *API) issueLoginResponse(w http.ResponseWriter, r *http.Request, u store.User) {
	if u.Role == kernel.RoleVet && u.EmailVerifiedAt == nil {
		writeErr(w, r, http.StatusForbidden, "email_not_verified", "email_not_verified")
		return
	}
	if u.TOTPEnabled {
		mfa, err := a.tokens.IssueMFA(u.ID, u.Email, u.Role, u.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, mfa)
		return
	}
	pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pair)
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
	if !totp.Validate(req.Code, secret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_2fa_code")
		return
	}
	pair, err := a.tokens.Issue(id.UserID, id.Email, id.Role, id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
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
		"secret":       key.Secret(),
		"otpauthUrl":   key.URL(),
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
	if !totp.Validate(req.Code, secret) {
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
