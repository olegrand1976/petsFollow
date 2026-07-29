package kernel

type Role string

const (
	RoleVet               Role = "vet"
	RoleClient            Role = "client"
	RoleAdmin             Role = "admin"
	RoleDev               Role = "dev"
	RoleCommercial        Role = "commercial"
	RoleCommercialManager Role = "commercial_manager"
	RoleCarePro           Role = "care_pro"
	RoleVetAssistant      Role = "vet_assistant"
	RoleSecretary         Role = "secretary"
)

type ProfessionalSpecialty string

const (
	SpecialtyVetLight    ProfessionalSpecialty = "vet_light"
	SpecialtyFarrier     ProfessionalSpecialty = "farrier"
	SpecialtyPhysio      ProfessionalSpecialty = "physio"
	SpecialtyBehaviorist ProfessionalSpecialty = "behaviorist"
	SpecialtyGroomer     ProfessionalSpecialty = "groomer"
	SpecialtyBreeder     ProfessionalSpecialty = "breeder"
)

func ValidRole(role Role) bool {
	switch role {
	case RoleVet, RoleClient, RoleAdmin, RoleDev, RoleCommercial, RoleCommercialManager, RoleCarePro, RoleVetAssistant, RoleSecretary:
		return true
	default:
		return false
	}
}

// IsOpsRole reports platform ops roles (admin full + DEV support IT).
func IsOpsRole(role Role) bool {
	return role == RoleAdmin || role == RoleDev
}

// IsPracticeStaff reports whether the role belongs to a veterinary practice team.
func IsPracticeStaff(role Role) bool {
	return role == RoleVet || role == RoleVetAssistant || role == RoleSecretary
}

// IsProRole reports roles that get an automatic personal (client) profile on registration / EnsureUserProfiles.
func IsProRole(role Role) bool {
	return role == RoleVet || role == RoleCarePro || role == RoleCommercial || role == RoleCommercialManager ||
		role == RoleVetAssistant || role == RoleSecretary || role == RoleDev || role == RoleAdmin
}

func ValidSpecialty(s ProfessionalSpecialty) bool {
	switch s {
	case SpecialtyVetLight, SpecialtyFarrier, SpecialtyPhysio, SpecialtyBehaviorist, SpecialtyGroomer, SpecialtyBreeder:
		return true
	default:
		return false
	}
}

func IsCarePro(role Role) bool {
	return role == RoleCarePro
}

// IsSalesForce reports whether the role belongs to the commercial sales force.
func IsSalesForce(role Role) bool {
	return role == RoleCommercial || role == RoleCommercialManager
}

type SessionStatus string

const (
	SessionInProgress        SessionStatus = "in_progress"
	SessionPendingValidation SessionStatus = "pending_validation"
	SessionValidated         SessionStatus = "validated"
	SessionCancelled         SessionStatus = "cancelled"
)

type AvailabilityStatus string

const (
	AvailabilityAvailable   AvailabilityStatus = "available"
	AvailabilityUnavailable AvailabilityStatus = "unavailable"
)

type TimelineType string

const (
	TimelineMessage   TimelineType = "message"
	TimelineHeartRate TimelineType = "heartrate"
	TimelineWeight    TimelineType = "weight"
	TimelineEvent     TimelineType = "event"
	TimelineCare      TimelineType = "care"
	TimelineVisit     TimelineType = "visit"
)

func CalculateBPM(tapCount, durationSec int) int {
	if durationSec <= 0 {
		return 0
	}
	return (tapCount * 60) / durationSec
}

// SupportsHeartRateControl is false for "other" and unknown species.
func SupportsHeartRateControl(species string) bool {
	switch species {
	case "dog", "cat", "horse":
		return true
	default:
		return false
	}
}

// IsHeartRateDeltaAlert is true when the current BPM rose by at least delta
// compared to the previous validated reading. No previous → no alert.
func IsHeartRateDeltaAlert(currentBPM int, previousBPM *int, delta int) bool {
	if previousBPM == nil || delta <= 0 {
		return false
	}
	return currentBPM-*previousBPM >= delta
}
