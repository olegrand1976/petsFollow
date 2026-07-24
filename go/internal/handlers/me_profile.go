package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type updateMeReq struct {
	FullName string `json:"fullName"`
}

type changePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (a *API) updateMe(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req updateMeReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.FullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "full_name_required")
		return
	}
	if err := a.store.UpdateUserFullName(r.Context(), id.UserID, req.FullName); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	data, err := a.store.GetUserMe(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, data)
}

func (a *API) changeMePassword(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req changePasswordReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if len(req.NewPassword) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return
	}
	if err := a.store.ChangeUserPassword(r.Context(), id.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "wrong_password")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) deleteMe(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	switch id.Role {
	case kernel.RoleClient:
		a.deleteClientMe(w, r, id.UserID)
	case kernel.RoleVet, kernel.RoleCommercial, kernel.RoleCommercialManager, kernel.RoleCarePro:
		a.deleteProMe(w, r, id.UserID)
	default:
		// admin : pas d'auto-suppression (dernier accès plateforme).
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	}
}

// deleteClientMe — effacement RGPD complet : DB, puis purge best-effort des
// médias (GCS/local) et annulation des abonnements Stripe.
func (a *API) deleteClientMe(w http.ResponseWriter, r *http.Request, userID string) {
	if err := a.purgeClientAccount(r.Context(), userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

// deleteProMe — anonymisation du compte Pro + purge de l'avatar.
func (a *API) deleteProMe(w http.ResponseWriter, r *http.Request, userID string) {
	if err := a.anonymizeProAccount(r.Context(), userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

// exportMe — portabilité RGPD (art. 20) : export JSON des données du compte.
func (a *API) exportMe(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	data, err := a.store.ExportUserData(r.Context(), id.UserID, id.Role == kernel.RoleClient)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="petsfollow-export.json"`)
	httpx.WriteData(w, http.StatusOK, data)
}

func (a *API) purgeMediaObjects(ctx context.Context, keys []string) {
	if a.media == nil {
		return
	}
	for _, k := range keys {
		if k == "" {
			continue
		}
		if err := a.media.Delete(ctx, k); err != nil {
			fmt.Printf("account deletion: delete media object %s failed: %v\n", k, err)
		}
	}
}

type emailPrefsReq struct {
	EmailOnMessage       bool `json:"emailOnMessage"`
	EmailOnHeartRate     bool `json:"emailOnHeartrate"`
	EmailOnVisitRequest  bool `json:"emailOnVisitRequest"`
}

func (a *API) getVetEmailPrefs(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	prefs, err := a.store.GetEmailPrefs(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"emailOnMessage":      prefs.OnMessage,
		"emailOnHeartrate":    prefs.OnHeartRate,
		"emailOnVisitRequest": prefs.OnVisitRequest,
	})
}

func (a *API) updateVetEmailPrefs(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req emailPrefsReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.UpdateEmailPrefs(r.Context(), id.UserID, req.EmailOnMessage, req.EmailOnHeartRate, req.EmailOnVisitRequest); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"emailOnMessage":      req.EmailOnMessage,
		"emailOnHeartrate":    req.EmailOnHeartRate,
		"emailOnVisitRequest": req.EmailOnVisitRequest,
	})
}
