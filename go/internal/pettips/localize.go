package pettips

import (
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

const defaultLimit = 3

// ForHousehold builds localized tip items for the given species set.
func ForHousehold(locale string, householdSpecies []string, now time.Time, limit int) []Item {
	if limit <= 0 {
		limit = defaultLimit
	}
	defs := Select(householdSpecies, now, limit)
	if len(defs) == 0 {
		return []Item{}
	}
	out := make([]Item, 0, len(defs))
	for _, d := range defs {
		species := d.Species
		if species == nil {
			species = []string{}
		}
		out = append(out, Item{
			ID:       d.ID,
			Species:  append([]string(nil), species...),
			Priority: priorityLabel(d.Priority),
			Title:    i18n.T(locale, "pet_tips."+d.ID+".title", nil),
			Body:     i18n.T(locale, "pet_tips."+d.ID+".body", nil),
		})
	}
	return out
}
