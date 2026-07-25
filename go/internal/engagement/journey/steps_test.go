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
	// 14 timed + 3 event; addon upsells (d45/d60/d75) removed.
	if len(seen) < 17 {
		t.Fatalf("expected >=17 steps, got %d", len(seen))
	}
	for _, banned := range []string{"d45_care_plus", "d60_horse", "d75_kennel"} {
		if seen[banned] {
			t.Fatalf("upsell step %s should be removed", banned)
		}
	}
}

func TestAppendUTM(t *testing.T) {
	got := appendUTM("https://example.com/app", "d2_first_measure")
	if !strings.Contains(got, "utm_content=d2_first_measure") {
		t.Fatalf("unexpected utm url: %s", got)
	}
}

func TestAnnualNearRenewal(t *testing.T) {
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	until := now.AddDate(0, 0, 30)
	seg := store.JourneyClientSegment{HasAnnualPlan: true, AnnualValidUntil: &until}
	if !AnnualNearRenewal(seg, now) {
		t.Fatal("expected near renewal")
	}
	far := now.AddDate(0, 0, 120)
	seg.AnnualValidUntil = &far
	if AnnualNearRenewal(seg, now) {
		t.Fatal("120 days is not near")
	}
}
