package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/handlers"
)

func stubGoogleClaims(t *testing.T, claims handlers.GoogleIDTokenClaims) {
	t.Helper()
	handlers.TestSetGoogleIDTokenValidator(func(ctx context.Context, rawToken, clientID string) (handlers.GoogleIDTokenClaims, error) {
		if rawToken == "" || clientID == "" {
			return handlers.GoogleIDTokenClaims{}, errors.New("missing token or client id")
		}
		return claims, nil
	})
	t.Cleanup(func() { handlers.TestSetGoogleIDTokenValidator(nil) })
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
	handlers.TestSetGoogleIDTokenValidator(func(ctx context.Context, rawToken, clientID string) (handlers.GoogleIDTokenClaims, error) {
		return handlers.GoogleIDTokenClaims{}, errors.New("bad token")
	})
	t.Cleanup(func() { handlers.TestSetGoogleIDTokenValidator(nil) })

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
	stubGoogleClaims(t, handlers.GoogleIDTokenClaims{
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

func TestGoogleLoginCreateClientOK(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	email := uniqueEmail("google-new")
	sub := "sub-" + email
	stubGoogleClaims(t, handlers.GoogleIDTokenClaims{
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
}

func TestGoogleLoginClientOnlyForVetEmail(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetGoogleOAuthClientID("test-google-client.apps.googleusercontent.com")
	stubGoogleClaims(t, handlers.GoogleIDTokenClaims{
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

	stubGoogleClaims(t, handlers.GoogleIDTokenClaims{
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
