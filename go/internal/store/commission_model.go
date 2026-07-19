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

// catalogTTC mirrors go/internal/billing/domain.go plan/addon amounts (avoid import cycle).
var catalogSubscriptions = []struct {
	code string
	ttc  int
	rec  bool
}{
	{"annual", 3500, false},
	{"triennial", 9500, true},
	{"quinquennial", 14500, false},
}

var catalogAddons = []struct {
	code string
	ttc  int
}{
	{"family", 3900},
	{"kennel", 11900},
	{"care_plus", 1900},
	{"horse", 3900},
}

// SubscriptionPlanRates returns indicative rates for annual / triennial / quinquennial.
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

// AddonPlanRates returns indicative commercial (+ vet when applicable) rates for addons.
func AddonPlanRates() []PlanRateInfo {
	out := make([]PlanRateInfo, 0, len(catalogAddons))
	for _, a := range catalogAddons {
		ht := HTVACents(a.ttc)
		commBps := CommercialRateBpsForAddon(a.code)
		vetBps := VetRateBpsForAddon(a.code)
		out = append(out, PlanRateInfo{
			Code:              a.code,
			TTCCents:          a.ttc,
			HTVACents:         ht,
			VetRateBpsMax:     vetBps,
			VetCentsMax:       CommercialCommissionCents(ht, vetBps),
			CommercialRateBps: commBps,
			CommercialCents:   CommercialCommissionCents(ht, commBps),
		})
	}
	return out
}

// DefaultBonusRules returns static SPIFF definitions (progress filled by callers).
func DefaultBonusRules() []BonusRule {
	return []BonusRule{
		{Code: "vet_tier_31", Audience: "vet", AmountCents: 5000, TitleKey: "bonus.vetTier31", Status: "available"},
		{Code: "commercial_ramp", Audience: "commercial", AmountCents: 2500, TitleKey: "bonus.commercialRamp", Status: "available"},
		{Code: "commercial_mix", Audience: "commercial", AmountCents: 5000, TitleKey: "bonus.commercialMix", Status: "available"},
	}
}
