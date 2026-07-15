package email

import (
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

func TestRenderBrandedEmail_containsBrandingAndCTA(t *testing.T) {
	html := renderBrandedEmail(brandedEmailContent{
		Lang:            "fr",
		Tagline:         "Suivi cardiaque vétérinaire",
		Greeting:        "Bonjour Dr Test,",
		Intro:           "Bienvenue sur petsFollow Pro.",
		CTALabel:        "Confirmer mon compte",
		CTAURL:          "http://localhost:3002/confirm-email?token=abc",
		Expiry:          "Ce lien expire dans 48 heures.",
		Disclaimer:      "Ignorez cet email si besoin.",
		Preheader:       "Activez votre compte.",
		FooterPoweredBy: "petsFollow — propulsé par LL-IT Software & Computer",
		FooterVisit:     "Visiter le site LL-IT Software & Computer",
		Brand: brandAssets{
			LLITLogoURL:    "http://localhost:3002/brand/ll-it-logo.png",
			LLITWebsiteURL: "https://ll-it-sc.be",
			SiteURL:        "http://localhost:3002",
		},
	})

	checks := []string{
		"petsFollow",
		"Pro",
		"#1B3A4B",
		"#2A9D8F",
		"Bonjour Dr Test,",
		"Confirmer mon compte",
		"http://localhost:3002/confirm-email?token=abc",
		"Ce lien expire dans 48 heures.",
		"ll-it-logo.png",
		"https://ll-it-sc.be",
		"LL-IT Software &amp; Computer",
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered email to contain %q", want)
		}
	}
}

func TestRenderBrandedEmail_escapesHTML(t *testing.T) {
	html := renderBrandedEmail(brandedEmailContent{
		Greeting: `<script>alert("x")</script>`,
		CTAURL:   `"><img onerror=alert(1)>`,
		Detail:   `<b>raw</b>`,
	})

	if strings.Contains(html, "<script>") {
		t.Fatal("expected greeting to be escaped")
	}
	if strings.Contains(html, `href=""><img`) {
		t.Fatal("expected CTA URL to be escaped in href attribute")
	}
	if strings.Contains(html, "<b>raw</b>") {
		t.Fatal("expected detail to be escaped")
	}
}

func TestRenderPasswordResetEmail_containsResetURL(t *testing.T) {
	n := NewNotifier("127.0.0.1", 9, "test@petsfollow.test", "http://localhost:3002", "https://ll-it-sc.be")
	locale := "fr"
	html := renderBrandedEmail(brandedEmailContent{
		Lang:            locale,
		Tagline:         i18n.T(locale, "emails.password_reset_tagline", nil),
		Greeting:        i18n.T(locale, "emails.password_reset_greeting", map[string]string{"fullName": "Dr Reset"}),
		Intro:           i18n.T(locale, "emails.password_reset_intro", nil),
		CTALabel:        i18n.T(locale, "emails.password_reset_cta", nil),
		CTAURL:          "http://localhost:3002/reset-password?token=demo-reset",
		Expiry:          i18n.T(locale, "emails.password_reset_expiry", nil),
		Disclaimer:      i18n.T(locale, "emails.password_reset_disclaimer", nil),
		Preheader:       i18n.T(locale, "emails.password_reset_preheader", nil),
		FooterPoweredBy: i18n.T(locale, "emails.footer_powered_by", nil),
		FooterVisit:     i18n.T(locale, "emails.footer_visit_llit", nil),
		Brand:           n.brandURLs(),
	})
	for _, want := range []string{
		"Réinitialiser mon mot de passe",
		"http://localhost:3002/reset-password?token=demo-reset",
		"Ce lien expire dans 1 heure.",
		"ll-it-logo.png",
		"https://ll-it-sc.be",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered email to contain %q", want)
		}
	}
	if strings.Contains(html, "emails.password_reset_") {
		t.Fatal("expected i18n keys to be resolved, not raw")
	}
	_ = n
}

func TestAllEmailLocalesResolve(t *testing.T) {
	keys := []string{
		"emails.confirm_registration_subject",
		"emails.password_reset_subject",
		"emails.heartrate_validated_subject",
		"emails.heartrate_validated_intro",
		"emails.new_message_subject",
		"emails.new_message_intro",
		"emails.footer_powered_by",
		"emails.footer_visit_llit",
	}
	for _, loc := range []string{"fr", "en", "nl"} {
		for _, key := range keys {
			got := i18n.T(loc, key, map[string]string{"fullName": "Ada", "bpm": "120", "body": "hi"})
			if got == key {
				t.Errorf("%s/%s unresolved", loc, key)
			}
		}
	}
}
