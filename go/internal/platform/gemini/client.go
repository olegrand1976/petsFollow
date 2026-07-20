package gemini

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

const defaultModel = "gemini-3.5-flash"

type Mapper interface {
	Configured() bool
	SuggestColumnMapping(ctx context.Context, headers []string, sampleRows []map[string]string) (*MappingSuggestion, error)
}

type Client struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type MappingSuggestion struct {
	Email      *string         `json:"email"`
	FullName   *string         `json:"fullName"`
	Locale     *string         `json:"locale"`
	Ignored    []string        `json:"ignored,omitempty"`
	Confidence float64         `json:"confidence"`
	Raw        json.RawMessage `json:"-"`
}

func New(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) Configured() bool {
	return c != nil && strings.TrimSpace(c.APIKey) != ""
}

func (c *Client) SuggestColumnMapping(ctx context.Context, headers []string, sampleRows []map[string]string) (*MappingSuggestion, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	payload := map[string]any{
		"headers":    headers,
		"sampleRows": sampleRows,
		"targets":    []string{"email", "fullName", "locale"},
	}
	payloadJSON, _ := json.Marshal(payload)
	prompt := `You map veterinary clinic client list spreadsheet columns to petsFollow fields.
Return ONLY valid JSON (no markdown) with this shape:
{"email":"<exact header or null>","fullName":"<exact header or null>","locale":"<exact header or null>","ignored":["..."],"confidence":0.0}
Rules:
- email: column that looks like an email address
- fullName: client display name (may be split first/last — prefer the best single column for full name)
- locale: language/locale if present (fr/nl/en/es), else null
- Use exact header strings from the input headers list
- confidence between 0 and 1
Input:
` + string(payloadJSON)

	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":      0.1,
			"responseMimeType": "application/json",
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.Model, c.APIKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini_request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini_http_%d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}

	text, raw, err := extractText(respBody)
	if err != nil {
		return nil, err
	}
	sug, err := ParseSuggestion(text)
	if err != nil {
		return nil, err
	}
	sug.Raw = raw
	normalizeSuggestion(sug, headers)
	return sug, nil
}

func ParseSuggestion(text string) (*MappingSuggestion, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var sug MappingSuggestion
	if err := json.Unmarshal([]byte(text), &sug); err != nil {
		return nil, fmt.Errorf("gemini_parse: %w", err)
	}
	return &sug, nil
}

func normalizeSuggestion(sug *MappingSuggestion, headers []string) {
	set := map[string]struct{}{}
	for _, h := range headers {
		set[h] = struct{}{}
	}
	sug.Email = keepHeader(sug.Email, set)
	sug.FullName = keepHeader(sug.FullName, set)
	sug.Locale = keepHeader(sug.Locale, set)
	if sug.Confidence < 0 {
		sug.Confidence = 0
	}
	if sug.Confidence > 1 {
		sug.Confidence = 1
	}
}

func keepHeader(p *string, set map[string]struct{}) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" || v == "null" {
		return nil
	}
	if _, ok := set[v]; !ok {
		return nil
	}
	return &v
}

func extractText(respBody []byte) (string, json.RawMessage, error) {
	var envelope struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return "", nil, fmt.Errorf("gemini_envelope: %w", err)
	}
	if len(envelope.Candidates) == 0 || len(envelope.Candidates[0].Content.Parts) == 0 {
		return "", nil, fmt.Errorf("gemini_empty")
	}
	text := envelope.Candidates[0].Content.Parts[0].Text
	return text, json.RawMessage(respBody), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
