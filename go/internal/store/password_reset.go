package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetResult struct {
	Token   string
	Email   string
	FullName string
	Locale  string
}

// RequestPasswordReset creates a reset token for password-auth users.
// Returns ErrNotFound when the email is unknown or the account has no password (Google-only).
func (s *Store) RequestPasswordReset(ctx context.Context, email string) (PasswordResetResult, error) {
	u, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return PasswordResetResult{}, err
	}
	if u.PasswordHash == "" {
		return PasswordResetResult{}, ErrNotFound
	}

	token := uuid.NewString()
	expires := time.Now().Add(time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PasswordResetResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE identity.password_reset_tokens SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL`, u.ID); err != nil {
		return PasswordResetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.password_reset_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), u.ID, token, expires); err != nil {
		return PasswordResetResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PasswordResetResult{}, err
	}

	locale := u.PreferredLocale
	if locale == "" {
		locale = "fr"
	}
	return PasswordResetResult{
		Token:    token,
		Email:    u.Email,
		FullName: u.FullName,
		Locale:   locale,
	}, nil
}

// ResetPassword applies a new password from a valid unused token.
func (s *Store) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password_too_short")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id::text, expires_at, used_at
		FROM identity.password_reset_tokens WHERE token = $1`, token).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if usedAt != nil {
		return ErrNotFound
	}
	if time.Now().After(expiresAt) {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `UPDATE identity.users SET password_hash = $2 WHERE id = $1`, userID, string(newHash)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.password_reset_tokens SET used_at = NOW() WHERE token = $1`, token); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
