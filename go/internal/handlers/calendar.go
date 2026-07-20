package handlers

import (
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

type putScheduleReq struct {
	ClientBookingEnabled  bool                 `json:"clientBookingEnabled"`
	SlotDurationMinutes   int                  `json:"slotDurationMinutes"`
	VacationsDeclaredYear *int                 `json:"vacationsDeclaredYear"`
	Slots                 []store.ScheduleSlot `json:"slots"`
}

func (a *API) getVetSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	sched, err := a.store.GetVetSchedule(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sched)
}

func (a *API) putVetSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req putScheduleReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.SlotDurationMinutes == 0 {
		req.SlotDurationMinutes = 30
	}
	if req.Slots == nil {
		req.Slots = []store.ScheduleSlot{}
	}
	sched, err := a.store.PutVetSchedule(r.Context(), id.PracticeID, req.ClientBookingEnabled, req.SlotDurationMinutes, req.VacationsDeclaredYear, req.Slots)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			msg := err.Error()
			code := "invalid_schedule"
			if strings.Contains(msg, "schedule_incomplete") {
				code = "schedule_incomplete"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sched)
}

func (a *API) listVetVacations(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	items, err := a.store.ListVacations(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type createVacationReq struct {
	StartsOn string `json:"startsOn"`
	EndsOn   string `json:"endsOn"`
	Label    string `json:"label"`
}

func (a *API) createVetVacation(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req createVacationReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.StartsOn == "" || req.EndsOn == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	v, err := a.store.CreateVacation(r.Context(), id.PracticeID, req.StartsOn, req.EndsOn, req.Label)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_vacation")
		return
	}
	httpx.WriteData(w, http.StatusCreated, v)
}

func (a *API) deleteVetVacation(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	if err := a.store.DeleteVacation(r.Context(), id.PracticeID, chi.URLParam(r, "id")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "vacation_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *API) getVetCalendar(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	from, to, ok := parseFromTo(w, r)
	if !ok {
		return
	}
	visits, err := a.store.ListPracticeVisitsInRange(r.Context(), id.PracticeID, from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	vacations, err := a.store.ListVacations(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pending, err := a.store.ListPracticePendingVetActions(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"visits":    visits,
		"vacations": vacations,
		"pending":   pending,
		"from":      from.UTC(),
		"to":        to.UTC(),
	})
}

func (a *API) getPracticeAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	practiceID := chi.URLParam(r, "practiceID")
	if practiceID == "" {
		practiceID = id.PracticeID
	}
	// Client must be linked to practice
	if _, err := a.store.GetClientByPractice(r.Context(), practiceID, id.UserID); err != nil {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_linked")
		return
	}
	enabled, _, err := a.store.ClientBookingEnabled(r.Context(), practiceID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !enabled {
		httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": false, "slots": []any{}})
		return
	}
	from, to, ok := parseFromTo(w, r)
	if !ok {
		return
	}
	slots, err := a.store.ListAvailableSlots(r.Context(), practiceID, from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"enabled": true, "slots": slots})
}

func parseFromTo(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 14)
	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			t, err = time.Parse("2006-01-02", fromStr)
		}
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_from")
			return time.Time{}, time.Time{}, false
		}
		from = t
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			t, err = time.Parse("2006-01-02", toStr)
		}
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_to")
			return time.Time{}, time.Time{}, false
		}
		to = t
	}
	if !to.After(from) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_range")
		return time.Time{}, time.Time{}, false
	}
	if to.Sub(from) > 62*24*time.Hour {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "range_too_large")
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}
