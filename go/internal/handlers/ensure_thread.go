package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type ensureThreadReq struct {
	ClientUserID string `json:"clientUserId"`
	PracticeID   string `json:"practiceId"`
	PetID        string `json:"petId"`
}

func (a *API) ensureThread(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}

	var req ensureThreadReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}

	switch {
	case kernel.IsPracticeStaff(id.Role):
		a.ensureThreadStaff(w, r, id, req)
	case id.Role == kernel.RoleClient:
		a.ensureThreadClient(w, r, id, req)
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	}
}

func (a *API) ensureThreadStaff(w http.ResponseWriter, r *http.Request, id authx.Identity, req ensureThreadReq) {
	if !a.allowPracticePerm(r, id, "messaging") {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if strings.TrimSpace(req.ClientUserID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "client_required")
		return
	}
	clientID := strings.TrimSpace(req.ClientUserID)
	client, err := a.store.GetUserByID(r.Context(), clientID)
	if err != nil || client.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
		return
	}
	pets, err := a.store.ListPetsByClientForVet(r.Context(), id.PracticeID, clientID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(pets) == 0 && client.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	petID := strings.TrimSpace(req.PetID)
	thread, err := a.store.GetOrCreateThreadForPet(r.Context(), id.PracticeID, clientID, id.UserID, petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, thread)
}

func (a *API) ensureThreadClient(w http.ResponseWriter, r *http.Request, id authx.Identity, req ensureThreadReq) {
	practiceID := strings.TrimSpace(req.PracticeID)
	petID := strings.TrimSpace(req.PetID)
	if practiceID == "" || petID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "practice_and_pet_required")
		return
	}
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil || pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if pid := strings.TrimSpace(pet.PracticeID); pid != "" && pid != practiceID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	vetID, err := a.store.GetVetForClient(r.Context(), id.UserID, practiceID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusForbidden, "forbidden", "vet_link_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	thread, err := a.store.GetOrCreateThreadForPet(r.Context(), practiceID, id.UserID, vetID, petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, thread)
}
