package handlers_test

import (
	"net/http"
	"testing"
)

func TestClientOverviewForVet(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	meCode, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", clientTok, nil)
	if meCode != http.StatusOK {
		t.Fatalf("me %d %#v", meCode, meEnv)
	}
	clientID, _ := dataMap(t, meEnv)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing client userId: %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID+"/overview", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	data := dataMap(t, env)
	if _, ok := data["petCount"].(float64); !ok {
		t.Fatalf("petCount missing: %#v", data)
	}
	for _, key := range []string{"unreadHeartrate", "alertSessions7d", "overdueCareCount", "pendingVisits", "shareCount"} {
		if _, ok := data[key].(float64); !ok {
			t.Fatalf("%s missing or not number: %#v", key, data)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/clients/00000000-0000-0000-0000-000000000099/overview", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("unknown client want 404 got %d %#v", code, env)
	}
}
