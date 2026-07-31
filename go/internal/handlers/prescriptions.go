package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/prescription"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPrescriptionRoutes(pr chi.Router) {
	pr.Get("/vet/prescriptions", a.listPrescriptions)
	pr.Post("/vet/prescriptions", a.createPrescription)
	pr.Get("/vet/prescriptions/{id}", a.getPrescription)
	pr.Patch("/vet/prescriptions/{id}", a.patchPrescription)
	pr.Delete("/vet/prescriptions/{id}", a.deletePrescription)
	pr.Get("/vet/prescriptions/{id}/pdf", a.getPrescriptionPDF)
}

func (a *API) requirePrescriptionsEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.PrescriptionsEnabled {
		writeErr(w, r, http.StatusNotFound, "prescriptions_disabled", "prescriptions_disabled")
		return false
	}
	return true
}

func (a *API) writePrescriptionErr(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case errors.Is(err, prescription.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, prescription.ErrNotDraft):
		writeErr(w, r, http.StatusConflict, "prescription_not_draft", "prescription_not_draft")
	case errors.Is(err, prescription.ErrEmptyMeds), errors.Is(err, prescription.ErrInvalidMeds):
		writeErr(w, r, http.StatusBadRequest, "invalid_medications", "invalid_medications")
	case errors.Is(err, prescription.ErrInvalidFormat):
		writeErr(w, r, http.StatusBadRequest, "invalid_format", "invalid_format")
	case errors.Is(err, prescription.ErrInvalidStatus):
		writeErr(w, r, http.StatusBadRequest, "invalid_status", "invalid_status")
	case errors.Is(err, prescription.ErrPayloadTooLarge):
		writeErr(w, r, http.StatusBadRequest, "payload_too_large", "payload_too_large")
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
	default:
		return false
	}
	return true
}

func (a *API) listPrescriptions(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != prescription.StatusDraft &&
		status != prescription.StatusSigned && status != prescription.StatusSent && status != prescription.StatusArchived {
		writeErr(w, r, http.StatusBadRequest, "invalid_status", "invalid_status")
		return
	}
	petID := strings.TrimSpace(r.URL.Query().Get("petId"))
	if petID != "" && !isUUID(petID) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	items, err := a.store.ListPrescriptions(r.Context(), id.PracticeID, status, petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.Prescription{}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) createPrescription(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	var body struct {
		PetID       string          `json:"petId"`
		Notes       string          `json:"notes"`
		PaperFormat string          `json:"paperFormat"`
		ValidUntil  *string         `json:"validUntil"`
		Medications json.RawMessage `json:"medications"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	petID := strings.TrimSpace(body.PetID)
	if petID == "" || !isUUID(petID) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	can, err := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermWriteNotes)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !can {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	meds, err := prescription.NormalizeMedications(body.Medications)
	if err != nil {
		a.writePrescriptionErr(w, r, err)
		return
	}
	var validUntil *time.Time
	if body.ValidUntil != nil && strings.TrimSpace(*body.ValidUntil) != "" {
		t, perr := parsePrescriptionTime(strings.TrimSpace(*body.ValidUntil))
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "invalid_valid_until", "invalid_valid_until")
			return
		}
		validUntil = &t
	}
	doc, err := a.store.CreatePrescriptionDraft(
		r.Context(), id.PracticeID, id.UserID, pet, meds, body.Notes, body.PaperFormat, validUntil,
	)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) prescriptionIDParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	raw := chi.URLParam(r, "id")
	if !isUUID(raw) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return "", false
	}
	return raw, true
}

func (a *API) getPrescription(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	rxID, ok := a.prescriptionIDParam(w, r)
	if !ok {
		return
	}
	doc, err := a.store.GetPrescription(r.Context(), id.PracticeID, rxID)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) patchPrescription(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	rxID, ok := a.prescriptionIDParam(w, r)
	if !ok {
		return
	}
	var body struct {
		Notes       *string         `json:"notes"`
		PaperFormat *string         `json:"paperFormat"`
		ValidUntil  *string         `json:"validUntil"`
		Medications json.RawMessage `json:"medications"`
		Status      *string         `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	// V1: refuse non-draft status transitions.
	if body.Status != nil {
		st := strings.TrimSpace(*body.Status)
		if st != "" && st != prescription.StatusDraft {
			writeErr(w, r, http.StatusBadRequest, "invalid_status", "invalid_status")
			return
		}
	}
	patch := store.PrescriptionPatch{
		Notes:       body.Notes,
		PaperFormat: body.PaperFormat,
	}
	if body.Medications != nil {
		raw := body.Medications
		patch.Medications = &raw
	}
	if body.ValidUntil != nil {
		v := strings.TrimSpace(*body.ValidUntil)
		if v == "" {
			patch.ClearValid = true
		} else {
			t, perr := parsePrescriptionTime(v)
			if perr != nil {
				writeErr(w, r, http.StatusBadRequest, "invalid_valid_until", "invalid_valid_until")
				return
			}
			patch.ValidUntil = &t
		}
	}
	doc, err := a.store.PatchPrescriptionDraft(r.Context(), id.PracticeID, rxID, patch)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) deletePrescription(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	rxID, ok := a.prescriptionIDParam(w, r)
	if !ok {
		return
	}
	if err := a.store.DeletePrescriptionDraft(r.Context(), id.PracticeID, rxID); err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) getPrescriptionPDF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	rxID, ok := a.prescriptionIDParam(w, r)
	if !ok {
		return
	}
	doc, err := a.store.GetPrescription(r.Context(), id.PracticeID, rxID)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	meds, err := prescription.NormalizeMedications(doc.Medications)
	if err != nil {
		a.writePrescriptionErr(w, r, err)
		return
	}
	var issued, valid *time.Time
	if doc.DateIssued != nil {
		if t, perr := time.Parse(time.RFC3339, *doc.DateIssued); perr == nil {
			issued = &t
		}
	}
	if doc.ValidUntil != nil {
		if t, perr := time.Parse(time.RFC3339, *doc.ValidUntil); perr == nil {
			valid = &t
		}
	}
	pdfBytes, err := prescription.BuildPDF(prescription.PDFInput{
		PaperFormat:    doc.PaperFormat,
		CountryCode:    doc.CountryCode,
		PracticeName:   doc.PracticeName,
		Prescriber:     doc.VeterinaryName,
		OwnerName:      doc.OwnerName,
		PetName:        doc.PetName,
		PetSpecies:     doc.PetSpecies,
		DateIssued:     issued,
		ValidUntil:     valid,
		Notes:          doc.Notes,
		Medications:    meds,
		DraftWatermark: doc.Status == prescription.StatusDraft,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	fname := "prescription.pdf"
	if isUUID(doc.ID) {
		fname = "prescription-" + doc.ID + ".pdf"
	}
	w.Header().Set("Content-Disposition", `inline; filename="`+fname+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func parsePrescriptionTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid time")
}
