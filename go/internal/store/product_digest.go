package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProductDigest is one calendar day's functional changelog (Europe/Brussels date).
type ProductDigest struct {
	DigestDate       time.Time
	Headline         string
	BodyText         string
	HeadlineByLocale map[string]string
	BodyByLocale     map[string]string
	CommitsJSON      json.RawMessage
	Status           string
	GeneratedAt      *time.Time
	SentAt           *time.Time
	Meta             json.RawMessage
}

// DigestRecipient is an internal staff user targeted by the product digest.
type DigestRecipient struct {
	ID              string
	Email           string
	FullName        string
	PreferredLocale string
	Role            string
}

// UpsertProductDigest inserts or replaces the digest for a given date (unless already sent).
func (s *Store) UpsertProductDigest(ctx context.Context, d ProductDigest) error {
	headlineLocale, err := json.Marshal(d.HeadlineByLocale)
	if err != nil {
		return err
	}
	bodyLocale, err := json.Marshal(d.BodyByLocale)
	if err != nil {
		return err
	}
	if d.CommitsJSON == nil {
		d.CommitsJSON = []byte("[]")
	}
	if d.Meta == nil {
		d.Meta = []byte("{}")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO ops.product_digests (
			digest_date, headline, body_text, headline_by_locale, body_by_locale,
			commits_json, status, generated_at, meta
		) VALUES ($1::date, $2, $3, $4::jsonb, $5::jsonb, $6::jsonb, $7, COALESCE($8, NOW()), $9::jsonb)
		ON CONFLICT (digest_date) DO UPDATE SET
			headline = EXCLUDED.headline,
			body_text = EXCLUDED.body_text,
			headline_by_locale = EXCLUDED.headline_by_locale,
			body_by_locale = EXCLUDED.body_by_locale,
			commits_json = EXCLUDED.commits_json,
			status = EXCLUDED.status,
			generated_at = EXCLUDED.generated_at,
			meta = EXCLUDED.meta
		WHERE ops.product_digests.status IS DISTINCT FROM 'sent'`,
		d.DigestDate, d.Headline, d.BodyText, headlineLocale, bodyLocale,
		d.CommitsJSON, d.Status, d.GeneratedAt, d.Meta)
	return err
}

// GetProductDigest returns the digest for digestDate (date-only).
func (s *Store) GetProductDigest(ctx context.Context, digestDate time.Time) (*ProductDigest, error) {
	var d ProductDigest
	var headlineLocale, bodyLocale []byte
	err := s.pool.QueryRow(ctx, `
		SELECT digest_date, headline, body_text, headline_by_locale, body_by_locale, commits_json, status,
			generated_at, sent_at, meta
		FROM ops.product_digests WHERE digest_date = $1::date`, digestDate).Scan(
		&d.DigestDate, &d.Headline, &d.BodyText, &headlineLocale, &bodyLocale, &d.CommitsJSON, &d.Status,
		&d.GeneratedAt, &d.SentAt, &d.Meta,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.HeadlineByLocale = map[string]string{}
	d.BodyByLocale = map[string]string{}
	if len(headlineLocale) > 0 {
		_ = json.Unmarshal(headlineLocale, &d.HeadlineByLocale)
	}
	if len(bodyLocale) > 0 {
		_ = json.Unmarshal(bodyLocale, &d.BodyByLocale)
	}
	return &d, nil
}

// MarkProductDigestSent sets status=sent and sent_at.
func (s *Store) MarkProductDigestSent(ctx context.Context, digestDate time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE ops.product_digests
		SET status = 'sent', sent_at = NOW()
		WHERE digest_date = $1::date`, digestDate)
	return err
}

// ListDigestRecipients returns staff for product digest emails.
// Skips *.petsfollow.test demo emails (same rule as weekly).
func (s *Store) ListDigestRecipients(ctx context.Context) ([]DigestRecipient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, email, COALESCE(full_name,''), COALESCE(preferred_locale,'fr'), role
		FROM identity.users
		WHERE role IN ('admin', 'commercial', 'commercial_manager')
		  AND email IS NOT NULL AND TRIM(email) <> ''
		  AND LOWER(email) NOT LIKE '%@petsfollow.test'
		ORDER BY role, email`)
	if err != nil {
		return nil, err
	}
	return scanDigestRecipients(rows)
}

// RecordProductDigestSend inserts an idempotent send row. Returns true if newly inserted.
func (s *Store) RecordProductDigestSend(ctx context.Context, digestDate time.Time, userID string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO ops.product_digest_sends (digest_date, user_id)
		VALUES ($1::date, $2::uuid)
		ON CONFLICT DO NOTHING`, digestDate, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ClearProductDigestSend removes a send row so a later run can retry after SMTP failure.
func (s *Store) ClearProductDigestSend(ctx context.Context, digestDate time.Time, userID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM ops.product_digest_sends
		WHERE digest_date = $1::date AND user_id = $2::uuid`, digestDate, userID)
	return err
}

func scanProductDigestRows(rows pgx.Rows) ([]ProductDigest, error) {
	defer rows.Close()
	var out []ProductDigest
	for rows.Next() {
		var d ProductDigest
		var headlineLocale, bodyLocale []byte
		if err := rows.Scan(
			&d.DigestDate, &d.Headline, &d.BodyText, &headlineLocale, &bodyLocale, &d.CommitsJSON, &d.Status,
			&d.GeneratedAt, &d.SentAt, &d.Meta,
		); err != nil {
			return nil, err
		}
		d.HeadlineByLocale = map[string]string{}
		d.BodyByLocale = map[string]string{}
		if len(headlineLocale) > 0 {
			_ = json.Unmarshal(headlineLocale, &d.HeadlineByLocale)
		}
		if len(bodyLocale) > 0 {
			_ = json.Unmarshal(bodyLocale, &d.BodyByLocale)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListProductDigests returns ready/sent digests newest-first (optional limit; 0 = no limit).
func (s *Store) ListProductDigests(ctx context.Context, limit int) ([]ProductDigest, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 90 {
		limit = 90
	}
	rows, err := s.pool.Query(ctx, `
		SELECT digest_date, headline, body_text, headline_by_locale, body_by_locale, commits_json, status,
			generated_at, sent_at, meta
		FROM ops.product_digests
		WHERE status IN ('ready', 'sent')
		ORDER BY digest_date DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return scanProductDigestRows(rows)
}

// ListProductDigestsSince returns ready/sent digests with digest_date >= since (ascending).
func (s *Store) ListProductDigestsSince(ctx context.Context, since time.Time) ([]ProductDigest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT digest_date, headline, body_text, headline_by_locale, body_by_locale, commits_json, status,
			generated_at, sent_at, meta
		FROM ops.product_digests
		WHERE status IN ('ready', 'sent') AND digest_date >= $1::date
		ORDER BY digest_date ASC`, since)
	if err != nil {
		return nil, err
	}
	return scanProductDigestRows(rows)
}

func scanDigestRecipients(rows pgx.Rows) ([]DigestRecipient, error) {
	defer rows.Close()
	var out []DigestRecipient
	for rows.Next() {
		var r DigestRecipient
		if err := rows.Scan(&r.ID, &r.Email, &r.FullName, &r.PreferredLocale, &r.Role); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListWeeklyDigestRecipients returns daily staff plus active practice reference vets.
// Skips *.petsfollow.test demo emails.
func (s *Store) ListWeeklyDigestRecipients(ctx context.Context) ([]DigestRecipient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, email, COALESCE(full_name,''), COALESCE(preferred_locale,'fr'), role
		FROM (
			SELECT u.id, u.email, u.full_name, u.preferred_locale, u.role
			FROM identity.users u
			WHERE u.role IN ('admin', 'commercial', 'commercial_manager')
			  AND u.email IS NOT NULL AND TRIM(u.email) <> ''
			  AND LOWER(u.email) NOT LIKE '%@petsfollow.test'

			UNION

			SELECT u.id, u.email, u.full_name, u.preferred_locale, u.role
			FROM practice.practices p
			JOIN identity.users u ON u.id = p.reference_vet_user_id
			WHERE p.reference_vet_user_id IS NOT NULL
			  AND u.email IS NOT NULL AND TRIM(u.email) <> ''
			  AND LOWER(u.email) NOT LIKE '%@petsfollow.test'
			  AND EXISTS (
				SELECT 1 FROM practice.team_members tm
				WHERE tm.practice_id = p.id
				  AND tm.user_id = u.id
				  AND tm.status = 'active'
			  )
		) r
		ORDER BY role, email`)
	if err != nil {
		return nil, err
	}
	return scanDigestRecipients(rows)
}

// RecordProductDigestWeeklySend inserts an idempotent weekly send row. Returns true if newly inserted.
func (s *Store) RecordProductDigestWeeklySend(ctx context.Context, weekStart time.Time, userID string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO ops.product_digest_weekly_sends (week_start, user_id)
		VALUES ($1::date, $2::uuid)
		ON CONFLICT DO NOTHING`, weekStart, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ClearProductDigestWeeklySend removes a weekly send row so a later run can retry after SMTP failure.
func (s *Store) ClearProductDigestWeeklySend(ctx context.Context, weekStart time.Time, userID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM ops.product_digest_weekly_sends
		WHERE week_start = $1::date AND user_id = $2::uuid`, weekStart, userID)
	return err
}
