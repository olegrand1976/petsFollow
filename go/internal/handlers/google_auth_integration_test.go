package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/handlers"
)

func stubGoogleClaims(t *testing.T, api *testAPI, claims handlers.GoogleIDTokenClaims) {
	t.Helper()
	api.api.TestSetGoogleIDTokenValidator(func(ctx context.Context, rawToken, clientID string) (handlers.GoogleIDTokenClaims, error) {
		if rawToken == "" || clientID == "" {
			return handlers.GoogleIDTokenClaims{}, errors.New("missing token or client id")
		}
		return claims, nil
	})
	t.Cleanup(func() { api.api.TestSetGoogleIDTokenValidator(nil) })
}

func TestGoogleLoginNotConfigured(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "x",
		"audience": "client",
		"consent":  true,
	})
	if code != http.StatusNotImplemented {
		t.Fatalf("want 501 got %d %#v", code, env)
	}
	if errMsgKey(env) != "not_configured" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestGoogleLoginInvalidToken(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	api.api.TestSetGoogleIDTokenValidator(func(ctx context.Context, rawToken, clientID string) (handlers.GoogleIDTokenClaims, error) {
		return handlers.GoogleIDTokenClaims{}, errors.New("bad token")
	})
	t.Cleanup(func() { api.api.TestSetGoogleIDTokenValidator(nil) })

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "bad",
		"audience": "client",
		"consent":  true,
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d %#v", code, env)
	}
	if errMsgKey(env) != "invalid_google_token" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestGoogleLoginCreateClientRequiresConsent(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-noconsent")
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-" + email, Email: email, EmailVerified: true, Name: "Google User",
	})

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "client",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d %#v", code, env)
	}
	if errMsgKey(env) != "consent_required" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestGoogleLoginCreateProRequiresConsent(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-pro-noconsent")
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-" + email, Email: email, EmailVerified: true, Name: "Google Vet",
	})

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "pro",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d %#v", code, env)
	}
	if errMsgKey(env) != "consent_required" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestGoogleLoginExistingProWithoutConsent(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-pro-existing")
	password := "VetDemo123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Existing Vet",
		"practiceName": "Cab Existing Google", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	if token == "" || token == confirmPath {
		t.Fatalf("confirmPath %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}

	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-" + email, Email: email, EmailVerified: true, Name: "Existing Vet",
	})
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "pro",
	})
	if code != http.StatusOK {
		t.Fatalf("google existing pro without consent want 200 got %d %#v", code, env)
	}
	if _, ok := dataMap(t, env)["accessToken"].(string); !ok {
		t.Fatalf("tokens missing %#v", env)
	}
}

func TestGoogleLoginExistingClientWithoutConsent(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-client-existing")
	password := "ClientDemo123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Existing Client", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}

	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-" + email, Email: email, EmailVerified: true, Name: "Existing Client",
	})
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "client",
	})
	if code != http.StatusOK {
		t.Fatalf("google existing client without consent want 200 got %d %#v", code, env)
	}
	if _, ok := dataMap(t, env)["accessToken"].(string); !ok {
		t.Fatalf("tokens missing %#v", env)
	}
}

func TestGoogleLoginCreateProOK(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-pro-new")
	sub := "sub-" + email
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: sub, Email: email, EmailVerified: true, Name: "Google Vet",
	})

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "pro",
		"consent":  true,
	})
	if code != http.StatusOK {
		t.Fatalf("want 200 got %d %#v", code, env)
	}
	data := dataMap(t, env)
	tok, _ := data["accessToken"].(string)
	if tok == "" || data["refreshToken"] == nil {
		t.Fatalf("tokens missing %#v", data)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["role"] != "vet" {
		t.Fatalf("role %#v", me)
	}
	if me["termsAcceptedAt"] == nil || me["termsAcceptedAt"] == "" {
		t.Fatalf("termsAcceptedAt missing %#v", me)
	}
}

func TestGoogleLoginCreateClientOK(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-new")
	sub := "sub-" + email
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: sub, Email: email, EmailVerified: true, Name: "Google New",
	})

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "client",
		"consent":  true,
	})
	if code != http.StatusOK {
		t.Fatalf("want 200 got %d %#v", code, env)
	}
	data := dataMap(t, env)
	tok, _ := data["accessToken"].(string)
	if tok == "" || data["refreshToken"] == nil {
		t.Fatalf("tokens missing %#v", data)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["role"] != "client" {
		t.Fatalf("role %#v", me)
	}
	if me["emailVerified"] != true {
		t.Fatalf("emailVerified missing %#v", me)
	}
	if me["email"] != email {
		t.Fatalf("email %#v", me)
	}
	if me["termsAcceptedAt"] == nil || me["termsAcceptedAt"] == "" {
		t.Fatalf("termsAcceptedAt missing %#v", me)
	}
}

func TestGoogleLoginClientOnlyForVetEmail(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-vet-demo-google", Email: "vet.demo@petsfollow.test",
		EmailVerified: true, Name: "Vet Demo",
	})

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "client",
		"consent":  true,
	})
	if code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %#v", code, env)
	}
	if errMsgKey(env) != "google_client_only" {
		t.Fatalf("error %#v", env["error"])
	}
}

func TestGoogleLoginLinksUnverifiedClientAndVerifies(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-unverified")
	password := "ClientDemo123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Unverified Client", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusForbidden || errCode(env) != "email_not_verified" {
		t.Fatalf("precondition login want email_not_verified got %d %#v", code, env)
	}

	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-" + email, Email: email, EmailVerified: true, Name: "Unverified Client",
	})
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "client",
		"consent":  true,
	})
	if code != http.StatusOK {
		t.Fatalf("google link want 200 got %d %#v", code, env)
	}
	tok, _ := dataMap(t, env)["accessToken"].(string)
	if tok == "" {
		t.Fatalf("tokens missing %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["emailVerified"] != true {
		t.Fatalf("email still unverified %#v", me)
	}
	if me["googleLinked"] != true {
		t.Fatalf("googleLinked %#v", me)
	}
}

func TestGoogleLoginProAudienceDoesNotSetTermsViaConsent(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-pro")
	// Create a vet-less path: audience pro with unknown email creates a vet (RegisterGoogleVet).
	// We only assert that consent on Pro path does not rely on client terms semantics —
	// here we link an existing client via wrong audience first is forbidden; instead
	// create client without terms then google as pro → google_pro_only.
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": "ClientDemo123!", "fullName": "Client", "consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	// Clear terms to simulate provisioned-like state.
	if _, err := api.pool.Exec(t.Context(), `
		UPDATE identity.users SET terms_accepted_at = NULL WHERE email = $1`, email); err != nil {
		t.Fatal(err)
	}
	stubGoogleClaims(t, api, handlers.GoogleIDTokenClaims{
		Subject: "sub-pro-" + email, Email: email, EmailVerified: true, Name: "Client",
	})
	// Pro audience on a client role → google_pro_only (no terms write).
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"idToken":  "tok",
		"audience": "pro",
		"consent":  true,
	})
	if code != http.StatusForbidden || errMsgKey(env) != "google_pro_only" {
		t.Fatalf("want google_pro_only got %d %#v", code, env)
	}
	var terms any
	if err := api.pool.QueryRow(t.Context(), `
		SELECT terms_accepted_at FROM identity.users WHERE email = $1`, email).Scan(&terms); err != nil {
		t.Fatal(err)
	}
	if terms != nil {
		t.Fatalf("pro path must not set terms_accepted_at, got %#v", terms)
	}
}
