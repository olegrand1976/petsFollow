package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EmailConfirmationResult struct {
	Token    string
	Email    string
	FullName string
	Locale   string
}

// RequestEmailConfirmation issues a fresh 48h verification token for an unverified
// password-auth user. Returns ErrNotFound when the email is unknown, already verified,
// or has no password (Google-only).
func (s *Store) RequestEmailConfirmation(ctx context.Context, email string) (EmailConfirmationResult, error) {
	u, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return EmailConfirmationResult{}, err
	}
	if u.EmailVerifiedAt != nil || u.PasswordHash == "" {
		return EmailConfirmationResult{}, ErrNotFound
	}

	token := uuid.NewString()
	expires := time.Now().Add(48 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return EmailConfirmationResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE identity.email_verification_tokens SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL`, u.ID); err != nil {
		return EmailConfirmationResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), u.ID, token, expires); err != nil {
		return EmailConfirmationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return EmailConfirmationResult{}, err
	}

	locale := u.PreferredLocale
	if locale == "" {
		locale = "fr"
	}
	return EmailConfirmationResult{
		Token:    token,
		Email:    u.Email,
		FullName: u.FullName,
		Locale:   locale,
	}, nil
}
