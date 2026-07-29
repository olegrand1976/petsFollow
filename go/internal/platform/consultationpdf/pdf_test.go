package consultationpdf

import "testing"

func TestBuildPDFProducesMagic(t *testing.T) {
	b, err := BuildPDF(PDFInput{
		Marketing: Marketing{
			SiteURL:         "https://example.test",
			RegisterURL:     "https://example.test/register",
			ProductsURL:     "https://example.test/produits",
			CommercialName:  "Camille",
			CommercialPhone: "+32000000000",
		},
		PetName:      "Bella",
		Species:      "dog",
		OwnerName:    "Demo Client",
		PracticeName: "VetPlus",
		VisitWhen:    "2026-01-15 10:00",
		Reports: []ReportSection{
			{AuthorName: "Dr Demo", FinalizedAt: "2026-01-15", BodyText: "Examen clinique OK. Suivi dans 2 semaines."},
			{AuthorName: "Assistante", FinalizedAt: "2026-01-15", BodyText: "Vaccination à jour."},
		},
		Locale: "fr",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || string(b[:5]) != "%PDF-" {
		t.Fatalf("bad pdf magic len=%d", len(b))
	}
}
