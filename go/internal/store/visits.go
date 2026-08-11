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
	ID          string     `json:"id"`
	PetID       string     `json:"petId"`
	PracticeID  string     `json:"practiceId"`
	SiteID      string     `json:"siteId,omitempty"`
	SiteName    string     `json:"siteName,omitempty"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	Status      string     `json:"status"`
	Notes       string     `json:"notes"`
	// CallbackPhone: temporary contact for walk-in / Nouveau client bookings (not on the immutable slot user).
	CallbackPhone       string     `json:"callbackPhone,omitempty"`
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
	// PreconsultAlert is "urgent" when submitted intake is AI-red or declared high.
	PreconsultAlert string `json:"preconsultAlert,omitempty"`
	// RequestPreconsult: VetPro opted in to send public preconsult questionnaire.
	RequestPreconsult bool `json:"requestPreconsult,omitempty"`
	// ConsultationSession: walk-in CR flow — excluded from agenda overlap / slot busy.
	ConsultationSession bool `json:"consultationSession,omitempty"`
	// IsWalkinPlaceholder: visit still on system "Nouvel animal" — needs identify before CR finalize.
	IsWalkinPlaceholder bool `json:"isWalkinPlaceholder,omitempty"`
	// WaitingRoomAt: desk marked client as arrived / in waiting room.
	WaitingRoomAt *time.Time `json:"waitingRoomAt,omitempty"`
	// HasFinalReport: at least one visit_reports row with status=final (client list enrichment).
	HasFinalReport bool `json:"hasFinalReport,omitempty"`
	// ReportStatus: owner-only enrichment — "final" | "draft" (never draft body text).
	ReportStatus string `json:"reportStatus,omitempty"`
	// Permission is set for care_pro list responses (read | write_notes | full).
	Permission string `json:"permission,omitempty"`
	// Calendar resources (optional).
	AssigneeUserID string `json:"assigneeUserId,omitempty"`
	AssigneeName   string `json:"assigneeName,omitempty"`
	RoomID         string `json:"roomId,omitempty"`
	RoomName       string `json:"roomName,omitempty"`
}

type CreateVisitInput struct {
	PetID           string
	PracticeID      string
	SiteID          string // empty → primary site
	Source          string // client | vet | care_pro
	Notes           string
	CallbackPhone   string // required when pet is walk-in placeholder
	ScheduledAt     *time.Time
	DurationMinutes *int
	VisitTypeID     *string
	AssigneeUserID  string // optional staff user id
	RoomID          string // optional room under site
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
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status, COALESCE(v.notes,''), v.source, v.created_at,
			'', '', '',
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			COALESCE(p.is_walkin_placeholder, false), COALESCE(v.callback_phone,'')
		FROM visits.visits v
		LEFT JOIN practice.sites si ON si.id = v.site_id
		LEFT JOIN pets.pets p ON p.id = v.pet_id
		WHERE v.pet_id = $1 AND v.deleted_at IS NULL
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisitsFull(rows)
}

func (s *Store) ListPracticeVisitsByStatus(ctx context.Context, practiceID, status string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			COALESCE(p.is_walkin_placeholder, false), COALESCE(v.callback_phone,'')
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		LEFT JOIN practice.sites si ON si.id = v.site_id
		WHERE v.practice_id = $1 AND v.status = $2 AND v.deleted_at IS NULL
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC
		LIMIT 100`, practiceID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisitsFull(rows)
}

// ConsultationListItem is a cabinet consultation with report summary flags (no audio keys).
// Included: walk-ins (consultation_session) and agenda visits with a persisted CR.
type ConsultationListItem struct {
	Visit
	HasReport        bool   `json:"hasReport"`
	HasAudio         bool   `json:"hasAudio"`
	AudioDurationSec int    `json:"audioDurationSec,omitempty"`
	ReportStatus     string `json:"reportStatus,omitempty"`
}

// ListConsultationsFilter filters practice consultations (walk-ins + visits with CR).
type ListConsultationsFilter struct {
	SiteID   string // empty or "all" = all sites
	Status   string
	Query    string
	From     *time.Time
	To       *time.Time
	HasAudio *bool
	Limit    int
	Offset   int
}

// ListPracticeConsultations returns cabinet consultations newest first:
// walk-ins (consultation_session) or visits with a persisted report.
func (s *Store) ListPracticeConsultations(ctx context.Context, practiceID string, f ListConsultationsFilter) ([]ConsultationListItem, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	offset := max(f.Offset, 0)

	args := []any{practiceID}
	where := []string{
		`v.practice_id = $1`,
		sqlVisitIsConsultationHistory,
		`v.deleted_at IS NULL`,
	}
	argN := 2

	if sid := strings.TrimSpace(f.SiteID); sid != "" && sid != "all" {
		resolved, err := s.ResolveSiteID(ctx, practiceID, sid, true)
		if err != nil {
			return nil, err
		}
		where = append(where, fmt.Sprintf(`v.site_id = $%d`, argN))
		args = append(args, resolved)
		argN++
	}

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
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			COALESCE(p.is_walkin_placeholder, false), COALESCE(v.callback_phone,''),
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
				SELECT MAX(r.audio_duration_sec)
				FROM visits.visit_reports r
				WHERE r.visit_id = v.id
				  AND r.audio_duration_sec IS NOT NULL
				  AND r.audio_duration_sec > 0
			), 0) AS audio_duration_sec,
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
		LEFT JOIN practice.sites si ON si.id = v.site_id
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
			&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.SiteName, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
			&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
			&v.AddressText, &v.Lat, &v.Lng, &v.ConsultationSession, &v.IsWalkinPlaceholder, &v.CallbackPhone,
			&item.HasReport, &item.HasAudio, &item.AudioDurationSec, &item.ReportStatus,
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

func (s *Store) ListPracticePendingVetActions(ctx context.Context, practiceID, siteID string) ([]Visit, error) {
	args := []any{practiceID}
	siteFilter := ""
	if sid := strings.TrimSpace(siteID); sid != "" && sid != "all" {
		resolved, err := s.ResolveSiteID(ctx, practiceID, sid, true)
		if err != nil {
			return nil, err
		}
		siteFilter = ` AND v.site_id = $2`
		args = append(args, resolved)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			COALESCE(p.is_walkin_placeholder, false), COALESCE(v.callback_phone,'')
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		LEFT JOIN practice.sites si ON si.id = v.site_id
		WHERE v.practice_id = $1
		  AND v.pending_action_by = 'vet'
		  AND v.status IN ('requested', 'reschedule_pending')
		  AND v.deleted_at IS NULL`+siteFilter+`
		ORDER BY COALESCE(v.scheduled_at, v.proposed_scheduled_at, v.created_at) DESC
		LIMIT 100`, args...)
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
			&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.SiteName, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
			&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
			&v.AddressText, &v.Lat, &v.Lng, &v.ConsultationSession, &v.IsWalkinPlaceholder, &v.CallbackPhone,
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

func (s *Store) resolveCreateVisitSiteID(ctx context.Context, in *CreateVisitInput) error {
	siteID, err := s.ResolveBookingSiteID(ctx, in.PracticeID, in.SiteID)
	if err != nil {
		return err
	}
	in.SiteID = siteID
	return nil
}

func (s *Store) validateCreateVisitResources(ctx context.Context, in *CreateVisitInput) error {
	if err := s.ValidateVisitAssignee(ctx, in.PracticeID, in.AssigneeUserID); err != nil {
		return err
	}
	return s.ValidateVisitRoom(ctx, in.PracticeID, in.SiteID, in.RoomID)
}

// nullUUIDArg returns nil for empty string so Postgres stores NULL.
func nullUUIDArg(id string) any {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	return id
}

// checkResourceOverlapsTx rejects hard conflicts on assignee and room (when set).
func (s *Store) checkResourceOverlapsTx(ctx context.Context, tx pgx.Tx, practiceID, siteID, assigneeUserID, roomID string, start time.Time, durMin int, excludeVisitID string) error {
	end := start.Add(time.Duration(durMin) * time.Minute)
	if aid := strings.TrimSpace(assigneeUserID); aid != "" {
		var n int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE practice_id = $1
			  AND assignee_user_id = $2::uuid
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND ($5 = '' OR id::text <> $5)
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $4
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $6) || ' minutes')::interval > $3`,
			practiceID, aid, start, end, excludeVisitID, durMin,
		).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: assignee_busy", ErrValidation)
		}
	}
	if rid := strings.TrimSpace(roomID); rid != "" {
		var n int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE room_id = $1::uuid
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND ($4 = '' OR id::text <> $4)
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
			rid, start, end, excludeVisitID, durMin,
		).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: room_busy", ErrValidation)
		}
	}
	_ = siteID // site already locked by caller; resources are global to assignee/room
	return nil
}

// CheckVisitResourceConflicts reports assignee_busy / room_busy for a proposed slot (no lock).
func (s *Store) CheckVisitResourceConflicts(ctx context.Context, practiceID, assigneeUserID, roomID string, start time.Time, durMin int, excludeVisitID string) error {
	end := start.Add(time.Duration(durMin) * time.Minute)
	if aid := strings.TrimSpace(assigneeUserID); aid != "" {
		var n int
		if err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE practice_id = $1
			  AND assignee_user_id = $2::uuid
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND ($5 = '' OR id::text <> $5)
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $4
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $6) || ' minutes')::interval > $3`,
			practiceID, aid, start, end, excludeVisitID, durMin,
		).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: assignee_busy", ErrValidation)
		}
	}
	if rid := strings.TrimSpace(roomID); rid != "" {
		var n int
		if err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE room_id = $1::uuid
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND ($4 = '' OR id::text <> $4)
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
			rid, start, end, excludeVisitID, durMin,
		).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w: room_busy", ErrValidation)
		}
	}
	return nil
}

func (s *Store) CreateVisit(ctx context.Context, in CreateVisitInput) (Visit, error) {
	if err := s.resolveCreateVisitSiteID(ctx, &in); err != nil {
		return Visit{}, err
	}
	if err := s.validateCreateVisitResources(ctx, &in); err != nil {
		return Visit{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback(ctx)

	var siteActive bool
	if err := tx.QueryRow(ctx, `
		SELECT active FROM practice.sites
		WHERE id = $1 AND practice_id = $2 FOR UPDATE`, in.SiteID, in.PracticeID,
	).Scan(&siteActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Visit{}, fmt.Errorf("%w: invalid_site", ErrValidation)
		}
		return Visit{}, err
	}
	if !siteActive {
		return Visit{}, fmt.Errorf("%w: site_inactive", ErrValidation)
	}

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
	err = tx.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, site_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by, request_preconsult, consultation_session, visit_type_id, assignee_user_id, room_id, callback_phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id::text, pet_id::text, practice_id::text, site_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by, COALESCE(request_preconsult,false), COALESCE(consultation_session,false),
			COALESCE(visit_type_id::text,''), COALESCE(assignee_user_id::text,''), COALESCE(room_id::text,''), COALESCE(callback_phone,'')`,
		id, in.PetID, in.PracticeID, in.SiteID, in.ScheduledAt, status, in.Notes, in.Source, in.DurationMinutes, pending, in.RequestPreconsult, in.ConsultationSession, in.VisitTypeID,
		nullUUIDArg(in.AssigneeUserID), nullUUIDArg(in.RoomID), strings.TrimSpace(in.CallbackPhone),
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy, &v.RequestPreconsult, &v.ConsultationSession, &v.VisitTypeID,
		&v.AssigneeUserID, &v.RoomID, &v.CallbackPhone)
	if err != nil {
		return Visit{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Visit{}, err
	}
	return v, nil
}

// CreateVisitBooked serializes bookings for a site and re-checks overlap under lock.
func (s *Store) CreateVisitBooked(ctx context.Context, in CreateVisitInput) (Visit, error) {
	if in.ScheduledAt == nil {
		return Visit{}, fmt.Errorf("%w: scheduled_required", ErrValidation)
	}
	if err := s.resolveCreateVisitSiteID(ctx, &in); err != nil {
		return Visit{}, err
	}
	if err := s.validateCreateVisitResources(ctx, &in); err != nil {
		return Visit{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback(ctx)

	// Lock site row first (same resource as DeactivateSite) then re-check active.
	var siteActive bool
	if err := tx.QueryRow(ctx, `
		SELECT active FROM practice.sites
		WHERE id = $1 AND practice_id = $2 FOR UPDATE`, in.SiteID, in.PracticeID,
	).Scan(&siteActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Visit{}, fmt.Errorf("%w: invalid_site", ErrValidation)
		}
		return Visit{}, err
	}
	if !siteActive {
		return Visit{}, fmt.Errorf("%w: site_inactive", ErrValidation)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (site_id, practice_id)
		VALUES ($1, $2) ON CONFLICT (site_id) DO NOTHING`, in.SiteID, in.PracticeID); err != nil {
		return Visit{}, err
	}
	var locked string
	if err := tx.QueryRow(ctx, `
		SELECT site_id::text FROM practice.vet_schedule WHERE site_id = $1 FOR UPDATE`, in.SiteID,
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
		// Legacy single-queue: only visits with neither assignee nor room compete for the site slot.
		// Assigned/roomed visits may run in parallel on the same site (model 2).
		unassignedOnly := strings.TrimSpace(in.AssigneeUserID) == "" && strings.TrimSpace(in.RoomID) == ""
		if unassignedOnly {
			var n int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*)::int FROM visits.visits
				WHERE site_id = $1
				  AND deleted_at IS NULL
				  AND status IN ('requested', 'confirmed', 'reschedule_pending')
				  AND COALESCE(consultation_session, false) = false
				  AND assignee_user_id IS NULL
				  AND room_id IS NULL
				  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
				  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
				  AND COALESCE(proposed_scheduled_at, scheduled_at)
				      + (COALESCE(duration_minutes, $4) || ' minutes')::interval > $2`,
				in.SiteID, *in.ScheduledAt, end, dur,
			).Scan(&n); err != nil {
				return Visit{}, err
			}
			if n > 0 {
				return Visit{}, fmt.Errorf("%w: slot_taken", ErrValidation)
			}
		}
		if err := s.checkResourceOverlapsTx(ctx, tx, in.PracticeID, in.SiteID, in.AssigneeUserID, in.RoomID, *in.ScheduledAt, dur, ""); err != nil {
			return Visit{}, err
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
		INSERT INTO visits.visits (id, pet_id, practice_id, site_id, scheduled_at, status, notes, source, duration_minutes, pending_action_by, request_preconsult, consultation_session, visit_type_id, assignee_user_id, room_id, callback_phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id::text, pet_id::text, practice_id::text, site_id::text, scheduled_at, status, COALESCE(notes,''), source, created_at,
			duration_minutes, proposed_scheduled_at, pending_action_by, COALESCE(request_preconsult,false), COALESCE(consultation_session,false),
			COALESCE(visit_type_id::text,''), COALESCE(assignee_user_id::text,''), COALESCE(room_id::text,''), COALESCE(callback_phone,'')`,
		id, in.PetID, in.PracticeID, in.SiteID, in.ScheduledAt, status, in.Notes, in.Source, in.DurationMinutes, pending, in.RequestPreconsult, in.ConsultationSession, in.VisitTypeID,
		nullUUIDArg(in.AssigneeUserID), nullUUIDArg(in.RoomID), strings.TrimSpace(in.CallbackPhone),
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy, &v.RequestPreconsult, &v.ConsultationSession, &v.VisitTypeID,
		&v.AssigneeUserID, &v.RoomID, &v.CallbackPhone)
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
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status, COALESCE(v.notes,''), v.source, v.created_at,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.request_preconsult,false), COALESCE(v.consultation_session,false),
			v.waiting_room_at,
			COALESCE(v.assignee_user_id::text,''), COALESCE(au.full_name,''),
			COALESCE(v.room_id::text,''), COALESCE(rm.name,''),
			COALESCE(v.callback_phone,''), COALESCE(p.is_walkin_placeholder, false),
			COALESCE(v.visit_type_id::text,''), COALESCE(vt.name,'')
		FROM visits.visits v
		LEFT JOIN practice.sites si ON si.id = v.site_id
		LEFT JOIN identity.users au ON au.id = v.assignee_user_id
		LEFT JOIN practice.rooms rm ON rm.id = v.room_id
		LEFT JOIN pets.pets p ON p.id = v.pet_id
		LEFT JOIN practice.visit_types vt ON vt.id = v.visit_type_id
		WHERE v.id = $1 AND v.deleted_at IS NULL`, id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.SiteName, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
		&v.AddressText, &v.Lat, &v.Lng, &v.RequestPreconsult, &v.ConsultationSession, &v.WaitingRoomAt,
		&v.AssigneeUserID, &v.AssigneeName, &v.RoomID, &v.RoomName,
		&v.CallbackPhone, &v.IsWalkinPlaceholder,
		&v.VisitTypeID, &v.VisitTypeName)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

// GetPracticeVisit returns a non-deleted visit for the practice, with pet/owner names.
func (s *Store) GetPracticeVisit(ctx context.Context, practiceID, visitID string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, COALESCE(v.site_id::text,''), COALESCE(si.name,''),
			v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.consultation_session, false),
			COALESCE(p.is_walkin_placeholder, false), COALESCE(v.callback_phone,'')
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		LEFT JOIN practice.sites si ON si.id = v.site_id
		WHERE v.id = $1 AND v.practice_id = $2 AND v.deleted_at IS NULL`,
		visitID, practiceID,
	).Scan(
		&v.ID, &v.PetID, &v.PracticeID, &v.SiteID, &v.SiteName, &v.ScheduledAt, &v.Status,
		&v.Notes, &v.Source, &v.CreatedAt,
		&v.PetName, &v.ClientName, &v.ClientID,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
		&v.AddressText, &v.Lat, &v.Lng, &v.ConsultationSession, &v.IsWalkinPlaceholder, &v.CallbackPhone,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

// UpdateVisitNotes sets agenda / desk notes (calendar.manage — not clinical CR).
// If callbackPhone is non-nil, also updates the walk-in callback number (empty clears).
func (s *Store) UpdateVisitNotes(ctx context.Context, id, notes string, callbackPhone *string) (Visit, error) {
	if len(notes) > 1000 {
		notes = notes[:1000]
	}
	var tag pgconn.CommandTag
	var err error
	if callbackPhone != nil {
		phone := strings.TrimSpace(*callbackPhone)
		if len(phone) > 40 {
			phone = phone[:40]
		}
		tag, err = s.pool.Exec(ctx, `
			UPDATE visits.visits SET notes = $2, callback_phone = $3 WHERE id = $1 AND deleted_at IS NULL`,
			id, notes, phone)
	} else {
		tag, err = s.pool.Exec(ctx, `
			UPDATE visits.visits SET notes = $2 WHERE id = $1 AND deleted_at IS NULL`, id, notes)
	}
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

// RescheduleVisitDirect sets scheduled_at immediately (staff unilateral move).
// Status stays confirmed (or becomes confirmed from reschedule_pending only).
func (s *Store) RescheduleVisitDirect(ctx context.Context, id string, at time.Time) (Visit, error) {
	cur, err := s.GetVisit(ctx, id)
	if err != nil {
		return Visit{}, err
	}
	if cur.ConsultationSession {
		return Visit{}, ErrNotFound
	}
	if cur.Status != "confirmed" && cur.Status != "reschedule_pending" {
		return Visit{}, ErrNotFound
	}
	dur := 30
	if cur.DurationMinutes != nil {
		dur = *cur.DurationMinutes
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (site_id, practice_id)
		VALUES ($1, $2) ON CONFLICT (site_id) DO NOTHING`, cur.SiteID, cur.PracticeID); err != nil {
		return Visit{}, err
	}
	var locked string
	if err := tx.QueryRow(ctx, `
		SELECT site_id::text FROM practice.vet_schedule WHERE site_id = $1 FOR UPDATE`, cur.SiteID,
	).Scan(&locked); err != nil {
		return Visit{}, err
	}
	end := at.Add(time.Duration(dur) * time.Minute)
	unassignedOnly := strings.TrimSpace(cur.AssigneeUserID) == "" && strings.TrimSpace(cur.RoomID) == ""
	if unassignedOnly {
		var n int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE site_id = $1
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND assignee_user_id IS NULL
			  AND room_id IS NULL
			  AND id::text <> $4
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
			cur.SiteID, at, end, id, dur,
		).Scan(&n); err != nil {
			return Visit{}, err
		}
		if n > 0 {
			return Visit{}, fmt.Errorf("%w: slot_taken", ErrValidation)
		}
	}
	if err := s.checkResourceOverlapsTx(ctx, tx, cur.PracticeID, cur.SiteID, cur.AssigneeUserID, cur.RoomID, at, dur, id); err != nil {
		return Visit{}, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE visits.visits
		SET scheduled_at = $2,
			proposed_scheduled_at = NULL,
			pending_action_by = NULL,
			status_before_reschedule = NULL,
			status = 'confirmed'
		WHERE id = $1
		  AND deleted_at IS NULL
		  AND COALESCE(consultation_session, false) = false
		  AND status IN ('confirmed', 'reschedule_pending')`,
		id, at,
	)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return Visit{}, err
	}
	return s.GetVisit(ctx, id)
}

// SetVisitResources updates assignee and/or room (nil pointer = leave unchanged, empty string = clear).
func (s *Store) SetVisitResources(ctx context.Context, id string, assigneeUserID, roomID *string) (Visit, error) {
	cur, err := s.GetVisit(ctx, id)
	if err != nil {
		return Visit{}, err
	}
	nextAssignee := cur.AssigneeUserID
	if assigneeUserID != nil {
		nextAssignee = strings.TrimSpace(*assigneeUserID)
	}
	nextRoom := cur.RoomID
	if roomID != nil {
		nextRoom = strings.TrimSpace(*roomID)
	}
	// New assignee must be on the calendar; keeping an existing (possibly hidden) assignee is allowed.
	requireInCalendar := nextAssignee != "" && nextAssignee != cur.AssigneeUserID
	if err := s.validateVisitAssignee(ctx, cur.PracticeID, nextAssignee, requireInCalendar); err != nil {
		return Visit{}, err
	}
	if err := s.ValidateVisitRoom(ctx, cur.PracticeID, cur.SiteID, nextRoom); err != nil {
		return Visit{}, err
	}
	if cur.ScheduledAt != nil && !cur.ConsultationSession {
		dur := 30
		if cur.DurationMinutes != nil {
			dur = *cur.DurationMinutes
		}
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return Visit{}, err
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.vet_schedule (site_id, practice_id)
			VALUES ($1, $2) ON CONFLICT (site_id) DO NOTHING`, cur.SiteID, cur.PracticeID); err != nil {
			return Visit{}, err
		}
		var locked string
		if err := tx.QueryRow(ctx, `
			SELECT site_id::text FROM practice.vet_schedule WHERE site_id = $1 FOR UPDATE`, cur.SiteID,
		).Scan(&locked); err != nil {
			return Visit{}, err
		}
		if err := s.checkResourceOverlapsTx(ctx, tx, cur.PracticeID, cur.SiteID, nextAssignee, nextRoom, *cur.ScheduledAt, dur, id); err != nil {
			return Visit{}, err
		}
		// Clearing both resources lands in the legacy unassigned queue.
		if strings.TrimSpace(nextAssignee) == "" && strings.TrimSpace(nextRoom) == "" {
			end := cur.ScheduledAt.Add(time.Duration(dur) * time.Minute)
			var n int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*)::int FROM visits.visits
				WHERE site_id = $1
				  AND deleted_at IS NULL
				  AND status IN ('requested', 'confirmed', 'reschedule_pending')
				  AND COALESCE(consultation_session, false) = false
				  AND assignee_user_id IS NULL
				  AND room_id IS NULL
				  AND id::text <> $4
				  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
				  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
				  AND COALESCE(proposed_scheduled_at, scheduled_at)
				      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
				cur.SiteID, *cur.ScheduledAt, end, id, dur,
			).Scan(&n); err != nil {
				return Visit{}, err
			}
			if n > 0 {
				return Visit{}, fmt.Errorf("%w: slot_taken", ErrValidation)
			}
		}
		tag, err := tx.Exec(ctx, `
			UPDATE visits.visits
			SET assignee_user_id = $2, room_id = $3
			WHERE id = $1 AND deleted_at IS NULL`,
			id, nullUUIDArg(nextAssignee), nullUUIDArg(nextRoom))
		if err != nil {
			return Visit{}, err
		}
		if tag.RowsAffected() == 0 {
			return Visit{}, ErrNotFound
		}
		if err := tx.Commit(ctx); err != nil {
			return Visit{}, err
		}
		return s.GetVisit(ctx, id)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET assignee_user_id = $2, room_id = $3
		WHERE id = $1 AND deleted_at IS NULL`,
		id, nullUUIDArg(nextAssignee), nullUUIDArg(nextRoom))
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
}

// MarkVisitWaitingRoom sets waiting_room_at = now() when currently null. Returns (visit, newlyMarked, err).
func (s *Store) MarkVisitWaitingRoom(ctx context.Context, id string) (Visit, bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits
		SET waiting_room_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND waiting_room_at IS NULL
		  AND status IN ('requested', 'confirmed', 'reschedule_pending')`, id)
	if err != nil {
		return Visit{}, false, err
	}
	v, gerr := s.GetVisit(ctx, id)
	if gerr != nil {
		return Visit{}, false, gerr
	}
	return v, tag.RowsAffected() > 0, nil
}

// ClearVisitWaitingRoom clears the waiting-room flag.
func (s *Store) ClearVisitWaitingRoom(ctx context.Context, id string) (Visit, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visits SET waiting_room_at = NULL
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return Visit{}, err
	}
	if tag.RowsAffected() == 0 {
		return Visit{}, ErrNotFound
	}
	return s.GetVisit(ctx, id)
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

// SoftDeleteVisit marks a consultation-history visit as soft-deleted (hidden from
// /consultations and calendar). Eligible = walk-in OR (done/cancelled + persisted CR).
func (s *Store) SoftDeleteVisit(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visits v
		SET deleted_at = now()
		WHERE v.id = $1
		  AND v.deleted_at IS NULL
		  AND `+sqlVisitSoftDeleteEligible+`
		RETURNING v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status, COALESCE(v.notes,''), v.source, v.created_at,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng, COALESCE(v.request_preconsult,false), COALESCE(v.consultation_session,false)`,
		id,
	).Scan(&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
		&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
		&v.AddressText, &v.Lat, &v.Lng, &v.RequestPreconsult, &v.ConsultationSession)
	if errors.Is(err, pgx.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
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
			  AND v.deleted_at IS NULL
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
			status_before_reschedule = NULL,
			callback_phone = ''
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
	cur, err := s.GetVisit(ctx, id)
	if err != nil {
		return Visit{}, err
	}
	if cur.Status != "reschedule_pending" || cur.ProposedScheduledAt == nil || cur.ConsultationSession {
		return Visit{}, ErrNotFound
	}
	at := *cur.ProposedScheduledAt
	dur := 30
	if cur.DurationMinutes != nil {
		dur = *cur.DurationMinutes
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (site_id, practice_id)
		VALUES ($1, $2) ON CONFLICT (site_id) DO NOTHING`, cur.SiteID, cur.PracticeID); err != nil {
		return Visit{}, err
	}
	var locked string
	if err := tx.QueryRow(ctx, `
		SELECT site_id::text FROM practice.vet_schedule WHERE site_id = $1 FOR UPDATE`, cur.SiteID,
	).Scan(&locked); err != nil {
		return Visit{}, err
	}
	end := at.Add(time.Duration(dur) * time.Minute)
	unassignedOnly := strings.TrimSpace(cur.AssigneeUserID) == "" && strings.TrimSpace(cur.RoomID) == ""
	if unassignedOnly {
		var n int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM visits.visits
			WHERE site_id = $1
			  AND deleted_at IS NULL
			  AND status IN ('requested', 'confirmed', 'reschedule_pending')
			  AND COALESCE(consultation_session, false) = false
			  AND assignee_user_id IS NULL
			  AND room_id IS NULL
			  AND id::text <> $4
			  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
			  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
			  AND COALESCE(proposed_scheduled_at, scheduled_at)
			      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
			cur.SiteID, at, end, id, dur,
		).Scan(&n); err != nil {
			return Visit{}, err
		}
		if n > 0 {
			return Visit{}, fmt.Errorf("%w: slot_taken", ErrValidation)
		}
	}
	if err := s.checkResourceOverlapsTx(ctx, tx, cur.PracticeID, cur.SiteID, cur.AssigneeUserID, cur.RoomID, at, dur, id); err != nil {
		return Visit{}, err
	}
	tag, err := tx.Exec(ctx, `
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
	if err := tx.Commit(ctx); err != nil {
		return Visit{}, err
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
