package billit

import (
	"fmt"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

// MapDocument transforms an internal document into a Billit OrderDTO.
func MapDocument(doc invoicing.Document) (OrderDTO, error) {
	cp := invoicing.NormalizeCounterparty(doc.Counterparty)
	if err := invoicing.ValidateCounterparty(cp); err != nil {
		return OrderDTO{}, err
	}
	orderType, err := mapOrderType(doc.Type)
	if err != nil {
		return OrderDTO{}, err
	}
	now := doc.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	orderDate := now.Format("2006-01-02")
	expiry := now.AddDate(0, 0, 30).Format("2006-01-02")

	ids := countryIdentifiers(cp)
	customer := CustomerDTO{
		Name:        cp.Name,
		VATNumber:   effectiveVAT(cp),
		PartyType:   "Customer",
		Email:       cp.Email,
		Street:      cp.Street,
		City:        cp.City,
		Zipcode:     cp.Postal,
		CountryCode: cp.Country,
		Identifiers: ids,
	}
	if cp.Street != "" || cp.City != "" || cp.Postal != "" {
		customer.Addresses = []Address{{
			AddressType: "InvoiceAddress",
			Name:        cp.Name,
			Street:      cp.Street,
			City:        cp.City,
			Zipcode:     cp.Postal,
			CountryCode: cp.Country,
		}}
	}

	lines := make([]OrderLine, 0, len(doc.Lines))
	for _, l := range doc.Lines {
		lines = append(lines, OrderLine{
			Quantity:      l.Quantity,
			UnitPriceExcl: float64(l.UnitPriceExclCents) / 100.0,
			Description:   l.Description,
			VATPercentage: l.VATPercent,
		})
	}
	currency := doc.Currency
	if currency == "" {
		currency = "EUR"
	}
	return OrderDTO{
		OrderType:      orderType,
		OrderDirection: "Income",
		OrderNumber:    doc.Number,
		OrderDate:      orderDate,
		ExpiryDate:     expiry,
		Currency:       currency,
		Customer:       customer,
		OrderLines:     lines,
	}, nil
}

func mapOrderType(t invoicing.DocType) (string, error) {
	switch t {
	case invoicing.DocInvoice:
		return "Invoice", nil
	case invoicing.DocCreditNote:
		return "CreditNote", nil
	case invoicing.DocProforma:
		return "ProForma", nil
	default:
		return "", fmt.Errorf("unsupported doc type %q", t)
	}
}

func effectiveVAT(cp invoicing.Counterparty) string {
	if cp.VATNumber != "" {
		return cp.VATNumber
	}
	if cp.Country == "ES" && cp.TaxID != "" {
		return cp.TaxID
	}
	return ""
}

func countryIdentifiers(cp invoicing.Counterparty) []Identifier {
	var out []Identifier
	switch cp.Country {
	case "BE":
		// VATNumber is enough for Peppol BE; CompanyNumber kept on domain for UI only.
	case "FR":
		if cp.SIRET != "" {
			out = append(out, Identifier{IdentifierType: "SIRET", Identifier: cp.SIRET})
		}
		if cp.SIREN != "" {
			out = append(out, Identifier{IdentifierType: "SIREN", Identifier: cp.SIREN})
		}
	case "IT":
		if cp.CodiceDestinatario != "" {
			out = append(out, Identifier{IdentifierType: "SDICODDEST", Identifier: cp.CodiceDestinatario})
		}
		if cp.PEC != "" {
			out = append(out, Identifier{IdentifierType: "SDIPEC", Identifier: cp.PEC})
		}
	case "ES":
		if cp.TaxID != "" && !strings.EqualFold(cp.TaxID, cp.VATNumber) {
			out = append(out, Identifier{IdentifierType: "VAT", Identifier: cp.TaxID})
		}
	}
	return out
}
