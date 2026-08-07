package handlers_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Bornes de corps et compression : le webhook Stripe est public et non
// authentifié, le profil cabinet bufferise son corps pour le relire deux fois.
// Sans limite, un POST suffit à faire monter le tas d'une instance Cloud Run.

// countingReader compte ce que le handler consomme réellement : un webhook
// rejeté après avoir tout avalé n'aurait rien borné du tout.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func TestStripeWebhookDoesNotBufferOversizedBody(t *testing.T) {
	api := newTestAPI(t)

	const sent = 16 << 20 // 16 Mo, très au-delà d'un event Stripe (quelques Ko)
	body := &countingReader{r: io.LimitReader(neverEndingByte('a'), sent)}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhooks/stripe", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
		t.Fatalf("webhook = %d %s, want 413 (or 400)", rec.Code, rec.Body.String())
	}
	// Marge d'un buffer de lecture au-dessus de la borne du handler.
	if maxRead := int64(1<<20) + 64<<10; body.n > maxRead {
		t.Fatalf("handler a lu %d octets, borne attendue ~%d", body.n, maxRead)
	}
}

type neverEndingByte byte

func (b neverEndingByte) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

func TestVetProfileRejectsOversizedBody(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	payload := map[string]any{"displayName": strings.Repeat("x", 512<<10)}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/vet/profile", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized vet profile = %d %s, want 413", rec.Code, rec.Body.String())
	}
}

func TestVetProfileAcceptsNormalBody(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/profile", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("read profile %d %#v", code, env)
	}
	profile := dataMap(t, env)

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", vetTok, profile)
	if code != http.StatusOK {
		t.Fatalf("write back profile %d %#v", code, env)
	}
}

// La liste des animaux d’un cabinet est le plus gros JSON servi au Pro : sans
// compression, chaque rafraîchissement repasse tout en clair sur le réseau.
func TestJSONResponsesAreCompressed(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vet/pets", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("clients = %d %s", rec.Code, rec.Body.String())
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}

	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer zr.Close()
	plain, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(plain, &env); err != nil {
		t.Fatalf("decoded body is not JSON: %v", err)
	}
	if _, ok := env["data"]; !ok {
		t.Fatalf("missing data envelope: %s", plain)
	}
}

func TestNonGzipClientStillGetsPlainJSON(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vet/pets", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("clients = %d %s", rec.Code, rec.Body.String())
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want none", enc)
	}
	var env map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not plain JSON: %v", err)
	}
}
