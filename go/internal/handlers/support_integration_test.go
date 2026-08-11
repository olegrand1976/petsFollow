package handlers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
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
			"windowMinutes":  15,
			"consoleErrors":  []any{map[string]any{"message": "TypeError: x", "ts": "2026-07-26T10:00:00Z"}},
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
		"source":      "flutter_client",
		"subject":     "Spoof source",
		"message":     "should not stick as flutter_client",
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

	for _, st := range []string{"to_test", "done"} {
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
			"status": st,
		})
		if code != http.StatusOK {
			t.Fatalf("patch status %s %d %#v", st, code, env)
		}
		if dataMap(t, env)["status"] != st {
			t.Fatalf("expected %s: %#v", st, env)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin get after patch %d %#v", code, env)
	}
	history, _ := dataMap(t, env)["statusHistory"].([]any)
	if len(history) < 2 {
		t.Fatalf("expected status history events, got %#v", dataMap(t, env)["statusHistory"])
	}
	foundDone := false
	for _, raw := range history {
		ev, _ := raw.(map[string]any)
		if ev["toStatus"] == "done" {
			foundDone = true
			break
		}
	}
	if !foundDone {
		t.Fatalf("expected done event in history: %#v", history)
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

func TestSupportTicketStatusWorkflowAndAttachments(t *testing.T) {
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetOtherTok := loginToken(t, api.handler, "vet.lyon@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":      "nuxt_pro",
		"subject":     "Workflow attachment",
		"message":     "need status + file",
		"diagnostics": map[string]any{},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)

	for _, st := range []string{"in_progress", "to_test", "done"} {
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
			"status": st,
		})
		if code != http.StatusOK {
			t.Fatalf("patch %s %d %#v", st, code, env)
		}
		if dataMap(t, env)["status"] != st {
			t.Fatalf("expected %s: %#v", st, env)
		}
	}
	// Illegal jump: done → to_test (must go closed or stay done).
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
		"status": "to_test",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 invalid_transition got %d %#v", code, env)
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj["msgKey"] != "invalid_transition" && errObj["message"] != "invalid_transition" {
		t.Fatalf("want invalid_transition msgKey %#v", env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
		"status": "closed",
	})
	if code != http.StatusOK || dataMap(t, env)["status"] != "closed" {
		t.Fatalf("patch closed %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	history, _ := dataMap(t, env)["statusHistory"].([]any)
	if len(history) < 5 {
		t.Fatalf("expected >=5 status events, got %d %#v", len(history), history)
	}

	pdf := []byte("%PDF-1.4\n% support note\n")
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "note.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pdf); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/support/tickets/"+ticketID+"/attachments", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload attachment %d %#v", rec.Code, envelope)
	}
	att := dataMap(t, envelope)
	attID, _ := att["id"].(string)
	if attID == "" {
		t.Fatalf("missing attachment id: %#v", att)
	}
	if _, ok := att["objectKey"]; ok {
		t.Fatalf("must not expose objectKey: %#v", att)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get after upload %d %#v", code, env)
	}
	atts, _ := dataMap(t, env)["attachments"].([]any)
	if len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %#v", atts)
	}

	dl := "/api/v1/admin/support/tickets/" + ticketID + "/attachments/" + attID + "/download"
	req = httptest.NewRequest(http.MethodGet, dl, nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download %d %s", rec.Code, rec.Body.String())
	}
	if !bytes.Equal(rec.Body.Bytes(), pdf) {
		t.Fatalf("unexpected download body")
	}

	req = httptest.NewRequest(http.MethodGet, dl, nil)
	req.Header.Set("Authorization", "Bearer "+vetOtherTok)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("vet must not download support attachment, got 200")
	}
}

func TestSupportTicketNotifySoftFail(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetSupportInboxEmail("ops-support@example.invalid")
	api.api.TestReplaceNotifier(email.NewNotifierAuth(
		"smtp.invalid.petsfollow", 587, "support@petsfollow.test",
		"smtp-user", "", "http://localhost:3002", "https://ll-it-sc.be",
	))

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":      "nuxt_pro",
		"subject":     "Notify soft-fail",
		"message":     "SMTP down must not block create",
		"diagnostics": map[string]any{},
	})
	if code != http.StatusCreated {
		t.Fatalf("create with failing SMTP %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)
	if ticketID == "" {
		t.Fatalf("missing ticket id: %#v", env)
	}

	for _, st := range []string{"in_progress", "to_test"} {
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, adminTok, map[string]any{
			"status": st,
		})
		if code != http.StatusOK {
			t.Fatalf("patch status %s with failing SMTP %d %#v", st, code, env)
		}
		if dataMap(t, env)["status"] != st {
			t.Fatalf("expected %s: %#v", st, env)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/support/tickets/"+ticketID+"/replies", adminTok, map[string]any{
		"body": "Comment despite SMTP fail",
	})
	if code != http.StatusCreated {
		t.Fatalf("reply with failing SMTP %d %#v", code, env)
	}
}
