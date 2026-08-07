package handlers_test

import (
	"net/http"
	"testing"
)

func TestPetTipsClientOnlyAndLocalized(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/pet-tips", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet want 403 got %d %#v", code, env)
	}
	if errMsgKey(env) != "client_only" {
		t.Fatalf("error %#v", env["error"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/pet-tips", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client want 200 got %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("client.demo has pets — expected tips %#v", data)
	}
	if len(items) > 3 {
		t.Fatalf("limit 3 got %d", len(items))
	}
	first, _ := items[0].(map[string]any)
	if first["id"] == nil || first["title"] == nil || first["body"] == nil {
		t.Fatalf("item %#v", first)
	}
	title := first["title"].(string)
	if title == "" || title == "pet_tips."+first["id"].(string)+".title" {
		t.Fatalf("unlocalized title %#v", first)
	}
}

func TestVetNewsForbiddenForClient(t *testing.T) {
	t.Setenv("VET_NEWS_ENABLED", "true")
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/news", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("client want 403 got %d %#v", code, env)
	}
}
