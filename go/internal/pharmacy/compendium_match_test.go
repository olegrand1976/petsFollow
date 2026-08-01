package pharmacy

import "testing"

func TestSuggestCNKAutoFill(t *testing.T) {
	extracted := ExtractedMedication{
		Name:               "VETORYL 20 mg caps",
		PharmaceuticalForm: "gélule",
		PackSize:           "caps 30",
		Manufacturer:       "Dechra",
	}
	refs := []RefMedMatchInput{
		{CNK: "2712001", Name: "Vetoryl 20 mg capsules", PharmaceuticalForm: "gélule", PackSize: "30", Manufacturer: "Dechra"},
		{CNK: "2712002", Name: "Amoxi Vet 500 mg", PharmaceuticalForm: "comprimé", PackSize: "20"},
	}
	got := SuggestCNK(extracted, refs)
	if !got.AutoFill || got.SuggestedCNK != "2712001" {
		t.Fatalf("expected autofill 2712001 got %#v", got)
	}
	if len(got.Candidates) < 1 || got.Candidates[0].CNK != "2712001" {
		t.Fatalf("candidates %#v", got.Candidates)
	}
}

func TestSuggestCNKAmbiguousNoAutoFill(t *testing.T) {
	extracted := ExtractedMedication{Name: "VETORYL caps"}
	refs := []RefMedMatchInput{
		{CNK: "1", Name: "VETORYL 10 mg caps"},
		{CNK: "2", Name: "VETORYL 20 mg caps"},
	}
	got := SuggestCNK(extracted, refs)
	if got.AutoFill {
		t.Fatalf("should not autofill ambiguous %#v", got)
	}
	if len(got.Candidates) < 2 {
		t.Fatalf("want candidates %#v", got.Candidates)
	}
}

func TestClassifyAfterMatch(t *testing.T) {
	m := ExtractedMedication{Name: "VETMULIN 100 mg/g"}
	out, st, code, _ := ClassifyAfterMatch(m, CNKMatchResult{
		SuggestedCNK: "2888001",
		Score:        0.92,
		AutoFill:     true,
		Candidates:   []CNKMatchCandidate{{CNK: "2888001", Name: "Vetmulin", Score: 0.92}},
	})
	if st != "pending" || out.CNK != "2888001" || code != "" {
		t.Fatalf("autofill classify %s %s %#v", st, code, out)
	}

	out, st, code, _ = ClassifyAfterMatch(m, CNKMatchResult{
		Candidates: []CNKMatchCandidate{{CNK: "1", Name: "A", Score: 0.5}, {CNK: "2", Name: "B", Score: 0.45}},
	})
	if st != "pending" || code != "cnk_unmatched" || out.CNK != "" {
		t.Fatalf("unmatched %s %s %#v", st, code, out)
	}

	out, st, code, _ = ClassifyAfterMatch(m, CNKMatchResult{})
	if st != "error" || code != "missing_cnk" {
		t.Fatalf("empty match %s %s", st, code)
	}
}
