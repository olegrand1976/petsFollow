package handlers_test

import (
	"net/http"
	"testing"
)

// Export + DELETE /me sur un client jetable (pas de seed demo).
func TestRGPDExportAndDeleteDisposableClient(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("rgpd-client")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "RGPD Disposable",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	const prefix = "/confirm-email?token="
	if len(confirmPath) > len(prefix) {
		token = confirmPath[len(prefix):]
	}
	if token == "" {
		t.Fatalf("missing confirm token: %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", access, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	data := dataMap(t, env)
	profile, _ := data["profile"].(map[string]any)
	if profile == nil {
		t.Fatalf("export missing profile: %#v", data)
	}
	if _, hasHash := profile["password_hash"]; hasHash {
		t.Fatal("export must not include password_hash")
	}
	if _, ok := data["pets"]; !ok {
		t.Fatalf("client export missing pets key: %#v", data)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("delete me %d %#v", code, env)
	}
	if ok, _ := dataMap(t, env)["ok"].(bool); !ok {
		t.Fatalf("expected ok true: %#v", env)
	}

	// Login après purge doit échouer.
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code == http.StatusOK {
		t.Fatalf("login after delete should fail, got 200 %#v", env)
	}
}
