package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerAdminRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/admin/metrics/overview", a.adminMetricsOverview)
		pr.Get("/admin/users", a.adminListUsers)
		pr.Get("/admin/payments", a.adminListPayments)
		pr.Get("/admin/commercials", a.adminListCommercials)
		pr.Post("/admin/commercials", a.adminCreateCommercial)
		pr.Patch("/admin/commercials/{id}/assign", a.adminAssignVet)
		pr.Get("/admin/vets", a.adminListVets)
		pr.Post("/admin/vets", a.adminCreateVet)
		pr.Post("/admin/clients", a.adminCreateClient)
		pr.Get("/admin/commercials/{id}/commissions", a.adminCommercialCommissions)
		pr.Get("/admin/prospects", a.adminListProspects)
	})
}

func (a *API) adminListCommercials(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListAllCommercials(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

type createCommercialReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

func (a *API) adminCreateCommercial(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req createCommercialReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return
	}
	if _, err := a.store.GetUserByEmail(r.Context(), req.Email); err == nil {
		writeErr(w, r, http.StatusConflict, "conflict", "email_already_exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	userID, err := a.store.CreateCommercialUser(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": userID, "email": req.Email})
}

type assignVetReq struct {
	VetUserID string `json:"vetUserId"`
}

func (a *API) adminAssignVet(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	commercialID := chi.URLParam(r, "id")
	var req assignVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.VetUserID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if err := a.store.AssignVetToCommercial(r.Context(), req.VetUserID, commercialID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "assigned", "vetUserId": req.VetUserID, "commercialId": commercialID})
}

func (a *API) adminCommercialCommissions(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	summary, err := a.store.GetCommercialCommissionSummary(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, summary)
}

func (a *API) adminListProspects(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !store.ValidProspectStatus(status) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	rows, err := a.store.ListAllProspects(r.Context(), status)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleAdmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "admin_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) adminMetricsOverview(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	m, err := a.store.AdminMetricsOverview(r.Context(), from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, m)
}

func (a *API) adminListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit
	rows, err := a.store.ListAdminUsers(r.Context(), r.URL.Query().Get("role"), from, to, limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) adminListPayments(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit
	rows, err := a.store.ListAdminPayments(r.Context(), from, to, r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}
