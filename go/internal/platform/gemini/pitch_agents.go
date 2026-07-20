package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type VetTurnResult struct {
	Reply           string `json:"reply"`
	Action          string `json:"action"` // continue | book_appointment | hang_up_not_interested
	AppointmentSlot string `json:"appointmentSlot,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

type CoachResult struct {
	Score          float64            `json:"score"`
	Dimensions     map[string]float64 `json:"dimensions"`
	Strengths      []string           `json:"strengths"`
	Improvements   []string           `json:"improvements"`
	CoachingTips   []string           `json:"coachingTips"`
	ScriptCoverage []string           `json:"scriptCoverage"`
}

type AnalyzerResult struct {
	VetChanges struct {
		Apply        bool            `json:"apply"`
		Changelog    string          `json:"changelog"`
		ContentJSON  json.RawMessage `json:"content"`
	} `json:"vetChanges"`
	CoachChanges struct {
		Apply       bool            `json:"apply"`
		Changelog   string          `json:"changelog"`
		ContentJSON json.RawMessage `json:"content"`
	} `json:"coachChanges"`
	Rationale  []string `json:"rationale"`
	NoOpReason string   `json:"noOpReason"`
}

func BuildVetSystemPrompt(contentJSON json.RawMessage, interestLevel string) string {
	var c struct {
		BasePersona  string            `json:"basePersona"`
		ProductFacts string            `json:"productFacts"`
		Difficulty   map[string]string `json:"difficulty"`
		Tools        string            `json:"tools"`
	}
	_ = json.Unmarshal(contentJSON, &c)
	diff := c.Difficulty[interestLevel]
	if diff == "" {
		diff = c.Difficulty["neutre"]
	}
	return strings.Join([]string{
		c.BasePersona,
		"Faits produit: " + c.ProductFacts,
		"Niveau d'intérêt / difficulté pour CET appel: " + interestLevel + ". " + diff,
		c.Tools,
		`À chaque tour, réponds UNIQUEMENT en JSON:
{"reply":"texte oral du véto","action":"continue|book_appointment|hang_up_not_interested","appointmentSlot":"optionnel","reason":"optionnel"}`,
	}, "\n\n")
}

func (c *Client) VetTurn(ctx context.Context, systemPrompt string, history []map[string]string, userLine string) (*VetTurnResult, error) {
	var b strings.Builder
	b.WriteString("Historique de l'appel:\n")
	for _, h := range history {
		b.WriteString(h["role"])
		b.WriteString(": ")
		b.WriteString(h["text"])
		b.WriteString("\n")
	}
	b.WriteString("\nDernière réplique du commercial: ")
	b.WriteString(userLine)
	b.WriteString("\nRéponds en JSON.")

	raw, err := c.GenerateJSON(ctx, systemPrompt, b.String(), 0.8)
	if err != nil {
		return nil, err
	}
	raw = stripFences(raw)
	var res VetTurnResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		// Fallback: treat whole text as reply
		return &VetTurnResult{Reply: raw, Action: "continue"}, nil
	}
	if res.Reply == "" {
		res.Reply = "Oui ?"
	}
	if res.Action == "" {
		res.Action = "continue"
	}
	return &res, nil
}

func (c *Client) CoachCall(ctx context.Context, coachContent json.RawMessage, scriptHints string, interestLevel, outcome string, transcript json.RawMessage) (*CoachResult, error) {
	var cfg struct {
		System string   `json:"system"`
		Rubric []string `json:"rubric"`
		Rules  string   `json:"rules"`
	}
	_ = json.Unmarshal(coachContent, &cfg)
	system := cfg.System
	if system == "" {
		system = "Tu es un coach commercial. Réponds en JSON."
	}
	user := fmt.Sprintf(`Analyse cet appel d'entraînement petsFollow.
interest_level=%s outcome=%s
hints_script=%s
rules=%s
rubric=%v
transcript=%s

Retourne JSON:
{"score":0-10,"dimensions":{"opener":0-10,"listening":0-10,"objections":0-10,"offerClarity":0-10,"cta":0-10},"strengths":[],"improvements":[],"coachingTips":[],"scriptCoverage":[]}`,
		interestLevel, outcome, scriptHints, cfg.Rules, cfg.Rubric, string(transcript))

	raw, err := c.GenerateJSON(ctx, system, user, 0.3)
	if err != nil {
		return nil, err
	}
	raw = stripFences(raw)
	var res CoachResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, fmt.Errorf("coach_parse: %w", err)
	}
	return &res, nil
}

func (c *Client) AnalyzeFeedback(ctx context.Context, vetContent, coachContent json.RawMessage, feedbackSummary string) (*AnalyzerResult, error) {
	system := `Tu es un analyseur qualité pour agents d'entraînement commercial petsFollow.
Tu peux proposer des versions améliorées des prompts véto (vet_live) et coach.
Ne lève JAMAIS les garde-fous produit (pas de boîtier, pas de % sur TTC, inscription ≠ revenu).
Si trop peu de signal, noOpReason non vide et apply=false.
Réponds UNIQUEMENT en JSON.`
	user := fmt.Sprintf(`Prompts courants:
vet_live=%s
coach=%s

Retours commerciaux agrégés:
%s

JSON attendu:
{"vetChanges":{"apply":false,"changelog":"","content":{}},"coachChanges":{"apply":false,"changelog":"","content":{}},"rationale":[],"noOpReason":""}
Si apply=true, content doit être l'objet content_json COMPLET (même schéma que l'actuel), pas un patch partiel.`,
		string(vetContent), string(coachContent), feedbackSummary)

	raw, err := c.GenerateJSON(ctx, system, user, 0.2)
	if err != nil {
		return nil, err
	}
	raw = stripFences(raw)
	var res AnalyzerResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, fmt.Errorf("analyzer_parse: %w", err)
	}
	return &res, nil
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
