package gemini

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/afsca"
)

const afscaClassifyTimeout = 4 * time.Second

type afscaScopeIn struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Label string `json:"label,omitempty"`
}

type afscaScopeOut struct {
	Items []struct {
		URL   string `json:"url"`
		Scope string `json:"scope"` // small|large|both
	} `json:"items"`
}

const afscaScopeSystem = `Tu classifies des newsletters AFSCA (Belgique) pour des vétérinaires praticiens.
Pour chaque entrée, attribue un scope :
- "small" : pertinence principale animaux de compagnie (chien, chat, NAC, furet…)
- "large" : pertinence principale animaux de production / élevage / équidés (bovins, ovins, porcs, volailles, chevaux…)
- "both" : transversal (réglementation générale, antibio, recrutement AFSCA, santé publique sans espèce claire, ou les deux)

Réponds UNIQUEMENT en JSON valide : {"items":[{"url":"…","scope":"small|large|both"},…]}
Conserve toutes les URL fournies, une entrée par URL.`

// ClassifyAfscaScopes tags newsletter items with small|large|both via Gemini lite.
// On failure, timeout, or if Gemini is not configured, returns heuristic classification.
func (c *Client) ClassifyAfscaScopes(ctx context.Context, items []afsca.Item) []afsca.Item {
	if len(items) == 0 {
		return items
	}
	fallback := afsca.HeuristicClassify(items)
	if c == nil || !c.Configured() {
		return fallback
	}
	in := make([]afscaScopeIn, 0, len(items))
	for _, it := range items {
		in = append(in, afscaScopeIn{URL: it.URL, Title: it.Title, Label: it.Label})
	}
	payload, err := json.Marshal(in)
	if err != nil {
		return fallback
	}
	user := "Entrées JSON :\n" + string(payload)

	callCtx, cancel := context.WithTimeout(ctx, afscaClassifyTimeout)
	defer cancel()
	raw, err := c.GenerateJSONLite(callCtx, afscaScopeSystem, user, 0.1)
	if err != nil {
		return fallback
	}
	var out afscaScopeOut
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		trimmed := strings.TrimSpace(raw)
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		if err2 := json.Unmarshal([]byte(trimmed), &out); err2 != nil {
			return fallback
		}
	}
	byURL := make(map[string]string, len(out.Items))
	for _, row := range out.Items {
		if row.URL == "" {
			continue
		}
		byURL[row.URL] = afsca.NormalizeScope(row.Scope)
	}
	tagged := make([]afsca.Item, len(items))
	for i, it := range items {
		if s, ok := byURL[it.URL]; ok {
			it.Scope = s
		} else {
			it.Scope = fallback[i].Scope
		}
		tagged[i] = it
	}
	return tagged
}
