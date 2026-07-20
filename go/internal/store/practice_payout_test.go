package store

import "testing"

func TestIsVetPayoutProfileComplete(t *testing.T) {
	base := PracticeProfile{
		CompanyLegalName:      "Cabinet VetPlus SRL",
		VATNumber:             "BE0123456789",
		CompanyNumber:         "0123.456.789",
		LegalForm:             "srl",
		BillingSameAsPractice: true,
		AddressLine1:          "Rue Test 1",
		City:                  "Bruxelles",
		PostalCode:            "1000",
		PayoutIBAN:            "BE68539007547034",
		PayoutAccountHolder:   "Cabinet VetPlus",
	}
	if !IsVetPayoutProfileComplete(base) {
		t.Fatal("expected complete profile")
	}

	incomplete := base
	incomplete.PayoutIBAN = ""
	if IsVetPayoutProfileComplete(incomplete) {
		t.Fatal("expected incomplete without IBAN")
	}

	billing := base
	billing.BillingSameAsPractice = false
	billing.BillingAddressLine1 = ""
	if IsVetPayoutProfileComplete(billing) {
		t.Fatal("expected incomplete without billing address")
	}
	billing.BillingAddressLine1 = "Av. Facture 2"
	billing.BillingCity = "Liège"
	billing.BillingPostalCode = "4000"
	if !IsVetPayoutProfileComplete(billing) {
		t.Fatal("expected complete with distinct billing address")
	}
}
