package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type createWeightReq struct {
	WeightKg float64 `json:"weightKg"`
	Comment  *string `json:"comment"`
}

func (a *API) createWeightReading(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	pet, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil || pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	if !a.requirePetPractice(w, r, pet) {
		return
	}
	var req createWeightReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	reading, err := a.store.CreateWeightReading(
		r.Context(), pet.ID, id.UserID, id.UserID, pet.PracticeID, req.WeightKg, req.Comment,
	)
	if errors.Is(err, store.ErrInvalidInput) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_weight")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, reading)
}

func (a *API) listWeightReadings(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	readings, err := a.store.ListWeightReadings(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, readings)
}
