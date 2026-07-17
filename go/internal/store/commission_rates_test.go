package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestApplyVetPlanFactor(t *testing.T) {
	if got := store.ApplyVetPlanFactor(1200, "triennial"); got != 1200 {
		t.Fatalf("triennial max = 1200, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(1200, "annual"); got != 804 {
		t.Fatalf("annual max = 804 (8%%), got %d", got)
	}
	if got := store.ApplyVetPlanFactor(1200, "quinquennial"); got != 804 {
		t.Fatalf("quinquennial max = 804, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(700, "triennial"); got != 700 {
		t.Fatalf("triennial entry = 700, got %d", got)
	}
	if got := store.ApplyVetPlanFactor(700, "annual"); got != 469 {
		t.Fatalf("annual entry = 469, got %d", got)
	}
}

func TestCommercialRateBpsForPlan(t *testing.T) {
	if got := store.CommercialRateBpsForPlan("annual"); got != 800 {
		t.Fatalf("annual = 800, got %d", got)
	}
	if got := store.CommercialRateBpsForPlan("triennial"); got != 1200 {
		t.Fatalf("triennial = 1200, got %d", got)
	}
	if got := store.CommercialRateBpsForPlan("quinquennial"); got != 800 {
		t.Fatalf("quinquennial = 800, got %d", got)
	}
	if got := store.CommercialRateBpsForAddon("family"); got != 1000 {
		t.Fatalf("addon = 1000, got %d", got)
	}
}

func TestDefaultVetCommissionTiers(t *testing.T) {
	tiers := store.DefaultVetCommissionTiers()
	if len(tiers) != 4 {
		t.Fatalf("want 4 tiers, got %d", len(tiers))
	}
	if tiers[0].RateBps != 700 || tiers[3].RateBps != 1200 {
		t.Fatalf("unexpected tier rates: %#v", tiers)
	}
}
