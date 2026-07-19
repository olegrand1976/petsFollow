package billing

import (
	"errors"
	"fmt"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

type PlanCode string

const (
	PlanAnnual        PlanCode = "annual"
	PlanTriennial     PlanCode = "triennial"
	PlanQuinquennial  PlanCode = "quinquennial"
)

type BillingMode string

const (
	ModeOneTime      BillingMode = "one_time"
	ModeSubscription BillingMode = "subscription"
)

type EntitlementStatus string

const (
	StatusPending   EntitlementStatus = "pending"
	StatusActive    EntitlementStatus = "active"
	StatusPastDue   EntitlementStatus = "past_due"
	StatusExpired   EntitlementStatus = "expired"
	StatusCancelled EntitlementStatus = "cancelled"
)

var (
	ErrInvalidPlan        = errors.New("invalid plan")
	ErrInvalidBillingMode = errors.New("invalid billing mode")
	ErrPaymentRequired    = errors.New("payment required")
	ErrInvalidAddon       = errors.New("invalid addon")
)

type AddonCode string

const (
	AddonFamily   AddonCode = "family"
	AddonCarePlus AddonCode = "care_plus"
	AddonHorse    AddonCode = "horse"
)

// AddonDurationDays is the validity window for a purchased addon.
const AddonDurationDays = 365

type Addon struct {
	Code         AddonCode `json:"code"`
	Label        string    `json:"label"`
	AmountCents  int       `json:"amountCents"`
	Currency     string    `json:"currency"`
	DurationDays int       `json:"durationDays"`
}

func AllAddons() []Addon {
	return AllAddonsForLocale("fr")
}

func AllAddonsForLocale(locale string) []Addon {
	locale = i18n.NormalizeLocale(locale)
	return []Addon{
		{Code: AddonFamily, Label: i18n.T(locale, "billing.addon_family_label", nil), AmountCents: 5500, Currency: "eur", DurationDays: AddonDurationDays},
		{Code: AddonCarePlus, Label: i18n.T(locale, "billing.addon_care_plus_label", nil), AmountCents: 1900, Currency: "eur", DurationDays: AddonDurationDays},
		{Code: AddonHorse, Label: i18n.T(locale, "billing.addon_horse_label", nil), AmountCents: 3900, Currency: "eur", DurationDays: AddonDurationDays},
	}
}

func ParseAddonCode(s string) (AddonCode, error) {
	switch AddonCode(s) {
	case AddonFamily, AddonCarePlus, AddonHorse:
		return AddonCode(s), nil
	default:
		return "", ErrInvalidAddon
	}
}

func GetAddon(code AddonCode) (Addon, error) {
	for _, a := range AllAddons() {
		if a.Code == code {
			return a, nil
		}
	}
	return Addon{}, ErrInvalidAddon
}

func AddonValidUntil(from time.Time, addon Addon) time.Time {
	return from.Add(time.Duration(addon.DurationDays) * 24 * time.Hour)
}

func AddonPriceIDEnvKey(code AddonCode) string {
	return fmt.Sprintf("STRIPE_PRICE_ADDON_%s", addonEnvSuffix(code))
}

func addonEnvSuffix(code AddonCode) string {
	switch code {
	case AddonFamily:
		return "FAMILY"
	case AddonCarePlus:
		return "CARE_PLUS"
	case AddonHorse:
		return "HORSE"
	default:
		return "UNKNOWN"
	}
}

type Plan struct {
	Code         PlanCode    `json:"code"`
	Label        string      `json:"label"`
	AmountCents  int         `json:"amountCents"`
	Currency     string      `json:"currency"`
	DurationDays int         `json:"durationDays"`
	Recommended  bool        `json:"recommended"`
}

type PlanOffer struct {
	Plan        Plan        `json:"plan"`
	BillingMode BillingMode `json:"billingMode"`
	Summary     string      `json:"summary"`
}

func AllPlans() []Plan {
	return AllPlansForLocale("fr")
}

func AllPlansForLocale(locale string) []Plan {
	locale = i18n.NormalizeLocale(locale)
	return []Plan{
		{Code: PlanAnnual, Label: i18n.T(locale, "billing.annual_label", nil), AmountCents: 3500, Currency: "eur", DurationDays: 365},
		{Code: PlanTriennial, Label: i18n.T(locale, "billing.triennial_label", nil), AmountCents: 9500, Currency: "eur", DurationDays: 1095, Recommended: true},
		{Code: PlanQuinquennial, Label: i18n.T(locale, "billing.quinquennial_label", nil), AmountCents: 14500, Currency: "eur", DurationDays: 1825},
	}
}

func GetPlanForLocale(code PlanCode, locale string) (Plan, error) {
	for _, p := range AllPlansForLocale(locale) {
		if p.Code == code {
			return p, nil
		}
	}
	return Plan{}, ErrInvalidPlan
}

func ParsePlanCode(s string) (PlanCode, error) {
	switch PlanCode(s) {
	case PlanAnnual, PlanTriennial, PlanQuinquennial:
		return PlanCode(s), nil
	default:
		return "", ErrInvalidPlan
	}
}

func ParseBillingMode(s string) (BillingMode, error) {
	switch BillingMode(s) {
	case ModeOneTime, ModeSubscription:
		return BillingMode(s), nil
	default:
		return "", ErrInvalidBillingMode
	}
}

// SupportsBillingMode reports whether a plan can be sold with the given mode.
// Stripe recurring Prices cannot exceed a 3-year interval, so quinquennial
// (5 years) is one_time only — entitlement duration stays 1825 days in-app.
func SupportsBillingMode(plan PlanCode, mode BillingMode) bool {
	if mode == ModeSubscription && plan == PlanQuinquennial {
		return false
	}
	return mode == ModeOneTime || mode == ModeSubscription
}

func GetPlan(code PlanCode) (Plan, error) {
	for _, p := range AllPlans() {
		if p.Code == code {
			return p, nil
		}
	}
	return Plan{}, ErrInvalidPlan
}

func PlanSummary(plan Plan, mode BillingMode) string {
	return PlanSummaryForLocale(plan, mode, "fr")
}

func PlanSummaryForLocale(plan Plan, mode BillingMode, locale string) string {
	locale = i18n.NormalizeLocale(locale)
	switch mode {
	case ModeSubscription:
		switch plan.Code {
		case PlanAnnual:
			return i18n.T(locale, "billing.annual_sub_summary", nil)
		case PlanTriennial:
			return i18n.T(locale, "billing.triennial_sub_summary", nil)
		case PlanQuinquennial:
			// Unreachable via ListPlans/StartCheckout; keep one_time copy if called.
			return i18n.T(locale, "billing.quinquennial_onetime_summary", nil)
		default:
			return plan.Label
		}
	default:
		switch plan.Code {
		case PlanAnnual:
			return i18n.T(locale, "billing.annual_onetime_summary", nil)
		case PlanTriennial:
			return i18n.T(locale, "billing.triennial_onetime_summary", nil)
		case PlanQuinquennial:
			return i18n.T(locale, "billing.quinquennial_onetime_summary", nil)
		default:
			return plan.Label
		}
	}
}

func ValidUntil(from time.Time, plan Plan) time.Time {
	return from.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)
}

func (s EntitlementStatus) AllowsAccess() bool {
	switch s {
	case StatusActive, StatusPastDue, StatusCancelled:
		return true
	default:
		return false
	}
}

func PriceIDEnvKey(plan PlanCode, mode BillingMode) string {
	return fmt.Sprintf("STRIPE_PRICE_%s_%s", planEnvSuffix(plan), modeEnvSuffix(mode))
}

func planEnvSuffix(plan PlanCode) string {
	switch plan {
	case PlanAnnual:
		return "ANNUAL"
	case PlanTriennial:
		return "TRIENNIAL"
	case PlanQuinquennial:
		return "QUINQUENNIAL"
	default:
		return "UNKNOWN"
	}
}

func modeEnvSuffix(mode BillingMode) string {
	switch mode {
	case ModeOneTime:
		return "ONETIME"
	case ModeSubscription:
		return "SUB"
	default:
		return "UNKNOWN"
	}
}
