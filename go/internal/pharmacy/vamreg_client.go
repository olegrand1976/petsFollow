package pharmacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VamregClient posts antibiotic declarations (or dry-runs).
// Live declaration HTTP contract is not covered by the FAMHP readonly ICD
// (v20260701 software-house GET lists) — see VamregAFMPSClient for reference lists.
type VamregClient struct {
	BaseURL    string
	APIKey     string
	DryRun     bool
	HTTPClient *http.Client
}

// VamregDeclareRequest is the payload sent to the authority gateway.
type VamregDeclareRequest struct {
	DAFID      string       `json:"dafId"`
	DAFNumber  string       `json:"dafNumber"`
	PracticeID string       `json:"practiceId"`
	Lines      []VamregLine `json:"lines"`
}

// VamregLine is one antibiotic line on a finalized DAF.
type VamregLine struct {
	MedicationCNK  string          `json:"medicationCnk"`
	MedicationName string          `json:"medicationName"`
	AMMNumber      string          `json:"ammNumber"`
	Qty            float64         `json:"qty"`
	Unit           string          `json:"unit"`
	Payload        json.RawMessage `json:"payload"`
}

// VamregDeclareResult captures the gateway outcome for audit.
type VamregDeclareResult struct {
	DryRun       bool            `json:"dryRun"`
	HTTPStatus   int             `json:"httpStatus,omitempty"`
	ResponseBody json.RawMessage `json:"responseBody,omitempty"`
}

// Declare sends the payload or returns a dry-run success without network I/O.
// Live mode (!DryRun) requires BaseURL — empty URL is an error, never a silent dry-run.
// Auth/path here are provisional until FAMHP provides the write ICD (readonly lists → VamregAFMPSClient).
func (c *VamregClient) Declare(ctx context.Context, req VamregDeclareRequest) (VamregDeclareResult, error) {
	if c == nil || c.DryRun {
		return VamregDeclareResult{
			DryRun:       true,
			HTTPStatus:   200,
			ResponseBody: json.RawMessage(`{"status":"dry_run_ok"}`),
		}, nil
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return VamregDeclareResult{}, fmt.Errorf("vamreg_base_url_required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return VamregDeclareResult{}, err
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	url := strings.TrimRight(c.BaseURL, "/") + "/v1/declarations"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return VamregDeclareResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Stable key = DAF id so Asynq retries do not create duplicate authority records.
	if id := strings.TrimSpace(req.DAFID); id != "" {
		httpReq.Header.Set("Idempotency-Key", id)
	}
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	res, err := httpClient.Do(httpReq)
	if err != nil {
		return VamregDeclareResult{}, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	out := VamregDeclareResult{HTTPStatus: res.StatusCode, ResponseBody: json.RawMessage(raw)}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return out, fmt.Errorf("vamreg_http_%d", res.StatusCode)
	}
	if len(out.ResponseBody) == 0 || !json.Valid(out.ResponseBody) {
		out.ResponseBody = json.RawMessage(`{}`)
	}
	return out, nil
}
