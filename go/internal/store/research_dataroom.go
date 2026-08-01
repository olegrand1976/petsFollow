package store

import (
	"context"
	"fmt"
	"time"
)

// ResearchDataRoomEvent is a micro-event safe for Data room export
// (no practice_id_hash / source_hash / city / payload — payload stays warehouse-only).
type ResearchDataRoomEvent struct {
	EventWeek   string `json:"eventWeek"`
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
	Species     string `json:"species"`
	AgeBand     string `json:"ageBand"`
	SignalType  string `json:"signalType"`
}

// ResearchDataRoomEvents returns anonymized micro-events only for grains with
// ≥ ResearchKAnonymity distinct practice_id_hash (multi-cabinet k-anonymity).
// Access is network-wide once the caller has membership in any dataroom_enabled group
// (groups gate privilege; they do not scope the event set).
func (s *Store) ResearchDataRoomEvents(ctx context.Context, species, signal, country, postal, from, to string, limit int) ([]ResearchDataRoomEvent, error) {
	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return nil, fmt.Errorf("invalid_from_date")
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return nil, fmt.Errorf("invalid_to_date")
		}
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `
		WITH eligible AS (
		  SELECT event_week, postal_code, country_code, species, signal_type
		  FROM research.anon_events
		  GROUP BY event_week, postal_code, country_code, species, signal_type
		  HAVING COUNT(DISTINCT practice_id_hash) >= $7
		)
		SELECT e.event_week::text, e.postal_code, e.country_code, e.species, e.age_band,
		       e.signal_type
		FROM research.anon_events e
		JOIN eligible k ON k.event_week = e.event_week
		  AND k.postal_code = e.postal_code
		  AND k.country_code = e.country_code
		  AND k.species = e.species
		  AND k.signal_type = e.signal_type
		WHERE ($1 = '' OR e.species = $1)
		  AND ($2 = '' OR e.signal_type = $2)
		  AND ($3 = '' OR e.country_code = $3)
		  AND ($4 = '' OR e.postal_code = $4)
		  AND ($5 = '' OR e.event_week >= $5::date)
		  AND ($6 = '' OR e.event_week <= $6::date)
		ORDER BY e.event_week DESC, e.created_at DESC
		LIMIT $8`,
		species, signal, country, postal, from, to, ResearchKAnonymity, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchDataRoomEvent
	for rows.Next() {
		var e ResearchDataRoomEvent
		if err := rows.Scan(&e.EventWeek, &e.PostalCode, &e.CountryCode, &e.Species,
			&e.AgeBand, &e.SignalType); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []ResearchDataRoomEvent{}
	}
	return out, rows.Err()
}
