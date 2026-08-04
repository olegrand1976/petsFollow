package sms

import (
	"regexp"
	"strings"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// Indicatifs des marchés couverts pour convertir un numéro national (0…).
var regionPrefixes = map[string]string{
	"BE": "+32",
	"FR": "+33",
	"NL": "+31",
	"LU": "+352",
}

// NormalizeE164 tente une normalisation légère et déterministe vers E.164 :
// séparateurs usuels retirés, 00 → +, 0 national → indicatif de defaultRegion.
// Tout numéro qui ne matche pas E.164 après ça est rejeté (jamais de devinette) —
// l'appelant journalise un skip invalid_phone.
func NormalizeE164(raw, defaultRegion string) (string, bool) {
	s := strings.TrimSpace(raw)
	for _, sep := range []string{" ", ".", "-", "/", "(", ")"} {
		s = strings.ReplaceAll(s, sep, "")
	}
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "00") {
		s = "+" + s[2:]
	} else if strings.HasPrefix(s, "0") {
		prefix, ok := regionPrefixes[strings.ToUpper(strings.TrimSpace(defaultRegion))]
		if !ok {
			return "", false
		}
		s = prefix + s[1:]
	}
	if !e164Pattern.MatchString(s) {
		return "", false
	}
	return s, true
}
