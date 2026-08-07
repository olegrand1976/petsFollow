package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerInvoicingRoutes(r chi.Router) {
	if a.invoicing == nil || !a.cfg.BillitEnabled {
		return
	}
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/practices/me/invoicing/connection", a.invoicingGetConnection)
		pr.Post("/practices/me/invoicing/connect/start", a.invoicingConnectStart)
		pr.Post("/practices/me/invoicing/connect/complete", a.invoicingConnectComplete)
		pr.Post("/practices/me/invoicing/connect/refresh", a.invoicingConnectRefresh)
		pr.Get("/practices/me/invoicing/documents", a.invoicingListDocuments)
		pr.Post("/practices/me/invoicing/documents", a.invoicingCreateDocument)
		pr.Get("/practices/me/invoicing/documents/{id}", a.invoicingGetDocument)
		pr.Post("/practices/me/invoicing/documents/{id}/send", a.invoicingSendDocument)
		pr.Get("/admin/invoicing/connections", a.adminInvoicingConnections)
		pr.Get("/admin/invoicing/saas-targets", a.adminInvoicingSaasTargets)
		pr.Post("/admin/invoicing/connections/{practiceId}/mark-partner-invoiced", a.adminInvoicingMarkPartner)
		pr.Post("/admin/invoicing/practices/{practiceId}/saas-billing", a.adminInvoicingSaasBilling)
		pr.Post("/admin/invoicing/connections/{practiceId}/saas-draft", a.adminInvoicingSaasDraft)
		pr.Post("/admin/invoicing/connections/{practiceId}/saas-documents/{docId}/send", a.adminInvoicingSaasSend)
	})
}

func (a *API) registerInvoicingPublicRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	if a.invoicing == nil || !a.cfg.BillitEnabled {
		return
	}
	r.Group(func(pr chi.Router) {
		if rateLimit != nil {
			pr.Use(rateLimit)
		}
		pr.Get("/public/proforma/{token}", a.getPublicProforma)
		pr.Post("/public/proforma/{token}/accept", a.postPublicProformaAccept)
	})
}

func (a *API) invoicingGetConnection(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAnyPracticePerm(w, r, "clients.write", "practice.settings")
	if !ok {
		return
	}
	c, err := a.invoicing.GetConnection(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// PartyID réservé aux gestionnaires Billit (practice.settings) — pas aux seuls clients.write.
	if !a.allowPracticePerm(r, id, "practice.settings") {
		c.BillitPartyID = ""
	}
	httpx.WriteData(w, http.StatusOK, c)
}

func (a *API) invoicingConnectStart(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	res, err := a.invoicing.ConnectStart(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, res)
}

func (a *API) invoicingConnectComplete(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	var req invoicing.ConnectCompleteInput
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_body")
		return
	}
	c, err := a.invoicing.ConnectComplete(r.Context(), id.PracticeID, req)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, c)
}

func (a *API) invoicingConnectRefresh(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	c, err := a.invoicing.ConnectRefresh(r.Context(), id.PracticeID)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, c)
}

func (a *API) invoicingListDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	items, err := a.invoicing.ListDocuments(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) invoicingCreateDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	var req invoicing.CreateDocumentInput
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_body")
		return
	}
	req.VisitID = strings.TrimSpace(req.VisitID)
	var visit store.Visit
	if req.VisitID != "" {
		var err error
		visit, err = a.store.GetVisit(r.Context(), req.VisitID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if visit.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusBadRequest, "visit_mismatch", "visit_mismatch")
			return
		}
		if err := a.store.AssertVisitIdentified(r.Context(), req.VisitID); err != nil {
			if a.writeWalkinErr(w, r, err) {
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	req.DAFID = strings.TrimSpace(req.DAFID)
	if req.DAFID != "" {
		daf, err := a.store.GetDAF(r.Context(), id.PracticeID, req.DAFID)
		if err != nil {
			if errors.Is(err, pharmacy.ErrDAFNotFound) {
				writeErr(w, r, http.StatusBadRequest, "daf_mismatch", "daf_mismatch")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if daf.Status != "finalized" {
			writeErr(w, r, http.StatusBadRequest, "daf_not_finalized", "daf_not_finalized")
			return
		}
		if req.VisitID != "" {
			if daf.VisitID == "" || daf.VisitID != req.VisitID {
				writeErr(w, r, http.StatusBadRequest, "daf_visit_mismatch", "daf_visit_mismatch")
				return
			}
			if daf.PetID != "" && visit.PetID != "" && daf.PetID != visit.PetID {
				writeErr(w, r, http.StatusBadRequest, "daf_visit_mismatch", "daf_visit_mismatch")
				return
			}
		}
	}
	doc, err := a.invoicing.CreateDocument(r.Context(), id.PracticeID, id.UserID, req)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) invoicingGetDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	doc, err := a.invoicing.GetDocument(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) invoicingSendDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	doc, err := a.invoicing.SendDocument(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	out := map[string]any{
		"id":             doc.ID,
		"practiceId":     doc.PracticeID,
		"type":           doc.Type,
		"status":         doc.Status,
		"source":         doc.Source,
		"number":         doc.Number,
		"billitOrderId":  doc.BillitOrderID,
		"idempotencyKey": doc.IdempotencyKey,
		"counterparty":   doc.Counterparty,
		"currency":       doc.Currency,
		"totalExclCents": doc.TotalExclCents,
		"totalVatCents":  doc.TotalVATCents,
		"totalInclCents": doc.TotalInclCents,
		"relatedDocumentId": doc.RelatedDocumentID,
		"visitId":        doc.VisitID,
		"dafId":          doc.DAFID,
		"peppolStatus":   doc.PeppolStatus,
		"sentAt":         doc.SentAt,
		"acceptedAt":     doc.AcceptedAt,
		"tokenExpiresAt": doc.TokenExpiresAt,
		"createdBy":      doc.CreatedBy,
		"createdAt":      doc.CreatedAt,
		"updatedAt":      doc.UpdatedAt,
		"lines":          doc.Lines,
	}
	if doc.Type == invoicing.DocProforma && doc.PublicToken != "" {
		cta := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/proforma/" + doc.PublicToken
		a.afterProformaIssued(r.Context(), id.UserID, id.PracticeID, doc, cta)
		if a.cfg.DevSeedEnabled {
			out["acceptPath"] = "/proforma/" + doc.PublicToken
		}
	}
	httpx.WriteData(w, http.StatusOK, out)
}

func (a *API) getPublicProforma(w http.ResponseWriter, r *http.Request) {
	view, err := a.invoicing.GetPublicProforma(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, view)
}

func (a *API) postPublicProformaAccept(w http.ResponseWriter, r *http.Request) {
	res, err := a.invoicing.AcceptProforma(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	// Minimal public payload — no practice internals / Billit ids.
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"proforma": map[string]any{
			"id":     res.Proforma.ID,
			"status": res.Proforma.Status,
		},
		"invoice": map[string]any{
			"id":             res.Invoice.ID,
			"status":         res.Invoice.Status,
			"totalInclCents": res.Invoice.TotalInclCents,
			"currency":       res.Invoice.Currency,
		},
	})
}

func (a *API) afterProformaIssued(ctx context.Context, authorUserID, practiceID string, doc invoicing.Document, ctaURL string) {
	if a.notifier != nil {
		locale := "fr"
		practiceName := ""
		if party, _, err := a.store.GetPracticeInvoicingProfile(ctx, practiceID); err == nil {
			practiceName = party.LegalName
		}
		to := strings.TrimSpace(doc.Counterparty.Email)
		if to != "" {
			if err := a.notifier.SendProformaForValidation(
				to, locale, doc.Counterparty.Name, practiceName,
				formatMoneyEUR(doc.TotalInclCents), ctaURL,
			); err != nil {
				fmt.Printf("proforma email doc %s: %v\n", doc.ID, err)
			}
		}
	}
	a.recordProformaTimelineEvent(ctx, authorUserID, doc, ctaURL)
}

func formatMoneyEUR(cents int64) string {
	return fmt.Sprintf("%.2f €", float64(cents)/100.0)
}

func (a *API) recordProformaTimelineEvent(ctx context.Context, authorUserID string, doc invoicing.Document, ctaURL string) {
	petID := ""
	if doc.VisitID != "" {
		if v, err := a.store.GetVisit(ctx, doc.VisitID); err == nil {
			petID = v.PetID
		}
	}
	if petID == "" && doc.DAFID != "" {
		if d, err := a.store.GetDAF(ctx, doc.PracticeID, doc.DAFID); err == nil {
			petID = d.PetID
		}
	}
	if petID == "" || authorUserID == "" {
		return
	}
	title := "Proforma à valider"
	body := fmt.Sprintf("Devis %s — validez pour facturation.", formatMoneyEUR(doc.TotalInclCents))
	meta := map[string]any{
		"kind":       "proforma_pending",
		"documentId": doc.ID,
		"url":        ctaURL,
		"cta":        "Valider",
	}
	_ = a.store.InsertDossierEventWithMeta(ctx, petID, authorUserID, title, body, meta)
}

func (a *API) adminInvoicingConnections(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	items, err := a.invoicing.ListAdminConnections(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) adminInvoicingSaasTargets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	items, err := a.invoicing.ListSaasTargets(r.Context(), 200, 0, false)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) adminInvoicingMarkPartner(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceId")
	if err := a.invoicing.MarkPartnerListed(r.Context(), practiceID); err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) adminInvoicingSaasDraft(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceId")
	if !isUUID(practiceID) {
		writeErr(w, r, http.StatusBadRequest, "invalid_id", "invalid_id")
		return
	}
	doc, err := a.invoicing.CreateSaasDraft(r.Context(), practiceID, id.UserID)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) adminInvoicingSaasBilling(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceId")
	if !isUUID(practiceID) {
		writeErr(w, r, http.StatusBadRequest, "invalid_id", "invalid_id")
		return
	}
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Enabled == nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "enabled_required")
		return
	}
	if err := a.invoicing.SetSaasBillingEnabled(r.Context(), practiceID, *body.Enabled); err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"practiceId": practiceID, "saasBillingEnabled": *body.Enabled})
}

func (a *API) adminInvoicingSaasSend(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	practiceID := chi.URLParam(r, "practiceId")
	docID := chi.URLParam(r, "docId")
	if !isUUID(practiceID) || !isUUID(docID) {
		writeErr(w, r, http.StatusBadRequest, "invalid_id", "invalid_id")
		return
	}
	doc, err := a.invoicing.SendSaasDocument(r.Context(), practiceID, docID)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) writeInvoicingErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, invoicing.ErrDisabled):
		writeErr(w, r, http.StatusNotFound, "not_found", "invoicing_disabled")
	case errors.Is(err, invoicing.ErrProfileIncomplete):
		writeErr(w, r, http.StatusUnprocessableEntity, "practice_profile_incomplete", "practice_profile_incomplete")
	case errors.Is(err, invoicing.ErrInvalidState):
		writeErr(w, r, http.StatusBadRequest, "invalid_connect_state", "invalid_connect_state")
	case errors.Is(err, invoicing.ErrNotConnected):
		writeErr(w, r, http.StatusConflict, "invoicing_not_connected", "invoicing_not_connected")
	case errors.Is(err, invoicing.ErrConnectionNotActive):
		writeErr(w, r, http.StatusConflict, "connection_not_active", "connection_not_active")
	case errors.Is(err, invoicing.ErrDocNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "document_not_found")
	case errors.Is(err, invoicing.ErrProformaTokenInvalid):
		writeErr(w, r, http.StatusNotFound, "not_found", "proforma_token_invalid")
	case errors.Is(err, invoicing.ErrProformaExpired):
		writeErr(w, r, http.StatusGone, "gone", "proforma_expired")
	case errors.Is(err, invoicing.ErrProformaAlreadyAccepted):
		writeErr(w, r, http.StatusConflict, "conflict", "proforma_already_accepted")
	case errors.Is(err, invoicing.ErrDocNotDraft):
		writeErr(w, r, http.StatusConflict, "document_not_draft", "document_not_draft")
	case errors.Is(err, invoicing.ErrSendInProgress):
		writeErr(w, r, http.StatusConflict, "document_send_in_progress", "document_send_in_progress")
	case errors.Is(err, invoicing.ErrPeppolNotForType):
		writeErr(w, r, http.StatusBadRequest, "peppol_not_allowed", "peppol_not_allowed")
	case errors.Is(err, invoicing.ErrInvalidType):
		writeErr(w, r, http.StatusBadRequest, "invalid_type", "invalid_type")
	case errors.Is(err, invoicing.ErrRelatedDocument):
		writeErr(w, r, http.StatusBadRequest, "related_document_invalid", "related_document_invalid")
	case errors.Is(err, invoicing.ErrRelatedInvoiceNotOnBillit):
		writeErr(w, r, http.StatusBadRequest, "related_invoice_not_on_billit", "related_invoice_not_on_billit")
	case errors.Is(err, invoicing.ErrOrderPersistFailed):
		writeErr(w, r, http.StatusConflict, "order_persist_failed", "order_persist_failed")
	case errors.Is(err, invoicing.ErrLinesRequired):
		writeErr(w, r, http.StatusBadRequest, "lines_required", "lines_required")
	case errors.Is(err, invoicing.ErrInvalidLines):
		writeErr(w, r, http.StatusBadRequest, "invalid_lines", "invalid_lines")
	case errors.Is(err, invoicing.ErrInvalidCounterparty):
		reason := strings.TrimPrefix(err.Error(), invoicing.ErrInvalidCounterparty.Error())
		reason = strings.TrimPrefix(reason, ": ")
		reason = strings.TrimSpace(reason)
		msgKey := "invalid_counterparty"
		if reason != "" {
			msgKey = reason
		}
		var details any
		if reason != "" {
			details = map[string]string{"reason": reason}
		}
		writeErrDetails(w, r, http.StatusBadRequest, "invalid_counterparty", msgKey, details)
	case errors.Is(err, invoicing.ErrDocsQuotaExceeded):
		writeErr(w, r, http.StatusConflict, "docs_quota_exceeded", "docs_quota_exceeded")
	case errors.Is(err, invoicing.ErrPartnerNotEligible):
		writeErr(w, r, http.StatusConflict, "partner_mark_not_eligible", "partner_mark_not_eligible")
	case errors.Is(err, invoicing.ErrMasterNotConfigured):
		writeErr(w, r, http.StatusServiceUnavailable, "saas_master_not_configured", "saas_master_not_configured")
	case errors.Is(err, invoicing.ErrSaasNotEligible):
		writeErr(w, r, http.StatusConflict, "saas_not_eligible", "saas_not_eligible")
	case errors.Is(err, invoicing.ErrSaasBillingDisabled):
		writeErr(w, r, http.StatusConflict, "saas_billing_disabled", "saas_billing_disabled")
	case errors.Is(err, invoicing.ErrSecrets):
		writeErr(w, r, http.StatusConflict, "invoicing_secrets_mismatch", "invoicing_secrets_mismatch")
	case errors.Is(err, invoicing.ErrGateway):
		detail := strings.TrimPrefix(err.Error(), invoicing.ErrGateway.Error())
		detail = strings.TrimPrefix(detail, ": ")
		detail = strings.TrimSpace(detail)
		if len(detail) > 200 {
			detail = detail[:200]
		}
		var details any
		if detail != "" {
			details = map[string]string{"gateway": detail}
		}
		writeErrDetails(w, r, http.StatusBadGateway, "bad_gateway", "invoicing_gateway_error", details)
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}
