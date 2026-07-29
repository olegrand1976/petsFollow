package handlers_test

import (
	"net/http"
	"testing"
)

func TestDevRoleSupportAndBillingGate(t *testing.T) {
	api := newTestAPI(t)
	devTok := loginToken(t, api.handler, "dev.demo@petsfollow.test", "AdminDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets?limit=5", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev list tickets %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/stats", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev support stats %d %#v", code, env)
	}
	stats := dataMap(t, env)
	if _, ok := stats["byStatus"]; !ok {
		t.Fatalf("expected byStatus in stats: %#v", stats)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/runtime-flags", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev runtime flags %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/users", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev list users %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/payments", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev payments expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/staging/seed", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev staging seed expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/payments", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin payments %d %#v", code, env)
	}

	// Create + patch ticket as DEV.
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":  "nuxt_pro",
		"subject": "DEV patch test",
		"message": "ticket for DEV status change",
	})
	if code != http.StatusCreated {
		t.Fatalf("create ticket %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)
	if ticketID == "" {
		t.Fatalf("missing ticket id")
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, devTok, map[string]any{
		"status": "in_progress",
	})
	if code != http.StatusOK {
		t.Fatalf("dev patch ticket %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "in_progress" {
		t.Fatalf("expected in_progress, got %#v", dataMap(t, env)["status"])
	}
}

func TestDevProfileSwitchToVet(t *testing.T) {
	api := newTestAPI(t)
	devTok := loginToken(t, api.handler, "dev.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/profiles", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list profiles %d %#v", code, env)
	}
	raw := env["data"]
	list, ok := raw.([]any)
	if !ok {
		// wrapped differently?
		if m, mok := raw.(map[string]any); mok {
			list, _ = m["profiles"].([]any)
		}
	}
	if len(list) < 2 {
		t.Fatalf("expected multi profiles for DEV, got %#v", env)
	}
	var vetProfileID string
	for _, it := range list {
		m, _ := it.(map[string]any)
		if m["role"] == "vet" {
			vetProfileID, _ = m["id"].(string)
		}
	}
	if vetProfileID == "" {
		t.Fatalf("DEV missing vet profile: %#v", list)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", devTok, map[string]any{
		"profileId": vetProfileID,
	})
	if code != http.StatusOK {
		t.Fatalf("switch to vet %d %#v", code, env)
	}
	data := dataMap(t, env)
	tok, _ := data["accessToken"].(string)
	if tok == "" {
		t.Fatalf("expected new access token after switch: %#v", data)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me after switch %d %#v", code, env)
	}
	if dataMap(t, env)["role"] != "vet" {
		t.Fatalf("expected role vet after switch, got %#v", dataMap(t, env)["role"])
	}
}
