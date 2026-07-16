package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"golang.org/x/crypto/bcrypt"
)

// CommercialCommissionRateBps is the flat commercial commission rate (15%).
const CommercialCommissionRateBps = 1500

// CommercialCommissionCents returns the flat 15% commission on a base amount.
func CommercialCommissionCents(baseAmountCents int) int {
	return baseAmountCents * CommercialCommissionRateBps / 10000
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
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND period_ym=$2`, commercialUserID, month).Scan(&monthEarned)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1`, commercialUserID).Scan(&lifetime)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(base_amount_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='subscription_flat'`, commercialUserID).Scan(&subRevenue)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(base_amount_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='addon_pct'`, commercialUserID).Scan(&addonRevenue)

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
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND period_ym=$2`, commercialUserID, month).Scan(&monthEarned)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1`, commercialUserID).Scan(&lifetime)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='subscription_flat'`, commercialUserID).Scan(&subCommission)
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(commission_cents),0)::int FROM billing.commercial_commission_ledger
		WHERE commercial_user_id=$1 AND source_type='addon_pct'`, commercialUserID).Scan(&addonCommission)

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

	return map[string]any{
		"rateBps":                    CommercialCommissionRateBps,
		"monthPeriodYm":              month,
		"monthEarnedCents":           monthEarned,
		"lifetimeEarnedCents":        lifetime,
		"subscriptionCommissionCents": subCommission,
		"addonCommissionCents":       addonCommission,
		"recentLedger":               recent,
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
	return period, nil
}

// AccrueCommercialForSubscription accrues a flat commission for the commercial
// assigned to the vet of a subscription pet, once entitlement is active.
func (s *Store) AccrueCommercialForSubscription(ctx context.Context, petID string) error {
	ent, err := s.GetEntitlementByPetID(ctx, petID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if ent.BillingMode != "subscription" {
		return nil
	}
	if ent.Status != "active" && ent.Status != "past_due" && ent.Status != "cancelled" {
		return nil
	}

	var ownerUserID, practiceID string
	err = s.pool.QueryRow(ctx, `
		SELECT owner_user_id::text, practice_id::text FROM pets.pets WHERE id=$1`, petID).Scan(&ownerUserID, &practiceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	vetUserID, commercialUserID, err := s.resolveVetCommercial(ctx, ownerUserID, practiceID)
	if err != nil {
		return err
	}
	if commercialUserID == "" {
		return nil
	}

	period, err := s.ResolveOpenCommercialPeriodYM(ctx, PeriodYM(time.Now()))
	if err != nil {
		return err
	}
	commission := CommercialCommissionCents(ent.AmountCents)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO billing.commercial_commission_ledger (
			id, commercial_user_id, vet_user_id, client_user_id, source_type, source_id,
			base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at
		) VALUES ($1,$2,$3,$4,'subscription_flat',$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (source_type, source_id) DO NOTHING`,
		uuid.NewString(), commercialUserID, vetUserID, ownerUserID, ent.ID,
		ent.AmountCents, CommercialCommissionRateBps, commission, period)
	return err
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

	vetUserID, commercialUserID, err := s.resolveVetCommercial(ctx, addon.OwnerUserID, practiceID)
	if err != nil {
		return err
	}
	if commercialUserID == "" {
		return nil
	}

	period, err := s.ResolveOpenCommercialPeriodYM(ctx, PeriodYM(time.Now()))
	if err != nil {
		return err
	}
	commission := CommercialCommissionCents(addon.AmountCents)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO billing.commercial_commission_ledger (
			id, commercial_user_id, vet_user_id, client_user_id, source_type, source_id,
			base_amount_cents, rate_bps, commission_cents, period_ym, accrued_at
		) VALUES ($1,$2,$3,$4,'addon_pct',$5,$6,$7,$8,$9,NOW())
		ON CONFLICT (source_type, source_id) DO NOTHING`,
		uuid.NewString(), commercialUserID, vetUserID, addon.OwnerUserID, addon.ID,
		addon.AmountCents, CommercialCommissionRateBps, commission, period)
	return err
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
	q += ` LIMIT 1`
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
		WHERE billing_mode='subscription' AND status IN ('active','past_due','cancelled')
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
