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

type createBloodPressureReq struct {
	SystolicMmHg  int     `json:"systolicMmHg"`
	DiastolicMmHg int     `json:"diastolicMmHg"`
	MeanMmHg      *int    `json:"meanMmHg"`
	Method        string  `json:"method"`
	Site          *string `json:"site"`
	Comment       *string `json:"comment"`
}

func (a *API) createBloodPressureReading(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
	if !ok {
		return
	}
	if id.Role == kernel.RoleClient {
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
		if !a.requirePremiumAccess(w, r, pet.ID) {
			return
		}
	} else if kernel.IsPracticeStaff(id.Role) {
		if !a.allowPracticePerm(r, id, "pets.write_clinical") {
			writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
			return
		}
	} else if id.Role != kernel.RoleAdmin {
		// care_pro and other roles cannot write BP (clinical measure).
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	var req createBloodPressureReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	reading, err := a.store.CreateBloodPressureReading(
		r.Context(),
		pet.ID, pet.OwnerUserID, id.UserID, pet.PracticeID,
		req.SystolicMmHg, req.DiastolicMmHg, req.MeanMmHg,
		store.NormalizeBPMethod(req.Method), req.Site, req.Comment,
	)
	if errors.Is(err, store.ErrInvalidBloodPressure) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_blood_pressure")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, reading)
}

func (a *API) listBloodPressureReadings(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	readings, err := a.store.ListBloodPressureReadings(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, readings)
}
