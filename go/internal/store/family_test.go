package store_test

import (
	"errors"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestHouseholdLimits(t *testing.T) {
	if store.FamilyMinPets != 2 || store.KennelMinPets != 6 {
		t.Fatalf("unexpected limits: familyMin=%d kennelMin=%d", store.FamilyMinPets, store.KennelMinPets)
	}
}

func TestCheckFamilyPurchasePetCount(t *testing.T) {
	cases := []struct {
		n       int
		wantErr error
	}{
		{0, store.ErrFamilyRequiresTwoPets},
		{1, store.ErrFamilyRequiresTwoPets},
		{2, nil},
		{3, nil},
		{8, nil},
	}
	for _, tc := range cases {
		err := store.CheckFamilyPurchasePetCount(tc.n)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("n=%d: got %v want %v", tc.n, err, tc.wantErr)
		}
	}
}

func TestCheckKennelPurchasePetCount(t *testing.T) {
	if !errors.Is(store.CheckKennelPurchasePetCount(5), store.ErrKennelRequiresSixPets) {
		t.Fatal("5 pets must reject kennel")
	}
	if err := store.CheckKennelPurchasePetCount(6); err != nil {
		t.Fatalf("6 pets should allow kennel: %v", err)
	}
}

func TestApplyDiscountCents(t *testing.T) {
	if got := store.ApplyDiscountCents(3500, 1000); got != 3150 {
		t.Fatalf("−10%% of 3500: got %d", got)
	}
	if got := store.ApplyDiscountCents(9500, 1500); got != 8075 {
		t.Fatalf("−15%% of 9500: got %d", got)
	}
	if got := store.HouseholdDiscountBps(true, false); got != store.FamilyPetDiscountBps {
		t.Fatalf("family discount bps: %d", got)
	}
	if got := store.HouseholdDiscountBps(true, true); got != store.KennelPetDiscountBps {
		t.Fatalf("kennel wins over family: %d", got)
	}
}

// Commercial Annual totals: 1st pet full price, later pets discounted, + household addon.
// Mirrors ResolvePetCheckoutAmount (others ≥ 1) + catalogue amounts in billing/domain.go.
func TestAnnualHouseholdCheckoutTotals(t *testing.T) {
	const (
		annualCents = 3500
		familyCents = 3900
		kennelCents = 11900
	)
	familyPet := store.ApplyDiscountCents(annualCents, store.FamilyPetDiscountBps) // 3150
	kennelPet := store.ApplyDiscountCents(annualCents, store.KennelPetDiscountBps) // 2975

	cases := []struct {
		name string
		got  int
		want int
	}{
		{"without Family — 2 pets", annualCents + annualCents, 7000},
		{"with Family — 2 pets", annualCents + familyPet + familyCents, 10550},
		{"with Family — 4 pets", annualCents + 3*familyPet + familyCents, 16850},
		{"with Kennel — 6 pets", annualCents + 5*kennelPet + kennelCents, 30275},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s: got %d want %d", tc.name, tc.got, tc.want)
		}
	}
}
