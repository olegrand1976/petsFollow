package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Room is a work room under a practice site.
type Room struct {
	ID         string    `json:"id"`
	PracticeID string    `json:"practiceId"`
	SiteID     string    `json:"siteId"`
	Name       string    `json:"name"`
	Active     bool      `json:"active"`
	SortOrder  int       `json:"sortOrder"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CreateRoomInput struct {
	Name      string
	SortOrder *int
}

type PatchRoomInput struct {
	Name      *string
	Active    *bool
	SortOrder *int
}

const roomSelectCols = `
	id::text, practice_id::text, site_id::text, name, active, sort_order, created_at, updated_at`

func scanRoom(row pgx.Row) (Room, error) {
	var r Room
	err := row.Scan(&r.ID, &r.PracticeID, &r.SiteID, &r.Name, &r.Active, &r.SortOrder, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (s *Store) ListRooms(ctx context.Context, practiceID, siteID string, includeInactive bool) ([]Room, error) {
	resolved, err := s.ResolveSiteID(ctx, practiceID, siteID, true)
	if err != nil {
		return nil, err
	}
	q := `
		SELECT ` + roomSelectCols + `
		FROM practice.rooms
		WHERE practice_id = $1 AND site_id = $2`
	if !includeInactive {
		q += ` AND active`
	}
	q += ` ORDER BY sort_order, name`
	rows, err := s.pool.Query(ctx, q, practiceID, resolved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Room
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []Room{}
	}
	return out, rows.Err()
}

func (s *Store) GetRoom(ctx context.Context, practiceID, siteID, roomID string) (Room, error) {
	resolved, err := s.ResolveSiteID(ctx, practiceID, siteID, true)
	if err != nil {
		return Room{}, err
	}
	r, err := scanRoom(s.pool.QueryRow(ctx, `
		SELECT `+roomSelectCols+` FROM practice.rooms
		WHERE id = $1 AND practice_id = $2 AND site_id = $3`, roomID, practiceID, resolved))
	if errors.Is(err, pgx.ErrNoRows) {
		return Room{}, ErrNotFound
	}
	return r, err
}

func (s *Store) CreateRoom(ctx context.Context, practiceID, siteID string, in CreateRoomInput) (Room, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Room{}, fmt.Errorf("%w: name_required", ErrValidation)
	}
	resolved, err := s.ResolveSiteID(ctx, practiceID, siteID, false)
	if err != nil {
		return Room{}, err
	}
	var siteActive bool
	if err := s.pool.QueryRow(ctx, `
		SELECT active FROM practice.sites WHERE id = $1 AND practice_id = $2`, resolved, practiceID,
	).Scan(&siteActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Room{}, fmt.Errorf("%w: invalid_site", ErrValidation)
		}
		return Room{}, err
	}
	if !siteActive {
		return Room{}, fmt.Errorf("%w: site_inactive", ErrValidation)
	}
	sortOrder := 0
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	} else {
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(MAX(sort_order), -1) + 1 FROM practice.rooms WHERE site_id = $1`, resolved,
		).Scan(&sortOrder)
	}
	id := uuid.NewString()
	r, err := scanRoom(s.pool.QueryRow(ctx, `
		INSERT INTO practice.rooms (id, practice_id, site_id, name, active, sort_order)
		VALUES ($1, $2, $3, $4, TRUE, $5)
		RETURNING `+roomSelectCols, id, practiceID, resolved, name, sortOrder))
	if err != nil {
		if isUniqueViolation(err) {
			return Room{}, fmt.Errorf("%w: room_name_taken", ErrValidation)
		}
		return Room{}, err
	}
	return r, nil
}

func (s *Store) PatchRoom(ctx context.Context, practiceID, siteID, roomID string, in PatchRoomInput) (Room, error) {
	cur, err := s.GetRoom(ctx, practiceID, siteID, roomID)
	if err != nil {
		return Room{}, err
	}
	name := cur.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return Room{}, fmt.Errorf("%w: name_required", ErrValidation)
		}
	}
	active := cur.Active
	if in.Active != nil {
		active = *in.Active
	}
	sortOrder := cur.SortOrder
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	}
	r, err := scanRoom(s.pool.QueryRow(ctx, `
		UPDATE practice.rooms
		SET name = $4, active = $5, sort_order = $6, updated_at = NOW()
		WHERE id = $1 AND practice_id = $2 AND site_id = $3
		RETURNING `+roomSelectCols, roomID, practiceID, cur.SiteID, name, active, sortOrder))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Room{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return Room{}, fmt.Errorf("%w: room_name_taken", ErrValidation)
		}
		return Room{}, err
	}
	return r, nil
}

// DeactivateRoom soft-deactivates a room (visits keep room_id until cleared).
func (s *Store) DeactivateRoom(ctx context.Context, practiceID, siteID, roomID string) (Room, error) {
	active := false
	return s.PatchRoom(ctx, practiceID, siteID, roomID, PatchRoomInput{Active: &active})
}

// ValidateVisitRoom ensures room belongs to practice+site and is active (when set).
func (s *Store) ValidateVisitRoom(ctx context.Context, practiceID, siteID, roomID string) error {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return nil
	}
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.rooms
			WHERE id = $1 AND practice_id = $2 AND site_id = $3 AND active
		)`, roomID, practiceID, siteID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: invalid_room", ErrValidation)
	}
	return nil
}

// ValidateVisitAssignee ensures user is an active team member included in the calendar.
func (s *Store) ValidateVisitAssignee(ctx context.Context, practiceID, userID string) error {
	return s.validateVisitAssignee(ctx, practiceID, userID, true)
}

// validateVisitAssignee checks active membership; when requireInCalendar, also include_in_calendar.
func (s *Store) validateVisitAssignee(ctx context.Context, practiceID, userID string, requireInCalendar bool) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.team_members
			WHERE practice_id = $1 AND user_id = $2 AND status = 'active'
			  AND ($3 OR COALESCE(include_in_calendar, true))
		)`, practiceID, userID, !requireInCalendar).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: invalid_assignee", ErrValidation)
	}
	return nil
}
