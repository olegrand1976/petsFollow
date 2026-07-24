package store

import (
	"strings"
	"testing"
)

func TestParseStripeAmountCents(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want int
	}{
		{"3,50", 350},
		{"35,00", 3500},
		{"95,00", 9500},
		{"3.50", 350},
		{"350", 350}, // bare integer = cents (Stripe unit_amount)
		{"€3,50", 350},
	}
	for _, tc := range cases {
		got, err := parseStripeAmountCents(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %d want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseStripeAmountCentsErrors(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "   ", "abc", "€", "—"} {
		if _, err := parseStripeAmountCents(in); err == nil {
			t.Fatalf("%q: expected error", in)
		}
	}
}

func TestInferStripePlanMapping(t *testing.T) {
	t.Parallel()
	cases := []struct {
		interval string
		count    int
		amount   int
		name     string
		slug     string
		plan     string
		mode     string
	}{
		{"month", 1, 350, "Mensualy reccurent", "", "monthly", "subscription"},
		{"year", 1, 3500, "Annual recurrent", "", "annual", "subscription"},
		{"year", 3, 9500, "3 years recurrent", "", "triennial", "subscription"},
		{"year", 5, 14500, "5 years", "", "quinquennial", "subscription"},
		{"", 0, 14500, "Quinquennial one-time", "", "quinquennial", "one_time"},
		{"", 0, 3500, "Annual one-time", "Annual one-time", "annual", "one_time"},
		{"", 0, 9500, "3 years one-time", "Triennial one-time", "triennial", "one_time"},
		{"", 0, 999, "Monthly reccurent petsFollow", "", "monthly", "subscription"},
		{"", 0, 0, "Unknown product", "", "", ""},
	}
	for _, tc := range cases {
		plan, mode := InferStripePlanMapping(tc.interval, tc.count, tc.amount, tc.name, tc.slug)
		if plan != tc.plan || mode != tc.mode {
			t.Fatalf("%s: got %s/%s want %s/%s", tc.name, plan, mode, tc.plan, tc.mode)
		}
	}
}

func TestParseStripeProductsCSV(t *testing.T) {
	t.Parallel()
	csv := `"id","Name","Date (UTC)","Description","Url","Tax Code","plan_slug (metadata)"
"prod_UwYk7HoNdxO5pK","Mensualy reccurent",2026-07-24 09:42:00,"Mensualy reccurent petsFollow",,"txcd_10000000",
"prod_UflpZS9LKSOh5e","Annual one-time",2026-06-09 14:05:00,"Annual one-time petsFollow",,"txcd_10000000","Annual one-time"
`
	products, err := ParseStripeProductsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("len=%d", len(products))
	}
	if products[0].StripeProductID != "prod_UwYk7HoNdxO5pK" {
		t.Fatalf("id=%s", products[0].StripeProductID)
	}
	if products[1].MetadataPlanSlug != "Annual one-time" {
		t.Fatalf("slug=%q", products[1].MetadataPlanSlug)
	}
}

func TestParseStripePricesCSV(t *testing.T) {
	t.Parallel()
	csv := `"Price ID","Product ID","Product Name","Product Statement Descriptor","Product Tax Code","Description","Created (UTC)","Amount","Currency","Interval","Interval Count","Usage Type","Aggregate Usage","Billing Scheme","Trial Period Days","Tax Behavior","plan_slug (metadata)","interval (metadata)"
"price_1Twfd5RYOt1sDKOD4Aw5Bd3G","prod_UwYk7HoNdxO5pK","Mensualy reccurent",,"txcd_10000000",,2026-07-24 09:42:00,"3,50","eur","month",1,"licensed",,"per_unit",,"unspecified",,
"price_1TtqcZRYOt1sDKODosBdjKCU","prod_UflpZS9LKSOh5e","Annual one-time","petsFollow Annual 1x","txcd_10000000","petsFollow Abonnement annuel (non récurrent)",2026-07-16 14:50:00,"35,00","eur",,,,,"per_unit",,"inclusive",,
`
	prices, err := ParseStripePricesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 2 {
		t.Fatalf("len=%d", len(prices))
	}
	if prices[0].AmountCents != 350 {
		t.Fatalf("amount=%d", prices[0].AmountCents)
	}
	if prices[0].PlanCode == nil || *prices[0].PlanCode != "monthly" {
		t.Fatalf("plan=%v", prices[0].PlanCode)
	}
	if prices[0].BillingMode == nil || *prices[0].BillingMode != "subscription" {
		t.Fatalf("mode=%v", prices[0].BillingMode)
	}
	if prices[1].PlanCode == nil || *prices[1].PlanCode != "annual" {
		t.Fatalf("plan=%v", prices[1].PlanCode)
	}
	if prices[1].BillingMode == nil || *prices[1].BillingMode != "one_time" {
		t.Fatalf("mode=%v", prices[1].BillingMode)
	}
}

func TestParseStripeProductsCSVEmpty(t *testing.T) {
	t.Parallel()
	products, err := ParseStripeProductsCSV(strings.NewReader(`"id","Name"
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 0 {
		t.Fatalf("len=%d", len(products))
	}
}

func TestParseStripePricesCSVInvalidAmount(t *testing.T) {
	t.Parallel()
	csv := `"Price ID","Product ID","Amount"
"price_bad","prod_x","not-a-price"
`
	if _, err := ParseStripePricesCSV(strings.NewReader(csv)); err == nil {
		t.Fatal("expected invalid amount error")
	}
}

func TestParseStripeCSVMalformed(t *testing.T) {
	t.Parallel()
	if _, err := ParseStripeProductsCSV(strings.NewReader("")); err == nil {
		t.Fatal("expected error on empty reader")
	}
}
