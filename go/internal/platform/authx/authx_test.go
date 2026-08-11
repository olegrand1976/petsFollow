package authx

import (
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestParseRejectsNonHMACAlg(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	// Hand-crafted header with alg=none (classic confusion). Must not parse as access.
	noneTok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEiLCJlbWFpbCI6InZldEB0ZXN0LmNvbSIsInJvbGUiOiJ2ZXQiLCJ0eXAiOiJhY2Nlc3MifQ."
	if _, err := issuer.Parse(noneTok); err == nil {
		t.Fatal("expected alg=none rejected by Parse")
	}
	if _, err := issuer.ParseMFA(noneTok); err == nil {
		t.Fatal("expected alg=none rejected by ParseMFA")
	}
}

func TestIssueAndParseAccessToken(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	pair, err := issuer.IssueProfile("user-1", "vet@test.com", kernel.RoleVet, "practice-1", "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected tokens")
	}
	id, err := issuer.Parse(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "user-1" || id.Email != "vet@test.com" || id.Role != kernel.RoleVet || id.PracticeID != "practice-1" {
		t.Fatalf("unexpected identity %+v", id)
	}
}

func TestParseRejectsRefreshAsAccess(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	pair, err := issuer.IssueProfile("user-1", "vet@test.com", kernel.RoleVet, "practice-1", "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issuer.Parse(pair.RefreshToken); err == nil {
		t.Fatal("expected refresh token to be rejected by Parse")
	}
}

func TestParseRefresh(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	pair, err := issuer.IssueProfile("user-1", "vet@test.com", kernel.RoleVet, "practice-1", "", 3)
	if err != nil {
		t.Fatal(err)
	}
	id, err := issuer.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "user-1" || id.Email != "vet@test.com" || id.Role != kernel.RoleVet || id.PracticeID != "practice-1" {
		t.Fatalf("unexpected identity %+v", id)
	}
	// La révocation se joue au refresh : sans ce claim, tout token resterait valide.
	if id.TokenVersion != 3 {
		t.Fatalf("expected token version 3, got %d", id.TokenVersion)
	}
	if _, err := issuer.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("expected access token to be rejected by ParseRefresh")
	}
}

func TestIssueAndParseMFA(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	mfa, err := issuer.IssueMFA("user-1", "vet@test.com", kernel.RoleVet, "practice-1")
	if err != nil {
		t.Fatal(err)
	}
	if !mfa.Requires2FA || mfa.MFAToken == "" {
		t.Fatalf("unexpected mfa %+v", mfa)
	}
	id, err := issuer.ParseMFA(mfa.MFAToken)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "user-1" {
		t.Fatalf("unexpected user %s", id.UserID)
	}
	if _, err := issuer.Parse(mfa.MFAToken); err == nil {
		t.Fatal("expected MFA token rejected by Parse")
	}
}

func TestJourneyUnsubscribeToken(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Minute, time.Hour)
	tok, err := issuer.IssueJourneyUnsubscribe("client-1", "client@test.com")
	if err != nil {
		t.Fatal(err)
	}
	id, err := issuer.ParseJourneyUnsubscribe(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != "client-1" || id.Email != "client@test.com" || id.Role != kernel.RoleClient {
		t.Fatalf("unexpected identity %+v", id)
	}
	if _, err := issuer.Parse(tok); err == nil {
		t.Fatal("expected journey token rejected by Parse")
	}
}
