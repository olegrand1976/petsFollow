package gemini

import (
	"strings"
	"unicode"
)

const maxVisitReportRunes = 100_000

// NormalizeVisitReportText strips NULs, applies NFKC-ish cleanup (space collapse edges),
// and caps length so Gemini / store never choke on pathological payloads.
func NormalizeVisitReportText(raw string) string {
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if r == 0 {
			continue
		}
		// Drop other C0 controls except tab/newline/carriage-return.
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			continue
		}
		if unicode.Is(unicode.Cc, r) && r != '\t' && r != '\n' && r != '\r' {
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	runes := []rune(out)
	if len(runes) > maxVisitReportRunes {
		out = string(runes[:maxVisitReportRunes])
	}
	return out
}
