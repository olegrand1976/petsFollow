package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type VisitReport struct {
	ID                   string     `json:"id"`
	VisitID              string     `json:"visitId"`
	AuthorUserID         string     `json:"authorUserId"`
	Status               string     `json:"status"`
	BodyText             string     `json:"bodyText"`
	AudioURL             string     `json:"audioUrl,omitempty"`
	AudioObjectKey       string     `json:"-"`
	HasAudio             bool       `json:"hasAudio"`
	TranscriptText       string     `json:"transcriptText,omitempty"`
	ImprovedText         string     `json:"improvedText,omitempty"`
	ClientAudioConsentAt *time.Time `json:"clientAudioConsentAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	FinalizedAt          *time.Time `json:"finalizedAt,omitempty"`
}

const visitReportReturning = `
	id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
	COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
	COALESCE(improved_text,''), client_audio_consent_at, created_at, updated_at, finalized_at`

func scanVisitReport(row pgx.Row) (VisitReport, error) {
	var r VisitReport
	err := row.Scan(
		&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.ClientAudioConsentAt, &r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt,
	)
	if err == nil {
		r.HasAudio = strings.TrimSpace(r.AudioObjectKey) != ""
	}
	return r, err
}

func (s *Store) UpsertVisitReport(ctx context.Context, visitID, authorUserID, bodyText string) (VisitReport, error) {
	id := uuid.NewString()
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		INSERT INTO visits.visit_reports (id, visit_id, author_user_id, status, body_text)
		VALUES ($1, $2, $3, 'draft', $4)
		ON CONFLICT (visit_id, author_user_id) DO UPDATE
			SET body_text = EXCLUDED.body_text, updated_at = NOW()
			WHERE visits.visit_reports.status = 'draft'
		RETURNING `+visitReportReturning,
		id, visitID, authorUserID, bodyText))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrConflict
	}
	return r, err
}

func (s *Store) GetVisitReport(ctx context.Context, visitID, authorUserID string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		SELECT `+visitReportReturning+`
		FROM visits.visit_reports WHERE visit_id=$1 AND author_user_id=$2`, visitID, authorUserID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) GetVisitReportByID(ctx context.Context, reportID string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		SELECT `+visitReportReturning+`
		FROM visits.visit_reports WHERE id=$1`, reportID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) UpdateVisitReportAudio(ctx context.Context, reportID, audioURL, objectKey string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET audio_url=$2, audio_object_key=$3, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning,
		reportID, audioURL, objectKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

// ClearVisitReportAudio wipes audio fields (after delete from media store). Allowed on draft or final.
func (s *Store) ClearVisitReportAudio(ctx context.Context, reportID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE visits.visit_reports
		SET audio_url='', audio_object_key='', updated_at=NOW()
		WHERE id=$1`, reportID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateVisitReportTranscript(ctx context.Context, reportID, transcript string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET transcript_text=$2, body_text=CASE WHEN COALESCE(body_text,'')='' THEN $2 ELSE body_text END, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning,
		reportID, transcript))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

// MarkVisitReportAudioConsent stamps client_audio_consent_at (idempotent if already set).
func (s *Store) MarkVisitReportAudioConsent(ctx context.Context, reportID string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET client_audio_consent_at = COALESCE(client_audio_consent_at, NOW()), updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning, reportID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) UpdateVisitReportImproved(ctx context.Context, reportID, improved string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET improved_text=$2, body_text=$2, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning,
		reportID, improved))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) FinalizeVisitReport(ctx context.Context, reportID string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET status='final', finalized_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning, reportID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

// sqlVisitReportIsPersisted — predicate on alias `r` (visits.visit_reports):
// any author content beyond an empty Ensure draft.
const sqlVisitReportIsPersisted = `(
  r.status = 'final'
  OR length(trim(COALESCE(r.body_text, ''))) > 0
  OR length(trim(COALESCE(r.transcript_text, ''))) > 0
  OR length(trim(COALESCE(r.improved_text, ''))) > 0
  OR length(trim(COALESCE(r.audio_object_key, ''))) > 0
)`

// VisitHasPersistedReport is true when any author saved content (not an empty Ensure draft).
func (s *Store) VisitHasPersistedReport(ctx context.Context, visitID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM visits.visit_reports r
			WHERE r.visit_id = $1::uuid
			  AND `+sqlVisitReportIsPersisted+`
		)`, visitID).Scan(&ok)
	return ok, err
}

// EnsureVisitReport returns existing or creates empty draft.
func (s *Store) EnsureVisitReport(ctx context.Context, visitID, authorUserID string) (VisitReport, error) {
	r, err := s.GetVisitReport(ctx, visitID, authorUserID)
	if err == nil {
		return r, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return VisitReport{}, err
	}
	return s.UpsertVisitReport(ctx, visitID, authorUserID, "")
}

// VisitReportSummary is a visit report with author display name for multi-author listing.
type VisitReportSummary struct {
	VisitReport
	AuthorFullName string `json:"authorFullName,omitempty"`
	Mine           bool   `json:"mine"`
}

// ListVisitReportsForVisit returns all reports for a visit (any author), newest first.
func (s *Store) ListVisitReportsForVisit(ctx context.Context, visitID, viewerUserID string) ([]VisitReportSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			r.id::text, r.visit_id::text, r.author_user_id::text, r.status, COALESCE(r.body_text,''),
			COALESCE(r.audio_url,''), COALESCE(r.audio_object_key,''), COALESCE(r.transcript_text,''),
			COALESCE(r.improved_text,''), r.client_audio_consent_at, r.created_at, r.updated_at, r.finalized_at,
			COALESCE(NULLIF(TRIM(u.full_name), ''), u.email) AS author_full_name
		FROM visits.visit_reports r
		JOIN identity.users u ON u.id = r.author_user_id
		WHERE r.visit_id = $1
		ORDER BY r.updated_at DESC`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisitReportSummary
	for rows.Next() {
		var sum VisitReportSummary
		var r VisitReport
		if err := rows.Scan(
			&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
			&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
			&r.ClientAudioConsentAt, &r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt,
			&sum.AuthorFullName,
		); err != nil {
			return nil, err
		}
		r.HasAudio = strings.TrimSpace(r.AudioObjectKey) != ""
		sum.VisitReport = r
		sum.Mine = r.AuthorUserID == viewerUserID
		out = append(out, sum)
	}
	if out == nil {
		out = []VisitReportSummary{}
	}
	return out, rows.Err()
}

// GetVisitReportWithAudio returns any draft report for the visit that still has audio stored.
func (s *Store) GetVisitReportWithAudio(ctx context.Context, visitID string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		SELECT `+visitReportReturning+`
		FROM visits.visit_reports
		WHERE visit_id = $1::uuid
		  AND status = 'draft'
		  AND length(trim(COALESCE(audio_object_key, ''))) > 0
		ORDER BY updated_at DESC
		LIMIT 1`, visitID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}
