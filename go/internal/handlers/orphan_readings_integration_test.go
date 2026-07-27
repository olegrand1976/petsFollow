package handlers_test

import (
	"context"
	"net/http"
	"testing"
)

// Orphan client (no practice): weight + heart-rate start must succeed.
func TestOrphanClientWeightAndHeartRateAllowed(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-orphan-readings")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Orphan Readings",
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
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email": email, "password": password,
		})
		if code != http.StatusOK {
			t.Fatalf("login %d %#v", code, env)
		}
		access, _ = dataMap(t, env)["accessToken"].(string)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", access, map[string]any{
		"name": "OrphanPet", "species": "dog",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create pet %d %#v", code, env)
	}
	pet, _ := dataMap(t, env)["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id %#v", env)
	}

	if _, err := api.pool.Exec(context.Background(), `
		UPDATE billing.pet_entitlements SET status='active', valid_until=NOW()+INTERVAL '1 year'
		WHERE pet_id=$1`, petID); err != nil {
		t.Fatalf("activate entitlement: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/weights", access, map[string]any{
		"weightKg": 8.2,
	})
	if code != http.StatusCreated {
		t.Fatalf("create weight without vet want 201 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", access, map[string]any{
		"durationSec": 60,
	})
	if code != http.StatusCreated {
		t.Fatalf("start HR without vet want 201 got %d %#v", code, env)
	}
}
