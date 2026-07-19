package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) requireHorsePackOwner(w http.ResponseWriter, r *http.Request, petID string) (ownerID string, ok bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return "", false
	}
	species, err := a.store.PetOwnedBy(r.Context(), petID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return "", false
	}
	if species != "horse" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "horse_pet_required")
		return "", false
	}
	has, err := a.store.HasActiveAddon(r.Context(), id.UserID, string(billing.AddonHorse))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return "", false
	}
	if !has {
		writeErr(w, r, http.StatusPaymentRequired, "addon_required", "horse_pack_required")
		return "", false
	}
	return id.UserID, true
}

func (a *API) listHorseContacts(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	ownerID, ok := a.requireHorsePackOwner(w, r, petID)
	if !ok {
		return
	}
	items, err := a.store.ListProfessionalContacts(r.Context(), petID, ownerID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type horseContactReq struct {
	Role     string `json:"role"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Notes    string `json:"notes"`
}

func (a *API) createHorseContact(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	ownerID, ok := a.requireHorsePackOwner(w, r, petID)
	if !ok {
		return
	}
	var req horseContactReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	name := strings.TrimSpace(req.FullName)
	if name == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "full_name_required")
		return
	}
	created, err := a.store.CreateProfessionalContact(r.Context(), petID, ownerID,
		strings.TrimSpace(req.Role), name, strings.TrimSpace(req.Phone), strings.TrimSpace(req.Email), strings.TrimSpace(req.Notes))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, created)
}

func (a *API) deleteHorseContact(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if err := a.store.DeleteProfessionalContact(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "contact_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) updateHorseContact(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req horseContactReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	name := strings.TrimSpace(req.FullName)
	if name == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "full_name_required")
		return
	}
	updated, err := a.store.UpdateProfessionalContact(r.Context(), chi.URLParam(r, "id"), id.UserID,
		strings.TrimSpace(req.Role), name, strings.TrimSpace(req.Phone), strings.TrimSpace(req.Email), strings.TrimSpace(req.Notes))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "contact_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) listHorseCompetitions(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	ownerID, ok := a.requireHorsePackOwner(w, r, petID)
	if !ok {
		return
	}
	items, err := a.store.ListCompetitions(r.Context(), petID, ownerID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type horseCompetitionReq struct {
	EventDate  string `json:"eventDate"`
	Title      string `json:"title"`
	Location   string `json:"location"`
	Discipline string `json:"discipline"`
	Result     string `json:"result"`
	Notes      string `json:"notes"`
}

func (a *API) createHorseCompetition(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	ownerID, ok := a.requireHorsePackOwner(w, r, petID)
	if !ok {
		return
	}
	var req horseCompetitionReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "title_required")
		return
	}
	eventDate := strings.TrimSpace(req.EventDate)
	if eventDate == "" {
		eventDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", eventDate); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_event_date")
		return
	}
	created, err := a.store.CreateCompetition(r.Context(), petID, ownerID, eventDate, title,
		strings.TrimSpace(req.Location), strings.TrimSpace(req.Discipline), strings.TrimSpace(req.Result), strings.TrimSpace(req.Notes))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, created)
}

func (a *API) deleteHorseCompetition(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if err := a.store.DeleteCompetition(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "competition_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) updateHorseCompetition(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req horseCompetitionReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "title_required")
		return
	}
	eventDate := strings.TrimSpace(req.EventDate)
	if eventDate == "" {
		eventDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", eventDate); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_event_date")
		return
	}
	updated, err := a.store.UpdateCompetition(r.Context(), chi.URLParam(r, "id"), id.UserID, eventDate, title,
		strings.TrimSpace(req.Location), strings.TrimSpace(req.Discipline), strings.TrimSpace(req.Result), strings.TrimSpace(req.Notes))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "competition_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}
