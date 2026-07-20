package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GenerateJSON calls Gemini generateContent and returns the model text (expect JSON).
func (c *Client) GenerateJSON(ctx context.Context, system, userPrompt string, temperature float64) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("gemini_not_configured")
	}
	if temperature < 0 {
		temperature = 0.4
	}
	contents := []map[string]any{
		{"role": "user", "parts": []map[string]string{{"text": userPrompt}}},
	}
	body := map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"temperature":      temperature,
			"responseMimeType": "application/json",
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
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.Model, c.APIKey,
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

// GenerateText is like GenerateJSON but without forcing JSON mime type.
func (c *Client) GenerateText(ctx context.Context, system, userPrompt string, temperature float64) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("gemini_not_configured")
	}
	if temperature < 0 {
		temperature = 0.7
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
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.Model, c.APIKey,
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
