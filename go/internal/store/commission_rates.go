package store

// Plan commission factors (percent of vet base tier rate), applied before the 5–8% clamp.
const (
	PlanFactorMonthlyPct      = 67
	PlanFactorAnnualPct       = 67
	PlanFactorTriennialPct    = 100
	PlanFactorQuinquennialPct = 67 // legacy grandfathered
	MinVetCommissionBps       = 500
	MaxVetCommissionBps       = 800
	// MinVetBaseTierBps is the lowest base rate that still yields ≥5% after ×0.67.
	MinVetBaseTierBps = 747
)

// Default commercial rates by plan/addon (bps of HTVA).
const (
	CommercialRateMonthlyBps      = 800
	CommercialRateAnnualBps       = 800
	CommercialRateTriennialBps    = 1200
	CommercialRateQuinquennialBps = 800  // legacy grandfathered
	CommercialRateAddonBps        = 1000 // legacy grandfathered
	// VetAddonRateBps is the flat vet rate on Family / Kennel (Care+/Horse = 0).
	VetAddonRateBps = 500
)

// DefaultVetCommissionTiers is the progressive ladder (7.50% → 8% base; effective 5–8% after plan factor).
func DefaultVetCommissionTiers() []CommissionTier {
	max10 := 10
	max30 := 30
	max60 := 60
	return []CommissionTier{
		{MinClients: 1, MaxClients: &max10, RateBps: 750},
		{MinClients: 11, MaxClients: &max30, RateBps: 767},
		{MinClients: 31, MaxClients: &max60, RateBps: 783},
		{MinClients: 61, MaxClients: nil, RateBps: 800},
	}
}

// VetPlanFactorPct returns the plan multiplier percent (monthly/annual ×0.67 · triennial ×1.00).
func VetPlanFactorPct(planCode string) int {
	switch planCode {
	case "triennial":
		return PlanFactorTriennialPct
	case "monthly":
		return PlanFactorMonthlyPct
	case "quinquennial":
		return PlanFactorQuinquennialPct
	default:
		return PlanFactorAnnualPct
	}
}

// ApplyVetPlanFactor scales a progressive tier rate by plan factor and clamps to [5%, 8%].
func ApplyVetPlanFactor(baseRateBps int, planCode string) int {
	if baseRateBps < 0 {
		baseRateBps = 0
	}
	factor := VetPlanFactorPct(planCode)
	rate := baseRateBps * factor / 100
	if rate < MinVetCommissionBps {
		return MinVetCommissionBps
	}
	if rate > MaxVetCommissionBps {
		return MaxVetCommissionBps
	}
	return rate
}

// CommercialRateBpsForPlan returns the commercial commission rate for a subscription plan.
func CommercialRateBpsForPlan(planCode string) int {
	switch planCode {
	case "triennial":
		return CommercialRateTriennialBps
	case "monthly":
		return CommercialRateMonthlyBps
	case "quinquennial":
		return CommercialRateQuinquennialBps
	default:
		return CommercialRateAnnualBps
	}
}

// CommercialRateBpsForAddon returns the commercial commission rate for an addon (legacy).
func CommercialRateBpsForAddon(_ string) int {
	return CommercialRateAddonBps
}

// VetRateBpsForAddon returns the vet commission rate for an addon (0 if none; legacy).
func VetRateBpsForAddon(addonCode string) int {
	switch addonCode {
	case "family", "kennel":
		return VetAddonRateBps
	default:
		return 0
	}
}
