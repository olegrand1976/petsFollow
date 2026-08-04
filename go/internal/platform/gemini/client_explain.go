package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ClientExplainCard is one educational card for a pet owner.
type ClientExplainCard struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Kind  string `json:"kind"` // term | medication | general
}

// ClientExplainResult is the structured vulgarization of a finalized CR.
type ClientExplainResult struct {
	Disclaimer string              `json:"disclaimer"`
	Cards      []ClientExplainCard `json:"cards"`
}

// ClientExplainInput drives the owner-facing CR explain prompt.
type ClientExplainInput struct {
	Locale    string // fr|nl|en|es|et|it|uk|ru
	PetName   string
	Species   string
	BodyTexts []string
}

func clientExplainLangName(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "nl":
		return "néerlandais"
	case "en":
		return "anglais"
	case "es":
		return "espagnol"
	case "et":
		return "estonien"
	case "it":
		return "italien"
	case "uk":
		return "ukrainien"
	case "ru":
		return "russe"
	default:
		return "français"
	}
}

func buildClientExplainSystem(locale string) string {
	lang := clientExplainLangName(locale)
	return fmt.Sprintf(`Tu es un assistant pédagogique pour propriétaires d'animaux sur petsFollow.
Tu vulgarises un compte-rendu vétérinaire FINALISÉ en langage simple et rassurant.

RÈGLES ABSOLUES :
- Ne modifie JAMAIS le diagnostic, ne le contredis pas, n'invente aucun fait médical absent du CR.
- Ne propose aucun traitement, dosage, ni arrêt de médicament. Renvoie TOUJOURS aux instructions exactes du vétérinaire.
- Ton rassurant, non alarmiste. Si un terme est grave, explique calmement et invite à suivre le véto.
- Langue de sortie : %s uniquement.
- 2 à 6 cartes max. kind ∈ term|medication|general.
- disclaimer : rappel que ce n'est pas un avis médical et qu'il faut suivre le vétérinaire.

Réponds UNIQUEMENT en JSON valide :
{"disclaimer":"…","cards":[{"title":"…","body":"…","kind":"term"}]}`, lang)
}

// ExplainVisitReportForClient produces owner-facing educational cards from final CR text.
func (c *Client) ExplainVisitReportForClient(ctx context.Context, in ClientExplainInput) (*ClientExplainResult, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	locale := strings.ToLower(strings.TrimSpace(in.Locale))
	if locale == "" {
		locale = "fr"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Animal: %s (%s)\n\nCompte-rendu vétérinaire (source, ne pas altérer):\n",
		strings.TrimSpace(in.PetName), strings.TrimSpace(in.Species)))
	for i, t := range in.BodyTexts {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("--- CR %d ---\n%s\n", i+1, t))
	}
	raw, err := c.GenerateJSONLite(ctx, buildClientExplainSystem(locale), b.String(), 0.2)
	if err != nil {
		return nil, err
	}
	return ParseClientExplainJSON(raw)
}

// ParseClientExplainJSON normalizes Gemini JSON into ClientExplainResult.
func ParseClientExplainJSON(raw string) (*ClientExplainResult, error) {
	trimmed := stripJSONFences(raw)
	var out ClientExplainResult
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		return nil, fmt.Errorf("gemini_explain_parse: %w", err)
	}
	if strings.TrimSpace(out.Disclaimer) == "" {
		out.Disclaimer = "Ceci n'est pas un avis médical. Suivez toujours les consignes de votre vétérinaire."
	}
	norm := make([]ClientExplainCard, 0, len(out.Cards))
	for _, card := range out.Cards {
		title := strings.TrimSpace(card.Title)
		body := strings.TrimSpace(card.Body)
		if title == "" || body == "" {
			continue
		}
		kind := strings.ToLower(strings.TrimSpace(card.Kind))
		switch kind {
		case "term", "medication", "general":
		default:
			kind = "general"
		}
		norm = append(norm, ClientExplainCard{Title: title, Body: body, Kind: kind})
	}
	out.Cards = norm
	if len(out.Cards) == 0 {
		return nil, fmt.Errorf("gemini_explain_empty")
	}
	return &out, nil
}

func stripJSONFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}
