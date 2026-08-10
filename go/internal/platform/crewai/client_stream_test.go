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

func TestClientAsyncSubmitAndStream(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		var body crewai.SubmitRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !body.Async {
			t.Fatal("expected async")
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(crewai.SubmitResponse{TaskID: "task-1", Status: "running"})
	})
	mux.HandleFunc("/v1/tasks/task-1/events", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("id: 1\nevent: step\ndata: {\"agent\":\"tri\",\"label\":\"A\"}\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("id: 2\nevent: final\ndata: {\"report\":\"## History\\n\\nok\"}\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("event: ping\ndata: {\"done\":true}\n\n"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := &crewai.Client{BaseURL: srv.URL, Secret: "sekrit", HTTPClient: srv.Client()}
	out, err := c.SubmitTask(context.Background(), crewai.SubmitRequest{
		Workflow: crewai.WorkflowCRImprove, Tenant: crewai.TenantPetsFollow, Async: true,
		Payload: map[string]any{"sourceText": "x"},
	})
	if err != nil || out.TaskID != "task-1" {
		t.Fatalf("%v %+v", err, out)
	}
	ch := make(chan crewai.StreamEvent, 8)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.StreamEvents(ctx, out.TaskID, ch)
		close(ch)
	}()
	var types []string
	for ev := range ch {
		types = append(types, ev.Type)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(types, ",")
	if !strings.Contains(joined, "step") || !strings.Contains(joined, "final") {
		t.Fatalf("events %v", types)
	}
}
