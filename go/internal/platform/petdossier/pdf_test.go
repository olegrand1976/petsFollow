package petdossier

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestTimelineLineHasNoReportFields(t *testing.T) {
	// Public dossier PDF must never grow hasReport / visitId on timeline rows.
	rt := reflect.TypeOf(TimelineLine{})
	if rt.NumField() != 3 {
		t.Fatalf("TimelineLine fields want 3 (When/Title/Body), got %d", rt.NumField())
	}
	for _, name := range []string{"When", "Title", "Body"} {
		if _, ok := rt.FieldByName(name); !ok {
			t.Fatalf("TimelineLine missing field %s", name)
		}
	}
	for _, forbidden := range []string{"HasReport", "ReportStatus", "VisitID", "Meta"} {
		if _, ok := rt.FieldByName(forbidden); ok {
			t.Fatalf("TimelineLine must not expose %s", forbidden)
		}
	}
}

func TestBuildPDFAndZip(t *testing.T) {
	pdf, err := BuildPDF(PDFInput{
		Marketing: Marketing{
			SiteURL:         "https://petsfollow.app",
			RegisterURL:     "https://petsfollow.app/register?invite=ABC",
			ProductsURL:     "https://petsfollow.app/produits",
			CommercialName:  "Camille Vente",
			CommercialPhone: "0470 12 34 56",
		},
		Pet: PetInfo{
			Name: "Rex", Species: "dog", Breed: "Labrador", OwnerName: "Client Démo",
		},
		Visits: []VisitLine{{
			When: "2026-01-01", Status: "done", PracticeName: "VetPlus", ProConsulted: "Dr Démo",
		}},
		HeartRates:    []HRLine{{When: "2026-01-02", BPM: "80", Alert: false}},
		Documents:     []DocLine{{Title: "Radio", FileName: "radio.pdf"}},
		HasHealthBook: true,
		GeneratedAt:   time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		Locale:        "fr",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Fatalf("expected PDF magic, got %q", pdf[:min(8, len(pdf))])
	}
	zipBytes, err := BuildZip([]ZipFile{
		{Name: "dossier.pdf", Data: pdf},
		{Name: "documents/radio.pdf", Data: []byte("%PDF-1.4 test")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(zipBytes) < 40 {
		t.Fatalf("zip too small: %d", len(zipBytes))
	}

	en, err := BuildPDF(PDFInput{
		Marketing:   Marketing{SiteURL: "https://petsfollow.app"},
		Pet:         PetInfo{Name: "Rex"},
		GeneratedAt: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		Locale:      "en",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(en, []byte("%PDF")) {
		t.Fatal("en PDF magic")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestBuildPDFCyrillic couvre les locales cyrilliques : libellés uk/ru ET
// données utilisateur en cyrillique (nom d'animal / de propriétaire). Avant le
// passage à une police UTF-8 (platform/pdffont), ces caractères sortaient en
// parasites via le traducteur cp1252 — y compris dans un PDF français.
func TestBuildPDFCyrillic(t *testing.T) {
	for _, loc := range []string{"uk", "ru", "fr"} {
		out, err := BuildPDF(PDFInput{
			Marketing: Marketing{SiteURL: "https://petsfollow.app"},
			Pet: PetInfo{
				Name:      "Мурчик",
				Species:   "cat",
				Breed:     "Сфінкс",
				OwnerName: "Олена Ґалаґан",
			},
			Visits:      []VisitLine{{When: "2026-01-01", Status: "done", PracticeName: "ВетПлюс", ProConsulted: "Др. Іваненко"}},
			Timeline:    []TimelineLine{{When: "2026-01-01", Title: "Щеплення", Body: "Щеплення виконано"}},
			GeneratedAt: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
			Locale:      loc,
		})
		if err != nil {
			t.Fatalf("%s: %v", loc, err)
		}
		if !bytes.HasPrefix(out, []byte("%PDF")) {
			t.Fatalf("%s: expected PDF magic", loc)
		}
	}
}
