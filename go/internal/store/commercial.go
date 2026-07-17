package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"golang.org/x/crypto/bcrypt"
)

// DefaultCommercialCommissionRateBps is used when settings row is missing (12%).
const DefaultCommercialCommissionRateBps = 1200

// CommercialCommissionCents returns the flat commercial commission on a base amount.
// rateBps may be 0 (explicit zero commission). Negative rates are clamped to 0.
func CommercialCommissionCents(baseAmountCents, rateBps int) int {
	if rateBps < 0 {
		rateBps = 0
	}
	if baseAmountCents < 0 {
		baseAmountCents = 0
	}
	return baseAmountCents * rateBps / 10000
}

func (s *Store) GetCommercialRateBps(ctx context.Context) (int, error) {
	var bps int
	err := s.pool.QueryRow(ctx, `
		SELECT commercial_rate_bps FROM billing.commission_settings WHERE id=1`).Scan(&bps)
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultCommercialCommissionRateBps, nil
	}
	if err != nil {
		// Fail-closed: never invent a monetary rate on DB errors.
		return 0, err
	}
	return bps, nil
}

func (s *Store) SetCommercialRateBps(ctx context.Context, rateBps int) error {
	if rateBps < 0 || rateBps > 5000 {
		return errors.New("invalid_rate_bps")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing.commission_settings (id, commercial_rate_bps, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE SET commercial_rate_bps=$1, updated_at=NOW()`, rateBps)
	return err
}

func (s *Store) EnsureCommissionSettings(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing.commission_settings (id, commercial_rate_bps)
		VALUES (1, $1)
		ON CONFLICT (id) DO NOTHING`, DefaultCommercialCommissionRateBps)
	return err
}

type CommercialVetRow struct {
	UserID       string `json:"userId"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	PracticeName string `json:"practiceName"`
	ClientCount  int    `json:"clientCount"`
}

type CommercialRow struct {
	UserID      string `json:"userId"`
	FullName    string `json:"fullName"`
	Email       string `json:"email"`
	ClientCount int    `json:"clientCount"`
}

type CommercialLedgerRow struct {
	ID              string    `json:"id"`
	SourceType      string    `json:"sourceType"`
	VetEmail        string    `json:"vetEmail"`
	ClientEmail     string    `json:"clientEmail"`
	BaseAmountCents int       `json:"baseAmountCents"`
	RateBps         int       `json:"rateBps"`
	CommissionCents int       `json:"commissionCents"`
	PeriodYM        string    `json:"periodYm"`
	AccruedAt       time.Time `json:"accruedAt"`
}

type EncodeVetInput struct {
	Email            string
	Password         string
	FullName         string
	PracticeName     string
	Phone            string
	City             string
	PostalCode       string
	AddressLine1     string
	ContactEmail     string
	PreferredLocale  string
	AutoReplyDefault string
}

func (s *Store) CreateCommercialUser(ctx context.Context, email, password, fullName string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	userID := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, $4, 'commercial', NULL, NOW())`,
		userID, email, string(hash), fullName)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Store) EncodeVetForCommercial(ctx context.Context, commercialUserID string, in EncodeVetInput) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	practiceID := uuid.NewString()
	userID := uuid.NewString()
	contactEmail := in.ContactEmail
	if contactEmail == "" {
		contactEmail = in.Email
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practices (id, name, phone, contact_email, address_line1, city, postal_code, profile_completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		practiceID, in.PracticeName, in.Phone, contactEmail, in.AddressLine1, in.City, in.PostalCode); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, assigned_commercial_id, preferred_locale)
		VALUES ($1, $2, $3, $4, 'vet', $5, NOW(), $6, $7)`,
		userID, in.Email, string(hash), in.FullName, practiceID, commercialUserID, i18n.NormalizeLocale(in.PreferredLocale)); err != nil {
		return "", err
	}
	autoReply := in.AutoReplyDefault
	if autoReply == "" {
		autoReply = "Je suis indisponible, je reviens vers vous rapidement."
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, 'available', $3)`, userID, practiceID, autoReply); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, true, true)`, userID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *Store) AssignVetToCommercial(ctx context.Context, vetUserID, commercialUserID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id=$2 WHERE id=$1 AND role='vet'`, vetUserID, commercialUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetAssignedCommercialID(ctx context.Context, vetUserID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`, vetUserID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}

func (s *Store) ListCommercialVets(ctx context.Context, commercialUserID string) ([]CommercialVetRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email, COALESCE(pr.name,''),
			COALESCE((SELECT COUNT(*)::int FROM practice.practice_clients pc WHERE pc.vet_user_id = u.id), 0)
		FROM identity.users u
		LEFT JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE u.role='vet' AND u.assigned_commercial_id=$1
		ORDER BY u.full_name`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialVetRow, 0)
	for rows.Next() {
		var v CommercialVetRow
		if err := rows.Scan(&v.UserID, &v.FullName, &v.Email, &v.PracticeName, &v.ClientCount); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) ListAllCommercials(ctx context.Context) ([]CommercialRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE((SELECT COUNT(*)::int FROM identity.users v WHERE v.role='vet' AND v.assigned_commercial_id = u.id), 0)
		FROM identity.users u
		WHERE u.role='commercial'
		ORDER BY u.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommercialRow, 0)
	for rows.Next() {
		var c CommercialRow
		if err := rows.Scan(&c.UserID, &c.FullName, &c.Email, &c.ClientCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CommercialOverview(ctx context.Context, commercialUserID string) (map[string]any, error) {
	var assignedVets, prospectsTotal, prospectsNew, prospectsConverted int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM identity.users WHERE role='vet' AND assigned_commercial_id=$1`,
		commercialUserID).Scan(&assignedVets); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (WHERE status='new')::int,
			COUNT(*) FILTER (WHERE status='converted')::int
		FROM sales.prospects WHERE commercial_user_id=$1`,
		commercialUserID).Scan(&prospectsTotal, &prospectsNew, &prospectsConverted); err != nil {
		return nil, err
	}

	month := PeriodYM(time.Now())
	var monthEarned, lifetime, subRevenue, addonRevenue int
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND period_ym=$2`, commercialUserID, month).Scan(&monthEarned); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1`, commercialUserID).Scan(&lifetime); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(base_amount_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type IN ('subscription_pct','subscription_mirror')`, commercialUserID).Scan(&subRevenue); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(base_amount_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='addon_pct'`, commercialUserID).Scan(&addonRevenue); err != nil {
		return nil, err
	}

	return map[string]any{
		"assignedVets":                  assignedVets,
		"prospectsTotal":                prospectsTotal,
		"prospectsNew":                  prospectsNew,
		"prospectsConverted":            prospectsConverted,
		"monthEarnedCents":              monthEarned,
		"lifetimeEarnedCents":           lifetime,
		"linkedSubscriptionRevenueCents": subRevenue,
		"linkedAddonRevenueCents":        addonRevenue,
	}, nil
}

func (s *Store) GetCommercialCommissionSummary(ctx context.Context, commercialUserID string) (map[string]any, error) {
	month := PeriodYM(time.Now())
	var monthEarned, lifetime, subCommission, addonCommission int
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND period_ym=$2`, commercialUserID, month).Scan(&monthEarned); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1`, commercialUserID).Scan(&lifetime); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type IN ('subscription_pct','subscription_mirror')`, commercialUserID).Scan(&subCommission); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='addon_pct'`, commercialUserID).Scan(&addonCommission); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT cl.id::text, cl.source_type, ve.email, ce.email,
			cl.base_amount_cents, cl.rate_bps, cl.commission_cents, cl.period_ym, cl.accrued_at
		FROM billing.commercial_commission_ledger cl
		JOIN identity.users ve ON ve.id = cl.vet_user_id
		JOIN identity.users ce ON ce.id = cl.client_user_id
		WHERE cl.commercial_user_id=$1
		ORDER BY cl.accrued_at DESC
		LIMIT 50`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	recent := make([]CommercialLedgerRow, 0)
	for rows.Next() {
		var r CommercialLedgerRow
		if err := rows.Scan(&r.ID, &r.SourceType, &r.VetEmail, &r.ClientEmail,
			&r.BaseAmountCents, &r.RateBps, &r.CommissionCents, &r.PeriodYM, &r.AccruedAt); err != nil {
			return nil, err
		}
		recent = append(recent, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rateBps, err := s.GetCommercialRateBps(ctx)
	if err != nil {
		return nil, err
	}
	payoutHistory, err := s.ListCommercialPayoutHistory(ctx, commercialUserID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"rateBps":                     rateBps,
		"monthPeriodYm":               month,
		"monthEarnedCents":            monthEarned,
		"lifetimeEarnedCents":         lifetime,
		"subscriptionCommissionCents": subCommission,
		"addonCommissionCents":        addonCommission,
		"recentLedger":                recent,
		"payoutHistory":               payoutHistory,
	}, nil
}

func (s *Store) ResolveOpenCommercialPeriodYM(ctx context.Context, preferred string) (string, error) {
	period := preferred
	for i := 0; i < 24; i++ {
		var status string
		err := s.pool.QueryRow(ctx, `SELECT status FROM billing.commercial_payout_runs WHERE period_ym=$1`, period).Scan(&status)
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
	return "", fmt.Errorf("no_open_commercial_period")
}

// lockOpenCommercialPeriodYM locks (or creates) an open commercial payout run and returns its period.
// Callers must hold tx until ledger insert commits so close cannot race mid-accrual.
func (s *Store) lockOpenCommercialPeriodYM(ctx context.Context, tx pgx.Tx, preferred string) (string, error) {
	period := preferred
	for i := 0; i < 24; i++ {
		var status string
		err := tx.QueryRow(ctx, `
			SELECT status FROM billing.commercial_payout_runs WHERE period_ym=$1 FOR UPDATE`, period).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			if _, err := tx.Exec(ctx, `
				INSERT INTO billing.commercial_payout_runs (id, period_ym, status)
				VALUES ($1, $2, 'open')
				ON CONFLICT (period_ym) DO NOTHING`, uuid.NewString(), period); err != nil {
				return "", err
			}
			err = tx.QueryRow(ctx, `
				SELECT status FROM billing.commercial_payout_runs WHERE period_ym=$1 FOR UPDATE`, period).Scan(&status)
			if err != nil {
				return "", err
			}
		} else if err != nil {
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
	return "", fmt.Errorf("no_open_commercial_period")
}

// AccrueCommercialForSubscription accrues a flat commercial % on the pet entitlement.
// Resolves vet+commercial via practice_clients (same path as addons), not the vet ledger.
func (s *Store) AccrueCommercialForSubscription(ctx context.Context, petID string) error {
	ent, err := s.GetEntitlementByPetID(ctx, petID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if ent.Status != "active" && ent.Status != "past_due" && ent.Status != "cancelled" {
		return nil
	}

	var practiceID string
	err = s.pool.QueryRow(ctx, `
		SELECT practice_id::text FROM pets.pets WHERE id=$1`, petID).Scan(&practiceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	vetUserID, commercialUserID, err := s.resolveVetCommercial(ctx, ent.OwnerUserID, practiceID)
	if err != nil {
		return err
	}
	if commercialUserID == "" || vetUserID == "" {
		return nil
	}

	rateBps, err := s.GetCommercialRateBps(ctx)
	if err != nil {
		return err
	}
	commission := CommercialCommissionCents(ent.AmountCents, rateBps)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	period, err := s.lockOpenCommercialPeriodYM(ctx, tx, PeriodYM(time.Now()))
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.commercial_commission_ledger (
			id, commercial_user_id, vet_user_id, client_user_id, source_type, source_id,
			base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at
		) VALUES ($1,$2,$3,$4,'subscription_pct',$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (source_type, source_id) DO NOTHING`,
		uuid.NewString(), commercialUserID, vetUserID, ent.OwnerUserID, ent.ID,
		ent.AmountCents, rateBps, commission, period); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// AccrueCommercialForAddon accrues a flat commission for the commercial assigned
// to the vet of an addon buyer's practice.
func (s *Store) AccrueCommercialForAddon(ctx context.Context, addonID string) error {
	addon, err := s.GetAddonEntitlement(ctx, addonID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if addon.Status != "active" {
		return nil
	}

	var practiceID string
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,'') FROM identity.users WHERE id=$1`, addon.OwnerUserID).Scan(&practiceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if practiceID == "" {
		_ = s.pool.QueryRow(ctx, `
			SELECT practice_id::text FROM pets.pets
			WHERE owner_user_id=$1 ORDER BY created_at DESC LIMIT 1`, addon.OwnerUserID).Scan(&practiceID)
	}

	vetUserID, commercialUserID, err := s.resolveVetCommercial(ctx, addon.OwnerUserID, practiceID)
	if err != nil {
		return err
	}
	if commercialUserID == "" || vetUserID == "" {
		return nil
	}

	rateBps, err := s.GetCommercialRateBps(ctx)
	if err != nil {
		return err
	}
	commission := CommercialCommissionCents(addon.AmountCents, rateBps)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	period, err := s.lockOpenCommercialPeriodYM(ctx, tx, PeriodYM(time.Now()))
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.commercial_commission_ledger (
			id, commercial_user_id, vet_user_id, client_user_id, source_type, source_id,
			base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at
		) VALUES ($1,$2,$3,$4,'addon_pct',$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (source_type, source_id) DO NOTHING`,
		uuid.NewString(), commercialUserID, vetUserID, addon.OwnerUserID, addon.ID,
		addon.AmountCents, rateBps, commission, period); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// resolveVetCommercial finds the vet linked to a client (preferring the given
// practice) and the commercial assigned to that vet.
func (s *Store) resolveVetCommercial(ctx context.Context, clientUserID, practiceID string) (vetUserID, commercialUserID string, err error) {
	var commercial *string
	q := `
		SELECT pc.vet_user_id::text, u.assigned_commercial_id::text
		FROM practice.practice_clients pc
		JOIN identity.users u ON u.id = pc.vet_user_id
		WHERE pc.client_user_id=$1`
	args := []any{clientUserID}
	if practiceID != "" {
		q += ` AND pc.practice_id=$2`
		args = append(args, practiceID)
	}
	// Deterministic: newest client↔vet link wins.
	q += ` ORDER BY pc.created_at DESC NULLS LAST, pc.vet_user_id LIMIT 1`
	err = s.pool.QueryRow(ctx, q, args...).Scan(&vetUserID, &commercial)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	if commercial != nil {
		commercialUserID = *commercial
	}
	return vetUserID, commercialUserID, nil
}

func (s *Store) AccrueAllCommercialForActiveEntitlements(ctx context.Context) error {
	rows, err := s.pool.Query(ctx, `
		SELECT pet_id::text FROM billing.pet_entitlements
		WHERE status IN ('active','past_due','cancelled')
		ORDER BY created_at`)
	if err != nil {
		return err
	}
	petIDs := make([]string, 0)
	for rows.Next() {
		var petID string
		if err := rows.Scan(&petID); err != nil {
			rows.Close()
			return err
		}
		petIDs = append(petIDs, petID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, petID := range petIDs {
		if err := s.AccrueCommercialForSubscription(ctx, petID); err != nil {
			return err
		}
	}

	addonRows, err := s.pool.Query(ctx, `
		SELECT id::text FROM billing.addon_entitlements WHERE status='active' ORDER BY created_at`)
	if err != nil {
		return err
	}
	addonIDs := make([]string, 0)
	for addonRows.Next() {
		var id string
		if err := addonRows.Scan(&id); err != nil {
			addonRows.Close()
			return err
		}
		addonIDs = append(addonIDs, id)
	}
	addonRows.Close()
	if err := addonRows.Err(); err != nil {
		return err
	}
	for _, id := range addonIDs {
		if err := s.AccrueCommercialForAddon(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

type CommercialPayoutLine struct {
	ID                   string `json:"id"`
	RunID                string `json:"runId"`
	CommercialUserID     string `json:"commercialUserId"`
	CommercialEmail      string `json:"commercialEmail"`
	CommercialFullName   string `json:"commercialFullName"`
	LedgerCount          int    `json:"ledgerCount"`
	AmountCents          int    `json:"amountCents"`
	Status               string `json:"status"`
	PayoutIBAN           string `json:"payoutIban"`
	PayoutBIC            string `json:"payoutBic"`
	PayoutAccountHolder  string `json:"payoutAccountHolder"`
}

type CommercialPayoutProfile struct {
	IBAN          string `json:"iban"`
	BIC           string `json:"bic"`
	AccountHolder string `json:"accountHolder"`
}

func (s *Store) GetCommercialPayoutProfile(ctx context.Context, userID string) (CommercialPayoutProfile, error) {
	var p CommercialPayoutProfile
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(payout_iban,''), COALESCE(payout_bic,''), COALESCE(payout_account_holder,'')
		FROM identity.users WHERE id=$1`, userID).Scan(&p.IBAN, &p.BIC, &p.AccountHolder)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialPayoutProfile{}, ErrNotFound
	}
	return p, err
}

func (s *Store) UpdateCommercialPayoutProfile(ctx context.Context, userID string, p CommercialPayoutProfile) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users
		SET payout_iban=$2, payout_bic=$3, payout_account_holder=$4
		WHERE id=$1 AND role='commercial'`,
		userID, p.IBAN, p.BIC, p.AccountHolder)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetOrCreateCommercialPayoutRun(ctx context.Context, periodYM string) (PayoutRun, error) {
	var r PayoutRun
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, period_ym, status, closed_at, paid_at, COALESCE(note,''), created_at
		FROM billing.commercial_payout_runs WHERE period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt)
	if err == nil {
		return r, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PayoutRun{}, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO billing.commercial_payout_runs (id, period_ym, status)
		VALUES ($1, $2, 'open')
		ON CONFLICT (period_ym) DO NOTHING`, uuid.NewString(), periodYM)
	if err != nil {
		return PayoutRun{}, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, period_ym, status, closed_at, paid_at, COALESCE(note,''), created_at
		FROM billing.commercial_payout_runs WHERE period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt)
	return r, err
}

func (s *Store) ListCommercialPayoutRuns(ctx context.Context) ([]PayoutRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id::text, r.period_ym, r.status, r.closed_at, r.paid_at, COALESCE(r.note,''), r.created_at,
			COALESCE((SELECT SUM(amount_cents) FROM billing.commercial_payout_lines pl WHERE pl.run_id=r.id), 0),
			COALESCE((SELECT COUNT(*) FROM billing.commercial_payout_lines pl WHERE pl.run_id=r.id), 0)
		FROM billing.commercial_payout_runs r
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

func (s *Store) GetCommercialPayoutRunByPeriod(ctx context.Context, periodYM string) (PayoutRun, error) {
	var r PayoutRun
	err := s.pool.QueryRow(ctx, `
		SELECT r.id::text, r.period_ym, r.status, r.closed_at, r.paid_at, COALESCE(r.note,''), r.created_at,
			COALESCE((SELECT SUM(amount_cents) FROM billing.commercial_payout_lines pl WHERE pl.run_id=r.id), 0),
			COALESCE((SELECT COUNT(*) FROM billing.commercial_payout_lines pl WHERE pl.run_id=r.id), 0)
		FROM billing.commercial_payout_runs r WHERE r.period_ym=$1`, periodYM).Scan(
		&r.ID, &r.PeriodYM, &r.Status, &r.ClosedAt, &r.PaidAt, &r.Note, &r.CreatedAt, &r.TotalCents, &r.LineCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return PayoutRun{}, ErrNotFound
	}
	return r, err
}

func (s *Store) PreviewCommercialPeriodCommissions(ctx context.Context, periodYM string) ([]CommercialPayoutLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cl.commercial_user_id::text, u.email, u.full_name,
			COUNT(*)::int,
			COALESCE(SUM(cl.commission_cents),0)::int,
			COALESCE(u.payout_iban,''), COALESCE(u.payout_bic,''), COALESCE(u.payout_account_holder,'')
		FROM billing.commercial_commission_ledger cl
		JOIN identity.users u ON u.id = cl.commercial_user_id
		WHERE cl.period_ym=$1
		GROUP BY cl.commercial_user_id, u.email, u.full_name, u.payout_iban, u.payout_bic, u.payout_account_holder
		ORDER BY SUM(cl.commission_cents) DESC`, periodYM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CommercialPayoutLine
	for rows.Next() {
		var l CommercialPayoutLine
		l.Status = "pending"
		if err := rows.Scan(&l.CommercialUserID, &l.CommercialEmail, &l.CommercialFullName,
			&l.LedgerCount, &l.AmountCents, &l.PayoutIBAN, &l.PayoutBIC, &l.PayoutAccountHolder); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) ListCommercialPayoutLines(ctx context.Context, runID string) ([]CommercialPayoutLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pl.id::text, pl.run_id::text, pl.commercial_user_id::text, u.email, u.full_name,
			pl.ledger_count, pl.amount_cents, pl.status,
			COALESCE(u.payout_iban,''), COALESCE(u.payout_bic,''), COALESCE(u.payout_account_holder,'')
		FROM billing.commercial_payout_lines pl
		JOIN identity.users u ON u.id = pl.commercial_user_id
		WHERE pl.run_id=$1
		ORDER BY pl.amount_cents DESC, u.full_name`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CommercialPayoutLine
	for rows.Next() {
		var l CommercialPayoutLine
		if err := rows.Scan(&l.ID, &l.RunID, &l.CommercialUserID, &l.CommercialEmail, &l.CommercialFullName,
			&l.LedgerCount, &l.AmountCents, &l.Status,
			&l.PayoutIBAN, &l.PayoutBIC, &l.PayoutAccountHolder); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) AdminCommercialCommissionPeriodDetail(ctx context.Context, periodYM string) (map[string]any, error) {
	run, err := s.GetCommercialPayoutRunByPeriod(ctx, periodYM)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, ErrNotFound) {
		run = PayoutRun{PeriodYM: periodYM, Status: "open"}
	}
	var lines []CommercialPayoutLine
	total := 0
	if run.ID != "" && (run.Status == "closed" || run.Status == "paid") {
		lines, err = s.ListCommercialPayoutLines(ctx, run.ID)
		if err != nil {
			return nil, err
		}
		total = run.TotalCents
	} else {
		lines, err = s.PreviewCommercialPeriodCommissions(ctx, periodYM)
		if err != nil {
			return nil, err
		}
		for _, l := range lines {
			total += l.AmountCents
		}
	}
	if lines == nil {
		lines = []CommercialPayoutLine{}
	}
	return map[string]any{
		"run":        run,
		"lines":      lines,
		"totalCents": total,
		"lineCount":  len(lines),
	}, nil
}

func (s *Store) CloseCommercialPayoutRun(ctx context.Context, periodYM string) (PayoutRun, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PayoutRun{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.commercial_payout_runs (id, period_ym, status)
		VALUES ($1, $2, 'open')
		ON CONFLICT (period_ym) DO NOTHING`, uuid.NewString(), periodYM); err != nil {
		return PayoutRun{}, err
	}

	var runID, status string
	if err := tx.QueryRow(ctx, `
		SELECT id::text, status FROM billing.commercial_payout_runs WHERE period_ym=$1 FOR UPDATE`,
		periodYM).Scan(&runID, &status); err != nil {
		return PayoutRun{}, err
	}
	if status != "open" {
		return PayoutRun{}, ErrPayoutNotOpen
	}

	if _, err := tx.Exec(ctx, `DELETE FROM billing.commercial_payout_lines WHERE run_id=$1`, runID); err != nil {
		return PayoutRun{}, err
	}
	rows, err := tx.Query(ctx, `
		SELECT commercial_user_id::text, COUNT(*)::int, COALESCE(SUM(commission_cents),0)::int
		FROM billing.commercial_commission_ledger
		WHERE period_ym=$1
		GROUP BY commercial_user_id`, periodYM)
	if err != nil {
		return PayoutRun{}, err
	}
	var aggs []struct {
		userID      string
		ledgerCount int
		amount      int
	}
	for rows.Next() {
		var a struct {
			userID      string
			ledgerCount int
			amount      int
		}
		if err := rows.Scan(&a.userID, &a.ledgerCount, &a.amount); err != nil {
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
			INSERT INTO billing.commercial_payout_lines (id, run_id, commercial_user_id, ledger_count, amount_cents, status)
			VALUES ($1,$2,$3,$4,$5,'pending')`,
			uuid.NewString(), runID, a.userID, a.ledgerCount, a.amount); err != nil {
			return PayoutRun{}, err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE billing.commercial_payout_runs SET status='closed', closed_at=NOW()
		WHERE id=$1 AND status='open'`, runID); err != nil {
		return PayoutRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PayoutRun{}, err
	}
	return s.GetCommercialPayoutRunByPeriod(ctx, periodYM)
}

func (s *Store) MarkCommercialPayoutRunPaid(ctx context.Context, periodYM, note string) (PayoutRun, error) {
	run, err := s.GetCommercialPayoutRunByPeriod(ctx, periodYM)
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
		UPDATE billing.commercial_payout_runs SET status='paid', paid_at=COALESCE(paid_at, NOW()), note=$2 WHERE id=$1`,
		run.ID, note); err != nil {
		return PayoutRun{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE billing.commercial_payout_lines SET status='paid' WHERE run_id=$1`, run.ID); err != nil {
		return PayoutRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PayoutRun{}, err
	}
	return s.GetCommercialPayoutRunByPeriod(ctx, periodYM)
}

func (s *Store) ListCommercialPayoutHistory(ctx context.Context, commercialUserID string) ([]PayoutLineHistory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.period_ym, pl.amount_cents, pl.status, r.status, r.paid_at
		FROM billing.commercial_payout_lines pl
		JOIN billing.commercial_payout_runs r ON r.id = pl.run_id
		WHERE pl.commercial_user_id=$1
		ORDER BY r.period_ym DESC
		LIMIT 24`, commercialUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PayoutLineHistory
	for rows.Next() {
		var h PayoutLineHistory
		if err := rows.Scan(&h.PeriodYM, &h.AmountCents, &h.Status, &h.RunStatus, &h.PaidAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
