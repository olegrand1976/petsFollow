package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
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

func TestVetReferralProspect(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	commEmail := uniqueEmail("ref-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Ref Comm",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID := dataMap(t, env)["userId"].(string)
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	vetEmail := uniqueEmail("ref-vet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Ref",
		"practiceName": "Cabinet Ref", "phone": "0102030405", "city": "Lyon", "postalCode": "69001", "addressLine1": "1 rue Ref",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	_ = commID

	vetTok := loginToken(t, api.handler, vetEmail, "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prospects", vetTok, map[string]any{
		"practiceName": "Cabinet Recommandé", "contactName": "Dr Y", "city": "Nantes",
	})
	if code != http.StatusCreated {
		t.Fatalf("vet referral %d %#v", code, env)
	}
	p := dataMap(t, env)
	if p["source"] != "vet_referral" {
		t.Fatalf("expected vet_referral source, got %#v", p)
	}
	if p["commercialUserId"] != commID {
		t.Fatalf("expected commercial %s, got %#v", commID, p)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	found := false
	for _, item := range env["data"].([]any) {
		m := item.(map[string]any)
		if m["practiceName"] == "Cabinet Recommandé" && m["source"] == "vet_referral" {
			found = true
		}
	}
	if !found {
		t.Fatalf("commercial should see vet referral, got %#v", env)
	}
}

func TestCommercialFlatSubscriptionAccrual(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	if err := st.EnsureDefaultCommissionTiers(ctx); err != nil {
		t.Fatal(err)
	}
	_ = st.EnsureCommissionSettings(ctx)
	_ = st.SetCommercialRateBps(ctx, store.DefaultCommercialCommissionRateBps)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	commEmail := uniqueEmail("flat-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Flat Comm",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID := dataMap(t, env)["userId"].(string)
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	vetEmail := uniqueEmail("flat-vet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Flat",
		"practiceName": "Cabinet Flat", "phone": "0102030405", "city": "Lyon", "postalCode": "69001", "addressLine1": "2 rue Flat",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	vetID := dataMap(t, env)["userId"].(string)

	var practiceID string
	if err := api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE id=$1`, vetID).Scan(&practiceID); err != nil {
		t.Fatal(err)
	}

	clientID := uuid.NewString()
	clientHash, _ := bcrypt.GenerateFromPassword([]byte("ClientDemo123!"), bcrypt.DefaultCost)
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, 'Client Flat', 'client', $4, NOW())`,
		clientID, uniqueEmail("flat-client"), string(clientHash), practiceID); err != nil {
		t.Fatal(err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceID, clientID, vetID); err != nil {
		t.Fatal(err)
	}

	petID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
		VALUES ($1, $2, $3, 'Rex Flat', 'dog', 'lab', 'pending_payment')`, petID, practiceID, clientID); err != nil {
		t.Fatal(err)
	}

	baseCents := 2900
	ent, err := st.CreateEntitlement(ctx, petID, clientID, "annual", "subscription", baseCents)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	until := now.Add(365 * 24 * time.Hour)
	if err := st.ActivateEntitlement(ctx, store.ActivateEntitlementParams{
		PetID: petID, Status: "active", ValidFrom: now, ValidUntil: until,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.AccrueCommissionForPetActivation(ctx, petID); err != nil {
		t.Fatal(err)
	}

	wantRate := store.CommercialRateBpsForPlan("annual")
	wantCommission := store.CommissionFromTTCCents(baseCents, wantRate)

	var sourceType string
	var rateBps, commissionCents int
	var commercialID string
	if err := api.pool.QueryRow(ctx, `
		SELECT source_type, rate_bps, commission_cents, commercial_user_id::text
		FROM billing.commercial_commission_ledger
		WHERE source_id=$1 AND source_type='subscription_pct'`, ent.ID,
	).Scan(&sourceType, &rateBps, &commissionCents, &commercialID); err != nil {
		t.Fatalf("commercial flat accrual missing: %v", err)
	}
	if commercialID != commID {
		t.Fatalf("commercial %s want %s", commercialID, commID)
	}
	if rateBps != wantRate || commissionCents != wantCommission {
		t.Fatalf("flat commission got rate=%d cents=%d want rate=%d cents=%d",
			rateBps, commissionCents, wantRate, wantCommission)
	}
}
