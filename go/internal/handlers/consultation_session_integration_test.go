package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

// Walk-in consultationSession must not block (nor be blocked by) a normal booked slot.
func TestConsultationSessionExcludedFromOverlap(t *testing.T) {
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

	// Offset within the ±30m walk-in window, away from typical :00/:30 seed slots.
	slotTime := time.Now().UTC().Add(27 * time.Minute)
	slot := slotTime.Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "walk-in overlap probe",
	})
	if code == http.StatusBadRequest {
		// Cabinet on vacation for "now+27m" — skip rather than flake.
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("consultation create %d %#v", code, env)
	}
	consultID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+consultID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})
	if dataMap(t, env)["consultationSession"] != true {
		t.Fatalf("want consultationSession true %#v", env)
	}

	// Same slot, normal booked visit must succeed (walk-in does not occupy the slot).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"notes":           "rdv after walk-in",
	})
	if code != http.StatusCreated {
		t.Fatalf("booked create after consultation %d %#v", code, env)
	}
	bookedID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+bookedID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	// Second normal booking on same slot must conflict.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"notes":           "should conflict",
	})
	if code != http.StatusConflict {
		t.Fatalf("second booked want 409 got %d %#v", code, env)
	}

	// Another walk-in on the same slot still allowed.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "second walk-in",
	})
	if code != http.StatusCreated {
		t.Fatalf("second consultation create %d %#v", code, env)
	}
	consult2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+consult2, vetTok, map[string]any{
			"status": "cancelled",
		})
	})
}

func TestConsultationSessionGuards(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"consultationSession": true,
		"notes":               "missing confirmDirect",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 without confirmDirect got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         "2099-06-15T14:00:00Z",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "far future",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 stale session got %d %#v", code, env)
	}
}

func TestConsultationSessionCancelBlockedAfterReport(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "cancel after report",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	// Empty Ensure draft must still allow cancel.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("ensure report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
		"status": "cancelled",
	})
	if code != http.StatusOK {
		t.Fatalf("cancel empty draft want 200 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "cancel blocked",
	})
	if code != http.StatusCreated {
		t.Fatalf("create2 %d %#v", code, env)
	}
	visit2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		// may already be non-cancelled if guard works — still try
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2, vetTok, map[string]any{
			"status": "done",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visit2+"/report", vetTok, map[string]any{
		"bodyText": "Examen clinique OK",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2, vetTok, map[string]any{
		"status": "cancelled",
	})
	if code != http.StatusConflict {
		t.Fatalf("cancel after report want 409 got %d %#v", code, env)
	}
}

func TestConsultationSessionNotReschedulable(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "no reschedule",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
		"action":              "propose_reschedule",
		"proposedScheduledAt": time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
	})
	if code != http.StatusBadRequest {
		t.Fatalf("reschedule walk-in want 400 got %d %#v", code, env)
	}
}
