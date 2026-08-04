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

func (a *API) registerSpeciesRoutes(pr chi.Router) {
	pr.Get("/species", a.listSpecies)
}

func (a *API) registerAdminSpeciesRoutes(pr chi.Router) {
	pr.Get("/admin/species", a.adminListSpecies)
	pr.Post("/admin/species", a.adminCreateSpecies)
	pr.Patch("/admin/species/{code}", a.adminPatchSpecies)
}

// speciesCountryFor resolves which country's rules apply to the caller: the practice
// country for practice staff, BE otherwise. Admins may override with ?country=.
func (a *API) speciesCountryFor(r *http.Request, id authx.Identity) string {
	if q := strings.TrimSpace(r.URL.Query().Get("country")); q != "" && kernel.IsOpsRole(id.Role) {
		return store.NormalizeCountryCode(q)
	}
	if strings.TrimSpace(id.PracticeID) == "" {
		return "BE"
	}
	country, err := a.store.GetPracticeCountryCode(r.Context(), id.PracticeID)
	if err != nil {
		return "BE"
	}
	return country
}

// listSpecies serves the catalogue to Nuxt Pro, Flutter and Pro Light.
// Only active species are returned — the admin surface uses /admin/species.
func (a *API) listSpecies(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	country := a.speciesCountryFor(r, id)
	rows, err := a.store.ListActiveSpecies(r.Context(), country)
	if err != nil {
		writeInternal(w, r)
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"countryCode": country,
		"species":     rows,
	})
}

func (a *API) adminListSpecies(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	country := a.speciesCountryFor(r, id)
	rows, err := a.store.ListSpecies(r.Context(), country)
	if err != nil {
		writeInternal(w, r)
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"countryCode": country,
		"species":     rows,
	})
}

type speciesRuleBody struct {
	CountryCode            string `json:"countryCode"`
	IsLargeAnimal          bool   `json:"isLargeAnimal"`
	DAFRequired            string `json:"dafRequired"`
	IsFoodChain            bool   `json:"isFoodChain"`
	DefaultFoodChainStatus string `json:"defaultFoodChainStatus"`
}

func (a *API) adminCreateSpecies(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		Code              string            `json:"code"`
		Labels            map[string]string `json:"labels"`
		SortOrder         int               `json:"sortOrder"`
		IsActive          *bool             `json:"isActive"`
		SupportsHeartRate bool              `json:"supportsHeartRate"`
		Rule              *speciesRuleBody  `json:"rule"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	sp, err := a.store.CreateSpecies(r.Context(), store.Species{
		Code:              body.Code,
		Labels:            body.Labels,
		SortOrder:         body.SortOrder,
		IsActive:          active,
		SupportsHeartRate: body.SupportsHeartRate,
	})
	if a.writeSpeciesErr(w, r, err) {
		return
	}
	if body.Rule != nil {
		if err := a.store.UpsertSpeciesRule(r.Context(), sp.Code, store.SpeciesRule{
			CountryCode:            body.Rule.CountryCode,
			IsLargeAnimal:          body.Rule.IsLargeAnimal,
			DAFRequired:            body.Rule.DAFRequired,
			IsFoodChain:            body.Rule.IsFoodChain,
			DefaultFoodChainStatus: body.Rule.DefaultFoodChainStatus,
		}); a.writeSpeciesErr(w, r, err) {
			return
		}
	}
	httpx.WriteData(w, http.StatusCreated, sp)
}

func (a *API) adminPatchSpecies(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	code := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	var body struct {
		Labels            map[string]string `json:"labels"`
		SortOrder         *int              `json:"sortOrder"`
		IsActive          *bool             `json:"isActive"`
		SupportsHeartRate *bool             `json:"supportsHeartRate"`
		Rule              *speciesRuleBody  `json:"rule"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	sp, err := a.store.UpdateSpecies(r.Context(), code, store.SpeciesPatch{
		Labels:            body.Labels,
		SortOrder:         body.SortOrder,
		IsActive:          body.IsActive,
		SupportsHeartRate: body.SupportsHeartRate,
	})
	if a.writeSpeciesErr(w, r, err) {
		return
	}
	if body.Rule != nil {
		if err := a.store.UpsertSpeciesRule(r.Context(), code, store.SpeciesRule{
			CountryCode:            body.Rule.CountryCode,
			IsLargeAnimal:          body.Rule.IsLargeAnimal,
			DAFRequired:            body.Rule.DAFRequired,
			IsFoodChain:            body.Rule.IsFoodChain,
			DefaultFoodChainStatus: body.Rule.DefaultFoodChainStatus,
		}); a.writeSpeciesErr(w, r, err) {
			return
		}
	}
	httpx.WriteData(w, http.StatusOK, sp)
}

// writeSpeciesErr maps store errors to HTTP; returns true when a response was written.
func (a *API) writeSpeciesErr(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, store.ErrConflict):
		writeErr(w, r, http.StatusConflict, "species_exists", "species_exists")
	default:
		writeInternal(w, r)
	}
	return true
}
