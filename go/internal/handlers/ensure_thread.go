package handlers

import (
	"net/http"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type ensureThreadReq struct {
	ClientUserID string `json:"clientUserId"`
}

func (a *API) ensureThread(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req ensureThreadReq
	if err := httpx.DecodeJSON(r, &req); err != nil || strings.TrimSpace(req.ClientUserID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "client_required")
		return
	}
	clientID := strings.TrimSpace(req.ClientUserID)
	client, err := a.store.GetUserByID(r.Context(), clientID)
	if err != nil || client.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
		return
	}
	// Prefer pets of this practice as access proof (covers multi-practice links).
	pets, err := a.store.ListPetsByClientForVet(r.Context(), id.PracticeID, clientID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(pets) == 0 && client.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	thread, err := a.store.GetOrCreateThread(r.Context(), id.PracticeID, clientID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, thread)
}
