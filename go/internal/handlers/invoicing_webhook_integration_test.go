package handlers_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

const testBillitWebhookSecret = "test-billit-webhook-secret"

func TestInvoicingWebhookDeliveredAndUsage(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)

	access, practiceID := registerInvoicingPractice(t, api, "inv-wh")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connect/start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_wh_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("connect/complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create doc %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	// Simulate live path: order already on Billit, awaiting network confirmation.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orderID := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = $2, status = 'sending', peppol_status = 'sending'
		WHERE id = $1 AND practice_id = $3`, docID, orderID, practiceID); err != nil {
		t.Fatal(err)
	}

	payload := map[string]any{
		"OrderID": orderID, "EventType": "OrderDelivered", "Status": "delivered",
	}
	raw, _ := json.Marshal(payload)
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, raw)
	if code != http.StatusOK {
		t.Fatalf("webhook %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusOK {
		t.Fatalf("get doc %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("expected delivered %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/connection", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connection %d %#v", code, env)
	}
	usage, _ := dataMap(t, env)["usageThisMonth"].(float64)
	if usage < 1 {
		t.Fatalf("expected usage >= 1 got %#v", env)
	}

	// Replay → duplicate ack (idempotent).
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, raw)
	if code != http.StatusOK {
		t.Fatalf("replay %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "duplicate" {
		t.Fatalf("expected duplicate %#v", env)
	}
}

func TestInvoicingWebhookStatusProgression(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)
	access, practiceID := registerInvoicingPractice(t, api, "inv-wh-prog")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_wh_prog", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orderID := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = $2, status = 'sending', peppol_status = 'order_created'
		WHERE id = $1 AND practice_id = $3`, docID, orderID, practiceID); err != nil {
		t.Fatal(err)
	}

	rawSent, _ := json.Marshal(map[string]any{
		"OrderID": orderID, "EventType": "OrderSent", "Status": "sending",
	})
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, rawSent)
	if code != http.StatusOK {
		t.Fatalf("webhook sending %d %#v", code, env)
	}

	rawDelivered, _ := json.Marshal(map[string]any{
		"OrderID": orderID, "EventType": "OrderDelivered", "Status": "delivered",
	})
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, rawDelivered)
	if code != http.StatusOK {
		t.Fatalf("webhook delivered %d %#v", code, env)
	}
	if dataMap(t, env)["status"] == "duplicate" {
		t.Fatal("distinct status payload must not be treated as duplicate")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("expected delivered %#v", env)
	}
}

func TestInvoicingWebhookDoesNotReopenRejected(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)
	access, practiceID := registerInvoicingPractice(t, api, "inv-wh-rej")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_wh_rej", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orderID := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = $2, status = 'rejected', peppol_status = 'send_failed'
		WHERE id = $1 AND practice_id = $3`, docID, orderID, practiceID); err != nil {
		t.Fatal(err)
	}

	rawSending, _ := json.Marshal(map[string]any{
		"OrderID": orderID, "EventType": "weird", "Status": "unknown-code",
	})
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, rawSending)
	if code != http.StatusOK {
		t.Fatalf("webhook %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusRejected) {
		t.Fatalf("rejected must not reopen to sending %#v", env)
	}
}

func TestInvoicingWebhookBadSignature(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)
	body := []byte(`{"OrderID":1,"Status":"delivered"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoicing/webhooks/billit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Billit-Signature", "00")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestInvoicingWebhookUnknownOrderRetries(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)
	body := []byte(`{"OrderID":999001,"EventType":"OrderDelivered","Status":"delivered"}`)
	code, _ := postBillitWebhook(t, api.handler, testBillitWebhookSecret, body)
	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 got %d", code)
	}
	// Second attempt still 503 (event forgotten → not duplicate).
	code, _ = postBillitWebhook(t, api.handler, testBillitWebhookSecret, body)
	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 replay got %d", code)
	}
}

func TestInvoicingWebhookMessageDeliveredAndUsage(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)

	access, practiceID := registerInvoicingPractice(t, api, "inv-wh-msg")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connect/start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_wh_msg", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("connect/complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create doc %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orderID := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = $2, status = 'sending', peppol_status = 'sending'
		WHERE id = $1 AND practice_id = $3`, docID, orderID, practiceID); err != nil {
		t.Fatal(err)
	}

	// Billit Access Point Message/U shape (staging webhook).
	orderIDNum, _ := strconv.ParseInt(orderID, 10, 64)
	payload := map[string]any{
		"UpdatedEntityID":     700001,
		"UpdatedEntityType":   "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": map[string]any{
			"OrderMessage": map[string]any{
				"OrderID":          orderIDNum,
				"Success":          true,
				"TransportType":    "Peppol",
				"MessageDirection": "Outgoing",
			},
			"AdditionalMessageInformation": map[string]any{
				"EInvoiceFlowState": "Delivered",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, raw)
	if code != http.StatusOK {
		t.Fatalf("message webhook %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusOK {
		t.Fatalf("get doc %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("expected delivered %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/connection", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connection %d %#v", code, env)
	}
	usage, _ := dataMap(t, env)["usageThisMonth"].(float64)
	if usage < 1 {
		t.Fatalf("expected usage >= 1 got %#v", env)
	}
}

func TestInvoicingDocumentsIsolatedAcrossPractices(t *testing.T) {
	api := newTestAPI(t)

	accessA, _ := registerInvoicingPractice(t, api, "inv-iso-a")
	accessB, _ := registerInvoicingPractice(t, api, "inv-iso-b")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", accessA, nil)
	if code != http.StatusOK {
		t.Fatalf("A start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", accessA, map[string]any{
		"state": state, "partyId": "party_iso_a", "apiKey": "mock-key-a",
	})
	if code != http.StatusOK {
		t.Fatalf("A complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", accessA, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client A", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("A create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	// Practice B must not see / get practice A documents (anti-redirection).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents", accessB, nil)
	if code != http.StatusOK && code != http.StatusNotFound {
		// Not connected yet → connection/docs may 404 invoicing_disabled or empty list.
		// Connect B minimally so list is authorized.
		_ = env
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", accessB, nil)
	if code != http.StatusOK {
		t.Fatalf("B start %d %#v", code, env)
	}
	stateB, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", accessB, map[string]any{
		"state": stateB, "partyId": "party_iso_b", "apiKey": "mock-key-b",
	})
	if code != http.StatusOK {
		t.Fatalf("B complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents", accessB, nil)
	if code != http.StatusOK {
		t.Fatalf("B list %d %#v", code, env)
	}
	if rows, ok := env["data"].([]any); ok {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["id"] == docID {
				t.Fatalf("practice B must not list practice A doc %#v", rows)
			}
		}
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, accessB, nil)
	if code != http.StatusNotFound {
		t.Fatalf("practice B GET A doc want 404 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", accessB, nil)
	if code != http.StatusNotFound {
		t.Fatalf("practice B send A doc want 404 got %d %#v", code, env)
	}
}

func registerInvoicingPractice(t *testing.T, api *testAPI, prefix string) (accessToken, practiceID string) {
	t.Helper()
	email := uniqueEmail(prefix)
	password := "TestPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Inv WH", "practiceName": "Cabinet WH",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	practiceID, _ = dataMap(t, env)["practiceId"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE practice.practices
		SET company_legal_name = 'Cabinet WH SPRL',
		    vat_number = 'BE1000000021',
		    company_number = '0123456789',
		    contact_email = $2
		WHERE id = $1`, practiceID, email); err != nil {
		t.Fatal(err)
	}
	return access, practiceID
}

func postBillitWebhook(t *testing.T, h http.Handler, secret string, body []byte) (int, map[string]any) {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoicing/webhooks/billit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Billit-Signature", sig)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}
