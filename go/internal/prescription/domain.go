package prescription

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound        = errors.New("prescription_not_found")
	ErrNotDraft        = errors.New("prescription_not_draft")
	ErrInvalidStatus   = errors.New("prescription_invalid_status")
	ErrEmptyMeds       = errors.New("prescription_empty_medications")
	ErrInvalidMeds     = errors.New("prescription_invalid_medications")
	ErrInvalidFormat   = errors.New("prescription_invalid_format")
	ErrPayloadTooLarge = errors.New("prescription_payload_too_large")
)

const (
	StatusDraft    = "draft"
	StatusSigned   = "signed"
	StatusSent     = "sent"
	StatusArchived = "archived"
)

const (
	FormatA4 = "A4"
	FormatA5 = "A5"
)

// Soft limits for draft payloads (anti-abuse).
const (
	MaxMedications      = 50
	MaxNotesRunes       = 4000
	MaxCareAdviceRunes  = 4000
	MaxMedFieldRunes    = 500
)

// Medication is one line in the prescriptions.medications JSONB array.
type Medication struct {
	Name             string  `json:"name"`
	Dosage           string  `json:"dosage"`
	Form             string  `json:"form"`
	Quantity         string  `json:"quantity"`
	Posology         string  `json:"posology"`
	WithdrawalPeriod string  `json:"withdrawal_period"`
	CNK              string  `json:"cnk,omitempty"`
	RefMedicationID  *string `json:"ref_medication_id,omitempty"`
}

func NormalizePaperFormat(f string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(f)) {
	case "", FormatA4:
		return FormatA4, nil
	case FormatA5:
		return FormatA5, nil
	default:
		return "", ErrInvalidFormat
	}
}

func NormalizeNotes(notes string) (string, error) {
	s := strings.TrimSpace(notes)
	if utf8.RuneCountInString(s) > MaxNotesRunes {
		return "", ErrPayloadTooLarge
	}
	return s, nil
}

func NormalizeCareAdvice(advice string) (string, error) {
	s := strings.TrimSpace(advice)
	if utf8.RuneCountInString(s) > MaxCareAdviceRunes {
		return "", ErrPayloadTooLarge
	}
	return s, nil
}

func fieldOK(s string) bool {
	return utf8.RuneCountInString(s) <= MaxMedFieldRunes
}

// NormalizeMedications accepts an empty list (draft consignes without medication).
// nil / null / [] → empty slice. Non-empty lines still require a name.
func NormalizeMedications(raw json.RawMessage) ([]Medication, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []Medication{}, nil
	}
	var list []Medication
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, ErrInvalidMeds
	}
	if len(list) == 0 {
		return []Medication{}, nil
	}
	if len(list) > MaxMedications {
		return nil, ErrPayloadTooLarge
	}
	out := make([]Medication, 0, len(list))
	for _, m := range list {
		m.Name = strings.TrimSpace(m.Name)
		m.Dosage = strings.TrimSpace(m.Dosage)
		m.Form = strings.TrimSpace(m.Form)
		m.Quantity = strings.TrimSpace(m.Quantity)
		m.Posology = strings.TrimSpace(m.Posology)
		m.WithdrawalPeriod = strings.TrimSpace(m.WithdrawalPeriod)
		m.CNK = strings.TrimSpace(m.CNK)
		if m.Name == "" {
			return nil, ErrInvalidMeds
		}
		if !fieldOK(m.Name) || !fieldOK(m.Dosage) || !fieldOK(m.Form) ||
			!fieldOK(m.Quantity) || !fieldOK(m.Posology) || !fieldOK(m.WithdrawalPeriod) || !fieldOK(m.CNK) {
			return nil, ErrPayloadTooLarge
		}
		if m.RefMedicationID != nil {
			id := strings.TrimSpace(*m.RefMedicationID)
			if id == "" {
				m.RefMedicationID = nil
			} else {
				m.RefMedicationID = &id
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func MedicationsJSON(list []Medication) (json.RawMessage, error) {
	if list == nil {
		list = []Medication{}
	}
	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}
