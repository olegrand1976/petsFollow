package pdffont

import (
	"testing"

	"github.com/phpdave11/gofpdf"
)

// TestMetricCompatibleWithArial documente et verrouille la raison du choix de
// Liberation Sans : les largeurs latines doivent rester celles d'Arial, sinon
// les gabarits PDF existants (largeurs de cellules en mm) se décaleraient.
func TestMetricCompatibleWithArial(t *testing.T) {
	samples := []string{
		"Dossier animal - petsFollow",
		"Bartholomew (Chien, Labrador) — 12,5 kg",
		"Propriétaire : Éléonore Van Der Meersch",
		"AaBbCcXxYyZz 0123456789 %@/()",
	}
	for _, style := range []string{"", "B", "I"} {
		for _, size := range []float64{8, 9, 10, 11, 16} {
			for _, s := range samples {
				arial := gofpdf.New("P", "mm", "A4", "")
				arial.AddPage()
				arial.SetFont("Arial", style, size)
				// Comme en production : le core font mesure des octets cp1252,
				// pas des runes UTF-8 — sans traduction, chaque « é » compterait
				// double et la comparaison serait faussée.
				trArial := arial.UnicodeTranslatorFromDescriptor("")
				want := arial.GetStringWidth(trArial(s))

				lib := gofpdf.New("P", "mm", "A4", "")
				Register(lib)
				lib.AddPage()
				lib.SetFont(Family, style, size)
				got := lib.GetStringWidth(s)

				delta := got - want
				if delta < 0 {
					delta = -delta
				}
				// Tolérance 2 % : identique en pratique, marge pour l'arrondi
				// des unités de police (1/1000 em) entre les deux fontes.
				if want > 0 && delta/want > 0.02 {
					t.Errorf("style=%q size=%.0f %q: Liberation %.3fmm vs Arial %.3fmm (%.1f%%)",
						style, size, s, got, want, 100*delta/want)
				}
			}
		}
	}
}

func TestCyrillicRendersWithoutError(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	Register(pdf)
	pdf.AddPage()
	pdf.SetFont(Family, "", 11)
	// Ukrainien (ґ є і ї) + russe (ё ъ ы).
	pdf.MultiCell(180, 6, "Картка тварини — ґедзь, єнот, їжак, північ\nКарта животного — ёлка, объезд, рыба", "", "L", false)
	if err := pdf.Error(); err != nil {
		t.Fatalf("Cyrillic draw failed: %v", err)
	}
}
