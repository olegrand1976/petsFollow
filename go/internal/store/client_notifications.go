package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ClientNotificationPrefs struct {
	UserID    string `json:"userId"`
	HR        bool   `json:"hr"`
	Care      bool   `json:"care"`
	Visits    bool   `json:"visits"`
	Messages  bool   `json:"messages"`
	Discovery bool   `json:"discovery"`
	Billing   bool   `json:"billing"`
}

type DeviceToken struct {
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Store) UpsertDeviceToken(ctx context.Context, userID, token, platform string) (DeviceToken, error) {
	var dt DeviceToken
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notifications.device_tokens (user_id, token, platform, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, token) DO UPDATE SET platform = EXCLUDED.platform, updated_at = NOW()
		RETURNING user_id::text, token, platform, updated_at`,
		userID, token, platform,
	).Scan(&dt.UserID, &dt.Token, &dt.Platform, &dt.UpdatedAt)
	return dt, err
}

func (s *Store) GetClientNotificationPrefs(ctx context.Context, userID string) (ClientNotificationPrefs, error) {
	var p ClientNotificationPrefs
	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text, hr, care, visits, messages, discovery, billing
		FROM notifications.client_preferences WHERE user_id = $1`, userID,
	).Scan(&p.UserID, &p.HR, &p.Care, &p.Visits, &p.Messages, &p.Discovery, &p.Billing)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientNotificationPrefs{
			UserID: userID, HR: true, Care: true, Visits: true, Messages: true, Discovery: true, Billing: true,
		}, nil
	}
	return p, err
}

func (s *Store) UpdateClientNotificationPrefs(ctx context.Context, userID string, p ClientNotificationPrefs) (ClientNotificationPrefs, error) {
	var out ClientNotificationPrefs
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notifications.client_preferences (user_id, hr, care, visits, messages, discovery, billing)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			hr = EXCLUDED.hr, care = EXCLUDED.care, visits = EXCLUDED.visits,
			messages = EXCLUDED.messages, discovery = EXCLUDED.discovery, billing = EXCLUDED.billing
		RETURNING user_id::text, hr, care, visits, messages, discovery, billing`,
		userID, p.HR, p.Care, p.Visits, p.Messages, p.Discovery, p.Billing,
	).Scan(&out.UserID, &out.HR, &out.Care, &out.Visits, &out.Messages, &out.Discovery, &out.Billing)
	return out, err
}
