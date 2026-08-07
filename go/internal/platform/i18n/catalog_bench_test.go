package i18n

import "testing"

// Toute réponse d'erreur localisée passe par T, et chaque requête par
// ParseAcceptLanguage : deux chemins chauds côté allocations.

func BenchmarkTranslate(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = T("fr", "errors.invalid_credentials", nil)
	}
}

func BenchmarkTranslateWithVars(b *testing.B) {
	vars := map[string]string{"name": "Spirit"}
	b.ReportAllocs()
	for b.Loop() {
		_ = T("nl", "errors.invalid_credentials", vars)
	}
}

func BenchmarkParseAcceptLanguage(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = ParseAcceptLanguage("nl-BE,nl;q=0.9,fr;q=0.8,en;q=0.7")
	}
}
