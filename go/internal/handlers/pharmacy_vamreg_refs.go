package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyVamregRefRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/vamreg-refs", a.listPharmacyVamregRefs)
}

// GET /vet/pharmacy/vamreg-refs?kind=target_species|indication|pharmaceutical_form
func (a *API) listPharmacyVamregRefs(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	if _, ok := a.requirePracticePerm(w, r, "pharmacy.read"); !ok {
		return
	}
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	switch kind {
	case store.VamregRefKindTargetSpecies, store.VamregRefKindIndication, store.VamregRefKindPharmaceuticalForm:
	default:
		writeErr(w, r, http.StatusBadRequest, "validation_error", "kind_required")
		return
	}
	includeDeprecated := r.URL.Query().Get("includeDeprecated") == "1" || r.URL.Query().Get("includeDeprecated") == "true"
	items, err := a.store.ListVamregRefCodes(r.Context(), kind, includeDeprecated, 2000)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.VamregRefCode{}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"kind": kind, "items": items})
}

// POST /internal/pharmacy/vamreg-ref-sync — fetch AFMPS readonly lists; dryRun=true by default.
func (a *API) internalPharmacyVamregRefSync(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Pharmacy-Vamreg-Ref-Sync-Secret", a.cfg.PharmacyVamregRefSyncSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
		return
	}
	if a.vamregAFMPS == nil || strings.TrimSpace(a.cfg.VamregAfmpsAPIKey) == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "vamreg_afmps_not_configured")
		return
	}

	var body struct {
		DryRun *bool `json:"dryRun"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dryRun := true
	if body.DryRun != nil {
		dryRun = *body.DryRun
	}

	lists, err := pharmacy.FetchVamregRefLists(r.Context(), a.vamregAFMPS)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "vamreg_afmps_error", err.Error())
		return
	}

	result := store.VamregRefSyncResult{
		DryRun:   dryRun,
		Fetched:  map[string]int{},
		Upserted: map[string]int{},
		Sample:   map[string][]store.VamregRefCode{},
	}
	for kind, rows := range lists {
		result.Fetched[kind] = len(rows)
		sample := make([]store.VamregRefCode, 0, 3)
		for i, row := range rows {
			if i >= 3 {
				break
			}
			sample = append(sample, store.VamregRefCode{
				Kind: kind, Code: row.Code, LabelEN: row.En, LabelNL: row.Nl, LabelFR: row.Fr, Deprecated: row.Deprecated,
			})
		}
		result.Sample[kind] = sample
		if dryRun {
			continue
		}
		n, uerr := a.store.UpsertVamregRefCodes(r.Context(), kind, rows)
		if uerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		result.Upserted[kind] = n
	}
	if dryRun {
		result.Upserted = nil
	}
	httpx.WriteData(w, http.StatusOK, result)
}
