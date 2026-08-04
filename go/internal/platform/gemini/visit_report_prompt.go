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
	// LocaleHint: empty/auto = keep source language; fr|nl|en|es|et|it|uk|ru = force output language.
	LocaleHint string
}

// ResolveVisitReportOutputLang returns the natural-language name used in the improve prompt.
// Empty / "auto" → keep the transcription/source language.
func ResolveVisitReportOutputLang(localeHint string) string {
	switch strings.ToLower(strings.TrimSpace(localeHint)) {
	case "", "auto", "source":
		return ""
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
	case "fr":
		return "français"
	default:
		return ""
	}
}

// NormalizeVisitReportTargetLocale allowlists improve targetLocale.
// Returns normalized code (auto|fr|nl|en|es|et|it|uk|ru) and false if unsupported.
func NormalizeVisitReportTargetLocale(raw string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" || v == "auto" || v == "source" {
		return "auto", true
	}
	switch v {
	case "fr", "nl", "en", "es", "et", "it", "uk", "ru":
		return v, true
	default:
		return "", false
	}
}

func visitReportLangInstruction(lang string) string {
	if lang == "" {
		return `LANGUE (impératif) : rédige le compte-rendu STRICTEMENT dans la même langue que le texte source / la transcription.
Ne traduis pas. Si la source est en français, le CR est en français ; si elle est en néerlandais, le CR est en néerlandais ; idem pour les autres langues.`
	}
	return fmt.Sprintf(`LANGUE (impératif) : rédige le compte-rendu STRICTEMENT en %s (traduis si la source est dans une autre langue).`, lang)
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
	lang := ResolveVisitReportOutputLang(in.LocaleHint)
	langRule := visitReportLangInstruction(lang)

	switch in.Specialty {
	case kernel.SpecialtyFarrier:
		return fmt.Sprintf(`Tu es un assistant pour maréchal-ferrant. Reformule le compte-rendu clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**État des pieds :**
**Fer / type :**
**Observations :**
**Recommandations :**

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, langRule, country)
	case kernel.SpecialtyPhysio:
		return fmt.Sprintf(`Tu es un assistant en physiothérapie animale. Reformule le CR de séance clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**Motif :**
**Examen :**
**Techniques :**
**Exercices / plan :**

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, langRule, country)
	case kernel.SpecialtyBehaviorist:
		return fmt.Sprintf(`Tu es un assistant comportementaliste animalier. Reformule le CR clair,
structuré STRICTEMENT avec ces sections (titres exacts) en Markdown (**Section :**) :

**Contexte :**
**Comportements observés :**
**Analyse :**
**Plan d'accompagnement :**

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`, langRule, country)
	}

	return fmt.Sprintf(`Tu es un assistant vétérinaire. Reformule le compte-rendu de visite clair et professionnel,
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

%s

Règles impératives :
- N'invente aucun fait, symptôme, résultat d'examen, diagnostic ou médicaments absents de la source.
- Les sections « Diagnostic proposé » et « Médication proposée » sont des PROPOSITIONS IA à valider par le vétérinaire ; indique-le brièvement si pertinent (ex. « proposition — à confirmer »).
- Pays d'exercice de référence : %s. Pour la médication, privilégie la DCI et des dénominations courantes / autorisées dans ce pays ; n'invente pas de posologie non mentionnée dans la source.
- La prescription finale reste exclusivement sous la responsabilité du vétérinaire.
- Si une section n'a pas d'information dans la source, écris « Non précisé » (ou l'équivalent dans la langue du CR) ; ne pas inventer.`, langRule, country)
}
