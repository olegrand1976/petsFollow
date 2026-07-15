package i18n

import "testing"

func TestPasswordResetKeys(t *testing.T) {
	cases := []struct {
		loc  string
		want string
	}{
		{"fr", "petsFollow Pro — Réinitialisation du mot de passe"},
		{"en", "petsFollow Pro — Password reset"},
		{"nl", "petsFollow Pro — Wachtwoord resetten"},
		{"", "petsFollow Pro — Réinitialisation du mot de passe"}, // fallback FR
	}
	for _, tc := range cases {
		got := T(tc.loc, "emails.password_reset_subject", nil)
		if got == "emails.password_reset_subject" {
			t.Errorf("%q: key unresolved (returned key itself)", tc.loc)
			continue
		}
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.loc, got, tc.want)
		} else {
			t.Logf("%q -> %q OK", tc.loc, got)
		}
	}

	// Interpolation smoke
	greet := T("en", "emails.password_reset_greeting", map[string]string{"fullName": "Ada"})
	if greet != "Hello Ada," {
		t.Errorf("greeting interpolation: got %q", greet)
	}
}
