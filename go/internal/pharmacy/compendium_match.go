package pharmacy

import (
	"strings"
	"unicode"
)

// CNKMatchCandidate is one AFMPS dictionary hit for human review.
type CNKMatchCandidate struct {
	CNK                string  `json:"cnk"`
	Name               string  `json:"name"`
	PharmaceuticalForm string  `json:"pharmaceuticalForm,omitempty"`
	PackSize           string  `json:"packSize,omitempty"`
	Score              float64 `json:"score"`
}

// CNKMatchResult is the outcome of SuggestCNK for one extracted row.
type CNKMatchResult struct {
	SuggestedCNK string
	Score        float64
	Candidates   []CNKMatchCandidate
	AutoFill     bool // true when top hit is confident enough to prefill cnk
}

// RefMedMatchInput is a slim AFMPS row used for scoring.
type RefMedMatchInput struct {
	CNK                string
	Name               string
	PharmaceuticalForm string
	PackSize           string
	Manufacturer       string // optional, from afmps_meta if present
}

const (
	cnkMatchAutoThreshold   = 0.78
	cnkMatchGapThreshold    = 0.08
	cnkMatchMaxCandidates   = 8
	cnkMatchMinCandidateScr = 0.35
)

// SuggestCNK ranks AFMPS hits against an extracted monograph (no invented CNK).
func SuggestCNK(extracted ExtractedMedication, refs []RefMedMatchInput) CNKMatchResult {
	var out CNKMatchResult
	if len(refs) == 0 {
		return out
	}
	queryName := strings.TrimSpace(extracted.Name)
	if queryName == "" {
		return out
	}
	qNorm := normalizeMatchText(queryName)
	qBase := stripDosageTokens(qNorm)
	qForm := normalizeMatchText(extracted.PharmaceuticalForm)
	qPack := normalizeMatchText(extracted.PackSize)
	qMfr := normalizeMatchText(extracted.Manufacturer)

	cands := make([]CNKMatchCandidate, 0, len(refs))
	for _, ref := range refs {
		if strings.TrimSpace(ref.CNK) == "" || strings.TrimSpace(ref.Name) == "" {
			continue
		}
		rNorm := normalizeMatchText(ref.Name)
		rBase := stripDosageTokens(rNorm)
		score := tokenDice(qNorm, rNorm) * 0.55
		score += tokenDice(qBase, rBase) * 0.25
		if qBase != "" && qBase == rBase {
			score += 0.10
		}
		if brandToken(qNorm) != "" && brandToken(qNorm) == brandToken(rNorm) {
			score += 0.08
		}
		if qForm != "" && normalizeMatchText(ref.PharmaceuticalForm) != "" {
			if strings.Contains(normalizeMatchText(ref.PharmaceuticalForm), qForm) ||
				strings.Contains(qForm, normalizeMatchText(ref.PharmaceuticalForm)) {
				score += 0.12
			}
		}
		if qPack != "" && normalizeMatchText(ref.PackSize) != "" {
			if strings.Contains(normalizeMatchText(ref.PackSize), qPack) ||
				strings.Contains(qPack, normalizeMatchText(ref.PackSize)) {
				score += 0.05
			}
		}
		if qMfr != "" && normalizeMatchText(ref.Manufacturer) != "" {
			if strings.Contains(normalizeMatchText(ref.Manufacturer), qMfr) ||
				strings.Contains(qMfr, normalizeMatchText(ref.Manufacturer)) {
				score += 0.08
			}
		}
		if score > 1 {
			score = 1
		}
		if score < cnkMatchMinCandidateScr {
			continue
		}
		cands = append(cands, CNKMatchCandidate{
			CNK:                ref.CNK,
			Name:               ref.Name,
			PharmaceuticalForm: ref.PharmaceuticalForm,
			PackSize:           ref.PackSize,
			Score:              score,
		})
	}
	// insertion sort by score desc (small N)
	for i := 1; i < len(cands); i++ {
		j := i
		for j > 0 && cands[j].Score > cands[j-1].Score {
			cands[j], cands[j-1] = cands[j-1], cands[j]
			j--
		}
	}
	if len(cands) > cnkMatchMaxCandidates {
		cands = cands[:cnkMatchMaxCandidates]
	}
	out.Candidates = cands
	if len(cands) == 0 {
		return out
	}
	out.SuggestedCNK = cands[0].CNK
	out.Score = cands[0].Score
	gapOK := len(cands) == 1 || (cands[0].Score-cands[1].Score) >= cnkMatchGapThreshold
	out.AutoFill = cands[0].Score >= cnkMatchAutoThreshold && gapOK
	return out
}

func normalizeMatchText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
			continue
		}
		if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func brandToken(norm string) string {
	parts := strings.Fields(norm)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func stripDosageTokens(norm string) string {
	parts := strings.Fields(norm)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if looksLikeDosageToken(p) {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

func looksLikeDosageToken(p string) bool {
	if p == "mg" || p == "ml" || p == "g" || p == "kg" || p == "ui" || p == "%" {
		return true
	}
	hasDigit := false
	for _, r := range p {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	return hasDigit
}

func tokenDice(a, b string) float64 {
	ta := uniqueTokens(a)
	tb := uniqueTokens(b)
	if len(ta) == 0 || len(tb) == 0 {
		if a == b && a != "" {
			return 1
		}
		return 0
	}
	inter := 0
	for t := range ta {
		if tb[t] {
			inter++
		}
	}
	return (2 * float64(inter)) / float64(len(ta)+len(tb))
}

func uniqueTokens(s string) map[string]bool {
	m := make(map[string]bool)
	for p := range strings.FieldsSeq(s) {
		if p != "" {
			m[p] = true
		}
	}
	return m
}
