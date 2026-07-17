package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestCommercialCommissionCents(t *testing.T) {
	if got := store.CommercialCommissionCents(2900, 1200); got != 348 {
		t.Fatalf("12%% of 2900 = 348, got %d", got)
	}
	if got := store.CommercialCommissionCents(7500, 1200); got != 900 {
		t.Fatalf("12%% of 7500 = 900, got %d", got)
	}
	if got := store.CommercialCommissionCents(2900, 0); got != 0 {
		t.Fatalf("0%% of 2900 = 0, got %d", got)
	}
	if store.DefaultCommercialCommissionRateBps != 1200 {
		t.Fatalf("expected 1200 bps, got %d", store.DefaultCommercialCommissionRateBps)
	}
}
