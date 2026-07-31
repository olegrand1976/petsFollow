package handlers_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestClientContactPhoneCreateListGetPatch(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.phone.%s@petsfollow.test", uuid.NewString()[:8])
	phone := "0470 11 22 33"

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Phone Test", "contactPhone": phone,
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["contactPhone"] != phone {
		t.Fatalf("get contactPhone=%v want %q", got["contactPhone"], phone)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	found := false
	for _, raw := range list {
		row, _ := raw.(map[string]any)
		if row["userId"] == clientID {
			found = true
			if row["contactPhone"] != phone {
				t.Fatalf("list contactPhone=%v want %q", row["contactPhone"], phone)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created client not in list")
	}

	updated := "0499 00 11 22"
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": updated,
	})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	if dataMap(t, env)["contactPhone"] != updated {
		t.Fatalf("patch contactPhone=%v want %q", dataMap(t, env)["contactPhone"], updated)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": "",
	})
	if code != http.StatusOK {
		t.Fatalf("clear %d %#v", code, env)
	}
	cleared := dataMap(t, env)["contactPhone"]
	if cleared != nil && cleared != "" {
		t.Fatalf("clear contactPhone=%v want empty", cleared)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/clients/00000000-0000-0000-0000-000000000099", vetTok, map[string]any{
			"contactPhone": "0470 00 00 99",
		})
	if code != http.StatusNotFound {
		t.Fatalf("unknown client want 404 got %d %#v", code, env)
	}

	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, parcTok, map[string]any{
		"contactPhone": "0470 99 99 99",
	})
	if code != http.StatusNotFound {
		t.Fatalf("other practice want 404 got %d %#v", code, env)
	}

	tooLong := strings.Repeat("1", 41)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": tooLong,
	})
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("too long want 400 bad_request got %d %#v", code, env)
	}
}

func TestClientContactPhonePatchACL(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.phone.acl.%s@petsfollow.test", uuid.NewString()[:8])
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Phone ACL", "contactPhone": "0470 00 00 10",
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
  clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, map[string]any{
		"contactPhone": "0470 00 00 11",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth want 401 got %d %#v", code, env)
	}

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, clientTok, map[string]any{
		"contactPhone": "0470 00 00 12",
	})
	if code != http.StatusForbidden {
		t.Fatalf("client role want 403 got %d %#v", code, env)
	}

	commTok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, commTok, map[string]any{
		"contactPhone": "0470 00 00 13",
	})
	if code != http.StatusForbidden {
		t.Fatalf("commercial want 403 got %d %#v", code, env)
	}
}
