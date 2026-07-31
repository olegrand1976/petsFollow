package email

import (
	"net/smtp"
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

func TestPickSMTPAuthPrefersLOGIN(t *testing.T) {
	a, err := pickSMTPAuth("pro1.mail.ovh.net", "u", "p", "GSSAPI NTLM LOGIN")
	if err != nil {
		t.Fatal(err)
	}
	mech, proto, err := a.Start(&smtp.ServerInfo{Name: "pro1.mail.ovh.net", TLS: true, Auth: []string{"LOGIN"}})
	if err != nil {
		t.Fatal(err)
	}
	if mech != "LOGIN" || proto != nil {
		t.Fatalf("got mech=%q proto=%v", mech, proto)
	}
	user, err := a.Next([]byte("Username:"), true)
	if err != nil || string(user) != "u" {
		t.Fatalf("username step: %q %v", user, err)
	}
	pass, err := a.Next([]byte("Password:"), true)
	if err != nil || string(pass) != "p" {
		t.Fatalf("password step: %q %v", pass, err)
	}
}

func TestPickSMTPAuthPLAINWhenNoLOGIN(t *testing.T) {
	a, err := pickSMTPAuth("mail.example.com", "u", "p", "PLAIN")
	if err != nil {
		t.Fatal(err)
	}
	mech, _, err := a.Start(&smtp.ServerInfo{Name: "mail.example.com", TLS: true, Auth: []string{"PLAIN"}})
	if err != nil {
		t.Fatal(err)
	}
	if mech != "PLAIN" {
		t.Fatalf("got %q", mech)
	}
}
