package store

// Plan commission factors (percent of vet base tier rate), applied before the 12% cap.
const (
	PlanFactorAnnualPct       = 67
	PlanFactorTriennialPct    = 100
	PlanFactorQuinquennialPct = 67
	MaxVetCommissionBps       = 1200
)

// Default commercial rates by plan/addon (bps of HTVA).
const (
	CommercialRateAnnualBps       = 800
	CommercialRateTriennialBps    = 1200
	CommercialRateQuinquennialBps = 800
	CommercialRateAddonBps        = 1000
	// VetAddonRateBps is the flat vet rate on Family / Kennel (Care+/Horse = 0).
	VetAddonRateBps = 500
)

// DefaultVetCommissionTiers is the progressive ladder (7% → 9% → 11% → 12%).
func DefaultVetCommissionTiers() []CommissionTier {
	max10 := 10
	max30 := 30
	max60 := 60
	return []CommissionTier{
		{MinClients: 1, MaxClients: &max10, RateBps: 700},
		{MinClients: 11, MaxClients: &max30, RateBps: 900},
		{MinClients: 31, MaxClients: &max60, RateBps: 1100},
		{MinClients: 61, MaxClients: nil, RateBps: 1200},
	}
}

// VetPlanFactorPct returns the plan multiplier percent (67 / 100 / 67).
func VetPlanFactorPct(planCode string) int {
	switch planCode {
	case "triennial":
		return PlanFactorTriennialPct
	case "quinquennial":
		return PlanFactorQuinquennialPct
	default:
		return PlanFactorAnnualPct
	}
}

// ApplyVetPlanFactor scales a progressive tier rate by plan factor and caps at 12%.
func ApplyVetPlanFactor(baseRateBps int, planCode string) int {
	if baseRateBps < 0 {
		baseRateBps = 0
	}
	factor := VetPlanFactorPct(planCode)
	rate := baseRateBps * factor / 100
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
	case "quinquennial":
		return CommercialRateQuinquennialBps
	default:
		return CommercialRateAnnualBps
	}
}

// CommercialRateBpsForAddon returns the commercial commission rate for an addon.
func CommercialRateBpsForAddon(_ string) int {
	return CommercialRateAddonBps
}

// VetRateBpsForAddon returns the vet commission rate for an addon (0 if none).
func VetRateBpsForAddon(addonCode string) int {
	switch addonCode {
	case "family", "kennel":
		return VetAddonRateBps
	default:
		return 0
	}
}
