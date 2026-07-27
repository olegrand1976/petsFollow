package store

import "testing"

func TestNormalizeMedicationName(t *testing.T) {
	got := NormalizeMedicationName("  Amoxicilline Élevé  ")
	if got != "amoxicilline eleve" {
		t.Fatalf("got %q", got)
	}
}
