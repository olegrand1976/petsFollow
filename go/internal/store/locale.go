package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

func (s *Store) UpdateUserLocale(ctx context.Context, userID, locale string) error {
	locale = i18n.NormalizeLocale(locale)
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET preferred_locale = $2 WHERE id = $1`, userID, locale)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetUserPreferredLocale(ctx context.Context, userID string) (string, error) {
	var locale string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(preferred_locale,'fr') FROM identity.users WHERE id = $1`, userID).Scan(&locale)
	if errors.Is(err, pgx.ErrNoRows) {
		return "fr", ErrNotFound
	}
	return i18n.NormalizeLocale(locale), err
}
