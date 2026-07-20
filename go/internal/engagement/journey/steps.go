package journey

import (
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// PrefDiscovery / PrefBilling select which client preference gates the step.
const (
	PrefDiscovery = "discovery"
	PrefBilling   = "billing"
)

type StepKind string

const (
	KindTimed StepKind = "timed"
	KindEvent StepKind = "event"
)

type Step struct {
	Key        string
	Kind       StepKind
	OffsetDays int           // timed only
	Pref       string        // discovery | billing
	Cooldown   time.Duration // event only (0 = once forever)
	Eligible   func(seg store.JourneyClientSegment, now time.Time) (ok bool, skipReason string)
}

func TimedSteps() []Step {
	return []Step{
		{Key: "d0_welcome", Kind: KindTimed, OffsetDays: 0, Pref: PrefDiscovery, Eligible: always},
		{Key: "d1_activate", Kind: KindTimed, OffsetDays: 1, Pref: PrefDiscovery, Eligible: needsActivation},
		{Key: "d2_first_measure", Kind: KindTimed, OffsetDays: 2, Pref: PrefDiscovery, Eligible: needsFirstMeasure},
		{Key: "d4_routine", Kind: KindTimed, OffsetDays: 4, Pref: PrefDiscovery, Eligible: always},
		{Key: "d6_vet_link", Kind: KindTimed, OffsetDays: 6, Pref: PrefDiscovery, Eligible: always},
		{Key: "d10_visits", Kind: KindTimed, OffsetDays: 10, Pref: PrefDiscovery, Eligible: always},
		{Key: "d14_checkpoint", Kind: KindTimed, OffsetDays: 14, Pref: PrefDiscovery, Eligible: always},
		{Key: "d30_habit", Kind: KindTimed, OffsetDays: 30, Pref: PrefDiscovery, Eligible: always},
		{Key: "d45_care_plus", Kind: KindTimed, OffsetDays: 45, Pref: PrefDiscovery, Eligible: needsCarePlusUpsell},
		{Key: "d60_horse", Kind: KindTimed, OffsetDays: 60, Pref: PrefDiscovery, Eligible: needsHorseUpsell},
		{Key: "d75_kennel", Kind: KindTimed, OffsetDays: 75, Pref: PrefDiscovery, Eligible: needsKennelUpsell},
		{Key: "d90_quarter", Kind: KindTimed, OffsetDays: 90, Pref: PrefDiscovery, Eligible: always},
		{Key: "d120_seasonal", Kind: KindTimed, OffsetDays: 120, Pref: PrefDiscovery, Eligible: always},
		{Key: "d180_midyear", Kind: KindTimed, OffsetDays: 180, Pref: PrefDiscovery, Eligible: always},
		{Key: "d270_reengage", Kind: KindTimed, OffsetDays: 270, Pref: PrefDiscovery, Eligible: needsReengage},
		{Key: "d330_prerenew", Kind: KindTimed, OffsetDays: 330, Pref: PrefDiscovery, Eligible: always},
		{Key: "d365_anniversary", Kind: KindTimed, OffsetDays: 365, Pref: PrefDiscovery, Eligible: always},
	}
}

func EventSteps() []Step {
	return []Step{
		{Key: "evt_pending_payment", Kind: KindEvent, Pref: PrefBilling, Eligible: needsPendingPayment},
		{Key: "evt_past_due", Kind: KindEvent, Pref: PrefBilling, Eligible: needsPastDue},
		{Key: "evt_inactive_hr", Kind: KindEvent, Pref: PrefDiscovery, Cooldown: 90 * 24 * time.Hour, Eligible: needsInactiveHR},
	}
}

func AllStepKeys() []string {
	out := make([]string, 0, 24)
	for _, s := range TimedSteps() {
		out = append(out, s.Key)
	}
	for _, s := range EventSteps() {
		out = append(out, s.Key)
	}
	return out
}

func always(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	return true, ""
}

func needsActivation(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.PetCount == 0 {
		return true, ""
	}
	return false, "has_pet"
}

func needsFirstMeasure(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.ValidatedHRCount == 0 {
		return true, ""
	}
	return false, "has_validated_hr"
}

func needsCarePlusUpsell(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.ActiveAddons["care_plus"] {
		return false, "has_care_plus"
	}
	return true, ""
}

func needsHorseUpsell(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.HorseCount == 0 {
		return false, "no_horse"
	}
	if seg.ActiveAddons["horse"] {
		return false, "has_horse"
	}
	return true, ""
}

func needsKennelUpsell(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.PetCount < 6 {
		return false, "lt_6_pets"
	}
	if seg.ActiveAddons["kennel"] {
		return false, "has_kennel"
	}
	return true, ""
}

func needsReengage(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.DaysSinceLastHR == nil {
		return true, ""
	}
	if *seg.DaysSinceLastHR >= 60 {
		return true, ""
	}
	return false, "recent_hr"
}

func needsPendingPayment(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.HasPendingPayment && seg.PendingPaymentDays >= 3 {
		return true, ""
	}
	return false, "no_pending"
}

func needsPastDue(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.HasPastDue {
		return true, ""
	}
	return false, "not_past_due"
}

func needsInactiveHR(seg store.JourneyClientSegment, _ time.Time) (bool, string) {
	if seg.JourneyDays < 14 {
		return false, "before_d14"
	}
	if seg.DaysSinceLastHR == nil {
		return true, ""
	}
	if *seg.DaysSinceLastHR >= 21 {
		return true, ""
	}
	return false, "recent_hr"
}

// PrefEnabled reports whether the segment allows the preference channel.
func PrefEnabled(seg store.JourneyClientSegment, pref string) bool {
	switch pref {
	case PrefBilling:
		return seg.BillingPref
	default:
		return seg.DiscoveryPref
	}
}
