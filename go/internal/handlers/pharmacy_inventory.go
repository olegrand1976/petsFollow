package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyInventoryRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/inventory/sessions", a.listPharmacyInventorySessions)
	pr.Post("/vet/pharmacy/inventory/sessions", a.startPharmacyInventorySession)
	pr.Get("/vet/pharmacy/inventory/sessions/{id}", a.getPharmacyInventorySession)
	pr.Get("/vet/pharmacy/inventory/sessions/{id}/export.csv", a.exportPharmacyInventoryCSV)
	pr.Patch("/vet/pharmacy/inventory/sessions/{id}/lines/{lineId}", a.patchPharmacyInventoryLine)
	pr.Post("/vet/pharmacy/inventory/sessions/{id}/close", a.closePharmacyInventorySession)
	pr.Post("/vet/pharmacy/inventory/sessions/{id}/cancel", a.cancelPharmacyInventorySession)
}

func (a *API) listPharmacyInventorySessions(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	rows, err := a.store.ListInventorySessions(r.Context(), id.PracticeID, 20)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) startPharmacyInventorySession(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		DepositID string `json:"depositId"`
		Notes     string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	sess, err := a.store.StartInventorySession(r.Context(), id.PracticeID, id.UserID, body.DepositID, body.Notes)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "inventory_open_exists", "inventory_open_exists")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, sess)
}

func (a *API) getPharmacyInventorySession(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	sess, err := a.store.GetInventorySession(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) patchPharmacyInventoryLine(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		CountedQty float64 `json:"countedQty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	line, err := a.store.SetInventoryCountedQty(r.Context(), id.PracticeID, chi.URLParam(r, "id"), chi.URLParam(r, "lineId"), body.CountedQty)
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
	httpx.WriteData(w, http.StatusOK, line)
}

func (a *API) closePharmacyInventorySession(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		Counts []store.InventoryCountInput `json:"counts"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	settings, err := a.store.GetPharmacySettings(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	sess, err := a.store.CloseInventorySession(r.Context(), id.PracticeID, chi.URLParam(r, "id"), id.UserID, body.Counts, settings, time.Now())
	if err != nil {
		if errors.Is(err, store.ErrInventoryIncomplete) {
			writeErr(w, r, http.StatusBadRequest, "inventory_incomplete", "inventory_incomplete")
			return
		}
		if a.writePharmacyErr(w, r, err) {
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "inventory_not_open", "inventory_not_open")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) cancelPharmacyInventorySession(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	if err := a.store.CancelInventorySession(r.Context(), id.PracticeID, chi.URLParam(r, "id"), id.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (a *API) exportPharmacyInventoryCSV(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	sess, err := a.store.GetInventorySession(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	csv := store.FormatInventoryCSV(sess)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="inventory-`+sess.ID[:8]+`.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csv))
}
