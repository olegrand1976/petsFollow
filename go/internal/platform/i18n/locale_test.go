package i18n

import "testing"

func TestMatchSupported(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"es", "es", true},
		{"ES", "es", true},
		{"es-ES", "es", true},
		{"fr-FR", "fr", true},
		{"nl", "nl", true},
		{"en-GB", "en", true},
		{"xx", "", false},
		{"", "", false},
		{"  es-ES ;q=0.9", "es", true},
	}
	for _, tc := range cases {
		got, ok := MatchSupported(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("MatchSupported(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestNormalizeLocaleFallsBackToFr(t *testing.T) {
	if got := NormalizeLocale("xx"); got != "fr" {
		t.Fatalf("NormalizeLocale(xx) = %q, want fr", got)
	}
	if got := NormalizeLocale("es-ES"); got != "es" {
		t.Fatalf("NormalizeLocale(es-ES) = %q, want es", got)
	}
}
