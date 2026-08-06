package billit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

// Client is the live Billit HTTP gateway.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// DefaultBaseURL is the Billit production API. Staging / local live pilotes must
// pass https://api.sandbox.billit.be explicitly (isolated env, separate ApiKeys).
const DefaultBaseURL = "https://api.billit.be"

// SandboxBaseURL is the isolated test API (no real Peppol / government traffic).
const SandboxBaseURL = "https://api.sandbox.billit.be"

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) EnsureParty(_ context.Context, practice invoicing.PracticeParty) (string, error) {
	return "", fmt.Errorf("billit EnsureParty: use reseller connect flow (practice %s)", practice.PracticeID)
}

func (c *Client) CheckParty(ctx context.Context, partyID, apiKey string) (invoicing.GatewayStatus, error) {
	// Billit OpenAPI: GET /v1/parties/{partyID} (singular /v1/party is 405 on sandbox).
	path := "/v1/parties/" + url.PathEscape(strings.TrimSpace(partyID))
	res, err := c.doJSON(ctx, http.MethodGet, path, partyID, apiKey, nil)
	if err != nil {
		return invoicing.GatewayStatus{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return invoicing.GatewayStatus{Complete: false, Message: fmt.Sprintf("http_%d", res.StatusCode)},
			fmt.Errorf("billit check party: unauthorized")
	}
	if res.StatusCode >= 300 {
		return invoicing.GatewayStatus{Complete: false, Message: fmt.Sprintf("http_%d", res.StatusCode)},
			fmt.Errorf("billit check party: %s", truncate(string(body), 200))
	}
	return parsePartyStatus(body), nil
}

func parsePartyStatus(body []byte) invoicing.GatewayStatus {
	// Fail-closed: unknown / empty Billit payloads must not mark KYC complete.
	st := invoicing.GatewayStatus{Complete: false, Message: "unknown_party_shape"}
	if len(bytes.TrimSpace(body)) == 0 {
		st.Message = "empty_party_payload"
		return st
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		st.Message = "invalid_party_json"
		return st
	}
	if complete, ok, msg := readCompleteness(raw); ok {
		st.Complete = complete
		if msg != "" {
			st.Message = msg
		} else if complete {
			st.Message = "ok"
		} else {
			st.Message = "incomplete"
		}
		return st
	}
	for _, key := range []string{"Party", "party", "Company", "company", "Data", "data"} {
		if nested, ok := raw[key].(map[string]any); ok {
			if complete, found, msg := readCompleteness(nested); found {
				st.Complete = complete
				if msg != "" {
					st.Message = msg
				} else if complete {
					st.Message = "ok"
				} else {
					st.Message = "incomplete"
				}
				return st
			}
			if partyLooksReady(nested) {
				st.Complete = true
				st.Message = "ok"
				return st
			}
		}
	}
	// GET /v1/parties/{id} returns Party fields without a Complete flag.
	if partyLooksReady(raw) {
		st.Complete = true
		st.Message = "ok"
		return st
	}
	return st
}

// partyLooksReady: Billit Party DTO with identity fields (sandbox/prod have no Complete bool).
func partyLooksReady(raw map[string]any) bool {
	_, hasID := raw["PartyID"]
	if !hasID {
		_, hasID = raw["PartyId"]
	}
	name, _ := raw["Name"].(string)
	vat, _ := raw["VATNumber"].(string)
	if vat == "" {
		vat, _ = raw["VatNumber"].(string)
	}
	return hasID && strings.TrimSpace(name) != "" && strings.TrimSpace(vat) != ""
}

func readCompleteness(raw map[string]any) (complete bool, found bool, message string) {
	for _, key := range []string{"Complete", "IsComplete", "OnboardingComplete", "complete", "isComplete"} {
		if v, ok := raw[key]; ok {
			switch t := v.(type) {
			case bool:
				return t, true, ""
			case string:
				low := strings.ToLower(strings.TrimSpace(t))
				if low == "true" || low == "yes" || low == "ok" {
					return true, true, ""
				}
				if low == "false" || low == "no" {
					return false, true, low
				}
			}
		}
	}
	if v, ok := raw["Completeness"]; ok {
		switch t := v.(type) {
		case float64:
			return t >= 100, true, fmt.Sprintf("completeness=%.0f", t)
		case json.Number:
			f, _ := t.Float64()
			return f >= 100, true, fmt.Sprintf("completeness=%s", t.String())
		}
	}
	if v, ok := raw["Status"].(string); ok {
		low := strings.ToLower(strings.TrimSpace(v))
		switch low {
		case "active", "complete", "completed", "ready":
			return true, true, low
		case "pending", "incomplete", "kyc", "pending_kyc":
			return false, true, low
		}
	}
	return false, false, ""
}

func (c *Client) CreateDocument(ctx context.Context, partyID, apiKey string, doc invoicing.Document) (string, error) {
	ord, err := MapDocument(doc)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(ord)
	if err != nil {
		return "", err
	}
	res, err := c.doJSON(ctx, http.MethodPost, "/v1/orders", partyID, apiKey, payload)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("billit create order: http_%d %s", res.StatusCode, truncate(string(body), 200))
	}
	return parseOrderID(body)
}

func (c *Client) SendPeppol(ctx context.Context, partyID, apiKey, externalID, country string) error {
	orderID, err := strconv.ParseInt(externalID, 10, 64)
	if err != nil {
		return fmt.Errorf("billit send: invalid order id %q", externalID)
	}
	cmd := SendCommand{OrderID: orderID, OrderIDs: []int64{orderID}}
	// Italy uses SDI; omit Transport so Billit routes from customer Identifiers.
	// Other supported markets: explicit Peppol.
	if strings.ToUpper(strings.TrimSpace(country)) != "IT" {
		cmd.Transport = "Peppol"
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	res, err := c.doJSON(ctx, http.MethodPost, "/v1/orders/commands/send", partyID, apiKey, payload)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return fmt.Errorf("billit send: http_%d %s", res.StatusCode, truncate(string(body), 200))
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path, partyID, apiKey string, payload []byte) (*http.Response, error) {
	// Only retry safe GETs — POST create/send must not be replayed (duplicate Billit orders).
	maxAttempts := 1
	if method == http.MethodGet {
		maxAttempts = 3
	}
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return nil, err
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("PartyID", partyID)
		req.Header.Set("ApiKey", apiKey)
		res, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt+1 < maxAttempts {
				time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
				continue
			}
			break
		}
		if maxAttempts > 1 && (res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500) {
			res.Body.Close()
			lastErr = fmt.Errorf("billit transient http_%d", res.StatusCode)
			time.Sleep(time.Duration(attempt+1) * 150 * time.Millisecond)
			continue
		}
		return res, nil
	}
	return nil, lastErr
}

func parseOrderID(body []byte) (string, error) {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return "", fmt.Errorf("billit create order: empty response")
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return strconv.FormatInt(n, 10), nil
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", fmt.Errorf("billit create order: cannot parse id from %s", truncate(s, 120))
	}
	for _, key := range []string{"OrderID", "orderId", "id", "Id"} {
		if v, ok := obj[key]; ok {
			return fmt.Sprint(v), nil
		}
	}
	return "", fmt.Errorf("billit create order: cannot parse id from %s", truncate(s, 120))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
