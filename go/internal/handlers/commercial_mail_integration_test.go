package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestCommercialMailTemplatesSendTrackAndOptOut(t *testing.T) {
	api := newTestAPI(t)

	adminEmail := uniqueEmail("cm-mail-admin")
	insertVerifiedUser(t, api, string(kernel.RoleAdmin), adminEmail, "AdminDemo123!", "Admin Mail", nil)
	adminTok := loginToken(t, api.handler, adminEmail, "AdminDemo123!")

	commEmail := uniqueEmail("cm-mail-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Mail Commercial",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	tok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	slug := "test_intro_" + strings.ReplaceAll(commEmail, "@", "_")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/email-templates", tok, map[string]any{
		"slug":     slug,
		"name":     "Test intro",
		"category": "intro",
		"locale":   "fr",
		"subject":  "Bonjour {{practice_name}}",
		"bodyHtml": `<p>Hello {{contact_name}}</p><p><a href="https://example.com/demo">CTA</a></p>`,
	})
	if code != http.StatusCreated {
		t.Fatalf("create template: %d %#v", code, env)
	}
	tplID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", tok, map[string]any{
		"practiceName": "Cabinet Mail " + uuid.NewString()[:8],
		"contactName":  "Dr Mail",
		"contactEmail": uniqueEmail("prospect-mail"),
		"contactPhone": "0470" + fmt.Sprintf("%06d", time.Now().UnixNano()%1_000_000),
		"city":         "Bruxelles",
		"status":       "new",
	})
	if code != http.StatusCreated {
		t.Fatalf("create prospect: %d %#v", code, env)
	}
	prospectID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/email-templates/"+tplID+"/preview", tok, map[string]any{
		"prospectId": prospectID,
	})
	if code != http.StatusOK {
		t.Fatalf("preview: %d %#v", code, env)
	}
	if subj, _ := dataMap(t, env)["subject"].(string); !strings.Contains(subj, "Cabinet Mail") {
		t.Fatalf("expected substituted subject, got %#v", dataMap(t, env)["subject"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/emails", tok, map[string]any{
		"templateId": tplID,
	})
	if code != http.StatusCreated {
		t.Fatalf("send email: %d %#v", code, env)
	}
	send := dataMap(t, env)
	sendID, _ := send["id"].(string)
	if st, _ := send["status"].(string); st != "sent" {
		t.Fatalf("expected sent, got %#v", send)
	}
	clicks, _ := send["clicks"].([]any)
	if len(clicks) < 1 {
		t.Fatalf("expected tracked click row, got %#v", send)
	}
	if tok, _ := clicks[0].(map[string]any)["clickToken"].(string); tok != "" {
		t.Fatal("clickToken must not be returned to client")
	}

	var clickTok, openTok string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT c.click_token, s.open_token
		FROM sales.email_sends s
		JOIN sales.email_clicks c ON c.send_id = s.id
		WHERE s.id=$1
		LIMIT 1`, sendID).Scan(&clickTok, &openTok); err != nil {
		t.Fatalf("load tokens: %v", err)
	}
	if clickTok == "" || openTok == "" {
		t.Fatal("missing tokens in db")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/commercial-mail/c/"+clickTok, nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("click redirect want 302 got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/demo" {
		t.Fatalf("click location %#v", loc)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/emails/"+sendID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get send: %d %#v", code, env)
	}
	if bodyHTML, _ := dataMap(t, env)["bodyHtmlRendered"].(string); strings.Contains(bodyHTML, openTok) {
		t.Fatal("open token must be scrubbed from API body")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/public/commercial-mail/o/"+openTok, nil)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("open pixel status %d", rec.Code)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/emails/"+sendID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("detail after open: %d %#v", code, env)
	}
	if oc, _ := dataMap(t, env)["openCount"].(float64); oc < 1 {
		t.Fatalf("expected openCount>=1, got %#v", dataMap(t, env)["openCount"])
	}
	if ct, _ := dataMap(t, env)["clickTotal"].(float64); ct < 1 {
		t.Fatalf("expected clickTotal>=1, got %#v", dataMap(t, env)["clickTotal"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/public/commercial-mail/unsubscribe/"+openTok, nil)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unsubscribe status %d", rec.Code)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/emails", tok, map[string]any{
		"templateId": tplID,
	})
	if code != http.StatusConflict {
		t.Fatalf("expected opt-out conflict, got %d %#v", code, env)
	}
}
