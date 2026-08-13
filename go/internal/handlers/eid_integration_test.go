package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestEidWebEidChallenge_RedisRequiredOutsideLocal(t *testing.T) {
	api := newTestAPI(t)
	// Simulate staging/prod: no Redis on API, APP_ENV production, no DEV_SEED.
	t.Setenv("APP_ENV", "production")
	api.api.TestSetDevSeedEnabled(false)
	t.Cleanup(func() { api.api.TestSetDevSeedEnabled(true) })

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/eid/web-eid/challenge", vetTok, nil)
	if code != http.StatusServiceUnavailable || errorMsgKey(env) != "eid_redis_required" {
		t.Fatalf("want eid_redis_required 503 got %d %#v", code, env)
	}
}

func TestEidReadings_PurgeOld(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	var practiceID, userID string
	err := api.pool.QueryRow(ctx, `
		SELECT practice_id::text, id::text
		FROM identity.users
		WHERE email = 'vet.demo@petsfollow.test'`).Scan(&practiceID, &userID)
	if err != nil || practiceID == "" || userID == "" {
		t.Skip("seed vet/practice unavailable")
	}
	oldID := uuid.NewString()
	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.eid_readings (
			id, practice_id, user_id, tool, success, fields_read, niss_hash, error_code, created_at
		) VALUES ($1,$2,$3,'eid_viewer_xml',true,'{niss}','abc','',$4)`,
		oldID, practiceID, userID, time.Now().Add(-400*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	st := store.New(api.pool)
	n, err := st.PurgeOldEidReadings(ctx, time.Now().Add(-365*24*time.Hour), 500)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected purge >= 1, got %d", n)
	}
}

func TestEidImportViewer_BE(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	raw := loadEidFixture(t)
	code, env := doEidUpload(t, api.handler, vetTok, raw, "sample_valid.eid")
	if code != http.StatusOK {
		t.Fatalf("import %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["lastname"] != "Testeur" || data["firstname"] != "Camille" {
		t.Fatalf("identity %#v", data)
	}
	if data["niss"] != "96072399828" {
		t.Fatalf("niss %#v", data)
	}
	if data["country"] != "BE" {
		t.Fatalf("country %#v", data)
	}
	if _, hasPhoto := data["photo_jpeg_base64"]; hasPhoto {
		t.Fatalf("photo must not be returned to client: %#v", data)
	}
}

func TestEidImport_StripsPhotoFromResponse(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	raw := []byte(`<?xml version="1.0"?>
<Export>
  <surname>Dupont</surname>
  <firstname>Jean</firstname>
  <nationalnumber>85010112345</nationalnumber>
  <photo>` + "aGVsbG8=" + `</photo>
</Export>`)
	code, env := doEidUpload(t, api.handler, vetTok, raw, "with-photo.eid")
	if code != http.StatusOK {
		t.Fatalf("import %d %#v", code, env)
	}
	data := dataMap(t, env)
	if _, hasPhoto := data["photo_jpeg_base64"]; hasPhoto {
		t.Fatalf("photo leaked: %#v", data)
	}
	if data["firstname"] != "Jean" {
		t.Fatalf("identity %#v", data)
	}
}

func TestEidImport_EmptyIdentityRejected(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	raw := []byte(`<?xml version="1.0"?><Export><nationality>BE</nationality></Export>`)
	code, env := doEidUpload(t, api.handler, vetTok, raw, "empty.eid")
	if code != http.StatusUnprocessableEntity || errorMsgKey(env) != "eid_identity_empty" {
		t.Fatalf("want eid_identity_empty 422 got %d %#v", code, env)
	}
}

func TestEidImport_PDFRejected(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doEidUpload(t, api.handler, vetTok, []byte("%PDF-1.4 fake"), "card.pdf")
	if code != http.StatusUnprocessableEntity || errorMsgKey(env) != "eid_pdf_unsupported" {
		t.Fatalf("want eid_pdf_unsupported got %d %#v", code, env)
	}
}

func TestEidImport_Disabled(t *testing.T) {
	t.Setenv("EID_ENABLED", "false")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doEidUpload(t, api.handler, vetTok, loadEidFixture(t), "sample_valid.eid")
	if code != http.StatusNotFound || errorMsgKey(env) != "eid_disabled" {
		t.Fatalf("want eid_disabled 404 got %d %#v", code, env)
	}
}

func TestEidImport_RequiresAuth(t *testing.T) {
	api := newTestAPI(t)
	code, _ := doEidUpload(t, api.handler, "", loadEidFixture(t), "sample_valid.eid")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("want 401/403 got %d", code)
	}
}

func TestEidWebEidChallengeAndBadToken(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/eid/web-eid/challenge", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("challenge %d %#v", code, env)
	}
	data := dataMap(t, env)
	nonce, _ := data["nonce"].(string)
	if nonce == "" {
		t.Fatalf("missing nonce %#v", data)
	}
	origin, _ := data["origin"].(string)
	if origin == "" {
		t.Fatalf("missing origin %#v", data)
	}

	// Empty body must not consume nonce (parse before GetDel).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/eid/web-eid/verify", vetTok, map[string]any{})
	if code != http.StatusBadRequest || errorMsgKey(env) != "token_required" {
		t.Fatalf("want token_required 400 got %d %#v", code, env)
	}

	// Nonce still usable → invalid token yields 422 (not nonce_expired).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/eid/web-eid/verify", vetTok, map[string]any{
		"token": map[string]any{"unverifiedCertificate": "nope"},
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("verify want 422 got %d %#v", code, env)
	}
	if errorMsgKey(env) == "eid_nonce_expired" {
		t.Fatalf("nonce was burned by empty body: %#v", env)
	}

	// Second verify without new challenge → nonce expired.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/eid/web-eid/verify", vetTok, map[string]any{
		"token": map[string]any{"unverifiedCertificate": "nope"},
	})
	if code != http.StatusUnprocessableEntity || errorMsgKey(env) != "eid_nonce_expired" {
		t.Fatalf("want eid_nonce_expired got %d %#v", code, env)
	}
}

func TestEidWebEidChallenge_Disabled(t *testing.T) {
	t.Setenv("EID_ENABLED", "false")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/eid/web-eid/challenge", vetTok, nil)
	if code != http.StatusNotFound || errorMsgKey(env) != "eid_disabled" {
		t.Fatalf("want eid_disabled got %d %#v", code, env)
	}
}

func loadEidFixture(t *testing.T) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	path := filepath.Join(filepath.Dir(file), "..", "eid", "testdata", "sample_valid.eid")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func doEidUpload(t *testing.T, h http.Handler, token string, file []byte, filename string) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(file); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vet/eid/import", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var env map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &env)
	return rr.Code, env
}
