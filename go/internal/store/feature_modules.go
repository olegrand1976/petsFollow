package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type FeatureModules struct {
	UserID         string `json:"userId"`
	ModuleCarePlus bool   `json:"moduleCarePlus"`
	ModuleHorse    bool   `json:"moduleHorse"`
	ModuleKennel   bool   `json:"moduleKennel"`
	ModuleFamily   bool   `json:"moduleFamily"`
}

func (s *Store) GetFeatureModules(ctx context.Context, userID string) (FeatureModules, error) {
	var m FeatureModules
	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text,
			COALESCE(module_care_plus,false), COALESCE(module_horse,false),
			COALESCE(module_kennel,false), COALESCE(module_family,false)
		FROM notifications.client_preferences WHERE user_id=$1`, userID,
	).Scan(&m.UserID, &m.ModuleCarePlus, &m.ModuleHorse, &m.ModuleKennel, &m.ModuleFamily)
	if errors.Is(err, pgx.ErrNoRows) {
		return FeatureModules{UserID: userID}, nil
	}
	return m, err
}

func (s *Store) UpdateFeatureModules(ctx context.Context, userID string, m FeatureModules) (FeatureModules, error) {
	var out FeatureModules
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notifications.client_preferences (
			user_id, hr, care, visits, messages, discovery, billing,
			module_care_plus, module_horse, module_kennel, module_family
		) VALUES ($1, true, true, true, true, true, true, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			module_care_plus = EXCLUDED.module_care_plus,
			module_horse = EXCLUDED.module_horse,
			module_kennel = EXCLUDED.module_kennel,
			module_family = EXCLUDED.module_family
		RETURNING user_id::text, module_care_plus, module_horse, module_kennel, module_family`,
		userID, m.ModuleCarePlus, m.ModuleHorse, m.ModuleKennel, m.ModuleFamily,
	).Scan(&out.UserID, &out.ModuleCarePlus, &out.ModuleHorse, &out.ModuleKennel, &out.ModuleFamily)
	return out, err
}
