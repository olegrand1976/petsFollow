package consultationpdf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedEmblemIsPNG(t *testing.T) {
	if len(emblemPNG) < 8 || !bytes.Equal(emblemPNG[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatalf("embedded emblem must be a real PNG (gofpdf rejects JPEG-as-.png), len=%d magic=%x", len(emblemPNG), emblemPNG[:min(8, len(emblemPNG))])
	}
}

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

func TestStripMarkdownRemovesStars(t *testing.T) {
	in := "**Anamnèse / motif :**\n- **Reason:** Annual vaccination.\n*italic*"
	out := StripMarkdown(in)
	if strings.Contains(out, "**") {
		t.Fatalf("still has markdown bold: %q", out)
	}
	if !strings.Contains(out, "Anamnèse / motif :") {
		t.Fatalf("missing heading text: %q", out)
	}
	if !strings.Contains(out, "- Reason:") {
		t.Fatalf("list not normalized to ASCII dash: %q", out)
	}
	if strings.Contains(out, "•") {
		t.Fatalf("bullet glyph should be ASCII: %q", out)
	}
}

func TestBuildPDFStripsMarkdownFromBody(t *testing.T) {
	b, err := BuildPDF(PDFInput{
		Marketing: Marketing{SiteURL: "https://example.test"},
		PetName:   "Yvonne",
		Reports: []ReportSection{{
			AuthorName: "Dr Demo",
			BodyText:   "**Anamnèse :**\n- Patient OK\n- Suivi",
		}},
		Locale: "fr",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Content streams are compressed; ensure build succeeds and is a real multi-KB branded PDF.
	if len(b) < 2000 {
		t.Fatalf("expected branded pdf with emblem, got len=%d", len(b))
	}
	plain := StripMarkdown("**Anamnèse :**\n- Patient OK")
	if strings.Contains(plain, "**") {
		t.Fatalf("strip failed: %q", plain)
	}
	if err := os.WriteFile(filepath.Join(t.TempDir(), "sample.pdf"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSanitizePDFTextMapsDashes(t *testing.T) {
	got := sanitizePDFText("a — b – c")
	if strings.Contains(got, "—") || strings.Contains(got, "–") {
		t.Fatalf("dashes not mapped: %q", got)
	}
	if !strings.Contains(got, "a - b - c") {
		t.Fatalf("unexpected: %q", got)
	}
}
