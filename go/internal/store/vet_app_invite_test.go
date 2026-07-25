package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestNormalizeInviteCode(t *testing.T) {
	if got := store.NormalizeInviteCode("  ab12cd34  "); got != "AB12CD34" {
		t.Fatalf("got %q", got)
	}
	if got := store.NormalizeInviteCode(""); got != "" {
		t.Fatalf("empty got %q", got)
	}
}
