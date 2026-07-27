package pharmacy

import "testing"

func TestFormatDAFNumber(t *testing.T) {
	if got := FormatDAFNumber(2026, 42); got != "DAF-2026-000042" {
		t.Fatalf("got %s", got)
	}
}

func TestValidateVamregPayload(t *testing.T) {
	if err := ValidateVamregPayload(nil); err != ErrDAFVAMRegIncomplete {
		t.Fatalf("nil: %v", err)
	}
	if err := ValidateVamregPayload([]byte(`{"species":"dog","indication":"x","durationDays":5}`)); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := ValidateVamregPayload([]byte(`{"species":"dog","indication":"x","durationDays":0}`)); err != ErrDAFVAMRegIncomplete {
		t.Fatalf("duration: %v", err)
	}
}

func TestBuildDAFPDF(t *testing.T) {
	b, err := BuildDAFPDF(DAFPDFInput{
		DisplayNumber: "DAF-2026-000001",
		PracticeName:  "Cabinet Demo",
		Prescriber:    "Dr Demo",
		Lines: []DAFPDFLine{{
			Medication: "Amoxi", CNK: "2712345", AMM: "BE-V123", Lot: "L1",
			ExpiresOn: "2027-01-01", Qty: "2", Unit: "box",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || b[0] != '%' {
		t.Fatalf("unexpected pdf len=%d", len(b))
	}
}
