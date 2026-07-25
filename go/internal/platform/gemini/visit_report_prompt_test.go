package gemini

import (
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestBuildVisitReportImprovePromptVetSections(t *testing.T) {
	p := BuildVisitReportImprovePrompt(VisitReportPromptInput{CountryCode: "fr", LocaleHint: "fr"})
	for _, want := range []string{
		"Anamnèse / motif",
		"Examen clinique",
		"Observations",
		"Diagnostic proposé",
		"Médication proposée",
		"Plan / suivi",
		"Pays d'exercice de référence : FR",
		"PROPOSITIONS IA",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, p)
		}
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
