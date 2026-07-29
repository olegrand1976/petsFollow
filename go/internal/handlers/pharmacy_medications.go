package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyMedicationRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/medications/search", a.searchPharmacyMedications)
	pr.Patch("/vet/pharmacy/medications/{id}/withdrawal", a.patchPharmacyMedicationWithdrawal)
	pr.Patch("/vet/pets/{petID}/food-chain", a.patchVetPetFoodChain)
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

func (a *API) patchPharmacyMedicationWithdrawal(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		WithdrawalMeatDays *int `json:"withdrawalMeatDays"`
		WithdrawalMilkDays *int `json:"withdrawalMilkDays"`
		WithdrawalEggsDays *int `json:"withdrawalEggsDays"`
		FoodChainBanned    bool `json:"foodChainBanned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	med, err := a.store.UpsertPracticeMedicationWithdrawal(r.Context(), id.PracticeID, chi.URLParam(r, "id"),
		body.WithdrawalMeatDays, body.WithdrawalMilkDays, body.WithdrawalEggsDays, body.FoodChainBanned)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, med)
}

func (a *API) patchVetPetFoodChain(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		FoodChainStatus string `json:"foodChainStatus"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	petID := chi.URLParam(r, "petID")
	if err := a.store.SetPetFoodChainStatus(r.Context(), id.PracticeID, petID, body.FoodChainStatus); err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	st, _ := a.store.GetPetFoodChainStatus(r.Context(), petID)
	httpx.WriteData(w, http.StatusOK, map[string]string{"id": petID, "foodChainStatus": st})
}
