package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
)

// TestSeedMass densifies after seed.Run with reduced volumes (CI-friendly).
func TestSeedMass(t *testing.T) {
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
		t.Fatalf("seed.Run: %v", err)
	}

	opts := seed.MassOptions{
		PracticeCount:      2,
		ClientsPerPractice: 5,
		OrphanClients:      3,
		ExtraDemoClients:   3,
		CareProCount:       2,
	}
	if err := seed.RunMassWithOptions(ctx, pool, opts); err != nil {
		t.Fatalf("RunMassWithOptions: %v", err)
	}

	var vetCount, clientCount, careCount, petCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email LIKE 'mass.vet.%@petsfollow.test'`).Scan(&vetCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email LIKE 'mass.client.%@petsfollow.test'`).Scan(&clientCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email LIKE 'mass.care.%@petsfollow.test'`).Scan(&careCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pets.pets p
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE u.email LIKE 'mass.client.%@petsfollow.test'`).Scan(&petCount); err != nil {
		t.Fatal(err)
	}

	if vetCount != opts.PracticeCount {
		t.Fatalf("expected %d mass vets, got %d", opts.PracticeCount, vetCount)
	}
	wantClients := opts.PracticeCount*opts.ClientsPerPractice + opts.OrphanClients + opts.ExtraDemoClients
	if clientCount != wantClients {
		t.Fatalf("expected %d mass clients, got %d", wantClients, clientCount)
	}
	if careCount != opts.CareProCount {
		t.Fatalf("expected %d mass care_pros, got %d", opts.CareProCount, careCount)
	}
	if petCount < 5 {
		t.Fatalf("expected several mass pets, got %d", petCount)
	}

	// Commerciaux démo intactes + assignations mass.
	var camille, alex int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email='commercial.demo@petsfollow.test'`).Scan(&camille); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email='commercial.demo2@petsfollow.test'`).Scan(&alex); err != nil {
		t.Fatal(err)
	}
	if camille != 1 || alex != 1 {
		t.Fatalf("commercials must stay: camille=%d alex=%d", camille, alex)
	}

	var assigned int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users
		WHERE email LIKE 'mass.vet.%@petsfollow.test' AND assigned_commercial_id IS NOT NULL`).Scan(&assigned); err != nil {
		t.Fatal(err)
	}
	if assigned != opts.PracticeCount {
		t.Fatalf("expected all mass vets assigned to a commercial, got %d", assigned)
	}

	var ledger int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM billing.commission_ledger`).Scan(&ledger); err != nil {
		t.Fatal(err)
	}
	if ledger == 0 {
		t.Fatal("expected commission ledger rows after AccrueAll*")
	}

	var commercialLedger int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM billing.commercial_commission_ledger`).Scan(&commercialLedger); err != nil {
		t.Fatal(err)
	}
	if commercialLedger == 0 {
		t.Fatal("expected commercial commission ledger rows")
	}

	// Idempotence : second run no-op.
	if err := seed.RunMassWithOptions(ctx, pool, opts); err != nil {
		t.Fatalf("second RunMass: %v", err)
	}
	var vetAfter int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email LIKE 'mass.vet.%@petsfollow.test'`).Scan(&vetAfter); err != nil {
		t.Fatal(err)
	}
	if vetAfter != vetCount {
		t.Fatalf("idempotent re-run mutated vets: before=%d after=%d", vetCount, vetAfter)
	}
}
