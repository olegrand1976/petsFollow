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
	if len(byCode) != 3 {
		t.Fatalf("want 3 sellable plans in commission catalog, got %d", len(byCode))
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
	if _, ok := byCode["quinquennial"]; ok {
		t.Fatal("quinquennial must not appear in sellable commission catalog")
	}
}

func TestAddonCatalogMatchesBillingDomain(t *testing.T) {
	rates := store.AddonPlanRates()
	if len(rates) != 0 {
		t.Fatalf("addon commission catalog must be empty, got %d", len(rates))
	}
	if n := len(billing.AllAddons()); n != 0 {
		t.Fatalf("AllAddons must be empty, got %d", n)
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
	if got := store.CommissionFromTTCCents(350, 800); got != 23 {
		// HTVA(350)=289; 8%=23
		t.Fatalf("8%% of HTVA(350)=23, got %d", got)
	}
}
