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
Ne traduis pas. Si la source est en français, le CR est en français ; si elle est en néerlandais, le CR est en néerlandais ; idem pour les autres langues.
Les titres de section doivent aussi être dans cette même langue (traduis les titres de référence ci-dessous si besoin).`
	}
	return fmt.Sprintf(`LANGUE (impératif) : rédige le compte-rendu STRICTEMENT en %s (traduis si la source est dans une autre langue), y compris les titres de section listés ci-dessous.`, lang)
}

// resolveVisitReportSectionLocale maps LocaleHint to a section-title locale.
// auto/empty/unsupported → "fr" as semantic anchor (titles translated to source lang via lang instruction).
func resolveVisitReportSectionLocale(localeHint string) string {
	norm, ok := NormalizeVisitReportTargetLocale(localeHint)
	if !ok || norm == "auto" {
		return "fr"
	}
	return norm
}

// visitReportSupportedLocales is the allowlist used by section-title maps (parity tests).
var visitReportSupportedLocales = []string{"fr", "nl", "en", "es", "et", "it", "uk", "ru"}

type vetSectionTitles struct {
	Anamnesis, Clinical, Observations, Diagnosis, Medication, Plan string
}

// specialtySectionTitles is an ordered list of exactly 4 section titles for care_pro templates.
type specialtySectionTitles []string

var vetSectionsByLocale = map[string]vetSectionTitles{
	"fr": {
		Anamnesis: "Anamnèse / motif", Clinical: "Examen clinique", Observations: "Observations",
		Diagnosis: "Diagnostic proposé", Medication: "Médication proposée", Plan: "Plan / suivi",
	},
	"nl": {
		Anamnesis: "Anamnese / reden", Clinical: "Klinisch onderzoek", Observations: "Observaties",
		Diagnosis: "Voorgestelde diagnose", Medication: "Voorgestelde medicatie", Plan: "Plan / opvolging",
	},
	"en": {
		Anamnesis: "History / reason", Clinical: "Clinical examination", Observations: "Observations",
		Diagnosis: "Proposed diagnosis", Medication: "Proposed medication", Plan: "Plan / follow-up",
	},
	"es": {
		Anamnesis: "Anamnesis / motivo", Clinical: "Exploración clínica", Observations: "Observaciones",
		Diagnosis: "Diagnóstico propuesto", Medication: "Medicación propuesta", Plan: "Plan / seguimiento",
	},
	"et": {
		Anamnesis: "Anamnees / põhjus", Clinical: "Kliiniline läbivaatus", Observations: "Tähelepanekud",
		Diagnosis: "Pakutud diagnoos", Medication: "Pakutud ravimid", Plan: "Plaan / järelkontroll",
	},
	"it": {
		Anamnesis: "Anamnesi / motivo", Clinical: "Esame clinico", Observations: "Osservazioni",
		Diagnosis: "Diagnosi proposta", Medication: "Medicazione proposta", Plan: "Piano / follow-up",
	},
	"uk": {
		Anamnesis: "Анамнез / привід", Clinical: "Клінічний огляд", Observations: "Спостереження",
		Diagnosis: "Запропонований діагноз", Medication: "Запропонована медикація", Plan: "План / подальше спостереження",
	},
	"ru": {
		Anamnesis: "Анамнез / повод", Clinical: "Клинический осмотр", Observations: "Наблюдения",
		Diagnosis: "Предложенный диагноз", Medication: "Предложенная медикация", Plan: "План / наблюдение",
	},
}

var farrierSectionsByLocale = map[string]specialtySectionTitles{
	"fr": {"État des pieds", "Fer / type", "Observations", "Recommandations"},
	"nl": {"Toestand van de hoeven", "Beslag / type", "Observaties", "Aanbevelingen"},
	"en": {"Hoof condition", "Shoe / type", "Observations", "Recommendations"},
	"es": {"Estado de los cascos", "Herradura / tipo", "Observaciones", "Recomendaciones"},
	"et": {"Kabjade seisund", "Rauad / tüüp", "Tähelepanekud", "Soovitused"},
	"it": {"Stato degli zoccoli", "Ferro / tipo", "Osservazioni", "Raccomandazioni"},
	"uk": {"Стан копит", "Підкови / тип", "Спостереження", "Рекомендації"},
	"ru": {"Состояние копыт", "Подковы / тип", "Наблюдения", "Рекомендации"},
}

var physioSectionsByLocale = map[string]specialtySectionTitles{
	"fr": {"Motif", "Examen", "Techniques", "Exercices / plan"},
	"nl": {"Reden", "Onderzoek", "Technieken", "Oefeningen / plan"},
	"en": {"Reason", "Examination", "Techniques", "Exercises / plan"},
	"es": {"Motivo", "Exploración", "Técnicas", "Ejercicios / plan"},
	"et": {"Põhjus", "Läbivaatus", "Tehnikad", "Harjutused / plaan"},
	"it": {"Motivo", "Esame", "Tecniche", "Esercizi / piano"},
	"uk": {"Привід", "Огляд", "Техніки", "Вправи / план"},
	"ru": {"Повод", "Осмотр", "Техники", "Упражнения / план"},
}

var behavioristSectionsByLocale = map[string]specialtySectionTitles{
	"fr": {"Contexte", "Comportements observés", "Analyse", "Plan d'accompagnement"},
	"nl": {"Context", "Geobserveerd gedrag", "Analyse", "Begeleidingsplan"},
	"en": {"Context", "Observed behaviours", "Analysis", "Support plan"},
	"es": {"Contexto", "Comportamientos observados", "Análisis", "Plan de acompañamiento"},
	"et": {"Kontekst", "Täheldatud käitumine", "Analüüs", "Tugikava"},
	"it": {"Contesto", "Comportamenti osservati", "Analisi", "Piano di accompagnamento"},
	"uk": {"Контекст", "Спостережувана поведінка", "Аналіз", "План супроводу"},
	"ru": {"Контекст", "Наблюдаемое поведение", "Анализ", "План сопровождения"},
}

func vetTitlesForLocale(locale string) vetSectionTitles {
	if t, ok := vetSectionsByLocale[locale]; ok {
		return t
	}
	return vetSectionsByLocale["fr"]
}

func specialtyTitlesForLocale(m map[string]specialtySectionTitles, locale string) specialtySectionTitles {
	if t, ok := m[locale]; ok {
		return t
	}
	return m["fr"]
}

func formatExactSectionBlock(titles ...string) string {
	var b strings.Builder
	for _, t := range titles {
		b.WriteString("**")
		b.WriteString(t)
		b.WriteString(" :**\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func visitReportSectionModeIntro(auto bool) string {
	if auto {
		return `structuré STRICTEMENT avec ces sections (dans cet ordre) en Markdown.
Titres de référence (français) — traduis-les dans la langue du CR si la source n'est pas en français ; conserve le sens et l'ordre :`
	}
	return `structuré STRICTEMENT avec ces sections (titres exacts, dans cet ordre) en Markdown :`
}

func visitReportIsAutoLocale(localeHint string) bool {
	norm, ok := NormalizeVisitReportTargetLocale(localeHint)
	return !ok || norm == "auto"
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
	secLocale := resolveVisitReportSectionLocale(in.LocaleHint)
	auto := visitReportIsAutoLocale(in.LocaleHint)
	modeIntro := visitReportSectionModeIntro(auto)

	switch in.Specialty {
	case kernel.SpecialtyFarrier:
		t := specialtyTitlesForLocale(farrierSectionsByLocale, secLocale)
		return fmt.Sprintf(`Tu es un assistant pour maréchal-ferrant. Reformule le compte-rendu clair,
%s

%s

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`,
			modeIntro, formatExactSectionBlock(t...), langRule, country)
	case kernel.SpecialtyPhysio:
		t := specialtyTitlesForLocale(physioSectionsByLocale, secLocale)
		return fmt.Sprintf(`Tu es un assistant en physiothérapie animale. Reformule le CR de séance clair,
%s

%s

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`,
			modeIntro, formatExactSectionBlock(t...), langRule, country)
	case kernel.SpecialtyBehaviorist:
		t := specialtyTitlesForLocale(behavioristSectionsByLocale, secLocale)
		return fmt.Sprintf(`Tu es un assistant comportementaliste animalier. Reformule le CR clair,
%s

%s

%s
N'invente aucun fait absent de la source. Pas de HTML. Pays d'exercice de référence : %s.`,
			modeIntro, formatExactSectionBlock(t...), langRule, country)
	}

	t := vetTitlesForLocale(secLocale)
	return fmt.Sprintf(`Tu es un assistant vétérinaire. Reformule le compte-rendu de visite clair et professionnel,
%s

%s

Format Markdown obligatoire :
- Titres de section en gras (**Section :**) sur leur propre ligne.
- Sous-points d'examen en listes à puces (- **Sous-titre :** détail).
- Pas de HTML. Pas de blocs de code.

%s

Règles impératives :
- N'invente aucun fait, symptôme, résultat d'examen, diagnostic ou médicaments absents de la source.
- Les sections « %s » et « %s » sont des PROPOSITIONS IA à valider par le vétérinaire ; indique-le brièvement si pertinent (ex. « proposition — à confirmer »).
- Pays d'exercice de référence : %s. Pour la médication, privilégie la DCI et des dénominations courantes / autorisées dans ce pays ; n'invente pas de posologie non mentionnée dans la source.
- La prescription finale reste exclusivement sous la responsabilité du vétérinaire.
- Si une section n'a pas d'information dans la source, écris « Non précisé » (ou l'équivalent dans la langue du CR) ; ne pas inventer.`,
		modeIntro,
		formatExactSectionBlock(t.Anamnesis, t.Clinical, t.Observations, t.Diagnosis, t.Medication, t.Plan),
		langRule,
		t.Diagnosis,
		t.Medication,
		country)
}
