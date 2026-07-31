package store

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
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

// ListHouseholdUpcomingCare returns the next pending reminders across the owner's
// pets that already have premium access (active / past_due / cancelled, not expired).
func (s *Store) ListHouseholdUpcomingCare(ctx context.Context, ownerUserID string, limit int) ([]HouseholdCareItem, error) {
	if limit <= 0 {
		limit = 8
	}
	rows, err := s.pool.Query(ctx, `
		SELECT cr.id::text, cr.pet_id::text, p.name, cr.type, COALESCE(cr.title,''), cr.due_at, cr.status
		FROM care.reminders cr
		JOIN pets.pets p ON p.id = cr.pet_id
		JOIN billing.pet_entitlements e ON e.pet_id = p.id
		WHERE p.owner_user_id=$1
			AND cr.status='pending'
			AND e.status IN ('active','past_due','cancelled')
			AND (e.valid_until IS NULL OR e.valid_until > NOW())
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

// PetWithPendingEntitlement is one pet + billing row to insert atomically.
type PetWithPendingEntitlement struct {
	Pet         Pet
	PlanCode    string
	BillingMode string
	AmountCents int
}

// CreatePetWithPendingEntitlement inserts a pet and its pending entitlement in one transaction
// so a failed entitlement never leaves an orphan pet without billing row.
func (s *Store) CreatePetWithPendingEntitlement(ctx context.Context, p Pet, planCode, billingMode string, amountCents int) (Pet, error) {
	out, err := s.CreatePetsBatchWithPendingEntitlements(ctx, []PetWithPendingEntitlement{{
		Pet: p, PlanCode: planCode, BillingMode: billingMode, AmountCents: amountCents,
	}})
	if err != nil {
		return Pet{}, err
	}
	return out[0], nil
}

// CreatePetsBatchWithPendingEntitlements inserts all pets + pending entitlements in a single
// transaction (all-or-nothing). Owner advisory lock is taken once for the batch.
func (s *Store) CreatePetsBatchWithPendingEntitlements(ctx context.Context, items []PetWithPendingEntitlement) ([]Pet, error) {
	if len(items) == 0 {
		return nil, ErrValidation
	}
	ownerID := items[0].Pet.OwnerUserID
	for _, it := range items {
		if it.Pet.OwnerUserID != ownerID {
			return nil, ErrValidation
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, ownerID); err != nil {
		return nil, err
	}

	out := make([]Pet, 0, len(items))
	for _, it := range items {
		p := it.Pet
		p.ID = uuid.NewString()
		if p.PaymentStatus == "" {
			p.PaymentStatus = "pending_payment"
		}
		foodChain := kernel.DefaultFoodChainStatus(p.Species)
		if p.FoodChainStatus != "" {
			foodChain = p.FoodChainStatus
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, birth_date, weight_kg, photo_url, payment_status, litter_tag, microchip_number, health_book_number, food_chain_status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			RETURNING created_at`,
			p.ID, nullIfEmpty(strings.TrimSpace(p.PracticeID)), p.OwnerUserID, p.Name, p.Species, p.Breed, p.BirthDate, p.WeightKg, p.PhotoURL, p.PaymentStatus, p.LitterTag,
			p.MicrochipNumber, p.HealthBookNumber, foodChain,
		).Scan(&p.CreatedAt)
		p.FoodChainStatus = foodChain
		if err != nil {
			return nil, err
		}

		entID := uuid.NewString()
		var ent Entitlement
		err = tx.QueryRow(ctx, `
			INSERT INTO billing.pet_entitlements (id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency)
			VALUES ($1,$2,$3,$4,$5,'pending',$6,'eur')
			RETURNING id::text, pet_id::text, owner_user_id::text, plan_code, billing_mode, status, amount_cents, currency, created_at`,
			entID, p.ID, p.OwnerUserID, it.PlanCode, it.BillingMode, it.AmountCents,
		).Scan(&ent.ID, &ent.PetID, &ent.OwnerUserID, &ent.PlanCode, &ent.BillingMode, &ent.Status, &ent.AmountCents, &ent.Currency, &ent.CreatedAt)
		if err != nil {
			return nil, err
		}
		p.Entitlement = &ent
		out = append(out, p)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}
