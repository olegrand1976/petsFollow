package handlers

import (
	"errors"
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
	if req.CurrentPassword == "" {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "wrong_password")
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
	if id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if err := a.store.DeleteClientAccount(r.Context(), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

type emailPrefsReq struct {
	EmailOnMessage    bool `json:"emailOnMessage"`
	EmailOnHeartRate  bool `json:"emailOnHeartrate"`
}

func (a *API) getVetEmailPrefs(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	onMsg, onHR, err := a.store.GetEmailPrefs(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"emailOnMessage":   onMsg,
		"emailOnHeartrate": onHR,
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
	if err := a.store.UpdateEmailPrefs(r.Context(), id.UserID, req.EmailOnMessage, req.EmailOnHeartRate); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"emailOnMessage":   req.EmailOnMessage,
		"emailOnHeartrate": req.EmailOnHeartRate,
	})
}
