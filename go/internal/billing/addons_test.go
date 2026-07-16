package billing_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestParseAddonCode(t *testing.T) {
	for _, code := range []string{"family", "care_plus", "horse"} {
		if _, err := billing.ParseAddonCode(code); err != nil {
			t.Fatalf("expected valid addon %q: %v", code, err)
		}
	}
	if _, err := billing.ParseAddonCode("unknown"); err == nil {
		t.Fatal("expected invalid addon")
	}
}

func TestAllAddonsPrices(t *testing.T) {
	addons := billing.AllAddons()
	if len(addons) != 3 {
		t.Fatalf("expected 3 addons, got %d", len(addons))
	}
	byCode := map[billing.AddonCode]int{}
	for _, a := range addons {
		byCode[a.Code] = a.AmountCents
	}
	if byCode[billing.AddonFamily] != 4000 || byCode[billing.AddonCarePlus] != 1500 || byCode[billing.AddonHorse] != 3000 {
		t.Fatalf("unexpected addon prices: %#v", byCode)
	}
}
