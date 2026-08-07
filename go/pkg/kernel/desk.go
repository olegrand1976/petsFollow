package kernel

import "slices"

// AllowedDeskIdleMinutes are the selectable shared-desk lock delays (minutes).
var AllowedDeskIdleMinutes = []int{1, 2, 5, 10, 15, 30}

const DefaultDeskIdleMinutes = 2

func IsAllowedDeskIdleMinutes(min int) bool {
	return slices.Contains(AllowedDeskIdleMinutes, min)
}

// NormalizeDeskIdleMinutes returns min if allowed, otherwise DefaultDeskIdleMinutes.
func NormalizeDeskIdleMinutes(min int) int {
	if IsAllowedDeskIdleMinutes(min) {
		return min
	}
	return DefaultDeskIdleMinutes
}
