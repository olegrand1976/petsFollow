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
	notifier := email.NewNotifier("127.0.0.1", 9, "test@petsfollow.test")
	bill := billing.NewService(st, cfg)
	api := handlers.NewAPI(st, tokens, cfg, notifier, bill, nil)

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

func TestAuthResetInvalidToken(t *testing.T) {
	api := newTestAPI(t)
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/reset-password", map[string]any{
		"token": "nope", "password": "TestPass123!",
	})
	if code != http.StatusNotFound || errCode(env) != "not_found" {
		t.Fatalf("expected not_found, got %d %#v", code, env)
	}
}
