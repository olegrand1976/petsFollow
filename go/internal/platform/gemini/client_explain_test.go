package gemini

import "testing"

func TestParseClientExplainJSON(t *testing.T) {
	raw := `{
		"disclaimer": "Suivez le véto.",
		"cards": [
			{"title": "Souffle", "body": "Un souffle grade 2 est souvent léger.", "kind": "term"},
			{"title": "Médicament", "body": "Donnez-le comme indiqué par le vétérinaire.", "kind": "medication"},
			{"title": "", "body": "skip", "kind": "general"}
		]
	}`
	out, err := ParseClientExplainJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Cards) != 2 {
		t.Fatalf("cards=%d want 2", len(out.Cards))
	}
	if out.Cards[0].Kind != "term" || out.Cards[1].Kind != "medication" {
		t.Fatalf("kinds %#v", out.Cards)
	}
}

func TestParseClientTriageJSON(t *testing.T) {
	raw := "```json\n{\"reply\":\"Surveillez 12h.\",\"level\":\"GREEN\",\"watchSigns\":[\"vomissements\",\" \"],\"recommendedAction\":\"Observer\"}\n```"
	out, err := ParseClientTriageJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.Level != TriageGreen {
		t.Fatalf("level=%q", out.Level)
	}
	if len(out.WatchSigns) != 1 || out.WatchSigns[0] != "vomissements" {
		t.Fatalf("signs %#v", out.WatchSigns)
	}
}
