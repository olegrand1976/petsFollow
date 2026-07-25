package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestHTVACents(t *testing.T) {
	if store.DefaultVATRateBps != 2100 {
		t.Fatalf("expected 2100 bps VAT, got %d", store.DefaultVATRateBps)
	}
	if got := store.HTVACents(2900); got != 2396 {
		t.Fatalf("HTVA of 2900 = 2396, got %d", got)
	}
	if got := store.HTVACents(7900); got != 6528 {
		t.Fatalf("HTVA of 7900 = 6528, got %d", got)
	}
	if got := store.HTVACents(11500); got != 9504 {
		t.Fatalf("HTVA of 11500 = 9504, got %d", got)
	}
	if got := store.HTVACents(0); got != 0 {
		t.Fatalf("HTVA of 0 = 0, got %d", got)
	}
	if got := store.HTVACents(-100); got != 0 {
		t.Fatalf("HTVA of negative = 0, got %d", got)
	}
}

func TestCommissionFromTTCCents(t *testing.T) {
	// 12% of HTVA(2900)=2396 → 287
	if got := store.CommissionFromTTCCents(2900, 1200); got != 287 {
		t.Fatalf("12%% of HTVA(2900) = 287, got %d", got)
	}
	// 12% of HTVA(7900)=6528 → 783
	if got := store.CommissionFromTTCCents(7900, 1200); got != 783 {
		t.Fatalf("12%% of HTVA(7900) = 783, got %d", got)
	}
	if got := store.CommissionFromTTCCents(2900, 0); got != 0 {
		t.Fatalf("0%% of HTVA(2900) = 0, got %d", got)
	}
}
