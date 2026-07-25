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
func (s *Store) ListDigestRecipients(ctx context.Context) ([]DigestRecipient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, email, COALESCE(full_name,''), COALESCE(preferred_locale,'fr'), role
		FROM identity.users
		WHERE role IN ('admin', 'commercial', 'commercial_manager')
		ORDER BY role, email`)
	if err != nil {
		return nil, err
	}
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
