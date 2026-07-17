package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestFamilyLimits(t *testing.T) {
	if store.FamilyMinPets != 2 || store.FamilyMaxPets != 3 {
		t.Fatalf("unexpected family limits: min=%d max=%d", store.FamilyMinPets, store.FamilyMaxPets)
	}
}
