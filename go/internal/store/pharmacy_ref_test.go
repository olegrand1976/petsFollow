package store

import "testing"

func TestNormalizeMedicationName(t *testing.T) {
	got := NormalizeMedicationName("  Amoxicilline Élevé  ")
	if got != "amoxicilline eleve" {
		t.Fatalf("got %q", got)
	}
}

func TestParseMedicationLetter(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"a", "A", true},
		{" Z ", "Z", true},
		{"#", "#", true},
		{"", "", false},
		{"AB", "", false},
		{"1", "", false},
		{"@", "", false},
	}
	for _, tc := range cases {
		got, ok := ParseMedicationLetter(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("%q → (%q,%v) want (%q,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
