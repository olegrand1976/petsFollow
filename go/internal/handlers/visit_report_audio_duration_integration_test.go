package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

func TestTranscribePersistsAudioDurationAndList(t *testing.T) {
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

	slot := time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "audio duration probe",
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

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("clientAudioConsent", "true")
	_ = w.WriteField("audioDurationSec", "154")
	_ = w.WriteField("hint", "Transcription durée probe sans Gemini.")
	part, err := w.CreateFormFile("audio", "dictation.webm")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("fake-webm-bytes-for-duration-test"))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits/"+visitID+"/report/transcribe", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("transcribe %d %s", rec.Code, rec.Body.String())
	}
	var reportEnv map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &reportEnv); err != nil {
		t.Fatalf("decode transcribe: %v", err)
	}
	report := dataMap(t, reportEnv)
	if report["hasAudio"] != true {
		t.Fatalf("hasAudio want true %#v", report)
	}
	if dur, _ := report["audioDurationSec"].(float64); int(dur) != 154 {
		t.Fatalf("audioDurationSec want 154 got %#v", report["audioDurationSec"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations?hasAudio=true&q=audio+duration+probe", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	found := false
	for _, row := range env["data"].([]any) {
		m, _ := row.(map[string]any)
		if m["id"] != visitID {
			continue
		}
		found = true
		if m["hasAudio"] != true {
			t.Fatalf("list hasAudio %#v", m)
		}
		if dur, _ := m["audioDurationSec"].(float64); int(dur) != 154 {
			t.Fatalf("list audioDurationSec want 154 %#v", m)
		}
	}
	if !found {
		t.Fatalf("visit %s not listed %#v", visitID, env["data"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/visits/"+visitID+"/report/audio", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("audio stream %d %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "audio/webm" {
		t.Fatalf("content-type want audio/webm got %q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("empty audio body")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	var durDB int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(audio_duration_sec, 0) FROM visits.visit_reports
		WHERE visit_id = $1::uuid
		ORDER BY updated_at DESC
		LIMIT 1`, visitID).Scan(&durDB); err != nil {
		t.Fatalf("duration after finalize: %v", err)
	}
	if durDB != 154 {
		t.Fatalf("duration after finalize want 154 got %d", durDB)
	}
}
