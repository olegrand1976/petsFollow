package store

import "testing"

func TestValidatePreconsultAnswers(t *testing.T) {
	ok := PreconsultAnswers{
		ChiefComplaint: "Vomiting",
		Duration:       "few_days",
		Behavior:       "lethargic",
		Appetite:       "decreased",
		Thirst:         "normal",
		Elimination:    "normal",
		Urgency:        "medium",
		Comment:        "since Tuesday",
	}
	if err := ValidatePreconsultAnswers(ok); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	bad := ok
	bad.ChiefComplaint = ""
	if err := ValidatePreconsultAnswers(bad); err == nil {
		t.Fatal("expected chief_complaint_required")
	}
	bad = ok
	bad.Duration = "forever"
	if err := ValidatePreconsultAnswers(bad); err == nil {
		t.Fatal("expected invalid_duration")
	}
}
