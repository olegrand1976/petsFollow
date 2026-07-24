package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	BonusCodeCommercialRamp = "commercial_ramp"
	BonusCodeCommercialMix  = "commercial_mix"

	BonusStatusEarned = "earned"
	BonusStatusPaid   = "paid"

	commercialRampTargetPets  = 5
	commercialMixTargetPct    = 55
	commercialRampAmountCents = 2500
	commercialMixAmountCents  = 5000
)

var (
	ErrBonusNotEarned   = errors.New("bonus_not_earned")
	ErrBonusAlreadyPaid = errors.New("bonus_already_paid")
)

// CommercialBonusAward is a persisted SPIFF row (earned or paid).
type CommercialBonusAward struct {
	ID                 string     `json:"id"`
	CommercialUserID   string     `json:"commercialUserId"`
	CommercialFullName string     `json:"commercialFullName,omitempty"`
	CommercialEmail    string     `json:"commercialEmail,omitempty"`
	BonusCode          string     `json:"bonusCode"`
	AmountCents        int        `json:"amountCents"`
	Status             string     `json:"status"`
	PeriodYM           string     `json:"periodYm,omitempty"`
	VetUserID          string     `json:"vetUserId,omitempty"`
	VetEmail           string     `json:"vetEmail,omitempty"`
	VetFullName        string     `json:"vetFullName,omitempty"`
	Progress           int        `json:"progress"`
	Target             int        `json:"target"`
	EarnedAt           time.Time  `json:"earnedAt"`
	PaidAt             *time.Time `json:"paidAt,omitempty"`
}

// CommercialBonusTrackRow is an admin suivi row (live progress and/or award).
type CommercialBonusTrackRow struct {
	AwardID            string `json:"awardId,omitempty"`
	CommercialUserID   string `json:"commercialUserId"`
	CommercialFullName string `json:"commercialFullName"`
	CommercialEmail    string `json:"commercialEmail"`
	BonusCode          string `json:"bonusCode"`
	AmountCents        int    `json:"amountCents"`
	Status             string `json:"status"` // available | in_progress | earned | paid
	Progress           int    `json:"progress"`
	Target             int    `json:"target"`
	PeriodYM           string `json:"periodYm,omitempty"`
	VetUserID          string `json:"vetUserId,omitempty"`
	VetEmail           string `json:"vetEmail,omitempty"`
	VetFullName        string `json:"vetFullName,omitempty"`
}

func rampDedupeKey(commercialUserID, vetUserID string) string {
	return fmt.Sprintf("ramp:%s:%s", commercialUserID, vetUserID)
}

func mixDedupeKey(commercialUserID, periodYM string) string {
	return fmt.Sprintf("mix:%s:%s", commercialUserID, periodYM)
}

// SyncCommercialBonusAwards persists earned Ramp/Mix awards when thresholds are met.
// Existing paid/earned awards are never deleted when the rolling window slides.
func (s *Store) SyncCommercialBonusAwards(ctx context.Context, commercialUserID string) error {
	month := PeriodYM(time.Now())

	rows, err := s.pool.Query(ctx, `
		SELECT cl.vet_user_id::text, COUNT(*)::int
		FROM billing.commercial_commission_ledger cl
		WHERE cl.commercial_user_id=$1
		  AND cl.source_type='subscription_pct'
		  AND cl.accrued_at >= NOW() - INTERVAL '60 days'
		GROUP BY cl.vet_user_id
		HAVING COUNT(*) >= $2`, commercialUserID, commercialRampTargetPets)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var vetID string
		var cnt int
		if err := rows.Scan(&vetID, &cnt); err != nil {
			return err
		}
		if err := s.upsertBonusAward(ctx, CommercialBonusAward{
			CommercialUserID: commercialUserID,
			BonusCode:        BonusCodeCommercialRamp,
			AmountCents:      commercialRampAmountCents,
			Status:           BonusStatusEarned,
			VetUserID:        vetID,
			Progress:         cnt,
			Target:           commercialRampTargetPets,
		}, rampDedupeKey(commercialUserID, vetID)); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	var triennialN, subN int
	if err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE pe.plan_code='triennial')::int,
			COUNT(*)::int
		FROM billing.commercial_commission_ledger cl
		JOIN billing.pet_entitlements pe ON pe.id = cl.source_id
		WHERE cl.commercial_user_id=$1
		  AND cl.period_ym=$2
		  AND cl.source_type='subscription_pct'`, commercialUserID, month).Scan(&triennialN, &subN); err != nil {
		return err
	}
	pct := 0
	if subN > 0 {
		pct = triennialN * 100 / subN
	}
	if subN > 0 && pct >= commercialMixTargetPct {
		if err := s.upsertBonusAward(ctx, CommercialBonusAward{
			CommercialUserID: commercialUserID,
			BonusCode:        BonusCodeCommercialMix,
			AmountCents:      commercialMixAmountCents,
			Status:           BonusStatusEarned,
			PeriodYM:         month,
			Progress:         pct,
			Target:           commercialMixTargetPct,
		}, mixDedupeKey(commercialUserID, month)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) upsertBonusAward(ctx context.Context, a CommercialBonusAward, dedupeKey string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing.commercial_bonus_awards (
			id, commercial_user_id, bonus_code, amount_cents, status,
			period_ym, vet_user_id, progress, target, dedupe_key, earned_at
		) VALUES (
			$1,$2,$3,$4,'earned',
			NULLIF($5,''), NULLIF($6,'')::uuid, $7,$8,$9,NOW()
		)
		ON CONFLICT (dedupe_key) DO UPDATE SET
			progress = EXCLUDED.progress,
			target = EXCLUDED.target
		WHERE billing.commercial_bonus_awards.status = 'earned'`,
		uuid.NewString(), a.CommercialUserID, a.BonusCode, a.AmountCents,
		a.PeriodYM, a.VetUserID, a.Progress, a.Target, dedupeKey)
	return err
}

// MarkCommercialBonusPaid marks an earned award as paid by an admin.
func (s *Store) MarkCommercialBonusPaid(ctx context.Context, awardID, adminUserID string) (CommercialBonusAward, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT status FROM billing.commercial_bonus_awards WHERE id=$1`, awardID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialBonusAward{}, ErrNotFound
	}
	if err != nil {
		return CommercialBonusAward{}, err
	}
	if status == BonusStatusPaid {
		return CommercialBonusAward{}, ErrBonusAlreadyPaid
	}
	if status != BonusStatusEarned {
		return CommercialBonusAward{}, ErrBonusNotEarned
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE billing.commercial_bonus_awards
		SET status='paid', paid_at=NOW(), paid_by_admin_id=$2
		WHERE id=$1 AND status='earned'`, awardID, adminUserID)
	if err != nil {
		return CommercialBonusAward{}, err
	}
	return s.getCommercialBonusAward(ctx, awardID)
}

func (s *Store) getCommercialBonusAward(ctx context.Context, awardID string) (CommercialBonusAward, error) {
	var a CommercialBonusAward
	var period, vetID, vetEmail, vetName *string
	var paidAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.commercial_user_id::text, cu.full_name, cu.email,
			a.bonus_code, a.amount_cents, a.status, a.period_ym, a.vet_user_id::text,
			vu.email, vu.full_name, a.progress, a.target, a.earned_at, a.paid_at
		FROM billing.commercial_bonus_awards a
		JOIN identity.users cu ON cu.id = a.commercial_user_id
		LEFT JOIN identity.users vu ON vu.id = a.vet_user_id
		WHERE a.id=$1`, awardID).Scan(
		&a.ID, &a.CommercialUserID, &a.CommercialFullName, &a.CommercialEmail,
		&a.BonusCode, &a.AmountCents, &a.Status, &period, &vetID,
		&vetEmail, &vetName, &a.Progress, &a.Target, &a.EarnedAt, &paidAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialBonusAward{}, ErrNotFound
	}
	if err != nil {
		return CommercialBonusAward{}, err
	}
	if period != nil {
		a.PeriodYM = *period
	}
	if vetID != nil {
		a.VetUserID = *vetID
	}
	if vetEmail != nil {
		a.VetEmail = *vetEmail
	}
	if vetName != nil {
		a.VetFullName = *vetName
	}
	a.PaidAt = paidAt
	return a, nil
}

// ListCommercialBonusTrackRows returns admin suivi rows (awards + live in-progress).
func (s *Store) ListCommercialBonusTrackRows(ctx context.Context, statusFilter, commercialFilter string) ([]CommercialBonusTrackRow, error) {
	commercials, err := s.ListAllCommercials(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CommercialBonusTrackRow, 0)
	for _, c := range commercials {
		if commercialFilter != "" && c.UserID != commercialFilter {
			continue
		}
		if err := s.SyncCommercialBonusAwards(ctx, c.UserID); err != nil {
			return nil, err
		}
		rows, err := s.commercialBonusTrackForUser(ctx, c)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if statusFilter != "" && r.Status != statusFilter {
				continue
			}
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) commercialBonusTrackForUser(ctx context.Context, c CommercialRow) ([]CommercialBonusTrackRow, error) {
	month := PeriodYM(time.Now())
	out := make([]CommercialBonusTrackRow, 0)

	awards, err := s.listBonusAwardsForCommercial(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	awardedRampVets := map[string]bool{}
	mixAwardedForMonth := false
	for _, a := range awards {
		row := CommercialBonusTrackRow{
			AwardID:            a.ID,
			CommercialUserID:   c.UserID,
			CommercialFullName: c.FullName,
			CommercialEmail:    c.Email,
			BonusCode:          a.BonusCode,
			AmountCents:        a.AmountCents,
			Status:             a.Status,
			Progress:           a.Progress,
			Target:             a.Target,
			PeriodYM:           a.PeriodYM,
			VetUserID:          a.VetUserID,
			VetEmail:           a.VetEmail,
			VetFullName:        a.VetFullName,
		}
		out = append(out, row)
		if a.BonusCode == BonusCodeCommercialRamp && a.VetUserID != "" {
			awardedRampVets[a.VetUserID] = true
		}
		if a.BonusCode == BonusCodeCommercialMix && a.PeriodYM == month {
			mixAwardedForMonth = true
		}
	}

	// Live ramp progress for vets without an award yet.
	rampRows, err := s.pool.Query(ctx, `
		SELECT cl.vet_user_id::text, vu.email, vu.full_name, COUNT(*)::int
		FROM billing.commercial_commission_ledger cl
		JOIN identity.users vu ON vu.id = cl.vet_user_id
		WHERE cl.commercial_user_id=$1
		  AND cl.source_type='subscription_pct'
		  AND cl.accrued_at >= NOW() - INTERVAL '60 days'
		GROUP BY cl.vet_user_id, vu.email, vu.full_name
		HAVING COUNT(*) > 0
		ORDER BY COUNT(*) DESC`, c.UserID)
	if err != nil {
		return nil, err
	}
	defer rampRows.Close()
	for rampRows.Next() {
		var vetID, email, name string
		var cnt int
		if err := rampRows.Scan(&vetID, &email, &name, &cnt); err != nil {
			return nil, err
		}
		if awardedRampVets[vetID] {
			continue
		}
		status := "in_progress"
		if cnt >= commercialRampTargetPets {
			status = BonusStatusEarned
		}
		out = append(out, CommercialBonusTrackRow{
			CommercialUserID:   c.UserID,
			CommercialFullName: c.FullName,
			CommercialEmail:    c.Email,
			BonusCode:          BonusCodeCommercialRamp,
			AmountCents:        commercialRampAmountCents,
			Status:             status,
			Progress:           cnt,
			Target:             commercialRampTargetPets,
			VetUserID:          vetID,
			VetEmail:           email,
			VetFullName:        name,
		})
	}
	if err := rampRows.Err(); err != nil {
		return nil, err
	}

	if !mixAwardedForMonth {
		var triennialN, subN int
		if err := s.pool.QueryRow(ctx, `
			SELECT
				COUNT(*) FILTER (WHERE pe.plan_code='triennial')::int,
				COUNT(*)::int
			FROM billing.commercial_commission_ledger cl
			JOIN billing.pet_entitlements pe ON pe.id = cl.source_id
			WHERE cl.commercial_user_id=$1
			  AND cl.period_ym=$2
			  AND cl.source_type='subscription_pct'`, c.UserID, month).Scan(&triennialN, &subN); err != nil {
			return nil, err
		}
		pct := 0
		if subN > 0 {
			pct = triennialN * 100 / subN
		}
		status := "available"
		switch {
		case subN > 0 && pct >= commercialMixTargetPct:
			status = BonusStatusEarned
		case subN > 0:
			status = "in_progress"
		}
		if status != "available" {
			out = append(out, CommercialBonusTrackRow{
				CommercialUserID:   c.UserID,
				CommercialFullName: c.FullName,
				CommercialEmail:    c.Email,
				BonusCode:          BonusCodeCommercialMix,
				AmountCents:        commercialMixAmountCents,
				Status:             status,
				Progress:           pct,
				Target:             commercialMixTargetPct,
				PeriodYM:           month,
			})
		}
	}

	return out, nil
}

func (s *Store) listBonusAwardsForCommercial(ctx context.Context, commercialUserID string) ([]CommercialBonusAward, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.commercial_user_id::text, cu.full_name, cu.email,
			a.bonus_code, a.amount_cents, a.status, a.period_ym, a.vet_user_id::text,
			vu.email, vu.full_name, a.progress, a.target, a.earned_at, a.paid_at
		FROM billing.commercial_bonus_awards a
		JOIN identity.users cu ON cu.id = a.commercial_user_id
		LEFT JOIN identity.users vu ON vu.id = a.vet_user_id
		WHERE a.commercial_user_id=$1
		ORDER BY a.earned_at DESC`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialBonusAward, 0)
	for rows.Next() {
		var a CommercialBonusAward
		var period, vetID, vetEmail, vetName *string
		var paidAt *time.Time
		if err := rows.Scan(
			&a.ID, &a.CommercialUserID, &a.CommercialFullName, &a.CommercialEmail,
			&a.BonusCode, &a.AmountCents, &a.Status, &period, &vetID,
			&vetEmail, &vetName, &a.Progress, &a.Target, &a.EarnedAt, &paidAt,
		); err != nil {
			return nil, err
		}
		if period != nil {
			a.PeriodYM = *period
		}
		if vetID != nil {
			a.VetUserID = *vetID
		}
		if vetEmail != nil {
			a.VetEmail = *vetEmail
		}
		if vetName != nil {
			a.VetFullName = *vetName
		}
		a.PaidAt = paidAt
		out = append(out, a)
	}
	return out, rows.Err()
}
