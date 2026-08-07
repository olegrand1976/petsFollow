package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerProductDigestRoutes(r chi.Router) {
	r.Post("/internal/product-digest/ingest", a.internalIngestProductDigest)
	r.Post("/internal/product-digest/run", a.internalRunProductDigest)
	r.Post("/internal/product-digest/weekly-run", a.internalRunProductDigestWeekly)
}

func (a *API) registerProductDigestAuthedRoutes(r chi.Router) {
	r.Get("/product-digests", a.listProductDigests)
}

func (a *API) productDigestAuthorized(r *http.Request) bool {
	return secretHeaderOK(r, "X-Product-Digest-Secret", a.cfg.ProductDigestSecret)
}

func canViewProductDigests(role kernel.Role) bool {
	switch role {
	case kernel.RoleVet, kernel.RoleVetAssistant, kernel.RoleSecretary,
		kernel.RoleAdmin, kernel.RoleDev,
		kernel.RoleCommercial, kernel.RoleCommercialManager:
		return true
	default:
		return false
	}
}

func brusselsLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		return time.FixedZone("CET", 3600)
	}
	return loc
}

func brusselsDate(t time.Time) time.Time {
	local := t.In(brusselsLocation())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

// isoWeekStartMonday returns the Monday (UTC date-only) of the ISO week containing t.
func isoWeekStartMonday(t time.Time) time.Time {
	local := t.In(brusselsLocation())
	day := brusselsDate(local)
	// Go Weekday: Sunday=0 … Saturday=6. ISO Monday=1.
	wd := int(local.Weekday())
	if wd == 0 {
		wd = 7
	}
	return day.AddDate(0, 0, -(wd - 1))
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
	Date        string                       `json:"date"`
	Branch      string                       `json:"branch"`
	Environment string                       `json:"environment"`
	Commits     []gemini.ProductDigestCommit `json:"commits"`
}

// digestDisplayLabel prefers deploy environment (staging/production), then git branch, then APP_ENV, else "local".
func digestDisplayLabel(branch, environment, appEnvFallback string) string {
	if e := strings.TrimSpace(environment); e != "" {
		return e
	}
	if b := strings.TrimSpace(branch); b != "" {
		return b
	}
	if e := strings.TrimSpace(appEnvFallback); e != "" {
		return e
	}
	return "local"
}

func isTestDigestEnv(label string) bool {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "staging", "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}

// digestAudienceLabel appends the localized TEST-environment tag for non-prod labels.
func digestAudienceLabel(rawLabel, locale string) string {
	label := strings.TrimSpace(rawLabel)
	if label == "" {
		label = "local"
	}
	if !isTestDigestEnv(label) {
		return label
	}
	tag := strings.TrimSpace(i18n.T(locale, "emails.product_digest_test_env_tag", nil))
	if tag == "" {
		tag = "environnement de TEST"
	}
	return label + " — " + tag
}

func digestMetaBranch(meta json.RawMessage, appEnvFallback string) string {
	if len(meta) == 0 {
		return digestDisplayLabel("", "", appEnvFallback)
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		return digestDisplayLabel("", "", appEnvFallback)
	}
	branch, _ := m["branch"].(string)
	env, _ := m["environment"].(string)
	return digestDisplayLabel(branch, env, appEnvFallback)
}

func (a *API) listProductDigests(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if !canViewProductDigests(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	limit := 30
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	items, err := a.store.ListProductDigests(r.Context(), limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	locale := i18n.NormalizeLocale(r.Header.Get("Accept-Language"))
	if u, err := a.store.GetUserByID(r.Context(), id.UserID); err == nil {
		locale = i18n.NormalizeLocale(u.PreferredLocale)
	}
	out := make([]map[string]any, 0, len(items))
	for _, d := range items {
		headline := d.Headline
		if h := strings.TrimSpace(d.HeadlineByLocale[locale]); h != "" {
			headline = h
		}
		bodyText := d.BodyText
		if b := strings.TrimSpace(d.BodyByLocale[locale]); b != "" {
			bodyText = b
		}
		out = append(out, map[string]any{
			"date":     d.DigestDate.Format("2006-01-02"),
			"headline": headline,
			"body":     bodyText,
			"status":   d.Status,
			"branch":   digestAudienceLabel(digestMetaBranch(d.Meta, os.Getenv("APP_ENV")), locale),
		})
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": out})
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

	branch := strings.TrimSpace(body.Branch)
	environment := strings.TrimSpace(body.Environment)
	source := digestDisplayLabel(branch, environment, os.Getenv("APP_ENV"))
	commitsJSON, _ := json.Marshal(body.Commits)
	meta, _ := json.Marshal(map[string]any{
		"commitCount": len(body.Commits),
		"branch":      branch,
		"environment": environment,
		"source":      source,
	})

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
			"branch": source,
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
		"branch":      branch,
		"environment": environment,
		"source":      source,
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
		"branch":   source,
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
	rawBranch := digestMetaBranch(digest.Meta, os.Getenv("APP_ENV"))
	sent := 0
	skipped := 0
	failed := 0
	var lastAudience string
	for _, recip := range recipients {
		locale := i18n.NormalizeLocale(recip.PreferredLocale)
		audience := digestAudienceLabel(rawBranch, locale)
		lastAudience = audience
		headline := digest.Headline
		if h := strings.TrimSpace(digest.HeadlineByLocale[locale]); h != "" {
			headline = h
		}
		bodyText := digest.BodyText
		if b := strings.TrimSpace(digest.BodyByLocale[locale]); b != "" {
			bodyText = b
		}
		// Reserve send slot first (idempotent). On SMTP failure, clear the row so a later run can retry.
		inserted, err := a.store.RecordProductDigestSend(r.Context(), digestDate, recip.ID)
		if err != nil {
			failed++
			continue
		}
		if !inserted {
			skipped++
			continue
		}
		if err := a.notifier.SendProductDigest(recip.Email, locale, recip.FullName, dateLabel, audience, headline, bodyText); err != nil {
			_ = a.store.ClearProductDigestSend(r.Context(), digestDate, recip.ID)
			failed++
			continue
		}
		sent++
	}

	status := "partial"
	if sent > 0 && failed == 0 {
		status = "sent"
		if err := a.store.MarkProductDigestSent(r.Context(), digestDate); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	} else if sent == 0 && failed > 0 {
		status = "failed"
	} else if sent == 0 && skipped > 0 && failed == 0 {
		// All recipients already recorded from a prior successful send.
		status = "sent"
		_ = a.store.MarkProductDigestSent(r.Context(), digestDate)
	}

	if lastAudience == "" {
		lastAudience = digestAudienceLabel(rawBranch, "fr")
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":   status,
		"date":     digestDate.Format("2006-01-02"),
		"sent":     sent,
		"skipped":  skipped,
		"failed":   failed,
		"headline": digest.Headline,
		"branch":   lastAudience,
	})
}

func localizeDigestDay(d store.ProductDigest, locale string) (headline, body string) {
	headline = d.Headline
	if h := strings.TrimSpace(d.HeadlineByLocale[locale]); h != "" {
		headline = h
	}
	body = d.BodyText
	if b := strings.TrimSpace(d.BodyByLocale[locale]); b != "" {
		body = b
	}
	return headline, body
}

func buildWeeklyRollup(digests []store.ProductDigest, locale string) (headline, body string) {
	var parts []string
	for _, d := range digests {
		h, b := localizeDigestDay(d, locale)
		dayLabel := d.DigestDate.Format("02/01/2006")
		block := "• " + dayLabel
		if h != "" {
			block += " — " + h
		}
		if b != "" {
			block += "\n" + b
		}
		parts = append(parts, block)
	}
	body = strings.Join(parts, "\n\n")
	if len(digests) == 1 {
		h, _ := localizeDigestDay(digests[0], locale)
		headline = h
	} else {
		headline = ""
	}
	return headline, body
}

func (a *API) internalRunProductDigestWeekly(w http.ResponseWriter, r *http.Request) {
	if !a.productDigestAuthorized(r) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	today := brusselsDate(time.Now())
	var body struct {
		Date string `json:"date"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if raw := strings.TrimSpace(r.URL.Query().Get("date")); raw != "" {
		body.Date = raw
	}
	if body.Date != "" {
		if d, err := parseDigestDate(body.Date); err == nil {
			today = d
		} else {
			writeErr(w, r, http.StatusBadRequest, "invalid_date", "invalid_date")
			return
		}
	}

	weekStart := isoWeekStartMonday(today)
	// Rollup = ISO week Monday → today (matches weekStart idempotence key).
	since := weekStart

	digests, err := a.store.ListProductDigestsSince(r.Context(), since)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(digests) == 0 {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":    "noop",
			"weekStart": weekStart.Format("2006-01-02"),
			"reason":    "empty_week",
		})
		return
	}

	recipients, err := a.store.ListWeeklyDigestRecipients(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(recipients) == 0 {
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":    "noop",
			"weekStart": weekStart.Format("2006-01-02"),
			"reason":    "no_recipients",
			"days":      len(digests),
		})
		return
	}

	rawBranch := digestDisplayLabel("", "", os.Getenv("APP_ENV"))
	if len(digests) > 0 {
		rawBranch = digestMetaBranch(digests[len(digests)-1].Meta, os.Getenv("APP_ENV"))
	}
	weekLabel := weekStart.Format("02/01/2006")
	rangeLabel := since.Format("02/01/2006") + " – " + today.Format("02/01/2006")

	sent := 0
	skipped := 0
	failed := 0
	var lastAudience string
	for _, recip := range recipients {
		locale := i18n.NormalizeLocale(recip.PreferredLocale)
		audience := digestAudienceLabel(rawBranch, locale)
		lastAudience = audience
		headline, bodyText := buildWeeklyRollup(digests, locale)
		inserted, err := a.store.RecordProductDigestWeeklySend(r.Context(), weekStart, recip.ID)
		if err != nil {
			failed++
			continue
		}
		if !inserted {
			skipped++
			continue
		}
		if err := a.notifier.SendProductDigestWeekly(
			recip.Email, locale, recip.FullName, weekLabel, rangeLabel, audience, headline, bodyText,
		); err != nil {
			_ = a.store.ClearProductDigestWeeklySend(r.Context(), weekStart, recip.ID)
			failed++
			continue
		}
		sent++
	}

	status := "partial"
	if sent > 0 && failed == 0 {
		status = "sent"
	} else if sent == 0 && failed > 0 {
		status = "failed"
	} else if sent == 0 && skipped > 0 && failed == 0 {
		status = "sent"
	}

	if lastAudience == "" {
		lastAudience = digestAudienceLabel(rawBranch, "fr")
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":    status,
		"weekStart": weekStart.Format("2006-01-02"),
		"from":      since.Format("2006-01-02"),
		"to":        today.Format("2006-01-02"),
		"days":      len(digests),
		"sent":      sent,
		"skipped":   skipped,
		"failed":    failed,
		"branch":    lastAudience,
	})
}
