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

func (a *API) listMyVets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	vets, err := a.store.ListClientVets(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, vets)
}

type inviteVetReq struct {
	Email string `json:"email"`
}

func (a *API) inviteVet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req inviteVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Email == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	_ = a.store.InviteClientToVetByEmail(r.Context(), id.UserID, req.Email)
	httpx.WriteData(w, http.StatusOK, map[string]string{
		"message": t(r, "success.vet_invite_sent", nil),
	})
}

type primaryPracticeReq struct {
	PracticeID string `json:"practiceId"`
}

func (a *API) setPetPrimaryPractice(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req primaryPracticeReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.PracticeID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.SetPetPrimaryPractice(r.Context(), chi.URLParam(r, "petID"), id.UserID, req.PracticeID); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, r, http.StatusForbidden, "forbidden", "cannot_change_practice")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) requirePetOwner(w http.ResponseWriter, r *http.Request, petID, userID string) (store.Pet, bool) {
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return store.Pet{}, false
	}
	if pet.OwnerUserID != userID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return store.Pet{}, false
	}
	return pet, true
}

func (a *API) requirePetOwnerOrPractice(w http.ResponseWriter, r *http.Request, petID string, id authx.Identity) (store.Pet, bool) {
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return store.Pet{}, false
	}
	switch id.Role {
	case kernel.RoleClient:
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return store.Pet{}, false
		}
	case kernel.RoleVet:
		if pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return store.Pet{}, false
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return store.Pet{}, false
	}
	return pet, true
}

func (a *API) listCareReminders(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	switch id.Role {
	case kernel.RoleClient:
		if _, ok := a.requirePetOwner(w, r, petID, id.UserID); !ok {
			return
		}
	case kernel.RoleVet:
		pet, err := a.store.GetPet(r.Context(), petID)
		if err != nil || pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	reminders, err := a.store.ListCareReminders(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, reminders)
}

// getHousehold returns the Family household digest (privilege gated).
func (a *API) getHousehold(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	has, err := a.store.HasActiveAddon(r.Context(), id.UserID, string(billing.AddonFamily))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !has {
		writeErr(w, r, http.StatusPaymentRequired, "addon_required", "family_required")
		return
	}
	pets, err := a.store.ListPetsByOwner(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	upcoming, err := a.store.ListHouseholdUpcomingCare(r.Context(), id.UserID, 8)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if upcoming == nil {
		upcoming = []store.HouseholdCareItem{}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"maxPets":           store.FamilyMaxPets,
		"minPets":           store.FamilyMinPets,
		"petCount":          len(pets),
		"pets":              pets,
		"upcomingReminders": upcoming,
	})
}

type createCareReminderReq struct {
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	DueAt          *string `json:"dueAt"`
	DueDays        *int    `json:"dueDays"`
	Notes          string  `json:"notes"`
	RecurrenceDays *int    `json:"recurrenceDays"`
}

func (a *API) createCareReminder(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetOwnerOrPractice(w, r, petID, id)
	if !ok {
		return
	}
	var req createCareReminderReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	reminderType := req.Type
	if reminderType == "" {
		reminderType = "custom"
	}
	switch reminderType {
	case "vaccination", "deworming", "vet_check", "dental", "farrier", "fecal_egg", "custom", "medication":
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_care_type")
		return
	}

	ownerID := pet.OwnerUserID
	if id.Role == kernel.RoleClient {
		ownerID = id.UserID
	}
	needsCarePlus := reminderType == "custom" || reminderType == "medication"
	needsHorse := reminderType == "farrier" || reminderType == "fecal_egg"
	if needsCarePlus {
		okAddon, err := a.store.HasActiveAddon(r.Context(), ownerID, string(billing.AddonCarePlus))
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !okAddon {
			writeErr(w, r, http.StatusPaymentRequired, "addon_required", "care_plus_required")
			return
		}
	}
	if needsHorse {
		if pet.Species != "horse" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "horse_pet_required")
			return
		}
		okAddon, err := a.store.HasActiveAddon(r.Context(), ownerID, string(billing.AddonHorse))
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !okAddon {
			writeErr(w, r, http.StatusPaymentRequired, "addon_required", "horse_pack_required")
			return
		}
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = reminderType
	}
	if len(title) > 200 {
		title = title[:200]
	}
	notes := strings.TrimSpace(req.Notes)
	if len(notes) > 1000 {
		notes = notes[:1000]
	}
	dueAt := time.Now().AddDate(0, 0, 30)
	if req.DueAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.DueAt); err == nil {
			dueAt = t
		}
	} else if req.DueDays != nil && *req.DueDays > 0 {
		dueAt = time.Now().AddDate(0, 0, *req.DueDays)
	}
	created, err := a.store.CreateCareReminderFull(r.Context(), pet.ID, pet.PracticeID, reminderType, title, dueAt, notes, req.RecurrenceDays)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, created)
}

func (a *API) markCareReminderDone(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var updated store.CareReminder
	switch id.Role {
	case kernel.RoleClient:
		updated, err = a.store.MarkCareReminderDone(r.Context(), chi.URLParam(r, "id"), id.UserID)
	case kernel.RoleVet:
		updated, err = a.store.MarkCareReminderDoneByPractice(r.Context(), chi.URLParam(r, "id"), id.PracticeID)
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "reminder_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

type postponeCareReq struct {
	Days int `json:"days"`
}

func (a *API) postponeCareReminder(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req postponeCareReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Days != 7 && req.Days != 14 && req.Days != 30 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_postpone_days")
		return
	}
	var updated store.CareReminder
	switch id.Role {
	case kernel.RoleClient:
		updated, err = a.store.PostponeCareReminder(r.Context(), chi.URLParam(r, "id"), id.UserID, req.Days)
	case kernel.RoleVet:
		updated, err = a.store.PostponeCareReminderByPractice(r.Context(), chi.URLParam(r, "id"), id.PracticeID, req.Days)
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "reminder_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) listVisits(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	switch id.Role {
	case kernel.RoleClient:
		if _, ok := a.requirePetOwner(w, r, petID, id.UserID); !ok {
			return
		}
	case kernel.RoleVet:
		pet, err := a.store.GetPet(r.Context(), petID)
		if err != nil || pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	visits, err := a.store.ListVisits(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, visits)
}

type createVisitReq struct {
	ScheduledAt *string `json:"scheduledAt"`
	Notes       string  `json:"notes"`
}

func (a *API) createVisit(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetOwnerOrPractice(w, r, petID, id)
	if !ok {
		return
	}
	var req createVisitReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ScheduledAt); err == nil {
			scheduledAt = &t
		}
	}
	source := "client"
	if id.Role == kernel.RoleVet {
		source = "vet"
	}
	visit, err := a.store.CreateVisit(r.Context(), pet.ID, pet.PracticeID, source, req.Notes, scheduledAt)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, visit)
}

func (a *API) listVetVisits(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "requested"
	}
	switch status {
	case "requested", "confirmed", "done", "cancelled":
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	visits, err := a.store.ListPracticeVisitsByStatus(r.Context(), id.PracticeID, status)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, visits)
}

func (a *API) listVetOverdueCare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	items, err := a.store.ListOverdueCareReminders(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type updateVisitReq struct {
	Status string `json:"status"`
}

func (a *API) updateVisit(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req updateVisitReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Status == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	switch req.Status {
	case "requested", "confirmed", "done", "cancelled":
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	visit, err := a.store.GetVisit(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	switch id.Role {
	case kernel.RoleClient:
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
		switch req.Status {
		case "cancelled", "requested":
		default:
			writeErr(w, r, http.StatusForbidden, "forbidden", "client_cannot_set_status")
			return
		}
	case kernel.RoleVet:
		if pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	updated, err := a.store.UpdateVisitStatus(r.Context(), visit.ID, req.Status)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) getDiscovery(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	progress, err := a.store.GetDiscoveryProgress(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, progress)
}

type completeDiscoveryReq struct {
	CardKey string `json:"cardKey"`
}

func (a *API) completeDiscovery(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req completeDiscoveryReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.CardKey == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	progress, err := a.store.CompleteDiscoveryCard(r.Context(), id.UserID, req.CardKey)
	if err != nil {
		if err.Error() == "invalid_card_key" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_card_key")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, progress)
}

type deviceTokenReq struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (a *API) putDeviceToken(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req deviceTokenReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Token == "" || req.Platform == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	switch req.Platform {
	case "ios", "android", "web":
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_platform")
		return
	}
	dt, err := a.store.UpsertDeviceToken(r.Context(), id.UserID, req.Token, req.Platform)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, dt)
}

func (a *API) getClientNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	prefs, err := a.store.GetClientNotificationPrefs(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, prefs)
}

func (a *API) updateClientNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req struct {
		HR        *bool `json:"hr"`
		Care      *bool `json:"care"`
		Visits    *bool `json:"visits"`
		Messages  *bool `json:"messages"`
		Discovery *bool `json:"discovery"`
		Billing   *bool `json:"billing"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	current, err := a.store.GetClientNotificationPrefs(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if req.HR != nil {
		current.HR = *req.HR
	}
	if req.Care != nil {
		current.Care = *req.Care
	}
	if req.Visits != nil {
		current.Visits = *req.Visits
	}
	if req.Messages != nil {
		current.Messages = *req.Messages
	}
	if req.Discovery != nil {
		current.Discovery = *req.Discovery
	}
	if req.Billing != nil {
		current.Billing = *req.Billing
	}
	prefs, err := a.store.UpdateClientNotificationPrefs(r.Context(), id.UserID, current)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, prefs)
}

func (a *API) listVetLinkRequests(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	items, err := a.store.ListPendingVetLinkRequests(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) acceptVetLinkRequest(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	if err := a.store.AcceptVetLinkRequest(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "request_not_found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (a *API) rejectVetLinkRequest(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	if err := a.store.RejectVetLinkRequest(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "request_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "rejected"})
}
