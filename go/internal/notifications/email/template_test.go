package email

import (
	"strings"
	"testing"
)

func TestRenderConfirmRegistration_containsBrandingAndCTA(t *testing.T) {
	html := renderConfirmRegistration(confirmRegistrationContent{
		Tagline:    "Suivi cardiaque vétérinaire",
		Greeting:   "Bonjour Dr Test,",
		Intro:      "Bienvenue sur petsFollow Pro.",
		CTALabel:   "Confirmer mon compte",
		Expiry:     "Ce lien expire dans 48 heures.",
		Disclaimer: "Ignorez cet email si besoin.",
		Preheader:  "Activez votre compte.",
		ConfirmURL: "http://localhost:3002/confirm-email?token=abc",
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
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered email to contain %q", want)
		}
	}
}

func TestRenderConfirmRegistration_escapesHTML(t *testing.T) {
	html := renderConfirmRegistration(confirmRegistrationContent{
		Greeting:   `<script>alert("x")</script>`,
		ConfirmURL: `"><img onerror=alert(1)>`,
	})

	if strings.Contains(html, "<script>") {
		t.Fatal("expected greeting to be escaped")
	}
	if strings.Contains(html, `href=""><img`) {
		t.Fatal("expected confirm URL to be escaped in href attribute")
	}
}

func TestRenderPasswordResetEmail_containsResetURL(t *testing.T) {
	html := renderConfirmRegistration(confirmRegistrationContent{
		Tagline:    "Suivi cardiaque vétérinaire",
		Greeting:   "Bonjour Dr Reset,",
		Intro:      "Réinitialisez votre mot de passe.",
		CTALabel:   "Réinitialiser mon mot de passe",
		Expiry:     "Ce lien expire dans 1 heure.",
		Disclaimer: "Ignorez si non demandé.",
		Preheader:  "Reset password.",
		ConfirmURL: "http://localhost:3002/reset-password?token=demo-reset",
	})
	for _, want := range []string{
		"Réinitialiser mon mot de passe",
		"http://localhost:3002/reset-password?token=demo-reset",
		"Ce lien expire dans 1 heure.",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered email to contain %q", want)
		}
	}
}
