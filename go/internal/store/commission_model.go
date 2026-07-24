package store

// PlanRateInfo is an indicative commission row for UI / fiches (HTVA base).
type PlanRateInfo struct {
	Code              string `json:"code"`
	TTCCents          int    `json:"ttcCents"`
	HTVACents         int    `json:"htvaCents"`
	VetRateBpsMax     int    `json:"vetRateBpsMax"`
	VetCentsMax       int    `json:"vetCentsMax"`
	CommercialRateBps int    `json:"commercialRateBps"`
	CommercialCents   int    `json:"commercialCents"`
	Recommended       bool   `json:"recommended"`
}

// BonusRule describes a SPIFF / behavioural bonus (V1: display + optional progress).
type BonusRule struct {
	Code        string `json:"code"`
	Audience    string `json:"audience"` // vet | commercial
	AmountCents int    `json:"amountCents"`
	TitleKey    string `json:"titleKey"`
	Status      string `json:"status,omitempty"`   // available | in_progress | earned | paid
	Progress    *int   `json:"progress,omitempty"` // e.g. pets count or mix %
	Target      *int   `json:"target,omitempty"`
	AwardID     string `json:"awardId,omitempty"`
	VetUserID   string `json:"vetUserId,omitempty"`
	VetEmail    string `json:"vetEmail,omitempty"`
	VetFullName string `json:"vetFullName,omitempty"`
	PeriodYM    string `json:"periodYm,omitempty"`
}

// catalogTTC mirrors go/internal/billing/domain.go sellable plan amounts (avoid import cycle).
var catalogSubscriptions = []struct {
	code string
	ttc  int
	rec  bool
}{
	{"monthly", 350, false},
	{"annual", 3500, false},
	{"triennial", 9500, true},
}

// SubscriptionPlanRates returns indicative rates for monthly / annual / triennial.
func SubscriptionPlanRates() []PlanRateInfo {
	out := make([]PlanRateInfo, 0, len(catalogSubscriptions))
	for _, p := range catalogSubscriptions {
		ht := HTVACents(p.ttc)
		vetBps := ApplyVetPlanFactor(MaxVetCommissionBps, p.code)
		commBps := CommercialRateBpsForPlan(p.code)
		out = append(out, PlanRateInfo{
			Code:              p.code,
			TTCCents:          p.ttc,
			HTVACents:         ht,
			VetRateBpsMax:     vetBps,
			VetCentsMax:       CommercialCommissionCents(ht, vetBps),
			CommercialRateBps: commBps,
			CommercialCents:   CommercialCommissionCents(ht, commBps),
			Recommended:       p.rec,
		})
	}
	return out
}

// AddonPlanRates returns an empty sellable catalog (addons no longer sold).
// Legacy rates remain available via CommercialRateBpsForAddon / VetRateBpsForAddon.
func AddonPlanRates() []PlanRateInfo {
	return nil
}

// DefaultBonusRules returns static SPIFF definitions (progress filled by callers).
func DefaultBonusRules() []BonusRule {
	return []BonusRule{
		{Code: "vet_tier_31", Audience: "vet", AmountCents: 5000, TitleKey: "bonus.vetTier31", Status: "available"},
		{Code: "commercial_ramp", Audience: "commercial", AmountCents: 2500, TitleKey: "bonus.commercialRamp", Status: "available"},
		{Code: "commercial_mix", Audience: "commercial", AmountCents: 5000, TitleKey: "bonus.commercialMix", Status: "available"},
	}
}
