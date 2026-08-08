package handlers_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/internal/vetnews"
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

	st := store.New(api.pool)
	url := "https://www.anses.fr/fr/content/test-vet-news-" + strings.ReplaceAll(uniqueEmail("vn"), "@", "-")
	art := vetnews.Article{
		SourceID:   "anses",
		SourceName: "Anses — Santé animale",
		SourceURL:  url,
		Title:      "Rage : une réémergence préoccupante",
		Summary:    "Résumé test",
		Category:   "epidemio",
		Importance: "critical",
		Tags:       []string{"anses", "fr"},
		FetchedAt:  time.Now().UTC(),
	}
	outcome, err := st.UpsertVetNewsArticle(t.Context(), art)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != vetnews.UpsertInserted {
		t.Fatalf("first upsert want inserted got %v", outcome)
	}
	outcome, err = st.UpsertVetNewsArticle(t.Context(), art)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != vetnews.UpsertUnchanged {
		t.Fatalf("second upsert same hash want unchanged got %v", outcome)
	}
	art.Title = "Rage : titre mis à jour"
	outcome, err = st.UpsertVetNewsArticle(t.Context(), art)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != vetnews.UpsertUpdated {
		t.Fatalf("title change want updated got %v", outcome)
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
	sources, _ := data["sources"].([]any)
	if len(sources) != 6 {
		t.Fatalf("sources %#v", sources)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/news?importance=not-a-level", vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("bad importance want 400 got %d %#v", code, env)
	}
}
