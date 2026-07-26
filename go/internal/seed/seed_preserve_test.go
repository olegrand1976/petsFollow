package seed_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"golang.org/x/crypto/bcrypt"
)

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
				if os.Getenv(parts[0]) == "" {
					_ = os.Setenv(parts[0], parts[1])
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

// TestSeedPreservesProtectedRoles ensures admin/commercial/commercial_manager
// accounts (including non-demo staging users) are never wiped by seed.Run.
func TestSeedPreservesProtectedRoles(t *testing.T) {
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

	keepEmail := "keep+seed-preserve-" + uuid.NewString() + "@example.test"
	keepPass := "KeepPass123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(keepPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	keepID := uuid.NewString()
	if _, err := pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password
		) VALUES ($1, $2, $3, 'Keep Manager', 'commercial_manager', NULL, NOW(), true)`,
		keepID, keepEmail, string(hash)); err != nil {
		t.Fatalf("insert protected manager: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, keepID)
	})

	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("seed.Run: %v", err)
	}

	var (
		stillThere bool
		role       string
		hashAfter  string
		mustChange bool
	)
	err = pool.QueryRow(ctx, `
		SELECT true, role, password_hash, must_change_password
		FROM identity.users WHERE id=$1`, keepID).Scan(&stillThere, &role, &hashAfter, &mustChange)
	if err != nil {
		t.Fatalf("protected manager missing after seed: %v", err)
	}
	if role != "commercial_manager" {
		t.Fatalf("expected commercial_manager, got %q", role)
	}
	if !mustChange {
		t.Fatal("non-demo protected account must_change_password must stay untouched (true)")
	}
	if bcrypt.CompareHashAndPassword([]byte(hashAfter), []byte(keepPass)) != nil {
		t.Fatal("non-demo protected password_hash must stay untouched")
	}

	var adminEmail string
	err = pool.QueryRow(ctx, `
		SELECT email FROM identity.users
		WHERE email='admin.demo@petsfollow.test' AND role='admin'`).Scan(&adminEmail)
	if err != nil {
		t.Fatalf("admin.demo missing after seed: %v", err)
	}

	var seedMgrID, mgrEmail, commEmail, comm2Email string
	err = pool.QueryRow(ctx, `
		SELECT id::text, email FROM identity.users
		WHERE email='commercial.manager@petsfollow.test' AND role='commercial_manager'`).Scan(&seedMgrID, &mgrEmail)
	if err != nil {
		t.Fatalf("commercial.manager missing after seed: %v", err)
	}
	err = pool.QueryRow(ctx, `
		SELECT email FROM identity.users
		WHERE email='commercial.demo@petsfollow.test' AND role='commercial'`).Scan(&commEmail)
	if err != nil {
		t.Fatalf("commercial.demo missing after seed: %v", err)
	}
	err = pool.QueryRow(ctx, `
		SELECT email FROM identity.users
		WHERE email='commercial.demo2@petsfollow.test' AND role='commercial'`).Scan(&comm2Email)
	if err != nil {
		t.Fatalf("commercial.demo2 missing after seed: %v", err)
	}

	var demoTeamCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users
		WHERE email IN ('commercial.demo@petsfollow.test', 'commercial.demo2@petsfollow.test')
		  AND manager_user_id = $1::uuid`, seedMgrID).Scan(&demoTeamCount); err != nil {
		t.Fatal(err)
	}
	if demoTeamCount != 2 {
		t.Fatalf("expected both demo commercials linked to seed manager, got count=%d", demoTeamCount)
	}

	var vetCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM identity.users WHERE email='vet.demo@petsfollow.test'`).Scan(&vetCount); err != nil {
		t.Fatal(err)
	}
	if vetCount != 1 {
		t.Fatalf("expected vet.demo recreated, got count=%d", vetCount)
	}

	// Non-demo commercials must keep an existing manager across seed (COALESCE / untouched).
	// Demo emails are force-linked to commercial.manager — use a separate account here.
	altMgrID := uuid.NewString()
	altMgrEmail := "alt-mgr+" + uuid.NewString() + "@example.test"
	altHash, err := bcrypt.GenerateFromPassword([]byte("AltMgr123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password
		) VALUES ($1, $2, $3, 'Alt Manager', 'commercial_manager', NULL, NOW(), false)`,
		altMgrID, altMgrEmail, string(altHash)); err != nil {
		t.Fatalf("insert alt manager: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, altMgrID)
	})
	preserveID := uuid.NewString()
	preserveEmail := "commercial.preserve@petsfollow.test"
	if _, err := pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at,
			manager_user_id, must_change_password
		) VALUES ($1, $2, $3, 'Preserve Comm', 'commercial', NULL, NOW(), $4::uuid, true)
		ON CONFLICT (email) DO UPDATE SET
			manager_user_id = EXCLUDED.manager_user_id,
			password_hash = EXCLUDED.password_hash,
			must_change_password = true`,
		preserveID, preserveEmail, string(hash), altMgrID); err != nil {
		t.Fatalf("insert preserve commercial: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE email=$1`, preserveEmail)
	})
	// Detach demos to an alt manager — seed must re-attach them to Bérénice.
	if _, err := pool.Exec(ctx, `
		UPDATE identity.users SET manager_user_id=$1
		WHERE email IN ('commercial.demo@petsfollow.test', 'commercial.demo2@petsfollow.test')`,
		altMgrID); err != nil {
		t.Fatalf("detach demos to alt manager: %v", err)
	}
	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("second seed.Run: %v", err)
	}
	var preserveMgrAfter string
	err = pool.QueryRow(ctx, `
		SELECT manager_user_id::text FROM identity.users
		WHERE email=$1`, preserveEmail).Scan(&preserveMgrAfter)
	if err != nil {
		t.Fatalf("commercial.preserve after second seed: %v", err)
	}
	if preserveMgrAfter != altMgrID {
		t.Fatalf("commercial.preserve manager_user_id overwritten: got %q want %q", preserveMgrAfter, altMgrID)
	}
	if err := pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users
		WHERE email='commercial.manager@petsfollow.test'`).Scan(&seedMgrID); err != nil {
		t.Fatalf("seed manager after second seed: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users
		WHERE email IN ('commercial.demo@petsfollow.test', 'commercial.demo2@petsfollow.test')
		  AND manager_user_id = $1::uuid`, seedMgrID).Scan(&demoTeamCount); err != nil {
		t.Fatal(err)
	}
	if demoTeamCount != 2 {
		t.Fatalf("expected demos force-relinked to seed manager after re-seed, got count=%d", demoTeamCount)
	}

	smokeEmail := fmt.Sprintf("smoke-comm+%d@petsfollow.test", time.Now().UnixNano())
	if _, err := pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password
		) VALUES ($1, $2, $3, 'Smoke Comm', 'commercial', NULL, NOW(), true)`,
		uuid.NewString(), smokeEmail, string(hash)); err != nil {
		t.Fatalf("insert smoke commercial: %v", err)
	}
	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("third seed.Run: %v", err)
	}
	var smokeLeft int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM identity.users WHERE email=$1`, smokeEmail).Scan(&smokeLeft); err != nil {
		t.Fatal(err)
	}
	if smokeLeft != 0 {
		t.Fatalf("expected smoke commercial purged by seed, still present")
	}
}
