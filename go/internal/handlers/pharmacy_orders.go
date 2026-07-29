package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPharmacyOrderRoutes(pr chi.Router) {
	pr.Get("/vet/pharmacy/suppliers", a.listPharmacySuppliers)
	pr.Post("/vet/pharmacy/suppliers", a.upsertPharmacySupplier)
	pr.Get("/vet/pharmacy/orders", a.listPharmacyOrders)
	pr.Post("/vet/pharmacy/orders", a.createPharmacyOrder)
	pr.Get("/vet/pharmacy/orders/{id}", a.getPharmacyOrder)
	pr.Post("/vet/pharmacy/orders/{id}/send", a.sendPharmacyOrder)
	pr.Post("/vet/pharmacy/delivery-notes", a.createPharmacyDeliveryNote)
	pr.Get("/vet/pharmacy/delivery-notes/{id}", a.getPharmacyDeliveryNote)
}

func (a *API) listPharmacySuppliers(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	rows, err := a.store.ListSuppliers(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) upsertPharmacySupplier(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	su, err := a.store.UpsertSupplier(r.Context(), id.PracticeID, body.ID, body.Name, body.Email, body.Phone, body.Notes)
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
	httpx.WriteData(w, http.StatusOK, su)
}

func (a *API) listPharmacyOrders(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	rows, err := a.store.ListPurchaseOrders(r.Context(), id.PracticeID, 30)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) createPharmacyOrder(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		FromAlerts bool                           `json:"fromAlerts"`
		SupplierID string                         `json:"supplierId"`
		Notes      string                         `json:"notes"`
		Items      []store.PurchaseOrderItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	var (
		order store.PurchaseOrder
		err   error
	)
	if body.FromAlerts {
		order, err = a.store.CreatePurchaseOrderFromAlerts(r.Context(), id.PracticeID, id.UserID, body.SupplierID, body.Notes)
	} else {
		order, err = a.store.CreatePurchaseOrder(r.Context(), id.PracticeID, id.UserID, body.SupplierID, body.Notes, body.Items)
	}
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
	httpx.WriteData(w, http.StatusCreated, order)
}

func (a *API) getPharmacyOrder(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	order, err := a.store.GetPurchaseOrder(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, order)
}

func (a *API) sendPharmacyOrder(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	if a.pharmacyOrderSendRL != nil && !a.pharmacyOrderSendRL.Allow("po-send:"+id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
		return
	}
	if a.pharmacyOrderSendRL != nil && !a.pharmacyOrderSendRL.Allow("po-send-practice:"+id.PracticeID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
		return
	}
	orderID := chi.URLParam(r, "id")
	var body struct {
		ToEmail    string `json:"toEmail"`
		SupplierID string `json:"supplierId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	order, err := a.store.GetPurchaseOrder(r.Context(), id.PracticeID, orderID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if order.Status != "draft" {
		writeErr(w, r, http.StatusConflict, "order_not_draft", "order_not_draft")
		return
	}

	to, supplierID, err := a.resolvePurchaseOrderSendTo(r.Context(), id.PracticeID, order, body.ToEmail, body.SupplierID)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "supplier_email_required", "supplier_email_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.notifier == nil {
		writeErr(w, r, http.StatusBadGateway, "email_send_failed", "email_send_failed")
		return
	}

	order, err = a.store.SendPurchaseOrderLocked(r.Context(), id.PracticeID, orderID, to, supplierID, func(o store.PurchaseOrder) error {
		csv := store.FormatPurchaseOrderCSV(o)
		var rowsHTML strings.Builder
		rowsHTML.WriteString("<table border=\"1\" cellpadding=\"4\" cellspacing=\"0\"><tr><th>CNK</th><th>Médicament</th><th>Qté</th><th>Unité</th></tr>")
		for _, it := range o.Items {
			rowsHTML.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%g</td><td>%s</td></tr>",
				html.EscapeString(it.MedicationCNK), html.EscapeString(it.MedicationName), it.Qty, html.EscapeString(it.Unit)))
		}
		rowsHTML.WriteString("</table>")
		htmlBody := fmt.Sprintf("<p>Bonjour,</p><p>Commande réassort petsFollow (%d lignes).</p>%s<p>CSV en pièce jointe.</p>",
			len(o.Items), rowsHTML.String())
		return a.notifier.SendVetAlertWithCSV(to, "Commande réassort petsFollow", htmlBody, "commande-reassort.csv", []byte(csv))
	})
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "order_not_draft", "order_not_draft")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "order_send_in_progress", "order_send_in_progress")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "email_send_failed", "email_send_failed")
		return
	}
	httpx.WriteData(w, http.StatusOK, order)
}

// resolvePurchaseOrderSendTo forces toEmail to a practice supplier address (anti-spam).
// Empty toEmail → supplier email from body.SupplierID or order.SupplierID.
// When supplierID is known, only GetSupplier is used (no full list).
func (a *API) resolvePurchaseOrderSendTo(ctx context.Context, practiceID string, order store.PurchaseOrder, toEmail, supplierID string) (to, resolvedSupplierID string, err error) {
	toEmail = strings.TrimSpace(toEmail)
	supplierID = strings.TrimSpace(supplierID)
	if supplierID == "" {
		supplierID = strings.TrimSpace(order.SupplierID)
	}

	if supplierID != "" {
		su, gerr := a.store.GetSupplier(ctx, practiceID, supplierID)
		if gerr != nil {
			if errors.Is(gerr, store.ErrNotFound) {
				return "", "", store.ErrValidation
			}
			return "", "", gerr
		}
		suEmail := strings.TrimSpace(su.Email)
		if suEmail == "" {
			return "", "", store.ErrNotFound
		}
		if _, perr := mail.ParseAddress(suEmail); perr != nil {
			return "", "", store.ErrValidation
		}
		if toEmail == "" {
			return suEmail, su.ID, nil
		}
		if !strings.EqualFold(toEmail, suEmail) {
			return "", "", store.ErrValidation
		}
		return toEmail, su.ID, nil
	}

	if toEmail == "" {
		return "", "", store.ErrValidation
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return "", "", store.ErrValidation
	}
	suppliers, err := a.store.ListSuppliers(ctx, practiceID)
	if err != nil {
		return "", "", err
	}
	if len(suppliers) == 0 {
		return "", "", store.ErrNotFound
	}
	for _, su := range suppliers {
		em := strings.TrimSpace(su.Email)
		if strings.EqualFold(em, toEmail) {
			return em, su.ID, nil
		}
	}
	return "", "", store.ErrValidation
}

func (a *API) createPharmacyDeliveryNote(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.write")
	if !ok {
		return
	}
	var body struct {
		NoteNumber   string                        `json:"noteNumber"`
		SupplierID   string                        `json:"supplierId"`
		SupplierName string                        `json:"supplierName"`
		Notes        string                        `json:"notes"`
		Notify       bool                          `json:"notify"`
		Items        []store.DeliveryNoteItemInput `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	dn, err := a.store.ReceiveDeliveryNote(r.Context(), id.PracticeID, id.UserID,
		body.NoteNumber, body.SupplierID, body.SupplierName, body.Notes, body.Items, time.Now())
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "delivery_note_exists", "delivery_note_exists")
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
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if body.Notify && a.notifier != nil {
		emails, _ := a.store.ListPracticeVetEmails(r.Context(), id.PracticeID)
		subj := fmt.Sprintf("[petsFollow] BL %s — produits disponibles", dn.NoteNumber)
		bodyTxt := fmt.Sprintf("<p>Bon de livraison <strong>%s</strong> réceptionné (%d lignes).</p><p>Ouvrez /stock pour consulter les lots.</p>",
			html.EscapeString(dn.NoteNumber), len(dn.Items))
		for _, to := range emails {
			_ = a.notifier.SendVetAlert(to, subj, bodyTxt)
		}
		_ = a.store.MarkDeliveryNoteNotified(r.Context(), id.PracticeID, dn.ID)
		dn.NotifiedAt = time.Now().UTC().Format(time.RFC3339)
	}
	httpx.WriteData(w, http.StatusCreated, dn)
}

func (a *API) getPharmacyDeliveryNote(w http.ResponseWriter, r *http.Request) {
	if !a.requirePharmacyEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pharmacy.read")
	if !ok {
		return
	}
	dn, err := a.store.GetDeliveryNote(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, dn)
}
