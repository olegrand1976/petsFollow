package afsca

import (
	"regexp"
	"strings"
	"unicode"
)

// Animal scopes for newsletter tagging / practice preference.
const (
	ScopeSmall = "small" // companion / petits animaux
	ScopeLarge = "large" // production / gros animaux / équidés
	ScopeBoth  = "both"
)

// NormalizeScope returns small|large|both (default both).
func NormalizeScope(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ScopeSmall:
		return ScopeSmall
	case ScopeLarge:
		return ScopeLarge
	default:
		return ScopeBoth
	}
}

// MatchesScope reports whether an item tagged with itemScope should show for practiceScope.
func MatchesScope(practiceScope, itemScope string) bool {
	p := NormalizeScope(practiceScope)
	i := NormalizeScope(itemScope)
	if p == ScopeBoth || i == ScopeBoth {
		return true
	}
	return p == i
}

// FilterByScope keeps items matching the practice preference.
func FilterByScope(items []Item, practiceScope string) []Item {
	p := NormalizeScope(practiceScope)
	if p == ScopeBoth {
		return items
	}
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if MatchesScope(p, it.Scope) {
			out = append(out, it)
		}
	}
	return out
}

// HeuristicClassify tags items from title/label keywords (FR/NL/EN). Used when Gemini is off.
func HeuristicClassify(items []Item) []Item {
	out := make([]Item, len(items))
	for i, it := range items {
		it.Scope = heuristicScope(it.Title + " " + it.Label)
		out[i] = it
	}
	return out
}

func heuristicScope(text string) string {
	t := strings.ToLower(text)
	t = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' || r == '\'' {
			return r
		}
		return ' '
	}, t)
	largeHits := countPhraseOrTokenHits(t, largePhrases, largeTokens)
	smallHits := countPhraseOrTokenHits(t, smallPhrases, smallTokens)
	switch {
	case largeHits > 0 && smallHits == 0:
		return ScopeLarge
	case smallHits > 0 && largeHits == 0:
		return ScopeSmall
	case largeHits > 0 && smallHits > 0:
		return ScopeBoth
	default:
		return ScopeBoth
	}
}

func countPhraseOrTokenHits(text string, phrases, tokens []string) int {
	n := 0
	for _, p := range phrases {
		if p != "" && strings.Contains(text, p) {
			n++
		}
	}
	if len(tokens) == 0 {
		return n
	}
	tokSet := tokenize(text)
	for _, w := range tokens {
		if _, ok := tokSet[w]; ok {
			n++
		}
	}
	return n
}

var tokenSplit = regexp.MustCompile(`[^a-z0-9àâäæçéèêëïîôœùûüÿñ]+`)

func tokenize(text string) map[string]struct{} {
	parts := tokenSplit.Split(text, -1)
	out := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out[p] = struct{}{}
	}
	return out
}

// Multi-word phrases matched as substrings (safe — no short ambiguous tokens).
var largePhrases = []string{
	"fièvre catarrhale", "dermatose nodulaire", "nil occidental", "west nile",
	"westnijl", "animal de rente",
}

var smallPhrases = []string{
	"animal de compagnie", "gezelschapsdier", "animaux de compagnie",
}

// Single tokens matched with word boundaries (via tokenize).
var largeTokens = []string{
	"cheval", "chevaux", "équidé", "équidés", "equid", "equids", "paard", "paarden",
	"bovin", "bovins", "cattle", "vache", "vaches", "koe", "koeien",
	"ovin", "ovins", "ovine", "mouton", "moutons", "schaap", "schapen",
	"caprin", "caprins", "chèvre", "chèvres", "geit", "geiten",
	"porc", "porcs", "porcin", "varken", "varkens", "swine", "pig", "pigs",
	"volaille", "volailles", "pluimvee", "poultry", "aviaire", "vogelgriep", "newcastle",
	"blauwtong", "bluetongue", "ibr", "élevage", "abattoir", "slachthuis",
	"bétail", "vee", "horse", "horses", "sheep", "goat", "goats",
}

var smallTokens = []string{
	"chien", "chiens", "chat", "chats", "canin", "canine", "félin", "féline", "feline",
	"furet", "nac", "compagnon", "hond", "honden", "kat", "katten",
	"dog", "dogs", "cat", "cats", "puppy", "kitten",
}
