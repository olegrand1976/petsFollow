package seed

import (
	"context"
	"errors"
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
	if err := seedCommercial(ctx, tx); err != nil {
		return err
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
	if err := st.EnsureCommissionSettings(ctx); err != nil {
		return err
	}
	if err := st.AccrueAllActiveEntitlements(ctx); err != nil {
		return err
	}
	if err := st.AccrueAllCommercialForActiveEntitlements(ctx); err != nil {
		return err
	}
	if err := seedEnrichment(ctx, pool); err != nil {
		return err
	}
	if _, err := st.BackfillEmailJourneys(ctx); err != nil {
		return err
	}
	logSummary()
	return nil
}

func truncateAll(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `DELETE FROM notifications.notification_log`); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `TRUNCATE billing.commercial_payout_lines, billing.commercial_payout_runs, billing.commercial_commission_ledger,
		billing.addon_entitlements, sales.prospects,
		billing.payout_lines, billing.payout_runs, billing.commission_ledger, billing.commission_tiers,
		billing.commission_settings,
		billing.stripe_events, billing.pet_entitlements, billing.stripe_customers,
		identity.email_verification_tokens, identity.password_reset_tokens,
		notifications.client_preferences, notifications.device_tokens,
		discovery.email_sends, discovery.email_journey, discovery.progress,
		visits.visits, care.competitions, care.professional_contacts, care.reminders,
		notifications.notification_preferences, messaging.messages, messaging.threads, messaging.vet_availability,
		heartrate.sessions, pets.dossier_events, pets.pets,
		practice.client_import_rows, practice.client_import_jobs,
		practice.client_vet_link_requests, practice.invitations, practice.practice_clients, practice.practices, identity.users CASCADE`)
	return err
}

func seedCommercial(ctx context.Context, tx pgx.Tx) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordCommercial), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	commercialID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, 'commercial.demo@petsfollow.test', $2, 'Camille Vente', 'commercial', NULL, NOW())`,
		commercialID, string(hash)); err != nil {
		return err
	}
	// vet.demo is assigned to the demo commercial.
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id = $1
		WHERE email = 'vet.demo@petsfollow.test' AND role = 'vet'`, commercialID); err != nil {
		return err
	}
	return seedProspects(ctx, tx, commercialID)
}

func seedProspects(ctx context.Context, tx pgx.Tx, commercialID string) error {
	prospects := []struct {
		practiceName, contactName, contactEmail, contactPhone, city, notes, status string
		ageDays                                                                     int
	}{
		{"Clinique des Alpes", "Dr Sarah Alpes", "contact@alpes-vet.test", "0450112233", "Annecy", "Intéressée par le suivi cardiaque.", "qualified", 12},
		{"Cabinet du Vieux Port", "Dr Marc Port", "marc@vieuxport-vet.test", "0491223344", "Marseille", "Premier contact salon pro.", "contacted", 5},
		{"Vétérinaire Océan", "Dr Léa Océan", "lea@ocean-vet.test", "0240334455", "Nantes", "Demande de démo.", "new", 1},
		{"Centre Animalier Bordeaux", "Dr Hugo Giron", "hugo@bordeaux-vet.test", "0556445566", "Bordeaux", "A signé, onboarding en cours.", "converted", 30},
		{"Clinique Petite Patte", "Dr Nina Petit", "nina@petitepatte.test", "0388556677", "Strasbourg", "Pas de budget cette année.", "lost", 45},
	}
	for _, p := range prospects {
		if _, err := tx.Exec(ctx, `
			INSERT INTO sales.prospects (id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone, city, notes, status, status_changed_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW() - make_interval(days => $10), NOW() - make_interval(days => $10))`,
			uuid.NewString(), commercialID, p.practiceName, p.contactName, p.contactEmail, p.contactPhone, p.city, p.notes, p.status, p.ageDays); err != nil {
			return err
		}
	}
	return nil
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
	if p.seedPasswordReset {
		if _, err := tx.Exec(ctx, `
			INSERT INTO identity.password_reset_tokens (id, user_id, token, expires_at)
			VALUES ($1, $2, $3, NOW() + INTERVAL '7 days')`,
			uuid.NewString(), vetID, demoPasswordResetToken); err != nil {
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
	if c.seedDiscovery {
		if err := insertDiscoveryProgress(ctx, tx, clientID); err != nil {
			return err
		}
	}
	if c.seedActiveAddons {
		if err := seedClientActiveAddons(ctx, tx, clientID); err != nil {
			return err
		}
	}
	return nil
}

func seedClientActiveAddons(ctx context.Context, tx pgx.Tx, ownerUserID string) error {
	now := time.Now()
	from := now.Add(-30 * 24 * time.Hour)
	until := from.AddDate(0, 0, billing.AddonDurationDays)
	for _, code := range []billing.AddonCode{billing.AddonCarePlus, billing.AddonKennel, billing.AddonHorse} {
		addon, err := billing.GetAddon(code)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO billing.addon_entitlements (id, owner_user_id, addon_code, status, amount_cents, currency, valid_from, valid_until)
			VALUES ($1, $2, $3, 'active', $4, 'eur', $5, $6)`,
			uuid.NewString(), ownerUserID, string(code), addon.AmountCents, from, until); err != nil {
			return err
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
	for _, cr := range pet.careReminders {
		if err := insertCareReminder(ctx, tx, petID, reg.practiceID, cr); err != nil {
			return err
		}
	}
	for _, v := range pet.visits {
		if err := insertVisit(ctx, tx, petID, reg.practiceID, v); err != nil {
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

func insertCareReminder(ctx context.Context, tx pgx.Tx, petID, practiceID string, cr careReminderDef) error {
	status := cr.status
	if status == "" {
		status = "pending"
	}
	dueAt := time.Now().AddDate(0, 0, cr.dueDays)
	updatedAt := dueAt
	if status == "done" {
		updatedAt = time.Now().AddDate(0, 0, cr.dueDays)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.NewString(), petID, practiceID, cr.reminderType, cr.title, dueAt, status, updatedAt)
	return err
}

func insertVisit(ctx context.Context, tx pgx.Tx, petID, practiceID string, v visitDef) error {
	status := v.status
	if status == "" {
		status = "requested"
	}
	source := v.source
	if source == "" {
		source = "client"
	}
	var scheduledAt *time.Time
	if v.scheduledIn != 0 {
		t := time.Now().Add(v.scheduledIn)
		scheduledAt = &t
	}
	var pending *string
	if status == "requested" {
		p := "vet"
		if source == "vet" {
			p = "client"
		}
		pending = &p
	}
	var duration any
	if scheduledAt != nil {
		duration = 30
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, pending_action_by, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.NewString(), petID, practiceID, scheduledAt, status, v.notes, source, pending, duration)
	return err
}

func insertDiscoveryProgress(ctx context.Context, tx pgx.Tx, userID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO discovery.progress (user_id, started_at, completed_cards, streak_days, updated_at)
		VALUES ($1, NOW() - INTERVAL '2 days', '["day0","day2"]'::jsonb, 2, NOW())`,
		userID)
	return err
}

func seedEnrichment(ctx context.Context, pool *pgxpool.Pool) error {
	for _, practice := range demoPractices {
		for _, client := range practice.clients {
			if client.extraPracticeVet == "" {
				continue
			}
			var clientID, vetID, practiceID string
			err := pool.QueryRow(ctx, `
				SELECT u.id::text, v.id::text, v.practice_id::text
				FROM identity.users u
				JOIN identity.users v ON v.email = $2 AND v.role = 'vet'
				WHERE u.email = $1 AND u.role = 'client'`,
				client.email, client.extraPracticeVet).Scan(&clientID, &vetID, &practiceID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			if _, err := pool.Exec(ctx, `
				INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
				uuid.NewString(), practiceID, clientID, vetID); err != nil {
				return err
			}
		}
	}

	// Pending link request for Pro /requests inbox (client.marie → vet.demo).
	var marieID, vetDemoID, vetPlusID string
	err := pool.QueryRow(ctx, `
		SELECT c.id::text, v.id::text, v.practice_id::text
		FROM identity.users c
		JOIN identity.users v ON v.email = 'vet.demo@petsfollow.test' AND v.role = 'vet'
		WHERE c.email = 'client.marie@petsfollow.test' AND c.role = 'client'`).Scan(&marieID, &vetDemoID, &vetPlusID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil {
		if _, err := pool.Exec(ctx, `
			INSERT INTO practice.client_vet_link_requests (id, client_user_id, practice_id, vet_user_id, status)
			VALUES ($1, $2, $3, $4, 'pending')
			ON CONFLICT (client_user_id, practice_id) DO UPDATE SET
				vet_user_id = EXCLUDED.vet_user_id,
				status = 'pending',
				updated_at = NOW()`,
			uuid.NewString(), marieID, vetPlusID, vetDemoID); err != nil {
			return err
		}
	}
	return nil
}

func logSummary() {
	log.Println("--- Comptes démo petsFollow ---")
	log.Printf("Admin  : admin.demo@petsfollow.test / %s", passwordAdmin)
	log.Printf("Commerc: commercial.demo@petsfollow.test / %s (vet.demo assigné, 5 prospects)", passwordCommercial)
	log.Printf("Vétos  : *@petsfollow.test / %s", passwordVet)
	log.Println("  vet.demo@        — VetPlus (profil complet, messages non lus, BPM pending)")
	log.Println("  vet.parc@        — Clinique du Parc (alerte Chouchou)")
	log.Println("  vet.lyon@        — Lyon (indisponible, Nico pending payment)")
	log.Println("  vet.onboarding@  — profil cabinet à compléter (onboarding)")
	log.Println("  vet.unverified@  — email non confirmé (login bloqué)")
	log.Println("  vet.reset@       — token démo reset mot de passe")
	log.Printf("Clients: *@petsfollow.test / %s", passwordClient)
	log.Println("  client.demo@     — 6 animaux · addons Care+/Kennel/Horse actifs")
	log.Println("  client.vide@     — sans animal (kanban)")
	log.Println("  client.marie@    — Mimi + Chouchou · client.paul@ — Max")
	log.Println("  client.julie@    — Oscar · client.thomas@ — Luna + Nico (pending)")
	log.Printf("Confirm email : http://localhost:3002/confirm-email?token=%s", demoEmailConfirmToken)
	log.Printf("Reset password: http://localhost:3002/reset-password?token=%s", demoPasswordResetToken)
}
