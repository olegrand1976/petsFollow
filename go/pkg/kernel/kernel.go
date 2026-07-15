package kernel

type Role string

const (
	RoleVet    Role = "vet"
	RoleClient Role = "client"
	RoleAdmin  Role = "admin"
)

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
