package handlers_test

import (
	"context"
	"net/http"
	"testing"
)

func TestUpdatePetSpeciesResyncsFoodChainAndClearsDomicile(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	petName := "Rente-" + uniqueEmail("pet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", tok, map[string]any{
		"name": petName, "species": "cattle", "breed": "Blanc Bleu",
		"plan": "triennial", "billingMode": "subscription", "skipCheckout": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create cattle %d %#v", code, env)
	}
	pet, _ := dataMap(t, env)["pet"].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id %#v", env)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pets.pets WHERE id=$1`, petID)
	})
	if pet["foodChainStatus"] != "food_producing" {
		t.Fatalf("create cattle foodChainStatus=%v want food_producing", pet["foodChainStatus"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/pets/"+petID, tok, map[string]any{
		"name": petName, "species": "cattle", "breed": "Blanc Bleu",
		"domicileLocation": "Ferme Demo — Namur",
	})
	if code != http.StatusOK {
		t.Fatalf("set domicile %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get after domicile %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["domicileLocation"] != "Ferme Demo — Namur" {
		t.Fatalf("domicile=%v", got["domicileLocation"])
	}
	if got["foodChainStatus"] != "food_producing" {
		t.Fatalf("foodChain still %v", got["foodChainStatus"])
	}

	// Leave food-chain species without domicileLocation in body → clear housing + reset status.
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/pets/"+petID, tok, map[string]any{
		"name": petName, "species": "dog", "breed": "Mix",
	})
	if code != http.StatusOK {
		t.Fatalf("species→dog %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get after species change %d %#v", code, env)
	}
	got = dataMap(t, env)
	if got["species"] != "dog" {
		t.Fatalf("species=%v", got["species"])
	}
	if got["foodChainStatus"] != "companion" {
		t.Fatalf("foodChainStatus=%v want companion after leaving cattle", got["foodChainStatus"])
	}
	if loc, _ := got["domicileLocation"].(string); loc != "" {
		t.Fatalf("domicile should be cleared, got %q", loc)
	}

	// Enter production species again → default food_producing.
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/pets/"+petID, tok, map[string]any{
		"name": petName, "species": "alpaca", "breed": "Huacaya",
	})
	if code != http.StatusOK {
		t.Fatalf("species→alpaca %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get alpaca %d %#v", code, env)
	}
	got = dataMap(t, env)
	if got["foodChainStatus"] != "food_producing" {
		t.Fatalf("alpaca foodChainStatus=%v want food_producing", got["foodChainStatus"])
	}
}
