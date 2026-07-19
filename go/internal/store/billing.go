package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Entitlement struct {
	ID                    string     `json:"id"`
	PetID                 string     `json:"petId"`
	OwnerUserID           string     `json:"ownerUserId"`
	PlanCode              string     `json:"planCode"`
	BillingMode           string     `json:"billingMode"`
	Status                string     `json:"status"`
	AmountCents           int        `json:"amountCents"`
	Currency              string     `json:"currency"`
	ValidFrom             *time.Time `json:"validFrom,omitempty"`
	ValidUntil            *time.Time `json:"validUntil,omitempty"`
	StripeCheckoutSession string     `json:"stripeCheckoutSessionId,omitempty"`
	StripePaymentIntent   string     `json:"stripePaymentIntentId,omitempty"`
	StripeSubscription    string     `json:"stripeSubscriptionId,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
}

type ActivateEntitlementParams struct {
	PetID                 string
	Status                string
	ValidFrom             time.Time
	ValidUntil            time.Time
	StripeCheckoutSession string
	StripePaymentIntent   string
	StripeSubscription    string
}

func (e Entitlement) AllowsAccess() bool {
	switch e.Status {
	case "active", "past_due", "cancelled":
		return true
	default:
		return false
	}
}

func (s *Store) CreateEntitlement(ctx context.Context, petID, ownerID, plan, mode string, amountCents int) (Entitlement, error) {
	id := uuid.NewString()
	var e Entitlement
	err := s.pool.QueryRow(ctx, `
		INSERT INTO billing.pet_entitlements (id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,'eur')
		RETURNING id::text, pet_id::text, owner_user_id::text, plan_code, billing_mode, status, amount_cents, currency, created_at`,
		id, petID, ownerID, plan, mode, amountCents,
	).Scan(&e.ID, &e.PetID, &e.OwnerUserID, &e.PlanCode, &e.BillingMode, &e.Status, &e.AmountCents, &e.Currency, &e.CreatedAt)
	return e, err
}

func (s *Store) GetEntitlementByPetID(ctx context.Context, petID string) (Entitlement, error) {
	var e Entitlement
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, plan_code, billing_mode, status, amount_cents, currency,
			valid_from, valid_until, COALESCE(stripe_checkout_session_id,''), COALESCE(stripe_payment_intent_id,''),
			COALESCE(stripe_subscription_id,''), created_at
		FROM billing.pet_entitlements WHERE pet_id=$1`, petID,
	).Scan(&e.ID, &e.PetID, &e.OwnerUserID, &e.PlanCode, &e.BillingMode, &e.Status, &e.AmountCents, &e.Currency,
		&e.ValidFrom, &e.ValidUntil, &e.StripeCheckoutSession, &e.StripePaymentIntent, &e.StripeSubscription, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entitlement{}, ErrNotFound
	}
	return e, err
}

func (s *Store) GetEntitlementBySubscriptionID(ctx context.Context, subID string) (Entitlement, error) {
	var e Entitlement
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, plan_code, billing_mode, status, amount_cents, currency,
			valid_from, valid_until, COALESCE(stripe_checkout_session_id,''), COALESCE(stripe_payment_intent_id,''),
			COALESCE(stripe_subscription_id,''), created_at
		FROM billing.pet_entitlements WHERE stripe_subscription_id=$1`, subID,
	).Scan(&e.ID, &e.PetID, &e.OwnerUserID, &e.PlanCode, &e.BillingMode, &e.Status, &e.AmountCents, &e.Currency,
		&e.ValidFrom, &e.ValidUntil, &e.StripeCheckoutSession, &e.StripePaymentIntent, &e.StripeSubscription, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entitlement{}, ErrNotFound
	}
	return e, err
}

func (s *Store) SetEntitlementAmountCents(ctx context.Context, petID string, amountCents int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE billing.pet_entitlements SET amount_cents=$2, updated_at=NOW()
		WHERE pet_id=$1`, petID, amountCents)
	return err
}

func (s *Store) SetEntitlementCheckoutSession(ctx context.Context, petID, sessionID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE billing.pet_entitlements SET stripe_checkout_session_id=$2, updated_at=NOW() WHERE pet_id=$1`, petID, sessionID)
	return err
}

func (s *Store) ActivateEntitlement(ctx context.Context, p ActivateEntitlementParams) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		UPDATE billing.pet_entitlements SET
			status=$2, valid_from=$3, valid_until=$4,
			stripe_checkout_session_id=COALESCE(NULLIF($5,''), stripe_checkout_session_id),
			stripe_payment_intent_id=COALESCE(NULLIF($6,''), stripe_payment_intent_id),
			stripe_subscription_id=COALESCE(NULLIF($7,''), stripe_subscription_id),
			updated_at=NOW()
		WHERE pet_id=$1`, p.PetID, p.Status, p.ValidFrom, p.ValidUntil, p.StripeCheckoutSession, p.StripePaymentIntent, p.StripeSubscription)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE pets.pets SET payment_status='active', updated_at=NOW() WHERE id=$1`, p.PetID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateEntitlementStatus(ctx context.Context, petID, status string) error {
	paymentStatus := "active"
	switch status {
	case "expired", "pending":
		paymentStatus = status
	case "past_due":
		paymentStatus = "active"
	case "cancelled":
		paymentStatus = "expired"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE billing.pet_entitlements SET status=$2, updated_at=NOW() WHERE pet_id=$1`, petID, status); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE pets.pets SET payment_status=$2, updated_at=NOW() WHERE id=$1`, petID, paymentStatus); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) UpsertStripeCustomer(ctx context.Context, userID, stripeCustomerID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO billing.stripe_customers (user_id, stripe_customer_id) VALUES ($1,$2)
		ON CONFLICT (user_id) DO UPDATE SET stripe_customer_id=EXCLUDED.stripe_customer_id`, userID, stripeCustomerID)
	return err
}

func (s *Store) GetStripeCustomerID(ctx context.Context, userID string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT stripe_customer_id FROM billing.stripe_customers WHERE user_id=$1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) IsStripeEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM billing.stripe_events WHERE event_id=$1)`, eventID).Scan(&exists)
	return exists, err
}

func (s *Store) RecordStripeEvent(ctx context.Context, eventID, eventType string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO billing.stripe_events (event_id, event_type) VALUES ($1,$2) ON CONFLICT DO NOTHING`, eventID, eventType)
	return err
}

func (s *Store) HasActiveEntitlement(ctx context.Context, petID string) (bool, error) {
	ent, err := s.GetEntitlementByPetID(ctx, petID)
	if err != nil {
		return false, err
	}
	if !ent.AllowsAccess() {
		return false, nil
	}
	if ent.ValidUntil != nil && time.Now().After(*ent.ValidUntil) {
		return false, nil
	}
	return ent.Status == "active" || ent.Status == "past_due" || ent.Status == "cancelled", nil
}

func (s *Store) SetPetPaymentStatus(ctx context.Context, petID, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE pets.pets SET payment_status=$2, updated_at=NOW() WHERE id=$1`, petID, status)
	return err
}
