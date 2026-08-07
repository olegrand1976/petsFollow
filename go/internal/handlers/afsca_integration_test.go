package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/afsca"
)

func TestAfscaNewslettersBE(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	// Heal leftover pollution from older tests that mutated the shared demo practice.
	_, err := api.pool.Exec(t.Context(), `
		UPDATE practice.practices
		SET country_code = 'BE', animal_scope = 'both'
		WHERE id = (SELECT practice_id FROM identity.users WHERE email = 'vet.demo@petsfollow.test')`)
	if err != nil {
		t.Fatal(err)
	}

	fixture := `
<html><body><div class="paragraph__body"><p>07/08/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter559_fr.asp" target="_blank">Newsletter N°559</a><br>
Cas suspect Nil occidental</p>
<hr>
<p>08/07/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter557_fr.asp" target="_blank">Newsletter N°557</a><br>
Rage</p></div></body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(srv.Close)

	api.api.SetAfscaClient(&afsca.Client{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/afsca-newsletters?limit=5", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("status %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected items, got %#v", data)
	}
	first, _ := items[0].(map[string]any)
	if title, _ := first["title"].(string); !strings.Contains(title, "Nil") {
		t.Fatalf("title %#v", first)
	}
	if src, _ := data["sourceUrl"].(string); !strings.HasPrefix(src, srv.URL) {
		t.Fatalf("sourceUrl %q", src)
	}

	// Cap limit (handler clamps to MaxLimit; response still OK).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/afsca-newsletters?limit=999", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("limit clamp %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	items, _ = data["items"].([]any)
	if len(items) > afsca.MaxLimit {
		t.Fatalf("items len %d > MaxLimit", len(items))
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/overview", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov, _ := env["data"].(map[string]any)
	if cc, _ := ov["countryCode"].(string); cc != "BE" {
		t.Fatalf("countryCode %#v", ov["countryCode"])
	}
}

func TestAfscaNewslettersScopeFilter(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	_, err := api.pool.Exec(t.Context(), `
		UPDATE practice.practices
		SET country_code = 'BE', animal_scope = 'small'
		WHERE id = (SELECT practice_id FROM identity.users WHERE email = 'vet.demo@petsfollow.test')`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(t.Context(), `
			UPDATE practice.practices
			SET animal_scope = 'both'
			WHERE id = (SELECT practice_id FROM identity.users WHERE email = 'vet.demo@petsfollow.test')`)
	})

	fixture := `
<html><body><div class="paragraph__body">
<p>07/08/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter559_fr.asp" target="_blank">Newsletter N°559</a><br>
Cas suspect Nil occidental chez un cheval</p>
<hr>
<p>08/07/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter557_fr.asp" target="_blank">Newsletter N°557</a><br>
Rage : vigilance chiens importés</p>
</div></body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(srv.Close)
	api.api.SetAfscaClient(&afsca.Client{BaseURL: srv.URL, HTTPClient: srv.Client()})

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/afsca-newsletters?limit=10", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("status %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("small scope should keep dog item only (+not horse), got %#v", items)
	}
	first, _ := items[0].(map[string]any)
	if title, _ := first["title"].(string); !strings.Contains(strings.ToLower(title), "chien") {
		t.Fatalf("expected dog newsletter, got %#v", first)
	}
	if scope, _ := first["scope"].(string); scope != "small" {
		t.Fatalf("scope %#v", first["scope"])
	}
}

func TestAfscaNewslettersNonBE(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("afsca-fr")
	password := "TestPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr AFSCA FR", "practiceName": "Cabinet FR",
		"consent": true,
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
	vetTok, _ := dataMap(t, env)["accessToken"].(string)
	if vetTok == "" {
		t.Fatal("missing token")
	}

	_, err := api.pool.Exec(t.Context(), `
		UPDATE practice.practices
		SET country_code = 'FR'
		WHERE id = (SELECT practice_id FROM identity.users WHERE email = $1)`, email)
	if err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/afsca-newsletters", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("non-BE want 404, got %d %#v", code, env)
	}
	if errCode(env) != "afsca_not_available" {
		t.Fatalf("error %#v", env["error"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/overview", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov, _ := env["data"].(map[string]any)
	if cc, _ := ov["countryCode"].(string); cc != "FR" {
		t.Fatalf("countryCode %#v", ov["countryCode"])
	}
}
