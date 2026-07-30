package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/prescription"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyDAFRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/daf", a.listPharmacyDAF)
	pr.Post("/vet/pharmacy/daf", a.createPharmacyDAF)
	pr.Post("/vet/pharmacy/daf/preview-fefo", a.previewPharmacyDAFFEFO)
	pr.Get("/vet/pharmacy/daf/for-visit", a.getPharmacyDAFForVisit)
	pr.Put("/vet/pharmacy/daf/for-visit", a.upsertPharmacyDAFForVisit)
	pr.Post("/vet/pharmacy/daf/from-prescription", a.createPharmacyDAFFromPrescription)
	pr.Get("/vet/pharmacy/daf/{id}", a.getPharmacyDAF)
	pr.Patch("/vet/pharmacy/daf/{id}", a.patchPharmacyDAF)
	pr.Post("/vet/pharmacy/daf/{id}/finalize", a.finalizePharmacyDAF)
	pr.Post("/vet/pharmacy/daf/{id}/cancel", a.cancelPharmacyDAF)
	pr.Get("/vet/pharmacy/daf/{id}/pdf", a.getPharmacyDAFPDF)
	pr.Post("/vet/pharmacy/daf/{id}/pdf/regenerate", a.regeneratePharmacyDAFPDF)
	pr.Post("/vet/pharmacy/daf/{id}/vamreg/retry", a.retryPharmacyDAFVAMReg)
	pr.Get("/vet/pharmacy/daf/{id}/vamreg/audits", a.listPharmacyDAFVAMRegAudits)
	pr.Get("/vet/consultations/daf-drafts", a.listConsultationDAFDrafts)
}

func (a *API) writeDAFErr(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case errors.Is(err, pharmacy.ErrDAFNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, pharmacy.ErrDAFNotDraft):
		writeErr(w, r, http.StatusConflict, "daf_not_draft", "daf_not_draft")
	case errors.Is(err, pharmacy.ErrDAFNotFinalized):
		writeErr(w, r, http.StatusConflict, "daf_not_finalized", "daf_not_finalized")
	case errors.Is(err, pharmacy.ErrDAFEmpty):
		writeErr(w, r, http.StatusBadRequest, "daf_empty", "daf_empty")
	case errors.Is(err, pharmacy.ErrDAFAMMRequired):
		writeErr(w, r, http.StatusBadRequest, "daf_amm_required", "daf_amm_required")
	case errors.Is(err, pharmacy.ErrDAFVAMRegIncomplete):
		writeErr(w, r, http.StatusBadRequest, "daf_vamreg_incomplete", "daf_vamreg_incomplete")
	case errors.Is(err, pharmacy.ErrDAFNotAntibiotic):
		writeErr(w, r, http.StatusConflict, "daf_vamreg_not_applicable", "daf_vamreg_not_applicable")
	case errors.Is(err, pharmacy.ErrDAFAlreadyHasPDF):
		writeErr(w, r, http.StatusConflict, "daf_pdf_immutable", "daf_pdf_immutable")
	case errors.Is(err, pharmacy.ErrDAFTraceRequired):
		writeErr(w, r, http.StatusBadRequest, "daf_trace_required", "daf_trace_required")
	case errors.Is(err, pharmacy.ErrDAFVAMRegAlreadySent):
		writeErr(w, r, http.StatusConflict, "daf_vamreg_already_sent", "daf_vamreg_already_sent")
	case errors.Is(err, pharmacy.ErrDAFVAMRegInFlight):
		writeErr(w, r, http.StatusConflict, "daf_vamreg_in_flight", "daf_vamreg_in_flight")
	case errors.Is(err, pharmacy.ErrFoodChainWithdrawalRequired):
		writeErr(w, r, http.StatusBadRequest, "food_chain_withdrawal_required", "food_chain_withdrawal_required")
	case errors.Is(err, pharmacy.ErrFoodChainBannedMedication):
		writeErr(w, r, http.StatusConflict, "food_chain_banned_medication", "food_chain_banned_medication")
	case errors.Is(err, pharmacy.ErrStockInsufficient):
		if medID := pharmacy.StockMedID(err); medID != "" {
			writeErrDetails(w, r, http.StatusConflict, "stock_insufficient", "stock_insufficient", map[string]any{"medicationId": medID})
		} else {
			writeErr(w, r, http.StatusConflict, "stock_insufficient", "stock_insufficient")
		}
	case errors.Is(err, pharmacy.ErrStockUnavailableValidLots):
		if medID := pharmacy.StockMedID(err); medID != "" {
			writeErrDetails(w, r, http.StatusConflict, "stock_unavailable_valid_lots", "stock_unavailable_valid_lots", map[string]any{"medicationId": medID})
		} else {
			writeErr(w, r, http.StatusConflict, "stock_unavailable_valid_lots", "stock_unavailable_valid_lots")
		}
	case errors.Is(err, store.ErrValidation), errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
	default:
		return false
	}
	return true
}

func decodeDAFItems(raw []store.DAFItemInput) []store.DAFItemInput {
	out := make([]store.DAFItemInput, 0, len(raw))
	for _, it := range raw {
		it.MedicationID = strings.TrimSpace(it.MedicationID)
		it.DepositID = strings.TrimSpace(it.DepositID)
		it.AMMNumber = strings.TrimSpace(it.AMMNumber)
		if it.Unit == "" {
			it.Unit = "unit"
		}
		out = append(out, it)
	}
	return out
}

// validateDAFLinks ensures visit/pet/client belong to the practice and are coherent.
func (a *API) validateDAFLinks(w http.ResponseWriter, r *http.Request, practiceID, clientUserID, petID, visitID string) bool {
	clientUserID = strings.TrimSpace(clientUserID)
	petID = strings.TrimSpace(petID)
	visitID = strings.TrimSpace(visitID)
	if visitID == "" && petID == "" && clientUserID == "" {
		return true
	}
	if visitID != "" {
		visit, err := a.store.GetVisit(r.Context(), visitID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
				return false
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return false
		}
		if visit.PracticeID != practiceID {
			writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
			return false
		}
		if petID != "" && visit.PetID != petID {
			writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
			return false
		}
		if petID == "" {
			petID = visit.PetID
		}
	}
	if petID != "" {
		pet, err := a.store.GetPet(r.Context(), petID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
				return false
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return false
		}
		if pet.PracticeID != practiceID {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return false
		}
		if clientUserID != "" && pet.OwnerUserID != clientUserID {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return false
		}
	} else if clientUserID != "" {
		if _, err := a.store.GetClientByPractice(r.Context(), practiceID, clientUserID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
				return false
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return false
		}
	}
	return true
}

func (a *API) listPharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	items, err := a.store.ListDAF(r.Context(), id.PracticeID, r.URL.Query().Get("status"), year)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) createPharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		ClientUserID string               `json:"clientUserId"`
		PetID        string               `json:"petId"`
		VisitID      string               `json:"visitId"`
		Notes        string               `json:"notes"`
		Items        []store.DAFItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if !a.validateDAFLinks(w, r, id.PracticeID, body.ClientUserID, body.PetID, body.VisitID) {
		return
	}
	var doc store.DAFDocument
	var err error
	if strings.TrimSpace(body.VisitID) != "" {
		doc, err = a.store.UpsertDraftDAFForVisit(
			r.Context(), id.PracticeID, id.UserID,
			body.ClientUserID, body.PetID, body.VisitID, body.Notes, decodeDAFItems(body.Items),
		)
	} else {
		doc, err = a.store.CreateDAFDraft(
			r.Context(), id.PracticeID, id.UserID,
			body.ClientUserID, body.PetID, body.VisitID, body.Notes, decodeDAFItems(body.Items),
		)
	}
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) getPharmacyDAFForVisit(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	visitID := strings.TrimSpace(r.URL.Query().Get("visitId"))
	if visitID == "" {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	if !a.validateDAFLinks(w, r, id.PracticeID, "", "", visitID) {
		return
	}
	doc, err := a.store.GetDraftDAFByVisit(r.Context(), id.PracticeID, visitID)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) upsertPharmacyDAFForVisit(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		ClientUserID string               `json:"clientUserId"`
		PetID        string               `json:"petId"`
		VisitID      string               `json:"visitId"`
		Notes        string               `json:"notes"`
		Items        []store.DAFItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	body.VisitID = strings.TrimSpace(body.VisitID)
	if body.VisitID == "" {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	if !a.validateDAFLinks(w, r, id.PracticeID, body.ClientUserID, body.PetID, body.VisitID) {
		return
	}
	// Fill pet/client from visit when omitted.
	if body.PetID == "" || body.ClientUserID == "" {
		visit, err := a.store.GetVisit(r.Context(), body.VisitID)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
			return
		}
		if body.PetID == "" {
			body.PetID = visit.PetID
		}
		if body.ClientUserID == "" {
			pet, perr := a.store.GetPet(r.Context(), body.PetID)
			if perr == nil {
				body.ClientUserID = pet.OwnerUserID
			}
		}
	}
	doc, err := a.store.UpsertDraftDAFForVisit(
		r.Context(), id.PracticeID, id.UserID,
		body.ClientUserID, body.PetID, body.VisitID, body.Notes, decodeDAFItems(body.Items),
	)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) createPharmacyDAFFromPrescription(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	if !a.cfg.PrescriptionsEnabled {
		writeErr(w, r, http.StatusNotFound, "prescriptions_disabled", "prescriptions_disabled")
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		PrescriptionID string `json:"prescriptionId"`
		VisitID        string `json:"visitId"`
		Notes          string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	body.PrescriptionID = strings.TrimSpace(body.PrescriptionID)
	if body.PrescriptionID == "" {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	rx, err := a.store.GetPrescription(r.Context(), id.PracticeID, body.PrescriptionID)
	if err != nil {
		if errors.Is(err, prescription.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	meds, err := prescription.NormalizeMedications(rx.Medications)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	items := make([]store.DAFItemInput, 0, len(meds))
	for _, m := range meds {
		if m.RefMedicationID == nil || strings.TrimSpace(*m.RefMedicationID) == "" {
			continue
		}
		qty := 1.0
		if q := strings.TrimSpace(m.Quantity); q != "" {
			if parsed, perr := strconv.ParseFloat(q, 64); perr == nil && parsed > 0 {
				qty = parsed
			}
		}
		items = append(items, store.DAFItemInput{
			MedicationID: strings.TrimSpace(*m.RefMedicationID),
			Qty:          qty,
			Unit:         "unit",
			AMMNumber:    "",
		})
		if med, merr := a.store.GetRefMedicationForPractice(r.Context(), id.PracticeID, strings.TrimSpace(*m.RefMedicationID)); merr == nil {
			items[len(items)-1].AMMNumber = strings.TrimSpace(med.AMMNumber)
		}
	}
	if len(items) == 0 {
		writeErr(w, r, http.StatusBadRequest, "daf_empty", "daf_empty")
		return
	}
	notes := strings.TrimSpace(body.Notes)
	if notes == "" {
		notes = strings.TrimSpace(rx.Notes)
	}
	visitID := strings.TrimSpace(body.VisitID)
	if visitID != "" {
		if !a.validateDAFLinks(w, r, id.PracticeID, rx.OwnerID, rx.PetID, visitID) {
			return
		}
		doc, err := a.store.UpsertDraftDAFForVisit(
			r.Context(), id.PracticeID, id.UserID,
			rx.OwnerID, rx.PetID, visitID, notes, items,
		)
		if err != nil {
			if a.writeDAFErr(w, r, err) {
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusCreated, doc)
		return
	}
	doc, err := a.store.CreateDAFDraft(
		r.Context(), id.PracticeID, id.UserID,
		rx.OwnerID, rx.PetID, "", notes, items,
	)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) previewPharmacyDAFFEFO(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	var body struct {
		Items []store.DAFItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	lines, err := a.store.PreviewFEFOAllocation(r.Context(), id.PracticeID, decodeDAFItems(body.Items), time.Now())
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"lines": lines})
}

func (a *API) getPharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	doc, err := a.store.GetDAF(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) patchPharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		ClientUserID string               `json:"clientUserId"`
		PetID        string               `json:"petId"`
		VisitID      string               `json:"visitId"`
		Notes        string               `json:"notes"`
		Items        []store.DAFItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if !a.validateDAFLinks(w, r, id.PracticeID, body.ClientUserID, body.PetID, body.VisitID) {
		return
	}
	doc, err := a.store.ReplaceDAFDraftItems(r.Context(), id.PracticeID, chi.URLParam(r, "id"), body.ClientUserID, body.PetID, body.VisitID, body.Notes, decodeDAFItems(body.Items))
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) finalizePharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	dafID := chi.URLParam(r, "id")
	doc, err := a.store.FinalizeDAF(r.Context(), id.PracticeID, dafID, id.UserID, time.Now())
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// PDF after commit (failure does not roll back number).
	pdfErr := a.ensureDAFPDF(r, id.PracticeID, dafID)
	a.enqueueVamregIfNeeded(r.Context(), id.PracticeID, doc)
	a.recordDAFClientTimelineEvent(r.Context(), id.UserID, doc)
	doc, _ = a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	if pdfErr != nil {
		httpx.WriteData(w, http.StatusOK, map[string]any{"daf": doc, "pdfError": pdfErr.Error()})
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) enqueueVamregIfNeeded(ctx context.Context, practiceID string, doc store.DAFDocument) {
	if !doc.HasAntibiotic || doc.VamregStatus != "pending" {
		return
	}
	if a.vamregQ == nil {
		log.Printf("vamreg enqueue unavailable daf=%s — marking failed", doc.ID)
		if err := a.store.UpdateDAFVamregStatus(ctx, practiceID, doc.ID, "failed"); err != nil {
			log.Printf("vamreg mark failed daf=%s: %v", doc.ID, err)
		}
		return
	}
	if err := a.vamregQ.EnqueueDeclare(ctx, practiceID, doc.ID); err != nil {
		log.Printf("vamreg enqueue daf=%s: %v — marking failed for retry", doc.ID, err)
		if uerr := a.store.UpdateDAFVamregStatus(ctx, practiceID, doc.ID, "failed"); uerr != nil {
			log.Printf("vamreg mark failed daf=%s: %v", doc.ID, uerr)
		}
	}
}

func (a *API) retryPharmacyDAFVAMReg(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	dafID := chi.URLParam(r, "id")
	status, vamreg, hasAB, err := a.store.GetDAFVamregGate(r.Context(), id.PracticeID, dafID)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if status != "finalized" || !hasAB {
		writeErr(w, r, http.StatusConflict, "daf_vamreg_not_applicable", "daf_vamreg_not_applicable")
		return
	}
	if vamreg == "sent" {
		doc, _ := a.store.GetDAF(r.Context(), id.PracticeID, dafID)
		httpx.WriteData(w, http.StatusOK, doc)
		return
	}
	// failed + stuck pending (enqueue crash / worker never picked up) are both retryable.
	// VAMReg Idempotency-Key = dafID; ProcessDeclare no-ops once status is sent.
	if vamreg != "failed" && vamreg != "pending" {
		writeErr(w, r, http.StatusConflict, "daf_vamreg_not_applicable", "daf_vamreg_not_applicable")
		return
	}
	if err := a.store.UpdateDAFVamregStatus(r.Context(), id.PracticeID, dafID, "pending"); err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.vamregQ == nil {
		_ = a.store.UpdateDAFVamregStatus(r.Context(), id.PracticeID, dafID, "failed")
		writeErr(w, r, http.StatusBadGateway, "vamreg_enqueue_failed", "vamreg_enqueue_failed")
		return
	}
	if err := a.vamregQ.EnqueueDeclare(r.Context(), id.PracticeID, dafID); err != nil {
		log.Printf("vamreg retry enqueue daf=%s: %v", dafID, err)
		_ = a.store.UpdateDAFVamregStatus(r.Context(), id.PracticeID, dafID, "failed")
		writeErr(w, r, http.StatusBadGateway, "vamreg_enqueue_failed", "vamreg_enqueue_failed")
		return
	}
	doc, _ := a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) listPharmacyDAFVAMRegAudits(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	dafID := chi.URLParam(r, "id")
	if _, err := a.store.GetDAF(r.Context(), id.PracticeID, dafID); err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	rows, err := a.store.ListPharmacyJobAudits(r.Context(), id.PracticeID, "vamreg", dafID, 20)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.PharmacyJobAudit{}
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) cancelPharmacyDAF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		Reason  string `json:"reason"`
		Restock *bool  `json:"restock"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	restock := true
	if body.Restock != nil {
		restock = *body.Restock
	}
	doc, err := a.store.CancelDAF(r.Context(), id.PracticeID, chi.URLParam(r, "id"), id.UserID, body.Reason, restock)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) getPharmacyDAFPDF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	dafID := chi.URLParam(r, "id")
	doc, err := a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if doc.Status == "draft" {
		writeErr(w, r, http.StatusConflict, "daf_not_finalized", "daf_not_finalized")
		return
	}
	if doc.PDFObjectKey == "" {
		if err := a.ensureDAFPDF(r, id.PracticeID, dafID); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "pdf_error", "pdf_error")
			return
		}
		doc, _ = a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	}
	if a.media == nil || doc.PDFObjectKey == "" {
		writeErr(w, r, http.StatusNotFound, "not_found", "pdf_missing")
		return
	}
	rc, ct, err := a.media.Open(r.Context(), doc.PDFObjectKey)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pdf_missing")
		return
	}
	defer rc.Close()
	if ct == "" {
		ct = "application/pdf"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, doc.DisplayNumber))
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = io.Copy(w, rc)
}

func (a *API) regeneratePharmacyDAFPDF(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	dafID := chi.URLParam(r, "id")
	doc, err := a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	if err != nil {
		if a.writeDAFErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if doc.PDFObjectKey != "" {
		writeErr(w, r, http.StatusConflict, "daf_pdf_immutable", "daf_pdf_immutable")
		return
	}
	if err := a.ensureDAFPDF(r, id.PracticeID, dafID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "pdf_error", "pdf_error")
		return
	}
	doc, _ = a.store.GetDAF(r.Context(), id.PracticeID, dafID)
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) ensureDAFPDF(r *http.Request, practiceID, dafID string) error {
	if a.media == nil {
		return fmt.Errorf("media_not_configured")
	}
	doc, err := a.store.GetDAF(r.Context(), practiceID, dafID)
	if err != nil {
		return err
	}
	if doc.PDFObjectKey != "" {
		return nil
	}
	if doc.Status != "finalized" && doc.Status != "cancelled" {
		return pharmacy.ErrDAFNotFinalized
	}
	issued := time.Now()
	if doc.FinalizedAt != "" {
		if t, e := time.Parse(time.RFC3339Nano, doc.FinalizedAt); e == nil {
			issued = t
		} else if t, e := time.Parse("2006-01-02 15:04:05.999999-07", doc.FinalizedAt); e == nil {
			issued = t
		}
	}
	lines := make([]pharmacy.DAFPDFLine, 0, len(doc.Items))
	for _, it := range doc.Items {
		lines = append(lines, pharmacy.DAFPDFLine{
			Medication:         it.MedicationName,
			CNK:                it.MedicationCNK,
			AMM:                it.AMMNumber,
			Lot:                it.LotNumber,
			ExpiresOn:          it.ExpiresOn,
			Qty:                fmt.Sprintf("%g", it.Qty),
			Unit:               it.Unit,
			Antibiotic:         it.IsAntibiotic,
			WithdrawalMeatDays: it.WithdrawalMeatDays,
			WithdrawalMilkDays: it.WithdrawalMilkDays,
			WithdrawalEggsDays: it.WithdrawalEggsDays,
		})
	}
	foodChain := ""
	domicile := ""
	if doc.PetID != "" {
		foodChain, _ = a.store.GetPetFoodChainStatus(r.Context(), doc.PetID)
		domicile, _ = a.store.GetPetDomicileLocation(r.Context(), doc.PetID)
	}
	pdfBytes, err := pharmacy.BuildDAFPDF(pharmacy.DAFPDFInput{
		DisplayNumber:    doc.DisplayNumber,
		PracticeName:     doc.PracticeName,
		Prescriber:       doc.PrescriberName,
		ClientName:       doc.ClientName,
		PetName:          doc.PetName,
		FoodChainStatus:  foodChain,
		DomicileLocation: domicile,
		IssuedAt:         issued,
		Notes:            doc.Notes,
		Lines:            lines,
	})
	if err != nil {
		return err
	}
	sum := sha256.Sum256(pdfBytes)
	sha := hex.EncodeToString(sum[:])
	key := fmt.Sprintf("daf/%s/%s.pdf", practiceID, dafID)
	if _, err := a.media.Upload(r.Context(), key, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf"); err != nil {
		return err
	}
	if err := a.store.SetDAFPDFMeta(r.Context(), practiceID, dafID, key, sha); err != nil {
		if errors.Is(err, pharmacy.ErrDAFAlreadyHasPDF) {
			return nil
		}
		return err
	}
	return nil
}

func (a *API) recordDAFClientTimelineEvent(ctx context.Context, authorUserID string, doc store.DAFDocument) {
	if strings.TrimSpace(doc.PetID) == "" {
		return
	}
	locale := "fr"
	if ownerID := strings.TrimSpace(doc.ClientUserID); ownerID != "" {
		if loc, err := a.store.GetUserPreferredLocale(ctx, ownerID); err == nil && loc != "" {
			locale = loc
		}
	}
	title := i18n.T(locale, "daf.timeline_title", nil)
	body := i18n.T(locale, "daf.timeline_body", nil)
	if doc.DisplayNumber != "" {
		body = body + " (" + doc.DisplayNumber + ")"
	}
	// Non-PHI summary only (no drug names) for client-visible dossier_events.
	// event_type is shown as timeline title on Pro + Flutter.
	_ = a.store.InsertDossierEvent(ctx, doc.PetID, authorUserID, title, body)
}

func (a *API) listPetDAFDispenses(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermRead)
	if !ok {
		return
	}
	practiceID := id.PracticeID
	if practiceID == "" {
		practiceID = pet.PracticeID
	}
	if practiceID == "" {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	// Staff need pharmacy.read (owners use timeline dossier events — non-PHI).
	if !a.allowPracticePerm(r, id, "pharmacy.read") {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	docs, err := a.store.ListFinalizedDAFByPet(r.Context(), practiceID, petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	items := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		lines := make([]map[string]any, 0, len(d.Items))
		for _, it := range d.Items {
			line := map[string]any{
				"medicationName": it.MedicationName,
				"qty":            it.Qty,
				"unit":           it.Unit,
				"lotNumber":      it.LotNumber,
				"ammNumber":      it.AMMNumber,
				"expiresOn":      it.ExpiresOn,
				"medicationId":   it.MedicationID,
			}
			lines = append(lines, line)
		}
		items = append(items, map[string]any{
			"dafId":         d.ID,
			"displayNumber": d.DisplayNumber,
			"finalizedAt":   d.FinalizedAt,
			"items":         lines,
		})
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) listConsultationDAFDrafts(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("visitIds"))
	if raw == "" {
		httpx.WriteData(w, http.StatusOK, map[string]any{"drafts": map[string]string{}})
		return
	}
	parts := strings.Split(raw, ",")
	visitIDs := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && isUUID(p) {
			visitIDs = append(visitIDs, p)
		}
	}
	drafts, err := a.store.ListStaleDraftDAFByVisits(r.Context(), id.PracticeID, visitIDs, time.Hour)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"drafts": drafts})
}
