package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

// Tests d'intégration HTTP billing : checkout à la création de pet, webhook Stripe
// signé (activation entitlement), signature invalide refusée, ACL entitlement.
// Prérequis : DB up + make seed (comptes démo).

func webhookSecret() string {
	if s := os.Getenv("STRIPE_WEBHOOK_SECRET"); s != "" {
		return s
	}
	return "whsec_test"
}

// createPendingPet crée un pet côté client avec checkout et renvoie (petID, sessionID).
func createPendingPet(t *testing.T, api *testAPI, ownerTok, name string) (string, string) {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": name, "species": "dog", "breed": "Beagle",
		"plan": "annual", "billingMode": "subscription",
	})
	if code != http.StatusCreated {
		t.Fatalf("create pet %d %#v", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	sessionID, _ := data["sessionId"].(string)
	checkoutURL, _ := data["checkoutUrl"].(string)
	if petID == "" || sessionID == "" || checkoutURL == "" {
		t.Fatalf("expected pet+checkout, got %#v", data)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, petID)
	})
	return petID, sessionID
}

func getEntitlement(t *testing.T, api *testAPI, token, petID string) (int, map[string]any) {
	t.Helper()
	return doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/entitlement", token, nil)
}

func postWebhook(t *testing.T, api *testAPI, payload []byte, signature string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("Stripe-Signature", signature)
	}
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestBillingCheckoutWebhookActivatesEntitlement(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	ownerID, _ := dataMap(t, env)["userId"].(string)
	if ownerID == "" {
		t.Fatalf("missing owner id: %#v", env)
	}

	petID, sessionID := createPendingPet(t, api, ownerTok, "BillingTest-"+uniqueEmail("pet"))

	// Avant paiement : entitlement pending.
	code, env = getEntitlement(t, api, ownerTok, petID)
	if code != http.StatusOK {
		t.Fatalf("entitlement %d %#v", code, env)
	}
	ent := dataMap(t, env)
	if ent["status"] != "pending" || ent["planCode"] != "annual" {
		t.Fatalf("expected pending/annual, got %#v", ent)
	}

	// Reprise de checkout sur paiement pending.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/billing/checkout", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("resume checkout %d %#v", code, env)
	}
	if u, _ := dataMap(t, env)["checkoutUrl"].(string); u == "" {
		t.Fatalf("expected resume checkoutUrl, got %#v", env)
	}

	// Webhook signé checkout.session.completed → activation.
	payload, sig, err := billing.BuildTestWebhookPayload(webhookSecret(), "checkout.session.completed", map[string]any{
		"id":             sessionID,
		"payment_status": "paid",
		"customer":       "cus_test_" + ownerID,
		"subscription":   "sub_test_" + petID,
		"payment_intent": "pi_test_" + petID,
		"metadata": map[string]any{
			"pet_id":        petID,
			"owner_user_id": ownerID,
			"plan_code":     "annual",
			"billing_mode":  "subscription",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if code, body := postWebhook(t, api, payload, sig); code != http.StatusOK {
		t.Fatalf("webhook %d %s", code, body)
	}

	code, env = getEntitlement(t, api, ownerTok, petID)
	if code != http.StatusOK {
		t.Fatalf("entitlement after webhook %d %#v", code, env)
	}
	ent = dataMap(t, env)
	if ent["status"] != "active" {
		t.Fatalf("expected active entitlement, got %#v", ent)
	}
	if ent["stripeSubscriptionId"] != "sub_test_"+petID {
		t.Fatalf("expected subscription id recorded, got %#v", ent)
	}

	// Rejeu du même événement : idempotent (200, pas de double traitement).
	if code, body := postWebhook(t, api, payload, sig); code != http.StatusOK {
		t.Fatalf("webhook replay %d %s", code, body)
	}
}

func TestBillingWebhookRejectsBadSignature(t *testing.T) {
	api := newTestAPI(t)

	payload, _, err := billing.BuildTestWebhookPayload(webhookSecret(), "checkout.session.completed", map[string]any{
		"id": "cs_forged", "metadata": map[string]any{"pet_id": "00000000-0000-0000-0000-000000000000"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Signature absente.
	if code, body := postWebhook(t, api, payload, ""); code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing signature, got %d %s", code, body)
	}
	// Signature forgée avec un mauvais secret.
	_, badSig, err := billing.BuildTestWebhookPayload("whsec_wrong", "checkout.session.completed", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if code, body := postWebhook(t, api, payload, badSig); code != http.StatusBadRequest {
		t.Fatalf("expected 400 for forged signature, got %d %s", code, body)
	}
}

func TestBillingEntitlementAccessControl(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID, _ := createPendingPet(t, api, ownerTok, "BillingACL-"+uniqueEmail("pet"))

	// Propriétaire : OK.
	if code, env := getEntitlement(t, api, ownerTok, petID); code != http.StatusOK {
		t.Fatalf("owner entitlement %d %#v", code, env)
	}
	// Non authentifié : 401.
	if code, env := getEntitlement(t, api, "", petID); code != http.StatusUnauthorized {
		t.Fatalf("anonymous should be 401, got %d %#v", code, env)
	}
	// Autre client (autre cabinet) : 403.
	marieTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	if code, env := getEntitlement(t, api, marieTok, petID); code != http.StatusForbidden {
		t.Fatalf("other client should be 403, got %d %#v", code, env)
	}
	// Véto du même cabinet (VetPlus) : OK.
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	if code, env := getEntitlement(t, api, vetTok, petID); code != http.StatusOK {
		t.Fatalf("same-practice vet entitlement %d %#v", code, env)
	}
	// Véto d'un autre cabinet : 403.
	otherVetTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	if code, env := getEntitlement(t, api, otherVetTok, petID); code != http.StatusForbidden {
		t.Fatalf("other-practice vet should be 403, got %d %#v", code, env)
	}
	// Rôle non couvert (commercial) : deny-by-default 403.
	commercialTok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	if code, env := getEntitlement(t, api, commercialTok, petID); code != http.StatusForbidden {
		t.Fatalf("commercial should be 403, got %d %#v", code, env)
	}
}

func TestBillingCheckoutRejectsInvalidPlan(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": "BillingBadPlan-" + uniqueEmail("pet"), "species": "dog",
		"plan": "lifetime_gold", "billingMode": "subscription",
	})
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("expected bad_request for invalid plan, got %d %#v", code, env)
	}
}
