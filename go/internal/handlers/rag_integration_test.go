package handlers_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

type fakeRAGEmbedder struct{}

func (fakeRAGEmbedder) EmbedTexts(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = fakeEmbed(t)
	}
	return out, nil
}

func fakeEmbed(text string) []float32 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	seed := h.Sum64()
	v := make([]float32, gemini.EmbeddingDims)
	for i := range v {
		seed = seed*6364136223846793005 + 1
		v[i] = float32(int64(seed>>33)%2001-1000) / 1000.0
	}
	return v
}

func TestRAGDisabled(t *testing.T) {
	t.Setenv("AI_CR_ADVANCED_ENABLED", "false")
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/rag/documents", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d %s", rec.Code, rec.Body.String())
	}
}

func TestRAGPlatformUploadAndSearch(t *testing.T) {
	t.Setenv("AI_CR_ADVANCED_ENABLED", "true")
	t.Setenv("RAG_REINDEX_SECRET", "test-rag-secret")
	api := newTestAPI(t)
	if !ragSchemaReady(t, api) {
		return
	}
	bundle, err := media.New(config.Config{MediaLocalDir: t.TempDir(), APIPublicURL: "http://localhost:8291"})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)
	api.api.TestSetRAGEmbedder(fakeRAGEmbedder{})

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	body := "Posologie amoxicilline chien : 10 mg/kg deux fois par jour. " +
		"Ne pas utiliser chez le furet. Guide technique petsFollow RAG test " + sha256sum("platform")
	code, env := doRAGUpload(t, api.handler, adminTok, "/api/v1/admin/rag/documents", "guide.txt", "text/plain", []byte(body), "Guide posologie")
	if code != http.StatusAccepted {
		t.Fatalf("admin upload %d %#v", code, env)
	}
	doc := asMap(env["data"])
	docID, _ := doc["id"].(string)
	if docID == "" {
		t.Fatalf("missing id %#v", doc)
	}
	ready := waitRAGReady(t, api, adminTok, docID)
	if ready["status"] != "ready" {
		t.Fatalf("want ready got %#v", ready)
	}
	if chunkCountFromAny(ready["chunkCount"]) < 1 {
		t.Fatalf("want chunks got %#v", ready)
	}
	doc = ready

	// Pending practice doc must not appear in search.
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	pendingBody := "Document cabinet secret non approuve " + sha256sum("pending")
	code, env = doRAGUpload(t, api.handler, vetTok, "/api/v1/practices/me/rag/documents", "cabinet.txt", "text/plain", []byte(pendingBody), "Cabinet pending")
	if code != http.StatusCreated {
		t.Fatalf("practice upload %d %#v", code, env)
	}
	pending := asMap(env["data"])
	if pending["status"] != "pending" {
		t.Fatalf("want pending %#v", pending)
	}
	pendingID, _ := pending["id"].(string)

	searchBody, _ := json.Marshal(map[string]any{
		"query":      "posologie amoxicilline chien",
		"practiceId": "",
		"limit":      5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/rag/search", bytes.NewReader(searchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rag-Reindex-Secret", "test-rag-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search %d %s", rec.Code, rec.Body.String())
	}
	var searchEnv map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &searchEnv)
	hits, _ := searchEnv["data"].([]any)
	if len(hits) < 1 {
		t.Fatalf("want search hits got %#v", searchEnv)
	}
	for _, h := range hits {
		m := asMap(h)
		if m["documentId"] == pendingID {
			t.Fatalf("pending doc leaked into search: %#v", m)
		}
	}

	// Approve practice doc → async index → ready.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/rag/documents/"+pendingID+"/approve", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("approve %d %#v", code, env)
	}
	approved := waitRAGReady(t, api, adminTok, pendingID)
	if approved["status"] != "ready" {
		t.Fatalf("want ready after approve %#v", approved)
	}

	// Invalid status filter → 400
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/rag/documents?status=nope", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status want 400 got %d", rec.Code)
	}

	// Isolation: vet.parc must not see VetPlus practice chunks.
	var parcPracticeID string
	if err := api.pool.QueryRow(context.Background(),
		`SELECT practice_id::text FROM identity.users WHERE email='vet.parc@petsfollow.test'`).Scan(&parcPracticeID); err != nil {
		t.Fatalf("parc practice: %v", err)
	}
	searchBody, _ = json.Marshal(map[string]any{
		"query":      "Document cabinet secret non approuve",
		"practiceId": parcPracticeID,
		"limit":      5,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/internal/rag/search", bytes.NewReader(searchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rag-Reindex-Secret", "test-rag-secret")
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &searchEnv)
	hits, _ = searchEnv["data"].([]any)
	for _, h := range hits {
		m := asMap(h)
		if m["documentId"] == pendingID {
			t.Fatalf("other practice saw cabinet doc: %#v", m)
		}
	}

	// Cleanup
	_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/rag/documents/"+doc["id"].(string), adminTok, nil)
	_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/rag/documents/"+pendingID, adminTok, nil)
}

func TestRAGDownloadAndReindex(t *testing.T) {
	t.Setenv("AI_CR_ADVANCED_ENABLED", "true")
	t.Setenv("RAG_REINDEX_SECRET", "test-rag-secret")
	api := newTestAPI(t)
	if !ragSchemaReady(t, api) {
		return
	}
	bundle, err := media.New(config.Config{MediaLocalDir: t.TempDir(), APIPublicURL: "http://localhost:8291"})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)
	api.api.TestSetRAGEmbedder(fakeRAGEmbedder{})

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")

	body := "Download reindex fixture " + sha256sum("dl-reindex")
	code, env := doRAGUpload(t, api.handler, vetTok, "/api/v1/practices/me/rag/documents", "cab.txt", "text/plain", []byte(body), "Cab DL")
	if code != http.StatusCreated {
		t.Fatalf("practice upload %d %#v", code, env)
	}
	docID, _ := asMap(env["data"])["id"].(string)

	// Practice can download own pending doc.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/practices/me/rag/documents/"+docID+"/download", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("practice download %d %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Download reindex fixture")) {
		t.Fatalf("unexpected body %q", rec.Body.String())
	}

	// Other practice forbidden.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/practices/me/rag/documents/"+docID+"/download", nil)
	req.Header.Set("Authorization", "Bearer "+parcTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("parc download want 403 got %d", rec.Code)
	}

	// Admin download OK.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/rag/documents/"+docID+"/download", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin download %d", rec.Code)
	}

	// Approve then wait ready; admin reindex must force-rebuild ready docs (not a no-op).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/rag/documents/"+docID+"/approve", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("approve %d %#v", code, env)
	}
	ready := waitRAGReady(t, api, adminTok, docID)
	if ready["status"] != "ready" {
		t.Fatalf("want ready got %#v", ready)
	}
	chunksBefore, _ := ready["chunkCount"].(float64)
	if chunksBefore < 1 {
		t.Fatalf("expected chunks before reindex %#v", ready)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/rag/reindex", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin reindex %d %#v", code, env)
	}
	stats := asMap(env["data"])
	considered, _ := stats["considered"].(float64)
	indexed, _ := stats["indexed"].(float64)
	if considered < 1 {
		t.Fatalf("reindex considered want ≥1 got %#v", stats)
	}
	if indexed < 1 {
		t.Fatalf("reindex indexed want ≥1 got %#v", stats)
	}

	after := waitRAGReady(t, api, adminTok, docID)
	if after["status"] != "ready" {
		t.Fatalf("after reindex %#v", after)
	}

	_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/rag/documents/"+docID, adminTok, nil)
}

func TestAdminImproveRunStats(t *testing.T) {
	t.Setenv("AI_CR_ADVANCED_ENABLED", "true")
	api := newTestAPI(t)
	if !ragSchemaReady(t, api) {
		return
	}
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/rag/improve-stats?days=7", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("improve-stats %d %#v", code, env)
	}
	data := asMap(env["data"])
	if _, ok := data["total"]; !ok {
		t.Fatalf("missing total %#v", data)
	}
	if _, ok := data["byStatus"]; !ok {
		t.Fatalf("missing byStatus %#v", data)
	}
	if _, ok := data["latencyP50Ms"]; !ok {
		t.Fatalf("missing latencyP50Ms %#v", data)
	}
}

func waitRAGReady(t *testing.T, api *testAPI, adminTok, docID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/rag/documents", adminTok, nil)
		if code != http.StatusOK {
			t.Fatalf("list %d %#v", code, env)
		}
		rows, _ := env["data"].([]any)
		for _, row := range rows {
			m := asMap(row)
			if m["id"] != docID {
				continue
			}
			switch m["status"] {
			case "ready", "failed":
				return m
			}
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for rag doc %s", docID)
	return nil
}

func ragSchemaReady(t *testing.T, api *testAPI) bool {
	t.Helper()
	var ok bool
	err := api.pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')`).Scan(&ok)
	if err != nil || !ok {
		t.Skip("pgvector extension not available — recreate DB with pgvector/pgvector:pg16")
		return false
	}
	err = api.pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='rag' AND table_name='documents')`).Scan(&ok)
	if err != nil || !ok {
		t.Skip("rag schema missing — run migrations")
		return false
	}
	return true
}

func doRAGUpload(t *testing.T, h http.Handler, token, path, filename, contentType string, data []byte, title string) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("title", title)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	// Hint MIME for text parts when Go multipart defaults to octet-stream.
	_ = contentType
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var env map[string]any
	b, _ := io.ReadAll(rec.Body)
	_ = json.Unmarshal(b, &env)
	return rec.Code, env
}

func sha256sum(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func chunkCountFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
