package store

import "testing"

func TestCommercialCommissionCents(t *testing.T) {
	// Addon path still uses flat 15%.
	if got := CommercialCommissionCents(2500); got != 375 {
		t.Fatalf("15%% of 2500 = 375, got %d", got)
	}
	if got := CommercialCommissionCents(6000); got != 900 {
		t.Fatalf("15%% of 6000 = 900, got %d", got)
	}
	if CommercialCommissionRateBps != 1500 {
		t.Fatalf("expected 1500 bps, got %d", CommercialCommissionRateBps)
	}
}

func TestValidProspectStatus(t *testing.T) {
	for _, s := range []string{"new", "contacted", "qualified", "converted", "lost"} {
		if !ValidProspectStatus(s) {
			t.Fatalf("expected valid status %q", s)
		}
	}
	if ValidProspectStatus("unknown") {
		t.Fatal("expected invalid status")
	}
}

func TestValidProspectSource(t *testing.T) {
	if !ValidProspectSource("commercial") || !ValidProspectSource("vet_referral") {
		t.Fatal("expected commercial and vet_referral valid")
	}
	if ValidProspectSource("other") {
		t.Fatal("expected invalid source")
	}
}
