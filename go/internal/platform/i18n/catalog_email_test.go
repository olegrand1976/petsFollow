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
		"emails.heartrate_alert_subject",
		"emails.heartrate_alert_tagline",
		"emails.heartrate_alert_preheader",
		"emails.heartrate_alert_greeting",
		"emails.heartrate_alert_intro",
		"emails.heartrate_alert_disclaimer",
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
		"emails.preconsult_urgent_subject",
		"emails.preconsult_urgent_tagline",
		"emails.preconsult_urgent_preheader",
		"emails.preconsult_urgent_greeting",
		"emails.preconsult_urgent_intro",
		"emails.preconsult_urgent_detail",
		"emails.preconsult_urgent_cta",
		"emails.preconsult_urgent_disclaimer",
		"emails.preconsult_urgency_low",
		"emails.preconsult_urgency_medium",
		"emails.preconsult_urgency_high",
		"emails.visit_preconsult_subject",
		"emails.visit_preconsult_tagline",
		"emails.visit_preconsult_preheader",
		"emails.visit_preconsult_greeting",
		"emails.visit_preconsult_intro",
		"emails.visit_preconsult_detail",
		"emails.visit_preconsult_cta",
		"emails.visit_preconsult_disclaimer",
		"emails.proforma_validate_subject",
		"emails.proforma_validate_tagline",
		"emails.proforma_validate_preheader",
		"emails.proforma_validate_greeting",
		"emails.proforma_validate_intro",
		"emails.proforma_validate_detail",
		"emails.proforma_validate_cta",
		"emails.proforma_validate_disclaimer",
		"emails.visit_confirmed_affiliate_subject",
		"emails.visit_confirmed_affiliate_tagline",
		"emails.visit_confirmed_affiliate_preheader",
		"emails.visit_confirmed_affiliate_greeting",
		"emails.visit_confirmed_affiliate_intro",
		"emails.visit_confirmed_affiliate_detail",
		"emails.visit_confirmed_affiliate_cta",
		"emails.visit_confirmed_affiliate_disclaimer",
		"emails.footer_powered_by",
		"emails.footer_visit_llit",
		"emails.product_digest_subject",
		"emails.product_digest_tagline",
		"emails.product_digest_preheader",
		"emails.product_digest_greeting",
		"emails.product_digest_intro",
		"emails.product_digest_intro_with_headline",
		"emails.product_digest_disclaimer",
		"emails.product_digest_fallback_name",
		"emails.product_digest_test_env_tag",
		"emails.product_digest_weekly_subject",
		"emails.product_digest_weekly_tagline",
		"emails.product_digest_weekly_preheader",
		"emails.product_digest_weekly_greeting",
		"emails.product_digest_weekly_intro",
		"emails.product_digest_weekly_intro_with_headline",
		"emails.product_digest_weekly_disclaimer",
		"emails.product_digest_weekly_fallback_name",
		"emails.staging_seed_subject",
		"emails.staging_seed_tagline",
		"emails.staging_seed_schedule",
		"emails.staging_seed_preheader",
		"emails.staging_seed_greeting",
		"emails.staging_seed_intro",
		"emails.staging_seed_detail",
		"emails.staging_seed_cta",
		"emails.staging_seed_disclaimer",
		"emails.staging_seed_fallback_name",
		"emails.support_ticket_ops_subject",
		"emails.support_ticket_ops_tagline",
		"emails.support_ticket_ops_preheader",
		"emails.support_ticket_ops_greeting",
		"emails.support_ticket_ops_intro",
		"emails.support_ticket_ops_detail",
		"emails.support_ticket_ops_cta",
		"emails.support_ticket_ops_disclaimer",
		"emails.support_ticket_reply_subject",
		"emails.support_ticket_reply_tagline",
		"emails.support_ticket_reply_preheader",
		"emails.support_ticket_reply_greeting",
		"emails.support_ticket_reply_intro",
		"emails.support_ticket_reply_detail",
		"emails.support_ticket_reply_cta",
		"emails.support_ticket_reply_disclaimer",
		"emails.support_ticket_created_ack_subject",
		"emails.support_ticket_created_ack_tagline",
		"emails.support_ticket_created_ack_preheader",
		"emails.support_ticket_created_ack_greeting",
		"emails.support_ticket_created_ack_intro",
		"emails.support_ticket_created_ack_detail",
		"emails.support_ticket_created_ack_cta",
		"emails.support_ticket_created_ack_disclaimer",
		"emails.support_ticket_status_subject",
		"emails.support_ticket_status_tagline",
		"emails.support_ticket_status_preheader",
		"emails.support_ticket_status_greeting",
		"emails.support_ticket_status_intro",
		"emails.support_ticket_status_detail",
		"emails.support_ticket_status_cta",
		"emails.support_ticket_status_disclaimer",
		"emails.support_status_open",
		"emails.support_status_in_progress",
		"emails.support_status_to_test",
		"emails.support_status_done",
		"emails.support_status_closed",
		"emails.support_status_unknown",
		"emails.journey.unsubscribe",
	}
	vars := map[string]string{
		"fullName": "Ada", "bpm": "120", "vetName": "Dr. Vet", "practiceName": "VetPlus",
		"clientName": "Ada", "petName": "Rex", "when": "01/01/2026 10:00", "notes": "ok",
		"date": "22/07/2026", "headline": "Améliorations",
		"schedule": "dimanche 08:00",
		"ticketId": "abc", "subject": "Bug", "email": "a@b.c", "role": "vet", "source": "nuxt_pro",
		"message": "oops", "replyBody": "fix soon",
		"fromStatus": "Reçu", "toStatus": "Corrigé", "changedByName": "Admin Demo",
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
		"d10_visits", "d14_checkpoint", "d30_habit", "d90_quarter", "d120_seasonal",
		"d180_midyear", "d270_reengage", "d330_prerenew", "d365_anniversary",
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
