package journey

import (
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestNeedsFirstMeasureSkip(t *testing.T) {
	ok, reason := needsFirstMeasure(store.JourneyClientSegment{ValidatedHRCount: 1}, time.Now())
	if ok || reason != "has_validated_hr" {
		t.Fatalf("got ok=%v reason=%q", ok, reason)
	}
}

func TestNeedsKennelUpsell(t *testing.T) {
	ok, reason := needsKennelUpsell(store.JourneyClientSegment{PetCount: 5}, time.Now())
	if ok || reason != "lt_6_pets" {
		t.Fatalf("lt6: ok=%v reason=%q", ok, reason)
	}
	ok, reason = needsKennelUpsell(store.JourneyClientSegment{
		PetCount: 6, ActiveAddons: map[string]bool{"kennel": true},
	}, time.Now())
	if ok || reason != "has_kennel" {
		t.Fatalf("has kennel: ok=%v reason=%q", ok, reason)
	}
	ok, _ = needsKennelUpsell(store.JourneyClientSegment{PetCount: 6, ActiveAddons: map[string]bool{}}, time.Now())
	if !ok {
		t.Fatal("expected kennel upsell")
	}
}

func TestNeedsHorseUpsell(t *testing.T) {
	ok, reason := needsHorseUpsell(store.JourneyClientSegment{HorseCount: 0}, time.Now())
	if ok || reason != "no_horse" {
		t.Fatalf("got ok=%v reason=%q", ok, reason)
	}
	ok, _ = needsHorseUpsell(store.JourneyClientSegment{HorseCount: 1, ActiveAddons: map[string]bool{}}, time.Now())
	if !ok {
		t.Fatal("expected horse upsell")
	}
}

func TestNeedsInactiveHRBeforeD14(t *testing.T) {
	ok, reason := needsInactiveHR(store.JourneyClientSegment{JourneyDays: 10}, time.Now())
	if ok || reason != "before_d14" {
		t.Fatalf("got ok=%v reason=%q", ok, reason)
	}
	d := 30
	ok, _ = needsInactiveHR(store.JourneyClientSegment{JourneyDays: 20, DaysSinceLastHR: &d}, time.Now())
	if !ok {
		t.Fatal("expected inactive after 21d")
	}
}

func TestPrefEnabled(t *testing.T) {
	seg := store.JourneyClientSegment{DiscoveryPref: true, BillingPref: false}
	if !PrefEnabled(seg, PrefDiscovery) {
		t.Fatal("discovery should be on")
	}
	if PrefEnabled(seg, PrefBilling) {
		t.Fatal("billing should be off")
	}
}

func TestAllStepKeysUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, k := range AllStepKeys() {
		if seen[k] {
			t.Fatalf("duplicate step %s", k)
		}
		seen[k] = true
	}
	if len(seen) < 20 {
		t.Fatalf("expected >=20 steps, got %d", len(seen))
	}
}

func TestAppendUTM(t *testing.T) {
	got := appendUTM("https://example.com/app", "d2_first_measure")
	if !strings.Contains(got, "utm_content=d2_first_measure") {
		t.Fatalf("unexpected utm url: %s", got)
	}
}
