package handlers_test

import (
	"net/http"
	"testing"
)

// Care_pro opens a person-scoped thread with a granted client, sends a message;
// thread is distinct from the cabinet practice thread.
func TestMessagingCareProClientHappyPath(t *testing.T) {
	api := newTestAPI(t)
	careTok := loginToken(t, api.handler, "vetlight.demo@petsfollow.test", "CareProDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	_, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", clientTok, nil)
	clientID, _ := dataMap(t, meEnv)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing client userId: %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/care-pro/pets", careTok, nil)
	if code != http.StatusOK {
		t.Fatalf("care-pro pets %d %#v (make seed?)", code, env)
	}
	var spiritID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] == "Spirit" {
			spiritID, _ = p["id"].(string)
			break
		}
	}
	if spiritID == "" {
		t.Fatal("Spirit grant missing for vetlight.demo")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", careTok, map[string]any{
		"clientUserId": clientID,
		"petId":        spiritID,
	})
	if code != http.StatusOK {
		t.Fatalf("care_pro ensure thread %d %#v", code, env)
	}
	careThread := dataMap(t, env)
	careThreadID, _ := careThread["id"].(string)
	if careThreadID == "" {
		t.Fatalf("missing care_pro thread id: %#v", careThread)
	}
	if pid, _ := careThread["practiceId"].(string); pid != "" {
		t.Fatalf("care_pro thread must have empty practiceId, got %#v", careThread)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", vetTok, map[string]any{
		"clientUserId": clientID,
		"petId":        spiritID,
	})
	if code != http.StatusOK {
		t.Fatalf("vet ensure thread %d %#v", code, env)
	}
	vetThreadID, _ := dataMap(t, env)["id"].(string)
	if vetThreadID == "" || vetThreadID == careThreadID {
		t.Fatalf("care_pro and vet threads must be distinct: care=%s vet=%s", careThreadID, vetThreadID)
	}

	body := "care_pro terrain hello"
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads/"+careThreadID+"/messages", careTok, map[string]any{
		"body": body,
	})
	if code != http.StatusCreated {
		t.Fatalf("care_pro send %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", careTok, nil)
	if code != http.StatusOK {
		t.Fatalf("care_pro list threads %d %#v", code, env)
	}
	listed, _ := env["data"].([]any)
	found := false
	for _, row := range listed {
		m, _ := row.(map[string]any)
		if m["id"] == careThreadID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("care_pro thread missing from list: %#v", listed)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads/"+careThreadID+"/messages", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client list care_pro messages %d %#v", code, env)
	}
	msgs, _ := env["data"].([]any)
	got := false
	for _, item := range msgs {
		m, _ := item.(map[string]any)
		if m["body"] == body {
			got = true
			break
		}
	}
	if !got {
		t.Fatalf("client missing care_pro message: %#v", msgs)
	}

	reply := "client reply to care_pro"
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads/"+careThreadID+"/messages", clientTok, map[string]any{
		"body": reply,
	})
	if code != http.StatusCreated {
		t.Fatalf("client reply %d %#v", code, env)
	}
}

func TestMessagingCareProForbiddenWithoutGrant(t *testing.T) {
	api := newTestAPI(t)
	careTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	marieTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")

	_, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", marieTok, nil)
	marieID, _ := dataMap(t, meEnv)["userId"].(string)
	if marieID == "" {
		t.Fatalf("missing marie userId: %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", careTok, map[string]any{
		"clientUserId": marieID,
	})
	if code != http.StatusForbidden {
		t.Fatalf("expected forbidden without grant, got %d %#v", code, env)
	}
}

// Revoking pet_access must cut care_pro access to an existing person-scoped thread.
func TestMessagingCareProRevokedGrantBlocksSend(t *testing.T) {
	api := newTestAPI(t)
	careTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	_, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", ownerTok, nil)
	clientID, _ := dataMap(t, meEnv)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing client userId: %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("owner pets %d %#v", code, env)
	}
	var otherPetID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] != "Spirit" {
			otherPetID, _ = p["id"].(string)
			if otherPetID != "" {
				break
			}
		}
	}
	if otherPetID == "" {
		t.Skip("no non-Spirit pet for revoke ACL messaging check")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+otherPetID+"/shares", ownerTok, map[string]any{
		"email":      "farrier.demo@petsfollow.test",
		"permission": "write_notes",
	})
	if code != http.StatusCreated {
		t.Fatalf("share write_notes %d %#v", code, env)
	}
	granteeID, _ := dataMap(t, env)["granteeUserId"].(string)
	if granteeID == "" {
		t.Fatalf("missing granteeUserId: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", careTok, map[string]any{
		"clientUserId": clientID,
		"petId":        otherPetID,
	})
	if code != http.StatusOK {
		t.Fatalf("ensure thread with grant %d %#v", code, env)
	}
	threadID, _ := dataMap(t, env)["id"].(string)
	if threadID == "" {
		t.Fatalf("missing thread id: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete,
		"/api/v1/pets/"+otherPetID+"/shares/"+granteeID, ownerTok, nil)
	if code != http.StatusOK && code != http.StatusNoContent {
		t.Fatalf("revoke share %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads/"+threadID+"/messages", careTok, map[string]any{
		"body": "should be blocked after revoke",
	})
	if code != http.StatusForbidden {
		t.Fatalf("send after revoke want 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads/"+threadID+"/messages", careTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("list messages after revoke want 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", careTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list threads %d %#v", code, env)
	}
	for _, row := range env["data"].([]any) {
		m, _ := row.(map[string]any)
		if m["id"] == threadID {
			t.Fatalf("revoked thread must not appear in care_pro list: %#v", m)
		}
	}
}
