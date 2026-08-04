package sms

import "testing"

func TestNormalizeE164(t *testing.T) {
	cases := []struct {
		raw    string
		region string
		want   string
		ok     bool
	}{
		{"0470 12 34 56", "BE", "+32470123456", true},
		{"0470.12.34.56", "BE", "+32470123456", true},
		{"06 12 34 56 78", "FR", "+33612345678", true},
		{"0032 470 12 34 56", "BE", "+32470123456", true},
		{"+32 470/12.34.56", "BE", "+32470123456", true},
		{"+32470123456", "", "+32470123456", true},
		{"(0)470123456", "BE", "+32470123456", true},
		{"", "BE", "", false},
		{"   ", "BE", "", false},
		{"pas un numero", "BE", "", false},
		{"04701234", "XX", "", false},             // région inconnue
		{"12345", "BE", "", false},                // pas de préfixe interprétable
		{"+3212", "BE", "", false},                // trop court
		{"+3247012345678901234", "BE", "", false}, // trop long
	}
	for _, c := range cases {
		got, ok := NormalizeE164(c.raw, c.region)
		if got != c.want || ok != c.ok {
			t.Errorf("NormalizeE164(%q, %q) = (%q, %v), want (%q, %v)", c.raw, c.region, got, ok, c.want, c.ok)
		}
	}
}
