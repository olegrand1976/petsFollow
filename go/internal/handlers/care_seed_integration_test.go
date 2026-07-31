package handlers_test

import (
	"net/http"
	"testing"
)

// Creating a pet must not auto-seed Care reminders (manual only).
func TestCreatePetNoDefaultCareReminders(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", clientTok, map[string]any{
		"name":         "NoSeedCare",
		"species":      "dog",
		"plan":         "triennial",
		"billingMode":  "subscription",
		"skipCheckout": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create pet %d %#v", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id in %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/care-reminders", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list care-reminders %d %#v", code, env)
	}
	raw, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %#v", env["data"])
	}
	if len(raw) != 0 {
		t.Fatalf("expected no default care reminders, got %d: %#v", len(raw), raw)
	}
}
