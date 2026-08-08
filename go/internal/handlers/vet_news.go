package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/vetnews"
)

func (a *API) requireVetNewsEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.VetNewsEnabled {
		writeErr(w, r, http.StatusNotFound, "vet_news_disabled", "not_found")
		return false
	}
	return true
}

func (a *API) vetNewsIngester() *vetnews.Ingester {
	return vetnews.NewIngester(a.store)
}

// internalRunVetNews — CRON : ingest multi-sources (Anses, VE, CB, TVP, LPV, WOAH).
// Header X-Vet-News-Secret (env VET_NEWS_SECRET).
func (a *API) internalRunVetNews(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Vet-News-Secret", a.cfg.VetNewsSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.cfg.VetNewsEnabled {
		writeErr(w, r, http.StatusNotFound, "vet_news_disabled", "not_found")
		return
	}
	res := a.vetNewsIngester().Run(r.Context())
	log.Printf("vetnews run inserted=%d updated=%d unchanged=%d skipped=%d duration_ms=%d sources=%d",
		res.Inserted, res.Updated, res.Unchanged, res.Skipped, res.DurationMs, len(res.Sources))
	httpx.WriteData(w, http.StatusOK, res)
}

// listVetNews returns recent ingested articles for practice dashboards.
func (a *API) listVetNews(w http.ResponseWriter, r *http.Request) {
	if !a.requireVetNewsEnabled(w, r) {
		return
	}
	if _, ok := a.requirePracticePerm(w, r, "clients.read"); !ok {
		return
	}
	limit := 15
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	importance := strings.TrimSpace(r.URL.Query().Get("importance"))
	switch importance {
	case "", vetnews.ImportanceCritical, vetnews.ImportanceHigh, vetnews.ImportanceMedium, vetnews.ImportanceLow:
	default:
		writeErr(w, r, http.StatusBadRequest, "invalid_importance", "bad_request")
		return
	}
	filter := vetnews.ListFilter{
		SourceID:   strings.TrimSpace(r.URL.Query().Get("source")),
		Importance: importance,
		Limit:      limit,
	}
	items, err := a.store.ListVetNewsArticles(r.Context(), filter)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"items":   items,
		"legend":  vetnews.ImportanceLegend(),
		"sources": vetnews.SourceCatalog(),
	})
}
