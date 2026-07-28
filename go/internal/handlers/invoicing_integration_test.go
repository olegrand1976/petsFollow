package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/store"
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
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
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

func TestInvoicingCreateDocumentWithVisitID(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)
	practiceID, _ := pets[0].(map[string]any)["practiceId"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         time.Now().UTC().Format(time.RFC3339),
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "invoice visit link",
	})
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO invoicing.practice_connections (practice_id, billit_party_id, status, api_secret_ref, connected_at)
		VALUES ($1, 'party_demo', 'active', 'mock-secret', now())
		ON CONFLICT (practice_id) DO UPDATE
		SET status = 'active', billit_party_id = EXCLUDED.billit_party_id, api_secret_ref = EXCLUDED.api_secret_ref`,
		practiceID); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":    "invoice",
		"visitId": visitID,
		"counterparty": map[string]any{
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult linked", "quantity": 1, "unitPriceExclCents": 4500, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create doc with visit %d %#v", code, env)
	}
	if dataMap(t, env)["visitId"] != visitID {
		t.Fatalf("visitId=%v want %s %#v", dataMap(t, env)["visitId"], visitID, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":    "invoice",
		"visitId": "00000000-0000-4000-8000-000000000099",
		"counterparty": map[string]any{
			"name": "X", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "bad", "quantity": 1, "unitPriceExclCents": 100, "vatPercent": 21},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("foreign visit want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":  "invoice",
		"dafId": "00000000-0000-4000-8000-000000000099",
		"counterparty": map[string]any{
			"name": "X", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "bad daf", "quantity": 1, "unitPriceExclCents": 100, "vatPercent": 21},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("foreign daf want 400 got %d %#v", code, env)
	}

	// Second walk-in + DAF linked to first visit → mismatch when invoicing visit2+daf1.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         time.Now().UTC().Format(time.RFC3339),
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "invoice visit other",
	})
	if code != http.StatusCreated {
		t.Fatalf("create visit2 %d %#v", code, env)
	}
	visit2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	dafID := uuid.NewString()
	dafNum := time.Now().UnixNano()%900000 + 100000
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pharmacy.daf_documents (
			id, practice_id, daf_year, daf_number, status, pet_id, visit_id, prescriber_user_id,
			issued_at, finalized_at
		) VALUES (
			$1::uuid, $2::uuid, 2099, $5, 'finalized', $3::uuid, $4::uuid,
			(SELECT id FROM identity.users WHERE email = 'vet.demo@petsfollow.test' LIMIT 1),
			now(), now()
		)`, dafID, practiceID, petID, visitID, dafNum); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM pharmacy.daf_documents WHERE id = $1::uuid`, dafID)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":    "invoice",
		"visitId": visit2,
		"dafId":   dafID,
		"counterparty": map[string]any{
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "mismatch", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("daf/visit mismatch want 400 got %d %#v", code, env)
	}

	// Draft DAF rejected even with matching visit.
	draftID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pharmacy.daf_documents (
			id, practice_id, daf_year, status, pet_id, visit_id, prescriber_user_id
		) VALUES (
			$1::uuid, $2::uuid, 2099, 'draft', $3::uuid, $4::uuid,
			(SELECT id FROM identity.users WHERE email = 'vet.demo@petsfollow.test' LIMIT 1)
		)`, draftID, practiceID, petID, visitID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM pharmacy.daf_documents WHERE id = $1::uuid`, draftID)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":    "invoice",
		"visitId": visitID,
		"dafId":   draftID,
		"counterparty": map[string]any{
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "draft", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("draft daf want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", vetTok, map[string]any{
		"type":    "invoice",
		"visitId": visitID,
		"dafId":   dafID,
		"counterparty": map[string]any{
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "ok linked", "quantity": 1, "unitPriceExclCents": 2200, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("matching daf+visit want 201 got %d %#v", code, env)
	}
	if dataMap(t, env)["dafId"] != dafID {
		t.Fatalf("dafId=%v want %s", dataMap(t, env)["dafId"], dafID)
	}
}

func TestInvoicingProformaResendRefused(t *testing.T) {
	api := newTestAPI(t)
	access, _ := registerInvoicingPractice(t, api, "inv-pf")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_pf_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "proforma",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Devis", "quantity": 1, "unitPriceExclCents": 2000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("send %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusIssued) {
		t.Fatalf("want issued %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", access, nil)
	if code != http.StatusConflict {
		t.Fatalf("resend want 409 got %d %#v", code, env)
	}
}

func TestInvoicingQuotaExceeded(t *testing.T) {
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-quota")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_quota_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.practice_connections SET docs_included_monthly = 1 WHERE practice_id = $1`, practiceID); err != nil {
		t.Fatal(err)
	}

	create := func() string {
		t.Helper()
		c, e := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
			"type": "invoice",
			"counterparty": map[string]any{
				"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
				"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
			},
			"lines": []map[string]any{
				{"description": "Consult", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
			},
		})
		if c != http.StatusCreated {
			t.Fatalf("create %d %#v", c, e)
		}
		id, _ := dataMap(t, e)["id"].(string)
		return id
	}

	id1 := create()
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+id1+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("send1 %d %#v", code, env)
	}

	id2 := create()
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+id2+"/send", access, nil)
	if code != http.StatusConflict {
		t.Fatalf("send2 quota want 409 got %d %#v", code, env)
	}
}

func TestInvoicingQuotaCountsInFlightSending(t *testing.T) {
	// Live-like: a doc already in sending occupies the monthly slot even if usage_monthly is still 0.
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-quota-send")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_quota_s", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.practice_connections SET docs_included_monthly = 1 WHERE practice_id = $1`, practiceID); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "InFlight", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "A", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create inflight %d %#v", code, env)
	}
	inflightID, _ := dataMap(t, env)["id"].(string)
	inflightOrder := "ord_inflight_" + uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'sending', billit_order_id = $3, peppol_status = 'sending'
		WHERE id = $1 AND practice_id = $2`, inflightID, practiceID, inflightOrder); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Second", "country": "BE", "vatNumber": "BE0888888888",
			"street": "Rue 2", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "B", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create second %d %#v", code, env)
	}
	secondID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+secondID+"/send", access, nil)
	if code != http.StatusConflict {
		t.Fatalf("quota with inflight sending want 409 got %d %#v", code, env)
	}
}

func TestInvoicingRejectedCanRetry(t *testing.T) {
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-rej")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_rej_1", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'rejected', peppol_status = 'send_failed', billit_order_id = $3
		WHERE id = $1 AND practice_id = $2`, docID, practiceID, "ord_rej_"+uuid.NewString()); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("retry rejected %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("want delivered after retry %#v", env)
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

func TestInvoicingCreditNoteRequiresRelatedInvoice(t *testing.T) {
	api := newTestAPI(t)
	access, _ := registerInvoicingPractice(t, api, "inv-cn")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_cn", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	cp := map[string]any{
		"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
		"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
	}
	lines := []map[string]any{
		{"description": "X", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "credit_note", "counterparty": cp, "lines": lines,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("CN without related want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice", "counterparty": cp, "lines": lines,
	})
	if code != http.StatusCreated {
		t.Fatalf("invoice %d %#v", code, env)
	}
	invID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "credit_note", "relatedDocumentId": invID, "counterparty": cp, "lines": lines,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("CN on draft invoice want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+invID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("send invoice %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "credit_note", "relatedDocumentId": invID, "counterparty": cp, "lines": lines,
	})
	if code != http.StatusCreated {
		t.Fatalf("CN with related %d %#v", code, env)
	}
}

func TestInvoicingStaleSendingRejected(t *testing.T) {
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-stale")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_stale", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE0999999999",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'sending', billit_order_id = $2, peppol_status = 'sending',
		    updated_at = now() - interval '8 days'
		WHERE id = $1`, docID, "ord_stale_"+uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	st := store.New(api.pool)
	n, err := st.MarkStaleSendingDocuments(ctx, time.Now().Add(-7*24*time.Hour), 50)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected >=1 stale reject got %d", n)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusRejected) {
		t.Fatalf("want rejected %#v", env)
	}
	_ = practiceID
}

func TestInvoicingAdminMarkPartner(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("inv-admin-mp")
	password := "TestPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Inv Admin", "practiceName": "Cabinet Admin MP",
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
		SET company_legal_name = 'Cabinet Admin MP SPRL',
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
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_admin_mp", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/invoicing/connections", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin list %d %#v", code, env)
	}
	foundName := false
	if rows, ok := env["data"].([]any); ok {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["practiceId"] == practiceID {
				if name, _ := m["practiceName"].(string); name == "" {
					t.Fatalf("expected practiceName %#v", m)
				}
				foundName = true
			}
		}
	}
	if !foundName {
		t.Fatalf("practice not in admin list %#v", env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/invoicing/connections/"+practiceID+"/mark-partner-invoiced", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("mark partner %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/connection", access, nil)
	if code != http.StatusOK {
		t.Fatalf("connection %d %#v", code, env)
	}
	if dataMap(t, env)["partnerListedAt"] == nil {
		t.Fatalf("expected partnerListedAt %#v", env)
	}
}

func TestInvoicingAdminMarkPartnerNotEligible(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("inv-admin-mp-ne")
	password := "TestPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Inv Admin NE", "practiceName": "Cabinet Admin NE",
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
		SET company_legal_name = 'Cabinet Admin NE SPRL',
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
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_admin_ne", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.practice_connections SET status = 'pending_kyc' WHERE practice_id = $1`, practiceID); err != nil {
		t.Fatal(err)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/invoicing/connections/"+practiceID+"/mark-partner-invoiced", adminTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("want 409 partner_mark_not_eligible got %d %#v", code, env)
	}
	errBody, _ := env["error"].(map[string]any)
	if codeStr, _ := errBody["code"].(string); codeStr != "partner_mark_not_eligible" {
		t.Fatalf("want partner_mark_not_eligible %#v", env)
	}
}

func TestInvoicingAdminSaasDraft(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("inv-admin-saas")
	password := "TestPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Inv SaaS", "practiceName": "Cabinet SaaS",
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
		SET company_legal_name = 'Cabinet SaaS SPRL',
		    vat_number = 'BE0123456789',
		    company_number = '0123456789',
		    contact_email = $2,
		    address_line1 = 'Rue Demo 1',
		    city = 'Bruxelles',
		    postal_code = '1000'
		WHERE id = $1`, practiceID, email); err != nil {
		t.Fatal(err)
	}

	// Connect optional for Flux A — list admin still needs a row; create connection for UI path.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code == http.StatusNotFound {
		t.Skip("BILLIT_ENABLED off")
	}
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_saas", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/invoicing/connections", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin list %d %#v", code, env)
	}
	saasEnabled := false
	if rows, ok := env["data"].([]any); ok {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["practiceId"] == practiceID {
				saasEnabled, _ = m["saasDraftEnabled"].(bool)
			}
		}
	}
	if !saasEnabled {
		t.Fatalf("expected saasDraftEnabled on list %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/invoicing/connections/"+practiceID+"/saas-draft", adminTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("saas draft %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if doc["source"] != "saas_master" {
		t.Fatalf("want source saas_master %#v", doc)
	}
	if doc["status"] != "draft" {
		t.Fatalf("want draft %#v", doc)
	}
	if doc["billitOrderId"] == nil || doc["billitOrderId"] == "" {
		t.Fatalf("want billitOrderId %#v", doc)
	}
	if int(doc["totalExclCents"].(float64)) != 8800 {
		t.Fatalf("want 8800 HT %#v", doc)
	}
	// Idempotent same month
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/invoicing/connections/"+practiceID+"/saas-draft", adminTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("saas draft 2 %d %#v", code, env)
	}
	if dataMap(t, env)["id"] != doc["id"] {
		t.Fatalf("want same idempotent id %#v vs %#v", env, doc)
	}

	// Hidden from practice document list
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents", access, nil)
	if code != http.StatusOK {
		t.Fatalf("list docs %d %#v", code, env)
	}
	if rows, ok := env["data"].([]any); ok && len(rows) > 0 {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["id"] == doc["id"] {
				t.Fatalf("saas doc must not appear in practice list %#v", rows)
			}
		}
	}
}
