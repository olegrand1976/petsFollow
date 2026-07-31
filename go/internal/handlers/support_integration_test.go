package handlers_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestSupportTicketCreateListReply(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":  "nuxt_pro",
		"subject": "Bouton cassé",
		"message": "Le bouton valider ne répond plus sur /calendar",
		"diagnostics": map[string]any{
			"windowMinutes": 15,
			"consoleErrors": []any{map[string]any{"message": "TypeError: x", "ts": "2026-07-26T10:00:00Z"}},
			"networkEntries": []any{},
		},
		"appVersion": "web-dev",
		"route":      "/calendar",
		"locale":     "fr",
	})
	if code != http.StatusCreated {
		t.Fatalf("create ticket %d %#v", code, env)
	}
	ticket := dataMap(t, env)
	ticketID, _ := ticket["id"].(string)
	if ticketID == "" {
		t.Fatalf("missing ticket id: %#v", ticket)
	}
	if ticket["status"] != "open" {
		t.Fatalf("expected open, got %#v", ticket["status"])
	}
	if ticket["source"] != "nuxt_pro" {
		t.Fatalf("vet ticket source should be nuxt_pro, got %#v", ticket["source"])
	}

	// Spoof flutter_client must be ignored for vet (defaults to nuxt_pro unless flutter_pro_light).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":  "flutter_client",
		"subject": "Spoof source",
		"message": "should not stick as flutter_client",
		"diagnostics": map[string]any{},
	})
	if code != http.StatusCreated {
		t.Fatalf("create spoof ticket %d %#v", code, env)
	}
	if dataMap(t, env)["source"] != "nuxt_pro" {
		t.Fatalf("spoofed source not overridden: %#v", dataMap(t, env)["source"])
	}

	// Non-admin cannot list.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet list expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets?status=open", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin list %d %#v", code, env)
	}
	list := dataMap(t, env)
	items, _ := list["items"].([]any)
	found := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["id"] == ticketID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created ticket not in list: %#v", list)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin get %d %#v", code, env)
	}
	detail := dataMap(t, env)
	if detail["message"] == nil || detail["diagnostics"] == nil {
		t.Fatalf("detail missing message/diagnostics: %#v", detail)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/support/tickets/"+ticketID+"/replies", adminTok, map[string]any{
		"body": "Merci, on regarde ça.",
	})
	if code != http.StatusCreated {
		t.Fatalf("admin reply %d %#v", code, env)
	}
	replyData := dataMap(t, env)
	ticketAfter, _ := replyData["ticket"].(map[string]any)
	if ticketAfter["status"] != "in_progress" {
		t.Fatalf("expected in_progress after reply, got %#v", ticketAfter["status"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
		"status": "resolved",
	})
	if code != http.StatusOK {
		t.Fatalf("patch status %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "resolved" {
		t.Fatalf("expected resolved: %#v", env)
	}
}

func TestSupportTicketDiagnosticsTooLarge(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	big := strings.Repeat("x", store.MaxSupportDiagnosticsBytes+10)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", tok, map[string]any{
		"source":      "nuxt_pro",
		"subject":     "Too big",
		"message":     "payload",
		"diagnostics": map[string]any{"pad": big},
	})
	if code != http.StatusRequestEntityTooLarge && code != http.StatusBadRequest {
		t.Fatalf("expected 413 or 400, got %d %#v", code, env)
	}
	if code == http.StatusRequestEntityTooLarge && errCode(env) != "diagnostics_too_large" {
		t.Fatalf("expected diagnostics_too_large, got %#v", env)
	}
}

func TestSupportTicketSearch(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	marker := "UniqueSearchMarkerXYZ" + strings.ReplaceAll(uniqueEmail("q"), "@", "")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":      "nuxt_pro",
		"subject":     marker,
		"message":     "search probe",
		"diagnostics": map[string]any{},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets?q="+url.QueryEscape(marker), adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list q match %d %#v", code, env)
	}
	found := false
	for _, it := range dataMap(t, env)["items"].([]any) {
		m, _ := it.(map[string]any)
		if m["id"] == ticketID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ticket for q=%s: %#v", marker, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets?q=zzzinexistant999", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list q miss %d %#v", code, env)
	}
	items, _ := dataMap(t, env)["items"].([]any)
	if len(items) != 0 {
		t.Fatalf("expected empty for nonsense q, got %#v", items)
	}
}

func TestSupportTicketExportAndAnonymize(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	email := uniqueEmail("support-rgpd")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Support RGPD", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	subject := "RGPD export subject " + email
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", access, map[string]any{
		"source":  "flutter_client",
		"subject": subject,
		"message": "confidential bug details",
		"diagnostics": map[string]any{
			"windowMinutes": 15,
			"consoleErrors": []any{map[string]any{"message": "boom", "ts": "2026-07-26T12:00:00Z"}},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create ticket %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)
	if ticketID == "" {
		t.Fatalf("missing ticket id: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", access, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	exportData := dataMap(t, env)
	tickets, ok := exportData["supportTickets"].([]any)
	if !ok || len(tickets) == 0 {
		t.Fatalf("export missing supportTickets: %#v", exportData)
	}
	var exported map[string]any
	for _, it := range tickets {
		m, _ := it.(map[string]any)
		if m["id"] == ticketID || m["subject"] == subject {
			exported = m
			break
		}
	}
	if exported == nil {
		t.Fatalf("ticket not in export: %#v", tickets)
	}
	if exported["subject"] != subject {
		t.Fatalf("export subject mismatch: %#v", exported)
	}
	if exported["diagnostics"] == nil {
		t.Fatalf("export missing diagnostics: %#v", exported)
	}
	diag, _ := exported["diagnostics"].(map[string]any)
	if diag == nil {
		t.Fatalf("diagnostics not an object: %#v", exported["diagnostics"])
	}
	if _, has := diag["consoleErrors"]; !has {
		t.Fatalf("diagnostics missing consoleErrors: %#v", diag)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("delete me %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin get after anonymize %d %#v", code, env)
	}
	detail := dataMap(t, env)
	if detail["subject"] != "[anonymized]" || detail["message"] != "[anonymized]" {
		t.Fatalf("expected anonymized subject/message: %#v", detail)
	}
	if detail["createdBy"] != nil {
		t.Fatalf("expected createdBy null: %#v", detail["createdBy"])
	}
	anonDiag, _ := detail["diagnostics"].(map[string]any)
	if anonDiag == nil || len(anonDiag) != 0 {
		t.Fatalf("expected empty diagnostics object: %#v", detail["diagnostics"])
	}
}

func TestSupportTicketRateLimit(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("support-rl")
	password := "ClientPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Support RL", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	body := map[string]any{
		"source":  "flutter_client",
		"subject": "RL",
		"message": "rate limit probe",
		"diagnostics": map[string]any{},
	}
	for i := 0; i < store.SupportTicketsPerHour; i++ {
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", access, body)
		if code != http.StatusCreated {
			t.Fatalf("create #%d %d %#v", i+1, code, env)
		}
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", access, body)
	if code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d %#v", code, env)
	}
}
