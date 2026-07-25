package billing_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestParseAddonCode(t *testing.T) {
	for _, code := range []string{"family", "kennel", "care_plus", "horse"} {
		if _, err := billing.ParseAddonCode(code); err != nil {
			t.Fatalf("expected valid addon %q: %v", code, err)
		}
	}
	if _, err := billing.ParseAddonCode("unknown"); err == nil {
		t.Fatal("expected invalid addon")
	}
}

func TestAllAddonsPrices(t *testing.T) {
	if n := len(billing.AllAddons()); n != 0 {
		t.Fatalf("sellable addon catalog must be empty, got %d", n)
	}
	// Legacy defs still resolve for grandfathered entitlements / webhooks.
	for _, code := range []billing.AddonCode{
		billing.AddonFamily, billing.AddonKennel, billing.AddonCarePlus, billing.AddonHorse,
	} {
		if _, err := billing.GetAddon(code); err != nil {
			t.Fatalf("GetAddon(%s) legacy: %v", code, err)
		}
	}
	family, _ := billing.GetAddon(billing.AddonFamily)
	kennel, _ := billing.GetAddon(billing.AddonKennel)
	care, _ := billing.GetAddon(billing.AddonCarePlus)
	horse, _ := billing.GetAddon(billing.AddonHorse)
	if family.AmountCents != 3900 || kennel.AmountCents != 11900 ||
		care.AmountCents != 1900 || horse.AmountCents != 3900 {
		t.Fatalf("unexpected legacy addon prices: family=%d kennel=%d care=%d horse=%d",
			family.AmountCents, kennel.AmountCents, care.AmountCents, horse.AmountCents)
	}
}
