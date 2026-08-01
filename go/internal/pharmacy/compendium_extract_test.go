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
	st, code, _ := ClassifyExtractedRow(meds[0], false)
	if st != "pending" || code != "" {
		t.Fatalf("classify %s %s", st, code)
	}
	st, code, _ = ClassifyExtractedRow(meds[0], true)
	if st != "ready" || code != "" {
		t.Fatalf("confirmed classify %s %s", st, code)
	}
	st, code, _ = ClassifyExtractedRow(ExtractedMedication{Name: "X"}, false)
	if st != "error" || code != "missing_cnk" {
		t.Fatalf("expected missing_cnk got %s %s", st, code)
	}
	raw2 := `{"medications":[{"name":"VETMULIN","manufacturer":"Huvepharma","activeSubstance":"tiamuline","packSize":"sac 1 kg","prescriptionOnly":true}]}`
	meds2, err := ParseCompendiumExtractJSON(raw2)
	if err != nil || len(meds2) != 1 || meds2[0].Manufacturer != "Huvepharma" {
		t.Fatalf("enriched parse %#v %v", meds2, err)
	}
}
