package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerAdminRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Get("/admin/metrics/overview", a.adminMetricsOverview)
		pr.Get("/admin/users", a.adminListUsers)
		pr.Get("/admin/payments", a.adminListPayments)
	})
}

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleAdmin {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "admin only")
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
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}
