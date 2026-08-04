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
		{"et", "et", true},
		{"ET", "et", true},
		{"et-EE", "et", true},
		{"it", "it", true},
		{"IT", "it", true},
		{"it-IT", "it", true},
		{"uk", "uk", true},
		{"UK", "uk", true},
		{"uk-UA", "uk", true},
		{"ru", "ru", true},
		{"RU", "ru", true},
		{"ru-RU", "ru", true},
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
	if got := NormalizeLocale("et-EE"); got != "et" {
		t.Fatalf("NormalizeLocale(et-EE) = %q, want et", got)
	}
	if got := NormalizeLocale("it-IT"); got != "it" {
		t.Fatalf("NormalizeLocale(it-IT) = %q, want it", got)
	}
	if got := NormalizeLocale("uk-UA"); got != "uk" {
		t.Fatalf("NormalizeLocale(uk-UA) = %q, want uk", got)
	}
	if got := NormalizeLocale("ru-RU"); got != "ru" {
		t.Fatalf("NormalizeLocale(ru-RU) = %q, want ru", got)
	}
}

func TestIsCyrillicLocale(t *testing.T) {
	for _, loc := range []string{"uk", "ru", "uk-UA", "RU"} {
		if !IsCyrillicLocale(loc) {
			t.Errorf("IsCyrillicLocale(%q) = false, want true", loc)
		}
	}
	for _, loc := range []string{"fr", "nl", "en", "es", "et", "it", "xx"} {
		if IsCyrillicLocale(loc) {
			t.Errorf("IsCyrillicLocale(%q) = true, want false", loc)
		}
	}
}
