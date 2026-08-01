package handlers_test

import (
	"net/http"
	"testing"
)

func TestBloodPressureCreateListClientAndVet(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/blood-pressure", clientTok, map[string]any{
		"systolicMmHg":  140,
		"diastolicMmHg": 90,
		"method":        "doppler",
		"comment":       "à domicile",
	})
	if code != http.StatusCreated {
		t.Fatalf("create bp client %d %#v", code, env)
	}
	created := dataMap(t, env)
	if sys, _ := created["systolicMmHg"].(float64); sys != 140 {
		t.Fatalf("systolic=%v", created["systolicMmHg"])
	}
	if method, _ := created["method"].(string); method != "doppler" {
		t.Fatalf("method=%v", created["method"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/blood-pressure", vetTok, map[string]any{
		"systolicMmHg":  130,
		"diastolicMmHg": 80,
		"method":        "oscillometric",
		"site":          "queue",
	})
	if code != http.StatusCreated {
		t.Fatalf("create bp vet %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/blood-pressure", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list client %d %#v", code, env)
	}
	rows, ok := env["data"].([]any)
	if !ok || len(rows) < 2 {
		t.Fatalf("expected >=2 readings, got %#v", env["data"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("timeline %d %#v", code, env)
	}
	found := false
	for _, raw := range env["data"].([]any) {
		item, _ := raw.(map[string]any)
		if item["type"] == "blood_pressure" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("timeline missing blood_pressure entry: %#v", env["data"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/blood-pressure", clientTok, map[string]any{
		"systolicMmHg":  80,
		"diastolicMmHg": 120,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid bp want 400 got %d %#v", code, env)
	}

	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/blood-pressure", farrierTok, map[string]any{
		"systolicMmHg":  120,
		"diastolicMmHg": 80,
	})
	if code != http.StatusForbidden {
		t.Fatalf("care_pro bp want 403 got %d %#v", code, env)
	}
}
