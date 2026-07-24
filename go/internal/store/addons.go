package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AddonEntitlement struct {
	ID                    string     `json:"id"`
	OwnerUserID           string     `json:"ownerUserId"`
	AddonCode             string     `json:"addonCode"`
	Status                string     `json:"status"`
	AmountCents           int        `json:"amountCents"`
	Currency              string     `json:"currency"`
	ValidFrom             *time.Time `json:"validFrom,omitempty"`
	ValidUntil            *time.Time `json:"validUntil,omitempty"`
	StripeCheckoutSession string     `json:"stripeCheckoutSessionId,omitempty"`
	StripePaymentIntent   string     `json:"stripePaymentIntentId,omitempty"`
	StripeSubscriptionID  string     `json:"stripeSubscriptionId,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
}

const addonSelectCols = `
	id::text, owner_user_id::text, addon_code, status, amount_cents, currency,
	valid_from, valid_until, COALESCE(stripe_checkout_session_id,''), COALESCE(stripe_payment_intent_id,''),
	COALESCE(stripe_subscription_id,''), created_at`

func scanAddon(row pgx.Row) (AddonEntitlement, error) {
	var a AddonEntitlement
	err := row.Scan(&a.ID, &a.OwnerUserID, &a.AddonCode, &a.Status, &a.AmountCents, &a.Currency,
		&a.ValidFrom, &a.ValidUntil, &a.StripeCheckoutSession, &a.StripePaymentIntent,
		&a.StripeSubscriptionID, &a.CreatedAt)
	return a, err
}

func (s *Store) CreateAddonEntitlement(ctx context.Context, ownerUserID, addonCode string, amountCents int) (AddonEntitlement, error) {
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO billing.addon_entitlements (id, owner_user_id, addon_code, status, amount_cents, currency)
		VALUES ($1,$2,$3,'pending',$4,'eur')
		RETURNING `+addonSelectCols, id, ownerUserID, addonCode, amountCents)
	return scanAddon(row)
}

func (s *Store) GetAddonEntitlement(ctx context.Context, id string) (AddonEntitlement, error) {
	a, err := scanAddon(s.pool.QueryRow(ctx, `
		SELECT `+addonSelectCols+` FROM billing.addon_entitlements WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return AddonEntitlement{}, ErrNotFound
	}
	return a, err
}

func (s *Store) GetAddonBySubscriptionID(ctx context.Context, subscriptionID string) (AddonEntitlement, error) {
	if subscriptionID == "" {
		return AddonEntitlement{}, ErrNotFound
	}
	a, err := scanAddon(s.pool.QueryRow(ctx, `
		SELECT `+addonSelectCols+` FROM billing.addon_entitlements
		WHERE stripe_subscription_id=$1
		ORDER BY created_at DESC LIMIT 1`, subscriptionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AddonEntitlement{}, ErrNotFound
	}
	return a, err
}

// ActivateAddonEntitlement activates a pending addon and stores Stripe IDs.
// validUntil nil means lifetime (no expiry).
func (s *Store) ActivateAddonEntitlement(ctx context.Context, id string, validFrom time.Time, validUntil *time.Time, sessionID, paymentIntent, subscriptionID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE billing.addon_entitlements SET
			status='active', valid_from=$2, valid_until=$3,
			stripe_checkout_session_id=COALESCE(NULLIF($4,''), stripe_checkout_session_id),
			stripe_payment_intent_id=COALESCE(NULLIF($5,''), stripe_payment_intent_id),
			stripe_subscription_id=COALESCE(NULLIF($6,''), stripe_subscription_id),
			updated_at=NOW()
		WHERE id=$1 AND status='pending'`, id, validFrom, validUntil, sessionID, paymentIntent, subscriptionID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ExtendAddonBySubscriptionID renews an active/past_due addon. Cancelled/expired rows are ignored
// so a failed Stripe cancel cannot revive Family after a Kennel upgrade.
func (s *Store) ExtendAddonBySubscriptionID(ctx context.Context, subscriptionID string, validFrom, validUntil time.Time) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE billing.addon_entitlements SET
			status='active', valid_from=$2, valid_until=$3, updated_at=NOW()
		WHERE stripe_subscription_id=$1
			AND status IN ('active','past_due')`, subscriptionID, validFrom, validUntil)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateAddonStatusBySubscriptionID(ctx context.Context, subscriptionID, status string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE billing.addon_entitlements SET status=$2, updated_at=NOW()
		WHERE stripe_subscription_id=$1`, subscriptionID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAddonSubscriptionID returns the Stripe sub for an active/past_due/pending addon.
func (s *Store) GetAddonSubscriptionID(ctx context.Context, ownerUserID, addonCode string) (string, error) {
	var subID string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(stripe_subscription_id,'')
		FROM billing.addon_entitlements
		WHERE owner_user_id=$1 AND addon_code=$2
			AND status IN ('active','past_due','pending')
			AND COALESCE(stripe_subscription_id,'') <> ''
		ORDER BY created_at DESC LIMIT 1`, ownerUserID, addonCode).Scan(&subID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return subID, err
}

func (s *Store) ListOwnerAddons(ctx context.Context, ownerUserID string) ([]AddonEntitlement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+addonSelectCols+`
		FROM billing.addon_entitlements WHERE owner_user_id=$1 ORDER BY created_at DESC`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AddonEntitlement, 0)
	for rows.Next() {
		a, err := scanAddon(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) HasActiveAddon(ctx context.Context, ownerUserID, addonCode string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.addon_entitlements
			WHERE owner_user_id=$1 AND addon_code=$2
				AND status IN ('active','past_due')
				AND (valid_until IS NULL OR valid_until > NOW())
		)`, ownerUserID, addonCode).Scan(&exists)
	return exists, err
}

// HasActiveOrPendingAddon is true for an active/past_due entitlement or a recent pending checkout
// (24h). Used to enforce household exclusivity during the payment window.
func (s *Store) HasActiveOrPendingAddon(ctx context.Context, ownerUserID, addonCode string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.addon_entitlements
			WHERE owner_user_id=$1 AND addon_code=$2
				AND (
					(status='pending' AND created_at > NOW() - INTERVAL '24 hours')
					OR (status IN ('active','past_due') AND (valid_until IS NULL OR valid_until > NOW()))
				)
		)`, ownerUserID, addonCode).Scan(&exists)
	return exists, err
}

// CancelAddonEntitlement marks a pending/active/past_due addon as cancelled.
func (s *Store) CancelAddonEntitlement(ctx context.Context, id string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE billing.addon_entitlements
		SET status='cancelled', updated_at=NOW()
		WHERE id=$1 AND status IN ('pending','active','past_due')`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
