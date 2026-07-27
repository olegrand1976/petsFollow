package handlers_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestListVisitReportsIncludesMineAndAuthor(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-12-01T10:00:00Z",
		"notes":           "list authors probe visit",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	const probe = "list-authors-probe"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": probe,
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/reports", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list reports %d %#v", code, env)
	}
	items, _ := env["data"].([]any)
	if len(items) == 0 {
		t.Fatal("expected non-empty reports array")
	}
	first, _ := items[0].(map[string]any)
	if first == nil {
		t.Fatalf("first item not an object: %#v", items[0])
	}
	if mine, _ := first["mine"].(bool); !mine {
		t.Fatalf("want mine=true, got %#v", first["mine"])
	}
	bodyText, _ := first["bodyText"].(string)
	if !strings.Contains(bodyText, probe) {
		t.Fatalf("bodyText %q does not contain %q", bodyText, probe)
	}
	authorUserID, _ := first["authorUserId"].(string)
	if authorUserID == "" {
		t.Fatalf("authorUserId empty: %#v", first)
	}
}

func TestListVisitReportsMultiAuthorPeerReadable(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
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
		t.Skip("Spirit pet missing (seed incomplete)")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+spiritID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-12-02T10:00:00Z",
		"notes":           "multi-author CR visit",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	const peerBody = "peer-terrain-cr"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", farrierTok, map[string]any{
		"bodyText": peerBody,
	})
	if code != http.StatusOK {
		t.Fatalf("farrier put report %d %#v", code, env)
	}

	// Vet lists before writing own CR: must see peer content.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/reports", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list reports %d %#v", code, env)
	}
	items, _ := env["data"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected peer report, got %#v", items)
	}
	var peer map[string]any
	for _, row := range items {
		m, _ := row.(map[string]any)
		if mine, _ := m["mine"].(bool); !mine {
			peer = m
			break
		}
	}
	if peer == nil {
		t.Fatalf("expected a non-mine peer report, got %#v", items)
	}
	if body, _ := peer["bodyText"].(string); !strings.Contains(body, peerBody) {
		t.Fatalf("peer body %q missing %q", peer["bodyText"], peerBody)
	}

	const mineBody = "cabinet-mine-cr"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": mineBody,
	})
	if code != http.StatusOK {
		t.Fatalf("vet put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/reports", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list reports after mine %d %#v", code, env)
	}
	items, _ = env["data"].([]any)
	if len(items) < 2 {
		t.Fatalf("want ≥2 authors, got %#v", items)
	}
	var sawMine, sawPeer bool
	for _, row := range items {
		m, _ := row.(map[string]any)
		body, _ := m["bodyText"].(string)
		mine, _ := m["mine"].(bool)
		if mine && strings.Contains(body, mineBody) {
			sawMine = true
		}
		if !mine && strings.Contains(body, peerBody) {
			sawPeer = true
		}
	}
	if !sawMine || !sawPeer {
		t.Fatalf("want mine+peer bodies, got %#v", items)
	}
}
