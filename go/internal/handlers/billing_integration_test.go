package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
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

func TestBillingPortalAfterCheckout(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	ownerID, _ := dataMap(t, env)["userId"].(string)

	petID, sessionID := createPendingPet(t, api, ownerTok, "BillingPortal-"+uniqueEmail("pet"))
	payload, sig, err := billing.BuildTestWebhookPayload(webhookSecret(), "checkout.session.completed", map[string]any{
		"id":             sessionID,
		"payment_status": "paid",
		"customer":       billing.MockCustomerID(ownerID),
		"subscription":   "sub_portal_" + petID,
		"payment_intent": "pi_portal_" + petID,
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

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/billing/portal", ownerTok, map[string]any{
		"returnUrl": "petsfollow://home",
	})
	if code != http.StatusOK {
		t.Fatalf("portal %d %#v", code, env)
	}
	portalURL, _ := dataMap(t, env)["url"].(string)
	if portalURL == "" || !strings.Contains(portalURL, "/api/v1/billing/dev/mock-portal") {
		t.Fatalf("expected mock portal url, got %q", portalURL)
	}

	u, err := url.Parse(portalURL)
	if err != nil {
		t.Fatalf("parse portal url: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("mock portal page %d %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Mock Stripe portal") || !strings.Contains(body, "petsfollow://home") {
		t.Fatalf("unexpected mock portal html: %s", body)
	}
}

func TestBillingPortalSeededSubscription(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("expected pets list, got %#v", env["data"])
	}
	var bellaID string
	for _, item := range pets {
		m, _ := item.(map[string]any)
		if m["name"] == "Bella" {
			bellaID, _ = m["id"].(string)
			break
		}
	}
	if bellaID == "" {
		t.Fatal("seeded Bella not found — run make seed")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+bellaID+"/billing/portal", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("seeded portal %d %#v (re-seed if stripe_customers missing)", code, env)
	}
	if u, _ := dataMap(t, env)["url"].(string); !strings.Contains(u, "/billing/dev/mock-portal") {
		t.Fatalf("expected mock portal url, got %#v", env)
	}
}

func TestBillingCreatePetSkipCheckout(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	petName := "SkipPay-" + uniqueEmail("pet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": petName, "species": "dog", "breed": "Mix",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create skipCheckout %d %#v", code, env)
	}
	data := dataMap(t, env)
	if _, has := data["checkoutUrl"]; has {
		t.Fatalf("expected no checkoutUrl, got %#v", data)
	}
	pet, _ := data["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id: %#v", data)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, petID)
	})
	if status, _ := pet["paymentStatus"].(string); status != "pending_payment" {
		t.Fatalf("expected pending_payment, got %#v", pet)
	}
	ent, _ := pet["entitlement"].(map[string]any)
	if ent["status"] != "pending" || ent["planCode"] != "triennial" {
		t.Fatalf("expected pending triennial entitlement, got %#v", ent)
	}

	// Listed immediately after create-only.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	found := false
	for _, item := range list {
		m, _ := item.(map[string]any)
		if m["id"] == petID {
			found = true
			if m["name"] != petName {
				t.Fatalf("listed pet name %#v want %s", m["name"], petName)
			}
			listedEnt, _ := m["entitlement"].(map[string]any)
			if listedEnt["status"] != "pending" {
				t.Fatalf("listed entitlement not pending: %#v", listedEnt)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created pet %s not in list (%d pets)", petID, len(list))
	}

	// Resume checkout still works after create-only.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/billing/checkout", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("resume after skip %d %#v", code, env)
	}
	if u, _ := dataMap(t, env)["checkoutUrl"].(string); u == "" {
		t.Fatalf("expected checkoutUrl on resume, got %#v", env)
	}

	// Care mutation blocked while pending.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/care-reminders", ownerTok, map[string]any{
		"type": "vaccination", "title": "blocked",
	})
	if code != http.StatusPaymentRequired {
		t.Fatalf("expected 402 care create, got %d %#v", code, env)
	}

	// Household digest must not surface unpaid seeded reminders.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/household", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("household %d %#v", code, env)
	}
	upcoming, _ := dataMap(t, env)["upcomingReminders"].([]any)
	for _, item := range upcoming {
		m, _ := item.(map[string]any)
		if m["petId"] == petID {
			t.Fatalf("unpaid pet reminders leaked into household: %#v", m)
		}
	}
}

func TestBillingCreatePetRejectsInvalidBirthDate(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	bad := "not-a-date"
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": "BadBirth", "species": "dog", "birthDate": bad,
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid birth, got %d %#v", code, env)
	}
	if msgKey := errorMsgKey(env); msgKey != "invalid_birth_date" {
		t.Fatalf("expected invalid_birth_date, got %#v", env)
	}
}

func TestBillingCreatePetRequiresNameSpecies(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": "  ", "species": "dog",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 empty name, got %d %#v", code, env)
	}
	if msgKey := errorMsgKey(env); msgKey != "name_species_required" {
		t.Fatalf("expected name_species_required, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": "Rex", "species": "",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 empty species, got %d %#v", code, env)
	}
}

func TestBillingCreatePetsBatchRejectsInvalidPlan(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	// Count before — invalid second row must not leave the first pet committed.
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list before %d %#v", code, env)
	}
	before := len(env["data"].([]any))

	goodName := "BatchOk-" + uniqueEmail("pet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/batch", ownerTok, map[string]any{
		"pets": []any{
			map[string]any{
				"name": goodName, "species": "dog",
				"plan": "triennial", "billingMode": "subscription",
			},
			map[string]any{
				"name": "BatchBad-" + uniqueEmail("pet"), "species": "dog",
				"plan": "lifetime_gold", "billingMode": "subscription",
			},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid plan, got %d %#v", code, env)
	}
	if msgKey := errorMsgKey(env); msgKey != "invalid_plan" {
		t.Fatalf("expected invalid_plan, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list after %d %#v", code, env)
	}
	after := env["data"].([]any)
	if len(after) != before {
		t.Fatalf("batch validation fail must be atomic: before=%d after=%d", before, len(after))
	}
	for _, item := range after {
		m, _ := item.(map[string]any)
		if m["name"] == goodName {
			t.Fatalf("good pet leaked despite batch validation fail: %#v", m)
		}
	}
}

func TestBillingCreatePetsBatchAtomicOK(t *testing.T) {
	api := newTestAPI(t)
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	n1 := "BatchA-" + uniqueEmail("pet")
	n2 := "BatchB-" + uniqueEmail("pet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/batch", ownerTok, map[string]any{
		"pets": []any{
			map[string]any{"name": n1, "species": "dog", "plan": "annual", "billingMode": "subscription"},
			map[string]any{"name": n2, "species": "cat", "plan": "triennial", "billingMode": "subscription"},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("batch create %d %#v", code, env)
	}
	data := dataMap(t, env)
	pets, _ := data["pets"].([]any)
	if len(pets) != 2 {
		t.Fatalf("expected 2 pets, got %#v", data)
	}
	ids := make([]string, 0, 2)
	for _, item := range pets {
		m, _ := item.(map[string]any)
		id, _ := m["id"].(string)
		if id == "" {
			t.Fatalf("missing id: %#v", m)
		}
		ids = append(ids, id)
		ent, _ := m["entitlement"].(map[string]any)
		if ent["status"] != "pending" {
			t.Fatalf("expected pending entitlement: %#v", m)
		}
		t.Cleanup(func() {
			_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, id)
		})
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	found := map[string]bool{}
	for _, item := range env["data"].([]any) {
		m, _ := item.(map[string]any)
		for _, id := range ids {
			if m["id"] == id {
				found[id] = true
			}
		}
	}
	for _, id := range ids {
		if !found[id] {
			t.Fatalf("batch pet %s not listed", id)
		}
	}
}

// failingCheckoutGateway forces StartCheckout to fail after pet+entitlement commit.
type failingCheckoutGateway struct {
	inner billing.Gateway
}

func (g failingCheckoutGateway) CreateCheckoutSession(ctx context.Context, req billing.CheckoutRequest) (billing.CheckoutSession, error) {
	return billing.CheckoutSession{}, errStripeDown
}

func (g failingCheckoutGateway) CreatePortalSession(ctx context.Context, customerID, returnURL string) (billing.PortalSession, error) {
	return g.inner.CreatePortalSession(ctx, customerID, returnURL)
}

func (g failingCheckoutGateway) CancelSubscription(ctx context.Context, subscriptionID string) error {
	return g.inner.CancelSubscription(ctx, subscriptionID)
}

func (g failingCheckoutGateway) VerifyWebhook(payload []byte, signature string) (billing.StripeEvent, error) {
	return g.inner.VerifyWebhook(payload, signature)
}

var errStripeDown = errString("stripe unavailable")

type errString string

func (e errString) Error() string { return string(e) }

func TestBillingCreatePetCheckoutFailStill201(t *testing.T) {
	inner := billing.NewMockGateway(webhookSecret(), "http://localhost:8291", "test-url-secret")
	api := newTestAPIWithBilling(t, failingCheckoutGateway{inner: inner})
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	name := "CheckoutFail-" + uniqueEmail("pet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", ownerTok, map[string]any{
		"name": name, "species": "dog", "breed": "Mix",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": false,
	})
	if code != http.StatusCreated {
		t.Fatalf("expected 201 despite checkout fail, got %d %#v", code, env)
	}
	data := dataMap(t, env)
	if _, has := data["checkoutUrl"]; has {
		t.Fatalf("expected no checkoutUrl on stripe fail, got %#v", data)
	}
	pet, _ := data["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id: %#v", data)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, petID)
	})
	if status, _ := pet["paymentStatus"].(string); status != "pending_payment" {
		t.Fatalf("expected pending_payment, got %#v", pet)
	}
	ent, _ := pet["entitlement"].(map[string]any)
	if ent["status"] != "pending" {
		t.Fatalf("expected pending entitlement, got %#v", ent)
	}

	// Listed despite checkout failure.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	found := false
	for _, item := range env["data"].([]any) {
		m, _ := item.(map[string]any)
		if m["id"] == petID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("pet %s not listed after checkout fail", petID)
	}
}

// Orphan client (no practice) can create a pet; practiceId stays empty until link.
func TestBillingCreatePetWithoutPractice(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-orphan-pet")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Orphan Pet Owner",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	if idx := len("/confirm-email?token="); len(confirmPath) > idx {
		token = confirmPath[idx:]
	}
	if token == "" {
		t.Fatalf("missing confirm token: %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	petName := "Orphan-" + uniqueEmail("pet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", access, map[string]any{
		"name": petName, "species": "dog", "breed": "Mix",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create without practice %d %#v", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id: %#v", data)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, petID)
	})
	if pid, _ := pet["practiceId"].(string); pid != "" {
		t.Fatalf("expected empty practiceId, got %q", pid)
	}
	if status, _ := pet["paymentStatus"].(string); status != "pending_payment" {
		t.Fatalf("expected pending_payment, got %#v", pet)
	}

	var dbPractice *string
	err := api.pool.QueryRow(context.Background(), `
		SELECT practice_id::text FROM pets.pets WHERE id=$1`, petID).Scan(&dbPractice)
	if err != nil {
		t.Fatalf("db lookup: %v", err)
	}
	if dbPractice != nil {
		t.Fatalf("expected NULL practice_id in DB, got %v", *dbPractice)
	}

	// Care reminders not seeded without practice.
	var careCount int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM care.reminders WHERE pet_id=$1`, petID).Scan(&careCount); err != nil {
		t.Fatalf("care count: %v", err)
	}
	if careCount != 0 {
		t.Fatalf("expected 0 care reminders without practice, got %d", careCount)
	}

	// Activate entitlement while still orphan — cabinet features must gate.
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE billing.pet_entitlements
		SET status = 'active', valid_until = NOW() + INTERVAL '1 year'
		WHERE pet_id = $1`, petID); err != nil {
		t.Fatalf("activate entitlement: %v", err)
	}
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pets.pets SET payment_status = 'active' WHERE id = $1`, petID); err != nil {
		t.Fatalf("activate payment: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", access, map[string]any{
		"durationSec": 60,
	})
	if code != http.StatusCreated {
		t.Fatalf("orphan HR start want 201 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/weights", access, map[string]any{
		"weightKg": 12.5,
	})
	if code != http.StatusCreated {
		t.Fatalf("orphan weight want 201 got %d %#v", code, env)
	}

	var clientID, vetID, practiceID string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT id::text FROM identity.users WHERE email = $1`, email).Scan(&clientID); err != nil {
		t.Fatalf("client id: %v", err)
	}
	if err := api.pool.QueryRow(context.Background(), `
		SELECT id::text, practice_id::text FROM identity.users WHERE email = 'vet.demo@petsfollow.test'`).
		Scan(&vetID, &practiceID); err != nil {
		t.Fatalf("vet.demo: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM practice.practice_clients WHERE client_user_id = $1`, clientID)
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM identity.users WHERE id = $1`, clientID)
	})
	if _, err := api.pool.Exec(context.Background(), `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
		practiceID, clientID, vetID); err != nil {
		t.Fatalf("practice_clients: %v", err)
	}

	// First link allowed even with active entitlement.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/pets/"+petID+"/primary-practice", access, map[string]any{
		"practiceId": practiceID,
	})
	if code != http.StatusOK {
		t.Fatalf("primary-practice first link %d %#v", code, env)
	}
	var linkedPractice string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT practice_id::text FROM pets.pets WHERE id=$1`, petID).Scan(&linkedPractice); err != nil {
		t.Fatalf("practice after link: %v", err)
	}
	if linkedPractice != practiceID {
		t.Fatalf("expected practice %s, got %q", practiceID, linkedPractice)
	}
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM care.reminders WHERE pet_id=$1`, petID).Scan(&careCount); err != nil {
		t.Fatalf("care after link: %v", err)
	}
	if careCount != 0 {
		t.Fatalf("expected no auto care seed after first link, got %d", careCount)
	}
}

func errorMsgKey(env map[string]any) string {
	errObj, _ := env["error"].(map[string]any)
	if errObj == nil {
		return ""
	}
	if k, _ := errObj["msgKey"].(string); k != "" {
		return k
	}
	if k, _ := errObj["messageKey"].(string); k != "" {
		return k
	}
	return ""
}
