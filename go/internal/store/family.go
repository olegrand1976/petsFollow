package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Family pack rules (owner-scoped addon).
const (
	FamilyMinPets   = 2
	FamilyMaxPets   = 3
	familyAddonCode = "family"
)

var (
	ErrFamilyPetLimit        = errors.New("family_pet_limit")
	ErrFamilyRequiresTwoPets = errors.New("family_requires_two_pets")
)

type HouseholdCareItem struct {
	ID        string    `json:"id"`
	PetID     string    `json:"petId"`
	PetName   string    `json:"petName"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	DueAt     time.Time `json:"dueAt"`
	Status    string    `json:"status"`
	IsOverdue bool      `json:"isOverdue"`
}

// CheckFamilyPurchasePetCount validates 2 ≤ n ≤ 3 for buying / activating Family.
func CheckFamilyPurchasePetCount(n int) error {
	if n < FamilyMinPets {
		return ErrFamilyRequiresTwoPets
	}
	if n > FamilyMaxPets {
		return ErrFamilyPetLimit
	}
	return nil
}

// CheckFamilyCanAddPetCount blocks adding a pet when Family is engaged and n >= max.
func CheckFamilyCanAddPetCount(n int, familyEngaged bool) error {
	if familyEngaged && n >= FamilyMaxPets {
		return ErrFamilyPetLimit
	}
	return nil
}

func (s *Store) CountPetsByOwner(ctx context.Context, ownerUserID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pets.pets WHERE owner_user_id=$1`, ownerUserID).Scan(&n)
	return n, err
}

// ListHouseholdUpcomingCare returns the next pending reminders across the owner's pets.
func (s *Store) ListHouseholdUpcomingCare(ctx context.Context, ownerUserID string, limit int) ([]HouseholdCareItem, error) {
	if limit <= 0 {
		limit = 8
	}
	rows, err := s.pool.Query(ctx, `
		SELECT cr.id::text, cr.pet_id::text, p.name, cr.type, COALESCE(cr.title,''), cr.due_at, cr.status
		FROM care.reminders cr
		JOIN pets.pets p ON p.id = cr.pet_id
		WHERE p.owner_user_id=$1 AND cr.status='pending'
		ORDER BY cr.due_at ASC
		LIMIT $2`, ownerUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now()
	out := make([]HouseholdCareItem, 0)
	for rows.Next() {
		var item HouseholdCareItem
		if err := rows.Scan(&item.ID, &item.PetID, &item.PetName, &item.Type, &item.Title, &item.DueAt, &item.Status); err != nil {
			return nil, err
		}
		item.IsOverdue = item.DueAt.Before(now)
		out = append(out, item)
	}
	return out, rows.Err()
}

// AssertFamilyPurchaseEligible checks 2 ≤ pets ≤ 3 for buying / activating Family.
func (s *Store) AssertFamilyPurchaseEligible(ctx context.Context, ownerUserID string) error {
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	return CheckFamilyPurchasePetCount(n)
}

// AssertFamilyCanAddPet blocks a 4th pet while Family is active or pending checkout.
func (s *Store) AssertFamilyCanAddPet(ctx context.Context, ownerUserID string) error {
	engaged, err := s.HasActiveOrPendingAddon(ctx, ownerUserID, familyAddonCode)
	if err != nil {
		return err
	}
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	return CheckFamilyCanAddPetCount(n, engaged)
}

// CreatePetRespectingFamily atomically enforces the Family pet cap then inserts the pet.
func (s *Store) CreatePetRespectingFamily(ctx context.Context, p Pet) (Pet, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Pet{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, p.OwnerUserID); err != nil {
		return Pet{}, err
	}

	var engaged bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM billing.addon_entitlements
			WHERE owner_user_id=$1 AND addon_code=$2
				AND (
					(status='pending' AND created_at > NOW() - INTERVAL '24 hours')
					OR (status='active' AND (valid_until IS NULL OR valid_until > NOW()))
				)
		)`, p.OwnerUserID, familyAddonCode).Scan(&engaged); err != nil {
		return Pet{}, err
	}
	var n int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pets.pets WHERE owner_user_id=$1`, p.OwnerUserID).Scan(&n); err != nil {
		return Pet{}, err
	}
	if err := CheckFamilyCanAddPetCount(n, engaged); err != nil {
		return Pet{}, err
	}

	p.ID = uuid.NewString()
	if p.PaymentStatus == "" {
		p.PaymentStatus = "pending_payment"
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, birth_date, weight_kg, photo_url, payment_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at`,
		p.ID, p.PracticeID, p.OwnerUserID, p.Name, p.Species, p.Breed, p.BirthDate, p.WeightKg, p.PhotoURL, p.PaymentStatus,
	).Scan(&p.CreatedAt)
	if err != nil {
		return Pet{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Pet{}, err
	}
	return p, nil
}
