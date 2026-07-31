// Package petdossier builds a shareable animal dossier PDF + ZIP pack.
package petdossier

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

type Marketing struct {
	SiteURL         string
	RegisterURL     string
	ProductsURL     string
	CommercialName  string
	CommercialPhone string
	CommercialEmail string
}

type PetInfo struct {
	Name             string
	Species          string
	Breed            string
	BirthDate        string
	WeightKg         string
	MicrochipNumber  string
	HealthBookNumber string
	OwnerName        string
	OwnerEmail       string
}

type WeightLine struct {
	When     string
	WeightKg string
	Comment  string
}

type HRLine struct {
	When  string
	BPM   string
	Alert bool
}

type ReminderLine struct {
	Title string
	Due   string
}

// TimelineLine is dossier-PDF only (When / Title / Body).
// No hasReport / visitId — public dossier must never grow a consultation CTA.
type TimelineLine struct {
	When  string
	Title string
	Body  string
}

type VisitLine struct {
	When         string
	Status       string
	PracticeName string
	ProConsulted string
	Notes        string
}

type DocLine struct {
	Title    string
	FileName string
}

type PDFInput struct {
	Marketing     Marketing
	Pet           PetInfo
	Weights       []WeightLine
	HeartRates    []HRLine
	Reminders     []ReminderLine
	Timeline      []TimelineLine
	Visits        []VisitLine
	Documents     []DocLine
	HasHealthBook bool
	GeneratedAt   time.Time
	Locale        string // fr default labels
}

func BuildPDF(in PDFInput) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()
	// Core Arial is WinAnsi — translate UTF-8 (FR accents, em dash) before drawing.
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	L := labelsFor(in.Locale)

	site := strings.TrimRight(strings.TrimSpace(in.Marketing.SiteURL), "/")
	register := strings.TrimSpace(in.Marketing.RegisterURL)
	if register == "" && site != "" {
		register = site + "/register"
	}
	products := strings.TrimSpace(in.Marketing.ProductsURL)
	if products == "" && site != "" {
		products = site + "/produits"
	}

	// Header brand
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(27, 58, 75)
	pdf.CellFormat(0, 10, "petsFollow", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(42, 157, 143)
	if site != "" {
		pdf.CellFormat(0, 6, site, "", 1, "L", false, 0, site)
	}
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(27, 58, 75)
	pdf.MultiCell(0, 5, tr(L.Subtitle), "", "L", false)
	pdf.Ln(2)

	// Marketing CTA
	pdf.SetFillColor(247, 249, 251)
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(27, 58, 75)
	cta := L.CTA
	if register != "" {
		cta += "\n" + L.Register + register
	}
	if products != "" {
		cta += "\n" + L.Offer + products
	}
	pdf.MultiCell(0, 5, tr(cta), "1", "L", true)
	pdf.Ln(3)

	// Commercial band
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, tr(L.ContactTitle), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	name := strings.TrimSpace(in.Marketing.CommercialName)
	if name == "" {
		name = L.DefaultContact
	}
	pdf.CellFormat(0, 5, tr(name), "", 1, "L", false, 0, "")
	if phone := strings.TrimSpace(in.Marketing.CommercialPhone); phone != "" {
		pdf.CellFormat(0, 5, tr(L.Phone+phone), "", 1, "L", false, 0, "")
	}
	if email := strings.TrimSpace(in.Marketing.CommercialEmail); email != "" {
		pdf.CellFormat(0, 5, tr(L.Email+email), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	section := func(title string) {
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(27, 58, 75)
		pdf.CellFormat(0, 7, tr(title), "", 1, "L", false, 0, "")
		pdf.SetDrawColor(226, 230, 237)
		pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
		pdf.Ln(2)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(40, 40, 40)
	}
	line := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(45, 5, tr(label), "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(0, 5, tr(value), "", "L", false)
	}

	section(L.Animal)
	line(L.Name, in.Pet.Name)
	line(L.Species, in.Pet.Species)
	line(L.Breed, in.Pet.Breed)
	line(L.Birth, in.Pet.BirthDate)
	line(L.Weight, in.Pet.WeightKg)
	line(L.Chip, in.Pet.MicrochipNumber)
	line(L.HealthBook, in.Pet.HealthBookNumber)
	line(L.Owner, in.Pet.OwnerName)
	line(L.OwnerEmail, in.Pet.OwnerEmail)
	pdf.Ln(2)

	if len(in.Weights) > 0 {
		section(L.RecentWeights)
		for _, w := range in.Weights {
			txt := w.When + " — " + w.WeightKg + " kg"
			if w.Comment != "" {
				txt += " (" + w.Comment + ")"
			}
			pdf.MultiCell(0, 5, tr("• "+txt), "", "L", false)
		}
		pdf.Ln(1)
	}

	if len(in.HeartRates) > 0 {
		section(L.HeartRate)
		for _, h := range in.HeartRates {
			txt := h.When + " — " + h.BPM + " bpm"
			if h.Alert {
				txt += L.Alert
			}
			pdf.MultiCell(0, 5, tr("• "+txt), "", "L", false)
		}
		pdf.Ln(1)
	}

	if len(in.Reminders) > 0 {
		section(L.CareReminders)
		for _, r := range in.Reminders {
			pdf.MultiCell(0, 5, tr("• "+r.Title+" — "+r.Due), "", "L", false)
		}
		pdf.Ln(1)
	}

	if len(in.Visits) > 0 {
		section(L.Visits)
		for _, v := range in.Visits {
			head := v.When + " — " + v.Status
			if v.PracticeName != "" {
				head += " — " + v.PracticeName
			}
			pdf.SetFont("Arial", "B", 9)
			pdf.MultiCell(0, 5, tr(head), "", "L", false)
			pdf.SetFont("Arial", "", 9)
			if v.ProConsulted != "" {
				pdf.MultiCell(0, 5, tr(L.ProConsulted+v.ProConsulted), "", "L", false)
			}
			if v.Notes != "" {
				pdf.MultiCell(0, 5, tr(v.Notes), "", "L", false)
			}
			pdf.Ln(1)
		}
	}

	if len(in.Timeline) > 0 {
		section(L.Timeline)
		for _, t := range in.Timeline {
			pdf.SetFont("Arial", "B", 9)
			pdf.MultiCell(0, 5, tr(t.When+" — "+t.Title), "", "L", false)
			pdf.SetFont("Arial", "", 9)
			if t.Body != "" {
				pdf.MultiCell(0, 5, tr(t.Body), "", "L", false)
			}
		}
		pdf.Ln(1)
	}

	section(L.Attachments)
	if in.HasHealthBook {
		pdf.MultiCell(0, 5, tr(L.HealthBookIncl), "", "L", false)
	}
	if len(in.Documents) == 0 && !in.HasHealthBook {
		pdf.MultiCell(0, 5, tr(L.NoAttachments), "", "L", false)
	}
	for _, d := range in.Documents {
		label := d.Title
		if label == "" {
			label = d.FileName
		}
		pdf.MultiCell(0, 5, tr("• "+label+" ("+d.FileName+")"), "", "L", false)
	}
	pdf.Ln(4)

	// Footer marketing
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(107, 114, 128)
	gen := in.GeneratedAt
	if gen.IsZero() {
		gen = time.Now().UTC()
	}
	footer := fmt.Sprintf(L.FooterGenerated, gen.Format("2006-01-02 15:04"))
	if site != "" {
		footer += "\n" + site
	}
	if register != "" {
		footer += "\n" + L.FooterCreateAcct + register
	}
	if phone := strings.TrimSpace(in.Marketing.CommercialPhone); phone != "" {
		footer += "\n" + L.FooterCommercial + phone
	}
	pdf.MultiCell(0, 4, tr(footer), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
