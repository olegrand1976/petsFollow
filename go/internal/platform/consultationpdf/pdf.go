// Package consultationpdf builds a branded PDF for a shared consultation report.
package consultationpdf

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

type ReportSection struct {
	AuthorName  string
	FinalizedAt string
	BodyText    string
}

type PDFInput struct {
	Marketing    Marketing
	PetName      string
	Species      string
	Breed        string
	OwnerName    string
	PracticeName string
	VisitWhen    string
	Reports      []ReportSection
	GeneratedAt  time.Time
	Locale       string
}

type labels struct {
	Subtitle      string
	CTA           string
	Register      string
	Offer         string
	ContactTitle  string
	DefaultContact string
	Phone         string
	Email         string
	Pet           string
	Visit         string
	Practice      string
	ReportBy      string
	Finalized     string
	NoBody        string
}

func labelsFor(locale string) labels {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "nl":
		return labels{
			Subtitle: "Consultatieverslag — gedeeld via petsFollow",
			CTA: "petsFollow helpt dierenartsen bij opvolging, berichten en continuïteit van zorg.",
			Register: "Pro-account aanmaken: ",
			Offer: "Aanbod: ",
			ContactTitle: "Commercieel contact",
			DefaultContact: "petsFollow",
			Phone: "Tel. ",
			Email: "E-mail ",
			Pet: "Dier",
			Visit: "Bezoek",
			Practice: "Praktijk",
			ReportBy: "Verslag door ",
			Finalized: "Afgerond: ",
			NoBody: "(geen tekst)",
		}
	case "en":
		return labels{
			Subtitle: "Consultation report — shared via petsFollow",
			CTA: "petsFollow helps professionals follow patients, message owners and centralise care.",
			Register: "Create a pro account: ",
			Offer: "Offer: ",
			ContactTitle: "Sales contact",
			DefaultContact: "petsFollow",
			Phone: "Phone ",
			Email: "Email ",
			Pet: "Pet",
			Visit: "Visit",
			Practice: "Practice",
			ReportBy: "Report by ",
			Finalized: "Finalized: ",
			NoBody: "(empty)",
		}
	case "es":
		return labels{
			Subtitle: "Informe de consulta — compartido vía petsFollow",
			CTA: "petsFollow ayuda a los profesionales a seguir pacientes, hablar con dueños y centralizar cuidados.",
			Register: "Crear cuenta pro: ",
			Offer: "Oferta: ",
			ContactTitle: "Contacto comercial",
			DefaultContact: "petsFollow",
			Phone: "Tel. ",
			Email: "Email ",
			Pet: "Animal",
			Visit: "Visita",
			Practice: "Clínica",
			ReportBy: "Informe de ",
			Finalized: "Finalizado: ",
			NoBody: "(vacío)",
		}
	case "et":
		return labels{
			Subtitle: "Konsultatsiooni aruanne — jagatud petsFollowi kaudu",
			CTA: "petsFollow aitab professionaalidel patsiente jälgida, omanikega suhelda ja hooldust tsentraliseerida.",
			Register: "Loo pro konto: ",
			Offer: "Pakkumine: ",
			ContactTitle: "Müügikontakt",
			DefaultContact: "petsFollow",
			Phone: "Tel. ",
			Email: "E-post ",
			Pet: "Loom",
			Visit: "Külastus",
			Practice: "Kliinik",
			ReportBy: "Aruanne: ",
			Finalized: "Lõpetatud: ",
			NoBody: "(tühi)",
		}
	case "it":
		return labels{
			Subtitle: "Referto di consultazione — condiviso via petsFollow",
			CTA: "petsFollow aiuta i professionisti a seguire i pazienti, comunicare con i proprietari e centralizzare le cure.",
			Register: "Crea un account pro: ",
			Offer: "Offerta: ",
			ContactTitle: "Contatto commerciale",
			DefaultContact: "petsFollow",
			Phone: "Tel. ",
			Email: "Email ",
			Pet: "Animale",
			Visit: "Visita",
			Practice: "Clinica",
			ReportBy: "Referto di ",
			Finalized: "Finalizzato: ",
			NoBody: "(vuoto)",
		}
	default:
		return labels{
			Subtitle: "Compte-rendu de consultation — partagé via petsFollow",
			CTA: "petsFollow aide les professionnels à suivre leurs patients, échanger avec les propriétaires et centraliser les soins.",
			Register: "Créer un compte pro : ",
			Offer: "Offre : ",
			ContactTitle: "Contact commercial",
			DefaultContact: "petsFollow",
			Phone: "Tél. ",
			Email: "Email ",
			Pet: "Animal",
			Visit: "Visite",
			Practice: "Cabinet",
			ReportBy: "Compte-rendu par ",
			Finalized: "Finalisé : ",
			NoBody: "(vide)",
		}
	}
}

func BuildPDF(in PDFInput) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddPage()
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
		pdf.SetDrawColor(42, 157, 143)
		pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
		pdf.Ln(3)
	}

	section(L.Pet)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(40, 40, 40)
	line := in.PetName
	if sp := strings.TrimSpace(in.Species); sp != "" {
		line += " · " + sp
	}
	if br := strings.TrimSpace(in.Breed); br != "" {
		line += " · " + br
	}
	pdf.MultiCell(0, 5, tr(line), "", "L", false)
	if owner := strings.TrimSpace(in.OwnerName); owner != "" {
		pdf.MultiCell(0, 5, tr(owner), "", "L", false)
	}
	pdf.Ln(2)

	section(L.Visit)
	if when := strings.TrimSpace(in.VisitWhen); when != "" {
		pdf.MultiCell(0, 5, tr(when), "", "L", false)
	}
	if pr := strings.TrimSpace(in.PracticeName); pr != "" {
		pdf.MultiCell(0, 5, tr(L.Practice+": "+pr), "", "L", false)
	}
	pdf.Ln(2)

	for _, rep := range in.Reports {
		title := L.ReportBy + strings.TrimSpace(rep.AuthorName)
		if title == L.ReportBy {
			title = L.ReportBy + "—"
		}
		section(title)
		if fa := strings.TrimSpace(rep.FinalizedAt); fa != "" {
			pdf.SetFont("Arial", "I", 9)
			pdf.SetTextColor(100, 100, 100)
			pdf.MultiCell(0, 5, tr(L.Finalized+fa), "", "L", false)
			pdf.Ln(1)
		}
		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(40, 40, 40)
		body := strings.TrimSpace(rep.BodyText)
		if body == "" {
			body = L.NoBody
		}
		pdf.MultiCell(0, 5, tr(body), "", "L", false)
		pdf.Ln(3)
	}

	pdf.Ln(4)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	gen := in.GeneratedAt
	if gen.IsZero() {
		gen = time.Now().UTC()
	}
	pdf.MultiCell(0, 4, tr(fmt.Sprintf("petsFollow · %s", gen.Format("2006-01-02 15:04 UTC"))), "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
