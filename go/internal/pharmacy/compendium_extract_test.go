package pharmacy

import "testing"

func TestParseCompendiumExtractJSON(t *testing.T) {
	raw := `{"medications":[{"cnk":"2712345","name":"Amoxi Vet","atcCode":"J01CA04","isAntibiotic":true,"sourcePage":2}]}`
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(meds) != 1 || meds[0].CNK != "2712345" || !meds[0].IsAntibiotic {
		t.Fatalf("%#v", meds)
	}
	st, code, _ := ClassifyExtractedRow(meds[0])
	if st != "ready" || code != "" {
		t.Fatalf("classify %s %s", st, code)
	}
	st, code, _ = ClassifyExtractedRow(ExtractedMedication{Name: "X"})
	if st != "error" || code != "missing_cnk" {
		t.Fatalf("expected missing_cnk got %s %s", st, code)
	}
}
