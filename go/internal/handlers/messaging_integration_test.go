package handlers_test

import (
	"net/http"
	"testing"
)

// Vet → ensure thread → send message → client lists messages (happy path messagerie).
func TestMessagingVetClientHappyPath(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	_, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", clientTok, nil)
	clientID, _ := dataMap(t, meEnv)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing client userId: %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", vetTok, map[string]any{
		"clientUserId": clientID,
	})
	if code != http.StatusOK {
		t.Fatalf("ensure thread %d %#v", code, env)
	}
	thread := dataMap(t, env)
	threadID, _ := thread["id"].(string)
	if threadID == "" {
		t.Fatalf("missing thread id: %#v", thread)
	}

	// Practice list should expose clientName (and petName when pet-scoped).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list threads vet %d %#v", code, env)
	}
	listed, _ := env["data"].([]any)
	foundThread := false
	for _, row := range listed {
		m, _ := row.(map[string]any)
		if m["id"] != threadID {
			continue
		}
		foundThread = true
		if name, _ := m["clientName"].(string); name == "" {
			t.Fatalf("expected clientName on practice thread, got %#v", m)
		}
		break
	}
	if !foundThread {
		t.Fatalf("ensured thread not in practice list: %#v", listed)
	}

	body := "integration messaging hello"
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads/"+threadID+"/messages", vetTok, map[string]any{
		"body": body,
	})
	if code != http.StatusCreated {
		t.Fatalf("send message %d %#v", code, env)
	}
	msg := dataMap(t, env)
	if msg["body"] != body {
		t.Fatalf("unexpected message body: %#v", msg)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads/"+threadID+"/messages", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list messages client %d %#v", code, env)
	}
	raw, ok := env["data"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("expected messages array, got %#v", env["data"])
	}
	found := false
	for _, item := range raw {
		m, _ := item.(map[string]any)
		if m["body"] == body {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("client did not see vet message: %#v", raw)
	}

	t.Run("ensure_with_pet_exposes_petName", func(t *testing.T) {
		code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
		if code != http.StatusOK {
			t.Fatalf("list client pets %d %#v", code, env)
		}
		pets, _ := env["data"].([]any)
		if len(pets) == 0 {
			t.Fatal("expected at least one pet for client.demo")
		}
		pet, _ := pets[0].(map[string]any)
		petID, _ := pet["id"].(string)
		if petID == "" {
			t.Fatalf("missing pet id: %#v", pet)
		}

		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", vetTok, map[string]any{
			"clientUserId": clientID,
			"petId":        petID,
		})
		if code != http.StatusOK {
			t.Fatalf("ensure thread with pet %d %#v", code, env)
		}
		petThread := dataMap(t, env)
		petThreadID, _ := petThread["id"].(string)
		if petThreadID == "" {
			t.Fatalf("missing pet-scoped thread id: %#v", petThread)
		}

		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", vetTok, nil)
		if code != http.StatusOK {
			t.Fatalf("list threads vet %d %#v", code, env)
		}
		listed, _ := env["data"].([]any)
		found := false
		for _, row := range listed {
			m, _ := row.(map[string]any)
			if m["id"] != petThreadID {
				continue
			}
			found = true
			if name, _ := m["petName"].(string); name == "" {
				t.Fatalf("expected non-empty petName on pet-scoped thread, got %#v", m)
			}
			break
		}
		if !found {
			t.Fatalf("pet-scoped thread not in practice list: %#v", listed)
		}
	})

	// ACL : client orphelin ne peut pas lire le thread d'un autre.
	orphanEmail := uniqueEmail("msg-orphan")
	orphanPass := "ClientPass123!"
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": orphanEmail, "password": orphanPass, "fullName": "Orphan Msg",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register orphan %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	const prefix = "/confirm-email?token="
	if len(confirmPath) > len(prefix) {
		token = confirmPath[len(prefix):]
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm orphan %d %#v", code, env)
	}
	orphanTok, _ := dataMap(t, env)["accessToken"].(string)
	if orphanTok == "" {
		orphanTok = loginToken(t, api.handler, orphanEmail, orphanPass)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads/"+threadID+"/messages", orphanTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("orphan access want 403 got %d %#v", code, env)
	}
}
