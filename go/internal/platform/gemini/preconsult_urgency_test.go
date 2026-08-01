package gemini

import "testing"

func TestParsePreconsultUrgencyJSON(t *testing.T) {
	raw := "```json\n{\"urgency\":\"RED\",\"summary\":\" Détresse respiratoire — prioriser. \"}\n```"
	out, err := ParsePreconsultUrgencyJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.Urgency != TriageRed {
		t.Fatalf("urgency=%q want red", out.Urgency)
	}
	if out.Summary != "Détresse respiratoire — prioriser." {
		t.Fatalf("summary=%q", out.Summary)
	}
}

func TestParsePreconsultUrgencyJSONDefaultsUnknown(t *testing.T) {
	out, err := ParsePreconsultUrgencyJSON(`{"urgency":"purple","summary":"Cas ambigu"}`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Urgency != TriageOrange {
		t.Fatalf("urgency=%q want orange", out.Urgency)
	}
}

func TestParsePreconsultUrgencyJSONEmptySummary(t *testing.T) {
	if _, err := ParsePreconsultUrgencyJSON(`{"urgency":"green","summary":"  "}`); err == nil {
		t.Fatal("expected empty summary error")
	}
}
