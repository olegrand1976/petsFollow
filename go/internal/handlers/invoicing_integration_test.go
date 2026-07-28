package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

func TestInvoicingConnectAndSendMock(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("inv-billit")
	password := "TestPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Inv", "practiceName": "Cabinet Inv",
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
	if access == "" {
		t.Fatal("missing accessToken")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if practiceID == "" {
		t.Fatalf("missing practiceId %#v", env)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE practice.practices
		SET company_legal_name = 'Cabinet Inv SPRL',
		    vat_number = 'BE0123456789',
		    company_number = '0123456789',
		    contact_email = $2
		WHERE id = $1`, practiceID, email); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connect/start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	if state == "" {
		t.Fatalf("missing state %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_test_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("connect/complete %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.ConnActive) {
		t.Fatalf("expected active %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{"name": "Client", "vatNumber": "BE0999999999"},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 5000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create doc %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)
	if docID == "" {
		t.Fatalf("missing doc id %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("send %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("expected delivered %#v", env)
	}
}

func TestInvoicingGetConnectionHidesPartyIDWithoutSettings(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("inv-party")
	password := "TestPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Party", "practiceName": "Cabinet Party",
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
	practiceID, _ := dataMap(t, env)["practiceId"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE practice.practices
		SET company_legal_name = 'Cabinet Party SPRL',
		    vat_number = 'BE0123456789',
		    company_number = '0123456789',
		    contact_email = $2
		WHERE id = $1`, practiceID, email); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code == http.StatusNotFound {
		t.Skip("BILLIT_ENABLED off")
	}
	if code != http.StatusOK {
		t.Fatalf("connect/start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_secret_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("connect/complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/connection", access, nil)
	if code != http.StatusOK {
		t.Fatalf("ref GET connection %d %#v", code, env)
	}
	if got, _ := dataMap(t, env)["billitPartyId"].(string); got != "party_secret_1" {
		t.Fatalf("ref should see partyId, got %#v", env)
	}

	secEmail := uniqueEmail("inv-sec")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/team", access, map[string]any{
		"email": secEmail, "fullName": "Sec Party", "password": password, "teamRole": "secretary",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("invite secretary %d %#v", code, env)
	}
	secTok := loginToken(t, api.handler, secEmail, password)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/connection", secTok, nil)
	if code != http.StatusOK {
		t.Fatalf("secretary GET connection %d %#v", code, env)
	}
	if got, _ := dataMap(t, env)["billitPartyId"].(string); got != "" {
		t.Fatalf("secretary must not see partyId, got %q %#v", got, env)
	}
}

