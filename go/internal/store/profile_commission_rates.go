package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProfileCommissionRate is an editable rate for a professional profile key.
type ProfileCommissionRate struct {
	ProfileKey string    `json:"profileKey"`
	Label      string    `json:"label"`
	RateBps    int       `json:"rateBps"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (s *Store) ListProfileCommissionRates(ctx context.Context) ([]ProfileCommissionRate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT profile_key, label, rate_bps, updated_at
		FROM billing.profile_commission_rates
		ORDER BY profile_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProfileCommissionRate
	for rows.Next() {
		var r ProfileCommissionRate
		if err := rows.Scan(&r.ProfileKey, &r.Label, &r.RateBps, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []ProfileCommissionRate{}
	}
	return out, rows.Err()
}

func (s *Store) GetProfileCommissionRate(ctx context.Context, profileKey string) (ProfileCommissionRate, error) {
	var r ProfileCommissionRate
	err := s.pool.QueryRow(ctx, `
		SELECT profile_key, label, rate_bps, updated_at
		FROM billing.profile_commission_rates WHERE profile_key=$1`, profileKey).
		Scan(&r.ProfileKey, &r.Label, &r.RateBps, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfileCommissionRate{}, ErrNotFound
	}
	return r, err
}

func (s *Store) UpsertProfileCommissionRates(ctx context.Context, rates []ProfileCommissionRate) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, r := range rates {
		if r.ProfileKey == "" {
			continue
		}
		// Only update known seeded keys — never insert arbitrary profile_key via API.
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM billing.profile_commission_rates WHERE profile_key=$1)`,
			r.ProfileKey).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrValidation
		}
		if r.RateBps < 0 {
			r.RateBps = 0
		}
		if r.RateBps > 10000 {
			r.RateBps = 10000
		}
		_, err := tx.Exec(ctx, `
			UPDATE billing.profile_commission_rates
			SET label = COALESCE(NULLIF($2,''), label),
				rate_bps = $3,
				updated_at = NOW()
			WHERE profile_key = $1`,
			r.ProfileKey, r.Label, r.RateBps)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AccrueCareProForPetActivation prepares care_pro commission accrual.
// No-op while all specialty rates are 0%; ready to write ledger when rates > 0.
func (s *Store) AccrueCareProForPetActivation(ctx context.Context, petID string) error {
	var ownerID string
	err := s.pool.QueryRow(ctx, `SELECT owner_user_id::text FROM pets.pets WHERE id=$1`, petID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, COALESCE(u.professional_specialty,'')
		FROM practice.client_access ca
		JOIN identity.users u ON u.id = ca.grantee_user_id
		WHERE ca.client_user_id = $1 AND u.role = 'care_pro'`, ownerID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var careProID, specialty string
		if err := rows.Scan(&careProID, &specialty); err != nil {
			return err
		}
		if specialty == "" {
			continue
		}
		key := "care_pro." + specialty
		rate, err := s.GetProfileCommissionRate(ctx, key)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if rate.RateBps <= 0 {
			// Prepared ground: skip ledger until admin sets a non-zero rate.
			continue
		}
		// Future: insert care_pro commission ledger line (rate.RateBps of HT).
		_ = careProID
	}
	return rows.Err()
}
