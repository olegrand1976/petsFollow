package store

import "testing"

func TestNormalizeCountryCode(t *testing.T) {
	if got := NormalizeCountryCode("fr"); got != "FR" {
		t.Fatalf("got %s", got)
	}
	if got := NormalizeCountryCode("xx"); got != "BE" {
		t.Fatalf("invalid -> BE, got %s", got)
	}
	if got := NormalizeCountryCode(""); got != "BE" {
		t.Fatalf("empty -> BE, got %s", got)
	}
}
