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
		httpx.WriteError(w, http.StatusNotImplemented, "not_configured", "Google OAuth non configuré")
		return
	}
	var req googleLoginReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.IDToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "idToken requis")
		return
	}

	payload, err := validateGoogleIDToken(r.Context(), req.IDToken, a.cfg.GoogleOAuthClientID)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "token Google invalide")
		return
	}
	var gClaims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := payload.Claims(&gClaims); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "token Google invalide")
		return
	}
	email := gClaims.Email
	name := gClaims.Name
	if email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "email Google manquant")
		return
	}
	if !gClaims.EmailVerified {
		httpx.WriteError(w, http.StatusForbidden, "email_not_verified", "email Google non vérifié")
		return
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	u, err := a.resolveGoogleUser(r.Context(), email, name, payload.Subject)
	if err != nil {
		a.writeGoogleAuthError(w, err)
		return
	}
	a.issueLoginResponse(w, u)
}

func (a *API) resolveGoogleUser(ctx context.Context, email, fullName, googleSub string) (store.User, error) {
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
	return a.store.RegisterGoogleVet(ctx, store.RegisterGoogleVetInput{
		Email: email, FullName: fullName, GoogleSub: googleSub, PracticeName: practiceName,
	})
}

var (
	errGoogleProOnly           = errors.New("google pro only")
	errGoogleAccountMismatch   = errors.New("google account mismatch")
)

func (a *API) writeGoogleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errGoogleProOnly):
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "Google réservé aux profils Pro")
	case errors.Is(err, errGoogleAccountMismatch):
		httpx.WriteError(w, http.StatusConflict, "conflict", "ce compte Google est déjà lié à un autre utilisateur")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
	}
}

func (a *API) issueLoginResponse(w http.ResponseWriter, u store.User) {
	if u.Role == kernel.RoleVet && u.EmailVerifiedAt == nil {
		httpx.WriteError(w, http.StatusForbidden, "email_not_verified", "confirmez votre email avant de vous connecter")
		return
	}
	if u.TOTPEnabled {
		mfa, err := a.tokens.IssueMFA(u.ID, u.Email, u.Role, u.PracticeID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, mfa)
		return
	}
	pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "mfaToken et code requis")
		return
	}
	id, err := a.tokens.ParseMFA(req.MFAToken)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "session 2FA expirée")
		return
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || !enabled || secret == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "2FA non activée")
		return
	}
	if !totp.Validate(req.Code, secret) {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "code 2FA invalide")
		return
	}
	pair, err := a.tokens.Issue(id.UserID, id.Email, id.Role, id.PracticeID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, pair)
}

func (a *API) twoFactorStatus(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	_, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": enabled})
}

func (a *API) twoFactorSetup(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	_, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if enabled {
		httpx.WriteError(w, http.StatusConflict, "conflict", "2FA déjà activée")
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "petsFollow Pro",
		AccountName: id.Email,
		SecretSize:  20,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if err := a.store.SetTOTPSecret(r.Context(), id.UserID, key.Secret()); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	var req twoFactorCodeReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Code == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "code requis")
		return
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || secret == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "lancez d'abord la configuration 2FA")
		return
	}
	if enabled {
		httpx.WriteError(w, http.StatusConflict, "conflict", "2FA déjà activée")
		return
	}
	if !totp.Validate(req.Code, secret) {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "code 2FA invalide")
		return
	}
	if err := a.store.EnableTOTP(r.Context(), id.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	var req twoFactorDisableReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Code == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "code requis")
		return
	}
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if u.PasswordHash != "" {
		if req.Password == "" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "mot de passe incorrect")
			return
		}
	}
	secret, enabled, err := a.store.GetTOTPSecret(r.Context(), id.UserID)
	if err != nil || !enabled {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "2FA non activée")
		return
	}
	if !totp.Validate(req.Code, secret) {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "code 2FA invalide")
		return
	}
	if err := a.store.DisableTOTP(r.Context(), id.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
