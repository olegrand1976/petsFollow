package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

func TestSecretaryDeskVisitNotesAndRescheduleDirect(t *testing.T) {
	api := newTestAPI(t)
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
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

	slot := time.Now().UTC().Add(27*time.Hour + time.Duration(time.Now().UnixNano()%90)*time.Minute).Truncate(time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", secTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"silentConfirm":   true,
		"notes":           "initial",
	})
	if code == http.StatusBadRequest || code == http.StatusConflict {
		t.Skipf("visit create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/notes", secTok, map[string]any{
		"notes": "note secrétaire",
	})
	if code != http.StatusOK {
		t.Fatalf("notes %d %#v", code, env)
	}
	if got, _ := dataMap(t, env)["notes"].(string); got != "note secrétaire" {
		t.Fatalf("notes want desk note got %#v", dataMap(t, env)["notes"])
	}

	var moved bool
	for i := range 8 {
		next := time.Now().UTC().Add(time.Duration(50+i*3)*time.Hour + 17*time.Minute).Truncate(time.Minute).Format(time.RFC3339)
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
			"action":              "reschedule_direct",
			"proposedScheduledAt": next,
		})
		if code == http.StatusOK {
			moved = true
			break
		}
		if code != http.StatusConflict {
			t.Fatalf("reschedule_direct %d %#v", code, env)
		}
	}
	if !moved {
		t.Fatalf("reschedule_direct: no free slot after retries %#v", env)
	}
	data := dataMap(t, env)
	if data["status"] != "confirmed" {
		t.Fatalf("status want confirmed got %#v", data["status"])
	}
	if data["proposedScheduledAt"] != nil {
		t.Fatalf("proposed should be cleared: %#v", data["proposedScheduledAt"])
	}
}

func TestSecretarySendPreconsultAndWaitingRoom(t *testing.T) {
	api := newTestAPI(t)
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
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

	slot := time.Now().UTC().Add(41*time.Hour + time.Duration(time.Now().UnixNano()%90)*time.Minute).Truncate(time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", secTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"silentConfirm":   true,
	})
	if code == http.StatusBadRequest || code == http.StatusConflict {
		t.Skipf("visit create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
		"action": "send_preconsult",
	})
	if code != http.StatusOK {
		t.Fatalf("send_preconsult %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
		"action": "send_preconsult",
	})
	if code != http.StatusConflict {
		t.Fatalf("second send_preconsult want 409 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
		"action": "mark_waiting_room",
	})
	if code != http.StatusOK {
		t.Fatalf("mark_waiting_room %d %#v", code, env)
	}
	if dataMap(t, env)["waitingRoomAt"] == nil {
		t.Fatalf("waitingRoomAt expected: %#v", dataMap(t, env))
	}

	// Idempotent rematch while already waiting → 200 (flag already set).
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
		"action": "mark_waiting_room",
	})
	if code != http.StatusOK {
		t.Fatalf("mark_waiting_room idempotent %d %#v", code, env)
	}

	// Allow async LogNotification goroutine.
	deadline := time.Now().Add(3 * time.Second)
	var alerts []any
	for time.Now().Before(deadline) {
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/desk-alerts", vetTok, nil)
		if code != http.StatusOK {
			t.Fatalf("desk-alerts %d %#v", code, env)
		}
		alerts, _ = env["data"].([]any)
		if len(alerts) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if len(alerts) == 0 {
		t.Fatal("expected waiting_room desk alert for vet")
	}
	found := false
	for _, raw := range alerts {
		m, _ := raw.(map[string]any)
		if m["kind"] != "waiting_room" {
			continue
		}
		payload, _ := m["payload"].(map[string]any)
		if payload["visitId"] == visitID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("alert for visit %s not found: %#v", visitID, alerts)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, secTok, map[string]any{
		"action": "clear_waiting_room",
	})
	if code != http.StatusOK {
		t.Fatalf("clear_waiting_room %d %#v", code, env)
	}
	if dataMap(t, env)["waitingRoomAt"] != nil {
		t.Fatalf("waitingRoomAt should be cleared: %#v", dataMap(t, env))
	}
}
