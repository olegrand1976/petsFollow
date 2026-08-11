package store

import (
	"errors"
	"strings"
)

// ErrInvalidSupportTransition is returned when a status patch violates the workflow graph.
var ErrInvalidSupportTransition = errors.New("invalid_support_transition")

// supportStatusTransitions encodes open → in_progress → to_test → done (+ closed abandon / reopen).
// Same-status no-ops are handled by the caller before this check.
var supportStatusTransitions = map[string][]string{
	SupportStatusOpen:       {SupportStatusInProgress, SupportStatusClosed},
	SupportStatusInProgress: {SupportStatusToTest, SupportStatusOpen, SupportStatusClosed},
	SupportStatusToTest:     {SupportStatusDone, SupportStatusInProgress, SupportStatusClosed},
	SupportStatusDone:       {SupportStatusClosed},
	SupportStatusClosed:     {SupportStatusOpen},
}

// IsValidSupportStatusTransition reports whether from→to is allowed (false if either unknown).
func IsValidSupportStatusTransition(from, to string) bool {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" || !validSupportStatuses[from] || !validSupportStatuses[to] {
		return false
	}
	if from == to {
		return true
	}
	for _, next := range supportStatusTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// AllowedSupportStatusTransitions returns the allowed next statuses from from (excluding from itself).
func AllowedSupportStatusTransitions(from string) []string {
	from = strings.TrimSpace(from)
	out := supportStatusTransitions[from]
	if len(out) == 0 {
		return nil
	}
	cp := make([]string, len(out))
	copy(cp, out)
	return cp
}
