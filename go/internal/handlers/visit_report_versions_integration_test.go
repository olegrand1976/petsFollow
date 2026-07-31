package handlers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

func TestVisitReportTranscriptPutImproveVersionsFlow(t *testing.T) {
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(7 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "cr versions flow",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	// 1) Transcribe via hint (no Gemini).
	transcript1 := "Anamnèse brute : toux " + string([]byte{0x00, 0x01}) + "chronique «client» & chat."
	rec := postTranscribeHint(t, api, vetTok, visitID, transcript1)
	if rec.Code != http.StatusOK {
		t.Fatalf("transcribe %d %s", rec.Code, rec.Body.String())
	}
	rep := decodeReportData(t, rec.Body.Bytes())
	gotTranscript, _ := rep["transcriptText"].(string)
	gotBody, _ := rep["bodyText"].(string)
	if !strings.Contains(gotTranscript, "toux") || strings.Contains(gotTranscript, "\x00") {
		t.Fatalf("transcript normalized want toux without NUL, got %q", gotTranscript)
	}
	if gotBody == "" {
		t.Fatal("empty body should be filled from transcript")
	}

	// 2) PUT body distinct + transcript with escape corpus (optional transcriptText).
	escapeNotes := `Notes "quotes" 'apostrophe' «guillemets» back\slash &amp; <script>x</script> **bold**`
	editedBody := "**Anamnèse / motif :**\n\n- toux\n- appétit ok"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText":       editedBody,
		"transcriptText": escapeNotes,
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	rep = dataMap(t, env)
	if rep["bodyText"] != editedBody {
		t.Fatalf("body want edited, got %#v", rep["bodyText"])
	}
	tr, _ := rep["transcriptText"].(string)
	if !strings.Contains(tr, `"quotes"`) || !strings.Contains(tr, `back\slash`) {
		t.Fatalf("transcript escape round-trip failed: %q", tr)
	}
	if strings.Contains(tr, "<script>") {
		// Normalize does not strip HTML tags from source text — only C0. Script may remain in source;
		// XSS is a preview concern. Ensure round-trip kept the literal for editability.
	}
	improvedBefore, _ := rep["improvedText"].(string)

	// 3) Re-transcribe: v0 overwritten, body preserved.
	rec = postTranscribeHint(t, api, vetTok, visitID, "Deuxième dictée remplace v0")
	if rec.Code != http.StatusOK {
		t.Fatalf("re-transcribe %d %s", rec.Code, rec.Body.String())
	}
	rep = decodeReportData(t, rec.Body.Bytes())
	if rep["transcriptText"] != "Deuxième dictée remplace v0" {
		t.Fatalf("v0 overwrite failed: %#v", rep["transcriptText"])
	}
	if rep["bodyText"] != editedBody {
		t.Fatalf("body must be preserved on re-transcribe, got %#v", rep["bodyText"])
	}
	if improved, _ := rep["improvedText"].(string); improved != improvedBefore {
		t.Fatalf("improved should be untouched, got %q want %q", improved, improvedBefore)
	}

	// 4) PUT without transcriptText must not wipe transcript.
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "body only update",
	})
	if code != http.StatusOK {
		t.Fatalf("put body-only %d %#v", code, env)
	}
	rep = dataMap(t, env)
	if rep["transcriptText"] != "Deuxième dictée remplace v0" {
		t.Fatalf("omitted transcriptText wiped transcript: %#v", rep["transcriptText"])
	}

	// 5) Improve with sourceText must not wipe body on failure path (PUT keeps body).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve", vetTok, map[string]any{
		"sourceText": "Deuxième dictée remplace v0",
	})
	switch code {
	case http.StatusOK:
		rep = dataMap(t, env)
		if strings.TrimSpace(strField(rep, "improvedText")) == "" {
			t.Fatal("improve 200 without improvedText")
		}
		if strings.TrimSpace(strField(rep, "bodyText")) == "" {
			t.Fatal("improve 200 without bodyText")
		}
		v0 := strField(rep, "transcriptText")
		if v0 != "Deuxième dictée remplace v0" {
			t.Fatalf("improve must keep transcript, got %q", v0)
		}
	case http.StatusPaymentRequired, http.StatusServiceUnavailable, http.StatusBadGateway:
		t.Logf("improve skipped (entitlement/gemini): %d", code)
		// Body must still be the pre-improve PUT value (not overwritten by source).
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
		if code != http.StatusOK {
			t.Fatalf("get after failed improve %d %#v", code, env)
		}
		if dataMap(t, env)["bodyText"] != "body only update" {
			t.Fatalf("body wiped despite improve failure: %#v", dataMap(t, env)["bodyText"])
		}
	default:
		t.Fatalf("improve unexpected %d %#v", code, env)
	}

	// 6) Finalize + conflict on further PUT.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "final" {
		t.Fatalf("want final status, got %#v", dataMap(t, env)["status"])
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "should fail",
	})
	if code != http.StatusConflict {
		t.Fatalf("put after final want 409, got %d %#v", code, env)
	}

	// 7) Continuous improvement: mark / unmark reference on finalized CR.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/report/reference", vetTok, map[string]any{
		"isReference": true,
	})
	if code != http.StatusOK {
		t.Fatalf("mark reference %d %#v", code, env)
	}
	if dataMap(t, env)["isReference"] != true {
		t.Fatalf("want isReference true, got %#v", dataMap(t, env)["isReference"])
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/report/reference", vetTok, map[string]any{
		"isReference": false,
	})
	if code != http.StatusOK {
		t.Fatalf("unmark reference %d %#v", code, env)
	}
	if dataMap(t, env)["isReference"] != false {
		t.Fatalf("want isReference false, got %#v", dataMap(t, env)["isReference"])
	}
}

func TestVisitReportImproveInvalidTargetLocale(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(11 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "invalid locale improve",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve", vetTok, map[string]any{
		"sourceText":   "toux",
		"targetLocale": "de",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid targetLocale want 400, got %d %#v", code, env)
	}
}

func TestVisitReportReferenceRequiresFinal(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(13 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "reference draft",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "**Anamnèse / motif :**\n\ndraft",
	})
	if code != http.StatusOK {
		t.Fatalf("put draft %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/report/reference", vetTok, map[string]any{
		"isReference": true,
	})
	if code != http.StatusConflict {
		t.Fatalf("reference on draft want 409, got %d %#v", code, env)
	}

	// No report for another visit path: empty visit without ensure — PATCH without prior PUT
	slot2 := time.Now().UTC().Add(14 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot2,
		"notes":               "no report",
		"durationMinutes":     15,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit2 %d %#v", code, env)
	}
	visitID2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID2, vetTok, nil)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID2+"/report/reference", vetTok, map[string]any{
		"isReference": true,
	})
	if code != http.StatusNotFound {
		t.Fatalf("reference without report want 404, got %d %#v", code, env)
	}
}

func postTranscribeHint(t *testing.T, api *testAPI, tok, visitID, hint string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("clientAudioConsent", "true")
	_ = w.WriteField("hint", hint)
	part, err := w.CreateFormFile("audio", "dictation.webm")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("fake-webm-bytes"))
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits/"+visitID+"/report/transcribe", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	return rec
}

func decodeReportData(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode: %v body=%s", err, string(raw))
	}
	return dataMap(t, env)
}

func strField(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}
