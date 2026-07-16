package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Visit struct {
	ID          string     `json:"id"`
	PetID       string     `json:"petId"`
	PracticeID  string     `json:"practiceId"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	Status      string     `json:"status"`
	Notes       string     `json:"notes"`
	Source      string     `json:"source"`
	CreatedAt   time.Time  `json:"createdAt"`
	PetName     string     `json:"petName,omitempty"`
	ClientName  string     `json:"clientName,omitempty"`
	ClientID    string     `json:"clientId,omitempty"`
}

func (s *Store) ListVisits(ctx context.Context, petID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			'', '', ''
		FROM visits.visits WHERE pet_id = $1
		ORDER BY COALESCE(scheduled_at, created_at) DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisits(rows)
}

func (s *Store) ListPracticeVisitsByStatus(ctx context.Context, practiceID, status string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE v.practice_id = $1 AND v.status = $2
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC
		LIMIT 100`, practiceID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisits(rows)
}

func scanVisits(rows pgx.Rows) ([]Visit, error) {
	var out []Visit
	for rows.Next() {
		var v Visit
		if err := rows.Scan(
			&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) CreateVisit(ctx context.Context, petID, practiceID, source, notes string, scheduledAt *time.Time) (Visit, error) {
	id := uuid.NewString()
	status := "requested"
	if source == "vet" {
		status = "confirmed"
	}
	var v Visit
	err := s.pool.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at`,
		id, petID, practiceID, scheduledAt, status, notes, source,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt)
	return v, err
}

func (s *Store) GetVisit(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at
		FROM visits.visits WHERE id = $1`, id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) UpdateVisitStatus(ctx context.Context, id, status string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits SET status = $2
		WHERE id = $1
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at`,
		id, status,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}
