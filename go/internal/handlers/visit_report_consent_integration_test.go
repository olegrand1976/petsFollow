package handlers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranscribeRequiresClientAudioConsent(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-12-01T10:00:00Z",
		"notes":           "audio consent probe",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("audio", "probe.m4a")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("fake-audio-bytes"))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits/"+visitID+"/report/transcribe", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 without consent, got %d %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("audio_consent_required")) &&
		!bytes.Contains(rec.Body.Bytes(), []byte(`"code":"bad_request"`)) {
		t.Fatalf("expected consent error body, got %s", rec.Body.String())
	}
}
