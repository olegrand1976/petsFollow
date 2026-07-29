package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
	Email     string `json:"email"`
	VetUserID string `json:"vetUserId"`
}

func (a *API) lookupVets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if a.vetLookupRL != nil && !a.vetLookupRL.Allow("lookup:"+id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	hits, err := a.store.LookupVets(r.Context(), q, 10)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if hits == nil {
		hits = []store.VetLookupHit{}
	}
	httpx.WriteData(w, http.StatusOK, hits)
}

func (a *API) inviteVet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req inviteVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var result store.VetInviteResult
	switch {
	case strings.TrimSpace(req.VetUserID) != "":
		result, err = a.store.InviteClientToVetByID(r.Context(), id.UserID, req.VetUserID)
	case strings.TrimSpace(req.Email) != "":
		result, err = a.store.InviteClientToVetByEmail(r.Context(), id.UserID, req.Email)
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
}

type suggestVetReq struct {
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FullName     string `json:"fullName"`
	PracticeName string `json:"practiceName"`
}

func (a *API) suggestVet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if a.vetSuggestRL != nil && !a.vetSuggestRL.Allow("suggest:"+id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
		return
	}
	var req suggestVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	phone := strings.TrimSpace(req.Phone)
	if email == "" || phone == "" || !strings.Contains(email, "@") {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	// If the email already matches a vet on the platform, create a real invite instead.
	existing, ierr := a.store.InviteClientToVetByEmail(r.Context(), id.UserID, email)
	if ierr != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if existing.Found {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":       "invited",
			"found":        true,
			"practiceName": existing.PracticeName,
			"vetFullName":  existing.VetFullName,
		})
		return
	}
	lead, err := a.store.CreateVetLead(r.Context(), id.UserID, store.VetLeadInput{
		Email: email, Phone: phone, FullName: req.FullName, PracticeName: req.PracticeName,
	})
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	clientName := id.Email
	if u, uerr := a.store.GetUserByID(r.Context(), id.UserID); uerr == nil && u.FullName != "" {
		clientName = u.FullName
	}
	if to := strings.TrimSpace(a.cfg.OpsNotifyEmail); to != "" && a.notifier != nil {
		_ = a.notifier.SendVetLeadNotify(to, clientName, id.Email, lead.Email, lead.Phone, lead.FullName, lead.PracticeName)
	}
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"status": "suggested",
		"found":  false,
		"leadId": lead.ID,
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
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetOwner(w, r, petID, id.UserID); !ok {
		return
	}
	if err := a.store.SetPetPrimaryPractice(r.Context(), petID, id.UserID, req.PracticeID); err != nil {
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
	_ = a.store.StampClientPracticeIfEmpty(r.Context(), id.UserID, req.PracticeID)
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

// requirePetPractice rejects cabinet-scoped mutations when the pet has no linked practice yet.
func (a *API) requirePetPractice(w http.ResponseWriter, r *http.Request, pet store.Pet) bool {
	if strings.TrimSpace(pet.PracticeID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "vet_link_required")
		return false
	}
	return true
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
	if kernel.IsPracticeStaff(id.Role) && !a.allowPracticePerm(r, id, "pets.read") {
		writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
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

// getHousehold returns the household digest (all clients — no addon gate).
func (a *API) getHousehold(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
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
	// Le pack découle du nombre d'animaux (addons plus vendus — cf. politique tarifaire),
	// plus de dépendance à l'addon kennel legacy.
	pack := "standard"
	if len(pets) >= store.KennelMinPets {
		pack = "kennel"
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"familyMinPets":     store.FamilyMinPets,
		"kennelMinPets":     store.KennelMinPets,
		"pack":              pack,
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
	if id.Role == kernel.RoleClient && !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	if !a.requirePetPractice(w, r, pet) {
		return
	}
	if kernel.IsPracticeStaff(id.Role) {
		if !a.checkPracticePerm(w, r, id, "care.manage") {
			return
		}
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

	needsHorse := reminderType == "farrier" || reminderType == "fecal_egg"
	if needsHorse && pet.Species != "horse" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "horse_pet_required")
		return
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
	case kernel.RoleClient, kernel.RoleCarePro:
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
		if id.Role == kernel.RoleClient && !a.requirePremiumAccess(w, r, rem.PetID) {
			return
		}
		updated, err = a.store.MarkCareReminderDoneByID(r.Context(), rem.ID)
	default:
		if !a.checkPracticePerm(w, r, id, "care.manage") {
			return
		}
		updated, err = a.store.MarkCareReminderDoneByPractice(r.Context(), chi.URLParam(r, "id"), id.PracticeID)
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
	case kernel.RoleClient, kernel.RoleCarePro:
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
		if id.Role == kernel.RoleClient && !a.requirePremiumAccess(w, r, rem.PetID) {
			return
		}
		updated, err = a.store.PostponeCareReminderByID(r.Context(), rem.ID, req.Days)
	default:
		if !a.checkPracticePerm(w, r, id, "care.manage") {
			return
		}
		updated, err = a.store.PostponeCareReminderByPractice(r.Context(), chi.URLParam(r, "id"), id.PracticeID, req.Days)
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
	if err := a.store.AttachPreconsultStatuses(r.Context(), visits); err != nil {
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
	// hasFinalReport drives the client consultation CTA — owner-only (share/read are owner-gated).
	if id.Role == kernel.RoleClient && pet.OwnerUserID == id.UserID {
		if err := a.store.AttachFinalReportFlags(r.Context(), visits); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	httpx.WriteData(w, http.StatusOK, visits)
}

type createVisitReq struct {
	ScheduledAt       *string `json:"scheduledAt"`
	Notes             string  `json:"notes"`
	ConfirmDirect     bool    `json:"confirmDirect"`
	DurationMinutes   *int    `json:"durationMinutes"`
	VisitTypeID       *string `json:"visitTypeId"`
	RequestPreconsult bool    `json:"requestPreconsult"`
	// SilentConfirm skips client push/email when confirming immediately (walk-in consultation).
	SilentConfirm bool `json:"silentConfirm"`
	// ConsultationSession marks a walk-in CR flow (excluded from agenda overlap).
	ConsultationSession bool `json:"consultationSession"`
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
	if id.Role == kernel.RoleClient && !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	if !a.requirePetPractice(w, r, pet) {
		return
	}
	var req createVisitReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		t, perr := time.Parse(time.RFC3339, *req.ScheduledAt)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_scheduled_at")
			return
		}
		scheduledAt = &t
	}
	source := "client"
	switch {
	case id.Role == kernel.RoleCarePro:
		source = "care_pro"
	case kernel.IsPracticeStaff(id.Role):
		source = "vet"
	}
	actsAsPro := source == "vet" || source == "care_pro"

	confirmDirect := actsAsPro && req.ConfirmDirect
	if confirmDirect {
		// Practice staff with calendar.manage, pet full ACL, or care_pro terrain
		// (write_notes already checked via requirePetAccess) may confirm immediately.
		practiceStaff := a.allowPracticePerm(r, id, "calendar.manage") && id.PracticeID != "" && pet.PracticeID == id.PracticeID
		careProTerrain := id.Role == kernel.RoleCarePro
		if !practiceStaff && !careProTerrain {
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

	consultationSession := actsAsPro && req.ConsultationSession
	if consultationSession {
		if !confirmDirect {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_requires_confirm")
			return
		}
		if scheduledAt == nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_requires_schedule")
			return
		}
		// Walk-in only: reject far-future slots used to bypass agenda overlap.
		if scheduledAt.Sub(time.Now()).Abs() > 30*time.Minute {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_stale")
			return
		}
	}

	enabled, slotDur, err := a.store.ClientBookingEnabled(r.Context(), pet.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Clients always use cabinet slot duration (ignore client-supplied duration / type).
	duration := slotDur
	var visitTypeID *string
	if actsAsPro {
		if req.VisitTypeID != nil && strings.TrimSpace(*req.VisitTypeID) != "" {
			vt, verr := a.store.GetVisitType(r.Context(), pet.PracticeID, strings.TrimSpace(*req.VisitTypeID))
			if verr != nil {
				if errors.Is(verr, store.ErrNotFound) {
					writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_visit_type")
					return
				}
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
			if !vt.IsActive {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "visit_type_inactive")
				return
			}
			duration = vt.DurationMinutes
			idCopy := vt.ID
			visitTypeID = &idCopy
		} else if req.DurationMinutes != nil {
			duration = *req.DurationMinutes
		}
	}
	if _, err := store.NormalizeVisitDuration(duration); err != nil {
		if actsAsPro {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_duration")
			return
		}
		duration = 30
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
	} else if scheduledAt != nil && !consultationSession {
		// Timed RDV: block vacation (overlap under lock in CreateVisitBooked).
		// Walk-in consultation sessions skip vacation — terrain / cabinet immédiat.
		if onVac, err := a.store.IsOnVacation(r.Context(), pet.PracticeID, *scheduledAt); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		} else if onVac {
			writeErr(w, r, http.StatusBadRequest, "on_vacation", "on_vacation")
			return
		}
	}

	in := store.CreateVisitInput{
		PetID:               pet.ID,
		PracticeID:          pet.PracticeID,
		Source:              source,
		Notes:               req.Notes,
		ScheduledAt:         scheduledAt,
		DurationMinutes:     &duration,
		VisitTypeID:         visitTypeID,
		ConfirmDirect:       confirmDirect,
		ConsultationSession: consultationSession,
	}
	if req.RequestPreconsult && source == "vet" && a.allowPracticePerm(r, id, "calendar.manage") {
		in.RequestPreconsult = true
	}
	if scheduledAt == nil {
		in.DurationMinutes = nil
	}
	var visit store.Visit
	// Walk-in sessions skip agenda lock/overlap; timed RDV go through CreateVisitBooked.
	if scheduledAt != nil && !consultationSession {
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
	if actsAsPro && !confirmDirect && visit.Status == "requested" {
		a.pushVisitProposed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
	}
	if visit.Status == "confirmed" && !(req.SilentConfirm || in.ConsultationSession) {
		a.onVisitConfirmed(pet, visit)
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
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" || status == "pending" {
		visits, err := a.store.ListPracticePendingVetActions(r.Context(), id.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		_ = a.store.AttachPreconsultStatuses(r.Context(), visits)
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
	_ = a.store.AttachPreconsultStatuses(r.Context(), visits)
	httpx.WriteData(w, http.StatusOK, visits)
}

func (a *API) listVetConsultations(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	q := r.URL.Query()
	f := store.ListConsultationsFilter{
		Status: strings.TrimSpace(q.Get("status")),
		Query:  strings.TrimSpace(q.Get("q")),
	}
	switch f.Status {
	case "", "confirmed", "done", "cancelled", "requested", "reschedule_pending":
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	if raw := strings.TrimSpace(q.Get("from")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_from")
			return
		}
		f.From = &t
	}
	if raw := strings.TrimSpace(q.Get("to")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_to")
			return
		}
		f.To = &t
	}
	if raw := strings.TrimSpace(q.Get("hasAudio")); raw != "" {
		v := raw == "1" || strings.EqualFold(raw, "true")
		f.HasAudio = &v
	}
	if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_limit")
			return
		}
		f.Limit = n
	}
	if raw := strings.TrimSpace(q.Get("offset")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_offset")
			return
		}
		f.Offset = n
	}
	items, err := a.store.ListPracticeConsultations(r.Context(), id.PracticeID, f)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) listVetOverdueCare(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "care.manage")
	if !ok {
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
	RequestPreconsult   *bool   `json:"requestPreconsult"`
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
		ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermWriteNotes)
		if aerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !ok {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
	case kernel.RoleCarePro:
		ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermWriteNotes)
		if aerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !ok {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
	default:
		if !a.checkPracticePerm(w, r, id, "calendar.manage") {
			return
		}
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
	}

	actsAsVet := id.Role == kernel.RoleCarePro || kernel.IsPracticeStaff(id.Role)

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

	// Care_pro may complete shared terrain visits (done) but must not cancel/reschedule
	// cabinet or client bookings — only visits they originated (source=care_pro).
	if id.Role == kernel.RoleCarePro {
		switch action {
		case "cancel", "propose_reschedule", "accept_reschedule", "reject_reschedule", "reopen", "confirm":
			if visit.Source != "care_pro" {
				writeErr(w, r, http.StatusForbidden, "forbidden", "care_pro_visit_only")
				return
			}
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
		if req.RequestPreconsult != nil && actsAsVet && a.allowPracticePerm(r, id, "calendar.manage") {
			if err := a.store.SetVisitRequestPreconsult(r.Context(), visit.ID, *req.RequestPreconsult); err != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
		}
		updated, err = a.store.ConfirmVisit(r.Context(), visit.ID)
		if err == nil {
			if full, gerr := a.store.GetVisit(r.Context(), visit.ID); gerr == nil {
				updated = full
			}
			a.onVisitConfirmed(pet, updated)
		}
	case "cancel":
		if visit.Status != "requested" && visit.Status != "confirmed" && visit.Status != "reschedule_pending" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		// Walk-in discard must not race past a CR save (empty Ensure draft is OK to cancel).
		if visit.ConsultationSession {
			hasReport, herr := a.store.VisitHasPersistedReport(r.Context(), visit.ID)
			if herr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
			if hasReport {
				writeErr(w, r, http.StatusConflict, "conflict", "consultation_has_report")
				return
			}
		}
		updated, err = a.store.UpdateVisitStatus(r.Context(), visit.ID, "cancelled")
	case "done":
		if !actsAsVet {
			writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
			return
		}
		if visit.Status == "done" {
			updated = visit
			break
		}
		if visit.Status != "confirmed" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		updated, err = a.store.UpdateVisitStatus(r.Context(), visit.ID, "done")
	case "propose_reschedule":
		if visit.ConsultationSession {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_not_reschedulable")
			return
		}
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
		if visit.ConsultationSession {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_not_reschedulable")
			return
		}
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
			a.onVisitRescheduleAccepted(pet, updated)
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

// softDeleteVisit hides a walk-in consultation from /consultations history (deleted_at).
// Allowed even when done / with persisted CR — unlike cancel.
func (a *API) softDeleteVisit(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	// Soft-delete removes clinical history visibility — need calendar + clinical write.
	if !a.checkPracticePerm(w, r, id, "calendar.manage") {
		return
	}
	if !a.checkPracticePerm(w, r, id, "pets.write_clinical") {
		return
	}
	visit, err := a.store.GetVisit(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return
	}
	if visit.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	if !visit.ConsultationSession {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "not_consultation_session")
		return
	}
	updated, err := a.store.SoftDeleteVisit(r.Context(), visit.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
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
	id, ok := a.requirePracticePerm(w, r, "shares.manage")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "shares.manage")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "shares.manage")
	if !ok {
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
