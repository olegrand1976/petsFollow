package handlers_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestVetNewsDisabled(t *testing.T) {
	t.Setenv("VET_NEWS_ENABLED", "false")
	t.Setenv("VET_NEWS_SECRET", "test-vet-news")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/news", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("want 404 got %d %#v", code, env)
	}
	if errCode(env) != "vet_news_disabled" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestVetNewsListAndIngestAuth(t *testing.T) {
	t.Setenv("VET_NEWS_ENABLED", "true")
	t.Setenv("VET_NEWS_SECRET", "test-vet-news-secret")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/internal/vet-news/run", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("no secret want 401 got %d %#v", code, env)
	}

	// Seed one article directly (avoid live network in CI).
	url := "https://www.anses.fr/fr/content/test-vet-news-" + strings.ReplaceAll(uniqueEmail("vn"), "@", "-")
	_, err := api.pool.Exec(t.Context(), `
		INSERT INTO ops.vet_news_articles (
			source_id, source_name, source_url, title, summary, category, importance, tags
		) VALUES (
			'anses', 'Anses — Santé animale', $1,
			'Rage : une réémergence préoccupante',
			'Résumé test',
			'epidemio', 'critical', ARRAY['anses','fr']
		)`, url)
	if err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/news?limit=5", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected items %#v", data)
	}
	legend, _ := data["legend"].([]any)
	if len(legend) != 4 {
		t.Fatalf("legend %#v", legend)
	}
	first, _ := items[0].(map[string]any)
	if first["importance"] == nil || first["title"] == nil {
		t.Fatalf("item %#v", first)
	}
}
