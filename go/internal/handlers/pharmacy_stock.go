package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyStockRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/settings", a.getPharmacySettings)
	pr.Patch("/vet/pharmacy/settings", a.patchPharmacySettings)
	pr.Get("/vet/pharmacy/deposits", a.listPharmacyDeposits)
	pr.Post("/vet/pharmacy/deposits", a.createPharmacyDeposit)
	pr.Get("/vet/pharmacy/batches/export.csv", a.exportPharmacyBatchesCSV)
	pr.Get("/vet/pharmacy/batches", a.listPharmacyBatches)
	pr.Post("/vet/pharmacy/batches", a.receivePharmacyBatch)
	pr.Post("/vet/pharmacy/batches/{id}/adjust", a.adjustPharmacyBatch)
	pr.Post("/vet/pharmacy/batches/{id}/quarantine", a.quarantinePharmacyBatch)
	pr.Post("/vet/pharmacy/batches/{id}/waste", a.wastePharmacyBatch)
	pr.Get("/vet/pharmacy/expiry/summary", a.getPharmacyExpirySummary)
	pr.Get("/vet/pharmacy/movements", a.listPharmacyMovements)
	pr.Get("/vet/pharmacy/prices", a.listPharmacyPrices)
	pr.Put("/vet/pharmacy/prices/{medicationId}", a.putPharmacyPrice)
	pr.Get("/vet/pharmacy/prices/{medicationId}", a.getPharmacyPrice)
	pr.Put("/vet/pharmacy/reorder-thresholds", a.putPharmacyReorderThreshold)
	pr.Get("/vet/pharmacy/reorder-alerts", a.listPharmacyReorderAlerts)
}

func (a *API) writePharmacyErr(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case errors.Is(err, pharmacy.ErrInvalidExpiryOnReceipt):
		writeErr(w, r, http.StatusBadRequest, "invalid_expiry_on_receipt", "invalid_expiry_on_receipt")
	case errors.Is(err, pharmacy.ErrBatchExpired):
		writeErr(w, r, http.StatusConflict, "batch_expired", "batch_expired")
	case errors.Is(err, pharmacy.ErrBatchQuarantined):
		writeErr(w, r, http.StatusConflict, "batch_quarantined", "batch_quarantined")
	case errors.Is(err, pharmacy.ErrStockInsufficient):
		writeErr(w, r, http.StatusConflict, "stock_insufficient", "stock_insufficient")
	case errors.Is(err, pharmacy.ErrStockUnavailableValidLots):
		writeErr(w, r, http.StatusConflict, "stock_unavailable_valid_lots", "stock_unavailable_valid_lots")
	case errors.Is(err, pharmacy.ErrDAFTraceRequired):
		writeErr(w, r, http.StatusBadRequest, "daf_trace_required", "daf_trace_required")
	case errors.Is(err, pharmacy.ErrBatchNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, pharmacy.ErrBatchWasted):
		writeErr(w, r, http.StatusConflict, "batch_wasted", "batch_wasted")
	case errors.Is(err, pharmacy.ErrFoodChainWithdrawalRequired):
		writeErr(w, r, http.StatusBadRequest, "food_chain_withdrawal_required", "food_chain_withdrawal_required")
	case errors.Is(err, pharmacy.ErrFoodChainBannedMedication):
		writeErr(w, r, http.StatusConflict, "food_chain_banned_medication", "food_chain_banned_medication")
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
	case errors.Is(err, store.ErrConflict):
		writeErr(w, r, http.StatusConflict, "conflict", "conflict")
	default:
		return false
	}
	return true
}

func (a *API) getPharmacySettings(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	st, err := a.store.GetPharmacySettings(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, st)
}

func (a *API) patchPharmacySettings(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	st, err := a.store.GetPharmacySettings(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if v, ok := body["warnSoonDays"].(float64); ok {
		st.WarnSoonDays = int(v)
	}
	if v, ok := body["warnReturnDays"].(float64); ok {
		st.WarnReturnDays = int(v)
	}
	if v, ok := body["warnCriticalDays"].(float64); ok {
		st.WarnCriticalDays = int(v)
	}
	if v, ok := body["receiptWarnDays"].(float64); ok {
		st.ReceiptWarnDays = int(v)
	}
	if v, ok := body["allowExpiredReceipt"].(bool); ok {
		st.AllowExpiredReceipt = v
	}
	if v, ok := body["blockExpiredOnDaf"].(bool); ok {
		st.BlockExpiredOnDAF = v
	}
	if v, ok := body["blockExpiredOnAdjustOut"].(bool); ok {
		st.BlockExpiredOnAdjustOut = v
	}
	if v, ok := body["autoQuarantineExpired"].(bool); ok {
		st.AutoQuarantineExpired = v
	}
	if v, ok := body["expiryDigestEnabled"].(bool); ok {
		st.ExpiryDigestEnabled = v
	}
	if v, ok := body["notifyOnAutoQuarantine"].(bool); ok {
		st.NotifyOnAutoQuarantine = v
	}
	st.PracticeID = id.PracticeID
	out, err := a.store.UpsertPharmacySettings(r.Context(), st)
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, out)
}

func (a *API) listPharmacyDeposits(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	if _, err := a.store.EnsureDefaultDeposit(r.Context(), id.PracticeID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	items, err := a.store.ListMedicationDeposits(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) createPharmacyDeposit(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		Name      string `json:"name"`
		Code      string `json:"code"`
		IsDefault bool   `json:"isDefault"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	d, err := a.store.CreateMedicationDeposit(r.Context(), id.PracticeID, body.Name, body.Code, body.IsDefault)
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, d)
}

func (a *API) listPharmacyBatches(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	q := r.URL.Query()
	items, err := a.store.ListMedicationBatches(r.Context(), id.PracticeID, store.ListBatchesFilter{
		DepositID: q.Get("depositId"),
		Band:      q.Get("band"),
		Status:    q.Get("status"),
		Q:         q.Get("q"),
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) receivePharmacyBatch(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		DepositID    string  `json:"depositId"`
		MedicationID string  `json:"medicationId"`
		LotNumber    string  `json:"lotNumber"`
		ExpiresOn    string  `json:"expiresOn"`
		Qty          float64 `json:"qty"`
		Unit         string  `json:"unit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	exp, err := time.Parse("2006-01-02", strings.TrimSpace(body.ExpiresOn))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	settings, err := a.store.GetPharmacySettings(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	batch, softWarn, err := a.store.ReceiveMedicationBatch(r.Context(), store.ReceiptInput{
		PracticeID:   id.PracticeID,
		DepositID:    body.DepositID,
		MedicationID: body.MedicationID,
		LotNumber:    body.LotNumber,
		ExpiresOn:    exp,
		Qty:          body.Qty,
		Unit:         body.Unit,
		CreatedBy:    id.UserID,
	}, settings, time.Now())
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]any{"batch": batch, "shortDatedWarning": softWarn})
}

func (a *API) adjustPharmacyBatch(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	batchID := chi.URLParam(r, "id")
	var body struct {
		Delta  float64 `json:"delta"`
		Detail string  `json:"detail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	settings, _ := a.store.GetPharmacySettings(r.Context(), id.PracticeID)
	batch, err := a.store.AdjustBatchQty(r.Context(), id.PracticeID, batchID, id.UserID, body.Delta, body.Detail, settings, time.Now())
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, batch)
}

func (a *API) quarantinePharmacyBatch(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	batch, err := a.store.QuarantineBatch(r.Context(), id.PracticeID, chi.URLParam(r, "id"), id.UserID, body.Reason)
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, batch)
}

func (a *API) wastePharmacyBatch(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		Reason string   `json:"reason"`
		Qty    *float64 `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if body.Reason == "" {
		body.Reason = "expired"
	}
	batch, err := a.store.WasteBatch(r.Context(), id.PracticeID, chi.URLParam(r, "id"), id.UserID, body.Reason, body.Qty)
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, batch)
}

func (a *API) getPharmacyExpirySummary(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	sum, err := a.store.ExpirySummary(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sum)
}

func (a *API) listPharmacyMovements(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.store.ListStockMovements(r.Context(), id.PracticeID, store.ListMovementsFilter{
		DafID: strings.TrimSpace(r.URL.Query().Get("dafId")),
		Limit: limit,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) exportPharmacyBatchesCSV(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	items, err := a.store.ListMedicationBatches(r.Context(), id.PracticeID, store.ListBatchesFilter{})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="pharmacy-stock.csv"`)
	_, _ = w.Write([]byte("cnk;name;lot;deposit;expires_on;qty;unit;status;band\n"))
	for _, b := range items {
		line := fmt.Sprintf("%s;%s;%s;%s;%s;%g;%s;%s;%s\n",
			store.CSVEscape(b.MedicationCNK), store.CSVEscape(b.MedicationName),
			store.CSVEscape(b.LotNumber), store.CSVEscape(b.DepositCode),
			store.CSVEscape(b.ExpiresOn), b.QtyOnHand, store.CSVEscape(b.Unit),
			store.CSVEscape(b.Status), store.CSVEscape(b.ExpiryBand))
		_, _ = w.Write([]byte(line))
	}
}

// POST /internal/pharmacy/expiry-run — auto-quarantine + optional digest emails.
func (a *API) internalPharmacyExpiryRun(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Pharmacy-Expiry-Secret", a.cfg.PharmacyExpirySecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body struct {
		ForceDigest bool `json:"forceDigest"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	now := time.Now()
	practices, err := a.store.ListPracticesWithPharmacySettings(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var quarantinedTotal, digestsSent, notifySent int
	loc, _ := time.LoadLocation("Europe/Brussels")
	if loc == nil {
		loc = time.UTC
	}
	weekday := int(now.In(loc).Weekday())
	if weekday == 0 {
		weekday = 7 // ISO Sunday=7
	}

	for _, practiceID := range practices {
		settings, err := a.store.GetPharmacySettings(r.Context(), practiceID)
		if err != nil {
			continue
		}
		var n int
		if settings.AutoQuarantineExpired {
			n, _, _ = a.store.AutoQuarantineExpiredBatches(r.Context(), practiceID, now)
			quarantinedTotal += n
			if n > 0 && settings.NotifyOnAutoQuarantine && a.notifier != nil {
				emails, _ := a.store.ListPracticeVetEmails(r.Context(), practiceID)
				subj := fmt.Sprintf("[petsFollow] %d lot(s) mis en quarantaine (péremption)", n)
				bodyTxt := fmt.Sprintf("%d lot(s) périmé(s) ont été placés en quarantaine automatiquement. Séparez-les physiquement puis déclarez un waste (destruction ou retour fournisseur).", n)
				for _, to := range emails {
					_ = a.notifier.SendVetAlert(to, subj, bodyTxt)
					notifySent++
				}
			}
		}
		if a.notifier == nil {
			continue
		}
		sum, err := a.store.ExpirySummary(r.Context(), practiceID)
		if err != nil {
			continue
		}
		if !shouldSendPharmacyExpiryDigest(settings.ExpiryDigestEnabled, body.ForceDigest, weekday, settings.ExpiryDigestWeekday, sum) {
			continue
		}
		emails, _ := a.store.ListPracticeVetEmails(r.Context(), practiceID)
		subj := "[petsFollow] Digest péremption stock cabinet"
		bodyTxt := fmt.Sprintf("Critique=%d · À retourner=%d · Attention=%d · Périmé=%d · Quarantaine=%d\nOuvrez /stock pour agir. Lots en quarantaine : séparez-les physiquement puis déclarez un waste.",
			sum.Critical, sum.Return, sum.Soon, sum.Expired, sum.Quarantine)
		for _, to := range emails {
			_ = a.notifier.SendVetAlert(to, subj, bodyTxt)
			digestsSent++
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"practices":         len(practices),
		"quarantinedBatches": quarantinedTotal,
		"notifyEmails":      notifySent,
		"digestEmails":      digestsSent,
	})
}

func (a *API) listPharmacyPrices(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	rows, err := a.store.ListMedicationPrices(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.MedicationPrice{}
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) getPharmacyPrice(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	p, err := a.store.GetMedicationPrice(r.Context(), id.PracticeID, chi.URLParam(r, "medicationId"))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, p)
}

func (a *API) putPharmacyPrice(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		PurchasePriceCents int     `json:"purchasePriceCents"`
		SellPriceCents     int     `json:"sellPriceCents"`
		VATPercent         float64 `json:"vatPercent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if body.VATPercent == 0 {
		body.VATPercent = 21
	}
	p, err := a.store.UpsertMedicationPrice(r.Context(), id.PracticeID, chi.URLParam(r, "medicationId"), id.UserID,
		body.PurchasePriceCents, body.SellPriceCents, body.VATPercent)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, p)
}

func (a *API) putPharmacyReorderThreshold(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		MedicationID string  `json:"medicationId"`
		DepositID    string  `json:"depositId"`
		MinQty       float64 `json:"minQty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	th, err := a.store.UpsertReorderThreshold(r.Context(), id.PracticeID, body.MedicationID, body.DepositID, body.MinQty)
	if err != nil {
		if a.writePharmacyErr(w, r, err) {
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, th)
}

func (a *API) listPharmacyReorderAlerts(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	rows, err := a.store.ListReorderAlerts(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.ReorderAlert{}
	}
	httpx.WriteData(w, http.StatusOK, rows)
}
