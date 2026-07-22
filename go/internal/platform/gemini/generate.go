package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GenerateJSON calls Gemini generateContent on the primary model and returns text (expect JSON).
func (c *Client) GenerateJSON(ctx context.Context, system, userPrompt string, temperature float64) (string, error) {
	return c.generateContent(ctx, c.Model, system, userPrompt, temperature, true)
}

// GenerateJSONLite uses the lite model (import / analyzer batch).
func (c *Client) GenerateJSONLite(ctx context.Context, system, userPrompt string, temperature float64) (string, error) {
	return c.generateContent(ctx, c.effectiveLite(), system, userPrompt, temperature, true)
}

// GenerateText is like GenerateJSON but without forcing JSON mime type.
func (c *Client) GenerateText(ctx context.Context, system, userPrompt string, temperature float64) (string, error) {
	return c.generateContent(ctx, c.Model, system, userPrompt, temperature, false)
}

func (c *Client) generateContent(ctx context.Context, model, system, userPrompt string, temperature float64, forceJSON bool) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("gemini_not_configured")
	}
	if model == "" {
		model = defaultModel
	}
	if temperature < 0 {
		if forceJSON {
			temperature = 0.4
		} else {
			temperature = 0.7
		}
	}
	body := map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": userPrompt}}},
		},
		"generationConfig": map[string]any{
			"temperature": temperature,
		},
	}
	if forceJSON {
		body["generationConfig"].(map[string]any)["responseMimeType"] = "application/json"
	}
	if strings.TrimSpace(system) != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]string{{"text": system}},
		}
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model, c.APIKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini_request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini_http_%d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}
	text, _, err := extractText(respBody)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// GenerateTextStream streams spoken text tokens via streamGenerateContent (SSE).
// onDelta receives each text fragment; return a non-nil error to abort.
func (c *Client) GenerateTextStream(ctx context.Context, system, userPrompt string, temperature float64, onDelta func(string) error) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("gemini_not_configured")
	}
	if temperature < 0 {
		temperature = 0.8
	}
	body := map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": userPrompt}}},
		},
		"generationConfig": map[string]any{
			"temperature": temperature,
		},
	}
	if strings.TrimSpace(system) != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]string{{"text": system}},
		}
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		c.Model, c.APIKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini_stream_request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini_http_%d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	// Chunks can be large; raise buffer.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		text, err := extractStreamDelta(payload)
		if err != nil || text == "" {
			continue
		}
		full.WriteString(text)
		if onDelta != nil {
			if err := onDelta(text); err != nil {
				return full.String(), err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return full.String(), ctx.Err()
		}
		return full.String(), fmt.Errorf("gemini_stream_read: %w", err)
	}
	return full.String(), nil
}

func extractStreamDelta(payload string) (string, error) {
	var envelope struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}
	if len(envelope.Candidates) == 0 {
		return "", nil
	}
	var b strings.Builder
	for _, p := range envelope.Candidates[0].Content.Parts {
		b.WriteString(p.Text)
	}
	return b.String(), nil
}
