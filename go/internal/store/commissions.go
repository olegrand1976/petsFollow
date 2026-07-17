package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPayoutNotOpen   = errors.New("payout_not_open")
	ErrPayoutNotClosed = errors.New("payout_not_closed")
)

type CommissionTier struct {
	MinClients int  `json:"minClients"`
	MaxClients *int `json:"maxClients,omitempty"`
	RateBps    int  `json:"rateBps"`
}

type CommissionLedgerRow struct {
	ID              string    `json:"id"`
	VetUserID       string    `json:"vetUserId"`
	ClientUserID    string    `json:"clientUserId"`
	PetID           string    `json:"petId"`
	EntitlementID   string    `json:"entitlementId"`
	BaseAmountCents int       `json:"baseAmountCents"`
	RateBps         int       `json:"rateBps"`
	CommissionCents int       `json:"commissionCents"`
	PeriodYM        string    `json:"periodYm"`
	AccruedAt       time.Time `json:"accruedAt"`
	ClientEmail     string    `json:"clientEmail,omitempty"`
	PetName         string    `json:"petName,omitempty"`
	VetEmail        string    `json:"vetEmail,omitempty"`
	VetFullName     string    `json:"vetFullName,omitempty"`
}

type PayoutRun struct {
	ID         string     `json:"id"`
	PeriodYM   string     `json:"periodYm"`
	Status     string     `json:"status"`
	ClosedAt   *time.Time `json:"closedAt,omitempty"`
	PaidAt     *time.Time `json:"paidAt,omitempty"`
	Note       string     `json:"note"`
	CreatedAt  time.Time  `json:"createdAt"`
	TotalCents int        `json:"totalCents,omitempty"`
	LineCount  int        `json:"lineCount,omitempty"`
}

type PayoutLine struct {
	ID              string `json:"id"`
	RunID           string `json:"runId"`
	VetUserID       string `json:"vetUserId"`
	VetEmail        string `json:"vetEmail"`
	VetFullName     string `json:"vetFullName"`
	EligibleClients int    `json:"eligibleClients"`
	LedgerCount     int    `json:"ledgerCount"`
	AmountCents     int    `json:"amountCents"`
	Status          string `json:"status"`
}

type VetCommissionSummary struct {
	EligibleClients    int                   `json:"eligibleClients"`
	CurrentRateBps     int                   `json:"currentRateBps"`
	NextTierMinClients *int                  `json:"nextTierMinClients,omitempty"`
	MonthPeriodYM      string                `json:"monthPeriodYm"`
	MonthEarnedCents   int                   `json:"monthEarnedCents"`
	LifetimeEarnedCents   int                   `json:"lifetimeEarnedCents"`
	Tiers              []CommissionTier      `json:"tiers"`
	RecentLedger       []CommissionLedgerRow `json:"recentLedger"`
	PayoutHistory      []PayoutLineHistory   `json:"payoutHistory"`
}

type PayoutLineHistory struct {
	PeriodYM    string     `json:"periodYm"`
	AmountCents int        `json:"amountCents"`
	Status      string     `json:"status"`
	RunStatus   string     `json:"runStatus"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
}

func PeriodYM(t time.Time) string {
	return t.UTC().Format("2006-01")
}

func ValidPeriodYM(period string) bool {
	if len(period) != 7 || period[4] != '-' {
		return false
	}
	_, err := time.Parse("2006-01", period)
	return err == nil
}

func NextPeriodYM(period string) (string, error) {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return "", err
	}
	return t.AddDate(0, 1, 0).Format("2006-01"), nil
}

func (s *Store) ResolveOpenPeriodYM(ctx context.Context, preferred string) (string, error) {
	period := preferred
	for i := 0; i < 24; i++ {
		var status string
		err := s.pool.QueryRow(ctx, `SELECT status FROM billing.payout_runs WHERE period_ym=$1`, period).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return period, nil
		}
		if err != nil {
			return "", err
		}
		if status == "open" {
			return period, nil
		}
		next, err := NextPeriodYM(period)
		if err != nil {
			return "", err
		}
		period = next
	}
	return period, nil
}

func (s *Store) EnsureDefaultCommissionTiers(ctx context.Context) error {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM billing.commission_tiers`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	type tierDef struct {
		min, max, bps int
		hasMax        bool
	}
	tiers := []tierDef{
		{1, 4, 500, true},
		{5, 14, 800, true},
		{15, 39, 1000, true},
		{40, 0, 1200, false},
	}
	for _, t := range tiers {
		var max any
		if t.hasMax {
			max = t.max
		}
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO billing.commission_tiers (id, min_clients, max_clients, rate_bps)
			VALUES ($1, $2, $3, $4)`, uuid.NewString(), t.min, max, t.bps); err != nil {
			return err
		}
	}
	return nil
}

// ReplaceCommissionTiers replaces the full progressive ladder (admin).
func (s *Store) ReplaceCommissionTiers(ctx context.Context, tiers []CommissionTier) error {
	if len(tiers) == 0 {
		return errors.New("empty_tiers")
	}
	openEnded := 0
	prevMax := 0
	for i, t := range tiers {
		if t.MinClients < 1 || t.RateBps < 0 || t.RateBps > 5000 {
			return errors.New("invalid_tier")
		}
		if t.MaxClients == nil {
			openEnded++
			if i != len(tiers)-1 {
				return errors.New("open_ended_must_be_last")
			}
		} else {
			if *t.MaxClients < t.MinClients {
				return errors.New("invalid_tier_range")
			}
			prevMax = *t.MaxClients
			_ = prevMax
		}
		if i > 0 {
			prev := tiers[i-1]
			if prev.MaxClients == nil {
				return errors.New("invalid_tier_order")
			}
			if t.MinClients != *prev.MaxClients+1 {
				return errors.New("tiers_not_contiguous")
			}
		} else if t.MinClients != 1 {
			return errors.New("tiers_must_start_at_1")
		}
	}
	if openEnded != 1 {
		return errors.New("exactly_one_open_ended_tier")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Block concurrent accruals from reading an empty tiers table mid-replace.
	if _, err := tx.Exec(ctx, `LOCK TABLE billing.commission_tiers IN EXCLUSIVE MODE`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM billing.commission_tiers`); err != nil {
		return err
	}
	for _, t := range tiers {
		var max any
		if t.MaxClients != nil {
			max = *t.MaxClients
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO billing.commission_tiers (id, min_clients, max_clients, rate_bps)
			VALUES ($1, $2, $3, $4)`, uuid.NewString(), t.MinClients, max, t.RateBps); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) ListCommissionTiers(ctx context.Context) ([]CommissionTier, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT min_clients, max_clients, rate_bps
		FROM billing.commission_tiers
		ORDER BY min_clients`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CommissionTier
	for rows.Next() {
		var t CommissionTier
		var max *int
		if err := rows.Scan(&t.MinClients, &max, &t.RateBps); err != nil {
			return nil, err
		}
		t.MaxClients = max
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) ResolveCommissionRateBps(ctx context.Context, clientRank int) (int, error) {
	if clientRank < 1 {
		clientRank = 1
	}
	var rate int
	err := s.pool.QueryRow(ctx, `
		SELECT rate_bps FROM billing.commission_tiers
		WHERE min_clients <= $1 AND (max_clients IS NULL OR max_clients >= $1)
		ORDER BY min_clients DESC
		LIMIT 1`, clientRank).Scan(&rate)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return rate, err
}

func (s *Store) AccrueCommissionForPetActivation(ctx context.Context, petID string) error {
	ent, err := s.GetEntitlementByPetID(ctx, petID)
	if err != nil {
		return err
	}
	if ent.Status != "active" && ent.Status != "past_due" && ent.Status != "cancelled" {
		return nil
	}
	if err := s.EnsureDefaultCommissionTiers(ctx); err != nil {
		return err
	}

	var vetUserID string
	err = s.pool.QueryRow(ctx, `
		SELECT pc.vet_user_id::text
		FROM pets.pets p
		JOIN practice.practice_clients pc
			ON pc.client_user_id = p.owner_user_id AND pc.practice_id = p.practice_id
		WHERE p.id = $1
		ORDER BY pc.created_at DESC
		LIMIT 1`, petID).Scan(&vetUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Serialize accruals per vet to keep progressive ranks correct.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, vetUserID); err != nil {
		return err
	}

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM billing.commission_ledger WHERE entitlement_id=$1)`, ent.ID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return s.AccrueCommercialForSubscription(ctx, petID)
	}

	var alreadyClient bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.commission_ledger
			WHERE vet_user_id=$1 AND client_user_id=$2
		)`, vetUserID, ent.OwnerUserID).Scan(&alreadyClient); err != nil {
		return err
	}

	var rateBps int
	if alreadyClient {
		if err := tx.QueryRow(ctx, `
			SELECT rate_bps FROM billing.commission_ledger
			WHERE vet_user_id=$1 AND client_user_id=$2
			ORDER BY accrued_at ASC LIMIT 1`, vetUserID, ent.OwnerUserID).Scan(&rateBps); err != nil {
			return err
		}
	} else {
		var distinctClients int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(DISTINCT client_user_id) FROM billing.commission_ledger WHERE vet_user_id=$1`,
			vetUserID).Scan(&distinctClients); err != nil {
			return err
		}
		err = tx.QueryRow(ctx, `
			SELECT rate_bps FROM billing.commission_tiers
			WHERE min_clients <= $1 AND (max_clients IS NULL OR max_clients >= $1)
			ORDER BY min_clients DESC
			LIMIT 1`, distinctClients+1).Scan(&rateBps)
		if errors.Is(err, pgx.ErrNoRows) {
			// Never persist 0% silently (race with ReplaceCommissionTiers, or empty table).
			return fmt.Errorf("no_commission_tiers")
		} else if err != nil {
			return err
		}
	}

	preferred := PeriodYM(time.Now())
	period, err := s.ResolveOpenPeriodYM(ctx, preferred)
	if err != nil {
		return err
	}

	commission := ent.AmountCents * rateBps / 10000
	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.commission_ledger (
			id, vet_user_id, client_user_id, pet_id, entitlement_id,
			base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (entitlement_id) DO NOTHING`,
		uuid.NewString(), vetUserID, ent.OwnerUserID, petID, ent.ID,
		ent.AmountCents, rateBps, commission, period); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// Commercial commission is accrued after the vet ledger is committed.
	return s.AccrueCommercialForSubscription(ctx, petID)
}

func (s *Store) CountVetEligibleClients(ctx context.Context, vetUserID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT client_user_id) FROM billing.commission_ledger WHERE vet_user_id=$1`, vetUserID).Scan(&n)
	return n, err
}

func (s *Store) GetOrCreatePayoutRun(ctx context.Context, periodYM string) (PayoutRun, error) {
	var r PayoutRun
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, period_ym, status, closed_at, paid_at, COALESCE(note,''), created_at
		FROM billing.payout_runs WHERE period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt)
	if err == nil {
		return r, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PayoutRun{}, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO billing.payout_runs (id, period_ym, status)
		VALUES ($1, $2, 'open')
		ON CONFLICT (period_ym) DO NOTHING`, uuid.NewString(), periodYM)
	if err != nil {
		return PayoutRun{}, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, period_ym, status, closed_at, paid_at, COALESCE(note,''), created_at
		FROM billing.payout_runs WHERE period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt)
	return r, err
}

func (s *Store) ListPayoutRuns(ctx context.Context) ([]PayoutRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id::text, r.period_ym, r.status, r.closed_at, r.paid_at, COALESCE(r.note,''), r.created_at,
			COALESCE((SELECT SUM(amount_cents) FROM billing.payout_lines pl WHERE pl.run_id=r.id), 0),
			COALESCE((SELECT COUNT(*) FROM billing.payout_lines pl WHERE pl.run_id=r.id), 0)
		FROM billing.payout_runs r
		ORDER BY r.period_ym DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PayoutRun, 0)
	for rows.Next() {
		var r PayoutRun
		if err := rows.Scan(&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt, &r.TotalCents, &r.LineCount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetPayoutRunByPeriod(ctx context.Context, periodYM string) (PayoutRun, error) {
	var r PayoutRun
	err := s.pool.QueryRow(ctx, `
		SELECT r.id::text, r.period_ym, r.status, r.closed_at, r.paid_at, COALESCE(r.note,''), r.created_at,
			COALESCE((SELECT SUM(amount_cents) FROM billing.payout_lines pl WHERE pl.run_id=r.id), 0),
			COALESCE((SELECT COUNT(*) FROM billing.payout_lines pl WHERE pl.run_id=r.id), 0)
		FROM billing.payout_runs r WHERE r.period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt, &r.TotalCents, &r.LineCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return PayoutRun{}, ErrNotFound
	}
	return r, err
}

func (s *Store) ListPayoutLines(ctx context.Context, runID string) ([]PayoutLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pl.id::text, pl.run_id::text, pl.vet_user_id::text, u.email, u.full_name,
			pl.eligible_clients, pl.ledger_count, pl.amount_cents, pl.status
		FROM billing.payout_lines pl
		JOIN identity.users u ON u.id = pl.vet_user_id
		WHERE pl.run_id=$1
		ORDER BY pl.amount_cents DESC, u.full_name`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayoutLine
	for rows.Next() {
		var l PayoutLine
		if err := rows.Scan(&l.ID, &l.RunID, &l.VetUserID, &l.VetEmail, &l.VetFullName,
			&l.EligibleClients, &l.LedgerCount, &l.AmountCents, &l.Status); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) PreviewPeriodCommissions(ctx context.Context, periodYM string) ([]PayoutLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cl.vet_user_id::text, u.email, u.full_name,
			COUNT(DISTINCT cl.client_user_id)::int,
			COUNT(*)::int,
			COALESCE(SUM(cl.commission_cents),0)::int
		FROM billing.commission_ledger cl
		JOIN identity.users u ON u.id = cl.vet_user_id
		WHERE cl.period_ym=$1
		GROUP BY cl.vet_user_id, u.email, u.full_name
		ORDER BY SUM(cl.commission_cents) DESC`, periodYM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayoutLine
	for rows.Next() {
		var l PayoutLine
		l.Status = "pending"
		if err := rows.Scan(&l.VetUserID, &l.VetEmail, &l.VetFullName,
			&l.EligibleClients, &l.LedgerCount, &l.AmountCents); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) ClosePayoutRun(ctx context.Context, periodYM string) (PayoutRun, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PayoutRun{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.payout_runs (id, period_ym, status)
		VALUES ($1, $2, 'open')
		ON CONFLICT (period_ym) DO NOTHING`, uuid.NewString(), periodYM); err != nil {
		return PayoutRun{}, err
	}

	var runID, status string
	if err := tx.QueryRow(ctx, `
		SELECT id::text, status FROM billing.payout_runs WHERE period_ym=$1 FOR UPDATE`,
		periodYM).Scan(&runID, &status); err != nil {
		return PayoutRun{}, err
	}
	if status != "open" {
		return PayoutRun{}, ErrPayoutNotOpen
	}

	if _, err := tx.Exec(ctx, `DELETE FROM billing.payout_lines WHERE run_id=$1`, runID); err != nil {
		return PayoutRun{}, err
	}

	rows, err := tx.Query(ctx, `
		SELECT cl.vet_user_id::text,
			COUNT(DISTINCT cl.client_user_id)::int,
			COUNT(*)::int,
			COALESCE(SUM(cl.commission_cents),0)::int
		FROM billing.commission_ledger cl
		WHERE cl.period_ym=$1
		GROUP BY cl.vet_user_id`, periodYM)
	if err != nil {
		return PayoutRun{}, err
	}

	type aggRow struct {
		vetID                        string
		clients, ledgerCount, amount int
	}
	var aggs []aggRow
	for rows.Next() {
		var a aggRow
		if err := rows.Scan(&a.vetID, &a.clients, &a.ledgerCount, &a.amount); err != nil {
			rows.Close()
			return PayoutRun{}, err
		}
		aggs = append(aggs, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return PayoutRun{}, err
	}

	for _, a := range aggs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO billing.payout_lines (id, run_id, vet_user_id, eligible_clients, ledger_count, amount_cents, status)
			VALUES ($1,$2,$3,$4,$5,$6,'pending')`,
			uuid.NewString(), runID, a.vetID, a.clients, a.ledgerCount, a.amount); err != nil {
			return PayoutRun{}, err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE billing.payout_runs SET status='closed', closed_at=NOW()
		WHERE id=$1 AND status='open'`, runID); err != nil {
		return PayoutRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PayoutRun{}, err
	}
	return s.GetPayoutRunByPeriod(ctx, periodYM)
}

func (s *Store) MarkPayoutRunPaid(ctx context.Context, periodYM, note string) (PayoutRun, error) {
	run, err := s.GetPayoutRunByPeriod(ctx, periodYM)
	if err != nil {
		return PayoutRun{}, err
	}
	if run.Status != "closed" && run.Status != "paid" {
		return PayoutRun{}, ErrPayoutNotClosed
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PayoutRun{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		UPDATE billing.payout_runs SET status='paid', paid_at=COALESCE(paid_at, NOW()), note=$2 WHERE id=$1`,
		run.ID, note); err != nil {
		return PayoutRun{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE billing.payout_lines SET status='paid' WHERE run_id=$1`, run.ID); err != nil {
		return PayoutRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PayoutRun{}, err
	}
	return s.GetPayoutRunByPeriod(ctx, periodYM)
}

func (s *Store) VetCommissionSummary(ctx context.Context, vetUserID string) (VetCommissionSummary, error) {
	_ = s.EnsureDefaultCommissionTiers(ctx)
	tiers, err := s.ListCommissionTiers(ctx)
	if err != nil {
		return VetCommissionSummary{}, err
	}
	clients, err := s.CountVetEligibleClients(ctx, vetUserID)
	if err != nil {
		return VetCommissionSummary{}, err
	}
	rateRank := clients
	if rateRank == 0 {
		rateRank = 1
	}
	rate, err := s.ResolveCommissionRateBps(ctx, rateRank)
	if err != nil {
		return VetCommissionSummary{}, err
	}
	month := PeriodYM(time.Now())
	var monthEarned, lifetime int
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0) FROM billing.commission_ledger
		WHERE vet_user_id=$1 AND period_ym=$2`, vetUserID, month).Scan(&monthEarned)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0) FROM billing.commission_ledger
		WHERE vet_user_id=$1`, vetUserID).Scan(&lifetime)

	var nextMin *int
	for _, t := range tiers {
		if t.MinClients > clients {
			m := t.MinClients
			nextMin = &m
			break
		}
	}

	ledgerRows, err := s.pool.Query(ctx, `
		SELECT cl.id::text, cl.vet_user_id::text, cl.client_user_id::text, cl.pet_id::text, cl.entitlement_id::text,
			cl.base_amount_cents, cl.rate_bps, cl.commission_cents, cl.period_ym, cl.accrued_at,
			cu.email, p.name
		FROM billing.commission_ledger cl
		JOIN identity.users cu ON cu.id = cl.client_user_id
		JOIN pets.pets p ON p.id = cl.pet_id
		WHERE cl.vet_user_id=$1
		ORDER BY cl.accrued_at DESC
		LIMIT 50`, vetUserID)
	if err != nil {
		return VetCommissionSummary{}, err
	}
	defer ledgerRows.Close()
	var recent []CommissionLedgerRow
	for ledgerRows.Next() {
		var r CommissionLedgerRow
		if err := ledgerRows.Scan(&r.ID, &r.VetUserID, &r.ClientUserID, &r.PetID, &r.EntitlementID,
			&r.BaseAmountCents, &r.RateBps, &r.CommissionCents, &r.PeriodYM, &r.AccruedAt,
			&r.ClientEmail, &r.PetName); err != nil {
			return VetCommissionSummary{}, err
		}
		recent = append(recent, r)
	}

	histRows, err := s.pool.Query(ctx, `
		SELECT r.period_ym, pl.amount_cents, pl.status, r.status, r.paid_at
		FROM billing.payout_lines pl
		JOIN billing.payout_runs r ON r.id = pl.run_id
		WHERE pl.vet_user_id=$1
		ORDER BY r.period_ym DESC
		LIMIT 24`, vetUserID)
	if err != nil {
		return VetCommissionSummary{}, err
	}
	defer histRows.Close()
	var history []PayoutLineHistory
	for histRows.Next() {
		var h PayoutLineHistory
		if err := histRows.Scan(&h.PeriodYM, &h.AmountCents, &h.Status, &h.RunStatus, &h.PaidAt); err != nil {
			return VetCommissionSummary{}, err
		}
		history = append(history, h)
	}

	return VetCommissionSummary{
		EligibleClients:    clients,
		CurrentRateBps:     rate,
		NextTierMinClients: nextMin,
		MonthPeriodYM:      month,
		MonthEarnedCents:   monthEarned,
		LifetimeEarnedCents:   lifetime,
		Tiers:              tiers,
		RecentLedger:       recent,
		PayoutHistory:      history,
	}, nil
}

func (s *Store) AdminCommissionPeriodDetail(ctx context.Context, periodYM string) (map[string]any, error) {
	run, err := s.GetPayoutRunByPeriod(ctx, periodYM)
	notFound := errors.Is(err, ErrNotFound)
	if err != nil && !notFound {
		return nil, err
	}
	var lines []PayoutLine
	if !notFound && (run.Status == "closed" || run.Status == "paid") {
		lines, err = s.ListPayoutLines(ctx, run.ID)
		if err != nil {
			return nil, err
		}
	} else {
		lines, err = s.PreviewPeriodCommissions(ctx, periodYM)
		if err != nil {
			return nil, err
		}
		if notFound {
			run = PayoutRun{PeriodYM: periodYM, Status: "open"}
		}
	}
	total := 0
	for _, l := range lines {
		total += l.AmountCents
	}
	return map[string]any{
		"run":        run,
		"lines":      lines,
		"totalCents": total,
		"periodYm":   periodYM,
	}, nil
}

func (s *Store) AccrueAllActiveEntitlements(ctx context.Context) error {
	_ = s.EnsureDefaultCommissionTiers(ctx)
	rows, err := s.pool.Query(ctx, `
		SELECT pet_id::text FROM billing.pet_entitlements
		WHERE status IN ('active','past_due','cancelled')
		ORDER BY created_at`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var petID string
		if err := rows.Scan(&petID); err != nil {
			return err
		}
		if err := s.AccrueCommissionForPetActivation(ctx, petID); err != nil {
			return err
		}
	}
	return rows.Err()
}
