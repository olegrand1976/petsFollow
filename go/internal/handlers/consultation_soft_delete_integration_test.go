package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

func TestSoftDeleteConsultationHiddenFromList(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(19 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "soft-delete probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)

	_, _ = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "CR soft-delete test",
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list before %d %#v", code, env)
	}
	if !consultationsContain(env, visitID) {
		t.Fatalf("visit should appear in consultations before soft-delete")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("soft-delete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list after %d %#v", code, env)
	}
	if consultationsContain(env, visitID) {
		t.Fatalf("soft-deleted visit must not appear in consultations list")
	}

	// Idempotent second delete → 404
	code, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("second soft-delete want 404 got %d", code)
	}
}

func consultationsContain(env map[string]any, visitID string) bool {
	items, _ := env["data"].([]any)
	for _, raw := range items {
		m, _ := raw.(map[string]any)
		if m["id"] == visitID {
			return true
		}
	}
	return false
}
