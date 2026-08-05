package gemini

import (
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestBuildVisitReportImprovePromptVetSections(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "fr", LocaleHint: "auto"})
	for _, want := range []string{
		"Anamnèse / motif",
		"Examen clinique",
		"Observations",
		"Diagnostic proposé",
		"Médication proposée",
		"Plan / suivi",
		"Pays d'exercice de référence : FR",
		"PROPOSITIONS IA",
		"Markdown",
		"**Anamnèse / motif :**",
		"même langue que le texte source",
		"traduis-les dans la langue du CR",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, p)
		}
	}
	if strings.Contains(p, "titres exacts, dans cet ordre") {
		t.Fatal("auto mode must not force exact French titles wording")
	}
}

func TestBuildVisitReportImprovePromptForceLocaleNL(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "BE", LocaleHint: "nl"})
	if !strings.Contains(p, "STRICTEMENT en néerlandais") {
		t.Fatalf("expected forced Dutch:\n%s", p)
	}
	if strings.Contains(p, "même langue que le texte source") {
		t.Fatal("forced locale must not keep source-language rule")
	}
	for _, want := range []string{
		"**Anamnese / reden :**",
		"**Klinisch onderzoek :**",
		"**Observaties :**",
		"**Voorgestelde diagnose :**",
		"**Voorgestelde medicatie :**",
		"**Plan / opvolging :**",
		"titres exacts, dans cet ordre",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("missing Dutch section %q:\n%s", want, p)
		}
	}
	for _, fr := range []string{"Anamnèse / motif", "Examen clinique", "Diagnostic proposé", "Médication proposée"} {
		if strings.Contains(p, fr) {
			t.Fatalf("forced NL must not keep French title %q:\n%s", fr, p)
		}
	}
}

func TestBuildVisitReportImprovePromptForceLocaleEN(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "BE", LocaleHint: "en"})
	if !strings.Contains(p, "STRICTEMENT en anglais") {
		t.Fatalf("expected forced English:\n%s", p)
	}
	for _, want := range []string{
		"**History / reason :**",
		"**Clinical examination :**",
		"**Proposed diagnosis :**",
		"**Proposed medication :**",
		"**Plan / follow-up :**",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("missing English section %q:\n%s", want, p)
		}
	}
	if strings.Contains(p, "Anamnèse / motif") {
		t.Fatal("forced EN must not keep French anamnesis title")
	}
}

func TestBuildVisitReportImprovePromptForceLocaleFR(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "BE", LocaleHint: "fr"})
	if !strings.Contains(p, "STRICTEMENT en français") {
		t.Fatalf("expected forced French:\n%s", p)
	}
	if !strings.Contains(p, "titres exacts, dans cet ordre") {
		t.Fatal("forced FR must use exact-titles wording")
	}
	if strings.Contains(p, "traduis-les dans la langue du CR") {
		t.Fatal("forced FR must not use auto translate-titles wording")
	}
}

func TestResolveVisitReportOutputLang(t *testing.T) {
	if got := ResolveVisitReportOutputLang("auto"); got != "" {
		t.Fatalf("auto => empty, got %q", got)
	}
	if got := ResolveVisitReportOutputLang("nl"); got != "néerlandais" {
		t.Fatalf("nl => néerlandais, got %q", got)
	}
}

func TestNormalizeVisitReportTargetLocale(t *testing.T) {
	got, ok := NormalizeVisitReportTargetLocale("")
	if !ok || got != "auto" {
		t.Fatalf("empty => auto,true got %q %v", got, ok)
	}
	got, ok = NormalizeVisitReportTargetLocale("NL")
	if !ok || got != "nl" {
		t.Fatalf("NL => nl,true got %q %v", got, ok)
	}
	if _, ok := NormalizeVisitReportTargetLocale("de"); ok {
		t.Fatal("de must be rejected")
	}
}

func TestVisitReportSectionTitleMapsParity(t *testing.T) {
	specialtyMaps := []struct {
		name string
		m    map[string]specialtySectionTitles
	}{
		{"farrier", farrierSectionsByLocale},
		{"physio", physioSectionsByLocale},
		{"behaviorist", behavioristSectionsByLocale},
	}
	if len(vetSectionsByLocale) != len(visitReportSupportedLocales) {
		t.Fatalf("vet map size %d want %d", len(vetSectionsByLocale), len(visitReportSupportedLocales))
	}
	for _, loc := range visitReportSupportedLocales {
		loc := loc
		t.Run("vet/"+loc, func(t *testing.T) {
			v, ok := vetSectionsByLocale[loc]
			if !ok {
				t.Fatalf("missing locale %q", loc)
			}
			for _, field := range []string{v.Anamnesis, v.Clinical, v.Observations, v.Diagnosis, v.Medication, v.Plan} {
				if strings.TrimSpace(field) == "" {
					t.Fatalf("empty vet title for locale %q", loc)
				}
			}
		})
		for _, sm := range specialtyMaps {
			sm := sm
			t.Run(sm.name+"/"+loc, func(t *testing.T) {
				titles, ok := sm.m[loc]
				if !ok {
					t.Fatalf("missing locale %q", loc)
				}
				if len(titles) != 4 {
					t.Fatalf("want 4 titles, got %d", len(titles))
				}
				for _, title := range titles {
					if strings.TrimSpace(title) == "" {
						t.Fatalf("empty specialty title for locale %q", loc)
					}
				}
			})
		}
	}
}

func TestBuildVisitReportImprovePromptFarrierNoMeds(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{
		CountryCode: "BE",
		Specialty:   kernel.SpecialtyFarrier,
		LocaleHint:  "nl",
	})
	if strings.Contains(p, "Médication proposée") || strings.Contains(p, "Voorgestelde medicatie") {
		t.Fatal("farrier prompt must not include vet medication section")
	}
	if !strings.Contains(p, "Toestand van de hoeven") {
		t.Fatal("expected localized farrier sections")
	}
	if strings.Contains(p, "État des pieds") {
		t.Fatal("forced NL farrier must not keep French titles")
	}
}

func TestBuildVisitReportImprovePromptDefaultCountry(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "xx"})
	if !strings.Contains(p, "Pays d'exercice de référence : BE") {
		t.Fatalf("expected BE fallback:\n%s", p)
	}
}
