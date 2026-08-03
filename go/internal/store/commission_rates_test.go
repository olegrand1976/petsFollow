package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestApplyVetPlanFactor(t *testing.T) {
	if got := store.ApplyVetPlanFactor(800, "triennial"); got != 800 {
		t.Fatalf("triennial max = 800, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(1200, "triennial"); got != 800 {
		t.Fatalf("triennial over-cap clamps to 800, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(800, "annual"); got != 536 {
		t.Fatalf("annual max = 536 (5.36%%), got %d", got)
	}
	if got := store.ApplyVetPlanFactor(800, "monthly"); got != 536 {
		t.Fatalf("monthly max = 536 (5.36%%), got %d", got)
	}
	if got := store.ApplyVetPlanFactor(800, "quinquennial"); got != 536 {
		t.Fatalf("quinquennial legacy max = 536, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(750, "triennial"); got != 750 {
		t.Fatalf("triennial entry = 750, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(750, "annual"); got != 502 {
		t.Fatalf("annual entry = 502, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(750, "monthly"); got != 502 {
		t.Fatalf("monthly entry = 502, got %d", got)
	}
	// Below 5% after factor → floor at MinVetCommissionBps.
	if got := store.ApplyVetPlanFactor(700, "annual"); got != 500 {
		t.Fatalf("annual below-floor clamps to 500, got %d", got)
	}
}

func TestCommercialRateBpsForPlan(t *testing.T) {
	if got := store.CommercialRateBpsForPlan("monthly"); got != 800 {
		t.Fatalf("monthly = 800, got %d", got)
	}
	if got := store.CommercialRateBpsForPlan("annual"); got != 800 {
		t.Fatalf("annual = 800, got %d", got)
	}
	if got := store.CommercialRateBpsForPlan("triennial"); got != 1200 {
		t.Fatalf("triennial = 1200, got %d", got)
	}
	if got := store.CommercialRateBpsForPlan("quinquennial"); got != 800 {
		t.Fatalf("quinquennial legacy = 800, got %d", got)
	}
	if got := store.CommercialRateBpsForAddon("family"); got != 1000 {
		t.Fatalf("addon legacy = 1000, got %d", got)
	}
	if got := store.CommercialRateBpsForAddon("kennel"); got != 1000 {
		t.Fatalf("kennel commercial legacy = 1000, got %d", got)
	}
	if got := store.VetRateBpsForAddon("family"); got != 500 {
		t.Fatalf("family vet = 500, got %d", got)
	}
	if got := store.VetRateBpsForAddon("kennel"); got != 500 {
		t.Fatalf("kennel vet = 500, got %d", got)
	}
	if got := store.VetRateBpsForAddon("horse"); got != 0 {
		t.Fatalf("horse vet = 0, got %d", got)
	}
}

func TestDefaultVetCommissionTiers(t *testing.T) {
	tiers := store.DefaultVetCommissionTiers()
	if len(tiers) != 4 {
		t.Fatalf("want 4 tiers, got %d", len(tiers))
	}
	want := []int{750, 767, 783, 800}
	for i, w := range want {
		if tiers[i].RateBps != w {
			t.Fatalf("tier[%d] rate = %d, want %d", i, tiers[i].RateBps, w)
		}
		if tiers[i].RateBps < store.MinVetBaseTierBps || tiers[i].RateBps > store.MaxVetCommissionBps {
			t.Fatalf("tier[%d] rate %d outside [%d, %d]", i, tiers[i].RateBps, store.MinVetBaseTierBps, store.MaxVetCommissionBps)
		}
	}
}
