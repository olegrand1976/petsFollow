package invoicing_test

import (
	"errors"
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
			c:    invoicing.Counterparty{Name: "Cabinet Vet", Country: "be", VATNumber: "BE0123456789"},
		},
		{
			name:    "be_missing",
			c:       invoicing.Counterparty{Name: "X", Country: "BE"},
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
