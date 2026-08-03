package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

// Save CR on a confirmed consultation (without Terminer / done) must update lastVisitAt
// on GET /api/v1/vet/pets — matches animal timeline semantics.
func TestVetPetsLastVisitAtAfterConfirmedReportSave(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	slot := time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "lastVisitAt probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	if visitID == "" {
		t.Fatalf("missing visit id %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "CR saved — lastVisitAt should update without mark-done.",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	var lastVisitAt string
	for _, row := range rows {
		m, _ := row.(map[string]any)
		if m["id"] == petID {
			if v, ok := m["lastVisitAt"].(string); ok {
				lastVisitAt = v
			}
			break
		}
	}
	if lastVisitAt == "" {
		t.Fatalf("expected lastVisitAt for pet %s after confirmed CR save; pets=%#v", petID, env["data"])
	}
	parsed, err := time.Parse(time.RFC3339Nano, lastVisitAt)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, lastVisitAt)
	}
	if err != nil {
		t.Fatalf("parse lastVisitAt %q: %v", lastVisitAt, err)
	}
	if time.Since(parsed) > 2*time.Hour || parsed.After(time.Now().UTC().Add(time.Hour)) {
		t.Fatalf("lastVisitAt %v looks unrelated to the fresh consultation", parsed)
	}
}
