package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ProductDigestCommit is one git commit fed into the functional digest LLM.
type ProductDigestCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
	Body    string `json:"body,omitempty"`
	Author  string `json:"author,omitempty"`
}

// ProductDigestSummary is the multilingual functional summary produced by Gemini.
type ProductDigestSummary struct {
	Empty    bool                       `json:"empty"`
	Headline map[string]string          `json:"headline"` // fr/en/nl/es
	Body     map[string]string          `json:"body"`     // plain-text bullets per locale
	Reason   string                     `json:"reason,omitempty"`
}

const productDigestSystem = `Tu es le rédacteur produit de petsFollow (suivi cardiaque vétérinaire, faces Pro Nuxt + app pets Flutter + API).
À partir d'une liste de commits git, produis une synthèse FONCTIONNELLE pour l'équipe interne (admin, commerciaux, commercial managers).
Règles strictes :
- Langage métier, jamais technique (pas de noms de fichiers, packages, migrations, endpoints, refactor, CI, deps, typos purement internes).
- Décris ce que l'utilisateur / le cabinet / le commercial gagne ou voit changer.
- Ignore les commits purement techniques (chore, ci, test, docs internes, bump deps, fix lint) sauf s'ils ont un impact produit visible.
- Si aucun changement fonctionnel : empty=true et reason court.
- Sinon empty=false : 3 à 8 puces max, phrases courtes.
- Fournis headline + body pour fr, en, nl, es. body = texte plain avec puces "• " (une par ligne).
Réponds UNIQUEMENT en JSON valide.`

// SummarizeProductDigest turns git commits into a functional multilingual digest.
func (c *Client) SummarizeProductDigest(ctx context.Context, commits []ProductDigestCommit) (*ProductDigestSummary, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	payload, err := json.Marshal(commits)
	if err != nil {
		return nil, err
	}
	user := "Commits du jour (JSON) :\n" + string(payload) + `

Schéma de sortie JSON :
{
  "empty": false,
  "reason": "",
  "headline": {"fr":"…","en":"…","nl":"…","es":"…"},
  "body": {"fr":"• …\n• …","en":"…","nl":"…","es":"…"}
}`
	raw, err := c.GenerateJSONLite(ctx, productDigestSystem, user, 0.3)
	if err != nil {
		return nil, err
	}
	var out ProductDigestSummary
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		// Some models wrap JSON in markdown fences.
		trimmed := strings.TrimSpace(raw)
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		if err2 := json.Unmarshal([]byte(trimmed), &out); err2 != nil {
			return nil, fmt.Errorf("gemini_digest_parse: %w (raw=%s)", err, truncate(raw, 200))
		}
	}
	if out.Headline == nil {
		out.Headline = map[string]string{}
	}
	if out.Body == nil {
		out.Body = map[string]string{}
	}
	return &out, nil
}
