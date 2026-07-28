package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
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
	case errors.Is(err, invoicing.ErrLinesRequired):
		writeErr(w, r, http.StatusBadRequest, "lines_required", "lines_required")
	case errors.Is(err, invoicing.ErrDocsQuotaExceeded):
		writeErr(w, r, http.StatusConflict, "docs_quota_exceeded", "docs_quota_exceeded")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}
