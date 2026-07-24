package store

import (
	"context"
	"time"
)

// VetPetListItem is a practice-scoped pet row for the Pro animals list.
type VetPetListItem struct {
	ID                   string     `json:"id"`
	PracticeID           string     `json:"practiceId"`
	OwnerUserID          string     `json:"ownerUserId"`
	OwnerName            string     `json:"ownerName"`
	Name                 string     `json:"name"`
	Species              string     `json:"species"`
	Breed                string     `json:"breed"`
	BirthDate            *time.Time `json:"birthDate,omitempty"`
	PhotoURL             string     `json:"photoUrl,omitempty"`
	LastVisitAt          *time.Time `json:"lastVisitAt,omitempty"`
	LastHeartRateAt      *time.Time `json:"lastHeartRateAt,omitempty"`
	LastHeartRateBpm     *int       `json:"lastHeartRateBpm,omitempty"`
	UnreadHeartrateCount int        `json:"unreadHeartrateCount"`
}

func (s *Store) ListPetsForPractice(ctx context.Context, practiceID string) ([]VetPetListItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			p.id::text,
			p.practice_id::text,
			p.owner_user_id::text,
			COALESCE(u.full_name, ''),
			p.name,
			p.species,
			COALESCE(p.breed, ''),
			p.birth_date,
			COALESCE(p.photo_url, ''),
			lv.last_visit_at,
			lhs.last_hr_at,
			lhs.last_hr_bpm,
			COALESCE(ur.unread_count, 0)
		FROM pets.pets p
		JOIN identity.users u ON u.id = p.owner_user_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(v.scheduled_at, v.created_at) AS last_visit_at
			FROM visits.visits v
			WHERE v.pet_id = p.id AND v.status = 'done'
			ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC
			LIMIT 1
		) lv ON TRUE
		LEFT JOIN LATERAL (
			SELECT s.started_at AS last_hr_at, s.bpm AS last_hr_bpm
			FROM heartrate.sessions s
			WHERE s.pet_id = p.id
			  AND s.status = 'validated'
			  AND s.bpm IS NOT NULL
			ORDER BY s.started_at DESC
			LIMIT 1
		) lhs ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int AS unread_count
			FROM heartrate.sessions s
			WHERE s.pet_id = p.id
			  AND s.status = 'validated'
			  AND s.vet_seen_at IS NULL
		) ur ON TRUE
		WHERE p.practice_id = $1
		ORDER BY p.name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []VetPetListItem
	for rows.Next() {
		var item VetPetListItem
		if err := rows.Scan(
			&item.ID,
			&item.PracticeID,
			&item.OwnerUserID,
			&item.OwnerName,
			&item.Name,
			&item.Species,
			&item.Breed,
			&item.BirthDate,
			&item.PhotoURL,
			&item.LastVisitAt,
			&item.LastHeartRateAt,
			&item.LastHeartRateBpm,
			&item.UnreadHeartrateCount,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []VetPetListItem{}
	}
	return out, nil
}
