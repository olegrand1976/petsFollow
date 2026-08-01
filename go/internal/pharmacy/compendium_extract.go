package pharmacy

import (
	"context"
	"encoding/json"
	"fmt"
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

type extractPayload struct {
	Medications []ExtractedMedication `json:"medications"`
}

const compendiumExtractSystem = `You extract Belgian veterinary medicine catalogue monographs from PDF pages (e.g. Vetcompendium).
Return ONLY valid JSON with shape:
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
- cnk: national product code when explicitly present; otherwise empty string. NEVER invent CNK codes (Vetcompendium PDFs usually have none).
- name: product denomination including strength/form tokens when present (required)
- manufacturer: lab name in parentheses when present (Huvepharma, Dechra, …)
- activeSubstance + strength: principle actif and concentration
- pharmaceuticalForm + route: e.g. gélule/comprimé/granulés + po/im (split glued tokens like "gélulepo" → form gélule, route po)
- species: target species codes/names (Ca, Su, Bo, poule, …)
- posology: dosage block text (may be multi-line; keep concise)
- withdrawal*Days: integers when clearly stated (Viande/Lait/Œufs); omit/null if ambiguous or multi-dose
- packSize: packaging line (sac 1 kg, flacon 100 ml, caps 3 x 10, …)
- prescriptionOnly: true when marked R/
- isAntibiotic: true when clearly antibiotic / antibactérien / ATC J01*
- sourcePage: absolute page number when known
- Skip headers, footers, TOC, ads, copyright lines
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
		"The attached PDF may contain many pages. Extract medication monographs ONLY from absolute pages %d through %d inclusive. Ignore all other pages. Return JSON only. Do not invent CNK.",
		absStart, absEnd,
	)
	raw, err := e.Gemini.GenerateJSONWithMedia(ctx, compendiumExtractSystem, user, "application/pdf", pdfChunk, 0.1)
	if err != nil {
		return nil, err
	}
	return ParseCompendiumExtractJSON(raw)
}

// ParseCompendiumExtractJSON normalizes Gemini JSON into medication rows.
func ParseCompendiumExtractJSON(raw string) ([]ExtractedMedication, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty_extract")
	}
	var payload extractPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		var arr []ExtractedMedication
		if err2 := json.Unmarshal([]byte(raw), &arr); err2 != nil {
			return nil, fmt.Errorf("invalid_extract_json: %w", err)
		}
		payload.Medications = arr
	}
	out := make([]ExtractedMedication, 0, len(payload.Medications))
	for _, m := range payload.Medications {
		m.CNK = strings.TrimSpace(m.CNK)
		m.Name = strings.TrimSpace(m.Name)
		m.Manufacturer = strings.TrimSpace(m.Manufacturer)
		m.ActiveSubstance = strings.TrimSpace(m.ActiveSubstance)
		m.Strength = strings.TrimSpace(m.Strength)
		m.ATCCode = strings.TrimSpace(m.ATCCode)
		m.PharmaceuticalForm = strings.TrimSpace(m.PharmaceuticalForm)
		m.Route = strings.TrimSpace(m.Route)
		m.Posology = strings.TrimSpace(m.Posology)
		m.PackSize = strings.TrimSpace(m.PackSize)
		if len(m.Species) > 0 {
			cleaned := make([]string, 0, len(m.Species))
			for _, sp := range m.Species {
				sp = strings.TrimSpace(sp)
				if sp != "" {
					cleaned = append(cleaned, sp)
				}
			}
			m.Species = cleaned
		}
		if m.Name == "" && m.CNK == "" {
			continue
		}
		out = append(out, m)
	}
	return out, nil
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
