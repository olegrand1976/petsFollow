package pharmacy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
)

// ExtractedMedication is one drug row parsed from a PDF chunk (Vetcompendium-style).
type ExtractedMedication struct {
	CNK                string   `json:"cnk"`
	Name               string   `json:"name"`
	Manufacturer       string   `json:"manufacturer"`
	ActiveSubstance    string   `json:"activeSubstance"`
	Strength           string   `json:"strength"`
	ATCCode            string   `json:"atcCode"`
	PharmaceuticalForm string   `json:"pharmaceuticalForm"`
	Route              string   `json:"route"`
	Species            []string `json:"species"`
	Posology           string   `json:"posology"`
	WithdrawalMeatDays *int     `json:"withdrawalMeatDays,omitempty"`
	WithdrawalMilkDays *int     `json:"withdrawalMilkDays,omitempty"`
	WithdrawalEggsDays *int     `json:"withdrawalEggsDays,omitempty"`
	PackSize           string   `json:"packSize"`
	PrescriptionOnly   bool     `json:"prescriptionOnly"`
	IsAntibiotic       bool     `json:"isAntibiotic"`
	SourcePage         *int     `json:"sourcePage,omitempty"`
}

// extractedMedicationDTO accepts loosely-typed Gemini fields before coercion.
type extractedMedicationDTO struct {
	CNK                any `json:"cnk"`
	Name               any `json:"name"`
	Manufacturer       any `json:"manufacturer"`
	ActiveSubstance    any `json:"activeSubstance"`
	Strength           any `json:"strength"`
	ATCCode            any `json:"atcCode"`
	PharmaceuticalForm any `json:"pharmaceuticalForm"`
	Route              any `json:"route"`
	Species            any `json:"species"`
	Posology           any `json:"posology"`
	WithdrawalMeatDays any `json:"withdrawalMeatDays"`
	WithdrawalMilkDays any `json:"withdrawalMilkDays"`
	WithdrawalEggsDays any `json:"withdrawalEggsDays"`
	PackSize           any `json:"packSize"`
	PrescriptionOnly   any `json:"prescriptionOnly"`
	IsAntibiotic       any `json:"isAntibiotic"`
	SourcePage         any `json:"sourcePage"`
}

const compendiumExtractSystem = `You extract Belgian veterinary medicine catalogue monographs from PDF pages (e.g. Vetcompendium).
Return ONLY valid JSON with this closed shape (no extra keys, no markdown, no comments):
{"medications":[{
  "cnk":"",
  "name":"VETORYL 20 mg caps",
  "manufacturer":"Dechra",
  "activeSubstance":"trilostane",
  "strength":"20 mg",
  "atcCode":"",
  "pharmaceuticalForm":"gélule",
  "route":"po",
  "species":["Ca"],
  "posology":"...",
  "withdrawalMeatDays":null,
  "withdrawalMilkDays":null,
  "withdrawalEggsDays":null,
  "packSize":"caps 30",
  "prescriptionOnly":true,
  "isAntibiotic":false,
  "sourcePage":1
}]}
Rules:
- Use ONLY the keys listed above. Never invent fields or values not supported by the PDF text.
- cnk: national product code when explicitly present; otherwise "". NEVER invent CNK (Vetcompendium PDFs usually have none). Prefer a string.
- name: product denomination including strength/form tokens when present (required); string
- manufacturer: lab name in parentheses when present (Huvepharma, Dechra, …)
- activeSubstance + strength: principe actif and concentration as written (strings, e.g. strength "20 mg" not 20)
- pharmaceuticalForm + route: e.g. gélule/comprimé/granulés + po/im (split glued tokens like "gélulepo" → form gélule, route po)
- species: JSON array of short codes/names only (["Ca","Su"]); never a single string
- posology: dosage block text (concise; do not invent doses)
- withdrawal*Days: JSON integer days OR null only (e.g. 28, null). If the PDF says "3,5 j", ceil to integer days (4). Never emit floats, strings, or ranges. Use null if ambiguous, multi-dose, or negative.
- packSize: packaging line (sac 1 kg, flacon 100 ml, caps 3 x 10, …)
- prescriptionOnly / isAntibiotic: JSON booleans true/false only (never "oui", "R/", 1, 0)
- isAntibiotic: true only when clearly antibiotic / antibactérien / ATC J01*
- prescriptionOnly: true when marked R/
- sourcePage: absolute PDF page number as JSON integer within the chunk range (not the printed catalogue footer)
- Skip headers, footers, TOC, section dividers (e.g. "CATALOGUE"), ads, copyright lines
- One JSON object per distinct commercial presentation`

// CompendiumExtractor extracts medication rows from PDF page chunks via Gemini.
type CompendiumExtractor struct {
	Gemini *gemini.Client
}

func (e *CompendiumExtractor) ExtractChunk(ctx context.Context, pdfChunk []byte, absStart, absEnd int) ([]ExtractedMedication, error) {
	if e == nil || e.Gemini == nil || !e.Gemini.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	user := fmt.Sprintf(
		"The attached PDF contains only absolute pages %d through %d (PDF page 1 = absolute page %d). Extract all medication monographs from these pages. Two-column catalogue layout: NAME (Lab), substance: strength, form+route (may be glued e.g. comprimépo), Posologie, species codes, withdrawal, pack, R/. Return JSON only with the closed schema. Do not invent CNK or any value absent from the PDF. withdrawal*Days must be JSON integers or null (ceil fractions like 3.5→4). Booleans true/false only. species must be a JSON array. Strings for name/cnk/strength/packSize. Set sourcePage to absolute integers in %d–%d.",
		absStart, absEnd, absStart, absStart, absEnd,
	)
	raw, err := e.Gemini.GenerateJSONWithMedia(ctx, compendiumExtractSystem, user, "application/pdf", pdfChunk, 0.1)
	if err != nil {
		return nil, err
	}
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		return nil, err
	}
	RemapSourcePages(meds, absStart, absEnd)
	return DedupExtractedMedications(meds), nil
}

// DedupExtractedMedications keeps the first row per CNK (case-insensitive).
// Rows without CNK are deduped by name; empty name+CNK rows are dropped.
func DedupExtractedMedications(meds []ExtractedMedication) []ExtractedMedication {
	if len(meds) == 0 {
		return meds
	}
	seenCNK := make(map[string]struct{}, len(meds))
	seenName := make(map[string]struct{}, len(meds))
	out := make([]ExtractedMedication, 0, len(meds))
	for _, m := range meds {
		cnk := strings.ToLower(strings.TrimSpace(m.CNK))
		if cnk != "" {
			if _, ok := seenCNK[cnk]; ok {
				continue
			}
			seenCNK[cnk] = struct{}{}
			out = append(out, m)
			continue
		}
		nameKey := strings.ToLower(strings.TrimSpace(m.Name))
		if nameKey == "" {
			continue
		}
		if _, ok := seenName[nameKey]; ok {
			continue
		}
		seenName[nameKey] = struct{}{}
		out = append(out, m)
	}
	return out
}

// ParseCompendiumExtractJSON normalizes Gemini JSON into medication rows.
// Tolerates loosely typed fields; skips individual monographs that cannot be coerced.
// If the payload lists monographs but none survive coercion, returns empty_after_coerce
// (avoids silently advancing extract progress on a wiped chunk).
func ParseCompendiumExtractJSON(raw string) ([]ExtractedMedication, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty_extract")
	}

	var medRaws []json.RawMessage
	var envelope struct {
		Medications []json.RawMessage `json:"medications"`
	}
	envErr := json.Unmarshal([]byte(raw), &envelope)
	if envErr == nil && envelope.Medications != nil {
		medRaws = envelope.Medications
	} else if arrErr := json.Unmarshal([]byte(raw), &medRaws); arrErr != nil {
		if envErr != nil {
			return nil, fmt.Errorf("invalid_extract_json: %w", envErr)
		}
		return nil, fmt.Errorf("invalid_extract_json: %w", arrErr)
	}

	out := make([]ExtractedMedication, 0, len(medRaws))
	skipped := 0
	for _, row := range medRaws {
		m, ok := parseMedicationRaw(row)
		if !ok {
			skipped++
			continue
		}
		out = append(out, m)
	}
	if skipped > 0 {
		log.Printf("compendium extract parse: skipped %d/%d monograph(s)", skipped, len(medRaws))
	}
	if len(medRaws) > 0 && len(out) == 0 {
		return nil, fmt.Errorf("empty_after_coerce")
	}
	return out, nil
}

// parseMedicationRaw coerces one monograph. Keeps partial rows (lab/page/form…)
// even without name/CNK so humans can complete them in review.
func parseMedicationRaw(row json.RawMessage) (ExtractedMedication, bool) {
	var dto extractedMedicationDTO
	if err := json.Unmarshal(row, &dto); err != nil {
		return ExtractedMedication{}, false
	}
	m := dto.toExtracted()
	if m.Name == "" && m.CNK == "" && !hasPartialSignal(m) {
		return ExtractedMedication{}, false
	}
	return m, true
}

func hasPartialSignal(m ExtractedMedication) bool {
	return m.Manufacturer != "" ||
		m.ActiveSubstance != "" ||
		m.Strength != "" ||
		m.PharmaceuticalForm != "" ||
		m.PackSize != "" ||
		m.Posology != "" ||
		m.ATCCode != "" ||
		m.SourcePage != nil ||
		len(m.Species) > 0
}

func (d extractedMedicationDTO) toExtracted() ExtractedMedication {
	return ExtractedMedication{
		CNK:                coerceString(d.CNK),
		Name:               coerceString(d.Name),
		Manufacturer:       coerceString(d.Manufacturer),
		ActiveSubstance:    coerceString(d.ActiveSubstance),
		Strength:           coerceString(d.Strength),
		ATCCode:            coerceString(d.ATCCode),
		PharmaceuticalForm: coerceString(d.PharmaceuticalForm),
		Route:              coerceString(d.Route),
		Species:            coerceSpecies(d.Species),
		Posology:           coerceString(d.Posology),
		PackSize:           coerceString(d.PackSize),
		PrescriptionOnly:   coerceBool(d.PrescriptionOnly),
		IsAntibiotic:       coerceBool(d.IsAntibiotic),
		SourcePage:         coerceSourcePage(d.SourcePage),
		WithdrawalMeatDays: coerceWithdrawalDays(d.WithdrawalMeatDays),
		WithdrawalMilkDays: coerceWithdrawalDays(d.WithdrawalMilkDays),
		WithdrawalEggsDays: coerceWithdrawalDays(d.WithdrawalEggsDays),
	}
}

// coerceString turns Gemini scalars into trimmed strings (numbers → decimal without trailing .0).
// Objects/arrays → "" (caller may soft-skip the row if name+cnk empty).
func coerceString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return ""
		}
		if x == math.Trunc(x) && x >= float64(math.MinInt64) && x <= float64(math.MaxInt64) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case json.Number:
		return strings.TrimSpace(x.String())
	case bool:
		return strconv.FormatBool(x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return ""
		}
		var s string
		if err := json.Unmarshal(b, &s); err == nil {
			return strings.TrimSpace(s)
		}
		return ""
	}
}

// coerceWithdrawalDays accepts int / float / numeric string (incl. "3,5") from Gemini.
// Fractional values use Ceil (food-chain safe); negative / unusable → nil.
func coerceWithdrawalDays(v any) *int {
	f, ok := toNonNegFloat(v)
	if !ok {
		return nil
	}
	n := int(math.Ceil(f))
	if n < 0 {
		return nil
	}
	return &n
}

func toNonNegFloat(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	var f float64
	switch x := v.(type) {
	case float64:
		f = x
	case json.Number:
		var err error
		f, err = x.Float64()
		if err != nil {
			return 0, false
		}
	case string:
		s := strings.TrimSpace(strings.ReplaceAll(x, ",", "."))
		if s == "" || strings.EqualFold(s, "null") {
			return 0, false
		}
		var err error
		f, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
	case bool:
		return 0, false
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return 0, false
		}
		if err := json.Unmarshal(b, &f); err != nil {
			return 0, false
		}
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 {
		return 0, false
	}
	return f, true
}

func coerceBool(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case json.Number:
		f, err := x.Float64()
		return err == nil && f != 0
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		switch s {
		case "true", "1", "yes", "oui", "y":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func coerceSourcePage(v any) *int {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) || x < 1 {
			return nil
		}
		n := int(math.Round(x))
		if n < 1 {
			return nil
		}
		return &n
	case json.Number:
		f, err := x.Float64()
		if err != nil || f < 1 {
			return nil
		}
		n := int(math.Round(f))
		return &n
	case string:
		s := strings.TrimSpace(x)
		if s == "" || strings.EqualFold(s, "null") {
			return nil
		}
		f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if err != nil || f < 1 {
			return nil
		}
		n := int(math.Round(f))
		return &n
	default:
		return nil
	}
}

func coerceSpecies(v any) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		return []string{s}
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			s := coerceString(item)
			if s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return nil
		}
		var arr []string
		if err := json.Unmarshal(b, &arr); err == nil {
			cleaned := make([]string, 0, len(arr))
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if s != "" {
					cleaned = append(cleaned, s)
				}
			}
			if len(cleaned) == 0 {
				return nil
			}
			return cleaned
		}
		s := coerceString(x)
		if s != "" {
			return []string{s}
		}
		return nil
	}
}

// ClassifyExtractedRow sets staging status after AI extract, AFMPS match, or human edit.
// Valid CNK+name → pending (needs human confirm) unless confirmed=true → ready.
func ClassifyExtractedRow(m ExtractedMedication, confirmed bool) (status, errCode, errMsg string) {
	if strings.TrimSpace(m.Name) == "" {
		return "error", "missing_name", "name required"
	}
	if strings.TrimSpace(m.CNK) == "" {
		return "error", "missing_cnk", "cnk required for national dictionary"
	}
	if confirmed {
		return "ready", "", ""
	}
	return "pending", "", ""
}

// ClassifyAfterMatch applies AFMPS suggestion rules before human review.
// AutoFill → cnk prefilled, pending. Candidates only → pending + missing_cnk cleared wait pick.
// No candidates → error missing_cnk.
func ClassifyAfterMatch(m ExtractedMedication, match CNKMatchResult) (out ExtractedMedication, status, errCode, errMsg string) {
	out = m
	if strings.TrimSpace(out.Name) == "" {
		return out, "error", "missing_name", "name required"
	}
	if strings.TrimSpace(out.CNK) != "" {
		st, code, msg := ClassifyExtractedRow(out, false)
		return out, st, code, msg
	}
	if match.AutoFill && match.SuggestedCNK != "" {
		out.CNK = match.SuggestedCNK
		return out, "pending", "", ""
	}
	if len(match.Candidates) > 0 {
		// Wait for human to pick a candidate CNK.
		return out, "pending", "cnk_unmatched", "pick CNK from AFMPS suggestions"
	}
	return out, "error", "missing_cnk", "cnk required for national dictionary"
}
