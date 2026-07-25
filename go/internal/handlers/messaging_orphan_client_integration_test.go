package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Client self-signup (sans cabinet) : GET /messaging/threads doit renvoyer [] et non 500.
func TestListThreadsOrphanClientEmpty(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-orphan-client")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Orphan Client",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	if idx := len("/confirm-email?token="); len(confirmPath) > idx {
		token = confirmPath[idx:]
	}
	if token == "" {
		t.Fatalf("missing confirm token: %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		// confirm may not auto-login in all configs — fall back to password login
		code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email": email, "password": password,
		})
		if code != http.StatusOK {
			t.Fatalf("login %d %#v", code, env)
		}
		access, _ = dataMap(t, env)["accessToken"].(string)
	}
	if access == "" {
		t.Fatal("missing access token")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", access, nil)
	if code != http.StatusOK {
		t.Fatalf("threads expected 200, got %d %#v", code, env)
	}
	raw, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %#v", env["data"])
	}
	if len(raw) != 0 {
		t.Fatalf("expected empty threads for orphan client, got %#v", raw)
	}

	// Ensure JWT has no practice (self-signup).
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me %d %s", rec.Code, rec.Body.String())
	}
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	me := dataMap(t, envelope)
	if me["role"] != "client" {
		t.Fatalf("expected client, got %#v", me)
	}
	if pid, _ := me["practiceId"].(string); pid != "" {
		t.Fatalf("expected empty practiceId for orphan client, got %q", pid)
	}
}
