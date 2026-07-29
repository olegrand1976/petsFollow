package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"golang.org/x/crypto/bcrypt"
)

func TestSeedPreservesSupportTickets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadDotEnv()
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("initial seed.Run: %v", err)
	}

	ticketID := uuid.NewString()
	replyID := uuid.NewString()
	var adminID string
	if err := pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users WHERE email='admin.demo@petsfollow.test'`).Scan(&adminID); err != nil {
		t.Fatalf("admin id: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO ops.support_tickets (
			id, created_by, source, subject, message, status
		) VALUES ($1, $2::uuid, 'nuxt_pro', 'seed-preserve-ticket', 'please keep me', 'open')`,
		ticketID, adminID); err != nil {
		t.Fatalf("insert ticket: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO ops.support_ticket_replies (id, ticket_id, author_id, body)
		VALUES ($1, $2::uuid, $3::uuid, 'admin reply keep')`,
		replyID, ticketID, adminID); err != nil {
		t.Fatalf("insert reply: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM ops.support_ticket_replies WHERE id=$1`, replyID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM ops.support_tickets WHERE id=$1`, ticketID)
	})

	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("seed.Run after ticket: %v", err)
	}

	var subject, replyBody string
	if err := pool.QueryRow(ctx, `
		SELECT subject FROM ops.support_tickets WHERE id=$1`, ticketID).Scan(&subject); err != nil {
		t.Fatalf("ticket missing after seed: %v", err)
	}
	if subject != "seed-preserve-ticket" {
		t.Fatalf("ticket subject=%q", subject)
	}
	if err := pool.QueryRow(ctx, `
		SELECT body FROM ops.support_ticket_replies WHERE id=$1`, replyID).Scan(&replyBody); err != nil {
		t.Fatalf("reply missing after seed: %v", err)
	}
	if replyBody != "admin reply keep" {
		t.Fatalf("reply body=%q", replyBody)
	}
}

func TestSeedPreservesClientGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadDotEnv()
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("initial seed.Run: %v", err)
	}

	keepEmail := "preserve-client+" + uuid.NewString() + "@example.test"
	restore := seed.SetPreserveClientEmailsForTest([]string{keepEmail})
	t.Cleanup(restore)

	var practiceID, vetID string
	if err := pool.QueryRow(ctx, `
		SELECT p.id::text, u.id::text
		FROM practice.practices p
		JOIN identity.users u ON u.practice_id = p.id AND u.role = 'vet'
		WHERE p.name = 'Cabinet VetPlus Demo' AND u.email = 'vet.demo@petsfollow.test'
		LIMIT 1`).Scan(&practiceID, &vetID); err != nil {
		t.Fatalf("demo practice: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("PreserveClient123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	clientID := uuid.NewString()
	petID := uuid.NewString()
	threadID := uuid.NewString()
	msgID := uuid.NewString()
	pcID := uuid.NewString()
	entID := uuid.NewString()

	if _, err := pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at, terms_accepted_at
		) VALUES ($1, $2, $3, 'Preserve Client', 'client', $4::uuid, NOW(), NOW())`,
		clientID, keepEmail, string(hash), practiceID); err != nil {
		t.Fatalf("insert client: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, clientID)
	})

	if _, err := pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid)`,
		pcID, practiceID, clientID, vetID); err != nil {
		t.Fatalf("practice_clients: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, weight_kg, payment_status)
		VALUES ($1, $2::uuid, $3::uuid, 'PreservePet', 'dog', 'Beagle', 12.5, 'active')`,
		petID, practiceID, clientID); err != nil {
		t.Fatalf("pet: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid, $5::uuid)`,
		threadID, practiceID, clientID, vetID, petID); err != nil {
		t.Fatalf("thread: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO messaging.messages (id, thread_id, sender_user_id, body)
		VALUES ($1, $2::uuid, $3::uuid, 'keep this message')`,
		msgID, threadID, clientID); err != nil {
		t.Fatalf("message: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO billing.pet_entitlements (
			id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency
		) VALUES ($1, $2::uuid, $3::uuid, 'annual', 'subscription', 'active', 3500, 'eur')`,
		entID, petID, clientID); err != nil {
		t.Fatalf("entitlement: %v", err)
	}

	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("seed.Run with preserve client: %v", err)
	}

	var stillEmail, petName, msgBody, plan string
	var newPracticeID, newVetID string
	if err := pool.QueryRow(ctx, `
		SELECT email, COALESCE(practice_id::text,'') FROM identity.users WHERE id=$1`, clientID).
		Scan(&stillEmail, &newPracticeID); err != nil {
		t.Fatalf("client missing after seed: %v", err)
	}
	if stillEmail != keepEmail {
		t.Fatalf("email=%q", stillEmail)
	}
	if err := pool.QueryRow(ctx, `
		SELECT name, COALESCE(practice_id::text,'') FROM pets.pets WHERE id=$1`, petID).
		Scan(&petName, &newPracticeID); err != nil {
		t.Fatalf("pet missing after seed: %v", err)
	}
	if petName != "PreservePet" {
		t.Fatalf("pet name=%q", petName)
	}
	var practiceName string
	if err := pool.QueryRow(ctx, `SELECT name FROM practice.practices WHERE id=$1::uuid`, newPracticeID).
		Scan(&practiceName); err != nil {
		t.Fatalf("remapped practice: %v", err)
	}
	if practiceName != "Cabinet VetPlus Demo" {
		t.Fatalf("expected remapped VetPlus, got %q", practiceName)
	}
	if err := pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM messaging.threads WHERE id=$1`, threadID).Scan(&newVetID); err != nil {
		t.Fatalf("thread missing: %v", err)
	}
	var vetEmail string
	if err := pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1::uuid`, newVetID).Scan(&vetEmail); err != nil {
		t.Fatalf("thread vet: %v", err)
	}
	if vetEmail != "vet.demo@petsfollow.test" {
		t.Fatalf("expected remapped vet.demo, got %q", vetEmail)
	}
	if err := pool.QueryRow(ctx, `SELECT body FROM messaging.messages WHERE id=$1`, msgID).Scan(&msgBody); err != nil {
		t.Fatalf("message missing: %v", err)
	}
	if msgBody != "keep this message" {
		t.Fatalf("message=%q", msgBody)
	}
	if err := pool.QueryRow(ctx, `
		SELECT plan_code FROM billing.pet_entitlements WHERE id=$1`, entID).Scan(&plan); err != nil {
		t.Fatalf("entitlement missing: %v", err)
	}
	if plan != "annual" {
		t.Fatalf("plan=%q", plan)
	}
	var pcCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM practice.practice_clients
		WHERE client_user_id=$1::uuid AND practice_id=$2::uuid`, clientID, newPracticeID).Scan(&pcCount); err != nil {
		t.Fatalf("practice_clients: %v", err)
	}
	if pcCount != 1 {
		t.Fatalf("practice_clients count=%d", pcCount)
	}
}
