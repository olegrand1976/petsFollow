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
		Apply       bool            `json:"apply"`
		Changelog   string          `json:"changelog"`
		ContentJSON json.RawMessage `json:"content"`
	} `json:"vetChanges"`
	CoachChanges struct {
		Apply       bool            `json:"apply"`
		Changelog   string          `json:"changelog"`
		ContentJSON json.RawMessage `json:"content"`
	} `json:"coachChanges"`
	Rationale  []string `json:"rationale"`
	NoOpReason string   `json:"noOpReason"`
}

type vetPromptParts struct {
	BasePersona  string
	ProductFacts string
	Difficulty   string
	Tools        string
}

func parseVetPromptParts(contentJSON json.RawMessage, interestLevel string) vetPromptParts {
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
	return vetPromptParts{
		BasePersona:  c.BasePersona,
		ProductFacts: c.ProductFacts,
		Difficulty:   "Niveau d'intérêt / difficulté pour CET appel: " + interestLevel + ". " + diff,
		Tools:        c.Tools,
	}
}

// SpinPedagogyBlock is shared by Live, turn JSON, and text-stream prompts.
func SpinPedagogyBlock() string {
	return `Pédagogie SPIN Selling (entraînement commercial):
Tu joues l'ACHETEUR (vétérinaire), pas le coach. Ne dis jamais « utilise SPIN » à voix haute.
Challenge le commercial pour qu'il fasse émerger Situation, Problem, Implication, Need-payoff via de vraies objections métier (temps, abo, boîtier, adoption équipe, ROI cabinet).
Ne pitch pas petsFollow à sa place. Phrases courtes, ton parlé, sans Markdown.`
}

func BuildVetSystemPrompt(contentJSON json.RawMessage, interestLevel string) string {
	p := parseVetPromptParts(contentJSON, interestLevel)
	return strings.Join([]string{
		p.BasePersona,
		"Faits produit: " + p.ProductFacts,
		p.Difficulty,
		SpinPedagogyBlock(),
		p.Tools,
		`À chaque tour, réponds UNIQUEMENT en JSON:
{"reply":"texte oral du véto","action":"continue|book_appointment|hang_up_not_interested","appointmentSlot":"optionnel","reason":"optionnel"}`,
	}, "\n\n")
}

// BuildVetStreamPrompt: oral reply first, then a final JSON action line (not for TTS/UI).
func BuildVetStreamPrompt(contentJSON json.RawMessage, interestLevel string) string {
	p := parseVetPromptParts(contentJSON, interestLevel)
	turnRules := `- Français oral, 1 à 2 phrases MAX par tour, sans Markdown ni listes.
- Écoute d'abord le fond ; challenge SPIN via questions/objections, sans méta-coacher.
- Laisse le commercial finir son idée avant de répondre.`
	if interestLevel == "hostile" {
		turnRules = `- Niveau hostile: réponses très courtes (1 phrase), ton impatient, objections sèches.
- Tu peux couper / relancer vite si le pitch est faible ; raccroche vite via la ligne JSON action si besoin.
- Challenge SPIN sans méta-coacher, sans Markdown.`
	}
	return strings.Join([]string{
		p.BasePersona,
		"Faits produit: " + p.ProductFacts,
		p.Difficulty,
		SpinPedagogyBlock(),
		p.Tools,
		`Tu es AU TÉLÉPHONE (mode texte stream) avec un commercial petsFollow.
Règles:
` + turnRules + `
- Réponds d'abord UNIQUEMENT le texte oral.
- Ensuite, sur une NOUVELLE ligne et rien d'autre, un JSON:
{"action":"continue|book_appointment|hang_up_not_interested","appointmentSlot":"optionnel","reason":"optionnel"}
- N'inclus jamais ce JSON dans le texte oral.`,
	}, "\n\n")
}

// DisplayableStreamText hides a trailing JSON action line while tokens stream in.
func DisplayableStreamText(buf string) string {
	if buf == "" {
		return ""
	}
	i := strings.LastIndex(buf, "\n")
	if i < 0 {
		if strings.HasPrefix(strings.TrimSpace(buf), "{") {
			return ""
		}
		return buf
	}
	last := strings.TrimSpace(buf[i+1:])
	if strings.HasPrefix(last, "{") {
		return strings.TrimRight(buf[:i], "\r")
	}
	return buf
}

// ParseStreamVetReply splits oral text from a trailing action JSON line.
func ParseStreamVetReply(full string) VetTurnResult {
	full = strings.TrimSpace(full)
	if full == "" {
		return VetTurnResult{Reply: "Oui ?", Action: "continue"}
	}
	lines := strings.Split(full, "\n")
	var actionLine string
	cut := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "{") {
			var probe map[string]any
			if json.Unmarshal([]byte(t), &probe) == nil {
				if _, ok := probe["action"]; ok {
					actionLine = t
					cut = i
					break
				}
			}
		}
		break
	}
	reply := strings.TrimSpace(strings.Join(lines[:cut], "\n"))
	res := VetTurnResult{Reply: reply, Action: "continue"}
	if actionLine != "" {
		var meta struct {
			Action          string `json:"action"`
			AppointmentSlot string `json:"appointmentSlot"`
			Reason          string `json:"reason"`
		}
		if json.Unmarshal([]byte(actionLine), &meta) == nil {
			if meta.Action != "" {
				res.Action = meta.Action
			}
			res.AppointmentSlot = meta.AppointmentSlot
			res.Reason = meta.Reason
		}
	}
	if res.Reply == "" {
		res.Reply = "Oui ?"
	}
	return res
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

func (c *Client) VetTurnStream(
	ctx context.Context,
	systemPrompt string,
	history []map[string]string,
	userLine string,
	onDisplayDelta func(displayable string) error,
) (*VetTurnResult, error) {
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
	b.WriteString("\nRéponds: texte oral puis une ligne JSON action.")

	var emitted string
	var buf strings.Builder
	full, err := c.GenerateTextStream(ctx, systemPrompt, b.String(), 0.8, func(chunk string) error {
		buf.WriteString(chunk)
		disp := DisplayableStreamText(buf.String())
		delta, ok := utf8SafeSuffix(emitted, disp)
		if !ok || delta == "" {
			return nil
		}
		emitted = disp
		if onDisplayDelta != nil {
			return onDisplayDelta(delta)
		}
		return nil
	})
	if err != nil && ctx.Err() == nil {
		return nil, err
	}
	if full == "" {
		full = buf.String()
	}
	res := ParseStreamVetReply(full)
	if ctx.Err() != nil {
		// Interrupted: keep partial oral reply, never book/hangup.
		res.Action = "continue"
		res.AppointmentSlot = ""
		if disp := DisplayableStreamText(full); disp != "" {
			res.Reply = disp
		}
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
	rubric := cfg.Rubric
	if len(rubric) == 0 {
		rubric = []string{"opener", "listening", "objections", "offerClarity", "cta", "spin"}
	}
	dimsHint := strings.Join(rubric, `":0-10,"`)
	user := fmt.Sprintf(`Analyse cet appel d'entraînement petsFollow.
interest_level=%s outcome=%s
hints_script=%s
rules=%s
rubric=%v
transcript=%s

Retourne JSON:
{"score":0-10,"dimensions":{"%s":0-10},"strengths":[],"improvements":[],"coachingTips":[],"scriptCoverage":[]}
Évalue aussi si le commercial a fait émerger Situation/Problem/Implication/Need-payoff (dimension spin si présente).`,
		interestLevel, outcome, scriptHints, cfg.Rules, rubric, string(transcript), dimsHint)

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

	raw, err := c.GenerateJSONLite(ctx, system, user, 0.2)
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

// utf8SafeSuffix returns next[len(prev):] only when prev is a prefix of next (byte-safe for UTF-8).
func utf8SafeSuffix(prev, next string) (string, bool) {
	if len(next) <= len(prev) {
		return "", false
	}
	if !strings.HasPrefix(next, prev) {
		// Displayable text shrank (ex. ligne JSON) : pas de nouveau delta.
		return "", false
	}
	return next[len(prev):], true
}
