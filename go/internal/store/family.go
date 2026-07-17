package store

import (
	"context"
	"errors"
	"time"
)

// Family pack rules (owner-scoped addon).
const (
	FamilyMinPets = 2
	FamilyMaxPets = 3
)

var (
	ErrFamilyPetLimit        = errors.New("family_pet_limit")
	ErrFamilyRequiresTwoPets = errors.New("family_requires_two_pets")
	ErrFamilyRequired        = errors.New("family_required")
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

// AssertFamilyPurchaseEligible checks 2 ≤ pets ≤ 3 for buying the Family addon.
func (s *Store) AssertFamilyPurchaseEligible(ctx context.Context, ownerUserID string) error {
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	if n < FamilyMinPets {
		return ErrFamilyRequiresTwoPets
	}
	if n > FamilyMaxPets {
		return ErrFamilyPetLimit
	}
	return nil
}

// AssertFamilyCanAddPet blocks a 4th pet while Family is active.
func (s *Store) AssertFamilyCanAddPet(ctx context.Context, ownerUserID string) error {
	has, err := s.HasActiveAddon(ctx, ownerUserID, "family")
	if err != nil {
		return err
	}
	if !has {
		return nil
	}
	n, err := s.CountPetsByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	if n >= FamilyMaxPets {
		return ErrFamilyPetLimit
	}
	return nil
}
