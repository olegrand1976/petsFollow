package billing

import (
	"errors"
	"fmt"
	"time"
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
)

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
	return []Plan{
		{Code: PlanAnnual, Label: "25 € / an", AmountCents: 2500, Currency: "eur", DurationDays: 365},
		{Code: PlanTriennial, Label: "60 € / 3 ans", AmountCents: 6000, Currency: "eur", DurationDays: 1095, Recommended: true},
		{Code: PlanQuinquennial, Label: "75 € / 5 ans", AmountCents: 7500, Currency: "eur", DurationDays: 1825},
	}
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

func GetPlan(code PlanCode) (Plan, error) {
	for _, p := range AllPlans() {
		if p.Code == code {
			return p, nil
		}
	}
	return Plan{}, ErrInvalidPlan
}

func PlanSummary(plan Plan, mode BillingMode) string {
	switch mode {
	case ModeSubscription:
		switch plan.Code {
		case PlanAnnual:
			return "25 € / an, renouvelé automatiquement"
		case PlanTriennial:
			return "60 € tous les 3 ans, renouvelé automatiquement"
		case PlanQuinquennial:
			return "75 € tous les 5 ans, renouvelé automatiquement"
		default:
			return plan.Label
		}
	default:
		switch plan.Code {
		case PlanAnnual:
			return "25 € pour 1 an, paiement unique"
		case PlanTriennial:
			return "60 € pour 3 ans, paiement unique"
		case PlanQuinquennial:
			return "75 € pour 5 ans, paiement unique"
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
