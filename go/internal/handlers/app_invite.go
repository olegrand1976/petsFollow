package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"github.com/skip2/go-qrcode"
)

func (a *API) registerAppInviteRoutes(r chi.Router) {
	r.Get("/public/app-invite/{code}", a.getPublicAppInvite)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/me/app-invite", a.getMeAppInvite)
		pr.Get("/vet/app-invite", a.getVetAppInvite) // alias (vet-only) for Nuxt clients
		pr.Post("/me/vets/claim-invite", a.claimVetAppInvite)
	})
}

func (a *API) appInviteWebURL(code string) string {
	return strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/invite/" + code
}

func (a *API) appInviteDeepLink(code string) string {
	return "petsfollow://invite?code=" + code
}

func (a *API) encodeInviteQR(inviteURL string) (string, error) {
	png, err := qrcode.Encode(inviteURL, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

func (a *API) writeAppInvitePayload(w http.ResponseWriter, r *http.Request, inv store.AppInvite) {
	inviteURL := a.appInviteWebURL(inv.Code)
	qr, err := a.encodeInviteQR(inviteURL)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	downloadURL := strings.TrimSpace(a.cfg.PetsAppDownloadURL)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"code":          inv.Code,
		"role":          inv.Role,
		"inviteUrl":     inviteURL,
		"deepLink":      a.appInviteDeepLink(inv.Code),
		"downloadUrl":   downloadURL,
		"proSiteUrl":    strings.TrimRight(a.cfg.ProPublicSiteURL, "/"),
		"qrCodeDataUrl": qr,
		"practiceName":  inv.PracticeName,
		"displayName":   inv.DisplayName,
		"specialty":     inv.Specialty,
		// Compat Nuxt ProAppInviteModal
		"vetFullName": inv.DisplayName,
	})
}

func (a *API) getMeAppInvite(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	switch id.Role {
	case kernel.RoleVet, kernel.RoleCarePro, kernel.RoleCommercial, kernel.RoleCommercialManager:
		// ok
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "invite_role_denied")
		return
	}
	inv, err := a.store.EnsureAppInviteCode(r.Context(), id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrForbidden) {
			writeErr(w, r, http.StatusNotFound, "not_found", "invite_unavailable")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.writeAppInvitePayload(w, r, inv)
}

func (a *API) getVetAppInvite(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	inv, err := a.store.EnsureAppInviteCode(r.Context(), id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrForbidden) {
			writeErr(w, r, http.StatusNotFound, "not_found", "vet_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.writeAppInvitePayload(w, r, inv)
}

func (a *API) getPublicAppInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	inv, err := a.store.GetAppInviteByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "invite_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	downloadURL := strings.TrimSpace(a.cfg.PetsAppDownloadURL)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"code":         inv.Code,
		"role":         inv.Role,
		"practiceName": inv.PracticeName,
		"displayName":  inv.DisplayName,
		"vetFullName":  inv.DisplayName, // landing compat
		"specialty":    inv.Specialty,
		"downloadUrl":  downloadURL,
		"deepLink":     a.appInviteDeepLink(inv.Code),
		"inviteUrl":    a.appInviteWebURL(inv.Code),
	})
}

type claimInviteReq struct {
	Code string `json:"code"`
}

func (a *API) claimVetAppInvite(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req claimInviteReq
	if err := httpx.DecodeJSON(r, &req); err != nil || store.NormalizeInviteCode(req.Code) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "code_required")
		return
	}
	result, err := a.store.ClaimAppInvite(r.Context(), id.UserID, req.Code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "invite_not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_role")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
}

// tryClaimInvite soft-applies an invite code (invalid codes are ignored).
func (a *API) tryClaimInvite(r *http.Request, clientUserID, code string) {
	code = store.NormalizeInviteCode(code)
	if code == "" || clientUserID == "" {
		return
	}
	_, _ = a.store.ClaimAppInvite(r.Context(), clientUserID, code)
}
