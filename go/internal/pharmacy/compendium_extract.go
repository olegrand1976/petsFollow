package pharmacy

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
)

// ExtractedMedication is one drug row parsed from a PDF chunk.
type ExtractedMedication struct {
	CNK                string `json:"cnk"`
	Name               string `json:"name"`
	ATCCode            string `json:"atcCode"`
	PharmaceuticalForm string `json:"pharmaceuticalForm"`
	PackSize           string `json:"packSize"`
	IsAntibiotic       bool   `json:"isAntibiotic"`
	SourcePage         *int   `json:"sourcePage,omitempty"`
}

type extractPayload struct {
	Medications []ExtractedMedication `json:"medications"`
}

const compendiumExtractSystem = `You extract Belgian veterinary / human medicine catalogue rows from PDF pages.
Return ONLY valid JSON with shape:
{"medications":[{"cnk":"...","name":"...","atcCode":"","pharmaceuticalForm":"","packSize":"","isAntibiotic":false,"sourcePage":1}]}
Rules:
- cnk: national product code (CNK) when present; empty string if unknown
- name: product denomination (required if a row is listed)
- atcCode, pharmaceuticalForm, packSize: empty string when absent
- isAntibiotic: true when clearly antibiotic / antibactérien / ATC J01*
- sourcePage: absolute page number in the original document when known
- Skip headers, footers, TOC, advertisements
- Do not invent CNK codes`

// CompendiumExtractor extracts medication rows from PDF page chunks via Gemini.
type CompendiumExtractor struct {
	Gemini *gemini.Client
}

func (e *CompendiumExtractor) ExtractChunk(ctx context.Context, pdfChunk []byte, absStart, absEnd int) ([]ExtractedMedication, error) {
	if e == nil || e.Gemini == nil || !e.Gemini.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	user := fmt.Sprintf(
		"The attached PDF may contain many pages. Extract medication products ONLY from absolute pages %d through %d inclusive. Ignore all other pages. Return JSON only.",
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
		// tolerate top-level array
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
		m.ATCCode = strings.TrimSpace(m.ATCCode)
		m.PharmaceuticalForm = strings.TrimSpace(m.PharmaceuticalForm)
		m.PackSize = strings.TrimSpace(m.PackSize)
		if m.Name == "" && m.CNK == "" {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// ClassifyExtractedRow sets ready vs error (missing CNK/name) for staging.
func ClassifyExtractedRow(m ExtractedMedication) (status, errCode, errMsg string) {
	if strings.TrimSpace(m.Name) == "" {
		return "error", "missing_name", "name required"
	}
	if strings.TrimSpace(m.CNK) == "" {
		return "error", "missing_cnk", "cnk required for national dictionary"
	}
	return "ready", "", ""
}

// FormatBoolish helps tests / CSV-like flags.
func FormatBoolish(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// ParseOptionalInt ignores empty.
func ParseOptionalInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}
