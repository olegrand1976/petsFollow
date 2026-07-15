package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RegisterGoogleVetInput struct {
	Email            string
	FullName         string
	GoogleSub        string
	PracticeName     string
	PreferredLocale  string
	AutoReplyDefault string
}

func (s *Store) GetUserByGoogleSub(ctx context.Context, googleSub string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+` FROM identity.users WHERE google_sub=$1`, googleSub))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (s *Store) LinkGoogleAccount(ctx context.Context, userID, googleSub string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users
		SET google_sub = $2, auth_provider = CASE WHEN password_hash IS NULL THEN 'google' ELSE auth_provider END
		WHERE id = $1 AND (google_sub IS NULL OR google_sub = $2)`, userID, googleSub)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RegisterGoogleVet(ctx context.Context, in RegisterGoogleVetInput) (User, error) {
	practiceID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO practice.practices (id, name, contact_email) VALUES ($1, $2, $3)`,
		practiceID, in.PracticeName, in.Email); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, google_sub, auth_provider, preferred_locale)
		VALUES ($1, $2, NULL, $3, 'vet', $4, $5, $6, 'google', $7)`,
		userID, in.Email, in.FullName, practiceID, now, in.GoogleSub, in.PreferredLocale); err != nil {
		return User{}, err
	}
	autoReply := in.AutoReplyDefault
	if autoReply == "" {
		autoReply = "Je suis indisponible, je reviens vers vous rapidement."
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, 'available', $3)`,
		userID, practiceID, autoReply); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, true, true)`, userID); err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Store) SetTOTPSecret(ctx context.Context, userID, secret string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET totp_secret = $2, totp_enabled = FALSE WHERE id = $1`, userID, secret)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) EnableTOTP(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET totp_enabled = TRUE WHERE id = $1 AND totp_secret <> ''`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DisableTOTP(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET totp_enabled = FALSE, totp_secret = NULL WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetTOTPSecret(ctx context.Context, userID string) (string, bool, error) {
	var secret string
	var enabled bool
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(totp_secret,''), totp_enabled FROM identity.users WHERE id = $1`, userID).
		Scan(&secret, &enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, ErrNotFound
	}
	return secret, enabled, err
}
