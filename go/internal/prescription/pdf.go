package prescription

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

// PDFInput is the printable consignes (care instructions) content — not a legal prescription.
type PDFInput struct {
	PaperFormat    string
	CountryCode    string
	PracticeName   string
	Prescriber     string
	OwnerName      string
	PetName        string
	PetSpecies     string
	DateIssued     *time.Time
	ValidUntil     *time.Time
	Notes          string
	CareAdvice     string
	Medications    []Medication
	DraftWatermark bool
}

// BuildPDF renders an A4/A5 consignes sheet PDF (V1 preview / draft).
func BuildPDF(in PDFInput) ([]byte, error) {
	size := FormatA4
	if strings.EqualFold(in.PaperFormat, FormatA5) {
		size = FormatA5
	}
	pdf := gofpdf.New("P", "mm", size, "")
	title := documentTitle(in.CountryCode)
	pdf.SetTitle(title, false)
	pdf.AddPage()

	if in.DraftWatermark {
		pdf.SetFont("Arial", "B", 48)
		pdf.SetTextColor(220, 220, 220)
		pdf.TransformBegin()
		pdf.TransformRotate(45, 40, 140)
		pdf.Text(40, 140, "DRAFT")
		pdf.TransformEnd()
		pdf.SetTextColor(0, 0, 0)
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, title)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "practice"), sanitize(in.PracticeName)))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "prescriber"), sanitize(in.Prescriber)))
	pdf.Ln(5)
	if in.OwnerName != "" {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "owner"), sanitize(in.OwnerName)))
		pdf.Ln(5)
	}
	if in.PetName != "" {
		petLine := sanitize(in.PetName)
		if in.PetSpecies != "" {
			petLine = petLine + " (" + sanitize(in.PetSpecies) + ")"
		}
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "pet"), petLine))
		pdf.Ln(5)
	}
	if in.DateIssued != nil {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "issued"), in.DateIssued.Format("2006-01-02")))
		pdf.Ln(5)
	}
	if in.ValidUntil != nil {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", label(in.CountryCode, "valid"), in.ValidUntil.Format("2006-01-02")))
		pdf.Ln(5)
	}
	pdf.Ln(3)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, label(in.CountryCode, "medications"))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	for i, m := range in.Medications {
		pdf.SetFont("Arial", "B", 9)
		pdf.MultiCell(0, 5, fmt.Sprintf("%d. %s", i+1, sanitize(m.Name)), "", "", false)
		pdf.SetFont("Arial", "", 9)
		parts := []string{}
		if m.Dosage != "" {
			parts = append(parts, label(in.CountryCode, "dosage")+": "+sanitize(m.Dosage))
		}
		if m.Form != "" {
			parts = append(parts, label(in.CountryCode, "form")+": "+sanitize(m.Form))
		}
		if m.Quantity != "" {
			parts = append(parts, label(in.CountryCode, "qty")+": "+sanitize(m.Quantity))
		}
		if len(parts) > 0 {
			pdf.MultiCell(0, 4, strings.Join(parts, "  |  "), "", "", false)
		}
		if m.Posology != "" {
			pdf.MultiCell(0, 4, label(in.CountryCode, "posology")+": "+sanitize(m.Posology), "", "", false)
		}
		if m.WithdrawalPeriod != "" {
			pdf.MultiCell(0, 4, label(in.CountryCode, "withdrawal")+": "+sanitize(m.WithdrawalPeriod), "", "", false)
		}
		if m.CNK != "" {
			pdf.MultiCell(0, 4, "CNK: "+sanitize(m.CNK), "", "", false)
		}
		pdf.Ln(2)
	}

	if strings.TrimSpace(in.CareAdvice) != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, label(in.CountryCode, "careAdvice"))
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(0, 4, sanitize(in.CareAdvice), "", "", false)
	}

	if strings.TrimSpace(in.Notes) != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "I", 9)
		pdf.MultiCell(0, 4, label(in.CountryCode, "notes")+": "+sanitize(in.Notes), "", "", false)
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 7)
	pdf.MultiCell(0, 3, disclaimer(in.CountryCode), "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func documentTitle(cc string) string {
	switch strings.ToUpper(cc) {
	case "IT":
		return "Consegne di cura"
	case "ES":
		return "Instrucciones de cuidados"
	case "NL":
		return "Zorginstructies"
	default:
		return "Consignes"
	}
}

func label(cc, key string) string {
	// V1: French labels for BE/FR; light variants for IT/ES/NL.
	fr := map[string]string{
		"practice":    "Cabinet",
		"prescriber":  "Veterinaire",
		"owner":       "Proprietaire",
		"pet":         "Animal",
		"issued":      "Date",
		"valid":       "Valable jusqu'au",
		"medications": "Medicaments",
		"dosage":      "Dosage",
		"form":        "Forme",
		"qty":         "Quantite",
		"posology":    "Posologie",
		"withdrawal":  "Delai d'attente",
		"notes":       "Notes",
		"careAdvice":  "Conseils et soins",
	}
	it := map[string]string{
		"practice": "Clinica", "prescriber": "Veterinario", "owner": "Proprietario",
		"pet": "Animale", "issued": "Data", "valid": "Valida fino al",
		"medications": "Farmaci", "dosage": "Dosaggio", "form": "Forma",
		"qty": "Quantita", "posology": "Posologia", "withdrawal": "Tempo di sospensione",
		"notes": "Note", "careAdvice": "Consigli e cure",
	}
	es := map[string]string{
		"practice": "Clinica", "prescriber": "Veterinario", "owner": "Propietario",
		"pet": "Animal", "issued": "Fecha", "valid": "Valida hasta",
		"medications": "Medicamentos", "dosage": "Dosis", "form": "Forma",
		"qty": "Cantidad", "posology": "Posologia", "withdrawal": "Periodo de espera",
		"notes": "Notas", "careAdvice": "Consejos y cuidados",
	}
	nl := map[string]string{
		"practice": "Praktijk", "prescriber": "Dierenarts", "owner": "Eigenaar",
		"pet": "Dier", "issued": "Datum", "valid": "Geldig tot",
		"medications": "Geneesmiddelen", "dosage": "Dosering", "form": "Vorm",
		"qty": "Hoeveelheid", "posology": "Posologie", "withdrawal": "Wachttijd",
		"notes": "Notities", "careAdvice": "Advies en zorg",
	}
	var m map[string]string
	switch strings.ToUpper(cc) {
	case "IT":
		m = it
	case "ES":
		m = es
	case "NL":
		m = nl
	default:
		m = fr
	}
	if v, ok := m[key]; ok {
		return v
	}
	return key
}

func disclaimer(cc string) string {
	switch strings.ToUpper(cc) {
	case "IT":
		return "Foglio consegne generato da petsFollow (bozza). Non sostituisce una ricetta cartacea legale."
	case "ES":
		return "Ficha de instrucciones generada por petsFollow (borrador). No sustituye una receta legal en papel."
	case "NL":
		return "Zorgfiche gegenereerd door petsFollow (concept). Vervangt geen wettelijk papieren voorschrift."
	default:
		return "Fiche consignes generee par petsFollow (brouillon). Ne remplace pas une ordonnance legale (papier carbone)."
	}
}

// sanitize strips characters that break core fonts in gofpdf (latin-1 subset).
func sanitize(s string) string {
	replacer := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a", "ä", "a",
		"ù", "u", "û", "u", "ü", "u",
		"ô", "o", "ö", "o",
		"î", "i", "ï", "i",
		"ç", "c", "œ", "oe", "æ", "ae",
		"É", "E", "È", "E", "Ê", "E",
		"À", "A", "Â", "A",
		"Ù", "U", "Ç", "C",
		"ñ", "n", "Ñ", "N",
		"á", "a", "í", "i", "ó", "o", "ú", "u",
	)
	return replacer.Replace(s)
}
