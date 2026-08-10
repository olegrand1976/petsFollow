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
	AudioDurationSec     int        `json:"audioDurationSec,omitempty"`
	TranscriptText       string     `json:"transcriptText,omitempty"`
	ImprovedText         string     `json:"improvedText,omitempty"`
	IsReference          bool       `json:"isReference"`
	ClientAudioConsentAt *time.Time `json:"clientAudioConsentAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	FinalizedAt          *time.Time `json:"finalizedAt,omitempty"`
}

const visitReportReturning = `
	id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
	COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(audio_duration_sec, 0),
	COALESCE(transcript_text,''),
	COALESCE(improved_text,''), COALESCE(is_reference, false), client_audio_consent_at, created_at, updated_at, finalized_at`

func scanVisitReport(row pgx.Row) (VisitReport, error) {
	var r VisitReport
	err := row.Scan(
		&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.AudioDurationSec, &r.TranscriptText, &r.ImprovedText,
		&r.IsReference, &r.ClientAudioConsentAt, &r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt,
	)
	if err == nil {
		r.HasAudio = strings.TrimSpace(r.AudioObjectKey) != ""
	}
	return r, err
}

func (s *Store) UpsertVisitReport(ctx context.Context, visitID, authorUserID, bodyText string) (VisitReport, error) {
	return s.UpsertVisitReportFields(ctx, visitID, authorUserID, bodyText, nil)
}

// UpsertVisitReportFields upserts body_text and optionally transcript_text (when transcript != nil).
func (s *Store) UpsertVisitReportFields(ctx context.Context, visitID, authorUserID, bodyText string, transcript *string) (VisitReport, error) {
	id := uuid.NewString()
	var transcriptVal any
	updateTranscript := transcript != nil
	if updateTranscript {
		transcriptVal = *transcript
	}
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		INSERT INTO visits.visit_reports (id, visit_id, author_user_id, status, body_text, transcript_text)
		VALUES ($1, $2, $3, 'draft', $4, COALESCE($5, ''))
		ON CONFLICT (visit_id, author_user_id) DO UPDATE
			SET body_text = EXCLUDED.body_text,
				transcript_text = CASE WHEN $6 THEN COALESCE($5, '') ELSE visits.visit_reports.transcript_text END,
				updated_at = NOW()
			WHERE visits.visit_reports.status = 'draft'
		RETURNING `+visitReportReturning,
		id, visitID, authorUserID, bodyText, transcriptVal, updateTranscript))
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

func (s *Store) UpdateVisitReportAudio(ctx context.Context, reportID, audioURL, objectKey string, durationSec int) (VisitReport, error) {
	if durationSec < 0 {
		durationSec = 0
	}
	if durationSec > 7200 {
		durationSec = 7200
	}
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET audio_url=$2, audio_object_key=$3,
			audio_duration_sec=CASE WHEN $4 > 0 THEN $4 ELSE NULL END,
			updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING `+visitReportReturning,
		reportID, audioURL, objectKey, durationSec))
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

// ClearVisitReportAudio wipes audio blob fields (after delete from media store).
// audio_duration_sec is kept as non-PHI metadata for history display.
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

// UpdateVisitReportImprovedIfRunActive persists improved text only while the
// improve-advanced run is still queued/running — blocks cancel races.
func (s *Store) UpdateVisitReportImprovedIfRunActive(ctx context.Context, reportID, runID, improved string) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports vr
		SET improved_text = $2, body_text = $2, updated_at = NOW()
		FROM rag.improve_runs ir
		WHERE vr.id = $1::uuid
		  AND vr.status = 'draft'
		  AND ir.id = $3::uuid
		  AND ir.report_id = vr.id
		  AND ir.status IN ('queued', 'running')
		RETURNING vr.id::text, vr.visit_id::text, vr.author_user_id::text, vr.status, COALESCE(vr.body_text,''),
			COALESCE(vr.audio_url,''), COALESCE(vr.audio_object_key,''), COALESCE(vr.audio_duration_sec, 0),
			COALESCE(vr.transcript_text,''),
			COALESCE(vr.improved_text,''), COALESCE(vr.is_reference, false), vr.client_audio_consent_at,
			vr.created_at, vr.updated_at, vr.finalized_at`,
		reportID, improved, runID))
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

// SetVisitReportReference marks a finalized CR as a quality reference for continuous improvement.
func (s *Store) SetVisitReportReference(ctx context.Context, reportID string, isReference bool) (VisitReport, error) {
	r, err := scanVisitReport(s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET is_reference=$2, updated_at=NOW()
		WHERE id=$1 AND status='final'
		RETURNING `+visitReportReturning, reportID, isReference))
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

// sqlVisitIsConsultationHistory — predicate on alias `v` (visits.visits):
// walk-in session or visit with a persisted CR (same scope as /consultations list).
const sqlVisitIsConsultationHistory = `(
  COALESCE(v.consultation_session, false) = true
  OR EXISTS (
    SELECT 1 FROM visits.visit_reports r
    WHERE r.visit_id = v.id AND ` + sqlVisitReportIsPersisted + `
  )
)`

// sqlVisitSoftDeleteEligible — narrower than list: walk-in (any status), or
// done/cancelled visit with persisted CR. Blocks soft-delete of upcoming agenda RDVs
// that only have a draft CR (would remove them from the calendar).
const sqlVisitSoftDeleteEligible = `(
  COALESCE(v.consultation_session, false) = true
  OR (
    v.status IN ('done', 'cancelled')
    AND EXISTS (
      SELECT 1 FROM visits.visit_reports r
      WHERE r.visit_id = v.id AND ` + sqlVisitReportIsPersisted + `
    )
  )
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
			COALESCE(r.audio_url,''), COALESCE(r.audio_object_key,''), COALESCE(r.audio_duration_sec, 0),
			COALESCE(r.transcript_text,''),
			COALESCE(r.improved_text,''), COALESCE(r.is_reference, false), r.client_audio_consent_at, r.created_at, r.updated_at, r.finalized_at,
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
			&r.AudioURL, &r.AudioObjectKey, &r.AudioDurationSec, &r.TranscriptText, &r.ImprovedText,
			&r.IsReference, &r.ClientAudioConsentAt, &r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt,
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

// ClientFinalReport is a finalized CR safe to expose to the pet owner (no audio/PHI drafts).
type ClientFinalReport struct {
	ID          string     `json:"id"`
	AuthorName  string     `json:"authorName"`
	BodyText    string     `json:"bodyText"`
	FinalizedAt *time.Time `json:"finalizedAt,omitempty"`
}

// AttachFinalReportFlags sets HasFinalReport and ReportStatus (final|draft) on owner visit lists.
// Prefer final over draft when both exist; draft only if persisted content (not empty Ensure).
// Never loads body_text into the visit payload.
func (s *Store) AttachFinalReportFlags(ctx context.Context, visits []Visit) error {
	if len(visits) == 0 {
		return nil
	}
	ids := make([]string, len(visits))
	for i, v := range visits {
		ids[i] = v.ID
	}
	rows, err := s.pool.Query(ctx, `
		SELECT visit_id::text,
			CASE
				WHEN bool_or(status = 'final') THEN 'final'
				WHEN bool_or(status = 'draft' AND `+sqlVisitReportIsPersisted+`) THEN 'draft'
				ELSE ''
			END AS report_status
		FROM visits.visit_reports r
		WHERE visit_id = ANY($1::uuid[])
		GROUP BY visit_id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	statusByVisit := map[string]string{}
	for rows.Next() {
		var id, st string
		if err := rows.Scan(&id, &st); err != nil {
			return err
		}
		if st != "" {
			statusByVisit[id] = st
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range visits {
		st, ok := statusByVisit[visits[i].ID]
		if !ok {
			continue
		}
		visits[i].ReportStatus = st
		if st == "final" {
			visits[i].HasFinalReport = true
		}
	}
	return nil
}

// VisitHasFinalReport reports whether the visit has at least one finalized CR.
func (s *Store) VisitHasFinalReport(ctx context.Context, visitID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM visits.visit_reports
			WHERE visit_id = $1::uuid AND status = 'final'
		)`, visitID).Scan(&ok)
	return ok, err
}

// ListFinalVisitReportsForClient returns finalized CRs for a visit, oldest finalized first.
func (s *Store) ListFinalVisitReportsForClient(ctx context.Context, visitID string) ([]ClientFinalReport, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id::text,
			COALESCE(NULLIF(TRIM(u.full_name), ''), '—') AS author_name,
			COALESCE(r.body_text, ''),
			r.finalized_at
		FROM visits.visit_reports r
		JOIN identity.users u ON u.id = r.author_user_id
		WHERE r.visit_id = $1::uuid AND r.status = 'final'
		ORDER BY r.finalized_at ASC NULLS LAST, r.created_at ASC`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClientFinalReport
	for rows.Next() {
		var r ClientFinalReport
		if err := rows.Scan(&r.ID, &r.AuthorName, &r.BodyText, &r.FinalizedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []ClientFinalReport{}
	}
	return out, rows.Err()
}

// ListDraftVisitReportsWithBody returns draft CRs with non-empty body (ready to finalize for client).
func (s *Store) ListDraftVisitReportsWithBody(ctx context.Context, visitID string) ([]VisitReport, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+visitReportReturning+`
		FROM visits.visit_reports
		WHERE visit_id = $1::uuid
		  AND status = 'draft'
		  AND length(trim(COALESCE(body_text, ''))) > 0
		ORDER BY updated_at ASC`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisitReport
	for rows.Next() {
		r, err := scanVisitReport(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []VisitReport{}
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
