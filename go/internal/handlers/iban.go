package handlers

import (
	"strings"
	"unicode"
)

// normalizeIBAN uppercases and strips spaces.
func normalizeIBAN(s string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}

// validIBAN checks length, charset and ISO 13616 mod-97 checksum.
func validIBAN(iban string) bool {
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}
	for _, r := range iban {
		if !unicode.IsDigit(r) && (r < 'A' || r > 'Z') {
			return false
		}
	}
	rearranged := iban[4:] + iban[:4]
	mod := 0
	for _, r := range rearranged {
		if r >= '0' && r <= '9' {
			mod = (mod*10 + int(r-'0')) % 97
			continue
		}
		v := int(r-'A') + 10
		mod = (mod*10 + v/10) % 97
		mod = (mod*10 + v%10) % 97
	}
	return mod == 1
}

// validBIC accepts empty, or 8/11 alphanumeric SWIFT codes.
func validBIC(bic string) bool {
	if bic == "" {
		return true
	}
	if len(bic) != 8 && len(bic) != 11 {
		return false
	}
	for _, r := range bic {
		if !unicode.IsDigit(r) && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}
