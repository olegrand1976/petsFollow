package store

import (
	"context"
	"errors"
)

// Household / multi-pet addon rules.
const (
	FamilyMinPets = 2
	KennelMinPets = 6

	familyAddonCode = "family"
	kennelAddonCode = "kennel"
)

// Discount bps on pet-plan checkout when household addon is active and owner
// already has ≥1 other active pet entitlement.
const (
	FamilyPetDiscountBps = 1000 // −10 %
	KennelPetDiscountBps = 1500 // −15 %
)

var (
	ErrFamilyRequiresTwoPets = errors.New("family_requires_two_pets")
	ErrKennelRequiresSixPets = errors.New("kennel_requires_six_pets")
	ErrHouseholdExclusive    = errors.New("household_exclusive")
	ErrAddonAlreadyActive    = errors.New("addon_already_active")
)

// HouseholdDiscountBps returns the pet-plan discount for an owner (0 if none).
func HouseholdDiscountBps(hasFamily, hasKennel bool) int {
	if hasKennel {
		return KennelPetDiscountBps
	}
	if hasFamily {
		return FamilyPetDiscountBps
	}
	return 0
}

// ApplyDiscountCents applies bps discount to a TTC amount (rounded down).
func ApplyDiscountCents(amountCents, discountBps int) int {
	if amountCents <= 0 || discountBps <= 0 {
		return amountCents
	}
	if discountBps > 10000 {
		discountBps = 10000
	}
	return amountCents - (amountCents * discountBps / 10000)
}

// CheckFamilyPurchasePetCount validates n ≥ 2 for buying Family.
func CheckFamilyPurchasePetCount(n int) error {
	if n < FamilyMinPets {
		return ErrFamilyRequiresTwoPets
	}
	return nil
}

// CheckKennelPurchasePetCount validates n ≥ 6 for buying Kennel.
func CheckKennelPurchasePetCount(n int) error {
	if n < KennelMinPets {
		return ErrKennelRequiresSixPets
	}
	return nil
}

func (s *Store) AssertFamilyPurchaseEligible(ctx context.Context, ownerUserID string) error {
	if ok, err := s.HasActiveOrPendingAddon(ctx, ownerUserID, kennelAddonCode); err != nil {
		return err
	} else if ok {
		return ErrHouseholdExclusive
	}
	if ok, err := s.HasActiveOrPendingAddon(ctx, ownerUserID, familyAddonCode); err != nil {
		return err
	} else if ok {
		return ErrAddonAlreadyActive
	}
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	return CheckFamilyPurchasePetCount(n)
}

func (s *Store) AssertKennelPurchaseEligible(ctx context.Context, ownerUserID string) error {
	if ok, err := s.HasActiveOrPendingAddon(ctx, ownerUserID, kennelAddonCode); err != nil {
		return err
	} else if ok {
		return ErrAddonAlreadyActive
	}
	// Family pending checkout: refuse Kennel to avoid double charge / post-pay cancel.
	// Family active: allowed (upgrade — Family cancelled on Kennel activation).
	pendingFamily, err := s.hasPendingAddon(ctx, ownerUserID, familyAddonCode)
	if err != nil {
		return err
	}
	if pendingFamily {
		return ErrHouseholdExclusive
	}
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	return CheckKennelPurchasePetCount(n)
}

// AssertAddonNotAlreadyOwned blocks Care+ / Horse repurchase (lifetime one-time).
func (s *Store) AssertAddonNotAlreadyOwned(ctx context.Context, ownerUserID, addonCode string) error {
	ok, err := s.HasActiveOrPendingAddon(ctx, ownerUserID, addonCode)
	if err != nil {
		return err
	}
	if ok {
		return ErrAddonAlreadyActive
	}
	return nil
}

func (s *Store) hasPendingAddon(ctx context.Context, ownerUserID, addonCode string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.addon_entitlements
			WHERE owner_user_id=$1 AND addon_code=$2
				AND status='pending' AND created_at > NOW() - INTERVAL '24 hours'
		)`, ownerUserID, addonCode).Scan(&exists)
	return exists, err
}

// HasHouseholdAddon is true when Family or Kennel is active (shared foyer privileges).
func (s *Store) HasHouseholdAddon(ctx context.Context, ownerUserID string) (bool, error) {
	hasKennel, err := s.HasActiveAddon(ctx, ownerUserID, kennelAddonCode)
	if err != nil || hasKennel {
		return hasKennel, err
	}
	return s.HasActiveAddon(ctx, ownerUserID, familyAddonCode)
}

// DeactivateHouseholdAddon marks an active/pending/past_due household addon cancelled.
func (s *Store) DeactivateHouseholdAddon(ctx context.Context, ownerUserID, addonCode string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE billing.addon_entitlements
		SET status='cancelled', updated_at=NOW()
		WHERE owner_user_id=$1 AND addon_code=$2
			AND status IN ('active','pending','past_due')`, ownerUserID, addonCode)
	return err
}

// CountOtherActivePetEntitlements counts active paid pets for owner excluding petID.
func (s *Store) CountOtherActivePetEntitlements(ctx context.Context, ownerUserID, excludePetID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM billing.pet_entitlements e
		WHERE e.owner_user_id=$1
			AND e.pet_id <> $2
			AND e.status = 'active'
			AND (e.valid_until IS NULL OR e.valid_until > NOW())`,
		ownerUserID, excludePetID).Scan(&n)
	return n, err
}

// ResolvePetCheckoutAmount returns catalogue TTC (household discounts removed).
func (s *Store) ResolvePetCheckoutAmount(ctx context.Context, ownerUserID, petID string, catalogueCents int) (payCents int, discountBps int, err error) {
	_ = ctx
	_ = ownerUserID
	_ = petID
	return catalogueCents, 0, nil
}
