package i18n

import (
	"strings"
	"testing"
)

var smsKeys = []string{
	"sms.visit_confirmed",
	"sms.visit_reminder",
	"sms.visit_reschedule",
}

func TestAllSMSCatalogKeys(t *testing.T) {
	vars := map[string]string{"petName": "Rex", "when": "04/08/2026 10:30"}
	for _, loc := range Supported {
		for _, key := range smsKeys {
			got := T(loc, key, vars)
			if got == "" || got == key {
				t.Errorf("%s missing/unresolved key %s", loc, key)
			}
			if strings.Contains(got, "{") {
				t.Errorf("%s %s left an uninterpolated placeholder: %q", loc, key, got)
			}
		}
	}
}

func TestSMSInterpolation(t *testing.T) {
	got := T("en", "sms.visit_confirmed", map[string]string{"petName": "Bella", "when": "04/08/2026 10:30"})
	if got != "petsFollow: appointment confirmed for Bella on 04/08/2026 10:30." {
		t.Fatalf("got %q", got)
	}
}

// TestSMSLengthFitsSingleSegment garde chaque gabarit sur UN seul segment
// Telnyx. Le budget dépend de l'alphabet, d'où deux régimes :
//
//   - locales latines : GSM-7, 160 caractères par segment. Tout caractère hors
//     GSM-7 ferait basculer le message en UCS-2 et donc à 70 — on l'interdit.
//   - locales cyrilliques (uk, ru) : le cyrillique n'existe pas en GSM-7, ces
//     messages sont *nécessairement* en UCS-2, soit 70 caractères par segment.
//     Les gabarits y sont donc volontairement plus courts (pas de call-to-action
//     « ouvrez l'app » comme en français) : à 70 caractères, l'information du
//     créneau passe avant l'incitation. Ne pas rallonger sans vérifier ce test.
func TestSMSLengthFitsSingleSegment(t *testing.T) {
	// Nom d'animal long réaliste + date au format Europe/Brussels.
	vars := map[string]string{"petName": "Bartholomew", "when": "04/08/2026 10:30"}
	for _, loc := range Supported {
		cyrillic := IsCyrillicLocale(loc)
		limit := 160
		if cyrillic {
			limit = 70
		}
		for _, key := range smsKeys {
			got := T(loc, key, vars)
			if n := len([]rune(got)); n > limit {
				t.Errorf("%s %s = %d runes (>%d): %q", loc, key, n, limit, got)
			}
			if cyrillic {
				// UCS-2 accepte tout, mais un caractère d'un troisième script
				// signale une coquille de traduction (ex. idéogramme collé).
				if bad := nonCyrillicRunes(got); bad != "" {
					t.Errorf("%s %s contains runes outside Cyrillic/ASCII %q: %q", loc, key, bad, got)
				}
				continue
			}
			if bad := nonGSM7Runes(got); bad != "" {
				t.Errorf("%s %s contains non-GSM-7 runes %q (forces UCS-2, 70 chars/segment): %q", loc, key, bad, got)
			}
		}
	}
}

// gsm7Extra liste les caractères non-ASCII admis par l'alphabet GSM-7 (03.38)
// utiles pour nos locales latines.
const gsm7Extra = "£¥§¤èéùìòÇØøÅåÄÖÑÜäöñüàΔΦΓΛΩΠΨΣΘΞ€"

func nonGSM7Runes(s string) string {
	var bad []rune
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			if !strings.ContainsRune(gsm7Extra, r) {
				bad = append(bad, r)
			}
		}
	}
	return string(bad)
}

// cyrillicPunct : ponctuation non-ASCII tolérée dans un SMS cyrillique.
const cyrillicPunct = "«»—–’…€"

// nonCyrillicRunes renvoie les runes qui ne sont ni ASCII imprimable, ni
// cyrilliques (U+0400–U+04FF), ni de la ponctuation attendue.
func nonCyrillicRunes(s string) string {
	var bad []rune
	for _, r := range s {
		switch {
		case r >= 0x20 && r <= 0x7E:
		case r >= 0x0400 && r <= 0x04FF:
		case strings.ContainsRune(cyrillicPunct, r):
		default:
			bad = append(bad, r)
		}
	}
	return string(bad)
}
