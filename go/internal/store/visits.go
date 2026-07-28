package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	AddressText         string     `json:"addressText,omitempty"`
	Lat                 *float64   `json:"lat,omitempty"`
	Lng                 *float64   `json:"lng,omitempty"`
	// Visit type (optional catalogue entry); name/color hydrated on calendar lists.
	VisitTypeID    string `json:"visitTypeId,omitempty"`
	VisitTypeName  string `json:"visitTypeName,omitempty"`
	VisitTypeColor string `json:"visitTypeColor,omitempty"`
	// PreconsultStatus is pending|submitted|skipped when an intake exists.
	PreconsultStatus string `json:"preconsultStatus,omitempty"`
	// RequestPreconsult: VetPro opted in to send public preconsult questionnaire.
	RequestPreconsult bool `json:"requestPreconsult,omitempty"`
	// ConsultationSession: walk-in CR flow — excluded from agenda overlap / slot busy.
	ConsultationSession bool `json:"consultationSession,omitempty"`
	// Permission is set for care_pro list responses (read | write_notes | full).
	Permission string `json:"permission,omitempty"`
}

type CreateVisitInput struct {
	PetID             string
	PracticeID        string
	Source            string // client | vet | care_pro
	Notes             string
	ScheduledAt       *time.Time
	DurationMinutes   *int
	VisitTypeID       *string
	// ConfirmDirect: vet/care_pro creates already confirmed (skip client approval).
	ConfirmDirect bool
	// RequestPreconsult: when confirmed, send public preconsult questionnaire.
	RequestPreconsult bool
	// ConsultationSession: walk-in — skip overlap checks; does not block slots.
	ConsultationSession bool
}

func isProVisitSource(source string) bool {
	return source == "vet" || source == "care_pro"
}

func (s *Store) ListVisits(ctx context.Context, petID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			'', '', '',
			duration_minutes, proposed_scheduled_at, pending_action_by,
			COALESCE(address_text,''), lat, lng, COALESCE(consultation_session, false)
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
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false)
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

// ConsultationListItem is a walk-in consultation with report summary flags (no audio keys).
type ConsultationListItem struct {
	Visit
	HasReport    bool   `json:"hasReport"`
	HasAudio     bool   `json:"hasAudio"`
	ReportStatus string `json:"reportStatus,omitempty"`
}

// ListConsultationsFilter filters practice walk-in consultations.
type ListConsultationsFilter struct {
	Status   string
	Query    string
	From     *time.Time
	To       *time.Time
	HasAudio *bool
	Limit    int
	Offset   int
}

// ListPracticeConsultations returns consultation_session visits newest first.
func (s *Store) ListPracticeConsultations(ctx context.Context, practiceID string, f ListConsultationsFilter) ([]ConsultationListItem, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	args := []any{practiceID}
	where := []string{
		`v.practice_id = $1`,
		`COALESCE(v.consultation_session, false) = true`,
	}
	argN := 2

	if status := strings.TrimSpace(f.Status); status != "" {
		where = append(where, fmt.Sprintf(`v.status = $%d`, argN))
		args = append(args, status)
		argN++
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		where = append(where, fmt.Sprintf(
			`(p.name ILIKE $%d OR u.full_name ILIKE $%d OR COALESCE(u.email,'') ILIKE $%d OR COALESCE(v.notes,'') ILIKE $%d)`,
			argN, argN, argN, argN,
		))
		args = append(args, "%"+q+"%")
		argN++
	}
	if f.From != nil {
		where = append(where, fmt.Sprintf(`COALESCE(v.scheduled_at, v.created_at) >= $%d`, argN))
		args = append(args, *f.From)
		argN++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf(`COALESCE(v.scheduled_at, v.created_at) <= $%d`, argN))
		args = append(args, *f.To)
		argN++
	}
	if f.HasAudio != nil {
		if *f.HasAudio {
			where = append(where, `EXISTS (
				SELECT 1 FROM visits.visit_reports r
				WHERE r.visit_id = v.id AND r.status = 'draft'
				  AND length(trim(COALESCE(r.audio_object_key, ''))) > 0
			)`)
		} else {
			where = append(where, `NOT EXISTS (
				SELECT 1 FROM visits.visit_reports r
				WHERE r.visit_id = v.id AND r.status = 'draft'
				  AND length(trim(COALESCE(r.audio_object_key, ''))) > 0
			)`)
		}
	}

	args = append(args, limit, offset)
	limitPh := argN
	offsetPh := argN + 1

	q := fmt.Sprintf(`
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			EXISTS (
				SELECT 1 FROM visits.visit_reports r
				WHERE r.visit_id = v.id AND %s
			) AS has_report,
			EXISTS (
				SELECT 1 FROM visits.visit_reports r
				WHERE r.visit_id = v.id AND r.status = 'draft'
				  AND length(trim(COALESCE(r.audio_object_key, ''))) > 0
			) AS has_audio,
			COALESCE((
				SELECT CASE
					WHEN bool_or(r.status = 'final') THEN 'final'
					WHEN bool_or(%s) THEN 'draft'
					ELSE ''
				END
				FROM visits.visit_reports r
				WHERE r.visit_id = v.id
			), '') AS report_status
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE %s
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC
		LIMIT $%d OFFSET $%d`,
		sqlVisitReportIsPersisted, sqlVisitReportIsPersisted, strings.Join(where, " AND "),
		limitPh, offsetPh,
	)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ConsultationListItem
	for rows.Next() {
		var item ConsultationListItem
		var v Visit
		if err := rows.Scan(
			&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
			&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
			&v.AddressText, &v.Lat, &v.Lng, &v.ConsultationSession,
			&item.HasReport, &item.HasAudio, &item.ReportStatus,
		); err != nil {
			return nil, err
		}
		item.Visit = v
		out = append(out, item)
	}
	if out == nil {
		out = []ConsultationListItem{}
	}
	return out, rows.Err()
}

func (s *Store) ListPracticePendingVetActions(ctx context.Context, practiceID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false)
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
			&v.AddressText, &v.Lat, &v.Lng, &v.ConsultationSession,
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
	if isProVisitSource(in.Source) {
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
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by, request_preconsult, consultation_session, visit_type_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by, COALESCE(request_preconsult,false), COALESCE(consultation_session,false),
			COALESCE(visit_type_id::text,'')`,
		id, in.PetID, in.PracticeID, in.ScheduledAt, status, in.Notes, in.Source, in.DurationMinutes, pending, in.RequestPreconsult, in.ConsultationSession, in.VisitTypeID,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy, &v.RequestPreconsult, &v.ConsultationSession, &v.VisitTypeID)
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
	// Walk-in consultation sessions do not block (or get blocked by) agenda slots.
	if !in.ConsultationSession {
		end := in.ScheduledAt.Add(time.Duration(dur) * time.Minute)
		var n int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE practice_id = $1
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
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
	}

	status := "requested"
	var pending *string
	vet := "vet"
	client := "client"
	if isProVisitSource(in.Source) {
		if in.ConfirmDirect {
			status = "confirmed"
			pending = nil
		} else {
			pending = &client
		}
	} else {
		pending = &vet
	}

	id := uuid.NewString()
	var v Visit
	err = tx.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by, request_preconsult, consultation_session, visit_type_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id::text, pet_id::text, practice_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by, COALESCE(request_preconsult,false), COALESCE(consultation_session,false),
			COALESCE(visit_type_id::text,'')`,
		id, in.PetID, in.PracticeID, in.ScheduledAt, status, in.Notes, in.Source, in.DurationMinutes, pending, in.RequestPreconsult, in.ConsultationSession, in.VisitTypeID,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy, &v.RequestPreconsult, &v.ConsultationSession, &v.VisitTypeID)
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
			duration_minutes, proposed_scheduled_at, pending_action_by,
			COALESCE(address_text,''), lat, lng, COALESCE(request_preconsult,false), COALESCE(consultation_session,false)
		FROM visits.visits WHERE id = $1`, id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
		&v.AddressText, &v.Lat, &v.Lng, &v.RequestPreconsult, &v.ConsultationSession)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

// SetVisitRequestPreconsult toggles the opt-in flag (VetPro calendar.manage).
func (s *Store) SetVisitRequestPreconsult(ctx context.Context, id string, request bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits SET request_preconsult = $2 WHERE id = $1`, id, request)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateVisitLocation sets address. When clearCoords is true, lat/lng are set to NULL.
// Otherwise lat/lng are only overwritten when non-nil (preserve GPS on address-only PATCH).
func (s *Store) UpdateVisitLocation(ctx context.Context, id, addressText string, lat, lng *float64, clearCoords bool) (Visit, error) {
	var tag pgconn.CommandTag
	var err error
	if clearCoords {
		tag, err = s.pool.Exec(ctx, `
			UPDATE visits.visits
			SET address_text = $2, lat = NULL, lng = NULL
			WHERE id = $1`, id, addressText)
	} else {
		tag, err = s.pool.Exec(ctx, `
			UPDATE visits.visits
			SET address_text = $2,
				lat = COALESCE($3, lat),
				lng = COALESCE($4, lng)
			WHERE id = $1`, id, addressText, lat, lng)
	}
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

func (s *Store) UpdateVisitStatus(ctx context.Context, id, status string) (Visit, error) {
	var allowedFrom []string
	switch status {
	case "cancelled":
		allowedFrom = []string{"requested", "confirmed", "reschedule_pending"}
	case "done":
		allowedFrom = []string{"confirmed"}
	default:
		return Visit{}, fmt.Errorf("%w: invalid_status", ErrValidation)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits SET
			status = $2,
			proposed_scheduled_at = CASE WHEN $2 IN ('confirmed','done','cancelled') THEN NULL ELSE proposed_scheduled_at END,
			pending_action_by = CASE WHEN $2 IN ('confirmed','done','cancelled') THEN NULL ELSE pending_action_by END,
			status_before_reschedule = CASE WHEN $2 IN ('confirmed','done','cancelled') THEN NULL ELSE status_before_reschedule END
		WHERE id = $1 AND status = ANY($3::text[])`,
		id, status, allowedFrom,
	)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

// CancelStaleConsultationOrphans cancels walk-in visits still confirmed after olderThan
// with no persisted CR (empty Ensure draft counts as orphan). Predicate shared with
// VisitHasPersistedReport via sqlVisitReportIsPersisted.
func (s *Store) CancelStaleConsultationOrphans(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	tag, err := s.pool.Exec(ctx, `
		WITH stale AS (
			SELECT v.id
			FROM visits.visits v
			WHERE COALESCE(v.consultation_session, false) = true
			  AND v.status = 'confirmed'
			  AND v.scheduled_at IS NOT NULL
			  AND v.scheduled_at < $1
			  AND NOT EXISTS (
				SELECT 1 FROM visits.visit_reports r
				WHERE r.visit_id = v.id
				  AND `+sqlVisitReportIsPersisted+`
			  )
			ORDER BY v.scheduled_at ASC
			LIMIT $2
		)
		UPDATE visits.visits v
		SET status = 'cancelled',
			proposed_scheduled_at = NULL,
			pending_action_by = NULL,
			status_before_reschedule = NULL
		FROM stale
		WHERE v.id = stale.id`,
		olderThan, limit,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) ConfirmVisit(ctx context.Context, id string) (Visit, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET status = 'confirmed', pending_action_by = NULL, proposed_scheduled_at = NULL, status_before_reschedule = NULL
		WHERE id = $1 AND status = 'requested'`, id)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

func (s *Store) ProposeReschedule(ctx context.Context, id string, proposed time.Time, pendingBy string) (Visit, error) {
	if pendingBy != "vet" && pendingBy != "client" {
		return Visit{}, fmt.Errorf("%w: invalid_pending", ErrValidation)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET status = 'reschedule_pending',
			proposed_scheduled_at = $2,
			pending_action_by = $3,
			status_before_reschedule = CASE
				WHEN status = 'reschedule_pending' THEN COALESCE(status_before_reschedule, 'confirmed')
				ELSE status
			END
		WHERE id = $1 AND status IN ('requested', 'confirmed', 'reschedule_pending')`,
		id, proposed, pendingBy,
	)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

func (s *Store) AcceptReschedule(ctx context.Context, id string) (Visit, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET scheduled_at = proposed_scheduled_at,
			proposed_scheduled_at = NULL,
			pending_action_by = NULL,
			status_before_reschedule = NULL,
			status = 'confirmed'
		WHERE id = $1 AND status = 'reschedule_pending' AND proposed_scheduled_at IS NOT NULL`, id)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

func (s *Store) RejectReschedule(ctx context.Context, id string) (Visit, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET proposed_scheduled_at = NULL,
			status = COALESCE(NULLIF(status_before_reschedule, 'reschedule_pending'), 'requested'),
			pending_action_by = CASE
				WHEN COALESCE(NULLIF(status_before_reschedule, 'reschedule_pending'), 'requested') = 'requested' THEN 'vet'
				ELSE NULL
			END,
			status_before_reschedule = NULL
		WHERE id = $1 AND status = 'reschedule_pending'`, id)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

// ReopenVisitAsRequested restores a cancelled visit into pending vet action.
func (s *Store) ReopenVisitAsRequested(ctx context.Context, id string) (Visit, error) {
	vet := "vet"
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET status = 'requested', pending_action_by = $2, proposed_scheduled_at = NULL, status_before_reschedule = NULL
		WHERE id = $1 AND status = 'cancelled'`, id, vet)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}
