package store_test

import (
	"errors"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestFamilyLimits(t *testing.T) {
	if store.FamilyMinPets != 2 || store.FamilyMaxPets != 3 {
		t.Fatalf("unexpected family limits: min=%d max=%d", store.FamilyMinPets, store.FamilyMaxPets)
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
		{4, store.ErrFamilyPetLimit},
	}
	for _, tc := range cases {
		err := store.CheckFamilyPurchasePetCount(tc.n)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("n=%d: got %v want %v", tc.n, err, tc.wantErr)
		}
	}
}

func TestCheckFamilyCanAddPetCount(t *testing.T) {
	if err := store.CheckFamilyCanAddPetCount(3, false); err != nil {
		t.Fatalf("no family: 3 pets should allow add: %v", err)
	}
	if err := store.CheckFamilyCanAddPetCount(4, false); err != nil {
		t.Fatalf("no family: 4 pets should allow add: %v", err)
	}
	if err := store.CheckFamilyCanAddPetCount(2, true); err != nil {
		t.Fatalf("family: 2 pets should allow add: %v", err)
	}
	if !errors.Is(store.CheckFamilyCanAddPetCount(3, true), store.ErrFamilyPetLimit) {
		t.Fatal("family: 3 pets must block 4th")
	}
}
