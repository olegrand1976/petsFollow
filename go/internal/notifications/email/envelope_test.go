package email

import (
	"strings"
	"testing"
)

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

func TestSendConfirmFailsFastWhenSMTPPassEmpty(t *testing.T) {
	n := NewNotifierAuth(
		"smtp.invalid.petsfollow", 587,
		"petsFollow <noreply@petsfollow.app>",
		"noreply@petsfollow.app", "",
		"http://localhost:3002", "https://ll-it-sc.be",
	)
	err := n.SendConfirmRegistration("user@example.com", "fr", "User", "https://example.com/c")
	if err == nil {
		t.Fatal("expected error when SMTP_PASS empty")
	}
	if !strings.Contains(err.Error(), "SMTP_PASS empty") {
		t.Fatalf("got %v", err)
	}
}
