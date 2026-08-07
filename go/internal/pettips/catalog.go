// Package pettips serves practical, non-scientific tips for pet owners,
// filtered by the species present in their household.
package pettips

import (
	"sort"
	"strings"
	"time"
)

// Priority levels (lower = more important).
const (
	PriorityHigh   = 0
	PriorityMedium = 1
	PriorityLow    = 2
)

// Def is a tip definition (copy localized via i18n keys pet_tips.<id>.title/body).
type Def struct {
	ID       string
	Species  []string // empty = any species in household
	Priority int
	Months   []int // empty = all year; calendar months 1–12 (Europe/Brussels)
}

// Item is a localized tip ready for the API.
type Item struct {
	ID       string   `json:"id"`
	Species  []string `json:"species"`
	Priority string   `json:"priority"` // high|medium|low
	Title    string   `json:"title"`
	Body     string   `json:"body"`
}

// Catalog is the curated owner-facing tip list (not scientific / not Pro vet news).
var Catalog = []Def{
	{
		ID: "heat_safety", Species: []string{"dog", "cat"},
		Priority: PriorityHigh, Months: []int{6, 7, 8},
	},
	{
		ID: "parasite_season", Species: []string{"dog", "cat"},
		Priority: PriorityHigh, Months: []int{4, 5, 6, 7, 8, 9, 10},
	},
	{
		ID: "cold_paw_care", Species: []string{"dog", "cat"},
		Priority: PriorityMedium, Months: []int{11, 12, 1, 2},
	},
	{
		ID: "horse_pasture", Species: []string{"horse"},
		Priority: PriorityMedium, Months: []int{4, 5, 6, 7, 8, 9},
	},
	{
		ID: "hydration", Species: nil, // any
		Priority: PriorityMedium, Months: []int{6, 7, 8},
	},
	{
		ID: "seasonal_care_check", Species: nil,
		Priority: PriorityLow, Months: nil,
	},
}

func priorityLabel(p int) string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	default:
		return "low"
	}
}

func monthOK(months []int, now time.Time) bool {
	if len(months) == 0 {
		return true
	}
	m := int(now.Month())
	for _, x := range months {
		if x == m {
			return true
		}
	}
	return false
}

func speciesMatch(tipSpecies, household []string) bool {
	if len(tipSpecies) == 0 {
		return len(household) > 0
	}
	set := make(map[string]struct{}, len(household))
	for _, s := range household {
		set[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	for _, s := range tipSpecies {
		if _, ok := set[strings.ToLower(strings.TrimSpace(s))]; ok {
			return true
		}
	}
	return false
}

// NormalizeSpecies returns lowercased unique species codes.
func NormalizeSpecies(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// Select returns tip defs matching household species and calendar month, sorted by priority.
func Select(householdSpecies []string, now time.Time, limit int) []Def {
	species := NormalizeSpecies(householdSpecies)
	if len(species) == 0 || limit <= 0 {
		return nil
	}
	var matched []Def
	for _, d := range Catalog {
		if !monthOK(d.Months, now) {
			continue
		}
		if !speciesMatch(d.Species, species) {
			continue
		}
		matched = append(matched, d)
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Priority != matched[j].Priority {
			return matched[i].Priority < matched[j].Priority
		}
		return matched[i].ID < matched[j].ID
	})
	if len(matched) > limit {
		matched = matched[:limit]
	}
	return matched
}
