package handlers_test

import (
	"net/http"
	"testing"
)

func TestWeightReadingCreateListAndPetWeightSync(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/weights", clientTok, map[string]any{
		"weightKg": 12.5,
		"comment":  "après balade",
	})
	if code != http.StatusCreated {
		t.Fatalf("create weight %d %#v", code, env)
	}
	created := dataMap(t, env)
	if kg, _ := created["weightKg"].(float64); kg != 12.5 {
		t.Fatalf("weightKg=%v want 12.5", created["weightKg"])
	}
	if c, _ := created["comment"].(string); c != "après balade" {
		t.Fatalf("comment=%v", created["comment"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/weights", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list client %d %#v", code, env)
	}
	rows, ok := env["data"].([]any)
	if !ok || len(rows) == 0 {
		t.Fatalf("expected readings, got %#v", env["data"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/weights", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list vet %d %#v", code, env)
	}
	vetRows, _ := env["data"].([]any)
	if len(vetRows) == 0 {
		t.Fatalf("vet should see weight readings")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get pet %d %#v", code, env)
	}
	if kg, _ := dataMap(t, env)["weightKg"].(float64); kg != 12.5 {
		t.Fatalf("pet.weightKg=%v want 12.5 (synced)", dataMap(t, env)["weightKg"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/weights", clientTok, map[string]any{
		"weightKg": 0,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid weight want 400 got %d %#v", code, env)
	}
}
