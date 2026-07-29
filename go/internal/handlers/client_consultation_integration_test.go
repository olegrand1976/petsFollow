package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

func seedFinalConsultation(t *testing.T, api *testAPI) (clientTok, visitID, petID string) {
	t.Helper()
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok = loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID = activeDemoPetID(t, api.handler, clientTok)

	// Walk-in window: scheduledAt must be within ±30m of now.
	slot := time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "client consultation view probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ = dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "Examen clinique OK — probe client consultation.",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	// Mark visit as past so client timeline "consultations" section would pick it up.
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE visits.visits
		SET status = 'done', scheduled_at = NOW() - interval '2 days'
		WHERE id = $1`, visitID); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	return clientTok, visitID, petID
}

func TestClientConsultationReadAndShare(t *testing.T) {
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	clientTok, visitID, petID := seedFinalConsultation(t, api)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/visits", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list visits %d %#v", code, env)
	}
	foundFlag := false
	for _, row := range env["data"].([]any) {
		m, _ := row.(map[string]any)
		if m["id"] == visitID {
			foundFlag = m["hasFinalReport"] == true
			break
		}
	}
	if !foundFlag {
		t.Fatalf("expected hasFinalReport on visit %s %#v", visitID, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/client-consultation", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client consultation %d %#v", code, env)
	}
	data := dataMap(t, env)
	reports, _ := data["reports"].([]any)
	if len(reports) == 0 {
		t.Fatalf("expected final reports %#v", data)
	}
	body, _ := reports[0].(map[string]any)["bodyText"].(string)
	if !strings.Contains(body, "Examen clinique OK") {
		t.Fatalf("body %#v", body)
	}

	otherTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/client-consultation", otherTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("other client want 403 got %d %#v", code, env)
	}

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/client-consultation", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/consultation-shares", clientTok, map[string]any{
		"email": "pro.externe@petsfollow.test",
	})
	if code != http.StatusCreated {
		t.Fatalf("create share %d %#v", code, env)
	}
	token, _ := dataMap(t, env)["token"].(string)
	if token == "" {
		t.Fatalf("expected DEV token %#v", env)
	}

	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/consultation/"+token, nil)
	if code != http.StatusOK {
		t.Fatalf("meta %d %#v", code, env)
	}
	if dataMap(t, env)["petName"] == "" {
		t.Fatalf("expected petName %#v", env)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/consultation/"+token+"/download", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "pdf") {
		t.Fatalf("content-type=%q", rec.Header().Get("Content-Type"))
	}
	if string(rec.Body.Bytes()[:5]) != "%PDF-" {
		t.Fatalf("pdf magic %q", rec.Body.Bytes()[:5])
	}

	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pets.consultation_share_tokens SET expires_at = $1 WHERE token = $2`,
		time.Now().UTC().Add(-time.Minute), token); err != nil {
		t.Fatalf("expire: %v", err)
	}
	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/consultation/"+token, nil)
	if code != http.StatusGone {
		t.Fatalf("expired want 410 got %d %#v", code, env)
	}
}

func TestClientConsultationDraftExcluded(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	slot := time.Now().UTC().Add(8 * time.Minute).Format(time.RFC3339)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "draft only",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "brouillon seulement",
	})
	if code != http.StatusOK {
		t.Fatalf("put %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/client-consultation", clientTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("draft want 404 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/consultation-shares", clientTok, map[string]any{
		"email": "pro.externe@petsfollow.test",
	})
	if code != http.StatusNotFound {
		t.Fatalf("share draft want 404 got %d %#v", code, env)
	}
}
