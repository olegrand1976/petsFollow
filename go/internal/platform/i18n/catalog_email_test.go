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
		"emails.visit_request_subject",
		"emails.visit_request_tagline",
		"emails.visit_request_preheader",
		"emails.visit_request_greeting",
		"emails.visit_request_intro",
		"emails.visit_request_detail",
		"emails.visit_request_cta",
		"emails.visit_request_disclaimer",
		"emails.footer_powered_by",
		"emails.footer_visit_llit",
		"emails.journey.unsubscribe",
	}
	vars := map[string]string{
		"fullName": "Ada", "bpm": "120", "vetName": "Dr. Vet", "practiceName": "VetPlus",
		"clientName": "Ada", "petName": "Rex", "when": "01/01/2026 10:00", "notes": "ok",
	}
	for _, loc := range Supported {
		for _, key := range keys {
			got := T(loc, key, vars)
			if got == "" || got == key {
				t.Errorf("%s missing/unresolved key %s", loc, key)
			}
		}
	}
}

func TestJourneyEmailCatalogKeys(t *testing.T) {
	steps := []string{
		"d0_welcome", "d1_activate", "d2_first_measure", "d4_routine", "d6_vet_link",
		"d10_visits", "d14_checkpoint", "d30_habit", "d45_care_plus", "d60_horse",
		"d75_kennel", "d90_quarter", "d120_seasonal", "d180_midyear", "d270_reengage",
		"d330_prerenew", "d365_anniversary",
		"evt_pending_payment", "evt_past_due", "evt_inactive_hr",
	}
	fields := []string{"subject", "tagline", "preheader", "greeting", "intro", "cta", "disclaimer"}
	vars := map[string]string{"fullName": "Ada"}
	for _, loc := range Supported {
		for _, step := range steps {
			for _, field := range fields {
				key := "emails.journey." + step + "." + field
				got := T(loc, key, vars)
				if got == "" || got == key {
					t.Errorf("%s missing/unresolved key %s", loc, key)
				}
			}
		}
		near := T(loc, "emails.journey.d330_prerenew.intro_near", vars)
		if near == "" || near == "emails.journey.d330_prerenew.intro_near" {
			t.Errorf("%s missing d330 intro_near", loc)
		}
	}
}
