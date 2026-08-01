package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ClientTriageLevel is the urgency band for owner triage.
type ClientTriageLevel string

const (
	TriageGreen  ClientTriageLevel = "green"
	TriageOrange ClientTriageLevel = "orange"
	TriageRed    ClientTriageLevel = "red"
)

// ClientTriageTurn is one assistant reply from the triage model.
type ClientTriageTurn struct {
	Reply             string            `json:"reply"`
	Level             ClientTriageLevel `json:"level"`
	WatchSigns        []string          `json:"watchSigns"`
	RecommendedAction string            `json:"recommendedAction"`
}

// ClientTriageMessage is prior conversation context.
type ClientTriageMessage struct {
	Role string `json:"role"` // user | assistant
	Body string `json:"body"`
}

// ClientTriageInput drives the conversational triage prompt.
type ClientTriageInput struct {
	Locale   string
	PetName  string
	Species  string
	History  []ClientTriageMessage
	UserText string
}

func buildClientTriageSystem(locale string) string {
	lang := clientExplainLangName(locale)
	return fmt.Sprintf(`Tu es un assistant de TRIAGE pré-évaluation pour propriétaires d'animaux (petsFollow), 24/7.
Tu évalues le degré d'urgence — tu ne poses PAS de diagnostic et tu ne prescrits JAMAIS.

Niveaux :
- green : surveillance à domicile, signes à guetter
- orange : RDV recommandé sous 48h chez le vétérinaire
- red : urgence potentielle — contacter immédiatement le cabinet / service d'urgence local

RÈGLES :
- En cas de doute, monte d'un cran (green→orange, orange→red).
- Rouge si : intoxication (chocolat, xylitol, raisin…), détresse respiratoire, convulsions, hémorragie abondante, collapse, ballonnement/GDV suspect, trauma grave, non-réponse.
- Jamais de médicament, dosages, ni « ce n'est rien ».
- Langue : %s.
- reply : 2–5 phrases claires + prochaines étapes.
- watchSigns : liste courte (surtout green/orange).
- recommendedAction : une phrase d'action.

JSON uniquement :
{"reply":"…","level":"green|orange|red","watchSigns":["…"],"recommendedAction":"…"}`, lang)
}

// TriageClientMessage runs one conversational triage turn.
func (c *Client) TriageClientMessage(ctx context.Context, in ClientTriageInput) (*ClientTriageTurn, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	locale := strings.ToLower(strings.TrimSpace(in.Locale))
	if locale == "" {
		locale = "fr"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Contexte animal: %s (%s)\n\nHistorique:\n",
		strings.TrimSpace(in.PetName), strings.TrimSpace(in.Species)))
	for _, m := range in.History {
		role := strings.TrimSpace(m.Role)
		body := strings.TrimSpace(m.Body)
		if role == "" || body == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("[%s] %s\n", role, body))
	}
	b.WriteString(fmt.Sprintf("\n[user] %s\n", strings.TrimSpace(in.UserText)))
	raw, err := c.GenerateJSONLite(ctx, buildClientTriageSystem(locale), b.String(), 0.2)
	if err != nil {
		return nil, err
	}
	return ParseClientTriageJSON(raw)
}

// ParseClientTriageJSON normalizes Gemini JSON into ClientTriageTurn.
func ParseClientTriageJSON(raw string) (*ClientTriageTurn, error) {
	trimmed := stripJSONFences(raw)
	var out ClientTriageTurn
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		return nil, fmt.Errorf("gemini_triage_parse: %w", err)
	}
	out.Reply = strings.TrimSpace(out.Reply)
	out.RecommendedAction = strings.TrimSpace(out.RecommendedAction)
	level := ClientTriageLevel(strings.ToLower(strings.TrimSpace(string(out.Level))))
	switch level {
	case TriageGreen, TriageOrange, TriageRed:
		out.Level = level
	default:
		out.Level = TriageOrange
	}
	signs := make([]string, 0, len(out.WatchSigns))
	for _, s := range out.WatchSigns {
		s = strings.TrimSpace(s)
		if s != "" {
			signs = append(signs, s)
		}
	}
	out.WatchSigns = signs
	if out.Reply == "" {
		return nil, fmt.Errorf("gemini_triage_empty")
	}
	return &out, nil
}
