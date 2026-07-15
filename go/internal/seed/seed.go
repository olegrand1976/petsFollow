package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type ids struct {
	practiceID string
	vetID      string
	clientIDs  map[string]string // email -> user id
	petIDs     map[string]string // "clientEmail/petName" -> pet id
}

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := truncateAll(ctx, tx); err != nil {
		return err
	}
	if err := seedAdmin(ctx, tx); err != nil {
		return err
	}
	for _, practice := range demoPractices {
		if err := seedPractice(ctx, tx, practice); err != nil {
			return fmt.Errorf("practice %q: %w", practice.name, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
		UPDATE identity.users SET preferred_locale = 'nl'
		WHERE email = 'client.marie@petsfollow.test'`); err != nil {
		return err
	}
	st := store.New(pool)
	if err := st.EnsureDefaultCommissionTiers(ctx); err != nil {
		return err
	}
	if err := st.AccrueAllActiveEntitlements(ctx); err != nil {
		return err
	}
	logSummary()
	return nil
}

func truncateAll(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `DELETE FROM notifications.notification_log`); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `TRUNCATE billing.payout_lines, billing.payout_runs, billing.commission_ledger, billing.commission_tiers,
		billing.stripe_events, billing.pet_entitlements, billing.stripe_customers,
		identity.email_verification_tokens,
		notifications.notification_preferences, messaging.messages, messaging.threads, messaging.vet_availability,
		heartrate.sessions, pets.dossier_events, pets.pets, practice.invitations, practice.practice_clients, practice.practices, identity.users CASCADE`)
	return err
}

func seedAdmin(ctx context.Context, tx pgx.Tx) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordAdmin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, 'admin.demo@petsfollow.test', $2, 'Admin Ops', 'admin', NULL, NOW())`,
		uuid.NewString(), string(hash))
	return err
}

func seedPractice(ctx context.Context, tx pgx.Tx, p practiceDef) error {
	practiceID := uuid.NewString()
	vetID := uuid.NewString()
	vetHash, err := bcrypt.GenerateFromPassword([]byte(passwordVet), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practices (id, name, phone, contact_email, address_line1, address_line2, city, postal_code, website, profile_completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CASE WHEN $10 THEN NULL ELSE NOW() END)`,
		practiceID, p.name, p.phone, p.vetEmail, p.address, p.addressLine2, p.city, p.postalCode, p.website, p.incompleteProfile); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, $4, 'vet', $5, CASE WHEN $6 THEN NULL ELSE NOW() END)`,
		vetID, p.vetEmail, string(vetHash), p.vetName, practiceID, p.pendingEmailVerify); err != nil {
		return err
	}
	if p.pendingEmailVerify {
		if _, err := tx.Exec(ctx, `
			INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
			VALUES ($1, $2, $3, NOW() + INTERVAL '7 days')`,
			uuid.NewString(), vetID, demoEmailConfirmToken); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, $2, $3)`,
		vetID, p.notifyOnMessage, p.notifyOnHeartRate); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, $3, NULLIF($4, ''))`,
		vetID, practiceID, p.availability, p.autoReply); err != nil {
		return err
	}

	registry := &ids{
		practiceID: practiceID,
		vetID:      vetID,
		clientIDs:  make(map[string]string),
		petIDs:     make(map[string]string),
	}

	clientHash, err := bcrypt.GenerateFromPassword([]byte(passwordClient), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	for _, client := range p.clients {
		if err := seedClient(ctx, tx, registry, client, string(clientHash)); err != nil {
			return fmt.Errorf("client %q: %w", client.email, err)
		}
	}
	return nil
}

func seedClient(ctx context.Context, tx pgx.Tx, reg *ids, c clientDef, clientHash string) error {
	clientID := uuid.NewString()
	reg.clientIDs[c.email] = clientID

	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, $4, 'client', $5, NOW())`,
		clientID, c.email, clientHash, c.fullName, reg.practiceID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), reg.practiceID, clientID, reg.vetID); err != nil {
		return err
	}

	for _, pet := range c.pets {
		petKey := c.email + "/" + pet.name
		if err := seedPet(ctx, tx, reg, clientID, petKey, pet); err != nil {
			return fmt.Errorf("pet %q: %w", pet.name, err)
		}
	}

	// One messaging thread per client (pet_id = first pet if any)
	var threadPetID *string
	if len(c.pets) > 0 {
		petKey := c.email + "/" + c.pets[0].name
		if id, ok := reg.petIDs[petKey]; ok {
			threadPetID = &id
		}
	}
	threadID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1, $2, $3, $4, $5)`,
		threadID, reg.practiceID, clientID, reg.vetID, threadPetID); err != nil {
		return err
	}

	for _, pet := range c.pets {
		for _, msg := range pet.messages {
			senderID := clientID
			if msg.senderRole == "vet" {
				senderID = reg.vetID
			}
			if err := insertMessage(ctx, tx, threadID, senderID, msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func seedPet(ctx context.Context, tx pgx.Tx, reg *ids, clientID, petKey string, pet petDef) error {
	petID := uuid.NewString()
	reg.petIDs[petKey] = petID

	if _, err := tx.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, weight_kg, payment_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		petID, reg.practiceID, clientID, pet.name, pet.species, pet.breed, pet.weightKg, pet.paymentStatus); err != nil {
		return err
	}

	if err := seedEntitlement(ctx, tx, petID, clientID, pet); err != nil {
		return err
	}
	for _, hr := range pet.heartRates {
		if err := insertHeartRate(ctx, tx, petID, clientID, reg.practiceID, hr); err != nil {
			return err
		}
	}
	for _, ev := range pet.dossierEvents {
		authorID := reg.vetID
		if ev.authorRole == "client" {
			authorID = clientID
		}
		if err := insertDossierEvent(ctx, tx, petID, authorID, ev); err != nil {
			return err
		}
	}
	return nil
}

func seedEntitlement(ctx context.Context, tx pgx.Tx, petID, clientID string, pet petDef) error {
	plan, err := billing.GetPlan(pet.plan)
	if err != nil {
		return err
	}
	now := time.Now()
	var validFrom, validUntil *time.Time
	if pet.entitlement.AllowsAccess() || pet.entitlement == billing.StatusPending {
		from := now.Add(-30 * 24 * time.Hour)
		until := billing.ValidUntil(from, plan)
		validFrom = &from
		validUntil = &until
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO billing.pet_entitlements (id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency, valid_from, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'eur', $8, $9)`,
		uuid.NewString(), petID, clientID, pet.plan, pet.billingMode, pet.entitlement, plan.AmountCents, validFrom, validUntil)
	return err
}

func insertMessage(ctx context.Context, tx pgx.Tx, threadID, senderID string, msg messageDef) error {
	createdAt := time.Now().Add(msg.age)
	var readAt *time.Time
	if msg.read {
		t := createdAt.Add(30 * time.Minute)
		readAt = &t
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO messaging.messages (id, thread_id, sender_user_id, body, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.NewString(), threadID, senderID, msg.body, readAt, createdAt)
	return err
}

func insertHeartRate(ctx context.Context, tx pgx.Tx, petID, ownerID, practiceID string, hr heartRateDef) error {
	startedAt := time.Now().Add(hr.age)
	var endedAt, validatedAt *time.Time
	switch hr.status {
	case kernel.SessionValidated, kernel.SessionPendingValidation:
		end := startedAt.Add(time.Duration(hr.duration) * time.Second)
		endedAt = &end
		if hr.status == kernel.SessionValidated {
			validatedAt = &end
		}
	case kernel.SessionInProgress:
		// no ended_at
	case kernel.SessionCancelled:
		end := startedAt.Add(time.Duration(hr.duration/2) * time.Second)
		endedAt = &end
	default:
		return fmt.Errorf("unknown session status: %s", hr.status)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO heartrate.sessions (id, pet_id, owner_user_id, practice_id, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		uuid.NewString(), petID, ownerID, practiceID, hr.status, hr.tapCount, hr.duration, hr.bpm, hr.isAlert, startedAt, endedAt, validatedAt)
	return err
}

func insertDossierEvent(ctx context.Context, tx pgx.Tx, petID, authorID string, ev dossierEventDef) error {
	createdAt := time.Now().Add(ev.age)
	_, err := tx.Exec(ctx, `
		INSERT INTO pets.dossier_events (id, pet_id, author_user_id, event_type, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.NewString(), petID, authorID, ev.eventType, ev.content, createdAt)
	return err
}

func logSummary() {
	log.Println("--- Comptes démo petsFollow ---")
	log.Printf("Admin  : admin.demo@petsfollow.test / %s", passwordAdmin)
	log.Printf("Vétos  : *@petsfollow.test / %s", passwordVet)
	log.Println("  vet.demo@        — VetPlus (profil complet, messages non lus, BPM pending)")
	log.Println("  vet.parc@        — Clinique du Parc (alerte Chouchou)")
	log.Println("  vet.lyon@        — Lyon (indisponible, Nico pending payment)")
	log.Println("  vet.onboarding@  — profil cabinet à compléter (onboarding)")
	log.Println("  vet.unverified@  — email non confirmé (login bloqué)")
	log.Printf("Clients: *@petsfollow.test / %s", passwordClient)
	log.Println("  client.demo@     — Rex + Bella · client.vide@ sans animal (kanban)")
	log.Println("  client.marie@    — Mimi + Chouchou · client.paul@ — Max")
	log.Println("  client.julie@    — Oscar · client.thomas@ — Luna + Nico (pending)")
	log.Printf("Confirm email : http://localhost:3002/confirm-email?token=%s", demoEmailConfirmToken)
}
