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
	addons := billing.AllAddons()
	if len(addons) != 4 {
		t.Fatalf("expected 4 addons, got %d", len(addons))
	}
	byCode := map[billing.AddonCode]int{}
	for _, a := range addons {
		byCode[a.Code] = a.AmountCents
	}
	if byCode[billing.AddonFamily] != 3900 || byCode[billing.AddonKennel] != 11900 ||
		byCode[billing.AddonCarePlus] != 1900 || byCode[billing.AddonHorse] != 3900 {
		t.Fatalf("unexpected addon prices: %#v", byCode)
	}
}
