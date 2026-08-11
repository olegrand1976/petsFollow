package store_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func pharmacyStockStore(t *testing.T) (*store.Store, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("short")
	}
	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("DEV_SEED_ENABLED", "true")
	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("db: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := seed.Run(ctx, pool); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return store.New(pool), ctx
}

func TestPharmacyFEFOAndExpiry(t *testing.T) {
	st, ctx := pharmacyStockStore(t)

	u, err := st.GetUserByEmail(ctx, "vet.demo@petsfollow.test")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	practiceID, userID := u.PracticeID, u.ID

	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2777001", Name: "FEFO Demo Med", IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	dep, err := st.EnsureDefaultDeposit(ctx, practiceID)
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	settings, err := st.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	today := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

	_, _, err = st.ReceiveMedicationBatch(ctx, store.ReceiptInput{
		PracticeID: practiceID, DepositID: dep.ID, MedicationID: medID,
		LotNumber: "EXP1", ExpiresOn: today.AddDate(0, 0, -1), Qty: 5, CreatedBy: userID,
	}, settings, today)
	if !errors.Is(err, pharmacy.ErrInvalidExpiryOnReceipt) {
		t.Fatalf("expected invalid_expiry, got %v", err)
	}

	b1, soft, err := st.ReceiveMedicationBatch(ctx, store.ReceiptInput{
		PracticeID: practiceID, DepositID: dep.ID, MedicationID: medID,
		LotNumber: "A", ExpiresOn: today.AddDate(0, 0, 40), Qty: 3, CreatedBy: userID,
	}, settings, today)
	if err != nil {
		t.Fatalf("receipt A: %v", err)
	}
	if soft {
		t.Fatalf("40d should not soft-warn with receiptWarn=30")
	}
	b2, soft2, err := st.ReceiveMedicationBatch(ctx, store.ReceiptInput{
		PracticeID: practiceID, DepositID: dep.ID, MedicationID: medID,
		LotNumber: "B", ExpiresOn: today.AddDate(0, 0, 20), Qty: 4, CreatedBy: userID,
	}, settings, today)
	if err != nil {
		t.Fatalf("receipt B: %v", err)
	}
	if !soft2 {
		t.Fatalf("20d should soft-warn")
	}

	lines, err := st.AllocateFEFO(ctx, practiceID, medID, dep.ID, 5, userID, today)
	if err != nil {
		t.Fatalf("fefo: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %#v", lines)
	}
	if lines[0].BatchID != b2.ID || lines[0].Qty != 4 {
		t.Fatalf("first should be lot B qty 4: %#v", lines[0])
	}
	if lines[1].BatchID != b1.ID || lines[1].Qty != 1 {
		t.Fatalf("second should be lot A qty 1: %#v", lines[1])
	}

	a, err := st.GetMedicationBatch(ctx, practiceID, b1.ID)
	if err != nil || a.QtyOnHand != 2 {
		t.Fatalf("A remaining: %#v %v", a, err)
	}

	settings.AllowExpiredReceipt = true
	bExp, _, err := st.ReceiveMedicationBatch(ctx, store.ReceiptInput{
		PracticeID: practiceID, DepositID: dep.ID, MedicationID: medID,
		LotNumber: "Z", ExpiresOn: today.AddDate(0, 0, -2), Qty: 10, CreatedBy: userID,
	}, settings, today)
	if err != nil {
		t.Fatalf("expired receipt allowed: %v", err)
	}
	n, _, err := st.AutoQuarantineExpiredBatches(ctx, practiceID, today)
	if err != nil || n < 1 {
		t.Fatalf("auto quarantine: n=%d err=%v", n, err)
	}
	q, err := st.GetMedicationBatch(ctx, practiceID, bExp.ID)
	if err != nil || q.Status != "quarantine" {
		t.Fatalf("expected quarantine %#v %v", q, err)
	}

	_, err = st.WasteBatch(ctx, practiceID, bExp.ID, userID, "supplier_return", nil)
	if err != nil {
		t.Fatalf("waste: %v", err)
	}
}

func TestCreateMedicationDepositDuplicateCode(t *testing.T) {
	st, ctx := pharmacyStockStore(t)
	u, err := st.GetUserByEmail(ctx, "vet.demo@petsfollow.test")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	code := "DUP" + time.Now().Format("150405")
	if _, err := st.CreateMedicationDeposit(ctx, u.PracticeID, "Depot A", code, false); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = st.CreateMedicationDeposit(ctx, u.PracticeID, "Depot B", code, false)
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}
