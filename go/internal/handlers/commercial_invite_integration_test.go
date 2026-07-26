package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterClientWithCommercialInviteCode(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	hash, err := bcrypt.GenerateFromPassword([]byte("CommercialDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	commID := uuid.NewString()
	commEmail := fmt.Sprintf("smoke-inv-comm+%d@petsfollow.test", time.Now().UnixNano())
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at
		) VALUES ($1, $2, $3, 'Invite Comm', 'commercial', NOW())`,
		commID, commEmail, string(hash))
	if err != nil {
		t.Fatal(err)
	}

	inv, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Code == "" {
		t.Fatal("empty invite code")
	}

	clientEmail := fmt.Sprintf("smoke-inv-client+%d@petsfollow.test", time.Now().UnixNano())
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email":      clientEmail,
		"password":   "ClientDemo123!",
		"fullName":   "Invite Client",
		"consent":    true,
		"inviteCode": inv.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client status %d: %#v", code, env)
	}
	data := dataMap(t, env)
	if data["inviteStatus"] != "referred" {
		t.Fatalf("inviteStatus=%v want referred; env=%#v", data["inviteStatus"], env)
	}

	var clientID, referral, savedCode string
	err = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	if err != nil {
		t.Fatal(err)
	}
	err = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text, COALESCE(invite_code,'')
		FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&referral, &savedCode)
	if err != nil {
		t.Fatal(err)
	}
	if referral != commID {
		t.Fatalf("referral=%q want %q", referral, commID)
	}
	if savedCode != inv.Code {
		t.Fatalf("invite_code=%q want %q", savedCode, inv.Code)
	}

	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/referrals", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list referrals %d %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	found := false
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		if m["clientUserId"] == clientID {
			found = true
			if m["inviteCode"] != inv.Code {
				t.Fatalf("list inviteCode=%v want %s", m["inviteCode"], inv.Code)
			}
		}
	}
	if !found {
		t.Fatalf("client not in referrals list: %#v", rows)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/overview", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov := dataMap(t, env)
	if int(ov["referredClients"].(float64)) < 1 {
		t.Fatalf("referredClients=%v want >= 1", ov["referredClients"])
	}

	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.commercial_referrals WHERE client_user_id=$1`, clientID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.app_invite_codes WHERE user_id=$1`, commID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.users WHERE id IN ($1, $2)`, clientID, commID)
	})
}

func TestClaimCommercialInviteViaEndpoint(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	hash, err := bcrypt.GenerateFromPassword([]byte("CommercialDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	commID := uuid.NewString()
	commEmail := fmt.Sprintf("smoke-claim-comm+%d@petsfollow.test", time.Now().UnixNano())
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at
		) VALUES ($1, $2, $3, 'Claim Comm', 'commercial', NOW())`,
		commID, commEmail, string(hash))
	if err != nil {
		t.Fatal(err)
	}
	inv, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := fmt.Sprintf("smoke-claim-client+%d@petsfollow.test", time.Now().UnixNano())
	clientHash, _ := bcrypt.GenerateFromPassword([]byte("ClientDemo123!"), bcrypt.DefaultCost)
	clientID := uuid.NewString()
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at, terms_accepted_at
		) VALUES ($1, $2, $3, 'Claim Client', 'client', NOW(), NOW())`,
		clientID, clientEmail, string(clientHash))
	if err != nil {
		t.Fatal(err)
	}
	_ = st.EnsureUserProfiles(ctx, clientID)

	clientTok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", clientTok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim-invite %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["kind"] != "commercial" {
		t.Fatalf("kind=%v want commercial", data["kind"])
	}
	if data["status"] != "referred" {
		t.Fatalf("status=%v want referred", data["status"])
	}

	var referral string
	err = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&referral)
	if err != nil {
		t.Fatal(err)
	}
	if referral != commID {
		t.Fatalf("referral=%q want %q", referral, commID)
	}

	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.commercial_referrals WHERE client_user_id=$1`, clientID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.app_invite_codes WHERE user_id=$1`, commID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.users WHERE id IN ($1, $2)`, clientID, commID)
	})
}
