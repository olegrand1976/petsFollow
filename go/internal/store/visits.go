package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Visit struct {
	ID                  string     `json:"id"`
	PetID               string     `json:"petId"`
	PracticeID          string     `json:"practiceId"`
	ScheduledAt         *time.Time `json:"scheduledAt,omitempty"`
	Status              string     `json:"status"`
	Notes               string     `json:"notes"`
	Source              string     `json:"source"`
	CreatedAt           time.Time  `json:"createdAt"`
	PetName             string     `json:"petName,omitempty"`
	ClientName          string     `json:"clientName,omitempty"`
	ClientID            string     `json:"clientId,omitempty"`
	DurationMinutes     *int       `json:"durationMinutes,omitempty"`
	ProposedScheduledAt *time.Time `json:"proposedScheduledAt,omitempty"`
	PendingActionBy     *string    `json:"pendingActionBy,omitempty"`
}

type CreateVisitInput struct {
	PetID           string
	PracticeID      string
	Source          string // client | vet
	Notes           string
	ScheduledAt     *time.Time
	DurationMinutes *int
	// ConfirmDirect: vet creates already confirmed (skip client approval).
	ConfirmDirect bool
}

func (s *Store) ListVisits(ctx context.Context, petID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			'', '', '',
			duration_minutes, proposed_scheduled_at, pending_action_by
		FROM visits.visits WHERE pet_id = $1
		ORDER BY COALESCE(scheduled_at, created_at) DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisitsFull(rows)
}

func (s *Store) ListPracticeVisitsByStatus(ctx context.Context, practiceID, status string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by
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
	return scanVisitsFull(rows)
}

func (s *Store) ListPracticePendingVetActions(ctx context.Context, practiceID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE v.practice_id = $1
		  AND v.pending_action_by = 'vet'
		  AND v.status IN ('requested', 'reschedule_pending')
		ORDER BY COALESCE(v.scheduled_at, v.proposed_scheduled_at, v.created_at) DESC
		LIMIT 100`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisitsFull(rows)
}

func scanVisitsFull(rows pgx.Rows) ([]Visit, error) {
	var out []Visit
	for rows.Next() {
		var v Visit
		if err := rows.Scan(
			&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
			&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if out == nil {
		out = []Visit{}
	}
	return out, rows.Err()
}

func (s *Store) CreateVisit(ctx context.Context, in CreateVisitInput) (Visit, error) {
	id := uuid.NewString()
	status := "requested"
	var pending *string
	vet := "vet"
	client := "client"
	if in.Source == "vet" {
		if in.ConfirmDirect {
			status = "confirmed"
			pending = nil
		} else {
			status = "requested"
			pending = &client
		}
	} else {
		pending = &vet
	}
	var v Visit
	err := s.pool.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id, in.PetID, in.PracticeID, in.ScheduledAt, status, in.Notes, in.Source, in.DurationMinutes, pending,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	return v, err
}

// CreateVisitBooked serializes bookings for a practice and re-checks overlap under lock.
func (s *Store) CreateVisitBooked(ctx context.Context, in CreateVisitInput) (Visit, error) {
	if in.ScheduledAt == nil {
		return Visit{}, fmt.Errorf("%w: scheduled_required", ErrValidation)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (practice_id)
		VALUES ($1) ON CONFLICT (practice_id) DO NOTHING`, in.PracticeID); err != nil {
		return Visit{}, err
	}
	var locked string
	if err := tx.QueryRow(ctx, `
		SELECT practice_id::text FROM practice.vet_schedule WHERE practice_id = $1 FOR UPDATE`, in.PracticeID,
	).Scan(&locked); err != nil {
		return Visit{}, err
	}
	dur := 30
	if in.DurationMinutes != nil {
		dur = *in.DurationMinutes
	}
	end := in.ScheduledAt.Add(time.Duration(dur) * time.Minute)
	var n int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.visits
		WHERE practice_id = $1
		  AND status IN ('requested', 'confirmed', 'reschedule_pending')
		  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
		  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
		  AND COALESCE(proposed_scheduled_at, scheduled_at)
		      + (COALESCE(duration_minutes, $4) || ' minutes')::interval > $2`,
		in.PracticeID, *in.ScheduledAt, end, dur,
	).Scan(&n); err != nil {
		return Visit{}, err
	}
	if n > 0 {
		return Visit{}, fmt.Errorf("%w: slot_taken", ErrValidation)
	}

	id := uuid.NewString()
	vet := "vet"
	var v Visit
	err = tx.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by)
		VALUES ($1, $2, $3, $4, 'requested', $5, $6, $7, $8)
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id, in.PetID, in.PracticeID, in.ScheduledAt, in.Notes, in.Source, in.DurationMinutes, vet,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if err != nil {
		return Visit{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Visit{}, err
	}
	return v, nil
}

func (s *Store) GetVisit(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by
		FROM visits.visits WHERE id = $1`, id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) UpdateVisitStatus(ctx context.Context, id, status string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits SET
			status = $2,
			proposed_scheduled_at = CASE WHEN $2 IN ('confirmed','done','cancelled') THEN NULL ELSE proposed_scheduled_at END,
			pending_action_by = CASE WHEN $2 IN ('confirmed','done','cancelled') THEN NULL ELSE pending_action_by END
		WHERE id = $1
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id, status,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) ConfirmVisit(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits
		SET status = 'confirmed', pending_action_by = NULL, proposed_scheduled_at = NULL, status_before_reschedule = NULL
		WHERE id = $1 AND status = 'requested'
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) ProposeReschedule(ctx context.Context, id string, proposed time.Time, pendingBy string) (Visit, error) {
	if pendingBy != "vet" && pendingBy != "client" {
		return Visit{}, fmt.Errorf("%w: invalid_pending", ErrValidation)
	}
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits
		SET status = 'reschedule_pending',
			proposed_scheduled_at = $2,
			pending_action_by = $3,
			status_before_reschedule = CASE
				WHEN status = 'reschedule_pending' THEN COALESCE(status_before_reschedule, 'confirmed')
				ELSE status
			END
		WHERE id = $1 AND status IN ('requested', 'confirmed', 'reschedule_pending')
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id, proposed, pendingBy,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) AcceptReschedule(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits
		SET scheduled_at = proposed_scheduled_at,
			proposed_scheduled_at = NULL,
			pending_action_by = NULL,
			status_before_reschedule = NULL,
			status = 'confirmed'
		WHERE id = $1 AND status = 'reschedule_pending' AND proposed_scheduled_at IS NOT NULL
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) RejectReschedule(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits
		SET proposed_scheduled_at = NULL,
			status = COALESCE(NULLIF(status_before_reschedule, 'reschedule_pending'), 'requested'),
			pending_action_by = CASE
				WHEN COALESCE(NULLIF(status_before_reschedule, 'reschedule_pending'), 'requested') = 'requested' THEN 'vet'
				ELSE NULL
			END,
			status_before_reschedule = NULL
		WHERE id = $1 AND status = 'reschedule_pending'
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

// ReopenVisitAsRequested restores a cancelled/client reopen into pending vet action.
func (s *Store) ReopenVisitAsRequested(ctx context.Context, id string) (Visit, error) {
	vet := "vet"
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits
		SET status = 'requested', pending_action_by = $2, proposed_scheduled_at = NULL, status_before_reschedule = NULL
		WHERE id = $1
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by`,
		id, vet,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}
