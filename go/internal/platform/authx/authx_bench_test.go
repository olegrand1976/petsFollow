package authx

import (
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// Chaque requête authentifiée parse un JWT : c'est le coût fixe payé avant
// d'atteindre le moindre handler.

func benchIssuer() *TokenIssuer {
	return NewTokenIssuer("bench-signing-key-not-for-prod", 15*time.Minute, 30*24*time.Hour)
}

func BenchmarkIssueProfile(b *testing.B) {
	issuer := benchIssuer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := issuer.IssueProfile("user-1", "vet@petsfollow.test", kernel.RoleVet, "practice-1", "profile-1", 3); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAccessToken(b *testing.B) {
	issuer := benchIssuer()
	pair, err := issuer.IssueProfile("user-1", "vet@petsfollow.test", kernel.RoleVet, "practice-1", "profile-1", 3)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := issuer.Parse(pair.AccessToken); err != nil {
			b.Fatal(err)
		}
	}
}
