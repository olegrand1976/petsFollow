package crewai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

const (
	APIVersionHeader  = "X-Crew-API-Version"
	APIVersion        = "1"
	SecretHeader      = "X-Crew-Secret"
	TenantPetsFollow  = "petsfollow"
	WorkflowSmoke     = "staging_smoke_test"
	WorkflowCRImprove = "petsfollow_cr_improve"
)

// Client talks to the shared LL-IT crewai-orchestrator Cloud Run service.
type Client struct {
	BaseURL string
	Secret  string
	// UseIDToken attaches a Google ID token (Cloud Run IAM invoker) when true.
	UseIDToken bool
	HTTPClient *http.Client
}

func (c *Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (c *Client) base() string {
	return strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
}

// Configured reports whether the client has a base URL.
func (c *Client) Configured() bool {
	return c != nil && c.base() != ""
}

func (c *Client) authorize(ctx context.Context, req *http.Request) error {
	req.Header.Set(APIVersionHeader, APIVersion)
	if s := strings.TrimSpace(c.Secret); s != "" {
		req.Header.Set(SecretHeader, s)
	}
	if !c.UseIDToken {
		return nil
	}
	ts, err := idtoken.NewTokenSource(ctx, c.base())
	if err != nil {
		return fmt.Errorf("crewai_id_token: %w", err)
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("crewai_id_token: %w", err)
	}
	tok.SetAuthHeader(req)
	return nil
}

// Health hits GET /health (IAM may still be required when UseIDToken).
// Note: /healthz is intercepted by Google edge on Cloud Run (HTML 404).
func (c *Client) Health(ctx context.Context) error {
	if !c.Configured() {
		return fmt.Errorf("crewai_not_configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+"/health", nil)
	if err != nil {
		return err
	}
	if err := c.authorize(ctx, req); err != nil {
		return err
	}
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("crewai_health_%d", res.StatusCode)
	}
	return nil
}

// SubmitRequest is the multi-app task envelope.
type SubmitRequest struct {
	Workflow      string         `json:"workflow"`
	Tenant        string         `json:"tenant"`
	CorrelationID string         `json:"correlationId"`
	Payload       map[string]any `json:"payload"`
	Async         bool           `json:"async,omitempty"`
}

// Step is a progress event from the orchestrator.
type Step struct {
	Agent string `json:"agent"`
	Label string `json:"label"`
	At    string `json:"at,omitempty"`
}

// SubmitResponse is the sync Phase-2 task result (or 202 running when Async).
type SubmitResponse struct {
	TaskID        string `json:"taskId"`
	Status        string `json:"status"`
	Result        any    `json:"result"`
	Steps         []Step `json:"steps"`
	ErrorCode     string `json:"errorCode,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
	Workflow      string `json:"workflow,omitempty"`
	Tenant        string `json:"tenant,omitempty"`
}

// StreamEvent is one SSE frame from /v1/tasks/{id}/events.
type StreamEvent struct {
	ID   string
	Type string
	Data json.RawMessage
}

// SubmitTask POSTs /v1/tasks/submit. With Async=true expects HTTP 202 + status=running.
func (c *Client) SubmitTask(ctx context.Context, in SubmitRequest) (*SubmitResponse, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("crewai_not_configured")
	}
	if strings.TrimSpace(in.Workflow) == "" {
		return nil, fmt.Errorf("workflow_required")
	}
	if in.Payload == nil {
		in.Payload = map[string]any{}
	}
	body, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/v1/tasks/submit", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.authorize(ctx, req); err != nil {
		return nil, err
	}
	res, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("crewai_unauthorized")
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("crewai_forbidden")
	}
	if in.Async {
		if res.StatusCode != http.StatusAccepted && (res.StatusCode < 200 || res.StatusCode >= 300) {
			return nil, fmt.Errorf("crewai_http_%d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
		}
	} else if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("crewai_http_%d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out SubmitResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StreamEvents reads SSE from GET /v1/tasks/{taskID}/events until done/ctx cancel.
func (c *Client) StreamEvents(ctx context.Context, taskID string, out chan<- StreamEvent) error {
	if !c.Configured() {
		return fmt.Errorf("crewai_not_configured")
	}
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("task_id_required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+"/v1/tasks/"+taskID+"/events", nil)
	if err != nil {
		return err
	}
	if err := c.authorize(ctx, req); err != nil {
		return err
	}
	// Long-lived stream — don't use the short default client timeout.
	hc := &http.Client{Timeout: 0}
	if c.HTTPClient != nil {
		hc = c.HTTPClient
	}
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("crewai_unauthorized")
	}
	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("crewai_task_not_found")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("crewai_http_%d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	return readSSE(ctx, res.Body, out)
}

func readSSE(ctx context.Context, r io.Reader, out chan<- StreamEvent) error {
	br := bufio.NewReader(r)
	var id, evType string
	var data strings.Builder
	flush := func() error {
		if evType == "" && data.Len() == 0 {
			return nil
		}
		payload := strings.TrimRight(data.String(), "\n")
		ev := StreamEvent{ID: id, Type: evType, Data: json.RawMessage(payload)}
		if len(ev.Data) == 0 {
			ev.Data = json.RawMessage("{}")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- ev:
		}
		id, evType = "", ""
		data.Reset()
		return nil
	}
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			switch {
			case line == "":
				if err := flush(); err != nil {
					return err
				}
			case strings.HasPrefix(line, "id:"):
				id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			case strings.HasPrefix(line, "event:"):
				evType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				if data.Len() > 0 {
					data.WriteByte('\n')
				}
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				_ = flush()
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}
