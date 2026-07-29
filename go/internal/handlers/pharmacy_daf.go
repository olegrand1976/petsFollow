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
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyDAFRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/daf", a.listPharmacyDAF)
	pr.Post("/vet/pharmacy/daf", a.createPharmacyDAF)
	pr.Post("/vet/pharmacy/daf/preview-fefo", a.previewPharmacyDAFFEFO)
	pr.Get("/vet/pharmacy/daf/{id}", a.getPharmacyDAF)
	pr.Patch("/vet/pharmacy/daf/{id}", a.patchPharmacyDAF)
	pr.Post("/vet/pharmacy/daf/{id}/finalize", a.finalizePharmacyDAF)
	pr.Post("/vet/pharmacy/daf/{id}/cancel", a.cancelPharmacyDAF)
	pr.Get("/vet/pharmacy/daf/{id}/pdf", a.getPharmacyDAFPDF)
	pr.Post("/vet/pharmacy/daf/{id}/pdf/regenerate", a.regeneratePharmacyDAFPDF)
	pr.Post("/vet/pharmacy/daf/{id}/vamreg/retry", a.retryPharmacyDAFVAMReg)
	pr.Get("/vet/pharmacy/daf/{id}/vamreg/audits", a.listPharmacyDAFVAMRegAudits)
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
		writeErr(w, r, http.StatusConflict, "stock_insufficient", "stock_insufficient")
	case errors.Is(err, pharmacy.ErrStockUnavailableValidLots):
		writeErr(w, r, http.StatusConflict, "stock_unavailable_valid_lots", "stock_unavailable_valid_lots")
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
	doc, err := a.store.CreateDAFDraft(r.Context(), id.PracticeID, id.UserID, body.ClientUserID, body.PetID, body.VisitID, body.Notes, decodeDAFItems(body.Items))
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
	if doc.PetID != "" {
		foodChain, _ = a.store.GetPetFoodChainStatus(r.Context(), doc.PetID)
	}
	pdfBytes, err := pharmacy.BuildDAFPDF(pharmacy.DAFPDFInput{
		DisplayNumber:   doc.DisplayNumber,
		PracticeName:    doc.PracticeName,
		Prescriber:      doc.PrescriberName,
		ClientName:      doc.ClientName,
		PetName:         doc.PetName,
		FoodChainStatus: foodChain,
		IssuedAt:        issued,
		Notes:           doc.Notes,
		Lines:           lines,
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
