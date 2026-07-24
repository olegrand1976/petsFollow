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
	result, err := a.store.InviteClientToVetByEmail(r.Context(), id.UserID, req.Email)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
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
	return a.requirePetAccess(w, r, petID, id, store.PermRead)
}

func (a *API) requirePetAccess(w http.ResponseWriter, r *http.Request, petID string, id authx.Identity, need store.AccessPermission) (store.Pet, bool) {
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return store.Pet{}, false
	}
	ok, err := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, need)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return store.Pet{}, false
	}
	if !ok {
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
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	reminders, err := a.store.ListCareReminders(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if reminders == nil {
		reminders = []store.CareReminder{}
	}
	httpx.WriteData(w, http.StatusOK, reminders)
}

// getHousehold returns the Family/Kennel household digest (privilege gated).
func (a *API) getHousehold(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	has, err := a.store.HasHouseholdAddon(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !has {
		writeErr(w, r, http.StatusPaymentRequired, "addon_required", "household_required")
		return
	}
	hasKennel, _ := a.store.HasActiveAddon(r.Context(), id.UserID, string(billing.AddonKennel))
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
		"familyMinPets":     store.FamilyMinPets,
		"kennelMinPets":     store.KennelMinPets,
		"pack":              map[bool]string{true: "kennel", false: "family"}[hasKennel],
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
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
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
	case kernel.RoleCarePro:
		rem, rerr := a.store.GetCareReminder(r.Context(), chi.URLParam(r, "id"))
		if rerr != nil {
			if errors.Is(rerr, store.ErrNotFound) {
				writeErr(w, r, http.StatusNotFound, "not_found", "reminder_not_found")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if _, ok := a.requirePetAccess(w, r, rem.PetID, id, store.PermWriteNotes); !ok {
			return
		}
		updated, err = a.store.MarkCareReminderDoneByID(r.Context(), rem.ID)
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
	case kernel.RoleCarePro:
		rem, rerr := a.store.GetCareReminder(r.Context(), chi.URLParam(r, "id"))
		if rerr != nil {
			if errors.Is(rerr, store.ErrNotFound) {
				writeErr(w, r, http.StatusNotFound, "not_found", "reminder_not_found")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if _, ok := a.requirePetAccess(w, r, rem.PetID, id, store.PermWriteNotes); !ok {
			return
		}
		updated, err = a.store.PostponeCareReminderByID(r.Context(), rem.ID, req.Days)
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
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermRead)
	if !ok {
		return
	}
	visits, err := a.store.ListVisits(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	canNotes, err := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermWriteNotes)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !canNotes {
		for i := range visits {
			visits[i].Notes = ""
		}
	}
	httpx.WriteData(w, http.StatusOK, visits)
}

type createVisitReq struct {
	ScheduledAt     *string `json:"scheduledAt"`
	Notes           string  `json:"notes"`
	ConfirmDirect   bool    `json:"confirmDirect"`
	DurationMinutes *int    `json:"durationMinutes"`
}

func (a *API) createVisit(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
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
	if id.Role == kernel.RoleVet || id.Role == kernel.RoleCarePro {
		source = "vet"
	}

	confirmDirect := source == "vet" && req.ConfirmDirect
	if confirmDirect {
		practiceVet := id.Role == kernel.RoleVet && id.PracticeID != "" && pet.PracticeID == id.PracticeID
		if !practiceVet {
			fullOK, ferr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermFull)
			if ferr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
			if !fullOK {
				writeErr(w, r, http.StatusForbidden, "forbidden", "full_permission_required")
				return
			}
		}
	}

	enabled, slotDur, err := a.store.ClientBookingEnabled(r.Context(), pet.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Clients always use cabinet slot duration (ignore client-supplied duration).
	duration := slotDur
	if source == "vet" && req.DurationMinutes != nil {
		duration = *req.DurationMinutes
	}

	if source == "client" {
		if scheduledAt != nil {
			if !enabled {
				writeErr(w, r, http.StatusForbidden, "forbidden", "calendar_booking_disabled")
				return
			}
			if err := a.validateClientSlot(r, pet.PracticeID, *scheduledAt, duration, ""); err != nil {
				writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
				return
			}
		} else if enabled {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "slot_required")
			return
		}
	} else if scheduledAt != nil {
		// Vet-created timed visits: still block vacation / overlap (slot grid optional).
		if onVac, err := a.store.IsOnVacation(r.Context(), pet.PracticeID, *scheduledAt); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		} else if onVac {
			writeErr(w, r, http.StatusBadRequest, "on_vacation", "on_vacation")
			return
		}
		if overlap, err := a.store.HasVisitOverlap(r.Context(), pet.PracticeID, *scheduledAt, duration, ""); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		} else if overlap {
			writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
			return
		}
	}

	in := store.CreateVisitInput{
		PetID:           pet.ID,
		PracticeID:      pet.PracticeID,
		Source:          source,
		Notes:           req.Notes,
		ScheduledAt:     scheduledAt,
		DurationMinutes: &duration,
		ConfirmDirect:   confirmDirect,
	}
	if scheduledAt == nil {
		in.DurationMinutes = nil
	}
	var visit store.Visit
	if source == "client" && scheduledAt != nil {
		visit, err = a.store.CreateVisitBooked(r.Context(), in)
		if err != nil {
			if errors.Is(err, store.ErrValidation) {
				writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	} else {
		visit, err = a.store.CreateVisit(r.Context(), in)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	if source == "client" {
		a.notifyVetsVisitRequest(pet, visit)
	}
	if source == "vet" && !confirmDirect && visit.Status == "requested" {
		a.pushVisitProposed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
	}
	httpx.WriteData(w, http.StatusCreated, visit)
}

func (a *API) validateClientSlot(r *http.Request, practiceID string, start time.Time, duration int, excludeID string) error {
	onVac, err := a.store.IsOnVacation(r.Context(), practiceID, start)
	if err != nil {
		return errors.New("internal")
	}
	if onVac {
		return errors.New("on_vacation")
	}
	overlap, err := a.store.HasVisitOverlap(r.Context(), practiceID, start, duration, excludeID)
	if err != nil {
		return errors.New("internal")
	}
	if overlap {
		return errors.New("slot_taken")
	}
	slots, err := a.store.ListAvailableSlots(r.Context(), practiceID, start.Add(-time.Minute), start.Add(24*time.Hour))
	if err != nil {
		return errors.New("internal")
	}
	for _, sl := range slots {
		if sl.Start.Equal(start.UTC()) || sl.Start.Equal(start) {
			return nil
		}
		// tolerate small clock skew
		if sl.Start.Sub(start).Abs() < time.Second {
			return nil
		}
	}
	return errors.New("slot_unavailable")
}

func (a *API) listVetVisits(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" || status == "pending" {
		visits, err := a.store.ListPracticePendingVetActions(r.Context(), id.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, visits)
		return
	}
	switch status {
	case "requested", "confirmed", "done", "cancelled", "reschedule_pending":
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
	Status              string  `json:"status"`
	Action              string  `json:"action"`
	ProposedScheduledAt *string `json:"proposedScheduledAt"`
}

func (a *API) updateVisit(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req updateVisitReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
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
	case kernel.RoleVet:
		if pet.PracticeID != id.PracticeID {
			ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermFull)
			if aerr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
			if !ok {
				writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
				return
			}
		}
	case kernel.RoleCarePro:
		ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermFull)
		if aerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !ok {
			writeErr(w, r, http.StatusForbidden, "forbidden", "full_permission_required")
			return
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	actsAsVet := id.Role == kernel.RoleVet || id.Role == kernel.RoleCarePro

	action := req.Action
	if action == "" {
		switch req.Status {
		case "confirmed":
			action = "confirm"
		case "cancelled":
			action = "cancel"
		case "done":
			action = "done"
		case "requested":
			action = "reopen"
		default:
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_action")
			return
		}
	}

	var updated store.Visit
	switch action {
	case "confirm":
		if visit.Status == "reschedule_pending" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "use_accept_reschedule")
			return
		}
		if visit.Status != "requested" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		expected := "vet"
		if id.Role == kernel.RoleClient {
			expected = "client"
		}
		if visit.PendingActionBy == nil || *visit.PendingActionBy != expected {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_turn")
			return
		}
		updated, err = a.store.ConfirmVisit(r.Context(), visit.ID)
		if err == nil {
			a.pushVisitConfirmed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
		}
	case "cancel":
		if visit.Status != "requested" && visit.Status != "confirmed" && visit.Status != "reschedule_pending" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		updated, err = a.store.UpdateVisitStatus(r.Context(), visit.ID, "cancelled")
	case "done":
		if !actsAsVet {
			writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
			return
		}
		if visit.Status != "confirmed" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		updated, err = a.store.UpdateVisitStatus(r.Context(), visit.ID, "done")
	case "propose_reschedule":
		if visit.Status != "requested" && visit.Status != "confirmed" && visit.Status != "reschedule_pending" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		if req.ProposedScheduledAt == nil || *req.ProposedScheduledAt == "" {
			writeErr(w, r, http.StatusBadRequest, "proposed_required", "proposed_required")
			return
		}
		proposed, perr := time.Parse(time.RFC3339, *req.ProposedScheduledAt)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "invalid_proposed", "invalid_proposed")
			return
		}
		dur := 30
		if visit.DurationMinutes != nil {
			dur = *visit.DurationMinutes
		} else if _, slotDur, e := a.store.ClientBookingEnabled(r.Context(), pet.PracticeID); e == nil {
			dur = slotDur
		}
		overlap, oerr := a.store.HasVisitOverlap(r.Context(), pet.PracticeID, proposed, dur, visit.ID)
		if oerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if overlap {
			writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
			return
		}
		onVac, verr := a.store.IsOnVacation(r.Context(), pet.PracticeID, proposed)
		if verr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if onVac {
			writeErr(w, r, http.StatusBadRequest, "on_vacation", "on_vacation")
			return
		}
		pendingBy := "client"
		if id.Role == kernel.RoleClient {
			pendingBy = "vet"
		}
		updated, err = a.store.ProposeReschedule(r.Context(), visit.ID, proposed, pendingBy)
		if err == nil {
			if pendingBy == "client" {
				a.pushVisitReschedule(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
			} else {
				a.notifyVetsVisitRequest(pet, updated)
			}
		}
	case "accept_reschedule":
		if visit.PendingActionBy == nil {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_turn")
			return
		}
		if (id.Role == kernel.RoleClient && *visit.PendingActionBy != "client") ||
			(actsAsVet && *visit.PendingActionBy != "vet") {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_turn")
			return
		}
		if visit.ProposedScheduledAt != nil {
			dur := 30
			if visit.DurationMinutes != nil {
				dur = *visit.DurationMinutes
			}
			overlap, oerr := a.store.HasVisitOverlap(r.Context(), pet.PracticeID, *visit.ProposedScheduledAt, dur, visit.ID)
			if oerr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
			if overlap {
				writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
				return
			}
		}
		updated, err = a.store.AcceptReschedule(r.Context(), visit.ID)
		if err == nil {
			a.pushVisitConfirmed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
		}
	case "reject_reschedule":
		if visit.PendingActionBy == nil {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_turn")
			return
		}
		if (id.Role == kernel.RoleClient && *visit.PendingActionBy != "client") ||
			(actsAsVet && *visit.PendingActionBy != "vet") {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_turn")
			return
		}
		updated, err = a.store.RejectReschedule(r.Context(), visit.ID)
	case "reopen":
		if id.Role != kernel.RoleClient {
			writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
			return
		}
		if visit.Status != "cancelled" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		updated, err = a.store.ReopenVisitAsRequested(r.Context(), visit.ID)
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_action")
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
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
	// Re-enable discovery journey if the client turns the preference back on.
	if prefs.Discovery {
		_ = a.store.ResumeEmailJourney(r.Context(), id.UserID)
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
