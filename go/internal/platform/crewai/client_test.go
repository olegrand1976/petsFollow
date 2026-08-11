package crewai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/crewai"
)

func TestClientHealthAndSubmit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(crewai.SecretHeader) != "sekrit" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body crewai.SubmitRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Workflow != crewai.WorkflowSmoke {
			t.Fatalf("workflow %q", body.Workflow)
		}
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{
			TaskID: "t1", Status: "completed", Workflow: body.Workflow,
			Result: map[string]any{"ok": true},
			Steps:  []crewai.Step{{Agent: "smoke", Label: "Ping"}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	out, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{
		Workflow: crewai.WorkflowSmoke, Tenant: "platform", Payload: map[string]any{"x": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "completed" || out.TaskID != "t1" {
		t.Fatalf("%+v", out)
	}
}

func TestClientUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	c := &crewai.Client{BaseURL: srv.URL, Secret: "wrong"}
	_, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{Workflow: crewai.WorkflowSmoke})
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("got %v", err)
	}
}

func TestClientNotConfigured(t *testing.T) {
	c := &crewai.Client{}
	if c.Configured() {
		t.Fatal("empty client must not be configured")
	}
	if err := c.Health(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	c = &crewai.Client{BaseURL: "https://crew.example"}
	if c.Configured() {
		t.Fatal("URL without auth must not be configured")
	}
	c = &crewai.Client{BaseURL: "https://crew.example", Secret: "sekrit"}
	if !c.Configured() {
		t.Fatal("URL+secret must be configured")
	}
	c = &crewai.Client{BaseURL: "https://crew.example", UseIDToken: true}
	if !c.Configured() {
		t.Fatal("URL+UseIDToken must be configured")
	}
}

func TestClientSendsAPIVersionAndSecret(t *testing.T) {
	var gotHealth, gotSubmit http.Header
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		gotHealth = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		gotSubmit = r.Header.Clone()
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{TaskID: "t", Status: "completed"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{Workflow: crewai.WorkflowSmoke}); err != nil {
		t.Fatal(err)
	}
	for name, h := range map[string]http.Header{"health": gotHealth, "submit": gotSubmit} {
		if h.Get(crewai.APIVersionHeader) != crewai.APIVersion {
			t.Fatalf("%s: api version %q", name, h.Get(crewai.APIVersionHeader))
		}
		if h.Get(crewai.SecretHeader) != "sekrit" {
			t.Fatalf("%s: secret %q", name, h.Get(crewai.SecretHeader))
		}
	}
}

func TestWaitReadyColdStartThenOK(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	coldStarts := 0
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	if err := c.WaitReady(context.Background(), 10*time.Second, func() { coldStarts++ }); err != nil {
		t.Fatal(err)
	}
	if coldStarts != 1 {
		t.Fatalf("onColdStart calls = %d", coldStarts)
	}
	if calls < 2 {
		t.Fatalf("health calls = %d", calls)
	}
}

func TestWaitReadyWarmSkipsColdStart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	coldStarts := 0
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	if err := c.WaitReady(context.Background(), 5*time.Second, func() { coldStarts++ }); err != nil {
		t.Fatal(err)
	}
	if coldStarts != 0 {
		t.Fatalf("warm instance must not report cold start (%d)", coldStarts)
	}
}

func TestWaitReadyAuthErrorFailsFast(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	coldStarts := 0
	c := &crewai.Client{BaseURL: srv.URL, Secret: "wrong"}
	startedAt := time.Now()
	err := c.WaitReady(context.Background(), 30*time.Second, func() { coldStarts++ })
	if err == nil || !strings.Contains(err.Error(), "crewai_health_403") {
		t.Fatalf("got %v", err)
	}
	if coldStarts != 0 {
		t.Fatalf("auth error must not report cold start (%d)", coldStarts)
	}
	if time.Since(startedAt) > 5*time.Second {
		t.Fatal("auth error must fail fast, not burn the budget")
	}
}

func TestWaitReadyBudgetExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	err := c.WaitReady(context.Background(), 300*time.Millisecond, nil)
	if err == nil || !strings.Contains(err.Error(), "crewai_unavailable") {
		t.Fatalf("got %v", err)
	}
}

func TestWaitReadyNotConfigured(t *testing.T) {
	c := &crewai.Client{}
	if err := c.WaitReady(context.Background(), time.Second, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientCancelTask(t *testing.T) {
	var gotPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/cancel") {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get(crewai.SecretHeader) != "sekrit" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"taskId":"t-cancel","status":"cancelled"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	if err := c.CancelTask(context.Background(), "t-cancel"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/tasks/t-cancel/cancel" {
		t.Fatalf("path %q", gotPath)
	}
	if err := c.CancelTask(context.Background(), ""); err == nil {
		t.Fatal("empty task id")
	}
}

func TestClientForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	_, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{Workflow: crewai.WorkflowSmoke})
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("got %v", err)
	}
}

func TestClientWorkflowRequired(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit"}
	_, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{})
	if err == nil || !strings.Contains(err.Error(), "workflow_required") {
		t.Fatalf("got %v", err)
	}
	if hit {
		t.Fatal("must not call HTTP when workflow empty")
	}
}
