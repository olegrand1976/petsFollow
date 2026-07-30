package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func (a *API) registerPharmacyProtocolRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/protocols", a.listPharmacyProtocols)
}

func (a *API) listPharmacyProtocols(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	items, err := a.store.ListClinicalProtocols(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}
