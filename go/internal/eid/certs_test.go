package eid

import (
	"strings"
	"testing"
)

func TestLoadTrustedCAs(t *testing.T) {
	cas, err := LoadTrustedCAs()
	if err != nil {
		t.Fatal(err)
	}
	// Full sync (Citizen + Foreigner + roots) — see scripts/sync-eid-certs.sh.
	if len(cas) < 100 {
		t.Fatalf("expected full BE CA pack (>=100), got %d", len(cas))
	}

	var hasForeigner, hasRootCA6 bool
	for _, c := range cas {
		if !c.IsCA {
			t.Fatalf("non-CA slipped into trust store: CN=%q", c.Subject.CommonName)
		}
		cn := c.Subject.CommonName
		if strings.Contains(cn, "Foreigner CA") {
			hasForeigner = true
		}
		if cn == "Belgium Root CA6" {
			hasRootCA6 = true
		}
	}
	if !hasForeigner {
		t.Fatal("expected at least one CA with CN containing \"Foreigner CA\" (eidf*.crt)")
	}
	if !hasRootCA6 {
		t.Fatal("expected Belgium Root CA6 (belgiumrca6.crt) in trust store")
	}
}
