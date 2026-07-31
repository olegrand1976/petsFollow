package gemini

import (
	"fmt"
	"strings"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// VisitReportPromptInput drives the structured vet CR improve prompt.
type VisitReportPromptInput struct {
	CountryCode string
	Specialty   kernel.ProfessionalSpecialty // empty = practice vet
	LocaleHint  string                       // e.g. fr
}

// BuildVisitReportImprovePrompt returns the system prompt for Gemini improve.
func BuildVisitReportImprovePrompt(in VisitReportPromptInput) string {
	country := strings.ToUpper(strings.TrimSpace(in.CountryCode))
	switch country {
	case "BE", "FR", "NL", "LU", "DE", "ES", "IT", "PT", "AT", "CH", "GB", "IE", "PL", "EE", "US":
		// ok
	default:
		country = "BE"
	}
	lang := "français"
	switch strings.ToLower(strings.TrimSpace(in.LocaleHint)) {
	case "nl":
		lang = "néerlandais"
	case "en":
		lang = "anglais"
	case "es":
		lang = "espagnol"
	case "et":
		lang = "estonien"
	case "it":
		lang = "italien"
	}

	switch in.Specialty {
	case kernel.SpecialtyFarrier:
		return fmt.Sprintf(`Tu es un assistant pour maréchal-ferrant. Reformule le compte-rendu en %s clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**État des pieds :**
**Fer / type :**
**Observations :**
**Recommandations :**

N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, lang, country)
	case kernel.SpecialtyPhysio:
		return fmt.Sprintf(`Tu es un assistant en physiothérapie animale. Reformule le CR de séance en %s clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**Motif :**
**Examen :**
**Techniques :**
**Exercices / plan :**

N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, lang, country)
	case kernel.SpecialtyBehaviorist:
		return fmt.Sprintf(`Tu es un assistant comportementaliste animalier. Reformule le CR en %s clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**Contexte :**
**Comportements observés :**
**Analyse :**
**Plan d'accompagnement :**

N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, lang, country)
	}

	return fmt.Sprintf(`Tu es un assistant vétérinaire. Reformule le compte-rendu de visite en %s clair et professionnel,
structuré STRICTEMENT avec ces sections (titres exacts, dans cet ordre) en Markdown :

**Anamnèse / motif :**
**Examen clinique :**
**Observations :**
**Diagnostic proposé :**
**Médication proposée :**
**Plan / suivi :**

Format Markdown obligatoire :
- Titres de section en gras (**Section :**) sur leur propre ligne.
- Sous-points d'examen en listes à puces (- **Sous-titre :** détail).
- Pas de HTML. Pas de blocs de code.

Règles impératives :
- N'invente aucun fait, symptôme, résultat d'examen, diagnostic ou médicaments absents de la source.
- Les sections « Diagnostic proposé » et « Médication proposée » sont des PROPOSITIONS IA à valider par le vétérinaire ; indique-le brièvement si pertinent (ex. « proposition — à confirmer »).
- Pays d'exercice de référence : %s. Pour la médication, privilégie la DCI et des dénominations courantes / autorisées dans ce pays ; n'invente pas de posologie non mentionnée dans la source.
- La prescription finale reste exclusivement sous la responsabilité du vétérinaire.
- Si une section n'a pas d'information dans la source, écris « Non précisé » (ne pas inventer).`, lang, country)
}
