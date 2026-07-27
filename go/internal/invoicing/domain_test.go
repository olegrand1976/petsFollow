package invoicing_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

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
