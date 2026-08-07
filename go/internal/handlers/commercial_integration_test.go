package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func prospectListItems(t *testing.T, env map[string]any) []any {
	t.Helper()
	data, ok := env["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected prospect page object, got %#v", env["data"])
	}
	items, _ := data["items"].([]any)
	return items
}

func loginToken(t *testing.T, h http.Handler, email, password string) string {
	return loginTokenOpts(t, h, email, password, true)
}

// loginTokenRaw logs in without auto-accepting CGU (for RGPD consent-gate tests).
func loginTokenRaw(t *testing.T, h http.Handler, email, password string) string {
	return loginTokenOpts(t, h, email, password, false)
}

func loginTokenOpts(t *testing.T, h http.Handler, email, password string, acceptClientTerms bool) string {
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
	if !acceptClientTerms {
		return token
	}
	// Provisioned clients (vet/commercial create) start with termsAcceptedAt null;
	// Flutter gates on accept-terms — most integration tests accept once after login.
	code, env = doAuthJSON(t, h, http.MethodGet, "/api/v1/me", token, nil)
	if code == http.StatusOK {
		me := dataMap(t, env)
		if role, _ := me["role"].(string); role == "client" {
			if me["termsAcceptedAt"] == nil || me["termsAcceptedAt"] == "" {
				code, env = doAuthJSON(t, h, http.MethodPost, "/api/v1/me/accept-terms", token, map[string]any{
					"consent": true,
				})
				if code != http.StatusOK {
					t.Fatalf("accept-terms after login %s: %d %#v", email, code, env)
				}
			}
		}
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

func TestCommercialCreateClientsLinkedAndStandalone(t *testing.T) {
	api := newTestAPI(t)
	adminEmail := "admin.demo@petsfollow.test"
	var n int
	_ = api.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM identity.users WHERE email=$1`, adminEmail).Scan(&n)
	if n == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("AdminDemo123!"), bcrypt.DefaultCost)
		_, err := api.pool.Exec(context.Background(), `
			INSERT INTO identity.users (id, email, password_hash, full_name, role, email_verified_at)
			VALUES ($1, $2, $3, 'Admin Ops', 'admin', NOW())`, uuid.NewString(), adminEmail, string(hash))
		if err != nil {
			t.Fatal(err)
		}
	}
	adminTok := loginToken(t, api.handler, adminEmail, "AdminDemo123!")
	commEmail := uniqueEmail("comm-clients")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Comm Clients",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	vetEmail := uniqueEmail("vet-for-client")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Client",
		"practiceName": "Cabinet Client", "city": "Lyon",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	vetID, _ := env["data"].(map[string]any)["userId"].(string)
	if vetID == "" {
		t.Fatalf("missing vet userId %#v", env)
	}

	linkedEmail := uniqueEmail("linked-client")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", commTok, map[string]any{
		"email": linkedEmail, "password": "ClientDemo123!", "fullName": "Linked Client", "vetUserId": vetID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create linked client %d %#v", code, env)
	}

	standaloneEmail := uniqueEmail("solo-client")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", commTok, map[string]any{
		"email": standaloneEmail, "password": "ClientDemo123!", "fullName": "Solo Client",
	})
	if code != http.StatusCreated {
		t.Fatalf("create standalone client %d %#v", code, env)
	}
	var practiceID *string
	err := api.pool.QueryRow(context.Background(), `
		SELECT practice_id::text FROM identity.users WHERE email=$1`, standaloneEmail).Scan(&practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if practiceID != nil {
		t.Fatalf("standalone client should have NULL practice_id, got %v", *practiceID)
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
	for _, item := range prospectListItems(t, env) {
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
	for _, item := range prospectListItems(t, env) {
		m := item.(map[string]any)
		if m["practiceName"] == "Cabinet Recommandé" && m["source"] == "vet_referral" {
			found = true
		}
	}
	if !found {
		t.Fatalf("commercial should see vet referral, got %#v", env)
	}
}

func TestCommercialDirectoryProspectsPagination(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	commEmail := uniqueEmail("dir-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Dir Comm",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	marker := "BCE-DIR-" + uuid.NewString()[:8]
	for i := range 3 {
		_, err := api.pool.Exec(ctx, `
			INSERT INTO sales.prospects (
				id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
				city, notes, source, status
			) VALUES ($1, NULL, $2, '', '', '', $3, $4, 'directory', 'new')`,
			uuid.NewString(),
			"Cabinet "+marker+" "+string(rune('A'+i)),
			"City"+string(rune('A'+i)),
			marker+" notes",
		)
		if err != nil {
			t.Fatalf("insert directory prospect: %v", err)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/commercial/prospects?source=directory&q="+marker+"&limit=2&offset=0", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list directory %d %#v", code, env)
	}
	page := dataMap(t, env)
	total, _ := page["total"].(float64)
	if int(total) != 3 {
		t.Fatalf("expected total 3, got %#v", page)
	}
	items := prospectListItems(t, env)
	if len(items) != 2 {
		t.Fatalf("expected page size 2, got %#v", items)
	}
	for _, item := range items {
		m := item.(map[string]any)
		if m["source"] != "directory" {
			t.Fatalf("expected directory source, got %#v", m)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/commercial/prospects?source=directory&q="+marker+"&limit=2&offset=2", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list page 2 %d %#v", code, env)
	}
	items = prospectListItems(t, env)
	if len(items) != 1 {
		t.Fatalf("expected remaining 1, got %#v", items)
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

func TestCommercialBonusMixAndMarkPaid(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	_ = st.EnsureDefaultCommissionTiers(ctx)
	_ = st.EnsureCommissionSettings(ctx)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	commEmail := uniqueEmail("bonus-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Bonus Comm",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID := dataMap(t, env)["userId"].(string)
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	vetEmail := uniqueEmail("bonus-vet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Bonus",
		"practiceName": "Cabinet Bonus", "phone": "0102030405", "city": "Lyon", "postalCode": "69001", "addressLine1": "3 rue Bonus",
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
		VALUES ($1, $2, $3, 'Client Bonus', 'client', $4, NOW())`,
		clientID, uniqueEmail("bonus-client"), string(clientHash), practiceID); err != nil {
		t.Fatal(err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceID, clientID, vetID); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	until := now.Add(1095 * 24 * time.Hour)
	for range 2 {
		petID := uuid.NewString()
		if _, err := api.pool.Exec(ctx, `
			INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
			VALUES ($1, $2, $3, $4, 'dog', 'lab', 'pending_payment')`,
			petID, practiceID, clientID, "Pet Bonus "+uuid.NewString()[:8]); err != nil {
			t.Fatal(err)
		}
		if _, err := st.CreateEntitlement(ctx, petID, clientID, "triennial", "subscription", 9500); err != nil {
			t.Fatal(err)
		}
		if err := st.ActivateEntitlement(ctx, store.ActivateEntitlementParams{
			PetID: petID, Status: "active", ValidFrom: now, ValidUntil: until,
		}); err != nil {
			t.Fatal(err)
		}
		if err := st.AccrueCommissionForPetActivation(ctx, petID); err != nil {
			t.Fatal(err)
		}
	}

	if err := st.SyncCommercialBonusAwards(ctx, commID); err != nil {
		t.Fatal(err)
	}

	var mixCount int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM billing.commercial_bonus_awards
		WHERE commercial_user_id=$1 AND status='earned' AND bonus_code='commercial_mix'`, commID).Scan(&mixCount); err != nil {
		t.Fatal(err)
	}
	if mixCount != 1 {
		t.Fatalf("expected 1 mix award, got %d", mixCount)
	}

	if err := st.SyncCommercialBonusAwards(ctx, commID); err != nil {
		t.Fatal(err)
	}
	var totalAwards int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM billing.commercial_bonus_awards WHERE commercial_user_id=$1`, commID).Scan(&totalAwards); err != nil {
		t.Fatal(err)
	}
	if totalAwards != 1 {
		t.Fatalf("expected 1 award after re-sync, got %d", totalAwards)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/commissions", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("commissions %d %#v", code, env)
	}
	summary := dataMap(t, env)
	bonuses, ok := summary["bonuses"].([]any)
	if !ok || len(bonuses) < 1 {
		t.Fatalf("expected bonuses in summary, got %#v", summary["bonuses"])
	}
	foundMix := false
	for _, raw := range bonuses {
		b := raw.(map[string]any)
		if b["code"] == "commercial_mix" {
			foundMix = true
			if b["status"] != "earned" {
				t.Fatalf("mix status want earned, got %#v", b["status"])
			}
		}
		if b["code"] == "commercial_ramp" {
			t.Fatalf("commercial_ramp must not appear in bonuses %#v", bonuses)
		}
	}
	if !foundMix {
		t.Fatalf("missing mix in bonuses %#v", bonuses)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/commercial-bonuses", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin bonuses %d %#v", code, env)
	}
	adminData := dataMap(t, env)
	items, ok := adminData["items"].([]any)
	if !ok || len(items) < 1 {
		t.Fatalf("expected admin bonus items, got %#v", adminData)
	}
	if _, ok := adminData["trend"].([]any); !ok {
		t.Fatalf("expected trend series, got %#v", adminData["trend"])
	}
	if _, ok := adminData["comparison"].([]any); !ok {
		t.Fatalf("expected comparison, got %#v", adminData["comparison"])
	}
	if adminData["periodYm"] == nil || adminData["periodYm"] == "" {
		t.Fatalf("expected periodYm, got %#v", adminData["periodYm"])
	}
	kpi, ok := adminData["kpi"].(map[string]any)
	if !ok {
		t.Fatalf("expected kpi object, got %#v", adminData["kpi"])
	}
	if kpi["metCount"] == nil {
		t.Fatalf("expected kpi.metCount, got %#v", kpi)
	}

	var mixAwardID string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text FROM billing.commercial_bonus_awards
		WHERE commercial_user_id=$1 AND bonus_code='commercial_mix'`, commID).Scan(&mixAwardID); err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercial-bonuses/"+mixAwardID+"/mark-paid", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("mark paid %d %#v", code, env)
	}
	paid := dataMap(t, env)
	if paid["status"] != "paid" {
		t.Fatalf("expected paid status, got %#v", paid)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercial-bonuses/"+mixAwardID+"/mark-paid", adminTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("second mark-paid want 409, got %d %#v", code, env)
	}
}

func TestCommercialManagerTeamVisibility(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	mgrEmail := uniqueEmail("mgr")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": mgrEmail, "password": "CommercialDemo123!", "fullName": "Mgr Test", "role": "commercial_manager",
	})
	if code != http.StatusCreated {
		t.Fatalf("create manager %d %#v", code, env)
	}
	mgrID := dataMap(t, env)["userId"].(string)

	repEmail := uniqueEmail("rep")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": repEmail, "password": "CommercialDemo123!", "fullName": "Rep Test", "managerUserId": mgrID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create rep %d %#v", code, env)
	}

	otherEmail := uniqueEmail("other-rep")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": otherEmail, "password": "CommercialDemo123!", "fullName": "Other Rep",
	})
	if code != http.StatusCreated {
		t.Fatalf("create other %d %#v", code, env)
	}

	mgrTok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")
	repTok := loginToken(t, api.handler, repEmail, "CommercialDemo123!")
	otherTok := loginToken(t, api.handler, otherEmail, "CommercialDemo123!")

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", repTok, map[string]any{
		"practiceName": "Team Prospect", "status": "contacted",
	})
	if code != http.StatusCreated {
		t.Fatalf("rep prospect %d %#v", code, env)
	}
	prospectID := dataMap(t, env)["id"].(string)
	if dataMap(t, env)["firstContactedAt"] == nil {
		t.Fatalf("expected firstContactedAt on contacted create/update path, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", otherTok, map[string]any{
		"practiceName": "Outside Team",
	})
	if code != http.StatusCreated {
		t.Fatalf("other prospect %d %#v", code, env)
	}

	// Manager sees team member, not self, in team list
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/team", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("team %d %#v", code, env)
	}
	team, _ := env["data"].([]any)
	if len(team) != 1 {
		t.Fatalf("expected 1 team member, got %#v", env)
	}
	member := team[0].(map[string]any)
	if member["email"] != repEmail {
		t.Fatalf("expected rep in team, got %#v", member)
	}
	if member["userId"] == mgrID {
		t.Fatal("manager must not appear in team list")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/prospects", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("manager prospects %d %#v", code, env)
	}
	foundTeam, foundOutside := false, false
	for _, item := range env["data"].([]any) {
		m := item.(map[string]any)
		if m["practiceName"] == "Team Prospect" {
			foundTeam = true
		}
		if m["practiceName"] == "Outside Team" {
			foundOutside = true
		}
	}
	if !foundTeam {
		t.Fatal("manager should see team prospect")
	}
	if foundOutside {
		t.Fatal("manager must not see outside-team prospect")
	}

	// Peer commercial cannot access manager APIs
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/overview", repTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("rep forbidden on manager overview, got %d", code)
	}

	// Manager can patch team prospect status / appointment
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/commercial-manager/prospects/"+prospectID, mgrTok, map[string]any{
		"status": "qualified", "appointmentAt": time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})
	if code != http.StatusOK {
		t.Fatalf("manager patch %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "qualified" {
		t.Fatalf("expected qualified, got %#v", env)
	}

	// Overview self block present and team totals exclude manager production
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/overview", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov := dataMap(t, env)
	if ov["self"] == nil {
		t.Fatalf("expected self block, got %#v", ov)
	}

	// Invalid manager must not leave orphan commercial
	orphanEmail := uniqueEmail("orphan")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": orphanEmail, "password": "CommercialDemo123!", "fullName": "Orphan",
		"managerUserId": uuid.NewString(),
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid manager want 400, got %d %#v", code, env)
	}
	var n int
	_ = api.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM identity.users WHERE email=$1`, orphanEmail).Scan(&n)
	if n != 0 {
		t.Fatalf("orphan commercial created despite invalid manager")
	}

	// Filter commercialUserId outside team → 404
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/prospects?commercialUserId="+uuid.NewString(), mgrTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("outside team filter want 404, got %d", code)
	}
}

func TestProspectFirstEncodingClaimAndRelease(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	mgrEmail := uniqueEmail("claim-mgr")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": mgrEmail, "password": "CommercialDemo123!", "fullName": "Claim Manager", "role": "commercial_manager",
	})
	if code != http.StatusCreated {
		t.Fatalf("create manager %d %#v", code, env)
	}
	mgrID := dataMap(t, env)["userId"].(string)

	aEmail := uniqueEmail("claim-a")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": aEmail, "password": "CommercialDemo123!", "fullName": "Rep A", "managerUserId": mgrID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create A %d %#v", code, env)
	}
	bEmail := uniqueEmail("claim-b")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": bEmail, "password": "CommercialDemo123!", "fullName": "Rep B", "managerUserId": mgrID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create B %d %#v", code, env)
	}

	aTok := loginToken(t, api.handler, aEmail, "CommercialDemo123!")
	bTok := loginToken(t, api.handler, bEmail, "CommercialDemo123!")
	mgrTok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")

	phone := fmt.Sprintf("+32470%06d", time.Now().UnixNano()%1000000)
	practice := "Cabinet Claim " + uuid.NewString()[:8]
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", aTok, map[string]any{
		"practiceName": practice, "contactPhone": phone, "city": "Bruxelles",
	})
	if code != http.StatusCreated {
		t.Fatalf("create prospect %d %#v", code, env)
	}
	prospectID := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects/lookup?q="+phone, bTok, nil)
	if code != http.StatusOK {
		t.Fatalf("lookup %d %#v", code, env)
	}
	look := dataMap(t, env)
	if look["status"] != "owned" {
		t.Fatalf("want owned, got %#v", look)
	}
	if look["ownerName"] != "Rep A" {
		t.Fatalf("ownerName %#v", look["ownerName"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", bTok, map[string]any{
		"practiceName": practice, "contactPhone": phone, "city": "Bruxelles",
	})
	if code != http.StatusConflict {
		t.Fatalf("duplicate create want 409, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/claim", bTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("claim owned want 409, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial-manager/prospects/"+prospectID+"/release", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("release %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects/lookup?q="+url.QueryEscape(practice), bTok, nil)
	if code != http.StatusOK {
		t.Fatalf("lookup free %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "free" {
		t.Fatalf("want free after release, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/claim", bTok, nil)
	if code != http.StatusOK {
		t.Fatalf("claim free %d %#v", code, env)
	}
	if dataMap(t, env)["commercialUserId"] == "" {
		t.Fatalf("expected owner after claim %#v", env)
	}
}

func TestVetRegisterInviteCodeAndUnassignedPool(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	commEmail := uniqueEmail("invite-comm")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Invite Comm",
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID := dataMap(t, env)["userId"].(string)
	inv, err := store.New(api.pool).EnsureAppInviteCode(context.Background(), commID)
	if err != nil {
		t.Fatal(err)
	}

	vetWith := uniqueEmail("vet-with-code")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": vetWith, "password": "VetDemo123!", "fullName": "Véto Code",
		"practiceName": "Cabinet Code", "consent": true, "inviteCode": inv.Code,
	})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("register with code %d %#v", code, env)
	}
	var assigned string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`, vetWith).Scan(&assigned); err != nil {
		t.Fatal(err)
	}
	if assigned != commID {
		t.Fatalf("assigned=%s want %s", assigned, commID)
	}

	vetPool := uniqueEmail("vet-pool")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": vetPool, "password": "VetDemo123!", "fullName": "Véto Pool",
		"practiceName": "Cabinet Pool", "consent": true,
	})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("register without code %d %#v", code, env)
	}
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`, vetPool).Scan(&assigned); err != nil {
		t.Fatal(err)
	}
	if assigned != "" {
		t.Fatalf("want empty assigned for pool, got %s", assigned)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/vets/unassigned", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassigned list %d %#v", code, env)
	}
	found := false
	switch rows := env["data"].(type) {
	case []any:
		for _, raw := range rows {
			m, _ := raw.(map[string]any)
			if m["email"] == vetPool {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("pool vet not listed %#v", env["data"])
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": uniqueEmail("vet-bad"), "password": "VetDemo123!", "fullName": "Bad",
		"practiceName": "X", "consent": true, "inviteCode": "INVALID1",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("bad invite want 400, got %d %#v", code, env)
	}
}
