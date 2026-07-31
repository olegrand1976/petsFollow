// Package consultationpdf builds a branded PDF for a shared consultation report.
package consultationpdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/phpdave11/gofpdf"
)

//go:embed assets/emblem.png
var emblemPNG []byte

// Brand colors — documentation/13-CHARTE-GRAPHIQUE.md
var (
	colorNavy    = [3]int{27, 58, 75}     // #1B3A4B
	colorTeal    = [3]int{42, 157, 143}   // #2A9D8F
	colorMuted   = [3]int{107, 114, 128}  // #6B7280
	colorBody    = [3]int{27, 58, 75}
	colorBorder  = [3]int{226, 230, 237}  // #E2E6ED
	colorSurface = [3]int{247, 249, 251}  // #F7F9FB
	colorWhite   = [3]int{255, 255, 255}
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
	Subtitle       string
	CTA            string
	Register       string
	Offer          string
	ContactTitle   string
	DefaultContact string
	Phone          string
	Email          string
	Pet            string
	Visit          string
	Practice       string
	ReportBy       string
	Finalized      string
	NoBody         string
	Confidential   string
}

func labelsFor(locale string) labels {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "nl":
		return labels{
			Subtitle:       "Consultatieverslag — gedeeld via petsFollow",
			CTA:            "petsFollow helpt dierenartsen bij opvolging, berichten en continuïteit van zorg.",
			Register:       "Pro-account aanmaken: ",
			Offer:          "Aanbod: ",
			ContactTitle:   "Commercieel contact",
			DefaultContact: "petsFollow",
			Phone:          "Tel. ",
			Email:          "E-mail ",
			Pet:            "Dier",
			Visit:          "Bezoek",
			Practice:       "Praktijk",
			ReportBy:       "Verslag door ",
			Finalized:      "Afgerond: ",
			NoBody:         "(geen tekst)",
			Confidential:   "Vertrouwelijk — uitsluitend voor de geadresseerde professional.",
		}
	case "en":
		return labels{
			Subtitle:       "Consultation report — shared via petsFollow",
			CTA:            "petsFollow helps professionals follow patients, message owners and centralise care.",
			Register:       "Create a pro account: ",
			Offer:          "Offer: ",
			ContactTitle:   "Sales contact",
			DefaultContact: "petsFollow",
			Phone:          "Phone ",
			Email:          "Email ",
			Pet:            "Pet",
			Visit:          "Visit",
			Practice:       "Practice",
			ReportBy:       "Report by ",
			Finalized:      "Finalized: ",
			NoBody:         "(empty)",
			Confidential:   "Confidential — for the intended professional recipient only.",
		}
	case "es":
		return labels{
			Subtitle:       "Informe de consulta — compartido vía petsFollow",
			CTA:            "petsFollow ayuda a los profesionales a seguir pacientes, hablar con dueños y centralizar cuidados.",
			Register:       "Crear cuenta pro: ",
			Offer:          "Oferta: ",
			ContactTitle:   "Contacto comercial",
			DefaultContact: "petsFollow",
			Phone:          "Tel. ",
			Email:          "Email ",
			Pet:            "Animal",
			Visit:          "Visita",
			Practice:       "Clínica",
			ReportBy:       "Informe de ",
			Finalized:      "Finalizado: ",
			NoBody:         "(vacío)",
			Confidential:   "Confidencial — solo para el profesional destinatario.",
		}
	case "et":
		return labels{
			Subtitle:       "Konsultatsiooni aruanne — jagatud petsFollowi kaudu",
			CTA:            "petsFollow aitab professionaalidel patsiente jälgida, omanikega suhelda ja hooldust tsentraliseerida.",
			Register:       "Loo pro konto: ",
			Offer:          "Pakkumine: ",
			ContactTitle:   "Müügikontakt",
			DefaultContact: "petsFollow",
			Phone:          "Tel. ",
			Email:          "E-post ",
			Pet:            "Loom",
			Visit:          "Külastus",
			Practice:       "Kliinik",
			ReportBy:       "Aruanne: ",
			Finalized:      "Lõpetatud: ",
			NoBody:         "(tühi)",
			Confidential:   "Konfidentsiaalne — ainult adressaadist professionaalile.",
		}
	case "it":
		return labels{
			Subtitle:       "Referto di consultazione — condiviso via petsFollow",
			CTA:            "petsFollow aiuta i professionisti a seguire i pazienti, comunicare con i proprietari e centralizzare le cure.",
			Register:       "Crea un account pro: ",
			Offer:          "Offerta: ",
			ContactTitle:   "Contatto commerciale",
			DefaultContact: "petsFollow",
			Phone:          "Tel. ",
			Email:          "Email ",
			Pet:            "Animale",
			Visit:          "Visita",
			Practice:       "Clinica",
			ReportBy:       "Referto di ",
			Finalized:      "Finalizzato: ",
			NoBody:         "(vuoto)",
			Confidential:   "Riservato — solo per il professionista destinatario.",
		}
	default:
		return labels{
			Subtitle:       "Compte-rendu de consultation — partagé via petsFollow",
			CTA:            "petsFollow aide les professionnels à suivre leurs patients, échanger avec les propriétaires et centraliser les soins.",
			Register:       "Créer un compte pro : ",
			Offer:          "Offre : ",
			ContactTitle:   "Contact commercial",
			DefaultContact: "petsFollow",
			Phone:          "Tél. ",
			Email:          "Email ",
			Pet:            "Animal",
			Visit:          "Visite",
			Practice:       "Cabinet",
			ReportBy:       "Compte-rendu par ",
			Finalized:      "Finalisé : ",
			NoBody:         "(vide)",
			Confidential:   "Confidentiel — destiné uniquement au professionnel destinataire.",
		}
	}
}

var (
	reMDBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reMDItalic  = regexp.MustCompile(`\*([^*\n]+)\*`)
	reMDHeading = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	reMDCode    = regexp.MustCompile("`+")
)

// StripMarkdown turns common CR markdown into plain text suitable for Helvetica PDF.
func StripMarkdown(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = reMDBold.ReplaceAllString(s, "$1")
	s = reMDItalic.ReplaceAllString(s, "$1")
	s = reMDHeading.ReplaceAllString(s, "")
	s = reMDCode.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "**", "")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "- "):
			lines[i] = "- " + strings.TrimSpace(trim[2:])
		case strings.HasPrefix(trim, "* "):
			lines[i] = "- " + strings.TrimSpace(trim[2:])
		case strings.HasPrefix(trim, "• "):
			lines[i] = "- " + strings.TrimSpace(trim[len("• "):])
		default:
			lines[i] = line
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// sanitizePDFText maps characters that break WinAnsi / some mobile PDF viewers.
func sanitizePDFText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\u2014', '\u2013', '\u2212':
			b.WriteByte('-')
		case '\u2018', '\u2019', '\u201A':
			b.WriteByte('\'')
		case '\u201C', '\u201D', '\u201E':
			b.WriteByte('"')
		case '\u2026':
			b.WriteString("...")
		case '\u00A0':
			b.WriteByte(' ')
		case '\u2022': // bullet — ASCII hyphen for mobile PDF viewers
			b.WriteByte('-')
		default:
			if r == '\n' || r == '\t' || unicode.IsPrint(r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func BuildPDF(in PDFInput) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Top margin clears the navy brand band (subtitle is page-1 body only).
	pdf.SetMargins(14, 34, 14)
	pdf.SetAutoPageBreak(true, 22)
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

	gen := in.GeneratedAt
	if gen.IsZero() {
		gen = time.Now().UTC()
	}

	if len(emblemPNG) > 0 {
		opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		_ = pdf.RegisterImageOptionsReader("pf-emblem", opt, bytes.NewReader(emblemPNG))
		if err := pdf.Error(); err != nil {
			return nil, fmt.Errorf("consultationpdf: register emblem: %w", err)
		}
	}

	pdf.SetHeaderFuncMode(func() {
		drawBrandBand(pdf, site)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-16)
		pdf.SetDrawColor(colorTeal[0], colorTeal[1], colorTeal[2])
		pdf.SetLineWidth(0.4)
		pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
		pdf.Ln(2)
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(colorMuted[0], colorMuted[1], colorMuted[2])
		pdf.CellFormat(120, 4, tr(fmt.Sprintf("petsFollow · %s", gen.Format("2006-01-02 15:04 UTC"))), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 4, fmt.Sprintf("%d", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(colorNavy[0], colorNavy[1], colorNavy[2])
	pdf.MultiCell(0, 5, tr(sanitizePDFText(L.Subtitle)), "", "L", false)
	pdf.Ln(2)

	pdf.SetFillColor(colorSurface[0], colorSurface[1], colorSurface[2])
	pdf.SetDrawColor(colorBorder[0], colorBorder[1], colorBorder[2])
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(colorBody[0], colorBody[1], colorBody[2])
	cta := sanitizePDFText(L.CTA)
	if register != "" {
		cta += "\n" + sanitizePDFText(L.Register) + register
	}
	if products != "" {
		cta += "\n" + sanitizePDFText(L.Offer) + products
	}
	pdf.MultiCell(0, 5, tr(cta), "1", "L", true)

	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(colorNavy[0], colorNavy[1], colorNavy[2])
	pdf.CellFormat(0, 6, tr(sanitizePDFText(L.ContactTitle)), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(colorBody[0], colorBody[1], colorBody[2])
	name := strings.TrimSpace(in.Marketing.CommercialName)
	if name == "" {
		name = L.DefaultContact
	}
	pdf.CellFormat(0, 5, tr(sanitizePDFText(name)), "", 1, "L", false, 0, "")
	if phone := strings.TrimSpace(in.Marketing.CommercialPhone); phone != "" {
		pdf.CellFormat(0, 5, tr(sanitizePDFText(L.Phone+phone)), "", 1, "L", false, 0, "")
	}
	if email := strings.TrimSpace(in.Marketing.CommercialEmail); email != "" {
		pdf.CellFormat(0, 5, tr(sanitizePDFText(L.Email+email)), "", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(colorMuted[0], colorMuted[1], colorMuted[2])
	pdf.MultiCell(0, 4, tr(sanitizePDFText(L.Confidential)), "", "L", false)
	pdf.Ln(3)

	section := func(title string) {
		pdf.Ln(1)
		pdf.SetFillColor(colorNavy[0], colorNavy[1], colorNavy[2])
		pdf.SetTextColor(colorWhite[0], colorWhite[1], colorWhite[2])
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(0, 8, "  "+tr(sanitizePDFText(title)), "", 1, "L", true, 0, "")
		pdf.SetDrawColor(colorTeal[0], colorTeal[1], colorTeal[2])
		pdf.SetLineWidth(0.8)
		pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
		pdf.SetLineWidth(0.2)
		pdf.Ln(3)
		pdf.SetTextColor(colorBody[0], colorBody[1], colorBody[2])
	}

	section(L.Pet)
	pdf.SetFont("Arial", "", 10)
	line := in.PetName
	if sp := strings.TrimSpace(in.Species); sp != "" {
		line += " · " + sp
	}
	if br := strings.TrimSpace(in.Breed); br != "" {
		line += " · " + br
	}
	pdf.MultiCell(0, 5, tr(sanitizePDFText(line)), "", "L", false)
	if owner := strings.TrimSpace(in.OwnerName); owner != "" {
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(colorMuted[0], colorMuted[1], colorMuted[2])
		pdf.MultiCell(0, 5, tr(sanitizePDFText(owner)), "", "L", false)
		pdf.SetTextColor(colorBody[0], colorBody[1], colorBody[2])
	}
	pdf.Ln(1)

	section(L.Visit)
	pdf.SetFont("Arial", "", 10)
	if when := strings.TrimSpace(in.VisitWhen); when != "" {
		pdf.MultiCell(0, 5, tr(sanitizePDFText(when)), "", "L", false)
	}
	if pr := strings.TrimSpace(in.PracticeName); pr != "" {
		pdf.MultiCell(0, 5, tr(sanitizePDFText(L.Practice+": "+pr)), "", "L", false)
	}
	pdf.Ln(1)

	for _, rep := range in.Reports {
		author := strings.TrimSpace(rep.AuthorName)
		title := L.ReportBy + author
		if author == "" {
			title = L.ReportBy + "-"
		}
		section(title)
		if fa := strings.TrimSpace(rep.FinalizedAt); fa != "" {
			pdf.SetFont("Arial", "I", 9)
			pdf.SetTextColor(colorMuted[0], colorMuted[1], colorMuted[2])
			pdf.MultiCell(0, 5, tr(sanitizePDFText(L.Finalized+fa)), "", "L", false)
			pdf.Ln(1)
			pdf.SetTextColor(colorBody[0], colorBody[1], colorBody[2])
		}
		pdf.SetFont("Arial", "", 10)
		body := StripMarkdown(strings.TrimSpace(rep.BodyText))
		if body == "" {
			body = L.NoBody
		}
		pdf.MultiCell(0, 5, tr(sanitizePDFText(body)), "", "L", false)
		pdf.Ln(2)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawBrandBand is the repeating header (logo + wordmark only — no marketing subtitle).
func drawBrandBand(pdf *gofpdf.Fpdf, site string) {
	pdf.SetFillColor(colorNavy[0], colorNavy[1], colorNavy[2])
	pdf.Rect(0, 0, 210, 28, "F")
	pdf.SetFillColor(colorTeal[0], colorTeal[1], colorTeal[2])
	pdf.Rect(0, 28, 210, 1.5, "F")

	if pdf.GetImageInfo("pf-emblem") != nil {
		opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		pdf.ImageOptions("pf-emblem", 14, 5, 16, 16, false, opt, 0, "")
	}

	pdf.SetTextColor(colorWhite[0], colorWhite[1], colorWhite[2])
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(34, 6)
	pdf.CellFormat(0, 8, "petsFollow", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.SetX(34)
	if site != "" {
		pdf.SetTextColor(180, 230, 220)
		pdf.CellFormat(0, 5, site, "", 1, "L", false, 0, site)
	}
}

