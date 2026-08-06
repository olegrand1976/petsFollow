package handlers

import (
	"encoding/json"
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

func (a *API) registerCommercialCRMRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/commercial/prospects/{id}", a.commercialGetProspectDetail)
		pr.Get("/commercial/prospects/{id}/events", a.commercialListProspectEvents)
		pr.Post("/commercial/prospects/{id}/events", a.commercialCreateProspectEvent)
		pr.Get("/commercial/prospects/{id}/activities", a.commercialListProspectActivities)
		pr.Post("/commercial/prospects/{id}/activities", a.commercialCreateProspectActivity)
		pr.Patch("/commercial/activities/{id}", a.commercialPatchActivity)
		pr.Get("/commercial/activities", a.commercialListMyActivities)
		pr.Get("/commercial/agenda", a.commercialAgenda)
	})
}

func (a *API) registerCommercialManagerCRMRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/commercial-manager/activities", a.managerListActivities)
		pr.Post("/commercial-manager/activities", a.managerCreateActivity)
		pr.Get("/commercial-manager/agenda", a.managerAgenda)
	})
}

func (a *API) commercialGetProspectDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	if _, ok := a.loadProspectForMail(w, r, id, prospectID); !ok {
		return
	}
	detail, err := a.store.GetProspectDetail(r.Context(), prospectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) commercialListProspectEvents(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	if _, ok := a.loadProspectForMail(w, r, id, prospectID); !ok {
		return
	}
	items, err := a.store.ListProspectEvents(r.Context(), prospectID, 100)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type prospectEventReq struct {
	Kind string `json:"kind"`
	Body string `json:"body"`
}

func (a *API) commercialCreateProspectEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	p, ok := a.loadProspectForMail(w, r, id, prospectID)
	if !ok {
		return
	}
	if p.CommercialUserID == "" {
		writeErr(w, r, http.StatusConflict, "conflict", "prospect_unclaimed")
		return
	}
	var req prospectEventReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = string(store.EventNote)
	}
	if kind != string(store.EventNote) && kind != string(store.EventCall) && kind != string(store.EventMeeting) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_event_kind")
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	ev, err := a.store.CreateProspectEvent(r.Context(), prospectID, id.UserID, kind, req.Body, nil)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	_ = a.store.TouchProspectContacted(r.Context(), prospectID, id.UserID)
	httpx.WriteData(w, http.StatusCreated, ev)
}

func (a *API) commercialListProspectActivities(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	if _, ok := a.loadProspectForMail(w, r, id, prospectID); !ok {
		return
	}
	openOnly := r.URL.Query().Get("open") == "1" || r.URL.Query().Get("open") == "true"
	items, err := a.store.ListActivitiesByProspect(r.Context(), prospectID, openOnly)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type activityCreateReq struct {
	Kind           string  `json:"kind"`
	Title          string  `json:"title"`
	DueAt          *string `json:"dueAt"`
	AssigneeUserID string  `json:"assigneeUserId"`
	ProspectID     string  `json:"prospectId"`
}

func (a *API) commercialCreateProspectActivity(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	p, ok := a.loadProspectForMail(w, r, id, prospectID)
	if !ok {
		return
	}
	if p.CommercialUserID == "" {
		writeErr(w, r, http.StatusConflict, "conflict", "prospect_unclaimed")
		return
	}
	var req activityCreateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	assignee := p.CommercialUserID
	if id.Role == kernel.RoleCommercial {
		assignee = id.UserID
	}
	act, err := a.createActivityForProspect(r, id, p, assignee, req)
	if err != nil {
		writeActivityErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, act)
}

func (a *API) createActivityForProspect(r *http.Request, id authx.Identity, p store.Prospect, assignee string, req activityCreateReq) (store.Activity, error) {
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = string(store.ActivityFollowUp)
	}
	if !store.ValidActivityKind(kind) {
		return store.Activity{}, errBadActivityKind
	}
	var due *time.Time
	if req.DueAt != nil && strings.TrimSpace(*req.DueAt) != "" {
		t, err := parseFlexibleTime(*req.DueAt)
		if err != nil {
			return store.Activity{}, errBadDueAt
		}
		due = &t
	}
	act, err := a.store.CreateActivity(r.Context(), store.ActivityInput{
		ProspectID:     p.ID,
		AssigneeUserID: assignee,
		CreatedBy:      id.UserID,
		Kind:           kind,
		Title:          req.Title,
		DueAt:          due,
	})
	if err != nil {
		return store.Activity{}, err
	}
	meta := map[string]any{"activityId": act.ID, "kind": act.Kind}
	_, _ = a.store.CreateProspectEvent(r.Context(), p.ID, id.UserID, string(store.EventSystem),
		"Tâche créée : "+act.Title, meta)
	return act, nil
}

var (
	errBadActivityKind = errors.New("invalid_activity_kind")
	errBadDueAt        = errors.New("invalid_due_at")
	errBadStatus       = errors.New("invalid_activity_status")
)

func writeActivityErr(w http.ResponseWriter, r *http.Request, err error) {
	msg := err.Error()
	switch {
	case errors.Is(err, errBadActivityKind), msg == "invalid_activity_kind":
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_activity_kind")
	case errors.Is(err, errBadDueAt), msg == "invalid_due_at":
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_due_at")
	case errors.Is(err, errBadStatus), msg == "invalid_activity_status":
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_activity_status")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}

func parseFlexibleTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04", raw)
}

type activityPatchReq struct {
	Title    string  `json:"title"`
	Status   string  `json:"status"`
	DueAt    *string `json:"dueAt"`
	ClearDue bool    `json:"clearDue"`
}

func (a *API) commercialPatchActivity(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	actID := chi.URLParam(r, "id")
	existing, err := a.store.GetActivity(r.Context(), actID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !a.canAccessActivity(r, id, existing) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	var req activityPatchReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var due *time.Time
	if req.DueAt != nil && strings.TrimSpace(*req.DueAt) != "" {
		t, err := parseFlexibleTime(*req.DueAt)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_due_at")
			return
		}
		due = &t
	}
	prevStatus := existing.Status
	act, err := a.store.UpdateActivity(r.Context(), actID, req.Title, req.Status, due, req.ClearDue)
	if err != nil {
		writeActivityErr(w, r, err)
		return
	}
	if prevStatus != "done" && act.Status == "done" {
		_, _ = a.store.CreateProspectEvent(r.Context(), act.ProspectID, id.UserID, string(store.EventTaskDone),
			"Tâche terminée : "+act.Title, map[string]any{"activityId": act.ID})
		_ = a.store.TouchProspectContacted(r.Context(), act.ProspectID, id.UserID)
	}
	httpx.WriteData(w, http.StatusOK, act)
}

func (a *API) canAccessActivity(r *http.Request, id authx.Identity, act store.Activity) bool {
	if act.AssigneeUserID == id.UserID || act.CreatedBy == id.UserID {
		return true
	}
	if id.Role == kernel.RoleCommercialManager {
		ok, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, act.AssigneeUserID)
		return err == nil && ok
	}
	return false
}

func (a *API) commercialListMyActivities(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	overdue := r.URL.Query().Get("overdue") == "1" || r.URL.Query().Get("overdue") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.store.ListActivitiesForAssignee(r.Context(), id.UserID, status, overdue, limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) commercialAgenda(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	from, to, ok := parseAgendaRange(w, r)
	if !ok {
		return
	}
	items, err := a.store.ListAgenda(r.Context(), id.UserID, from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func parseAgendaRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	fromRaw := strings.TrimSpace(r.URL.Query().Get("from"))
	toRaw := strings.TrimSpace(r.URL.Query().Get("to"))
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// start of week Monday
	weekday := int(from.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	from = from.AddDate(0, 0, -(weekday - 1))
	to := from.AddDate(0, 0, 7)
	if fromRaw != "" {
		t, err := time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			t, err = time.Parse("2006-01-02", fromRaw)
		}
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_from")
			return time.Time{}, time.Time{}, false
		}
		from = t
	}
	if toRaw != "" {
		t, err := time.Parse(time.RFC3339, toRaw)
		if err != nil {
			t, err = time.Parse("2006-01-02", toRaw)
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
	return from, to, true
}

func (a *API) managerListActivities(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	commercialID := strings.TrimSpace(r.URL.Query().Get("commercialUserId"))
	if commercialID != "" {
		okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, commercialID)
		if err != nil || !okTeam {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_team_member")
			return
		}
	}
	status := r.URL.Query().Get("status")
	overdue := r.URL.Query().Get("overdue") == "1" || r.URL.Query().Get("overdue") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.store.ListTeamActivities(r.Context(), id.UserID, commercialID, status, overdue, limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) managerCreateActivity(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	var req activityCreateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.ProspectID) == "" || strings.TrimSpace(req.Title) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	p, ok := a.loadProspectForMail(w, r, id, req.ProspectID)
	if !ok {
		return
	}
	if p.CommercialUserID == "" {
		writeErr(w, r, http.StatusConflict, "conflict", "prospect_unclaimed")
		return
	}
	assignee := strings.TrimSpace(req.AssigneeUserID)
	if assignee == "" {
		assignee = p.CommercialUserID
	}
	if assignee != id.UserID {
		okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, assignee)
		if err != nil || !okTeam {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_team_member")
			return
		}
	}
	act, err := a.createActivityForProspect(r, id, p, assignee, req)
	if err != nil {
		writeActivityErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, act)
}

func (a *API) managerAgenda(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	commercialID := strings.TrimSpace(r.URL.Query().Get("commercialUserId"))
	if commercialID != "" {
		okTeam, err := a.store.IsManagerOfCommercial(r.Context(), id.UserID, commercialID)
		if err != nil || !okTeam {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_team_member")
			return
		}
	}
	from, to, ok := parseAgendaRange(w, r)
	if !ok {
		return
	}
	items, err := a.store.ListTeamAgenda(r.Context(), id.UserID, commercialID, from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

// recordProspectMutationEvents writes timeline entries after a successful prospect PATCH.
func (a *API) recordProspectMutationEvents(r *http.Request, actorUserID string, before, after store.Prospect) {
	if before.Status != after.Status {
		meta, _ := json.Marshal(map[string]string{"from": before.Status, "to": after.Status})
		var m map[string]any
		_ = json.Unmarshal(meta, &m)
		_, _ = a.store.CreateProspectEvent(r.Context(), after.ID, actorUserID, string(store.EventStatusChange),
			"Statut : "+before.Status+" → "+after.Status, m)
	}
	beforeAppt := ""
	afterAppt := ""
	if before.AppointmentAt != nil {
		beforeAppt = before.AppointmentAt.UTC().Format(time.RFC3339)
	}
	if after.AppointmentAt != nil {
		afterAppt = after.AppointmentAt.UTC().Format(time.RFC3339)
	}
	if beforeAppt != afterAppt || before.AppointmentOutcome != after.AppointmentOutcome {
		body := "RDV mis à jour"
		if afterAppt != "" {
			body = "RDV : " + after.AppointmentAt.Format("02/01/2006 15:04")
		}
		_, _ = a.store.CreateProspectEvent(r.Context(), after.ID, actorUserID, string(store.EventMeeting), body, map[string]any{
			"appointmentAt":      afterAppt,
			"appointmentOutcome": after.AppointmentOutcome,
		})
	}
}
