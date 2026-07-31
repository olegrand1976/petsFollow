package prescription

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeMedications(t *testing.T) {
	_, err := NormalizeMedications(nil)
	if err != ErrEmptyMeds {
		t.Fatalf("nil: %v", err)
	}
	_, err = NormalizeMedications([]byte(`[]`))
	if err != ErrEmptyMeds {
		t.Fatalf("empty: %v", err)
	}
	_, err = NormalizeMedications([]byte(`[{"name":""}]`))
	if err != ErrInvalidMeds {
		t.Fatalf("blank name: %v", err)
	}
	list, err := NormalizeMedications([]byte(`[{"name":" Amox ","dosage":"50mg","form":"tablet","quantity":"1 box","posology":"1/j","withdrawal_period":"","cnk":"123"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if list[0].Name != "Amox" || list[0].CNK != "123" {
		t.Fatalf("%+v", list[0])
	}

	tooMany := make([]map[string]string, MaxMedications+1)
	for i := range tooMany {
		tooMany[i] = map[string]string{"name": "X"}
	}
	raw, _ := json.Marshal(tooMany)
	if _, err := NormalizeMedications(raw); err != ErrPayloadTooLarge {
		t.Fatalf("too many: %v", err)
	}
	if _, err := NormalizeNotes(strings.Repeat("a", MaxNotesRunes+1)); err != ErrPayloadTooLarge {
		t.Fatalf("notes: %v", err)
	}
}

func TestBuildPDF(t *testing.T) {
	b, err := BuildPDF(PDFInput{
		PaperFormat:    FormatA5,
		CountryCode:    "BE",
		PracticeName:   "Cabinet Demo",
		Prescriber:     "Dr Demo",
		OwnerName:      "Client",
		PetName:        "Rex",
		PetSpecies:     "dog",
		DraftWatermark: true,
		Medications: []Medication{{
			Name: "Amoxicilline", Dosage: "50mg", Form: "tablet",
			Quantity: "1", Posology: "1x/jour 7j",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || string(b[:4]) != "%PDF" {
		t.Fatalf("not a pdf: %d bytes", len(b))
	}
}
