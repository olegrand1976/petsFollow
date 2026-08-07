package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	WalkinClientDisplayName = "Nouveau client"
	WalkinPetDisplayName    = "Nouvel animal"
	WalkinEmailDomain       = "petsfollow.invalid"
	WalkinPetSpecies        = "other"
)

// ErrWalkinImmutable is returned when mutating a walk-in placeholder client/pet.
var ErrWalkinImmutable = fmt.Errorf("%w: walkin_immutable", ErrForbidden)

// ErrWalkinIdentifyRequired is returned when clinical billing actions need a real identity.
var ErrWalkinIdentifyRequired = fmt.Errorf("%w: walkin_identify_required", ErrValidation)

type walkinQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// WalkinClientEmail returns the synthetic non-deliverable email for a practice walk-in slot.
func WalkinClientEmail(practiceID string) string {
	return fmt.Sprintf("walkin+%s@%s", strings.TrimSpace(practiceID), WalkinEmailDomain)
}

// IsWalkinPlaceholderEmail reports whether an email belongs to a walk-in slot account.
func IsWalkinPlaceholderEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return strings.HasPrefix(email, "walkin+") && strings.HasSuffix(email, "@"+WalkinEmailDomain)
}

// WalkinPlaceholders holds the system client + pet for a practice.
type WalkinPlaceholders struct {
	ClientID string `json:"clientId"`
	PetID    string `json:"petId"`
}

// EnsureWalkinPlaceholders creates the immutable Nouveau client / Nouvel animal pair if missing.
func (s *Store) EnsureWalkinPlaceholders(ctx context.Context, practiceID string) (WalkinPlaceholders, error) {
	return ensureWalkinPlaceholders(ctx, s.pool, practiceID)
}

// EnsureWalkinPlaceholdersTx is the transactional variant (seed).
func EnsureWalkinPlaceholdersTx(ctx context.Context, tx pgx.Tx, practiceID string) (WalkinPlaceholders, error) {
	return ensureWalkinPlaceholders(ctx, tx, practiceID)
}

func ensureWalkinPlaceholders(ctx context.Context, q walkinQuerier, practiceID string) (WalkinPlaceholders, error) {
	practiceID = strings.TrimSpace(practiceID)
	if practiceID == "" {
		return WalkinPlaceholders{}, fmt.Errorf("%w: practice_required", ErrValidation)
	}

	var out WalkinPlaceholders
	err := q.QueryRow(ctx, `
		SELECT u.id::text, p.id::text
		FROM identity.users u
		JOIN pets.pets p ON p.owner_user_id = u.id AND p.practice_id = u.practice_id AND p.is_walkin_placeholder
		WHERE u.practice_id = $1 AND u.is_walkin_placeholder AND u.role = 'client'
		LIMIT 1`, practiceID).Scan(&out.ClientID, &out.PetID)
	if err == nil {
		return out, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return WalkinPlaceholders{}, err
	}

	// Resolve a reference vet for practice_clients / thread (first vet of practice).
	var vetID string
	err = q.QueryRow(ctx, `
		SELECT id::text FROM identity.users
		WHERE practice_id = $1 AND role = 'vet'
		ORDER BY created_at ASC NULLS LAST, id ASC
		LIMIT 1`, practiceID).Scan(&vetID)
	if errors.Is(err, pgx.ErrNoRows) || strings.TrimSpace(vetID) == "" {
		return WalkinPlaceholders{}, fmt.Errorf("%w: practice_vet_required", ErrValidation)
	}
	if err != nil {
		return WalkinPlaceholders{}, err
	}

	clientID := uuid.NewString()
	petID := uuid.NewString()
	email := WalkinClientEmail(practiceID)

	if _, err := q.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, first_name, last_name, role, practice_id,
			email_verified_at, preferred_locale, must_change_password, is_walkin_placeholder
		) VALUES (
			$1, $2, NULL, $3, $4, '', 'client', $5,
			NOW(), $6, true, true
		)`,
		clientID, email, WalkinClientDisplayName, WalkinClientDisplayName, practiceID, i18n.NormalizeLocale("")); err != nil {
		// Race / retry: unique on email or one-walkin-per-practice.
		if isUniqueViolation(err) {
			return ensureWalkinPlaceholders(ctx, q, practiceID)
		}
		return WalkinPlaceholders{}, err
	}

	if _, err := q.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
		uuid.NewString(), practiceID, clientID, vetID); err != nil {
		return WalkinPlaceholders{}, err
	}

	if _, err := q.Exec(ctx, `
		INSERT INTO pets.pets (
			id, practice_id, owner_user_id, name, species, breed, payment_status,
			food_chain_status, is_walkin_placeholder
		) VALUES ($1, $2, $3, $4, $5, '', 'pending_payment', 'companion', true)`,
		petID, practiceID, clientID, WalkinPetDisplayName, WalkinPetSpecies); err != nil {
		if isUniqueViolation(err) {
			return ensureWalkinPlaceholders(ctx, q, practiceID)
		}
		return WalkinPlaceholders{}, err
	}

	return WalkinPlaceholders{ClientID: clientID, PetID: petID}, nil
}

// IsWalkinPlaceholderUser reports whether the user is the system walk-in slot.
func (s *Store) IsWalkinPlaceholderUser(ctx context.Context, userID string) (bool, error) {
	var flag bool
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(is_walkin_placeholder, false) FROM identity.users WHERE id = $1`, userID).Scan(&flag)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return flag, err
}

// IsWalkinPlaceholderPet reports whether the pet is the system walk-in slot.
func (s *Store) IsWalkinPlaceholderPet(ctx context.Context, petID string) (bool, error) {
	var flag bool
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(is_walkin_placeholder, false) FROM pets.pets WHERE id = $1`, petID).Scan(&flag)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return flag, err
}

// VisitNeedsClientIdentify is true when the visit still points at the walk-in pet.
func (s *Store) VisitNeedsClientIdentify(ctx context.Context, visitID string) (bool, error) {
	var walkin bool
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(p.is_walkin_placeholder, false)
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		WHERE v.id = $1 AND v.deleted_at IS NULL`, visitID).Scan(&walkin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return walkin, err
}

// AssertVisitIdentified returns ErrWalkinIdentifyRequired when the visit pet is still a placeholder.
func (s *Store) AssertVisitIdentified(ctx context.Context, visitID string) error {
	need, err := s.VisitNeedsClientIdentify(ctx, visitID)
	if err != nil {
		return err
	}
	if need {
		return ErrWalkinIdentifyRequired
	}
	return nil
}

// IdentifyVisitMode is existing | create.
type IdentifyVisitMode string

const (
	IdentifyVisitExisting IdentifyVisitMode = "existing"
	IdentifyVisitCreate   IdentifyVisitMode = "create"
)

// IdentifyVisitNewPet describes a pet to create under the target client.
type IdentifyVisitNewPet struct {
	Name    string
	Species string
	Breed   string
}

// IdentifyVisitClientInput reassigns a walk-in visit to a real client/pet.
type IdentifyVisitClientInput struct {
	PracticeID   string
	VisitID      string
	ActorUserID  string // vet/staff creating the client link
	Mode         IdentifyVisitMode
	ClientUserID string // existing mode
	PetID        string // existing pet under client (optional if NewPet set)
	NewPet       *IdentifyVisitNewPet
	// Create mode client fields
	FirstName    string
	LastName     string
	ContactPhone string
	Email        string
	Locale       string
}

// IdentifyVisitClientResult is the visit after pet_id reassignment.
type IdentifyVisitClientResult struct {
	Visit  Visit         `json:"visit"`
	Client ClientSummary `json:"client"`
	Pet    Pet           `json:"pet"`
}

// IdentifyVisitClient atomically creates/links a real identity and UPDATEs visits.pet_id only.
func (s *Store) IdentifyVisitClient(ctx context.Context, in IdentifyVisitClientInput) (IdentifyVisitClientResult, error) {
	in.PracticeID = strings.TrimSpace(in.PracticeID)
	in.VisitID = strings.TrimSpace(in.VisitID)
	in.ActorUserID = strings.TrimSpace(in.ActorUserID)
	if in.PracticeID == "" || in.VisitID == "" || in.ActorUserID == "" {
		return IdentifyVisitClientResult{}, fmt.Errorf("%w: identify_input", ErrValidation)
	}

	visit, err := s.GetPracticeVisit(ctx, in.PracticeID, in.VisitID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	need, err := s.IsWalkinPlaceholderPet(ctx, visit.PetID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	if !need {
		return IdentifyVisitClientResult{}, fmt.Errorf("%w: visit_already_identified", ErrValidation)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	defer tx.Rollback(ctx)

	var targetClientID string
	var targetPetID string

	switch in.Mode {
	case IdentifyVisitExisting:
		targetClientID = strings.TrimSpace(in.ClientUserID)
		if targetClientID == "" {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: client_required", ErrValidation)
		}
		walkinClient, err := s.IsWalkinPlaceholderUser(ctx, targetClientID)
		if err != nil {
			return IdentifyVisitClientResult{}, err
		}
		if walkinClient {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: cannot_identify_as_walkin", ErrValidation)
		}
		var linked bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM practice.practice_clients
				WHERE practice_id = $1 AND client_user_id = $2
			)`, in.PracticeID, targetClientID).Scan(&linked); err != nil {
			return IdentifyVisitClientResult{}, err
		}
		if !linked {
			return IdentifyVisitClientResult{}, ErrNotFound
		}
		targetPetID, err = s.resolveIdentifyPetTx(ctx, tx, in.PracticeID, targetClientID, in.PetID, in.NewPet)
		if err != nil {
			return IdentifyVisitClientResult{}, err
		}

	case IdentifyVisitCreate:
		first := strings.TrimSpace(in.FirstName)
		last := strings.TrimSpace(in.LastName)
		phone := strings.TrimSpace(in.ContactPhone)
		if first == "" && last == "" {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: name_required", ErrValidation)
		}
		if phone == "" {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: contact_phone_required", ErrValidation)
		}
		email := strings.TrimSpace(strings.ToLower(in.Email))
		if email == "" {
			email = fmt.Sprintf("desk+%s@%s", uuid.NewString(), WalkinEmailDomain)
		}
		if IsWalkinPlaceholderEmail(email) {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: invalid_email", ErrValidation)
		}
		clientID := uuid.NewString()
		fullName := strings.TrimSpace(first + " " + last)
		if _, err := tx.Exec(ctx, `
			INSERT INTO identity.users (
				id, email, password_hash, full_name, first_name, last_name, role, practice_id,
				email_verified_at, preferred_locale, must_change_password, contact_phone
			) VALUES ($1, $2, NULL, $3, $4, $5, 'client', $6, NOW(), $7, true, $8)`,
			clientID, email, fullName, first, last, in.PracticeID,
			i18n.NormalizeLocale(in.Locale), phone); err != nil {
			if isUniqueViolation(err) {
				return IdentifyVisitClientResult{}, fmt.Errorf("%w: email_taken", ErrConflict)
			}
			return IdentifyVisitClientResult{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
			VALUES ($1, $2, $3, $4)`,
			uuid.NewString(), in.PracticeID, clientID, in.ActorUserID); err != nil {
			return IdentifyVisitClientResult{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
			VALUES ($1, $2, $3, $4, NULL)`,
			uuid.NewString(), in.PracticeID, clientID, in.ActorUserID); err != nil {
			return IdentifyVisitClientResult{}, err
		}
		targetClientID = clientID
		if in.NewPet == nil {
			return IdentifyVisitClientResult{}, fmt.Errorf("%w: pet_required", ErrValidation)
		}
		targetPetID, err = s.insertIdentifyPetTx(ctx, tx, in.PracticeID, targetClientID, *in.NewPet)
		if err != nil {
			return IdentifyVisitClientResult{}, err
		}

	default:
		return IdentifyVisitClientResult{}, fmt.Errorf("%w: invalid_mode", ErrValidation)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE visits.visits SET pet_id = $2
		WHERE id = $1 AND practice_id = $3 AND deleted_at IS NULL`,
		in.VisitID, targetPetID, in.PracticeID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	if tag.RowsAffected() == 0 {
		return IdentifyVisitClientResult{}, ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return IdentifyVisitClientResult{}, err
	}
	if in.Mode == IdentifyVisitCreate {
		_ = s.MarkEmailJourneySkipped(ctx, targetClientID)
	}

	visitOut, err := s.GetPracticeVisit(ctx, in.PracticeID, in.VisitID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	clientOut, err := s.GetClientByPractice(ctx, in.PracticeID, targetClientID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	petOut, err := s.GetPet(ctx, targetPetID)
	if err != nil {
		return IdentifyVisitClientResult{}, err
	}
	return IdentifyVisitClientResult{Visit: visitOut, Client: clientOut, Pet: petOut}, nil
}

func (s *Store) resolveIdentifyPetTx(ctx context.Context, tx pgx.Tx, practiceID, clientID, petID string, newPet *IdentifyVisitNewPet) (string, error) {
	petID = strings.TrimSpace(petID)
	if petID != "" {
		var ownerID string
		var walkin bool
		err := tx.QueryRow(ctx, `
			SELECT owner_user_id::text, COALESCE(is_walkin_placeholder, false)
			FROM pets.pets WHERE id = $1 AND practice_id = $2`, petID, practiceID).Scan(&ownerID, &walkin)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		if err != nil {
			return "", err
		}
		if ownerID != clientID || walkin {
			return "", fmt.Errorf("%w: invalid_pet", ErrValidation)
		}
		return petID, nil
	}
	if newPet == nil {
		return "", fmt.Errorf("%w: pet_required", ErrValidation)
	}
	return s.insertIdentifyPetTx(ctx, tx, practiceID, clientID, *newPet)
}

func (s *Store) insertIdentifyPetTx(ctx context.Context, tx pgx.Tx, practiceID, ownerID string, np IdentifyVisitNewPet) (string, error) {
	name := strings.TrimSpace(np.Name)
	species := strings.TrimSpace(np.Species)
	if name == "" || species == "" {
		return "", fmt.Errorf("%w: name_species_required", ErrValidation)
	}
	planCode := "triennial"
	billingMode := "subscription"
	amountCents := 9500 // billing.GetPlan(PlanTriennial) — keep in sync with billing/domain.go
	petID := uuid.NewString()
	foodChain := kernel.DefaultFoodChainStatus(species)
	if _, err := tx.Exec(ctx, `
		INSERT INTO pets.pets (
			id, practice_id, owner_user_id, name, species, breed, payment_status, food_chain_status, is_walkin_placeholder
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending_payment', $7, false)`,
		petID, practiceID, ownerID, name, species, strings.TrimSpace(np.Breed), foodChain); err != nil {
		return "", err
	}
	entID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.pet_entitlements (id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, 'eur')`,
		entID, petID, ownerID, planCode, billingMode, amountCents); err != nil {
		return "", err
	}
	return petID, nil
}
