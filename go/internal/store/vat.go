package store

// DefaultVATRateBps is the Belgian standard VAT rate (21%) used to derive HTVA
// from client TTC amounts before applying partner commission rates.
const DefaultVATRateBps = 2100

// HTVACents converts a TTC amount in cents to HTVA cents (integer division).
// Negative inputs are clamped to 0.
func HTVACents(ttcCents int) int {
	if ttcCents < 0 {
		return 0
	}
	return ttcCents * 10000 / (10000 + DefaultVATRateBps)
}

// CommissionFromTTCCents returns commission cents on the HTVA base of a TTC amount.
func CommissionFromTTCCents(ttcCents, rateBps int) int {
	return CommercialCommissionCents(HTVACents(ttcCents), rateBps)
}
