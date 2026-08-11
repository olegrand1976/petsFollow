package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerSupportRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Use(a.requireTermsAcceptedMiddleware)
		pr.Post("/support/tickets", a.createSupportTicket)
		pr.Get("/admin/support/stats", a.adminSupportTicketStats)
		pr.Get("/admin/support/tickets", a.adminListSupportTickets)
		pr.Get("/admin/support/tickets/{id}", a.adminGetSupportTicket)
		pr.Patch("/admin/support/tickets/{id}", a.adminPatchSupportTicket)
		pr.Post("/admin/support/tickets/{id}/replies", a.adminReplySupportTicket)
		pr.Post("/admin/support/tickets/{id}/attachments", a.adminUploadSupportAttachment)
		pr.Get("/admin/support/tickets/{id}/attachments/{attachmentID}/download", a.adminDownloadSupportAttachment)
	})
}

type createSupportTicketReq struct {
	Source      string          `json:"source"`
	Subject     string          `json:"subject"`
	Message     string          `json:"message"`
	Diagnostics json.RawMessage `json:"diagnostics"`
	UserAgent   string          `json:"userAgent"`
	AppVersion  string          `json:"appVersion"`
	Locale      string          `json:"locale"`
	Route       string          `json:"route"`
}

func (a *API) createSupportTicket(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	const maxSupportTicketBodyBytes = store.MaxSupportDiagnosticsBytes + 64*1024
	r.Body = http.MaxBytesReader(w, r.Body, maxSupportTicketBodyBytes)
	var req createSupportTicketReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok || strings.Contains(err.Error(), "http: request body too large") {
			writeErr(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "payload_too_large")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if len(req.Diagnostics) > store.MaxSupportDiagnosticsBytes {
		writeErr(w, r, http.StatusRequestEntityTooLarge, "diagnostics_too_large", "diagnostics_too_large")
		return
	}

	ua := strings.TrimSpace(req.UserAgent)
	if ua == "" {
		ua = r.UserAgent()
	}
	locale := strings.TrimSpace(req.Locale)
	if locale == "" {
		locale = i18n.FromContext(r.Context())
	}

	source := store.ResolveSupportSource(id.Role, req.Source)

	ticket, err := a.store.CreateSupportTicket(r.Context(), store.CreateSupportTicketInput{
		CreatedBy:   id.UserID,
		Source:      source,
		Subject:     req.Subject,
		Message:     req.Message,
		Diagnostics: req.Diagnostics,
		UserAgent:   ua,
		AppVersion:  req.AppVersion,
		Locale:      locale,
		Route:       req.Route,
	})
	if err != nil {
		if errors.Is(err, store.ErrDiagnosticsTooLarge) {
			writeErr(w, r, http.StatusRequestEntityTooLarge, "diagnostics_too_large", "diagnostics_too_large")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	fullName := id.Email
	if u, err := a.store.GetUserByID(r.Context(), id.UserID); err == nil && strings.TrimSpace(u.FullName) != "" {
		fullName = u.FullName
	}
	a.notifySupportTicketCreated(r.Context(), ticket, fullName, id.Email, string(id.Role))

	httpx.WriteData(w, http.StatusCreated, ticket)
}

func (a *API) adminListSupportTickets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	status := r.URL.Query().Get("status")
	source := r.URL.Query().Get("source")
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := a.store.ListSupportTickets(r.Context(), status, source, q, limit, offset)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_filter")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	effectiveLimit := limit
	if effectiveLimit <= 0 || effectiveLimit > 100 {
		effectiveLimit = 50
	}
	if offset < 0 {
		offset = 0
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  effectiveLimit,
		"offset": offset,
	})
}

func (a *API) adminSupportTicketStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	stats, err := a.store.SupportTicketStats(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, stats)
}

func (a *API) adminGetSupportTicket(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	detail, err := a.store.GetSupportTicket(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

type patchSupportTicketReq struct {
	Status string `json:"status"`
}

func (a *API) adminPatchSupportTicket(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireAdminOrDev(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	var req patchSupportTicketReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	ticket, previous, err := a.store.UpdateSupportTicketStatus(r.Context(), id, req.Status, admin.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if previous != ticket.Status {
		a.notifySupportTicketStatusChanged(r.Context(), ticket, previous, ticket.Status, admin.UserID)
	}
	httpx.WriteData(w, http.StatusOK, ticket)
}

type replySupportTicketReq struct {
	Body string `json:"body"`
}

func (a *API) adminReplySupportTicket(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireAdminOrDev(w, r)
	if !ok {
		return
	}
	ticketID := chi.URLParam(r, "id")
	var req replySupportTicketReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	reply, ticket, previousStatus, err := a.store.AddSupportTicketReply(r.Context(), ticketID, admin.UserID, req.Body)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	if a.notifier != nil && ticket.CreatedBy != nil {
		u, err := a.store.GetUserByID(r.Context(), *ticket.CreatedBy)
		if err == nil && strings.TrimSpace(u.Email) != "" {
			locale := u.PreferredLocale
			if locale == "" {
				locale = "fr"
			}
			name := u.FullName
			if name == "" {
				name = u.Email
			}
			n := a.notifier
			to, subj, body, cta := u.Email, ticket.Subject, strings.TrimSpace(req.Body), a.supportCreatorCTAURL(ticket)
			go func() {
				_ = n.SendSupportTicketReply(to, locale, name, subj, body, cta)
			}()
		}
	}
	if previousStatus != ticket.Status {
		a.notifySupportTicketStatusChanged(r.Context(), ticket, previousStatus, ticket.Status, admin.UserID)
	}

	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"reply":  reply,
		"ticket": ticket,
	})
}

func (a *API) adminUploadSupportAttachment(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireAdminOrDev(w, r)
	if !ok {
		return
	}
	ticketID := chi.URLParam(r, "id")
	if _, err := a.store.GetSupportTicket(r.Context(), ticketID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	ct, size, fileName, objectKey, err := a.uploadDocumentFile(r, "support-attachments", ticketID)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	att, err := a.store.CreateSupportTicketAttachment(r.Context(), store.CreateSupportTicketAttachmentInput{
		TicketID:    ticketID,
		UploadedBy:  admin.UserID,
		FileName:    fileName,
		ContentType: ct,
		SizeBytes:   size,
		ObjectKey:   objectKey,
	})
	if err != nil {
		if a.media != nil && objectKey != "" {
			_ = a.media.Delete(r.Context(), objectKey)
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, att)
}

func (a *API) adminDownloadSupportAttachment(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	ticketID := chi.URLParam(r, "id")
	attID := chi.URLParam(r, "attachmentID")
	att, err := a.store.GetSupportTicketAttachment(r.Context(), attID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if att.TicketID != ticketID {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	key := strings.TrimSpace(att.ObjectKey)
	if key == "" || a.media == nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	rc, contentType, err := a.media.Open(r.Context(), key)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	defer rc.Close()
	if strings.TrimSpace(att.ContentType) != "" {
		contentType = att.ContentType
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	stem := strings.TrimSuffix(strings.TrimSpace(att.FileName), path.Ext(att.FileName))
	if strings.TrimSpace(stem) == "" {
		stem = "attachment"
	}
	ext, extErr := media.ExtForDocument(att.ContentType)
	if extErr != nil {
		ext = ""
	}
	filename := sanitizeFilename(stem) + ext
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = io.Copy(w, rc)
}

func supportStatusLabel(locale, status string) string {
	key := "emails.support_status_" + strings.TrimSpace(status)
	label := i18n.T(locale, key, nil)
	if label == key {
		return i18n.T(locale, "emails.support_status_unknown", nil)
	}
	return label
}

func (a *API) supportAdminURL(ticketID string) string {
	return strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/admin/support/" + ticketID
}

func (a *API) supportCreatorCTAURL(ticket store.SupportTicket) string {
	ctaURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	switch ticket.Source {
	case store.SupportSourceFlutterClient, store.SupportSourceFlutterProLight:
		if dl := strings.TrimSpace(a.cfg.PetsAppDownloadURL); dl != "" {
			return dl
		}
	}
	return ctaURL
}

// supportOpsRecipients returns SUPPORT_INBOX_EMAIL ∪ users with an admin[/dev]
// profile (demo *@petsfollow.test skipped by the store). Deduped by email.
func (a *API) supportOpsRecipients(ctx context.Context, includeDev bool) []store.SupportOpsRecipient {
	seen := map[string]struct{}{}
	out := []store.SupportOpsRecipient{}
	add := func(email, name, locale, role string) {
		email = strings.TrimSpace(email)
		if email == "" {
			return
		}
		key := strings.ToLower(email)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		if strings.TrimSpace(locale) == "" {
			locale = "fr"
		}
		if strings.TrimSpace(name) == "" {
			name = email
		}
		out = append(out, store.SupportOpsRecipient{
			Email:  email,
			Name:   name,
			Locale: locale,
			Role:   role,
		})
	}

	add(a.cfg.SupportInboxEmail, "Support", "fr", string(kernel.RoleAdmin))

	recs, err := a.store.ListSupportOpsRecipients(ctx)
	if err != nil {
		log.Printf("support ops recipients: %v", err)
		return out
	}
	for _, r := range recs {
		role := strings.TrimSpace(r.Role)
		if role == string(kernel.RoleAdmin) || (includeDev && role == string(kernel.RoleDev)) {
			add(r.Email, r.Name, r.Locale, role)
		}
	}
	return out
}

func (a *API) notifySupportTicketCreated(ctx context.Context, ticket store.SupportTicket, fullName, emailAddr, role string) {
	if a.notifier == nil {
		return
	}
	n := a.notifier
	adminURL := a.supportAdminURL(ticket.ID)
	creatorCTA := a.supportCreatorCTAURL(ticket)
	msgPreview := ticket.Message
	if len([]rune(msgPreview)) > 2000 {
		msgPreview = string([]rune(msgPreview)[:2000]) + "…"
	}
	ops := a.supportOpsRecipients(ctx, true)
	subject := ticket.Subject
	ticketID := ticket.ID
	source := ticket.Source
	creatorEmail := strings.TrimSpace(emailAddr)
	locale := strings.TrimSpace(ticket.Locale)
	if locale == "" {
		locale = "fr"
	}
	name := strings.TrimSpace(fullName)
	if name == "" {
		name = creatorEmail
	}
	// Soft-fail ops/ack async — never block ticket create on SMTP fan-out.
	go func() {
		for _, r := range ops {
			_ = n.SendSupportTicketOps(
				r.Email, r.Locale, ticketID, subject,
				fullName, emailAddr, role, source, msgPreview, adminURL,
			)
		}
		if creatorEmail == "" {
			return
		}
		_ = n.SendSupportTicketCreatedAck(
			creatorEmail, locale, name, subject, ticketID, creatorCTA,
		)
	}()
}

func (a *API) notifySupportTicketStatusChanged(ctx context.Context, ticket store.SupportTicket, fromStatus, toStatus, changedByUserID string) {
	if a.notifier == nil || fromStatus == toStatus {
		return
	}
	n := a.notifier
	changedByName := changedByUserID
	if u, err := a.store.GetUserByID(ctx, changedByUserID); err == nil {
		if strings.TrimSpace(u.FullName) != "" {
			changedByName = u.FullName
		} else if strings.TrimSpace(u.Email) != "" {
			changedByName = u.Email
		}
	}

	adminURL := a.supportAdminURL(ticket.ID)
	creatorCTA := a.supportCreatorCTAURL(ticket)
	subject := ticket.Subject
	ticketID := ticket.ID
	ops := a.supportOpsRecipients(ctx, false)

	type dest struct {
		email, locale, name, detailURL string
	}
	seen := map[string]struct{}{}
	dests := make([]dest, 0, len(ops)+1)
	for _, r := range ops {
		key := strings.ToLower(r.Email)
		seen[key] = struct{}{}
		dests = append(dests, dest{email: r.Email, locale: r.Locale, name: r.Name, detailURL: adminURL})
	}
	if ticket.CreatedBy != nil {
		if u, err := a.store.GetUserByID(ctx, *ticket.CreatedBy); err == nil && strings.TrimSpace(u.Email) != "" {
			key := strings.ToLower(strings.TrimSpace(u.Email))
			if _, ok := seen[key]; !ok {
				locale := u.PreferredLocale
				if locale == "" {
					locale = "fr"
				}
				name := u.FullName
				if name == "" {
					name = u.Email
				}
				dests = append(dests, dest{email: u.Email, locale: locale, name: name, detailURL: creatorCTA})
			}
		}
	}
	if len(dests) == 0 {
		return
	}

	go func() {
		for _, d := range dests {
			_ = n.SendSupportTicketStatusChanged(
				d.email, d.locale, d.name, subject, ticketID,
				supportStatusLabel(d.locale, fromStatus), supportStatusLabel(d.locale, toStatus),
				changedByName, d.detailURL,
			)
		}
	}()
}
