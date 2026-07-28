package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func (a *API) registerPharmacyMedicationRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/medications/search", a.searchPharmacyMedications)
}

// requirePharmacyEnabled gates pharmacy endpoints behind PHARMACY_ENABLED.
func (a *API) requirePharmacyEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
		return false
	}
	return true
}

// GET /vet/pharmacy/medications/search?q=&limit=
func (a *API) searchPharmacyMedications(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	if _, ok := a.requirePracticePerm(w, r, "pharmacy.read"); !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	items, err := a.store.SearchRefMedications(r.Context(), q, limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}
