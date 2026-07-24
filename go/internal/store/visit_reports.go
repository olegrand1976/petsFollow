package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type VisitReport struct {
	ID             string     `json:"id"`
	VisitID        string     `json:"visitId"`
	AuthorUserID   string     `json:"authorUserId"`
	Status         string     `json:"status"`
	BodyText       string     `json:"bodyText"`
	AudioURL       string     `json:"audioUrl,omitempty"`
	AudioObjectKey string     `json:"-"`
	TranscriptText string     `json:"transcriptText,omitempty"`
	ImprovedText   string     `json:"improvedText,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	FinalizedAt    *time.Time `json:"finalizedAt,omitempty"`
}

func (s *Store) UpsertVisitReport(ctx context.Context, visitID, authorUserID, bodyText string) (VisitReport, error) {
	id := uuid.NewString()
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		INSERT INTO visits.visit_reports (id, visit_id, author_user_id, status, body_text)
		VALUES ($1, $2, $3, 'draft', $4)
		ON CONFLICT (visit_id, author_user_id) DO UPDATE
			SET body_text = EXCLUDED.body_text, updated_at = NOW()
			WHERE visits.visit_reports.status = 'draft'
		RETURNING id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at`,
		id, visitID, authorUserID, bodyText,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrConflict
	}
	return r, err
}

func (s *Store) GetVisitReport(ctx context.Context, visitID, authorUserID string) (VisitReport, error) {
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at
		FROM visits.visit_reports WHERE visit_id=$1 AND author_user_id=$2`, visitID, authorUserID,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) GetVisitReportByID(ctx context.Context, reportID string) (VisitReport, error) {
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at
		FROM visits.visit_reports WHERE id=$1`, reportID,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) UpdateVisitReportAudio(ctx context.Context, reportID, audioURL, objectKey string) (VisitReport, error) {
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET audio_url=$2, audio_object_key=$3, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at`,
		reportID, audioURL, objectKey,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
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
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET transcript_text=$2, body_text=CASE WHEN COALESCE(body_text,'')='' THEN $2 ELSE body_text END, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at`,
		reportID, transcript,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) UpdateVisitReportImproved(ctx context.Context, reportID, improved string) (VisitReport, error) {
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET improved_text=$2, body_text=$2, updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at`,
		reportID, improved,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
}

func (s *Store) FinalizeVisitReport(ctx context.Context, reportID string) (VisitReport, error) {
	var r VisitReport
	err := s.pool.QueryRow(ctx, `
		UPDATE visits.visit_reports
		SET status='final', finalized_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='draft'
		RETURNING id::text, visit_id::text, author_user_id::text, status, COALESCE(body_text,''),
			COALESCE(audio_url,''), COALESCE(audio_object_key,''), COALESCE(transcript_text,''),
			COALESCE(improved_text,''), created_at, updated_at, finalized_at`,
		reportID,
	).Scan(&r.ID, &r.VisitID, &r.AuthorUserID, &r.Status, &r.BodyText,
		&r.AudioURL, &r.AudioObjectKey, &r.TranscriptText, &r.ImprovedText,
		&r.CreatedAt, &r.UpdatedAt, &r.FinalizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReport{}, ErrNotFound
	}
	return r, err
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
