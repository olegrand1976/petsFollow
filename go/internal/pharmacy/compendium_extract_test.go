package pharmacy

import (
	"strings"
	"testing"
)

func TestDedupExtractedMedications(t *testing.T) {
	in := []ExtractedMedication{
		{CNK: "111", Name: "A"},
		{CNK: "111", Name: "A dup"},
		{CNK: "222", Name: "B"},
		{CNK: "", Name: "NoCNK"},
		{CNK: "", Name: "nocnk"},
	}
	out := DedupExtractedMedications(in)
	if len(out) != 3 {
		t.Fatalf("got %d %#v", len(out), out)
	}
	if out[0].Name != "A" || out[1].CNK != "222" || out[2].Name != "NoCNK" {
		t.Fatalf("%#v", out)
	}
}

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

// Gemini often returns fractional withdrawal days (Vetcompendium "3,5 j") as JSON numbers.
// A hard *int unmarshal used to fail the whole chunk → extract_failed mid-job.
func TestParseCompendiumExtractJSON_FractionalWithdrawalDays(t *testing.T) {
	raw := `{"medications":[{
		"name":"Exemple Lait 3.5",
		"manufacturer":"DemoLab",
		"withdrawalMeatDays":28,
		"withdrawalMilkDays":3.5,
		"withdrawalEggsDays":"1,5",
		"packSize":"flacon 100 ml"
	}]}`
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(meds) != 1 {
		t.Fatalf("len=%d", len(meds))
	}
	m := meds[0]
	if m.WithdrawalMeatDays == nil || *m.WithdrawalMeatDays != 28 {
		t.Fatalf("meat %#v", m.WithdrawalMeatDays)
	}
	if m.WithdrawalMilkDays == nil || *m.WithdrawalMilkDays != 4 { // Ceil(3.5)=4
		t.Fatalf("milk %#v", m.WithdrawalMilkDays)
	}
	if m.WithdrawalEggsDays == nil || *m.WithdrawalEggsDays != 2 { // Ceil(1.5)=2
		t.Fatalf("eggs %#v", m.WithdrawalEggsDays)
	}
}

func TestParseCompendiumExtractJSON_LooseTypes(t *testing.T) {
	raw := `{"medications":[{
		"name":"Loose Types Vet",
		"cnk":2712345,
		"strength":20,
		"species":"Ca",
		"prescriptionOnly":"oui",
		"isAntibiotic":1,
		"sourcePage":453.2,
		"packSize":"caps 10"
	}]}`
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(meds) != 1 {
		t.Fatalf("len=%d", len(meds))
	}
	m := meds[0]
	if m.CNK != "2712345" {
		t.Fatalf("cnk %q", m.CNK)
	}
	if m.Strength != "20" {
		t.Fatalf("strength %q", m.Strength)
	}
	if len(m.Species) != 1 || m.Species[0] != "Ca" {
		t.Fatalf("species %#v", m.Species)
	}
	if !m.PrescriptionOnly {
		t.Fatal("prescriptionOnly expected true")
	}
	if !m.IsAntibiotic {
		t.Fatal("isAntibiotic expected true")
	}
	if m.SourcePage == nil || *m.SourcePage != 453 {
		t.Fatalf("sourcePage %#v", m.SourcePage)
	}
}

func TestParseCompendiumExtractJSON_SoftSkipBadRow(t *testing.T) {
	// Second element is not an object → skipped; first kept.
	raw := `{"medications":[{"name":"Good Med","cnk":""}, "not-an-object", {"name":"Also Good","species":["Su"]}]}`
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(meds) != 2 {
		t.Fatalf("len=%d %#v", len(meds), meds)
	}
	if meds[0].Name != "Good Med" || meds[1].Name != "Also Good" {
		t.Fatalf("%#v", meds)
	}
}

func TestParseCompendiumExtractJSON_PartialRowKept(t *testing.T) {
	// No name/CNK but lab + page recovered → keep for manual completion.
	raw := `{"medications":[{
		"manufacturer":"Dechra",
		"pharmaceuticalForm":"gélule",
		"sourcePage":453,
		"packSize":"caps 30"
	}]}`
	meds, err := ParseCompendiumExtractJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(meds) != 1 {
		t.Fatalf("len=%d", len(meds))
	}
	m := meds[0]
	if m.Name != "" || m.CNK != "" {
		t.Fatalf("expected empty identity %#v", m)
	}
	if m.Manufacturer != "Dechra" || m.SourcePage == nil || *m.SourcePage != 453 {
		t.Fatalf("%#v", m)
	}
	st, code, _ := ClassifyExtractedRow(m, false)
	if st != "error" || code != "missing_name" {
		t.Fatalf("classify %s %s", st, code)
	}
}

func TestParseCompendiumExtractJSON_EmptyAfterCoerce(t *testing.T) {
	raw := `{"medications":["not-an-object", {"name":""}, null]}`
	_, err := ParseCompendiumExtractJSON(raw)
	if err == nil || !strings.Contains(err.Error(), "empty_after_coerce") {
		t.Fatalf("expected empty_after_coerce, got %v", err)
	}
}

func TestParseCompendiumExtractJSON_EmptyMedicationsOK(t *testing.T) {
	meds, err := ParseCompendiumExtractJSON(`{"medications":[]}`)
	if err != nil || len(meds) != 0 {
		t.Fatalf("empty array should succeed: %#v %v", meds, err)
	}
}

func TestCoerceWithdrawalDays(t *testing.T) {
	if got := coerceWithdrawalDays(nil); got != nil {
		t.Fatalf("nil → %#v", got)
	}
	if got := coerceWithdrawalDays("null"); got != nil {
		t.Fatalf("null string → %#v", got)
	}
	if got := coerceWithdrawalDays("not-a-number"); got != nil {
		t.Fatalf("junk → %#v", got)
	}
	if got := coerceWithdrawalDays(-2.0); got != nil {
		t.Fatalf("negative → %#v", got)
	}
	if got := coerceWithdrawalDays(3.1); got == nil || *got != 4 { // Ceil
		t.Fatalf("ceil 3.1 → %#v", got)
	}
	if got := coerceWithdrawalDays(0); got == nil || *got != 0 {
		t.Fatalf("zero → %#v", got)
	}
}

func TestCoerceBool(t *testing.T) {
	if !coerceBool(true) || coerceBool(false) {
		t.Fatal("bool passthrough")
	}
	if !coerceBool("oui") || !coerceBool(1.0) || coerceBool("non") {
		t.Fatal("loose bool")
	}
	if coerceBool("r/") {
		t.Fatal("r/ must not coerce to true")
	}
}

func TestCoerceSpecies(t *testing.T) {
	if got := coerceSpecies("Ca"); len(got) != 1 || got[0] != "Ca" {
		t.Fatalf("%#v", got)
	}
	if got := coerceSpecies([]any{"Ca", " Su ", ""}); len(got) != 2 || got[1] != "Su" {
		t.Fatalf("%#v", got)
	}
}

func TestCoerceString(t *testing.T) {
	if got := coerceString(2712345.0); got != "2712345" {
		t.Fatalf("%q", got)
	}
	if got := coerceString(20.5); got != "20.5" {
		t.Fatalf("%q", got)
	}
	if got := coerceString([]any{"x"}); got != "" {
		t.Fatalf("array → %q", got)
	}
}
