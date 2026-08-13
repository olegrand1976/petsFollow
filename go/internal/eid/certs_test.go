package eid

import "testing"

func TestLoadTrustedCAs(t *testing.T) {
	cas, err := LoadTrustedCAs()
	if err != nil {
		t.Fatal(err)
	}
	if len(cas) < 3 {
		t.Fatalf("expected several CA certs, got %d", len(cas))
	}
}
