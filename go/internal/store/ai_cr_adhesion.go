package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Adhesion step keys (practice.ai_cr_email_sends.step_key).
const (
	AiCrStepJ0Activation = "j0_activation"
	AiCrStepJ3Nudge      = "j3_nudge"
	AiCrStepJ7Nudge      = "j7_nudge"
	AiCrStepJ14NPS       = "j14_nps"
	AiCrStepJ15Digest    = "j15_digest"
	AiCrStepJ30Digest    = "j30_digest"
	AiCrStepJ45NPS       = "j45_nps"
	AiCrStepJ60ROI       = "j60_roi"
	AiCrStepJ75Convert   = "j75_convert"
	AiCrStepJ85Urgency   = "j85_urgency"
	AiCrStepJ90Last      = "j90_last"
)

type AiCrAdhesionCandidate struct {
	PracticeID   string
	PracticeName string
	Status       string
	ActivatedAt  time.Time
	TrialEndsAt  time.Time
}

// ListAiCrAdhesionCandidates returns modules eligible for adoption drip emails.
// Includes active (IA included in Pro) and legacy trial rows.
func (s *Store) ListAiCrAdhesionCandidates(ctx context.Context) ([]AiCrAdhesionCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.practice_id::text, COALESCE(p.name,''), m.status, m.activated_at, m.trial_ends_at
		FROM practice.ai_cr_modules m
		JOIN practice.practices p ON p.id = m.practice_id
		WHERE m.status IN ('trial', 'active')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AiCrAdhesionCandidate, 0)
	for rows.Next() {
		var c AiCrAdhesionCandidate
		if err := rows.Scan(&c.PracticeID, &c.PracticeName, &c.Status, &c.ActivatedAt, &c.TrialEndsAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// TryClaimAiCrEmailSend inserts a send row; returns true if this caller claimed the step.
func (s *Store) TryClaimAiCrEmailSend(ctx context.Context, practiceID, stepKey, status string, meta map[string]any) (bool, error) {
	if status == "" {
		status = "sent"
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte("{}")
	}
	var claimed string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO practice.ai_cr_email_sends (practice_id, step_key, status, sent_at, meta)
		VALUES ($1::uuid, $2, $3, NOW(), $4::jsonb)
		ON CONFLICT (practice_id, step_key) DO NOTHING
		RETURNING practice_id::text`, practiceID, stepKey, status, string(raw)).Scan(&claimed)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return claimed != "", nil
}

// HasAiCrEmailSend reports whether a step was already claimed (sent or skipped).
func (s *Store) HasAiCrEmailSend(ctx context.Context, practiceID, stepKey string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.ai_cr_email_sends
			WHERE practice_id = $1::uuid AND step_key = $2
		)`, practiceID, stepKey).Scan(&exists)
	return exists, err
}

// DeleteAiCrEmailSend releases a claim so a failed send can be retried on the next run.
func (s *Store) DeleteAiCrEmailSend(ctx context.Context, practiceID, stepKey string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM practice.ai_cr_email_sends
		WHERE practice_id = $1::uuid AND step_key = $2`, practiceID, stepKey)
	return err
}

// AiCrAdhesionDaysSince returns whole days since activation (UTC floor).
func AiCrAdhesionDaysSince(activatedAt, now time.Time) int {
	if now.Before(activatedAt) {
		return 0
	}
	return int(now.Sub(activatedAt).Hours() / 24)
}
