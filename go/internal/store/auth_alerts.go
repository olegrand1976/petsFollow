package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	AuthAlertSMTPConfirmFail   = "smtp_confirm_fail"
	AuthAlertLoginFailSpike    = "login_fail_spike"
	AuthAlertUnverifiedSpike   = "email_not_verified_spike"
	AuthAlertRegisterFailSpike = "register_fail_spike"
	AuthAlertUnverifiedStuck   = "unverified_stuck"
)

// RecentAuthAlertExists is true if the same kind+fingerprint was recorded within window.
func (s *Store) RecentAuthAlertExists(ctx context.Context, kind, fingerprint string, window time.Duration) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ops.auth_alerts
		WHERE kind = $1 AND fingerprint = $2 AND created_at >= NOW() - ($3 * INTERVAL '1 second')`,
		kind, fingerprint, int64(window/time.Second)).Scan(&n)
	return n > 0, err
}

func (s *Store) InsertAuthAlert(ctx context.Context, kind, fingerprint, detail, ticketID string) error {
	var ticket any
	if ticketID != "" {
		ticket = ticketID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ops.auth_alerts (id, kind, fingerprint, detail, ticket_id)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.NewString(), kind, fingerprint, detail, ticket)
	return err
}

// CountUnverifiedPasswordClientsOlderThan counts non-demo clients still awaiting email confirm.
func (s *Store) CountUnverifiedPasswordClientsOlderThan(ctx context.Context, age time.Duration) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users
		WHERE role = 'client'
		  AND email_verified_at IS NULL
		  AND COALESCE(password_hash,'') <> ''
		  AND email NOT LIKE '%@petsfollow.test'
		  AND created_at <= NOW() - ($1 * INTERVAL '1 second')`,
		int64(age/time.Second)).Scan(&n)
	return n, err
}
