package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func doAuthJSON(t *testing.T, h http.Handler, method, path, token string, body any) (int, map[string]any) {
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}

func loginToken(t *testing.T, h http.Handler, email, password string) string {
	t.Helper()
	code, env := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code != http.StatusOK {
		t.Fatalf("login %s status %d: %#v", email, code, env)
	}
	token, _ := dataMap(t, env)["accessToken"].(string)
	if token == "" {
		t.Fatal("missing accessToken")
	}
	return token
}

func TestCommercialAdminCreateAndEncodeVet(t *testing.T) {
	api := newTestAPI(t)

	// Ensure admin exists (seed may not have run); create if needed
	adminEmail := "admin.demo@petsfollow.test"
	adminPass := "AdminDemo123!"
	var n int
	_ = api.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM identity.users WHERE email=$1`, adminEmail).Scan(&n)
	if n == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		_, err := api.pool.Exec(context.Background(), `
			INSERT INTO identity.users (id, email, password_hash, full_name, role, email_verified_at)
			VALUES ($1, $2, $3, 'Admin Ops', 'admin', NOW())`, uuid.NewString(), adminEmail, string(hash))
		if err != nil {
			t.Fatal(err)
		}
	}

	adminTok := loginToken(t, api.handler, adminEmail, adminPass)

	commEmail := uniqueEmail("commercial")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Cam Test",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}

	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	// Commercial forbidden on admin metrics
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/metrics/overview", commTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("commercial should be forbidden on admin, got %d", code)
	}

	vetEmail := uniqueEmail("encoded-vet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Encoded",
		"practiceName": "Cabinet Encoded", "phone": "0102030405", "city": "Paris", "postalCode": "75001", "addressLine1": "1 rue Test",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/vets", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list vets %d %#v", code, env)
	}
	list, ok := env["data"].([]any)
	if !ok || len(list) < 1 {
		t.Fatalf("expected assigned vets list, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", commTok, map[string]any{
		"practiceName": "Prospect Test", "contactName": "Dr X", "status": "new",
	})
	if code != http.StatusCreated {
		t.Fatalf("create prospect %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/overview", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov := dataMap(t, env)
	if int(ov["assignedVets"].(float64)) < 1 {
		t.Fatalf("expected assignedVets >= 1, got %#v", ov)
	}

	// Vet cannot access commercial routes
	vetTok := loginToken(t, api.handler, vetEmail, "VetDemo123!")
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/overview", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet should be forbidden on commercial, got %d", code)
	}

	// Admin sees all prospects
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/prospects", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin prospects %d %#v", code, env)
	}
}

func TestCommercialRBACIsolation(t *testing.T) {
	api := newTestAPI(t)
	adminEmail := "admin.demo@petsfollow.test"
	var n int
	_ = api.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM identity.users WHERE email=$1`, adminEmail).Scan(&n)
	if n == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("AdminDemo123!"), bcrypt.DefaultCost)
		_, _ = api.pool.Exec(context.Background(), `
			INSERT INTO identity.users (id, email, password_hash, full_name, role, email_verified_at)
			VALUES ($1, $2, $3, 'Admin Ops', 'admin', NOW())`, uuid.NewString(), adminEmail, string(hash))
	}
	adminTok := loginToken(t, api.handler, adminEmail, "AdminDemo123!")

	c1 := uniqueEmail("c1")
	c2 := uniqueEmail("c2")
	for _, email := range []string{c1, c2} {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
			"email": email, "password": "CommercialDemo123!", "fullName": email,
		})
		if code != http.StatusCreated {
			t.Fatalf("create %s: %d %#v", email, code, env)
		}
	}
	tok1 := loginToken(t, api.handler, c1, "CommercialDemo123!")
	tok2 := loginToken(t, api.handler, c2, "CommercialDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", tok1, map[string]any{
		"practiceName": "Only C1",
	})
	if code != http.StatusCreated {
		t.Fatalf("prospect %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects", tok2, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	for _, item := range list {
		m, _ := item.(map[string]any)
		if m["practiceName"] == "Only C1" {
			t.Fatal("commercial 2 must not see commercial 1 prospect")
		}
	}
}
