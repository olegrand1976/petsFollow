package kernel

type Role string

const (
	RoleVet               Role = "vet"
	RoleClient            Role = "client"
	RoleAdmin             Role = "admin"
	RoleCommercial        Role = "commercial"
	RoleCommercialManager Role = "commercial_manager"
	RoleCarePro           Role = "care_pro"
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
	case RoleVet, RoleClient, RoleAdmin, RoleCommercial, RoleCommercialManager, RoleCarePro:
		return true
	default:
		return false
	}
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

func IsHeartRateAlert(bpm, minBPM, maxBPM int) bool {
	return bpm < minBPM || bpm > maxBPM
}
