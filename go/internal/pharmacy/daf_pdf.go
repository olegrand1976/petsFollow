package pharmacy

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

// DAFPDFInput is the printable DAF content (lot + AMM required on each line).
type DAFPDFInput struct {
	DisplayNumber    string
	PracticeName     string
	Prescriber       string
	ClientName       string
	PetName          string
	FoodChainStatus  string // companion | food_producing | excluded_from_food_chain
	DomicileLocation string
	IssuedAt         time.Time
	Notes            string
	Lines            []DAFPDFLine
}

type DAFPDFLine struct {
	Medication         string
	CNK                string
	AMM                string
	Lot                string
	ExpiresOn          string
	Qty                string
	Unit               string
	Antibiotic         bool
	WithdrawalMeatDays *int
	WithdrawalMilkDays *int
	WithdrawalEggsDays *int
}

// BuildDAFPDF renders a simple AFMPS-oriented DAF PDF.
func BuildDAFPDF(in DAFPDFInput) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(in.DisplayNumber, false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Document d'Administration et de Fourniture (DAF)")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("Numero : %s", in.DisplayNumber))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Cabinet : %s", in.PracticeName))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Prescripteur : %s", in.Prescriber))
	pdf.Ln(6)
	if in.ClientName != "" {
		pdf.Cell(0, 7, fmt.Sprintf("Client : %s", in.ClientName))
		pdf.Ln(6)
	}
	if in.PetName != "" {
		pdf.Cell(0, 7, fmt.Sprintf("Animal : %s", in.PetName))
		pdf.Ln(6)
	}
	if in.DomicileLocation != "" {
		pdf.MultiCell(0, 5, fmt.Sprintf("Domicile / ecurie : %s", in.DomicileLocation), "", "", false)
		pdf.Ln(2)
	}
	if in.FoodChainStatus != "" {
		label := in.FoodChainStatus
		switch in.FoodChainStatus {
		case "companion":
			label = "animal de compagnie (hors chaine alimentaire)"
		case "food_producing":
			label = "animal de rente / chaine alimentaire"
		case "excluded_from_food_chain":
			label = "exclu de la chaine alimentaire"
		}
		pdf.MultiCell(0, 5, fmt.Sprintf("Statut chaine alimentaire : %s", label), "", "", false)
		pdf.Ln(2)
	}
	pdf.Cell(0, 7, fmt.Sprintf("Emis le : %s", in.IssuedAt.Format("2006-01-02 15:04")))
	pdf.Ln(10)
	if in.Notes != "" {
		pdf.MultiCell(0, 5, "Notes : "+in.Notes, "", "", false)
		pdf.Ln(4)
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(45, 7, "Medicament", "1", 0, "", false, 0, "")
	pdf.CellFormat(22, 7, "AMM", "1", 0, "", false, 0, "")
	pdf.CellFormat(22, 7, "Lot", "1", 0, "", false, 0, "")
	pdf.CellFormat(22, 7, "DLC", "1", 0, "", false, 0, "")
	pdf.CellFormat(25, 7, "Qte", "1", 0, "", false, 0, "")
	pdf.CellFormat(18, 7, "AB", "1", 0, "", false, 0, "")
	pdf.CellFormat(36, 7, "Attente V/L/O", "1", 1, "", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	for _, l := range in.Lines {
		name := l.Medication
		if l.CNK != "" {
			name = name + " (" + l.CNK + ")"
		}
		ab := ""
		if l.Antibiotic {
			ab = "oui"
		}
		pdf.CellFormat(45, 7, truncatePDF(name, 28), "1", 0, "", false, 0, "")
		pdf.CellFormat(22, 7, truncatePDF(l.AMM, 12), "1", 0, "", false, 0, "")
		pdf.CellFormat(22, 7, truncatePDF(l.Lot, 12), "1", 0, "", false, 0, "")
		pdf.CellFormat(22, 7, l.ExpiresOn, "1", 0, "", false, 0, "")
		pdf.CellFormat(25, 7, l.Qty+" "+l.Unit, "1", 0, "", false, 0, "")
		pdf.CellFormat(18, 7, ab, "1", 0, "", false, 0, "")
		pdf.CellFormat(36, 7, formatWithdrawal(l), "1", 1, "", false, 0, "")
	}
	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(0, 4, "Mentions AFMPS : numero de lot et numero d'AMM obligatoires. Temps d'attente V/L/O = viande / lait / oeufs (jours). DLC indiquee a titre informatif. Document genere par petsFollow — ne remplace pas un logiciel DAF certifie.", "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncatePDF(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "."
}

func formatWithdrawal(l DAFPDFLine) string {
	parts := make([]string, 0, 3)
	if l.WithdrawalMeatDays != nil {
		parts = append(parts, fmt.Sprintf("V:%d", *l.WithdrawalMeatDays))
	} else {
		parts = append(parts, "V:-")
	}
	if l.WithdrawalMilkDays != nil {
		parts = append(parts, fmt.Sprintf("L:%d", *l.WithdrawalMilkDays))
	} else {
		parts = append(parts, "L:-")
	}
	if l.WithdrawalEggsDays != nil {
		parts = append(parts, fmt.Sprintf("O:%d", *l.WithdrawalEggsDays))
	} else {
		parts = append(parts, "O:-")
	}
	return strings.Join(parts, " ")
}
