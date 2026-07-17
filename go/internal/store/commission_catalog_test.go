package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestSubscriptionCatalogMatchesBillingDomain(t *testing.T) {
	byCode := map[string]store.PlanRateInfo{}
	for _, r := range store.SubscriptionPlanRates() {
		byCode[r.Code] = r
	}
	for _, p := range billing.AllPlans() {
		got, ok := byCode[string(p.Code)]
		if !ok {
			t.Fatalf("missing plan %s in commission catalog", p.Code)
		}
		if got.TTCCents != p.AmountCents {
			t.Fatalf("%s TTC catalog=%d domain=%d", p.Code, got.TTCCents, p.AmountCents)
		}
		if got.Recommended != p.Recommended {
			t.Fatalf("%s recommended mismatch", p.Code)
		}
	}
}

func TestAddonCatalogMatchesBillingDomain(t *testing.T) {
	byCode := map[string]store.PlanRateInfo{}
	for _, r := range store.AddonPlanRates() {
		byCode[r.Code] = r
	}
	for _, a := range billing.AllAddons() {
		got, ok := byCode[string(a.Code)]
		if !ok {
			t.Fatalf("missing addon %s in commission catalog", a.Code)
		}
		if got.TTCCents != a.AmountCents {
			t.Fatalf("%s TTC catalog=%d domain=%d", a.Code, got.TTCCents, a.AmountCents)
		}
	}
}

func TestIndicativeTriennialCommissions(t *testing.T) {
	// 95 € TTC → HTVA 7851 ct; 12% = 942 ct
	ht := store.HTVACents(9500)
	if ht != 7851 {
		t.Fatalf("HTVA(9500)=7851, got %d", ht)
	}
	if got := store.CommissionFromTTCCents(9500, 1200); got != 942 {
		t.Fatalf("12%% of HTVA(9500)=942, got %d", got)
	}
	if got := store.CommissionFromTTCCents(3500, 800); got != 231 {
		// HTVA(3500)=2892; 8%=231
		t.Fatalf("8%% of HTVA(3500)=231, got %d", got)
	}
	if got := store.CommissionFromTTCCents(14500, 800); got != 958 {
		// HTVA(14500)=11983; 8%=958
		t.Fatalf("8%% of HTVA(14500)=958, got %d", got)
	}
}
