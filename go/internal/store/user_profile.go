package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Store) UpdateUserFullName(ctx context.Context, userID, fullName string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.users SET full_name = $2 WHERE id = $1`, userID, fullName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ChangeUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	var hash *string
	err := s.pool.QueryRow(ctx, `SELECT password_hash FROM identity.users WHERE id = $1`, userID).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if hash == nil || *hash == "" {
		return ErrForbidden
	}
	if bcrypt.CompareHashAndPassword([]byte(*hash), []byte(currentPassword)) != nil {
		return ErrForbidden
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `UPDATE identity.users SET password_hash = $2 WHERE id = $1`, userID, string(newHash))
	return err
}

func (s *Store) DeleteClientAccount(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pets.pets WHERE owner_user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM messaging.threads WHERE client_user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM practice.practice_clients WHERE client_user_id = $1`, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM identity.users WHERE id = $1 AND role = 'client'`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateEmailPrefs(ctx context.Context, vetID string, onMessage, onHeartRate bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, $2, $3)
		ON CONFLICT (vet_user_id) DO UPDATE SET
			email_on_message = EXCLUDED.email_on_message,
			email_on_heartrate = EXCLUDED.email_on_heartrate`,
		vetID, onMessage, onHeartRate)
	return err
}

func (s *Store) GetEmailPrefs(ctx context.Context, vetID string) (onMessage, onHeartRate bool, err error) {
	return s.EmailPrefs(ctx, vetID)
}
