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
		{"", "petsFollow Pro — Réinitialisation du mot de passe"},
	}
	for _, tc := range cases {
		got := T(tc.loc, "emails.password_reset_subject", nil)
		if got == "emails.password_reset_subject" {
			t.Errorf("%q: key unresolved (returned key itself)", tc.loc)
			continue
		}
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.loc, got, tc.want)
		}
	}

	greet := T("en", "emails.password_reset_greeting", map[string]string{"fullName": "Ada"})
	if greet != "Hello Ada," {
		t.Errorf("greeting interpolation: got %q", greet)
	}
}

func TestAllEmailCatalogKeys(t *testing.T) {
	keys := []string{
		"emails.confirm_registration_subject",
		"emails.confirm_registration_tagline",
		"emails.confirm_registration_preheader",
		"emails.confirm_registration_greeting",
		"emails.confirm_registration_intro",
		"emails.confirm_registration_cta",
		"emails.confirm_registration_expiry",
		"emails.confirm_registration_disclaimer",
		"emails.password_reset_subject",
		"emails.password_reset_tagline",
		"emails.password_reset_preheader",
		"emails.password_reset_greeting",
		"emails.password_reset_intro",
		"emails.password_reset_cta",
		"emails.password_reset_expiry",
		"emails.password_reset_disclaimer",
		"emails.heartrate_validated_subject",
		"emails.heartrate_validated_tagline",
		"emails.heartrate_validated_preheader",
		"emails.heartrate_validated_greeting",
		"emails.heartrate_validated_intro",
		"emails.heartrate_validated_disclaimer",
		"emails.new_message_subject",
		"emails.new_message_tagline",
		"emails.new_message_preheader",
		"emails.new_message_greeting",
		"emails.new_message_intro",
		"emails.new_message_disclaimer",
		"emails.app_download_subject",
		"emails.app_download_tagline",
		"emails.app_download_preheader",
		"emails.app_download_greeting",
		"emails.app_download_intro",
		"emails.app_download_cta",
		"emails.app_download_disclaimer",
		"emails.footer_powered_by",
		"emails.footer_visit_llit",
	}
	vars := map[string]string{"fullName": "Ada", "bpm": "120", "vetName": "Dr. Vet", "practiceName": "VetPlus"}
	for _, loc := range Supported {
		for _, key := range keys {
			got := T(loc, key, vars)
			if got == "" || got == key {
				t.Errorf("%s missing/unresolved key %s", loc, key)
			}
		}
	}
}
