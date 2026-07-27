package pharmacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDAFNotFound           = errors.New("daf_not_found")
	ErrDAFNotDraft           = errors.New("daf_not_draft")
	ErrDAFNotFinalized       = errors.New("daf_not_finalized")
	ErrDAFEmpty              = errors.New("daf_empty")
	ErrDAFAMMRequired        = errors.New("daf_amm_required")
	ErrDAFVAMRegIncomplete   = errors.New("daf_vamreg_incomplete")
	ErrDAFAlreadyHasPDF      = errors.New("daf_pdf_immutable")
)

// FormatDAFNumber returns DAF-YYYY-NNNNNN.
func FormatDAFNumber(year int, number int64) string {
	return fmt.Sprintf("DAF-%d-%06d", year, number)
}

// VamregPayload fields required before finalize on antibiotic lines.
type VamregPayload struct {
	Species      string `json:"species"`
	Indication   string `json:"indication"`
	DurationDays int    `json:"durationDays"`
}

// ValidateVamregPayload returns ErrDAFVAMRegIncomplete when required fields are missing.
func ValidateVamregPayload(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return ErrDAFVAMRegIncomplete
	}
	var p VamregPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return ErrDAFVAMRegIncomplete
	}
	if strings.TrimSpace(p.Species) == "" || strings.TrimSpace(p.Indication) == "" || p.DurationDays <= 0 {
		return ErrDAFVAMRegIncomplete
	}
	return nil
}
