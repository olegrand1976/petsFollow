package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MockGateway struct {
	WebhookSecret string
	APIPublicURL  string
}

func NewMockGateway(webhookSecret, apiPublicURL string) *MockGateway {
	if webhookSecret == "" {
		webhookSecret = "whsec_test"
	}
	if apiPublicURL == "" {
		apiPublicURL = "http://localhost:8291"
	}
	return &MockGateway{WebhookSecret: webhookSecret, APIPublicURL: apiPublicURL}
}

func (g *MockGateway) CreateCheckoutSession(_ context.Context, req CheckoutRequest) (CheckoutSession, error) {
	id := "cs_mock_" + uuid.NewString()
	q := url.Values{}
	for k, v := range req.Metadata {
		q.Set(k, v)
	}
	q.Set("session_id", id)
	checkoutURL := fmt.Sprintf("%s/api/v1/billing/dev/mock-complete?%s", strings.TrimRight(g.APIPublicURL, "/"), q.Encode())
	return CheckoutSession{ID: id, URL: checkoutURL}, nil
}

func (g *MockGateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (PortalSession, error) {
	return PortalSession{URL: fmt.Sprintf("%s/billing/portal/mock?customer=%s&return=%s",
		strings.TrimRight(g.APIPublicURL, "/"), customerID, url.QueryEscape(returnURL))}, nil
}

func (g *MockGateway) CancelSubscription(_ context.Context, subscriptionID string) error {
	return nil
}

func (g *MockGateway) VerifyWebhook(payload []byte, signature string) (StripeEvent, error) {
	if g.WebhookSecret != "" && !verifyStripeSignature(payload, signature, g.WebhookSecret) {
		return StripeEvent{}, fmt.Errorf("invalid webhook signature")
	}
	var envelope struct {
		ID   string         `json:"id"`
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return StripeEvent{}, err
	}
	if envelope.ID == "" {
		envelope.ID = "evt_mock_" + uuid.NewString()
	}
	return StripeEvent{ID: envelope.ID, Type: envelope.Type, Data: envelope.Data}, nil
}

func BuildTestWebhookPayload(secret, eventType string, data map[string]any) ([]byte, string, error) {
	eventID := "evt_test_" + uuid.NewString()
	body, err := json.Marshal(map[string]any{
		"id":   eventID,
		"type": eventType,
		"data": map[string]any{"object": data},
	})
	if err != nil {
		return nil, "", err
	}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	signed := fmt.Sprintf("%s.%s", ts, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	sig := hex.EncodeToString(mac.Sum(nil))
	header := fmt.Sprintf("t=%s,v1=%s", ts, sig)
	return body, header, nil
}

func verifyStripeSignature(payload []byte, signatureHeader, secret string) bool {
	if signatureHeader == "" {
		return false
	}
	var timestamp, sigV1 string
	for _, part := range strings.Split(signatureHeader, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			sigV1 = kv[1]
		}
	}
	if sigV1 == "" {
		return false
	}
	signed := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sigV1))
}

var _ Gateway = (*MockGateway)(nil)
