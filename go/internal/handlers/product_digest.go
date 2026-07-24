package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerProductDigestRoutes(r chi.Router) {
	r.Post("/internal/product-digest/ingest", a.internalIngestProductDigest)
	r.Post("/internal/product-digest/run", a.internalRunProductDigest)
}

func (a *API) productDigestAuthorized(r *http.Request) bool {
	secret := a.cfg.ProductDigestSecret
	return secret != "" && r.Header.Get("X-Product-Digest-Secret") == secret
}

func brusselsDate(t time.Time) time.Time {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		loc = time.FixedZone("CET", 3600)
	}
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func parseDigestDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return brusselsDate(time.Now()), nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func firstNonEmpty(m map[string]string, prefer string) string {
	if m == nil {
		return ""
	}
	if v := strings.TrimSpace(m[prefer]); v != "" {
		return v
	}
	for _, loc := range i18n.Supported {
		if v := strings.TrimSpace(m[loc]); v != "" {
			return v
		}
	}
	for _, v := range m {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type productDigestIngestBody struct {
	Date    string                       `json:"date"`
	Commits []gemini.ProductDigestCommit `json:"commits"`
}

func (a *API) internalIngestProductDigest(w http.ResponseWriter, r *http.Request) {
	if !a.productDigestAuthorized(r) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var body productDigestIngestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_body", "invalid_body")
		return
	}
	digestDate, err := parseDigestDate(body.Date)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_date", "invalid_date")
		return
	}

	commitsJSON, _ := json.Marshal(body.Commits)
	meta, _ := json.Marshal(map[string]any{"commitCount": len(body.Commits)})

	if len(body.Commits) == 0 {
		now := time.Now().UTC()
		err := a.store.UpsertProductDigest(r.Context(), store.ProductDigest{
			DigestDate:       digestDate,
			Headline:         "",
			BodyText:         "",
			HeadlineByLocale: map[string]string{},
			BodyByLocale:     map[string]string{},
			CommitsJSON:      commitsJSON,
			Status:           "empty",
			GeneratedAt:      &now,
			Meta:             meta,
		})
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status": "empty",
			"date":   digestDate.Format("2006-01-02"),
			"reason": "no_commits",
		})
		return
	}

	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
		return
	}

	summary, err := a.gemini.SummarizeProductDigest(r.Context(), body.Commits)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "gemini_failed", err.Error())
		return
	}

	now := time.Now().UTC()
	status := "ready"
	if summary.Empty {
		status = "empty"
	}
	headlineFR := firstNonEmpty(summary.Headline, "fr")
	bodyFR := firstNonEmpty(summary.Body, "fr")
	metaFull, _ := json.Marshal(map[string]any{
		"commitCount": len(body.Commits),
		"empty":       summary.Empty,
		"reason":      summary.Reason,
	})
	if err := a.store.UpsertProductDigest(r.Context(), store.ProductDigest{
		DigestDate:       digestDate,
		Headline:         headlineFR,
		BodyText:         bodyFR,
		HeadlineByLocale: summary.Headline,
		BodyByLocale:     summary.Body,
		CommitsJSON:      commitsJSON,
		Status:           status,
		GeneratedAt:      &now,
		Meta:             metaFull,
	}); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":   status,
		"date":     digestDate.Format("2006-01-02"),
		"headline": headlineFR,
		"empty":    summary.Empty,
		"reason":   summary.Reason,
	})
}

func (a *API) internalRunProductDigest(w http.ResponseWriter, r *http.Request) {
	if !a.productDigestAuthorized(r) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	dateRaw := r.URL.Query().Get("date")
	if dateRaw == "" {
		var body struct {
			Date string `json:"date"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		dateRaw = body.Date
	}
	digestDate, err := parseDigestDate(dateRaw)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_date", "invalid_date")
		return
	}

	digest, err := a.store.GetProductDigest(r.Context(), digestDate)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if digest == nil {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status": "noop",
			"date":   digestDate.Format("2006-01-02"),
			"reason": "missing_digest",
		})
		return
	}
	if digest.Status == "empty" {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status": "noop",
			"date":   digestDate.Format("2006-01-02"),
			"reason": "empty_digest",
		})
		return
	}
	if digest.Status != "ready" && digest.Status != "sent" {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":       "noop",
			"date":         digestDate.Format("2006-01-02"),
			"reason":       "not_ready",
			"digestStatus": digest.Status,
		})
		return
	}

	recipients, err := a.store.ListDigestRecipients(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	dateLabel := digestDate.Format("02/01/2006")
	sent := 0
	skipped := 0
	failed := 0
	for _, recip := range recipients {
		locale := i18n.NormalizeLocale(recip.PreferredLocale)
		headline := digest.Headline
		if h := strings.TrimSpace(digest.HeadlineByLocale[locale]); h != "" {
			headline = h
		}
		bodyText := digest.BodyText
		if b := strings.TrimSpace(digest.BodyByLocale[locale]); b != "" {
			bodyText = b
		}
		// Reserve send slot first (idempotent). On mail failure, leave the row so we do not spam retries forever;
		// count as failed for observability.
		inserted, err := a.store.RecordProductDigestSend(r.Context(), digestDate, recip.ID)
		if err != nil {
			failed++
			continue
		}
		if !inserted {
			skipped++
			continue
		}
		if err := a.notifier.SendProductDigest(recip.Email, locale, recip.FullName, dateLabel, headline, bodyText); err != nil {
			failed++
			continue
		}
		sent++
	}

	if err := a.store.MarkProductDigestSent(r.Context(), digestDate); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":   "sent",
		"date":     digestDate.Format("2006-01-02"),
		"sent":     sent,
		"skipped":  skipped,
		"failed":   failed,
		"headline": digest.Headline,
	})
}
