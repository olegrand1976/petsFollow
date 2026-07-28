package handlers

import (
	"errors"
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
		pr.Post("/admin/invoicing/connections/{practiceId}/mark-partner-invoiced", a.adminInvoicingMarkPartner)
		pr.Post("/admin/invoicing/connections/{practiceId}/saas-draft", a.adminInvoicingSaasDraft)
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
	httpx.WriteData(w, http.StatusOK, doc)
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
	doc, err := a.invoicing.CreateSaasDraft(r.Context(), practiceID, id.UserID)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
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
	case errors.Is(err, invoicing.ErrOrderPersistFailed):
		writeErr(w, r, http.StatusConflict, "order_persist_failed", "order_persist_failed")
	case errors.Is(err, invoicing.ErrLinesRequired):
		writeErr(w, r, http.StatusBadRequest, "lines_required", "lines_required")
	case errors.Is(err, invoicing.ErrInvalidLines):
		writeErr(w, r, http.StatusBadRequest, "invalid_lines", "invalid_lines")
	case errors.Is(err, invoicing.ErrInvalidCounterparty):
		writeErr(w, r, http.StatusBadRequest, "invalid_counterparty", "invalid_counterparty")
	case errors.Is(err, invoicing.ErrDocsQuotaExceeded):
		writeErr(w, r, http.StatusConflict, "docs_quota_exceeded", "docs_quota_exceeded")
	case errors.Is(err, invoicing.ErrPartnerNotEligible):
		writeErr(w, r, http.StatusConflict, "partner_mark_not_eligible", "partner_mark_not_eligible")
	case errors.Is(err, invoicing.ErrMasterNotConfigured):
		writeErr(w, r, http.StatusServiceUnavailable, "saas_master_not_configured", "saas_master_not_configured")
	case errors.Is(err, invoicing.ErrGateway):
		writeErr(w, r, http.StatusBadGateway, "bad_gateway", "invoicing_gateway_error")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}
