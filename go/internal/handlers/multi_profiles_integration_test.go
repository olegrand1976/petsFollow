package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPetShareAppearsInGranteePetList(t *testing.T) {
	api := newTestAPI(t)

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "client.demo@petsfollow.test", "password": "ClientDemo123!",
	})
	if code != http.StatusOK {
		t.Fatalf("owner login %d %#v (make seed?)", code, env)
	}
	ownerTok, _ := dataMap(t, env)["accessToken"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("owner pets %d %#v", code, env)
	}
	ownerPets, ok := env["data"].([]any)
	if !ok || len(ownerPets) == 0 {
		t.Fatalf("expected owner pets, got %#v", env)
	}
	pet, _ := ownerPets[0].(map[string]any)
	petID, _ := pet["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id: %#v", pet)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "client.marie@petsfollow.test", "password": "ClientDemo123!",
	})
	if code != http.StatusOK {
		t.Fatalf("grantee login %d %#v", code, env)
	}
	granteeTok, _ := dataMap(t, env)["accessToken"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/shares", ownerTok, map[string]any{
		"email":      "client.marie@petsfollow.test",
		"permission": "read",
	})
	if code != http.StatusCreated {
		t.Fatalf("share create %d %#v", code, env)
	}
	grant := dataMap(t, env)
	granteeID, _ := grant["granteeUserId"].(string)
	if granteeID == "" {
		t.Fatalf("missing granteeUserId: %#v", grant)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete,
			"/api/v1/pets/"+petID+"/shares/"+granteeID, ownerTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", granteeTok, nil)
	if code != http.StatusOK {
		t.Fatalf("grantee pets %d %#v", code, env)
	}
	found := false
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["id"] == petID {
			found = true
			if p["permission"] != "read" {
				t.Fatalf("expected permission=read, got %#v", p["permission"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("shared pet %s missing from grantee list: %#v", petID, env["data"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/shares", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list shares %d %#v", code, env)
	}

	// Vet of same practice can list shares.
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "vet.demo@petsfollow.test", "password": "VetDemo123!",
	})
	if code != http.StatusOK {
		t.Fatalf("vet login %d %#v", code, env)
	}
	vetTok, _ := dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/shares", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet list shares %d %#v", code, env)
	}
}

func TestRegisterClientConfirmLogin(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-client")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Smoke Client",
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
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusOK {
		t.Fatalf("login %d %#v", code, env)
	}
	tok, _ := dataMap(t, env)["accessToken"].(string)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me %d %s", rec.Code, rec.Body.String())
	}
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if dataMap(t, envelope)["role"] != "client" {
		t.Fatalf("expected client role, got %#v", envelope)
	}
}
