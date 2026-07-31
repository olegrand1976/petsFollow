package handlers_test

import (
	"net/http"
	"testing"
)

func TestClientEnsureThreadWithPet(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/vets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me/vets %d %#v", code, env)
	}
	vets, _ := env["data"].([]any)
	if len(vets) == 0 {
		t.Fatal("expected at least one linked vet for client.demo")
	}
	vet, _ := vets[0].(map[string]any)
	practiceID, _ := vet["practiceId"].(string)
	if practiceID == "" {
		t.Fatalf("missing practiceId %#v", vet)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", clientTok, map[string]any{
		"practiceId": practiceID,
		"petId":      petID,
	})
	if code != http.StatusOK {
		t.Fatalf("ensure thread %d %#v", code, env)
	}
	thread := dataMap(t, env)
	if thread["id"] == nil || thread["petId"] != petID {
		t.Fatalf("expected pet-scoped thread, got %#v", thread)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/messaging/threads", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list threads %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	found := false
	for _, raw := range list {
		row, _ := raw.(map[string]any)
		if row["id"] == thread["id"] {
			found = true
			if row["petName"] == nil || row["petName"] == "" {
				t.Fatalf("expected petName enrichment, got %#v", row)
			}
		}
	}
	if !found {
		t.Fatalf("created thread not in list %#v", list)
	}

	// Pet already stamped to a practice cannot open a thread on another cabinet.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", clientTok, map[string]any{
		"practiceId": "00000000-0000-4000-8000-000000000099",
		"petId":      petID,
	})
	if code != http.StatusForbidden {
		t.Fatalf("cross-practice ensure want 403 got %d %#v", code, env)
	}
}
