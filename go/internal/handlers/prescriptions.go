package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/prescription"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPrescriptionRoutes(pr chi.Router) {
	pr.Get("/vet/prescriptions", a.listPrescriptions)
	pr.Post("/vet/prescriptions", a.createPrescription)
	pr.Post("/vet/prescriptions/suggest-from-visit", a.suggestPrescriptionFromVisit)
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
	case errors.Is(err, prescription.ErrInvalidMeds):
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

// resolvePrescriptionVisit checks visit belongs to practice + pet and returns visitID (or "" if unset).
func (a *API) resolvePrescriptionVisit(w http.ResponseWriter, r *http.Request, practiceID, petID, visitID string) (string, bool) {
	visitID = strings.TrimSpace(visitID)
	if visitID == "" {
		return "", true
	}
	if !isUUID(visitID) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return "", false
	}
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
			return "", false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return "", false
	}
	if visit.PracticeID != practiceID || visit.PetID != petID {
		writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
		return "", false
	}
	return visitID, true
}

func (a *API) listPrescriptions(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	q := r.URL.Query()
	f := store.ListPrescriptionsFilter{
		Status: strings.TrimSpace(q.Get("status")),
		PetID:  strings.TrimSpace(q.Get("petId")),
		Query:  strings.TrimSpace(q.Get("q")),
	}
	if f.Status != "" && f.Status != prescription.StatusDraft &&
		f.Status != prescription.StatusSigned && f.Status != prescription.StatusSent && f.Status != prescription.StatusArchived {
		writeErr(w, r, http.StatusBadRequest, "invalid_status", "invalid_status")
		return
	}
	if f.PetID != "" && !isUUID(f.PetID) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	if raw := strings.TrimSpace(q.Get("from")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_from")
			return
		}
		f.From = &t
	}
	if raw := strings.TrimSpace(q.Get("to")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_to")
			return
		}
		f.To = &t
	}
	items, err := a.store.ListPrescriptions(r.Context(), id.PracticeID, f)
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
		VisitID     string          `json:"visitId"`
		Notes       string          `json:"notes"`
		CareAdvice  string          `json:"careAdvice"`
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
	visitID, ok := a.resolvePrescriptionVisit(w, r, id.PracticeID, pet.ID, body.VisitID)
	if !ok {
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
		r.Context(), id.PracticeID, id.UserID, pet, meds, body.Notes, body.CareAdvice, body.PaperFormat, validUntil, visitID,
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
		CareAdvice  *string         `json:"careAdvice"`
		PaperFormat *string         `json:"paperFormat"`
		ValidUntil  *string         `json:"validUntil"`
		VisitID     *string         `json:"visitId"`
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
	cur, err := a.store.GetPrescription(r.Context(), id.PracticeID, rxID)
	if err != nil {
		if a.writePrescriptionErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	patch := store.PrescriptionPatch{
		Notes:       body.Notes,
		CareAdvice:  body.CareAdvice,
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
	if body.VisitID != nil {
		vid, ok := a.resolvePrescriptionVisit(w, r, id.PracticeID, cur.PetID, *body.VisitID)
		if !ok {
			return
		}
		patch.VisitID = &vid
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
		CareAdvice:     doc.CareAdvice,
		Medications:    meds,
		DraftWatermark: doc.Status == prescription.StatusDraft,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	fname := "consignes.pdf"
	if isUUID(doc.ID) {
		fname = "consignes-" + doc.ID + ".pdf"
	}
	w.Header().Set("Content-Disposition", `inline; filename="`+fname+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func (a *API) suggestPrescriptionFromVisit(w http.ResponseWriter, r *http.Request) {
	if !a.requirePrescriptionsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	if a.vetSuggestRL != nil && !a.vetSuggestRL.Allow("consignes-suggest:"+id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
		return
	}
	var body struct {
		VisitID string `json:"visitId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	visitID := strings.TrimSpace(body.VisitID)
	if visitID == "" || !isUUID(visitID) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if visit.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
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

	reports, err := a.store.ListVisitReportsForVisit(r.Context(), visitID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	source := pickVisitReportTextForConsignes(reports)
	if strings.TrimSpace(source) == "" {
		writeErr(w, r, http.StatusBadRequest, "no_visit_report", "no_visit_report")
		return
	}
	source = truncateRunes(source, maxConsignesSuggestSourceRunes)

	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "gemini_not_configured")
		return
	}

	country := "BE"
	if contact, cerr := a.store.GetPracticeContact(r.Context(), visit.PracticeID); cerr == nil && contact.CountryCode != "" {
		country = store.NormalizeCountryCode(contact.CountryCode)
	}
	system := gemini.BuildConsignesSuggestPrompt(gemini.ConsignesSuggestPromptInput{
		CountryCode: country,
		PetName:     pet.Name,
		PetSpecies:  pet.Species,
	})
	raw, err := a.gemini.GenerateJSON(r.Context(), system, "Compte-rendu source:\n\n"+source, 0.2)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "internal")
		return
	}
	suggestion, err := gemini.ParseConsignesSuggestJSON(raw)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, suggestion)
}

const maxConsignesSuggestSourceRunes = 12000

func pickVisitReportTextForConsignes(reports []store.VisitReportSummary) string {
	// Prefer finalized, then improved, then body, then transcript — newest list first.
	bestFinal := ""
	bestDraft := ""
	for _, r := range reports {
		text := strings.TrimSpace(r.ImprovedText)
		if text == "" {
			text = strings.TrimSpace(r.BodyText)
		}
		if text == "" {
			text = strings.TrimSpace(r.TranscriptText)
		}
		if text == "" {
			continue
		}
		if r.Status == "final" && bestFinal == "" {
			bestFinal = text
		}
		if bestDraft == "" {
			bestDraft = text
		}
	}
	if bestFinal != "" {
		return bestFinal
	}
	return bestDraft
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
