package gemini

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/prescription"
)

// ConsignesSuggestPromptInput drives the CR → consignes extraction prompt.
type ConsignesSuggestPromptInput struct {
	CountryCode string
	PetName     string
	PetSpecies  string
}

// ConsignesSuggestion is a non-persisted proposal for the consignes wizard.
type ConsignesSuggestion struct {
	Medications []prescription.Medication `json:"medications"`
	CareAdvice  string                    `json:"careAdvice"`
	Notes       string                    `json:"notes,omitempty"`
}

// BuildConsignesSuggestPrompt returns the system prompt for extracting client care instructions
// from a visit report. Output is consignes (not a legal prescription / ordonnance).
func BuildConsignesSuggestPrompt(in ConsignesSuggestPromptInput) string {
	pet := strings.TrimSpace(in.PetName)
	if pet == "" {
		pet = "l'animal"
	}
	species := strings.TrimSpace(in.PetSpecies)
	speciesHint := ""
	if species != "" {
		speciesHint = fmt.Sprintf(" (espèce: %s)", species)
	}
	cc := strings.ToUpper(strings.TrimSpace(in.CountryCode))
	if cc == "" {
		cc = "BE"
	}

	return fmt.Sprintf(`Tu aides un vétérinaire à préparer des CONSIGNES client (soins, médicaments à donner, posologie, conseils)
à partir d'un compte-rendu de consultation déjà rédigé.

Contexte animal: %s%s
Cabinet / pays (indicatif): %s

RÈGLES ABSOLUES:
- Extrais UNIQUEMENT ce qui est déjà présent dans le CR (sections type « Médication proposée » / « Proposed medication » / « Voorgestelde medicatie », « Plan / suivi » / « Plan / follow-up » / « Plan / opvolging », recommandations).
- N'invente JAMAIS de médicament, dosage, posologie, quantité ou délai d'attente absents du texte.
- Si une posologie n'est pas explicite, laisse le champ posology vide ("").
- careAdvice = consignes de soins et conseils destinés au propriétaire (texte clair, phrases courtes).
- notes = remarques internes optionnelles pour le véto (peut être "").
- Ce n'est PAS une ordonnance légale : ne rédige pas de mentions réglementaires / eIDAS / signature.
- Réponds UNIQUEMENT en JSON valide, sans markdown.

Schéma JSON:
{"medications":[{"name":"…","dosage":"…","form":"…","quantity":"…","posology":"…","withdrawal_period":"…"}],"careAdvice":"…","notes":"…"}

medications peut être [] si le CR ne mentionne aucun médicament.
Chaque name de médicament non vide si la ligne est présente.`, pet, speciesHint, cc)
}

type consignesSuggestRaw struct {
	Medications []prescription.Medication `json:"medications"`
	CareAdvice  string                    `json:"careAdvice"`
	Notes       string                    `json:"notes"`
}

// ParseConsignesSuggestJSON parses and soft-normalizes a Gemini consignes suggestion.
// Empty medications are allowed (unlike draft create).
func ParseConsignesSuggestJSON(raw string) (ConsignesSuggestion, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var parsed consignesSuggestRaw
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		// try extract first {...}
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start < 0 || end <= start {
			return ConsignesSuggestion{}, err
		}
		if err2 := json.Unmarshal([]byte(trimmed[start:end+1]), &parsed); err2 != nil {
			return ConsignesSuggestion{}, err
		}
	}

	out := ConsignesSuggestion{
		CareAdvice: strings.TrimSpace(parsed.CareAdvice),
		Notes:      strings.TrimSpace(parsed.Notes),
	}
	if utf8Over(out.CareAdvice, prescription.MaxCareAdviceRunes) {
		out.CareAdvice = truncateRunes(out.CareAdvice, prescription.MaxCareAdviceRunes)
	}
	if utf8Over(out.Notes, prescription.MaxNotesRunes) {
		out.Notes = truncateRunes(out.Notes, prescription.MaxNotesRunes)
	}

	meds := make([]prescription.Medication, 0, len(parsed.Medications))
	for _, m := range parsed.Medications {
		m.Name = truncateRunes(strings.TrimSpace(m.Name), prescription.MaxMedFieldRunes)
		m.Dosage = truncateRunes(strings.TrimSpace(m.Dosage), prescription.MaxMedFieldRunes)
		m.Form = truncateRunes(strings.TrimSpace(m.Form), prescription.MaxMedFieldRunes)
		m.Quantity = truncateRunes(strings.TrimSpace(m.Quantity), prescription.MaxMedFieldRunes)
		m.Posology = truncateRunes(strings.TrimSpace(m.Posology), prescription.MaxMedFieldRunes)
		m.WithdrawalPeriod = truncateRunes(strings.TrimSpace(m.WithdrawalPeriod), prescription.MaxMedFieldRunes)
		m.CNK = truncateRunes(strings.TrimSpace(m.CNK), prescription.MaxMedFieldRunes)
		if m.Name == "" {
			continue
		}
		if len(meds) >= prescription.MaxMedications {
			break
		}
		meds = append(meds, m)
	}
	out.Medications = meds
	return out, nil
}

func utf8Over(s string, max int) bool {
	n := 0
	for range s {
		n++
		if n > max {
			return true
		}
	}
	return false
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}
