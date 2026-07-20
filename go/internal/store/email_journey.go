package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const JourneyStatusActive = "active"
const JourneyStatusPaused = "paused"
const JourneyStatusCompleted = "completed"

type EmailJourney struct {
	UserID     string
	AnchorAt   time.Time
	EnrolledAt time.Time
	Status     string
}

type EmailSend struct {
	UserID  string
	StepKey string
	SentAt  time.Time
	Status  string
	Meta    map[string]any
}

// EnsureDiscoveryStarted persists discovery.progress if missing (GET previously returned a virtual row).
func (s *Store) EnsureDiscoveryStarted(ctx context.Context, userID string) (DiscoveryProgress, error) {
	var p DiscoveryProgress
	var cardsJSON []byte
	err := s.pool.QueryRow(ctx, `
		INSERT INTO discovery.progress (user_id, started_at, completed_cards, streak_days, updated_at)
		VALUES ($1, NOW(), '[]'::jsonb, 0, NOW())
		ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING user_id::text, started_at, completed_cards, streak_days, updated_at`,
		userID,
	).Scan(&p.UserID, &p.StartedAt, &cardsJSON, &p.StreakDays, &p.UpdatedAt)
	if err != nil {
		return DiscoveryProgress{}, err
	}
	if err := json.Unmarshal(cardsJSON, &p.CompletedCards); err != nil {
		p.CompletedCards = []string{}
	}
	return p, nil
}

// EnrollEmailJourney starts the loyalty drip for a client. Idempotent.
func (s *Store) EnrollEmailJourney(ctx context.Context, userID string, anchorAt time.Time) error {
	if anchorAt.IsZero() {
		anchorAt = time.Now().UTC()
	}
	if _, err := s.EnsureDiscoveryStarted(ctx, userID); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO discovery.email_journey (user_id, anchor_at, enrolled_at, status)
		VALUES ($1, $2, NOW(), 'active')
		ON CONFLICT (user_id) DO NOTHING`, userID, anchorAt.UTC())
	return err
}

// BackfillEmailJourneys enrolls all clients not yet on the journey (anchor = discovery.started_at or users.created_at).
func (s *Store) BackfillEmailJourneys(ctx context.Context) (int, error) {
	ct, err := s.pool.Exec(ctx, `
		INSERT INTO discovery.progress (user_id, started_at, completed_cards, streak_days, updated_at)
		SELECT u.id, u.created_at, '[]'::jsonb, 0, NOW()
		FROM identity.users u
		WHERE u.role = 'client'
		ON CONFLICT (user_id) DO NOTHING`)
	if err != nil {
		return 0, err
	}
	_ = ct
	// Anchor = NOW() for backfill to avoid flooding legacy accounts with every overdue step.
	ct, err = s.pool.Exec(ctx, `
		INSERT INTO discovery.email_journey (user_id, anchor_at, enrolled_at, status)
		SELECT u.id, NOW(), NOW(), 'active'
		FROM identity.users u
		WHERE u.role = 'client'
		ON CONFLICT (user_id) DO NOTHING`)
	if err != nil {
		return 0, err
	}
	return int(ct.RowsAffected()), nil
}

func (s *Store) GetEmailJourney(ctx context.Context, userID string) (EmailJourney, error) {
	var j EmailJourney
	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text, anchor_at, enrolled_at, status
		FROM discovery.email_journey WHERE user_id = $1`, userID,
	).Scan(&j.UserID, &j.AnchorAt, &j.EnrolledAt, &j.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailJourney{}, ErrNotFound
	}
	return j, err
}

func (s *Store) ListActiveEmailJourneys(ctx context.Context, limit int) ([]EmailJourney, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, `
		SELECT user_id::text, anchor_at, enrolled_at, status
		FROM discovery.email_journey
		WHERE status = 'active'
		ORDER BY enrolled_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmailJourney, 0)
	for rows.Next() {
		var j EmailJourney
		if err := rows.Scan(&j.UserID, &j.AnchorAt, &j.EnrolledAt, &j.Status); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *Store) PauseEmailJourney(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE discovery.email_journey SET status = 'paused' WHERE user_id = $1`, userID)
	return err
}

func (s *Store) HasEmailSend(ctx context.Context, userID, stepKey string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM discovery.email_sends WHERE user_id = $1 AND step_key = $2
		)`, userID, stepKey).Scan(&exists)
	return exists, err
}

// LastEmailSendAt returns the last send time for a step, or zero if never sent.
func (s *Store) LastEmailSendAt(ctx context.Context, userID, stepKey string) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT sent_at FROM discovery.email_sends
		WHERE user_id = $1 AND step_key = $2`, userID, stepKey).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}
	return t, err
}

// RecordEmailSend inserts or refreshes a send row (used for cooldown events).
func (s *Store) RecordEmailSend(ctx context.Context, userID, stepKey, status string, meta map[string]any) error {
	if status == "" {
		status = "sent"
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte("{}")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO discovery.email_sends (user_id, step_key, sent_at, status, meta)
		VALUES ($1, $2, NOW(), $3, $4::jsonb)
		ON CONFLICT (user_id, step_key) DO UPDATE SET
			sent_at = EXCLUDED.sent_at,
			status = EXCLUDED.status,
			meta = EXCLUDED.meta`,
		userID, stepKey, status, string(raw))
	return err
}

// TryAdvisoryLock acquires a session-level advisory lock. Caller must UnlockAdvisoryLock.
func (s *Store) TryAdvisoryLock(ctx context.Context, key int64) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, key).Scan(&ok)
	return ok, err
}

func (s *Store) UnlockAdvisoryLock(ctx context.Context, key int64) error {
	_, err := s.pool.Exec(ctx, `SELECT pg_advisory_unlock($1)`, key)
	return err
}

// JourneyClientSegment aggregates data needed for drip eligibility.
type JourneyClientSegment struct {
	UserID              string
	Email               string
	FullName            string
	Locale              string
	PetCount            int
	HorseCount          int
	ValidatedHRCount    int
	DaysSinceLastHR     *int
	HasPendingPayment   bool
	PendingPaymentDays  int
	HasPastDue          bool
	HasAnnualPlan       bool
	AnnualValidUntil    *time.Time
	ActiveAddons        map[string]bool
	DiscoveryPref       bool
	BillingPref         bool
	JourneyDays         int
}

func (s *Store) LoadJourneyClientSegment(ctx context.Context, userID string, anchorAt time.Time, now time.Time) (JourneyClientSegment, error) {
	seg := JourneyClientSegment{
		UserID:       userID,
		ActiveAddons: map[string]bool{},
	}
	err := s.pool.QueryRow(ctx, `
		SELECT email, COALESCE(full_name,''), COALESCE(preferred_locale,'fr')
		FROM identity.users WHERE id = $1 AND role = 'client'`, userID,
	).Scan(&seg.Email, &seg.FullName, &seg.Locale)
	if errors.Is(err, pgx.ErrNoRows) {
		return JourneyClientSegment{}, ErrNotFound
	}
	if err != nil {
		return JourneyClientSegment{}, err
	}

	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (WHERE species = 'horse')::int,
			COALESCE(BOOL_OR(payment_status = 'pending_payment'), false),
			COALESCE(MAX(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400)
				FILTER (WHERE payment_status = 'pending_payment'), 0)::int
		FROM pets.pets WHERE owner_user_id = $1`, userID,
	).Scan(&seg.PetCount, &seg.HorseCount, &seg.HasPendingPayment, &seg.PendingPaymentDays); err != nil {
		return JourneyClientSegment{}, err
	}

	var lastHR *time.Time
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int, MAX(validated_at)
		FROM heartrate.sessions
		WHERE owner_user_id = $1 AND status = 'validated'`, userID,
	).Scan(&seg.ValidatedHRCount, &lastHR); err != nil {
		return JourneyClientSegment{}, err
	}
	if lastHR != nil {
		d := int(now.Sub(*lastHR).Hours() / 24)
		seg.DaysSinceLastHR = &d
	}

	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.pet_entitlements
			WHERE owner_user_id = $1 AND status = 'past_due'
		) OR EXISTS(
			SELECT 1 FROM billing.addon_entitlements
			WHERE owner_user_id = $1 AND status = 'past_due'
		)`, userID).Scan(&seg.HasPastDue); err != nil {
		return JourneyClientSegment{}, err
	}

	var annualUntil *time.Time
	err = s.pool.QueryRow(ctx, `
		SELECT valid_until FROM billing.pet_entitlements
		WHERE owner_user_id = $1 AND plan_code = 'annual' AND status IN ('active','past_due')
		ORDER BY valid_until NULLS LAST
		LIMIT 1`, userID).Scan(&annualUntil)
	if err == nil {
		seg.HasAnnualPlan = true
		seg.AnnualValidUntil = annualUntil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return JourneyClientSegment{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT addon_code FROM billing.addon_entitlements
		WHERE owner_user_id = $1
			AND status IN ('active','past_due')
			AND (valid_until IS NULL OR valid_until > NOW())`, userID)
	if err != nil {
		return JourneyClientSegment{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return JourneyClientSegment{}, err
		}
		seg.ActiveAddons[code] = true
	}
	if err := rows.Err(); err != nil {
		return JourneyClientSegment{}, err
	}

	prefs, err := s.GetClientNotificationPrefs(ctx, userID)
	if err != nil {
		return JourneyClientSegment{}, err
	}
	seg.DiscoveryPref = prefs.Discovery
	seg.BillingPref = prefs.Billing

	seg.JourneyDays = int(now.Sub(anchorAt).Hours() / 24)
	if seg.JourneyDays < 0 {
		seg.JourneyDays = 0
	}
	return seg, nil
}

// SetClientDiscoveryPrefOnly updates only the discovery preference (unsubscribe journey).
func (s *Store) SetClientDiscoveryPrefOnly(ctx context.Context, userID string, discovery bool) error {
	prefs, err := s.GetClientNotificationPrefs(ctx, userID)
	if err != nil {
		return err
	}
	prefs.Discovery = discovery
	_, err = s.UpdateClientNotificationPrefs(ctx, userID, prefs)
	return err
}
