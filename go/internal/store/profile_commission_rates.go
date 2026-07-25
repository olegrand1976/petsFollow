package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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

// AccrueCareProForPetActivation écrit une ligne de ledger `care_pro` par
// professionnel lié au foyer (taux admin par spécialité, base HT du plan).
// Idempotent via l'index partiel (entitlement_id, vet_user_id) WHERE source_kind='care_pro'.
func (s *Store) AccrueCareProForPetActivation(ctx context.Context, petID string) error {
	ent, err := s.GetEntitlementByPetID(ctx, petID)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if ent.Status != "active" && ent.Status != "past_due" && ent.Status != "cancelled" {
		return nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, COALESCE(u.professional_specialty,'')
		FROM practice.client_access ca
		JOIN identity.users u ON u.id = ca.grantee_user_id
		WHERE ca.client_user_id = $1 AND u.role = 'care_pro'`, ent.OwnerUserID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type accrual struct {
		careProID string
		rateBps   int
	}
	var accruals []accrual
	for rows.Next() {
		var careProID, specialty string
		if err := rows.Scan(&careProID, &specialty); err != nil {
			return err
		}
		if specialty == "" {
			continue
		}
		rate, err := s.GetProfileCommissionRate(ctx, "care_pro."+specialty)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if rate.RateBps <= 0 {
			continue
		}
		accruals = append(accruals, accrual{careProID: careProID, rateBps: rate.RateBps})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(accruals) == 0 {
		return nil
	}

	period, err := s.ResolveOpenPeriodYM(ctx, PeriodYM(time.Now()))
	if err != nil {
		return err
	}
	baseHT := HTVACents(ent.AmountCents)
	for _, a := range accruals {
		commission := CommercialCommissionCents(baseHT, a.rateBps)
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO billing.commission_ledger (
				id, vet_user_id, client_user_id, pet_id, entitlement_id,
				base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at, source_kind
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),'care_pro')
			ON CONFLICT (entitlement_id, vet_user_id) WHERE source_kind='care_pro' DO NOTHING`,
			uuid.NewString(), a.careProID, ent.OwnerUserID, petID, ent.ID,
			baseHT, a.rateBps, commission, period); err != nil {
			return err
		}
	}
	return nil
}
