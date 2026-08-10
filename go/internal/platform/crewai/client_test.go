package crewai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		t.Fatal("expected not configured")
	}
	if err := c.Health(context.Background()); err == nil {
		t.Fatal("expected error")
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
