package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type updateVisitNotesReq struct {
	Notes string `json:"notes"`
}

// updateVisitNotes PATCH /visits/{visitID}/notes — agenda desk note (calendar.manage).
func (a *API) updateVisitNotes(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	if !a.checkPracticePerm(w, r, id, "calendar.manage") {
		return
	}
	visit, err := a.store.GetVisit(r.Context(), chi.URLParam(r, "visitID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return
	}
	if visit.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	var req updateVisitNotesReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	updated, err := a.store.UpdateVisitNotes(r.Context(), visit.ID, strings.TrimSpace(req.Notes))
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

// visitInUnassignedQueue is true when the visit competes for the legacy single site file
// (no assignee and no room). Assigned/roomed visits may run in parallel with that file.
func visitInUnassignedQueue(v store.Visit) bool {
	return strings.TrimSpace(v.AssigneeUserID) == "" && strings.TrimSpace(v.RoomID) == ""
}

// writeVisitSlotConflicts enforces unassigned-queue and/or hard assignee/room overlaps.
// Returns false when a response was already written.
func (a *API) writeVisitSlotConflicts(
	w http.ResponseWriter, r *http.Request,
	practiceID, siteID string, v store.Visit, at time.Time, dur int, excludeID string,
) bool {
	if visitInUnassignedQueue(v) {
		overlap, oerr := a.store.HasVisitOverlap(r.Context(), practiceID, siteID, at, dur, excludeID)
		if oerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return false
		}
		if overlap {
			writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
			return false
		}
	}
	if rerr := a.store.CheckVisitResourceConflicts(r.Context(), practiceID, v.AssigneeUserID, v.RoomID, at, dur, excludeID); rerr != nil {
		if errors.Is(rerr, store.ErrValidation) {
			msg := rerr.Error()
			switch {
			case strings.Contains(msg, "assignee_busy"):
				writeErr(w, r, http.StatusConflict, "assignee_busy", "assignee_busy")
			case strings.Contains(msg, "room_busy"):
				writeErr(w, r, http.StatusConflict, "room_busy", "room_busy")
			default:
				writeErr(w, r, http.StatusConflict, "slot_taken", "slot_taken")
			}
			return false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return false
	}
	return true
}

// parseAndValidateVisitSlot shared by propose_reschedule and reschedule_direct.
func (a *API) parseAndValidateVisitSlot(
	w http.ResponseWriter, r *http.Request, pet store.Pet, visit store.Visit, proposedRaw *string,
) (time.Time, bool) {
	if visit.ConsultationSession {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "consultation_session_not_reschedulable")
		return time.Time{}, false
	}
	if visit.Status != "requested" && visit.Status != "confirmed" && visit.Status != "reschedule_pending" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return time.Time{}, false
	}
	if proposedRaw == nil || *proposedRaw == "" {
		writeErr(w, r, http.StatusBadRequest, "proposed_required", "proposed_required")
		return time.Time{}, false
	}
	proposed, perr := time.Parse(time.RFC3339, *proposedRaw)
	if perr != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_proposed", "invalid_proposed")
		return time.Time{}, false
	}
	dur := 30
	if visit.DurationMinutes != nil {
		dur = *visit.DurationMinutes
	} else if _, slotDur, e := a.store.ClientBookingEnabled(r.Context(), pet.PracticeID, visit.SiteID); e == nil {
		dur = slotDur
	}
	if !a.writeVisitSlotConflicts(w, r, pet.PracticeID, visit.SiteID, visit, proposed, dur, visit.ID) {
		return time.Time{}, false
	}
	onVac, verr := a.store.IsOnVacation(r.Context(), pet.PracticeID, visit.SiteID, proposed)
	if verr != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return time.Time{}, false
	}
	if onVac {
		writeErr(w, r, http.StatusBadRequest, "on_vacation", "on_vacation")
		return time.Time{}, false
	}
	return proposed, true
}

func (a *API) notifyClinicalStaffWaitingRoom(actorUserID string, pet store.Pet, visit store.Visit) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		ids, err := a.store.ListClinicalStaffUserIDs(ctx, pet.PracticeID)
		if err != nil || len(ids) == 0 {
			return
		}
		clientName := visit.ClientName
		if clientName == "" {
			if client, cerr := a.store.GetUserByID(ctx, pet.OwnerUserID); cerr == nil {
				clientName = client.FullName
				if clientName == "" {
					clientName = client.Email
				}
			}
		}
		payload := map[string]any{
			"visitId":    visit.ID,
			"petId":      pet.ID,
			"petName":    pet.Name,
			"clientName": clientName,
			"clientId":   pet.OwnerUserID,
		}
		actorUserID = strings.TrimSpace(actorUserID)
		for _, uid := range ids {
			if actorUserID != "" && uid == actorUserID {
				continue
			}
			_ = a.store.LogNotification(ctx, uid, "waiting_room", payload)
		}
	}()
}

func (a *API) listDeskAlerts(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	alerts, err := a.store.ListUnreadDeskAlerts(r.Context(), id.UserID, 30)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, alerts)
}

type markDeskAlertsReq struct {
	IDs []string `json:"ids"`
}

func (a *API) markDeskAlertsRead(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req markDeskAlertsReq
	_ = httpx.DecodeJSON(r, &req)
	n, err := a.store.MarkDeskAlertsRead(r.Context(), id.UserID, req.IDs)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"marked": n})
}
