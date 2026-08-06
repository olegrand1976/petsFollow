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
	SiteID                string               `json:"siteId"`
	ClientBookingEnabled  bool                 `json:"clientBookingEnabled"`
	SlotDurationMinutes   int                  `json:"slotDurationMinutes"`
	VacationsDeclaredYear *int                 `json:"vacationsDeclaredYear"`
	Slots                 []store.ScheduleSlot `json:"slots"`
}

func (a *API) getVetSchedule(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	sched, err := a.store.GetVetSchedule(r.Context(), id.PracticeID, siteIDQuery(r))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sched)
}

func (a *API) putVetSchedule(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
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
	siteID := req.SiteID
	if siteID == "" {
		siteID = siteIDQuery(r)
	}
	sched, err := a.store.PutVetSchedule(r.Context(), id.PracticeID, siteID, req.ClientBookingEnabled, req.SlotDurationMinutes, req.VacationsDeclaredYear, req.Slots)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			msg := err.Error()
			code := "invalid_schedule"
			if strings.Contains(msg, "schedule_incomplete") {
				code = "schedule_incomplete"
			} else if strings.Contains(msg, "site_inactive") || strings.Contains(msg, "invalid_site") {
				code = "invalid_site"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sched)
}

func (a *API) listVetVacations(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	items, err := a.store.ListVacations(r.Context(), id.PracticeID, siteIDQuery(r))
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type createVacationReq struct {
	SiteID   string `json:"siteId"`
	StartsOn string `json:"startsOn"`
	EndsOn   string `json:"endsOn"`
	Label    string `json:"label"`
}

func (a *API) createVetVacation(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req createVacationReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.StartsOn == "" || req.EndsOn == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	siteID := req.SiteID
	if siteID == "" {
		siteID = siteIDQuery(r)
	}
	v, err := a.store.CreateVacation(r.Context(), id.PracticeID, siteID, req.StartsOn, req.EndsOn, req.Label)
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			code := "invalid_vacation"
			if strings.Contains(err.Error(), "site_inactive") || strings.Contains(err.Error(), "invalid_site") {
				code = "invalid_site"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_vacation")
		return
	}
	httpx.WriteData(w, http.StatusCreated, v)
}

func (a *API) deleteVetVacation(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
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

func (a *API) listVetVisitTypes(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	activeOnly := r.URL.Query().Get("active") == "1" || r.URL.Query().Get("active") == "true"
	items, err := a.store.ListVisitTypes(r.Context(), id.PracticeID, activeOnly)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type putVisitTypesReq struct {
	Items []store.VisitTypeInput `json:"items"`
}

func (a *API) putVetVisitTypes(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req putVisitTypesReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	items, err := a.store.PutVisitTypes(r.Context(), id.PracticeID, req.Items)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_visit_type"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "duplicate_name"):
				code = "duplicate_name"
			case strings.Contains(msg, "invalid_duration"):
				code = "invalid_duration"
			case strings.Contains(msg, "invalid_color"):
				code = "invalid_color"
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) getVetCalendar(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	from, to, ok := parseFromTo(w, r)
	if !ok {
		return
	}
	visits, err := a.store.ListPracticeVisitsInRange(r.Context(), id.PracticeID, siteIDQuery(r), from, to)
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := a.store.AttachPreconsultStatuses(r.Context(), visits); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	vacations, err := a.store.ListVacations(r.Context(), id.PracticeID, siteIDQuery(r))
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pending, err := a.store.ListPracticePendingVetActions(r.Context(), id.PracticeID, siteIDQuery(r))
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := a.store.AttachPreconsultStatuses(r.Context(), pending); err != nil {
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
	contact, err := a.store.GetPracticeContact(r.Context(), practiceID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	base := map[string]any{
		"practicePhone": contact.Phone,
		"practiceName":  contact.PracticeName,
	}
	siteID := siteIDQuery(r)

	// Without siteId: expose sites; fall through to slots only when exactly one bookable site.
	if siteID == "" {
		sites, serr := a.store.ListSites(r.Context(), practiceID, false)
		if serr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		type siteAvail struct {
			SiteID   string `json:"siteId"`
			SiteName string `json:"siteName"`
			Enabled  bool   `json:"enabled"`
			Phone    string `json:"phone,omitempty"`
		}
		var siteRows []siteAvail
		var bookableIDs []string
		anyEnabled := false
		for _, s := range sites {
			en, _, e := a.store.ClientBookingEnabled(r.Context(), practiceID, s.ID)
			if e != nil {
				continue
			}
			phone := s.Phone
			if phone == "" {
				phone = contact.Phone
			}
			siteRows = append(siteRows, siteAvail{SiteID: s.ID, SiteName: s.Name, Enabled: en, Phone: phone})
			if en {
				anyEnabled = true
				bookableIDs = append(bookableIDs, s.ID)
			}
		}
		if siteRows == nil {
			siteRows = []siteAvail{}
		}
		base["sites"] = siteRows
		switch len(bookableIDs) {
		case 0:
			base["enabled"] = false
			base["slots"] = []any{}
			httpx.WriteData(w, http.StatusOK, base)
			return
		case 1:
			// Compat mono-bookable: fill slots for that site even if other inactive/non-bookable sites exist.
			siteID = bookableIDs[0]
		default:
			base["enabled"] = anyEnabled
			base["slots"] = []any{}
			httpx.WriteData(w, http.StatusOK, base)
			return
		}
	}

	enabled, _, err := a.store.ClientBookingEnabled(r.Context(), practiceID, siteID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if site, serr := a.store.GetSite(r.Context(), practiceID, siteID); serr == nil {
		base["siteId"] = site.ID
		base["siteName"] = site.Name
		if site.Phone != "" {
			base["practicePhone"] = site.Phone
		}
	}
	if !enabled {
		base["enabled"] = false
		base["slots"] = []any{}
		httpx.WriteData(w, http.StatusOK, base)
		return
	}
	from, to, ok := parseFromTo(w, r)
	if !ok {
		return
	}
	slots, err := a.store.ListAvailableSlots(r.Context(), practiceID, siteID, from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	base["enabled"] = true
	base["slots"] = slots
	httpx.WriteData(w, http.StatusOK, base)
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
