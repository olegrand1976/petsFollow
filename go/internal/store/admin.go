package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type AdminMetrics struct {
	TotalRevenueCents     int            `json:"totalRevenueCents"`
	PeriodRevenueCents    int            `json:"periodRevenueCents"`
	MRRCents              int            `json:"mrrCents"`
	UserCount             int            `json:"userCount"`
	PetCount              int            `json:"petCount"`
	ActiveEntitlements    int            `json:"activeEntitlements"`
	PendingPayments       int            `json:"pendingPayments"`
	PastDueCount          int            `json:"pastDueCount"`
	ConversionRatePercent float64        `json:"conversionRatePercent"`
	PlanBreakdown         map[string]int `json:"planBreakdown"`
	ModeBreakdown         map[string]int `json:"modeBreakdown"`
}

type AdminUserRow struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"fullName"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	PetCount     int       `json:"petCount"`
	PaymentLabel string    `json:"paymentLabel"`
}

type AdminPaymentRow struct {
	ID                string     `json:"id"`
	CreatedAt         time.Time  `json:"createdAt"`
	ClientEmail       string     `json:"clientEmail"`
	ClientName        string     `json:"clientName"`
	PetName           string     `json:"petName"`
	PlanCode          string     `json:"planCode"`
	BillingMode       string     `json:"billingMode"`
	AmountCents       int        `json:"amountCents"`
	Status            string     `json:"status"`
	StripeSessionID   string     `json:"stripeSessionId,omitempty"`
	StripeSubID       string     `json:"stripeSubscriptionId,omitempty"`
	ValidUntil        *time.Time `json:"validUntil,omitempty"`
}

func (s *Store) AdminMetricsOverview(ctx context.Context, from, to time.Time) (AdminMetrics, error) {
	var m AdminMetrics
	m.PlanBreakdown = map[string]int{}
	m.ModeBreakdown = map[string]int{}

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents) FILTER (WHERE status IN ('active','past_due','cancelled')), 0)::int
		FROM billing.pet_entitlements`).Scan(&m.TotalRevenueCents)
	if err != nil {
		return m, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents) FILTER (WHERE status IN ('active','past_due','cancelled') AND created_at >= $1 AND created_at < $2), 0)::int
		FROM billing.pet_entitlements`, from, to).Scan(&m.PeriodRevenueCents)
	if err != nil {
		return m, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(
			CASE plan_code
				WHEN 'annual' THEN amount_cents
				WHEN 'triennial' THEN amount_cents / 3
				WHEN 'quinquennial' THEN amount_cents / 5
				ELSE 0
			END
		) FILTER (WHERE billing_mode='subscription' AND status='active'), 0)::int
		FROM billing.pet_entitlements`).Scan(&m.MRRCents)
	if err != nil {
		return m, err
	}

	err = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM identity.users`).Scan(&m.UserCount)
	if err != nil {
		return m, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM pets.pets`).Scan(&m.PetCount)
	if err != nil {
		return m, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM billing.pet_entitlements WHERE status='active'`).Scan(&m.ActiveEntitlements)
	if err != nil {
		return m, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM billing.pet_entitlements WHERE status='pending'`).Scan(&m.PendingPayments)
	if err != nil {
		return m, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM billing.pet_entitlements WHERE status='past_due'`).Scan(&m.PastDueCount)
	if err != nil {
		return m, err
	}

	var paid int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM billing.pet_entitlements WHERE status IN ('active','past_due','cancelled')`).Scan(&paid)
	if m.PetCount > 0 {
		m.ConversionRatePercent = float64(paid) / float64(m.PetCount) * 100
	}

	rows, err := s.pool.Query(ctx, `SELECT plan_code, COUNT(*)::int FROM billing.pet_entitlements GROUP BY plan_code`)
	if err != nil {
		return m, err
	}
	defer rows.Close()
	for rows.Next() {
		var plan string
		var count int
		if err := rows.Scan(&plan, &count); err != nil {
			return m, err
		}
		m.PlanBreakdown[plan] = count
	}

	rows2, err := s.pool.Query(ctx, `SELECT billing_mode, COUNT(*)::int FROM billing.pet_entitlements GROUP BY billing_mode`)
	if err != nil {
		return m, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var mode string
		var count int
		if err := rows2.Scan(&mode, &count); err != nil {
			return m, err
		}
		m.ModeBreakdown[mode] = count
	}

	return m, nil
}

func (s *Store) ListAdminUsers(ctx context.Context, role string, from, to time.Time, limit, offset int) ([]AdminUserRow, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if role != "" {
		rows, err = s.pool.Query(ctx, `
			SELECT u.id::text, u.email, u.full_name, u.role, u.created_at,
				COUNT(p.id)::int,
				COALESCE(
					(SELECT CASE
						WHEN COUNT(*) FILTER (WHERE e.status='pending') > 0 THEN 'pending'
						WHEN COUNT(*) FILTER (WHERE e.status='active') > 0 THEN 'active'
						ELSE 'none'
					END FROM billing.pet_entitlements e JOIN pets.pets pp ON pp.id=e.pet_id WHERE pp.owner_user_id=u.id),
					'none')
			FROM identity.users u
			LEFT JOIN pets.pets p ON p.owner_user_id = u.id
			WHERE u.created_at >= $1 AND u.created_at < $2 AND u.role = $3
			GROUP BY u.id ORDER BY u.created_at DESC LIMIT $4 OFFSET $5`, from, to, role, limit, offset)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT u.id::text, u.email, u.full_name, u.role, u.created_at,
				COUNT(p.id)::int,
				COALESCE(
					(SELECT CASE
						WHEN COUNT(*) FILTER (WHERE e.status='pending') > 0 THEN 'pending'
						WHEN COUNT(*) FILTER (WHERE e.status='active') > 0 THEN 'active'
						ELSE 'none'
					END FROM billing.pet_entitlements e JOIN pets.pets pp ON pp.id=e.pet_id WHERE pp.owner_user_id=u.id),
					'none')
			FROM identity.users u
			LEFT JOIN pets.pets p ON p.owner_user_id = u.id
			WHERE u.created_at >= $1 AND u.created_at < $2
			GROUP BY u.id ORDER BY u.created_at DESC LIMIT $3 OFFSET $4`, from, to, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AdminUserRow
	for rows.Next() {
		var r AdminUserRow
		if err := rows.Scan(&r.ID, &r.Email, &r.FullName, &r.Role, &r.CreatedAt, &r.PetCount, &r.PaymentLabel); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListAdminPayments(ctx context.Context, from, to time.Time, status string, limit, offset int) ([]AdminPaymentRow, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = s.pool.Query(ctx, `
			SELECT e.id::text, e.created_at, u.email, u.full_name, p.name, e.plan_code, e.billing_mode,
				e.amount_cents, e.status, COALESCE(e.stripe_checkout_session_id,''), COALESCE(e.stripe_subscription_id,''), e.valid_until
			FROM billing.pet_entitlements e
			JOIN identity.users u ON u.id = e.owner_user_id
			JOIN pets.pets p ON p.id = e.pet_id
			WHERE e.created_at >= $1 AND e.created_at < $2 AND e.status = $3
			ORDER BY e.created_at DESC LIMIT $4 OFFSET $5`, from, to, status, limit, offset)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT e.id::text, e.created_at, u.email, u.full_name, p.name, e.plan_code, e.billing_mode,
				e.amount_cents, e.status, COALESCE(e.stripe_checkout_session_id,''), COALESCE(e.stripe_subscription_id,''), e.valid_until
			FROM billing.pet_entitlements e
			JOIN identity.users u ON u.id = e.owner_user_id
			JOIN pets.pets p ON p.id = e.pet_id
			WHERE e.created_at >= $1 AND e.created_at < $2
			ORDER BY e.created_at DESC LIMIT $3 OFFSET $4`, from, to, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AdminPaymentRow
	for rows.Next() {
		var r AdminPaymentRow
		if err := rows.Scan(&r.ID, &r.CreatedAt, &r.ClientEmail, &r.ClientName, &r.PetName, &r.PlanCode, &r.BillingMode,
			&r.AmountCents, &r.Status, &r.StripeSessionID, &r.StripeSubID, &r.ValidUntil); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
