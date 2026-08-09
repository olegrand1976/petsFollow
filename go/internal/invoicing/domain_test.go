package invoicing_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

func TestCountryFromVAT(t *testing.T) {
	cases := []struct {
		vat, want string
	}{
		{"BE0123456789", "BE"},
		{" be0123456789 ", "BE"},
		{"FR12345678901", "FR"},
		{"0123456789", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := invoicing.CountryFromVAT(tc.vat); got != tc.want {
			t.Fatalf("CountryFromVAT(%q)=%q want %q", tc.vat, got, tc.want)
		}
	}
}

func TestComputeTotals(t *testing.T) {
	excl, vat, incl := invoicing.ComputeTotals([]invoicing.Line{
		{Quantity: 1, UnitPriceExclCents: 10000, VATPercent: 21},
		{Quantity: 2, UnitPriceExclCents: 500, VATPercent: 6},
	})
	if excl != 11000 {
		t.Fatalf("excl=%d", excl)
	}
	if vat != 2100+60 {
		t.Fatalf("vat=%d", vat)
	}
	if incl != excl+vat {
		t.Fatalf("incl=%d", incl)
	}
}

func TestSealOpenPlain(t *testing.T) {
	ref, err := invoicing.SealAPIKey("plain_dev", "", "secret-key")
	if err != nil {
		t.Fatal(err)
	}
	got, err := invoicing.OpenAPIKey("plain_dev", "", ref)
	if err != nil || got != "secret-key" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestSealOpenEnc(t *testing.T) {
	ref, err := invoicing.SealAPIKey("local_enc", "test-material-32chars-minimum!!", "secret-key")
	if err != nil {
		t.Fatal(err)
	}
	got, err := invoicing.OpenAPIKey("local_enc", "test-material-32chars-minimum!!", ref)
	if err != nil || got != "secret-key" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestValidateCounterpartyMultiCountry(t *testing.T) {
	cases := []struct {
		name    string
		c       invoicing.Counterparty
		wantErr bool
	}{
		{
			name: "be_ok",
			c: invoicing.Counterparty{
				Name: "Cabinet Vet", Country: "be", VATNumber: "BE1000000021",
				Street: "Rue 1", City: "Bruxelles", Postal: "1000",
			},
		},
		{
			name:    "be_missing",
			c:       invoicing.Counterparty{Name: "X", Country: "BE"},
			wantErr: true,
		},
		{
			name: "be_missing_address",
			c: invoicing.Counterparty{
				Name: "Cabinet Vet", Country: "BE", VATNumber: "BE1000000021",
			},
			wantErr: true,
		},
		{
			name: "be_bad_checksum",
			c: invoicing.Counterparty{
				Name: "Cabinet Vet", Country: "BE", VATNumber: "BE0123456789",
				Street: "Rue 1", City: "Bruxelles", Postal: "1000",
			},
			wantErr: true,
		},
		{
			name: "fr_siret",
			c: invoicing.Counterparty{
				Name: "Clinique", Country: "FR", VATNumber: "FR12345678901", SIRET: "84249114400015",
			},
		},
		{
			name:    "fr_missing_vat",
			c:       invoicing.Counterparty{Name: "Clinique", Country: "FR", SIRET: "84249114400015"},
			wantErr: true,
		},
		{
			name:    "fr_bad_siret",
			c:       invoicing.Counterparty{Name: "Clinique", Country: "FR", VATNumber: "FR12345678901", SIRET: "123"},
			wantErr: true,
		},
		{
			name:    "be_missing_vat",
			c:       invoicing.Counterparty{Name: "X", Country: "BE", CompanyNumber: "0123456789"},
			wantErr: true,
		},
		{
			name: "it_codice",
			c: invoicing.Counterparty{
				Name: "Clinica", Country: "IT", VATNumber: "IT01234567890", CodiceDestinatario: "3PRL4IK",
			},
		},
		{
			name: "it_pec",
			c: invoicing.Counterparty{
				Name: "Clinica", Country: "IT", VATNumber: "IT01234567890", PEC: "info@pec.example.it",
			},
		},
		{
			name: "it_both",
			c: invoicing.Counterparty{
				Name: "Clinica", Country: "IT", VATNumber: "IT01234567890",
				CodiceDestinatario: "3PRL4IK", PEC: "info@pec.example.it",
			},
			wantErr: true,
		},
		{
			name: "es_nif",
			c:    invoicing.Counterparty{Name: "Clinica ES", Country: "ES", TaxID: "B12345678"},
		},
		{
			name:    "country_missing",
			c:       invoicing.Counterparty{Name: "X"},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := invoicing.ValidateCounterparty(tc.c)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestValidateLines(t *testing.T) {
	ok := []invoicing.Line{{Description: "Consult", Quantity: 1, UnitPriceExclCents: 1000, VATPercent: 21}}
	if err := invoicing.ValidateLines(ok); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		line invoicing.Line
	}{
		{"empty_desc", invoicing.Line{Description: " ", Quantity: 1, UnitPriceExclCents: 1000, VATPercent: 21}},
		{"qty_zero", invoicing.Line{Description: "X", Quantity: 0, UnitPriceExclCents: 1000, VATPercent: 21}},
		{"price_zero", invoicing.Line{Description: "X", Quantity: 1, UnitPriceExclCents: 0, VATPercent: 21}},
		{"vat_over", invoicing.Line{Description: "X", Quantity: 1, UnitPriceExclCents: 1000, VATPercent: 101}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := invoicing.ValidateLines([]invoicing.Line{tc.line})
			if err == nil || !errors.Is(err, invoicing.ErrInvalidLines) {
				t.Fatalf("want ErrInvalidLines got %v", err)
			}
		})
	}
}

func TestResolveWebhookApplyStatus(t *testing.T) {
	cases := []struct {
		prev, in, want invoicing.DocStatus
	}{
		{invoicing.StatusSending, invoicing.StatusDelivered, invoicing.StatusDelivered},
		{invoicing.StatusDelivered, invoicing.StatusSending, invoicing.StatusDelivered},
		{invoicing.StatusDelivered, invoicing.StatusRejected, invoicing.StatusDelivered},
		{invoicing.StatusRejected, invoicing.StatusSending, invoicing.StatusRejected},
		{invoicing.StatusRejected, invoicing.StatusDelivered, invoicing.StatusDelivered},
		{invoicing.StatusCancelled, invoicing.StatusSending, invoicing.StatusCancelled},
		{invoicing.StatusSending, invoicing.StatusRejected, invoicing.StatusRejected},
	}
	for _, tc := range cases {
		got := invoicing.ResolveWebhookApplyStatus(tc.prev, tc.in)
		if got != tc.want {
			t.Fatalf("%s+%s => %s want %s", tc.prev, tc.in, got, tc.want)
		}
	}
}

// Un particulier n'a ni TVA ni identifiant fiscal : la validation doit exiger
// l'adresse (mention légale) et l'email (canal de livraison SMTP).
func TestValidateCounterpartyIndividual(t *testing.T) {
	base := invoicing.Counterparty{
		Name: "Marie Dupont", CustomerKind: invoicing.KindIndividual, Country: "BE",
		Email: "marie@example.test", Street: "Rue 1", City: "Bruxelles", Postal: "1000",
	}
	if err := invoicing.ValidateCounterparty(invoicing.NormalizeCounterparty(base)); err != nil {
		t.Fatalf("valid individual rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(c *invoicing.Counterparty)
		want   string
	}{
		{"sans email", func(c *invoicing.Counterparty) { c.Email = "" }, "individual_email_required"},
		{"email sans @", func(c *invoicing.Counterparty) { c.Email = "marie" }, "individual_email_required"},
		{"sans rue", func(c *invoicing.Counterparty) { c.Street = "" }, "individual_address_required"},
		{"sans ville", func(c *invoicing.Counterparty) { c.City = "" }, "individual_address_required"},
		{"sans code postal", func(c *invoicing.Counterparty) { c.Postal = "" }, "individual_address_required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			tc.mutate(&c)
			err := invoicing.ValidateCounterparty(invoicing.NormalizeCounterparty(c))
			if err == nil || !errors.Is(err, invoicing.ErrInvalidCounterparty) {
				t.Fatalf("want ErrInvalidCounterparty got %v", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want reason %s got %v", tc.want, err)
			}
		})
	}

	// Une TVA saisie par erreur sur un particulier est effacée, pas propagée.
	dirty := base
	dirty.VATNumber = "BE1000000021"
	dirty.SIRET = "84249114400015"
	dirty.CodiceDestinatario = "3PRL4IK"
	clean := invoicing.NormalizeCounterparty(dirty)
	if clean.VATNumber != "" || clean.SIRET != "" || clean.SIREN != "" || clean.CodiceDestinatario != "" {
		t.Fatalf("individual must carry no fiscal id: %+v", clean)
	}
}

// Le défaut reste professionnel : une contrepartie sans kind ne doit pas
// basculer silencieusement en email.
func TestNormalizeCounterpartyDefaultsToBusiness(t *testing.T) {
	c := invoicing.NormalizeCounterparty(invoicing.Counterparty{
		Name: "Clinique", Country: "be", VATNumber: "BE1000000021",
	})
	if c.CustomerKind != invoicing.KindBusiness {
		t.Fatalf("kind=%q want business", c.CustomerKind)
	}
	if c.VATNumber != "BE1000000021" {
		t.Fatalf("business VAT must be preserved: %+v", c)
	}
}

// Une valeur inconnue ne doit pas ouvrir le chemin B2C (facture sans TVA
// envoyée par email) sur une faute de frappe côté client.
func TestNormalizeCounterpartyUnknownKindFallsBackToBusiness(t *testing.T) {
	c := invoicing.NormalizeCounterparty(invoicing.Counterparty{
		Name: "Clinique", Country: "BE", VATNumber: "BE1000000021", CustomerKind: "particulier",
	})
	if c.CustomerKind != invoicing.KindBusiness {
		t.Fatalf("kind=%q want business", c.CustomerKind)
	}
	if c.VATNumber != "BE1000000021" {
		t.Fatalf("unknown kind must not strip the VAT: %+v", c)
	}
	if got := invoicing.ResolveTransport(c); got != invoicing.TransportPeppol {
		t.Fatalf("transport=%q want Peppol", got)
	}
}

// Billit renvoie les mêmes libellés de statut quel que soit le transport : sans
// ce garde-fou, le webhook d'un envoi email afficherait « delivered » (Peppol).
func TestResolveWebhookPeppolStatus(t *testing.T) {
	cases := []struct{ prev, incoming, want string }{
		{"email_sending", "delivered", "email_delivered"},
		{"email_sending", "rejected", "email_rejected"},
		{"email_delivered", "email_delivered", "email_delivered"},
		{"sending", "delivered", "delivered"},
		{"", "delivered", "delivered"},
		{"email_sending", "", ""},
	}
	for _, tc := range cases {
		if got := invoicing.ResolveWebhookPeppolStatus(tc.prev, tc.incoming); got != tc.want {
			t.Fatalf("prev=%q incoming=%q → %q want %q", tc.prev, tc.incoming, got, tc.want)
		}
	}
}

func TestResolveTransport(t *testing.T) {
	cases := []struct {
		name string
		cp   invoicing.Counterparty
		want invoicing.Transport
	}{
		{"particulier BE", invoicing.Counterparty{CustomerKind: invoicing.KindIndividual, Country: "BE"}, invoicing.TransportSMTP},
		{"particulier IT", invoicing.Counterparty{CustomerKind: invoicing.KindIndividual, Country: "IT"}, invoicing.TransportSMTP},
		{"pro BE", invoicing.Counterparty{CustomerKind: invoicing.KindBusiness, Country: "BE"}, invoicing.TransportPeppol},
		{"pro IT", invoicing.Counterparty{CustomerKind: invoicing.KindBusiness, Country: "it"}, invoicing.TransportSDI},
		{"kind absent", invoicing.Counterparty{Country: "FR"}, invoicing.TransportPeppol},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := invoicing.ResolveTransport(tc.cp); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
	if !invoicing.TransportPeppol.IsEInvoiceNetwork() || !invoicing.TransportSDI.IsEInvoiceNetwork() {
		t.Fatal("Peppol/SDI must be network transports")
	}
	if invoicing.TransportSMTP.IsEInvoiceNetwork() {
		t.Fatal("SMTP must not be treated as an e-invoice network")
	}
}
