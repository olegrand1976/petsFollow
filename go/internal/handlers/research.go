package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerResearchRoutes(r chi.Router) {
	r.Get("/research/overview", a.researchOverview)
	r.Get("/research/heatmap", a.researchHeatmap)
	r.Get("/research/timeseries", a.researchTimeseries)
	r.Get("/research/alerts", a.researchAlerts)
	r.Get("/vet/practice/research-opt-in", a.getResearchOptIn)
	r.Post("/vet/practice/research-opt-in", a.postResearchOptIn)
	r.Delete("/vet/practice/research-opt-in", a.deleteResearchOptIn)
	r.Get("/admin/research/opt-ins", a.adminListResearchOptIns)
	a.registerResearchV2Routes(r)
}

func (a *API) requireResearchEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.ResearchEnabled {
		writeErr(w, r, http.StatusNotFound, "research_disabled", "research_disabled")
		return false
	}
	return true
}

func (a *API) requireResearchRole(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	if !a.requireResearchEnabled(w, r) {
		return authx.Identity{}, false
	}
	id, err := authx.FromContext(r.Context())
	if err != nil || !kernel.IsResearchRole(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return id, false
	}
	return id, true
}

func (a *API) researchOverview(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireResearchRole(w, r); !ok {
		return
	}
	o, err := a.store.ResearchOverview(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, o)
}

func (a *API) researchHeatmap(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireResearchRole(w, r); !ok {
		return
	}
	q := r.URL.Query()
	cells, err := a.store.ResearchHeatmap(r.Context(),
		strings.TrimSpace(q.Get("species")),
		strings.TrimSpace(q.Get("signal")),
		strings.TrimSpace(q.Get("from")),
		strings.TrimSpace(q.Get("to")),
	)
	if err != nil {
		if strings.Contains(err.Error(), "invalid_") {
			writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": cells})
}

func (a *API) researchTimeseries(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireResearchRole(w, r); !ok {
		return
	}
	q := r.URL.Query()
	pts, err := a.store.ResearchTimeseries(r.Context(),
		strings.TrimSpace(q.Get("species")),
		strings.TrimSpace(q.Get("signal")),
		strings.TrimSpace(q.Get("country")),
		strings.TrimSpace(q.Get("postalCode")),
	)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": pts})
}

func (a *API) researchAlerts(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireResearchRole(w, r); !ok {
		return
	}
	alerts, err := a.store.ResearchAlerts(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": alerts})
}

func (a *API) getResearchOptIn(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	st, err := a.store.GetResearchOptIn(r.Context(), id.PracticeID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, st)
}

func (a *API) postResearchOptIn(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	st, err := a.store.SetResearchOptIn(r.Context(), id.PracticeID, id.UserID, true)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, st)
}

func (a *API) deleteResearchOptIn(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	salt := store.ResearchAnonSalt(a.cfg)
	if salt == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "research_misconfigured", "research_anon_salt_required")
		return
	}
	st, err := a.store.OptOutResearchAndPurge(r.Context(), salt, id.PracticeID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, st)
}

func (a *API) adminListResearchOptIns(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	items, err := a.store.ListResearchOptInPractices(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) internalRunResearchETL(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Research-Etl-Secret", a.cfg.ResearchEtlSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.cfg.ResearchEnabled {
		writeErr(w, r, http.StatusNotFound, "research_disabled", "research_disabled")
		return
	}
	salt := store.ResearchAnonSalt(a.cfg)
	if salt == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "research_misconfigured", "research_anon_salt_required")
		return
	}
	res, err := a.store.RunResearchETL(r.Context(), salt)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, res)
}
