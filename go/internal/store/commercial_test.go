package store

import "testing"

func TestCommercialCommissionCents(t *testing.T) {
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
