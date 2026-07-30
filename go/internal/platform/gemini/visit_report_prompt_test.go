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
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, p)
		}
	}
}

func TestBuildVisitReportImprovePromptForceLocale(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "BE", LocaleHint: "nl"})
	if !strings.Contains(p, "STRICTEMENT en néerlandais") {
		t.Fatalf("expected forced Dutch:\n%s", p)
	}
	if strings.Contains(p, "même langue que le texte source") {
		t.Fatal("forced locale must not keep source-language rule")
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

func TestBuildVisitReportImprovePromptFarrierNoMeds(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{
		CountryCode: "BE",
		Specialty:   kernel.SpecialtyFarrier,
	})
	if strings.Contains(p, "Médication proposée") {
		t.Fatal("farrier prompt must not include vet medication section")
	}
	if !strings.Contains(p, "État des pieds") {
		t.Fatal("expected farrier sections")
	}
}

func TestBuildVisitReportImprovePromptDefaultCountry(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "xx"})
	if !strings.Contains(p, "Pays d'exercice de référence : BE") {
		t.Fatalf("expected BE fallback:\n%s", p)
	}
}
