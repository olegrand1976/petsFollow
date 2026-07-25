package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type testAPI struct {
	handler http.Handler
	pool    *pgxpool.Pool
}

func loadDotEnv() {
	dir, _ := os.Getwd()
	for i := 0; i < 8; i++ {
		envPath := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				k, v := parts[0], parts[1]
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadDotEnv()
	// Les tests d'intégration s'appuient sur confirmPath/resetPath (exposés uniquement en env
	// dev/demo) et sur le gateway billing mock (opt-in explicite).
	_ = os.Setenv("DEV_SEED_ENABLED", "true")
	_ = os.Setenv("BILLING_MOCK_ENABLED", "true")
	// Pas de throttling dans la suite d'intégration (nombreux logins depuis la même IP httptest).
	_ = os.Setenv("AUTH_RATE_LIMIT_PER_MIN", "0")
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	migCtx, migCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migCancel()
	if err := db.Migrate(migCtx, pool); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}

	st := store.New(pool)
	tokens := authx.NewTokenIssuer(cfg.JWTSigningKey, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	notifier := email.NewNotifier("127.0.0.1", 9, "test@petsfollow.test", "http://localhost:3002", "https://ll-it-sc.be")
	bill := billing.NewService(st, cfg)
	api := handlers.NewAPI(st, tokens, cfg, notifier, bill, nil, nil)

	r := httpx.NewBaseRouter()
	r.Route("/api/v1", api.Routes)

	t.Cleanup(func() { pool.Close() })
	return &testAPI{handler: r, pool: pool}
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s+%d@petsfollow.test", prefix, time.Now().UnixNano())
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "fr")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}

func dataMap(t *testing.T, envelope map[string]any) map[string]any {
	t.Helper()
	d, ok := envelope["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", envelope)
	}
	return d
}

func errCode(envelope map[string]any) string {
	e, _ := envelope["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestAuthRegisterConfirmLoginForgotReset(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-auth")
	password := "TestPass123!"
	newPassword := "NewPass456!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Test", "practiceName": "Cabinet Test",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register status %d: %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	if !strings.Contains(confirmPath, "token=") {
		t.Fatalf("missing confirmPath: %#v", env)
	}
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusForbidden || errCode(env) != "email_not_verified" {
		t.Fatalf("expected email_not_verified, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm status %d: %#v", code, env)
	}
	if dataMap(t, env)["accessToken"] == nil {
		t.Fatal("expected accessToken after confirm")
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": "wrong-password",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusOK {
		t.Fatalf("login status %d: %#v", code, env)
	}

	// Email casing / espaces ne doivent pas bloquer le login
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "  " + strings.ToUpper(email) + "  ", "password": password,
	})
	if code != http.StatusOK {
		t.Fatalf("login with mixed-case/padded email status %d: %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/forgot-password", map[string]any{
		"email": email,
	})
	if code != http.StatusOK {
		t.Fatalf("forgot status %d: %#v", code, env)
	}
	resetPath, _ := dataMap(t, env)["resetPath"].(string)
	if !strings.Contains(resetPath, "token=") {
		t.Fatalf("missing resetPath: %#v", env)
	}
	resetToken := strings.TrimPrefix(resetPath, "/reset-password?token=")

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/forgot-password", map[string]any{
		"email": "nobody-exists@petsfollow.test",
	})
	if code != http.StatusOK {
		t.Fatalf("forgot unknown email should be 200, got %d %#v", code, env)
	}
	if _, ok := dataMap(t, env)["resetPath"]; ok {
		t.Fatal("resetPath must not be present for unknown email")
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/reset-password", map[string]any{
		"token": resetToken, "password": "short",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected password_too_short, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/reset-password", map[string]any{
		"token": resetToken, "password": newPassword,
	})
	if code != http.StatusOK {
		t.Fatalf("reset status %d: %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("old password should fail, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": newPassword,
	})
	if code != http.StatusOK {
		t.Fatalf("new password login status %d: %#v", code, env)
	}
}

func TestAuthRegisterDuplicateEmail(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-dup")
	body := map[string]any{
		"email": email, "password": "TestPass123!", "fullName": "Dr Dup", "practiceName": "Cabinet Dup",
		"consent": true,
	}
	code, _ := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", body)
	if code != http.StatusCreated {
		t.Fatalf("first register %d", code)
	}
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", body)
	if code != http.StatusConflict || errCode(env) != "conflict" {
		t.Fatalf("expected conflict, got %d %#v", code, env)
	}
}

// Le consentement CGU/privacy est requis au register (RGPD art. 7).
func TestAuthRegisterRequiresConsent(t *testing.T) {
	api := newTestAPI(t)
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": uniqueEmail("e2e-noconsent"), "password": "TestPass123!",
		"fullName": "Dr NoConsent", "practiceName": "Cabinet NoConsent",
	})
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("expected consent_required 400, got %d %#v", code, env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": uniqueEmail("e2e-noconsent-client"), "password": "TestPass123!",
		"fullName": "Client NoConsent",
	})
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("expected consent_required 400 (client), got %d %#v", code, env)
	}
}

func TestAuthResetInvalidToken(t *testing.T) {
	api := newTestAPI(t)
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/reset-password", map[string]any{
		"token": "nope", "password": "TestPass123!",
	})
	if code != http.StatusNotFound || errCode(env) != "not_found" {
		t.Fatalf("expected not_found, got %d %#v", code, env)
	}
}

// TestAuthSeedRoleLogins vérifie que les comptes seed Pro peuvent se connecter
// et que /me expose le bon rôle (garde-fou régression login / déploiement).
func TestAuthSeedRoleLogins(t *testing.T) {
	api := newTestAPI(t)
	cases := []struct {
		email, password, role string
	}{
		{"admin.demo@petsfollow.test", "AdminDemo123!", "admin"},
		{"vet.demo@petsfollow.test", "VetDemo123!", "vet"},
		{"commercial.demo@petsfollow.test", "CommercialDemo123!", "commercial"},
		{"commercial.manager@petsfollow.test", "CommercialDemo123!", "commercial_manager"},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
				"email": tc.email, "password": tc.password,
			})
			if code != http.StatusOK {
				t.Fatalf("login %s status %d: %#v (seed manquant ? make seed)", tc.email, code, env)
			}
			tok, _ := dataMap(t, env)["accessToken"].(string)
			if tok == "" {
				t.Fatal("expected accessToken")
			}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			req.Header.Set("Authorization", "Bearer "+tok)
			rec := httptest.NewRecorder()
			api.handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("me status %d: %s", rec.Code, rec.Body.String())
			}
			var envelope map[string]any
			_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
			me := dataMap(t, envelope)
			if me["role"] != tc.role {
				t.Fatalf("expected role %q, got %#v", tc.role, me["role"])
			}
			if me["email"] != tc.email {
				t.Fatalf("expected email %q, got %#v", tc.email, me["email"])
			}
		})
	}

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "vet.demo@petsfollow.test", "password": "WrongPass999!",
	})
	if code != http.StatusUnauthorized || errCode(env) != "unauthorized" {
		t.Fatalf("bad password: expected unauthorized, got %d %#v", code, env)
	}
}

func TestAuthRefresh(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-refresh")
	password := "TestPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Refresh", "practiceName": "Cabinet Refresh",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register status %d: %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm status %d: %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusOK {
		t.Fatalf("login status %d: %#v", code, env)
	}
	loginData := dataMap(t, env)
	refreshToken, _ := loginData["refreshToken"].(string)
	if refreshToken == "" {
		t.Fatal("expected refreshToken after login")
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": refreshToken,
	})
	if code != http.StatusOK {
		t.Fatalf("refresh status %d: %#v", code, env)
	}
	refreshed := dataMap(t, env)
	accessToken, _ := refreshed["accessToken"].(string)
	newRefresh, _ := refreshed["refreshToken"].(string)
	if accessToken == "" || newRefresh == "" {
		t.Fatalf("expected new token pair: %#v", refreshed)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept-Language", "fr")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status %d: %s", rec.Code, rec.Body.String())
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": "not-a-jwt",
	})
	if code != http.StatusUnauthorized || errCode(env) != "unauthorized" {
		t.Fatalf("expected unauthorized for bad refresh, got %d %#v", code, env)
	}
}
