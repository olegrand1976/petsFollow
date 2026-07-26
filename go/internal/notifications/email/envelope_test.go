package email

import "testing"

func TestEnvelopeFrom(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"noreply@petsfollow.app", "noreply@petsfollow.app"},
		{"petsFollow <noreply@petsfollow.app>", "noreply@petsfollow.app"},
		{"  petsFollow <noreply@petsfollow.app>  ", "noreply@petsfollow.app"},
		{"petsFollow <noreply@ll-it-sc.be>", "noreply@ll-it-sc.be"},
		{"", ""},
		{"   ", ""},
	}
	for _, tc := range cases {
		if got := envelopeFrom(tc.in); got != tc.want {
			t.Errorf("envelopeFrom(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
