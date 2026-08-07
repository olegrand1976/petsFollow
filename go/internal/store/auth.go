package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type RegisterGoogleVetInput struct {
	Email            string
	FullName         string
	GoogleSub        string
	PracticeName     string
	PreferredLocale  string
	AutoReplyDefault string
	TermsAccepted    bool
}

type RegisterGoogleClientInput struct {
	Email           string
	FullName        string
	GoogleSub       string
	PreferredLocale string
	TermsAccepted   bool
}

// BumpTokenVersion revokes every JWT already issued for the user: refresh
// compares the token's version to this column.
func (s *Store) BumpTokenVersion(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE identity.users SET token_version = token_version + 1 WHERE id = $1`, userID)
	return err
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
	if _, err := insertPrimarySiteTx(ctx, tx, practiceID); err != nil {
		return User{}, err
	}
	var termsAt *time.Time
	if in.TermsAccepted {
		termsAt = &now
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, google_sub, auth_provider, preferred_locale, terms_accepted_at)
		VALUES ($1, $2, NULL, $3, 'vet', $4, $5, $6, 'google', $7, $8)`,
		userID, in.Email, in.FullName, practiceID, now, in.GoogleSub, i18n.NormalizeLocale(in.PreferredLocale), termsAt); err != nil {
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
	_ = s.EnsureUserProfiles(ctx, userID)
	_ = s.EnsureReferenceTeamMembership(ctx, practiceID, userID)
	return s.GetUserByID(ctx, userID)
}

// RegisterGoogleClient creates a verified client from Google Sign-In (create-if-absent).
func (s *Store) RegisterGoogleClient(ctx context.Context, in RegisterGoogleClientInput) (User, error) {
	userID := uuid.NewString()
	now := time.Now()
	var termsAt *time.Time
	if in.TermsAccepted {
		termsAt = &now
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id,
			email_verified_at, google_sub, auth_provider, preferred_locale, terms_accepted_at
		) VALUES ($1, $2, NULL, $3, 'client', NULL, $4, $5, 'google', $6, $7)`,
		userID, in.Email, in.FullName, now, in.GoogleSub, i18n.NormalizeLocale(in.PreferredLocale), termsAt)
	if err != nil {
		if isUniqueViolation(err) {
			// Concurrent sign-in won the race — return existing client and link google_sub if needed.
			if u, e := s.GetUserByGoogleSub(ctx, in.GoogleSub); e == nil {
				return u, nil
			}
			u, e := s.GetUserByEmail(ctx, in.Email)
			if e != nil {
				return User{}, err
			}
			if u.Role != kernel.RoleClient {
				return User{}, ErrConflict
			}
			if linkErr := s.LinkGoogleAccount(ctx, u.ID, in.GoogleSub); linkErr != nil {
				// Already linked to this sub → OK; other failures → conflict.
				if u2, e2 := s.GetUserByGoogleSub(ctx, in.GoogleSub); e2 == nil && u2.ID == u.ID {
					return u2, nil
				}
				return User{}, ErrConflict
			}
			return s.GetUserByID(ctx, u.ID)
		}
		return User{}, err
	}
	_ = s.EnrollEmailJourney(ctx, userID, time.Now().UTC())
	_ = s.EnsureUserProfiles(ctx, userID)
	return s.GetUserByID(ctx, userID)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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
