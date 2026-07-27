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

func TestNearbyCommercialsByGPSAndPostal(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("CommercialDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	nearID := uuid.NewString()
	farID := uuid.NewString()
	emailNear := fmt.Sprintf("smoke-near+%d@petsfollow.test", time.Now().UnixNano())
	emailFar := fmt.Sprintf("smoke-far+%d@petsfollow.test", time.Now().UnixNano())

	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES
			($1, $2, $3, 'Near Commercial', 'commercial', NOW(), 50.8503, 4.3517, 'Bruxelles', '1000'),
			($4, $5, $3, 'Far Commercial', 'commercial', NOW(), 43.2965, 5.3698, 'Marseille', '13001')`,
		nearID, emailNear, string(hash), farID, emailFar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.users WHERE id IN ($1, $2)`, nearID, farID)
	})

	code, env := doJSON(t, api.handler, http.MethodGet,
		"/api/v1/commercials/nearby?lat=50.85&lng=4.35&limit=10", nil)
	if code != http.StatusOK {
		t.Fatalf("nearby gps status %d: %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	foundNear, foundFar := false, false
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		id, _ := m["userId"].(string)
		if id == nearID {
			foundNear = true
		}
		if id == farID {
			foundFar = true
		}
	}
	if !foundNear {
		t.Fatalf("expected near commercial in GPS results: %#v", rows)
	}
	if foundFar {
		t.Fatalf("far commercial should be outside default radius: %#v", rows)
	}

	code, env = doJSON(t, api.handler, http.MethodGet,
		"/api/v1/commercials/nearby?postalCode=1000&limit=10", nil)
	if code != http.StatusOK {
		t.Fatalf("nearby postal status %d: %#v", code, env)
	}
	rows, _ = env["data"].([]any)
	foundNear = false
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		if m["userId"] == nearID {
			foundNear = true
		}
	}
	if !foundNear {
		t.Fatalf("expected near commercial for postal 1000: %#v", rows)
	}
}

func TestRegisterVetWithInviteCode(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("CommercialDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	commID := uuid.NewString()
	commEmail := fmt.Sprintf("smoke-assign-comm+%d@petsfollow.test", time.Now().UnixNano())
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES ($1, $2, $3, 'Assign Comm', 'commercial', NOW(), 50.85, 4.35, 'Bruxelles', '1000')`,
		commID, commEmail, string(hash))
	if err != nil {
		t.Fatal(err)
	}
	inv, err := store.New(api.pool).EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}

	vetEmail := fmt.Sprintf("smoke-nearby-vet+%d@petsfollow.test", time.Now().UnixNano())
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":        vetEmail,
		"password":     "VetDemo123!",
		"fullName":     "Nearby Vet",
		"practiceName": "Cabinet Nearby",
		"consent":      true,
		"inviteCode":   inv.Code,
		// nearby pick alone must NOT assign
		"assignedCommercialId": uuid.NewString(),
	})
	if code != http.StatusCreated {
		t.Fatalf("register status %d: %#v", code, env)
	}

	var assigned string
	err = api.pool.QueryRow(ctx, `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`,
		vetEmail).Scan(&assigned)
	if err != nil {
		t.Fatal(err)
	}
	if assigned != commID {
		t.Fatalf("assigned_commercial_id=%q want %q", assigned, commID)
	}

	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.app_invite_codes WHERE user_id=$1`, commID)
		_, _ = api.pool.Exec(ctx, `
			DELETE FROM identity.email_verification_tokens WHERE user_id IN (
				SELECT id FROM identity.users WHERE email=$1 OR id=$2::uuid)`, vetEmail, commID)
		_, _ = api.pool.Exec(ctx, `
			DELETE FROM messaging.vet_availability WHERE vet_user_id IN (
				SELECT id FROM identity.users WHERE email=$1)`, vetEmail)
		_, _ = api.pool.Exec(ctx, `
			DELETE FROM notifications.notification_preferences WHERE vet_user_id IN (
				SELECT id FROM identity.users WHERE email=$1)`, vetEmail)
		var practiceID string
		_ = api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE email=$1`, vetEmail).Scan(&practiceID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.users WHERE email=$1 OR id=$2::uuid`, vetEmail, commID)
		if practiceID != "" {
			_, _ = api.pool.Exec(ctx, `DELETE FROM practice.practices WHERE id=$1::uuid`, practiceID)
		}
	})
}

func TestRegisterClientCommercialReferralFallback(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	hash, err := bcrypt.GenerateFromPassword([]byte("CommercialDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	commID := uuid.NewString()
	commEmail := fmt.Sprintf("smoke-ref-comm+%d@petsfollow.test", time.Now().UnixNano())
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, email_verified_at,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES ($1, $2, $3, 'Referral Comm', 'commercial', NOW(), 50.85, 4.35, 'Bruxelles', '1000')`,
		commID, commEmail, string(hash))
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := fmt.Sprintf("smoke-ref-client+%d@petsfollow.test", time.Now().UnixNano())
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email":            clientEmail,
		"password":         "ClientDemo123!",
		"fullName":         "Referral Client",
		"consent":          true,
		"commercialUserId": commID,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client status %d: %#v", code, env)
	}

	var clientID, referral string
	err = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	if err != nil {
		t.Fatal(err)
	}
	err = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&referral)
	if err != nil {
		t.Fatal(err)
	}
	if referral != commID {
		t.Fatalf("referral=%q want %q", referral, commID)
	}

	practiceID := uuid.NewString()
	vetID := uuid.NewString()
	vetHash, _ := bcrypt.GenerateFromPassword([]byte("VetDemo123!"), bcrypt.DefaultCost)
	vetEmail := fmt.Sprintf("smoke-orphan-vet+%d@petsfollow.test", time.Now().UnixNano())
	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.practices (id, name, contact_email) VALUES ($1, 'Orphan Practice', $2)`,
		practiceID, "orphan@petsfollow.test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.pool.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, 'Orphan Vet', 'vet', $4, NOW())`,
		vetID, vetEmail, string(vetHash), practiceID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceID, clientID, vetID)
	if err != nil {
		t.Fatal(err)
	}

	_, commercialUserID, err := st.ResolveVetCommercial(ctx, clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if commercialUserID != commID {
		t.Fatalf("fallback commercial=%q want %q", commercialUserID, commID)
	}

	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.practice_clients WHERE client_user_id=$1`, clientID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.commercial_referrals WHERE client_user_id=$1`, clientID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.email_verification_tokens WHERE user_id=$1`, clientID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.users WHERE id IN ($1, $2, $3)`, clientID, vetID, commID)
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.practices WHERE id=$1`, practiceID)
	})
}
