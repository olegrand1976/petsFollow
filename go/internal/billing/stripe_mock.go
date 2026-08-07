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
	// URLSecret signs mock checkout callbacks. The mock endpoints are reachable
	// without a bearer token (the browser follows the checkout redirect), so the
	// signature is what proves the URL was minted by us.
	URLSecret string
}

func NewMockGateway(webhookSecret, apiPublicURL, urlSecret string) *MockGateway {
	if webhookSecret == "" {
		webhookSecret = "whsec_test"
	}
	if apiPublicURL == "" {
		apiPublicURL = "http://localhost:8291"
	}
	return &MockGateway{WebhookSecret: webhookSecret, APIPublicURL: apiPublicURL, URLSecret: urlSecret}
}

func (g *MockGateway) CreateCheckoutSession(_ context.Context, req CheckoutRequest) (CheckoutSession, error) {
	id := "cs_mock_" + uuid.NewString()
	q := url.Values{}
	for k, v := range req.Metadata {
		q.Set(k, v)
	}
	q.Set("session_id", id)
	if req.SuccessURL != "" {
		q.Set("success_url", req.SuccessURL)
	}
	q.Set(MockSignatureParam, SignMockURL(g.URLSecret, q))
	checkoutURL := fmt.Sprintf("%s/api/v1/billing/dev/mock-complete?%s", strings.TrimRight(g.APIPublicURL, "/"), q.Encode())
	return CheckoutSession{ID: id, URL: checkoutURL}, nil
}

// MockSignatureParam is the query parameter carrying the mock URL signature.
const MockSignatureParam = "sig"

// mockURLDomain prefixes the signed payload so the MAC only ever authenticates
// mock billing URLs.
const mockURLDomain = "petsfollow/billing-mock-url/v1\n"

// SignMockURL authenticates a mock billing URL over all its other parameters.
// url.Values.Encode sorts by key, so the signed payload is canonical.
func SignMockURL(secret string, q url.Values) string {
	signed := url.Values{}
	for k, v := range q {
		if k == MockSignatureParam {
			continue
		}
		signed[k] = v
	}
	mac := hmac.New(sha256.New, []byte(secret))
	// Domain separation: the secret is also the JWT signing key, so the label
	// keeps this MAC from ever colliding with a token signature.
	_, _ = mac.Write([]byte(mockURLDomain))
	_, _ = mac.Write([]byte(signed.Encode()))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyMockURL reports whether q carries a signature we minted.
func VerifyMockURL(secret string, q url.Values) bool {
	got := q.Get(MockSignatureParam)
	if got == "" {
		return false
	}
	return hmac.Equal([]byte(SignMockURL(secret, q)), []byte(got))
}

// MockCustomerID mirrors the customer id written by MockCompleteCheckout webhooks.
func MockCustomerID(ownerUserID string) string {
	return "cus_mock_" + ownerUserID
}

func (g *MockGateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (PortalSession, error) {
	q := url.Values{}
	q.Set("customer", customerID)
	q.Set("return", returnURL)
	q.Set(MockSignatureParam, SignMockURL(g.URLSecret, q))
	return PortalSession{URL: fmt.Sprintf("%s/api/v1/billing/dev/mock-portal?%s",
		strings.TrimRight(g.APIPublicURL, "/"), q.Encode())}, nil
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
