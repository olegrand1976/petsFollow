package crewai

import (
	"bytes"
	"context"
	"encoding/json"
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
}

// Step is a progress event from the orchestrator.
type Step struct {
	Agent string `json:"agent"`
	Label string `json:"label"`
	At    string `json:"at,omitempty"`
}

// SubmitResponse is the sync Phase-2 task result.
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

// SubmitTask POSTs /v1/tasks/submit.
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
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("crewai_http_%d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out SubmitResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
