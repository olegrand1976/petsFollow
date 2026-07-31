package handlers

import "github.com/olegrand1976/petsFollow/go/internal/store"

// shouldQueryPharmacyExpiryDigest is true when a summary fetch is worth doing
// (digest enabled + force or matching ISO weekday). Empty-band skip happens after.
func shouldQueryPharmacyExpiryDigest(enabled, force bool, weekdayISO, configuredWeekday int) bool {
	if !enabled {
		return false
	}
	return force || weekdayISO == configuredWeekday
}

// shouldSendPharmacyExpiryDigest gates the weekly stock expiry digest:
// enabled + (force OR matching ISO weekday) + non-empty alert bands.
func shouldSendPharmacyExpiryDigest(enabled, force bool, weekdayISO, configuredWeekday int, sum store.ExpirySummary) bool {
	if !shouldQueryPharmacyExpiryDigest(enabled, force, weekdayISO, configuredWeekday) {
		return false
	}
	return sum.Critical+sum.Expired+sum.Quarantine+sum.Return+sum.Soon > 0
}
