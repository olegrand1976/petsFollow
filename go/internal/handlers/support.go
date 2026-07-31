package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
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
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) || strings.Contains(err.Error(), "http: request body too large") {
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

	since := time.Now().UTC().Add(-time.Hour)
	n, err := a.store.CountRecentSupportTickets(r.Context(), id.UserID, since)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if n >= store.SupportTicketsPerHour {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
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

	if a.notifier != nil {
		to := strings.TrimSpace(a.cfg.SupportInboxEmail)
		adminURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/admin/support/" + ticket.ID
		msgPreview := ticket.Message
		if len([]rune(msgPreview)) > 2000 {
			msgPreview = string([]rune(msgPreview)[:2000]) + "…"
		}
		fullName := id.Email
		if u, err := a.store.GetUserByID(r.Context(), id.UserID); err == nil && strings.TrimSpace(u.FullName) != "" {
			fullName = u.FullName
		}
		_ = a.notifier.SendSupportTicketOps(
			to, "fr", ticket.ID, ticket.Subject,
			fullName, id.Email, string(id.Role), ticket.Source, msgPreview, adminURL,
		)
	}

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
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	var req patchSupportTicketReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	ticket, err := a.store.UpdateSupportTicketStatus(r.Context(), id, req.Status)
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
	reply, ticket, err := a.store.AddSupportTicketReply(r.Context(), ticketID, admin.UserID, req.Body)
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
			ctaURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
			switch ticket.Source {
			case store.SupportSourceFlutterClient, store.SupportSourceFlutterProLight:
				if dl := strings.TrimSpace(a.cfg.PetsAppDownloadURL); dl != "" {
					ctaURL = dl
				}
			}
			_ = a.notifier.SendSupportTicketReply(u.Email, locale, name, ticket.Subject, strings.TrimSpace(req.Body), ctaURL)
		}
	}

	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"reply":  reply,
		"ticket": ticket,
	})
}
