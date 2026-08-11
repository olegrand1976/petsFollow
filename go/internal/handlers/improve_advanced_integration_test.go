package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/crewai"
)

func TestImproveAdvancedSSE(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAiCrAdvancedEnabled(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{TaskID: "crew-task-1", Status: "running"})
	})
	mux.HandleFunc("/v1/tasks/crew-task-1/events", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		write := func(s string) { _, _ = io.WriteString(w, s); fl.Flush() }
		write("id: 1\nevent: step\ndata: {\"agent\":\"tri\",\"label\":\"Extraction\"}\n\n")
		time.Sleep(20 * time.Millisecond)
		write("id: 2\nevent: step\ndata: {\"agent\":\"clinicien\",\"label\":\"RAG\"}\n\n")
		time.Sleep(20 * time.Millisecond)
		write("id: 3\nevent: step\ndata: {\"agent\":\"redacteur\",\"label\":\"Sections\"}\n\n")
		write("id: 4\nevent: final\ndata: {\"report\":\"## Anamnèse / motif\\n\\nToux 3j\\n\"}\n\n")
		write("event: ping\ndata: {\"done\":true}\n\n")
	})
	crewSrv := httptest.NewServer(mux)
	t.Cleanup(crewSrv.Close)
	api.api.TestSetCrewAI(&crewai.Client{BaseURL: crewSrv.URL, Secret: "x", HTTPClient: crewSrv.Client()})

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != 200 {
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
		"notes":               "improve advanced sse",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"callbackPhone":       "0470000001",
	})
	if code != 201 && code != 200 {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := asMap(env["data"])["id"].(string)
	if visitID == "" {
		t.Fatalf("visit %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "Chien toux 3 jours",
	})
	if code != 200 {
		t.Fatalf("put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve-advanced", vetTok, map[string]any{
		"sourceText": "Chien toux 3 jours", "targetLocale": "fr",
	})
	if code != http.StatusAccepted {
		t.Fatalf("improve-advanced %d %#v", code, env)
	}
	runID, _ := asMap(env["data"])["runId"].(string)
	if runID == "" {
		t.Fatalf("runId %#v", env)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/visits/"+visitID+"/report/improve-advanced/"+runID+"/events", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		api.handler.ServeHTTP(rec, req)
	}()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("events SSE timeout")
	}
	if rec.Code != 200 {
		t.Fatalf("events %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: step") || !strings.Contains(body, "tri") {
		t.Fatalf("missing steps: %s", truncate(body, 400))
	}
	if !strings.Contains(body, "event: final") {
		t.Fatalf("missing final: %s", truncate(body, 600))
	}
	// Warm instance: crew_ready + running statuses, never crew_warming.
	if !strings.Contains(body, `"state":"crew_ready"`) || !strings.Contains(body, `"state":"running"`) {
		t.Fatalf("missing status events: %s", truncate(body, 600))
	}
	if strings.Contains(body, `"state":"crew_warming"`) {
		t.Fatalf("warm instance must not emit crew_warming: %s", truncate(body, 600))
	}

	deadline := time.Now().Add(5 * time.Second)
	var report map[string]any
	for time.Now().Before(deadline) {
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
		report = asMap(env["data"])
		if strings.Contains(str(report["bodyText"]), "Anamnèse") || strings.Contains(str(report["improvedText"]), "Anamnèse") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !strings.Contains(str(report["bodyText"])+str(report["improvedText"]), "Anamnèse") {
		t.Fatalf("report not updated %#v", report)
	}

	api.api.TestSetAiCrAdvancedEnabled(false)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve-advanced", vetTok, map[string]any{
		"sourceText": "x",
	})
	if code != http.StatusNotFound {
		t.Fatalf("flag off want 404 got %d %#v", code, env)
	}
}

func TestImproveAdvancedCancelStopsPersist(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAiCrAdvancedEnabled(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{TaskID: "crew-cancel-1", Status: "running"})
	})
	mux.HandleFunc("/v1/tasks/crew-cancel-1/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		_, _ = io.WriteString(w, "id: 1\nevent: step\ndata: {\"agent\":\"tri\",\"label\":\"slow\"}\n\n")
		fl.Flush()
		// Hold the stream open until client cancels (context) or timeout.
		select {
		case <-r.Context().Done():
		case <-time.After(6 * time.Second):
			_, _ = io.WriteString(w, "id: 2\nevent: final\ndata: {\"report\":\"## SHOULD NOT PERSIST\\n\"}\n\n")
			fl.Flush()
		}
	})
	crewSrv := httptest.NewServer(mux)
	t.Cleanup(crewSrv.Close)
	api.api.TestSetCrewAI(&crewai.Client{BaseURL: crewSrv.URL, Secret: "x", HTTPClient: crewSrv.Client()})

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != 200 {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(12 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "improve advanced cancel",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"callbackPhone":       "0470000001",
	})
	if code != 201 && code != 200 {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := asMap(env["data"])["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "Brouillon cancel",
	})
	if code != 200 {
		t.Fatalf("put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve-advanced", vetTok, map[string]any{
		"sourceText": "Brouillon cancel", "targetLocale": "fr",
	})
	if code != http.StatusAccepted {
		t.Fatalf("improve-advanced %d %#v", code, env)
	}
	runID, _ := asMap(env["data"])["runId"].(string)
	time.Sleep(80 * time.Millisecond)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve-advanced/"+runID+"/cancel", vetTok, nil)
	if code != 200 {
		t.Fatalf("cancel %d %#v", code, env)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
		report := asMap(env["data"])
		if strings.Contains(str(report["improvedText"]), "SHOULD NOT PERSIST") {
			t.Fatalf("cancel did not stop persist: %#v", report)
		}
		time.Sleep(50 * time.Millisecond)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
	report := asMap(env["data"])
	if strings.Contains(str(report["improvedText"])+str(report["bodyText"]), "SHOULD NOT PERSIST") {
		t.Fatalf("report polluted after cancel %#v", report)
	}
}

// setupImproveAdvancedVisit creates a consultation visit with a draft report
// and returns (vetToken, visitID). Cleanup deletes the visit.
func setupImproveAdvancedVisit(t *testing.T, api *testAPI, notes string, offset time.Duration) (string, string) {
	t.Helper()
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != 200 {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(offset).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               notes,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"callbackPhone":       "0470000001",
	})
	if code != 201 && code != 200 {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := asMap(env["data"])["id"].(string)
	if visitID == "" {
		t.Fatalf("visit %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": notes,
	})
	if code != 200 {
		t.Fatalf("put report %d %#v", code, env)
	}
	return vetTok, visitID
}

// collectImproveAdvancedSSE runs the improve then streams events to completion.
func collectImproveAdvancedSSE(t *testing.T, api *testAPI, vetTok, visitID string, timeout time.Duration) string {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve-advanced", vetTok, map[string]any{
		"sourceText": "Chien toux 3 jours", "targetLocale": "fr",
	})
	if code != http.StatusAccepted {
		t.Fatalf("improve-advanced %d %#v", code, env)
	}
	runID, _ := asMap(env["data"])["runId"].(string)
	if runID == "" {
		t.Fatalf("runId %#v", env)
	}
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/visits/"+visitID+"/report/improve-advanced/"+runID+"/events", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		api.handler.ServeHTTP(rec, req)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("events SSE timeout")
	}
	if rec.Code != 200 {
		t.Fatalf("events %d %s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestImproveAdvancedColdStartWarmup — minScale=0: health KO puis OK →
// SSE status crew_warming + crew_ready, run complet.
func TestImproveAdvancedColdStartWarmup(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAiCrAdvancedEnabled(true)

	healthCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		healthCalls++
		if healthCalls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{TaskID: "crew-cold-1", Status: "running"})
	})
	mux.HandleFunc("/v1/tasks/crew-cold-1/events", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		_, _ = io.WriteString(w, "id: 1\nevent: final\ndata: {\"report\":\"## Anamnèse / motif\\n\\nCold start OK\\n\"}\n\n")
		_, _ = io.WriteString(w, "event: ping\ndata: {\"done\":true}\n\n")
		fl.Flush()
	})
	crewSrv := httptest.NewServer(mux)
	t.Cleanup(crewSrv.Close)
	api.api.TestSetCrewAI(&crewai.Client{BaseURL: crewSrv.URL, Secret: "x", HTTPClient: crewSrv.Client()})

	vetTok, visitID := setupImproveAdvancedVisit(t, api, "improve advanced cold start", 13*time.Minute)
	body := collectImproveAdvancedSSE(t, api, vetTok, visitID, 10*time.Second)

	if !strings.Contains(body, `"state":"crew_warming"`) {
		t.Fatalf("missing crew_warming: %s", truncate(body, 600))
	}
	if !strings.Contains(body, `"state":"crew_ready"`) {
		t.Fatalf("missing crew_ready: %s", truncate(body, 600))
	}
	if !strings.Contains(body, `"state":"running"`) {
		t.Fatalf("missing running status: %s", truncate(body, 600))
	}
	if !strings.Contains(body, `"label":"crew_warming"`) {
		t.Fatalf("missing warm-up control step: %s", truncate(body, 600))
	}
	if !strings.Contains(body, "event: final") {
		t.Fatalf("missing final: %s", truncate(body, 800))
	}
	if healthCalls < 2 {
		t.Fatalf("health calls = %d", healthCalls)
	}
}

// TestImproveAdvancedCrewUnavailable — instance jamais up dans le budget →
// run failed + SSE error crewai_unavailable.
func TestImproveAdvancedCrewUnavailable(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAiCrAdvancedEnabled(true)
	api.api.TestSetCrewWarmupBudget(300 * time.Millisecond)
	t.Cleanup(func() { api.api.TestSetCrewWarmupBudget(90 * time.Second) })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, _ *http.Request) {
		t.Error("submit must not be called when instance never becomes ready")
	})
	crewSrv := httptest.NewServer(mux)
	t.Cleanup(crewSrv.Close)
	api.api.TestSetCrewAI(&crewai.Client{BaseURL: crewSrv.URL, Secret: "x", HTTPClient: crewSrv.Client()})

	vetTok, visitID := setupImproveAdvancedVisit(t, api, "improve advanced unavailable", 14*time.Minute)
	body := collectImproveAdvancedSSE(t, api, vetTok, visitID, 10*time.Second)

	if !strings.Contains(body, `"state":"crew_warming"`) {
		t.Fatalf("missing crew_warming: %s", truncate(body, 600))
	}
	if !strings.Contains(body, "crewai_unavailable") {
		t.Fatalf("missing crewai_unavailable error: %s", truncate(body, 800))
	}
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
