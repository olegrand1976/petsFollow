package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		    vat_number = 'BE1000000021',
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
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "X", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "X", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "Sophie Demo", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
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
				"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "InFlight", "country": "BE", "vatNumber": "BE1000000021",
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
			"name": "Second", "country": "BE", "vatNumber": "BE1000000120",
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

func TestInvoicingQuotaIgnoresSaasMasterSending(t *testing.T) {
	// Flux A docs in sending must not occupy practice Peppol quota slots.
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-quota-saas")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_quota_saas", "apiKey": "mock-key",
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
	saasID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO invoicing.documents (
			id, practice_id, type, status, source, billit_order_id, idempotency_key,
			counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
			peppol_status, created_at, updated_at
		) VALUES (
			$1, $2, 'invoice', 'sending', 'saas_master', $3, $4,
			'{"name":"LL-IT-SC","country":"BE"}'::jsonb, 'EUR', 8800, 1848, 10648,
			'sending', now(), now()
		)`, saasID, practiceID, "ord_saas_inflight_"+uuid.NewString(), "saas-quota-"+uuid.NewString()); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice",
		"counterparty": map[string]any{
			"name": "Practice", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Consult", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create practice %d %#v", code, env)
	}
	practiceDocID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+practiceDocID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("practice send with saas_master sending must succeed, got %d %#v", code, env)
	}
}

func TestInvoicingCreditNoteRejectsSaasMasterRelated(t *testing.T) {
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-cn-saas")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_cn_saas", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	saasID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO invoicing.documents (
			id, practice_id, type, status, source, billit_order_id, idempotency_key,
			counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
			peppol_status, created_at, updated_at
		) VALUES (
			$1, $2, 'invoice', 'delivered', 'saas_master', $3, $4,
			'{"name":"Cabinet","country":"BE","vatNumber":"BE0123456749"}'::jsonb, 'EUR', 8800, 1848, 10648,
			'delivered', now(), now()
		)`, saasID, practiceID, "ord_cn_saas_"+uuid.NewString(), "saas-cn-"+uuid.NewString()); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type":              "credit_note",
		"relatedDocumentId": saasID,
		"counterparty": map[string]any{
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
			"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
		},
		"lines": []map[string]any{
			{"description": "Avoir", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("CN on saas_master want 400 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "related_document_invalid" {
		t.Fatalf("want related_document_invalid %#v", env)
	}
}

func TestInvoicingStaleSendingSkipsSaasMaster(t *testing.T) {
	api := newTestAPI(t)
	_, practiceID := registerInvoicingPractice(t, api, "inv-stale-saas")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	saasID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO invoicing.documents (
			id, practice_id, type, status, source, billit_order_id, idempotency_key,
			counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
			peppol_status, created_at, updated_at
		) VALUES (
			$1, $2, 'invoice', 'sending', 'saas_master', $3, $4,
			'{"name":"Cabinet","country":"BE"}'::jsonb, 'EUR', 8800, 1848, 10648,
			'sending', now() - interval '8 days', now() - interval '8 days'
		)`, saasID, practiceID, "ord_stale_saas_"+uuid.NewString(), "saas-stale-"+uuid.NewString()); err != nil {
		t.Fatal(err)
	}

	st := store.New(api.pool)
	n, err := st.MarkStaleSendingDocuments(ctx, time.Now().Add(-7*24*time.Hour), 50)
	if err != nil {
		t.Fatal(err)
	}
	var status, source string
	if err := api.pool.QueryRow(ctx, `
		SELECT status, COALESCE(source, 'practice') FROM invoicing.documents WHERE id = $1`, saasID,
	).Scan(&status, &source); err != nil {
		t.Fatal(err)
	}
	if status != string(invoicing.StatusSending) {
		t.Fatalf("saas_master must stay sending, got %q (mark returned %d)", status, n)
	}
	if source != string(invoicing.SourceSaasMaster) {
		t.Fatalf("source=%q", source)
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
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
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
		    vat_number = 'BE1000000021',
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
		"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
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
	cnID, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+cnID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("CN send after invoice on Billit %d %#v", code, env)
	}
}

func TestInvoicingCreditNoteSendRequiresRelatedOnBillit(t *testing.T) {
	api := newTestAPI(t)
	access, practiceID := registerInvoicingPractice(t, api, "inv-cn-nobillit")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_cn_nb", "apiKey": "mock-key",
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	cp := map[string]any{
		"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
		"street": "Rue 1", "city": "Bruxelles", "postal": "1000",
	}
	lines := []map[string]any{
		{"description": "X", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "invoice", "counterparty": cp, "lines": lines,
	})
	if code != http.StatusCreated {
		t.Fatalf("invoice %d %#v", code, env)
	}
	invID, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+invID+"/send", access, nil)
	if code != http.StatusOK {
		t.Fatalf("send %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents", access, map[string]any{
		"type": "credit_note", "relatedDocumentId": invID, "counterparty": cp, "lines": lines,
	})
	if code != http.StatusCreated {
		t.Fatalf("CN %d %#v", code, env)
	}
	cnID, _ := dataMap(t, env)["id"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = NULL, number = NULL, status = 'delivered'
		WHERE id = $1 AND practice_id = $2`, invID, practiceID); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+cnID+"/send", access, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("CN send without Billit invoice want 400 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["msgKey"] != "related_invoice_not_on_billit" {
		t.Fatalf("msgKey %#v", env)
	}
}

func TestInvoicingSecretsMismatchOnSend(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetBillitSecretsBackend("local_enc")
	api.api.TestSetBillitSecretsKey("original-billit-secrets-key!!")

	access, _ := registerInvoicingPractice(t, api, "inv-sec-rot")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/start", access, nil)
	if code != http.StatusOK {
		t.Fatalf("start %d %#v", code, env)
	}
	state, _ := dataMap(t, env)["state"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/connect/complete", access, map[string]any{
		"state": state, "partyId": "party_sec", "apiKey": "mock-key-live",
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
			{"description": "X", "quantity": 1, "unitPriceExclCents": 1000, "vatPercent": 21},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	docID, _ := dataMap(t, env)["id"].(string)

	// Simulate BILLIT_SECRETS_KEY rotation after connect.
	api.api.TestSetBillitSecretsKey("rotated-billit-secrets-key!!!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/practices/me/invoicing/documents/"+docID+"/send", access, nil)
	if code != http.StatusConflict {
		t.Fatalf("want 409 secrets mismatch got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["msgKey"] != "invoicing_secrets_mismatch" {
		t.Fatalf("msgKey %#v", env)
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
			"name": "Client", "country": "BE", "vatNumber": "BE1000000021",
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
		    vat_number = 'BE1000000021',
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
		    vat_number = 'BE1000000021',
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
		    vat_number = 'BE1000000021',
		    company_number = '0123456789',
		    contact_email = $2,
		    address_line1 = 'Rue Demo 1',
		    city = 'Bruxelles',
		    postal_code = '1000',
		    profile_completed_at = COALESCE(profile_completed_at, NOW()),
		    saas_billing_enabled = TRUE
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

	// Hidden from practice document list + GET by id
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
	docID, _ := doc["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/practices/me/invoicing/documents/"+docID, access, nil)
	if code != http.StatusNotFound {
		t.Fatalf("saas GET by id want 404 got %d %#v", code, env)
	}

	// Webhook delivered on saas_master must not burn practice Peppol quota.
	api.api.TestSetBillitWebhookSecret(testBillitWebhookSecret)
	orderID, _ := doc["billitOrderId"].(string)
	var usageBefore int
	if err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(doc_count, 0) FROM invoicing.usage_monthly
		WHERE practice_id = $1 AND yyyymm = $2`,
		practiceID, time.Now().UTC().Year()*100+int(time.Now().UTC().Month()),
	).Scan(&usageBefore); err != nil {
		usageBefore = 0
	}
	raw, _ := json.Marshal(map[string]any{
		"OrderID": orderID, "EventType": "OrderDelivered", "Status": "delivered",
	})
	code, env = postBillitWebhook(t, api.handler, testBillitWebhookSecret, raw)
	if code != http.StatusOK {
		t.Fatalf("saas webhook %d %#v", code, env)
	}
	var usageAfter int
	if err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(doc_count, 0) FROM invoicing.usage_monthly
		WHERE practice_id = $1 AND yyyymm = $2`,
		practiceID, time.Now().UTC().Year()*100+int(time.Now().UTC().Month()),
	).Scan(&usageAfter); err != nil {
		usageAfter = 0
	}
	if usageAfter != usageBefore {
		t.Fatalf("saas delivered must not increment usage: before=%d after=%d", usageBefore, usageAfter)
	}

	// Option A: explicit admin send after draft (mock → delivered, still no quota burn).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/invoicing/saas-targets", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("saas-targets %d %#v", code, env)
	}
	var saasDocID string
	if rows, ok := env["data"].([]any); ok {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["practiceId"] != practiceID {
				continue
			}
			sd, _ := m["saasDocument"].(map[string]any)
			if sd == nil {
				t.Fatalf("expected saasDocument on targets after draft %#v", m)
			}
			saasDocID, _ = sd["id"].(string)
		}
	}
	if saasDocID == "" {
		t.Fatal("missing saasDocument id")
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'draft', peppol_status = 'saas_draft', sent_at = NULL
		WHERE id = $1`, saasDocID); err != nil {
		t.Fatal(err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/invoicing/connections/"+practiceID+"/saas-documents/"+saasDocID+"/send", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("saas send %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != string(invoicing.StatusDelivered) {
		t.Fatalf("want delivered after saas send %#v", env)
	}
	var usageAfterSend int
	if err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(doc_count, 0) FROM invoicing.usage_monthly
		WHERE practice_id = $1 AND yyyymm = $2`,
		practiceID, time.Now().UTC().Year()*100+int(time.Now().UTC().Month()),
	).Scan(&usageAfterSend); err != nil {
		usageAfterSend = 0
	}
	if usageAfterSend != usageBefore {
		t.Fatalf("saas send must not increment usage: before=%d after=%d", usageBefore, usageAfterSend)
	}
}

func TestInvoicingAdminSaasTargetsAndCron(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/invoicing/saas-targets", adminTok, nil)
	if code == http.StatusNotFound {
		t.Skip("BILLIT_ENABLED off")
	}
	if code != http.StatusOK {
		t.Fatalf("saas-targets %d %#v", code, env)
	}
	rows, ok := env["data"].([]any)
	if !ok || len(rows) == 0 {
		t.Fatalf("expected seeded BE practices as saas targets %#v", env)
	}
	var practiceID string
	for _, row := range rows {
		m, _ := row.(map[string]any)
		if en, _ := m["saasDraftEnabled"].(bool); en {
			practiceID, _ = m["practiceId"].(string)
			break
		}
	}
	if practiceID == "" {
		t.Fatalf("no saasDraftEnabled target %#v", env)
	}

	secret := "test-saas-invoices-secret"
	api.api.TestSetSaasInvoicesSecret(secret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/saas-invoices/run", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("cron without secret want 401 got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/internal/saas-invoices/run", nil)
	req.Header.Set("X-Saas-Invoices-Secret", secret)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cron with secret want 200 got %d %s", rec.Code, rec.Body.String())
	}
	var cronEnv map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&cronEnv); err != nil {
		t.Fatal(err)
	}
	data := dataMap(t, cronEnv)
	drafted, _ := data["drafted"].(float64)
	unchanged, _ := data["unchanged"].(float64)
	if drafted+unchanged < 1 {
		t.Fatalf("expected at least one draft/unchanged %#v", cronEnv)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/invoicing/saas-targets", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("saas-targets after cron %d %#v", code, env)
	}
	found := false
	if rows, ok := env["data"].([]any); ok {
		for _, row := range rows {
			m, _ := row.(map[string]any)
			if m["practiceId"] != practiceID {
				continue
			}
			sd, _ := m["saasDocument"].(map[string]any)
			if sd == nil || sd["id"] == nil {
				t.Fatalf("expected saasDocument after cron %#v", m)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("practice %s missing after cron %#v", practiceID, env)
	}
}
