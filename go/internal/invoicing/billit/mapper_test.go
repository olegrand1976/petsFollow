package billit_test

import (
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing/billit"
)

func TestMapDocumentCountries(t *testing.T) {
	base := invoicing.Document{
		Type:      invoicing.DocInvoice,
		CreatedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Currency:  "EUR",
		Lines: []invoicing.Line{{
			Description: "Consult", Quantity: 1, UnitPriceExclCents: 10000, VATPercent: 21,
		}},
	}

	t.Run("be", func(t *testing.T) {
		doc := base
		doc.Counterparty = invoicing.Counterparty{
			Name: "Vet BE", Country: "BE", VATNumber: "BE1000000021", CompanyNumber: "0123456789",
			Street: "Rue 1", City: "Bruxelles", Postal: "1000",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if ord.Customer.VATNumber != "BE1000000021" || ord.Customer.CountryCode != "BE" {
			t.Fatalf("%+v", ord.Customer)
		}
		if len(ord.OrderLines) != 1 || ord.OrderLines[0].UnitPriceExcl != 100 {
			t.Fatalf("%+v", ord.OrderLines)
		}
	})

	t.Run("fr", func(t *testing.T) {
		doc := base
		doc.Counterparty = invoicing.Counterparty{
			Name: "Clinique FR", Country: "FR", VATNumber: "FR12345678901", SIRET: "84249114400015",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if !hasID(ord, "SIRET", "84249114400015") || !hasID(ord, "SIREN", "842491144") {
			t.Fatalf("identifiers=%+v", ord.Customer.Identifiers)
		}
	})

	t.Run("it_codice", func(t *testing.T) {
		doc := base
		doc.Counterparty = invoicing.Counterparty{
			Name: "Clinica", Country: "IT", VATNumber: "IT01234567890", CodiceDestinatario: "3PRL4IK",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if !hasID(ord, "SDICODDEST", "3PRL4IK") {
			t.Fatalf("%+v", ord.Customer.Identifiers)
		}
	})

	t.Run("es", func(t *testing.T) {
		doc := base
		doc.Counterparty = invoicing.Counterparty{
			Name: "Clinica ES", Country: "ES", TaxID: "B12345678",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if ord.Customer.VATNumber != "B12345678" {
			t.Fatalf("vat=%q", ord.Customer.VATNumber)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		doc := base
		doc.Counterparty = invoicing.Counterparty{Name: "X", Country: "IT"}
		if _, err := billit.MapDocument(doc); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("credit_note_about_invoice", func(t *testing.T) {
		doc := base
		doc.Type = invoicing.DocCreditNote
		doc.AboutInvoiceNumber = "INV-100"
		doc.Counterparty = invoicing.Counterparty{
			Name: "Vet BE", Country: "BE", VATNumber: "BE1000000021",
			Street: "Rue 1", City: "Bruxelles", Postal: "1000",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if ord.OrderType != "CreditNote" {
			t.Fatalf("type=%s", ord.OrderType)
		}
		if ord.AboutInvoiceNumber != "INV-100" {
			t.Fatalf("AboutInvoiceNumber=%q", ord.AboutInvoiceNumber)
		}
	})

	t.Run("credit_note_omit_about_when_empty", func(t *testing.T) {
		doc := base
		doc.Type = invoicing.DocCreditNote
		doc.Counterparty = invoicing.Counterparty{
			Name: "Vet BE", Country: "BE", VATNumber: "BE1000000021",
			Street: "Rue 1", City: "Bruxelles", Postal: "1000",
		}
		ord, err := billit.MapDocument(doc)
		if err != nil {
			t.Fatal(err)
		}
		if ord.AboutInvoiceNumber != "" {
			t.Fatalf("want omit AboutInvoiceNumber got %q", ord.AboutInvoiceNumber)
		}
	})
}

func hasID(ord billit.OrderDTO, typ, val string) bool {
	for _, id := range ord.Customer.Identifiers {
		if id.IdentifierType == typ && id.Identifier == val {
			return true
		}
	}
	return false
}
