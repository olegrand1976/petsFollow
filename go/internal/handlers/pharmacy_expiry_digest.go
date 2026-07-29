package handlers

import "github.com/olegrand1976/petsFollow/go/internal/store"

// shouldSendPharmacyExpiryDigest gates the weekly stock expiry digest:
// enabled + (force OR matching ISO weekday) + non-empty alert bands.
func shouldSendPharmacyExpiryDigest(enabled, force bool, weekdayISO, configuredWeekday int, sum store.ExpirySummary) bool {
	if !enabled {
		return false
	}
	if !force && weekdayISO != configuredWeekday {
		return false
	}
	return sum.Critical+sum.Expired+sum.Quarantine+sum.Return+sum.Soon > 0
}
