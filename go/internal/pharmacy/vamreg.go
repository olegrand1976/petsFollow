package pharmacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDAFNotFound                 = errors.New("daf_not_found")
	ErrDAFNotDraft                 = errors.New("daf_not_draft")
	ErrDAFNotFinalized             = errors.New("daf_not_finalized")
	ErrDAFEmpty                    = errors.New("daf_empty")
	ErrDAFAMMRequired              = errors.New("daf_amm_required")
	ErrDAFVAMRegIncomplete         = errors.New("daf_vamreg_incomplete")
	ErrDAFAlreadyHasPDF            = errors.New("daf_pdf_immutable")
	ErrDAFNotAntibiotic            = errors.New("daf_not_antibiotic")
	ErrDAFVAMRegAlreadySent        = errors.New("daf_vamreg_already_sent")
	ErrDAFVAMRegInFlight           = errors.New("daf_vamreg_in_flight")
	ErrFoodChainWithdrawalRequired = errors.New("food_chain_withdrawal_required")
	ErrFoodChainBannedMedication   = errors.New("food_chain_banned_medication")
	// ErrDAFSpeciesNotApplicable — l'espèce du patient n'est pas productrice de denrées
	// alimentaires dans le pays de la clinique (ou l'individu en est exclu) : pas de DAF.
	ErrDAFSpeciesNotApplicable = errors.New("daf_species_not_applicable")
)

// FormatDAFNumber returns DAF-YYYY-NNNNNN.
func FormatDAFNumber(year int, number int64) string {
	return fmt.Sprintf("DAF-%d-%06d", year, number)
}

// VamregPosologyMaxLen caps posology payload size for VAMReg declare.
const VamregPosologyMaxLen = 500

// VamregPayload fields required before finalize on antibiotic lines.
type VamregPayload struct {
	Species      string `json:"species"`
	Indication   string `json:"indication"`
	DurationDays int    `json:"durationDays"`
	Posology     string `json:"posology"`
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
	posology := strings.TrimSpace(p.Posology)
	if strings.TrimSpace(p.Species) == "" ||
		strings.TrimSpace(p.Indication) == "" ||
		p.DurationDays <= 0 ||
		posology == "" ||
		len(posology) > VamregPosologyMaxLen {
		return ErrDAFVAMRegIncomplete
	}
	return nil
}
