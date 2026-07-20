package gemini

import "testing"

func TestParseSuggestion(t *testing.T) {
	raw := `{"email":"Courriel","fullName":"Nom","locale":null,"ignored":["Tel"],"confidence":0.9}`
	sug, err := ParseSuggestion(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sug.Email == nil || *sug.Email != "Courriel" {
		t.Fatalf("email %#v", sug.Email)
	}
	if sug.FullName == nil || *sug.FullName != "Nom" {
		t.Fatalf("name %#v", sug.FullName)
	}
	if sug.Locale != nil {
		t.Fatalf("locale %#v", sug.Locale)
	}
	if sug.Confidence != 0.9 {
		t.Fatalf("confidence %v", sug.Confidence)
	}
}

func TestParseSuggestionMarkdownFence(t *testing.T) {
	raw := "```json\n{\"email\":\"e\",\"fullName\":\"n\",\"confidence\":0.5}\n```"
	sug, err := ParseSuggestion(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sug.Email == nil || *sug.Email != "e" {
		t.Fatalf("got %#v", sug)
	}
}

func TestNormalizeSuggestionDropsUnknownHeaders(t *testing.T) {
	email := "Nope"
	name := "Nom"
	sug := &MappingSuggestion{Email: &email, FullName: &name, Confidence: 2}
	normalizeSuggestion(sug, []string{"Nom", "Email"})
	if sug.Email != nil {
		t.Fatal("expected email dropped")
	}
	if sug.FullName == nil || *sug.FullName != "Nom" {
		t.Fatal("expected name kept")
	}
	if sug.Confidence != 1 {
		t.Fatalf("confidence clamped to %v", sug.Confidence)
	}
}
