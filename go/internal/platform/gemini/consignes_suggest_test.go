package gemini

import (
	"strings"
	"testing"
)

func TestBuildConsignesSuggestPromptMentionsRules(t *testing.T) {
	p := BuildConsignesSuggestPrompt(ConsignesSuggestPromptInput{
		CountryCode: "BE",
		PetName:     "Rex",
		PetSpecies:  "dog",
	})
	for _, needle := range []string{
		"Rex",
		"dog",
		"BE",
		"N'invente JAMAIS",
		"careAdvice",
		"ordonnance",
		"medications",
	} {
		if !strings.Contains(p, needle) {
			t.Fatalf("prompt missing %q", needle)
		}
	}
}

func TestParseConsignesSuggestJSON(t *testing.T) {
	raw := `{
		"medications": [
			{"name":" Metacam ","dosage":"1mg","form":"liquid","quantity":"1","posology":"SID","withdrawal_period":""}
		],
		"careAdvice": "  Repos au calme  ",
		"notes": " interne "
	}`
	s, err := ParseConsignesSuggestJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Medications) != 1 || s.Medications[0].Name != "Metacam" {
		t.Fatalf("meds %#v", s.Medications)
	}
	if s.CareAdvice != "Repos au calme" {
		t.Fatalf("careAdvice %q", s.CareAdvice)
	}
	if s.Notes != "interne" {
		t.Fatalf("notes %q", s.Notes)
	}
}

func TestParseConsignesSuggestJSONEmptyMeds(t *testing.T) {
	s, err := ParseConsignesSuggestJSON(`{"medications":[],"careAdvice":"Boire de l'eau","notes":""}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Medications) != 0 || s.CareAdvice != "Boire de l'eau" {
		t.Fatalf("%+v", s)
	}
}

func TestParseConsignesSuggestJSONFenced(t *testing.T) {
	raw := "```json\n{\"medications\":[{\"name\":\"X\"}],\"careAdvice\":\"ok\"}\n```"
	s, err := ParseConsignesSuggestJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Medications) != 1 || s.CareAdvice != "ok" {
		t.Fatalf("%+v", s)
	}
}


func TestParseConsignesSuggestJSONTruncatesMedFields(t *testing.T) {
	long := strings.Repeat("x", 600)
	raw := `{"medications":[{"name":"` + long + `","dosage":"` + long + `"}],"careAdvice":"ok","notes":""}`
	s, err := ParseConsignesSuggestJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Medications) != 1 {
		t.Fatalf("meds %#v", s.Medications)
	}
	if len([]rune(s.Medications[0].Name)) != 500 {
		t.Fatalf("name runes %d", len([]rune(s.Medications[0].Name)))
	}
	if len([]rune(s.Medications[0].Dosage)) != 500 {
		t.Fatalf("dosage runes %d", len([]rune(s.Medications[0].Dosage)))
	}
}
