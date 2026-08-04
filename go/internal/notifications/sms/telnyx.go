package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultTelnyxBaseURL = "https://api.telnyx.com"

// TelnyxClient appelle l'API Messages v2 de Telnyx (stdlib uniquement).
// DryRun retourne un succès synthétique sans I/O réseau.
type TelnyxClient struct {
	APIKey             string
	MessagingProfileID string
	From               string
	BaseURL            string // override httptest ; défaut api.telnyx.com
	DryRun             bool
	HTTPClient         *http.Client
}

type telnyxMessageRequest struct {
	To                 string `json:"to"`
	Text               string `json:"text"`
	MessagingProfileID string `json:"messaging_profile_id,omitempty"`
	From               string `json:"from,omitempty"`
}

type telnyxMessageResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *TelnyxClient) Send(ctx context.Context, to, text string) (Result, error) {
	if c.DryRun {
		log.Printf("sms dry-run: to=%s len=%d", to, len(text))
		return Result{ProviderMessageID: "dry_run", DryRun: true}, nil
	}
	if c.APIKey == "" {
		return Result{}, fmt.Errorf("telnyx_api_key_required")
	}
	body, err := json.Marshal(telnyxMessageRequest{
		To:                 to,
		Text:               text,
		MessagingProfileID: c.MessagingProfileID,
		From:               c.From,
	})
	if err != nil {
		return Result{}, err
	}
	base := c.BaseURL
	if base == "" {
		base = defaultTelnyxBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v2/messages", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return Result{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Result{}, fmt.Errorf("telnyx_http_%d", res.StatusCode)
	}
	var parsed telnyxMessageResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, fmt.Errorf("telnyx_bad_response: %w", err)
	}
	return Result{ProviderMessageID: parsed.Data.ID}, nil
}
