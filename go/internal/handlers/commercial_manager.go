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

func (a *API) registerCommercialManagerRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/commercial-manager/overview", a.managerOverview)
		pr.Get("/commercial-manager/team", a.managerListTeam)
		pr.Get("/commercial-manager/team/{id}/overview", a.managerTeamMemberOverview)
		pr.Get("/commercial-manager/prospects", a.managerListProspects)
		pr.Patch("/commercial-manager/prospects/{id}", a.managerUpdateProspect)
		pr.Get("/commercial-manager/followups", a.managerFollowups)
	})
}

func (a *API) requireCommercialManager(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleCommercialManager {
		writeErr(w, r, http.StatusForbidden, "forbidden", "commercial_manager_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) requireCommercialOrManager(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || !kernel.IsSalesForce(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "commercial_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) managerOverview(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	ov, err := a.store.ManagerOverview(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, ov)
}

func (a *API) managerListTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	team, err := a.store.ListManagerTeam(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, team)
}

func (a *API) managerTeamMemberOverview(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	memberID := chi.URLParam(r, "id")
	belongs, err := a.store.CommercialBelongsToManager(r.Context(), memberID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !belongs {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	ov, err := a.store.CommercialOverview(r.Context(), memberID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, ov)
}

func (a *API) managerListProspects(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !store.ValidProspectStatus(status) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	commercialID := strings.TrimSpace(r.URL.Query().Get("commercialUserId"))
	if commercialID != "" {
		belongs, err := a.store.CommercialBelongsToManager(r.Context(), commercialID, id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !belongs {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
	}
	upcoming := r.URL.Query().Get("upcomingAppointments") == "1" || r.URL.Query().Get("upcomingAppointments") == "true"
	rows, err := a.store.ListManagerProspects(r.Context(), id.UserID, status, commercialID, upcoming)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) managerFollowups(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	data, err := a.store.ListManagerFollowups(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, data)
}

func (a *API) managerUpdateProspect(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	existing, err := a.store.GetProspectByID(r.Context(), prospectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if existing.Source != "directory" {
		belongs, err := a.store.CommercialBelongsToManager(r.Context(), existing.CommercialUserID, id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !belongs {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
	}
	in, ok := a.parseProspectUpdate(w, r, existing)
	if !ok {
		return
	}
	prospect, err := a.store.UpdateProspectAsManager(r.Context(), id.UserID, prospectID, in)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, prospect)
}

type prospectUpdateReq struct {
	PracticeName       string  `json:"practiceName"`
	ContactName        string  `json:"contactName"`
	ContactEmail       string  `json:"contactEmail"`
	ContactPhone       string  `json:"contactPhone"`
	City               string  `json:"city"`
	Notes              string  `json:"notes"`
	Status             string  `json:"status"`
	AppointmentAt      *string `json:"appointmentAt"`
	ClearAppointment   bool    `json:"clearAppointment"`
	AppointmentOutcome string  `json:"appointmentOutcome"`
	LostReason         string  `json:"lostReason"`
}

func (a *API) parseProspectUpdate(w http.ResponseWriter, r *http.Request, existing store.Prospect) (store.ProspectInput, bool) {
	var req prospectUpdateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return store.ProspectInput{}, false
	}
	in := store.ProspectInput{
		PracticeName:       existing.PracticeName,
		ContactName:        existing.ContactName,
		ContactEmail:       existing.ContactEmail,
		ContactPhone:       existing.ContactPhone,
		City:               existing.City,
		Notes:              existing.Notes,
		Status:             existing.Status,
		AppointmentAt:      existing.AppointmentAt,
		AppointmentOutcome: existing.AppointmentOutcome,
		LostReason:         existing.LostReason,
	}
	if req.PracticeName != "" {
		in.PracticeName = req.PracticeName
	}
	if req.ContactName != "" {
		in.ContactName = req.ContactName
	}
	if req.ContactEmail != "" {
		in.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		in.ContactPhone = req.ContactPhone
	}
	if req.City != "" {
		in.City = req.City
	}
	if req.Notes != "" {
		in.Notes = req.Notes
	}
	if req.Status != "" {
		if !store.ValidProspectStatus(req.Status) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return store.ProspectInput{}, false
		}
		in.Status = req.Status
	}
	if req.AppointmentOutcome != "" {
		if !store.ValidAppointmentOutcome(req.AppointmentOutcome) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_appointment_outcome")
			return store.ProspectInput{}, false
		}
		in.AppointmentOutcome = req.AppointmentOutcome
	}
	if req.LostReason != "" {
		in.LostReason = req.LostReason
	}
	in.ClearAppointment = req.ClearAppointment
	if req.AppointmentAt != nil {
		raw := strings.TrimSpace(*req.AppointmentAt)
		if raw == "" {
			in.ClearAppointment = true
			in.AppointmentAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				t, err = time.Parse("2006-01-02T15:04", raw)
			}
			if err != nil {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_appointment_at")
				return store.ProspectInput{}, false
			}
			in.AppointmentAt = &t
			if in.AppointmentOutcome == "" {
				in.AppointmentOutcome = "scheduled"
			}
		}
	}
	return in, true
}
