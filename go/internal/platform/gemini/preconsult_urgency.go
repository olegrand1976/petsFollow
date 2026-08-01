package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PreconsultUrgencyInput drives a one-shot urgency assessment from intake answers.
type PreconsultUrgencyInput struct {
	Locale         string
	PetName        string
	Species        string
	ChiefComplaint string
	Duration       string
	Behavior       string
	Appetite       string
	Thirst         string
	Elimination    string
	Urgency        string // client-declared: low|medium|high
	Comment        string
}

// PreconsultUrgencyAssessment is the informative AI judgment (does not override answers.urgency).
type PreconsultUrgencyAssessment struct {
	Urgency ClientTriageLevel `json:"urgency"`
	Summary string            `json:"summary"`
}

func buildPreconsultUrgencySystem(locale string) string {
	lang := clientExplainLangName(locale)
	return fmt.Sprintf(`Tu évalues l'urgence d'une pré-consultation vétérinaire (petsFollow) à partir du questionnaire client.
Tu ne poses PAS de diagnostic et tu ne prescrits JAMAIS.

Niveaux (urgency) :
- green : pas de signe d'urgence — prise en charge au créneau prévu
- orange : vigilance — le véto devrait prioriser / rappeler si besoin
- red : urgence potentielle — alerter le cabinet rapidement

RÈGLES :
- En cas de doute, monte d'un cran (green→orange, orange→red).
- Rouge si : intoxication, détresse respiratoire, convulsions, hémorragie abondante, collapse, ballonnement/GDV suspect, trauma grave, non-réponse.
- Tiens compte de l'urgence auto-déclarée (low/medium/high) sans l'ignorer, mais juge aussi les symptômes.
- summary : 1–2 phrases informatives pour le vétérinaire (langue : %s).

JSON uniquement :
{"urgency":"green|orange|red","summary":"…"}`, lang)
}

// AssessPreconsultUrgency runs a soft-fail-friendly Gemini judgment on preconsult answers.
func (c *Client) AssessPreconsultUrgency(ctx context.Context, in PreconsultUrgencyInput) (*PreconsultUrgencyAssessment, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	locale := strings.ToLower(strings.TrimSpace(in.Locale))
	if locale == "" {
		locale = "fr"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Animal: %s (%s)\n", strings.TrimSpace(in.PetName), strings.TrimSpace(in.Species)))
	b.WriteString(fmt.Sprintf("Motif: %s\n", strings.TrimSpace(in.ChiefComplaint)))
	b.WriteString(fmt.Sprintf("Durée: %s\n", strings.TrimSpace(in.Duration)))
	b.WriteString(fmt.Sprintf("Comportement: %s\n", strings.TrimSpace(in.Behavior)))
	b.WriteString(fmt.Sprintf("Appétit: %s\n", strings.TrimSpace(in.Appetite)))
	b.WriteString(fmt.Sprintf("Soif: %s\n", strings.TrimSpace(in.Thirst)))
	b.WriteString(fmt.Sprintf("Élimination: %s\n", strings.TrimSpace(in.Elimination)))
	b.WriteString(fmt.Sprintf("Urgence déclarée: %s\n", strings.TrimSpace(in.Urgency)))
	if cmt := strings.TrimSpace(in.Comment); cmt != "" {
		b.WriteString(fmt.Sprintf("Commentaire: %s\n", cmt))
	}
	raw, err := c.GenerateJSONLite(ctx, buildPreconsultUrgencySystem(locale), b.String(), 0.2)
	if err != nil {
		return nil, err
	}
	return ParsePreconsultUrgencyJSON(raw)
}

// ParsePreconsultUrgencyJSON normalizes Gemini JSON into PreconsultUrgencyAssessment.
func ParsePreconsultUrgencyJSON(raw string) (*PreconsultUrgencyAssessment, error) {
	trimmed := stripJSONFences(raw)
	var out PreconsultUrgencyAssessment
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		return nil, fmt.Errorf("gemini_preconsult_urgency_parse: %w", err)
	}
	out.Summary = strings.TrimSpace(out.Summary)
	rawLevel := strings.ToLower(strings.TrimSpace(string(out.Urgency)))
	level := ClientTriageLevel(rawLevel)
	switch level {
	case TriageGreen, TriageOrange, TriageRed:
		out.Urgency = level
	default:
		// Soft default; callers log assess failures separately when desired.
		out.Urgency = TriageOrange
	}
	if runes := []rune(out.Summary); len(runes) > 800 {
		out.Summary = string(runes[:800])
	}
	if out.Summary == "" {
		return nil, fmt.Errorf("gemini_preconsult_urgency_empty")
	}
	return &out, nil
}
