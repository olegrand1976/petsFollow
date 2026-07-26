package seed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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
	if err := seedDemoSchedules(ctx, pool, st); err != nil {
		return err
	}
	if err := seedCarePros(ctx, pool); err != nil {
		return err
	}
	if err := seedStripeCatalog(ctx, st); err != nil {
		return err
	}
	if err := seedProfilesTeamModules(ctx, pool, st); err != nil {
		return err
	}
	if _, err := st.BackfillEmailJourneys(ctx); err != nil {
		return err
	}
	logSummary()
	return nil
}

func seedDemoSchedules(ctx context.Context, pool *pgxpool.Pool, st *store.Store) error {
	rows, err := pool.Query(ctx, `
		SELECT id::text FROM practice.practices
		WHERE profile_completed_at IS NOT NULL
		ORDER BY name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	year := time.Now().Year()
	slots := []store.ScheduleSlot{
		{Weekday: 1, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 1, StartTime: "14:00", EndTime: "18:00"},
		{Weekday: 2, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 2, StartTime: "14:00", EndTime: "18:00"},
		{Weekday: 3, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 4, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 4, StartTime: "14:00", EndTime: "18:00"},
		{Weekday: 5, StartTime: "09:00", EndTime: "12:00"},
	}
	for rows.Next() {
		var practiceID string
		if err := rows.Scan(&practiceID); err != nil {
			return err
		}
		if _, err := st.PutVetSchedule(ctx, practiceID, true, 30, &year, slots); err != nil {
			return err
		}
	}
	return rows.Err()
}

func payoutLegalName(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	return p.name + " SRL"
}
func payoutVAT(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	return "BE0123456789"
}
func payoutCompanyNumber(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	return "0123.456.789"
}
func payoutLegalForm(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	return "srl"
}
func payoutIBAN(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	// Valid Belgian IBAN (checksum) for demo payouts.
	return "BE68539007547034"
}
func payoutHolder(p practiceDef) string {
	if p.incompleteProfile {
		return ""
	}
	return p.vetName
}

// protectedSalesRoles are never deleted by seed (staging real accounts + demos).
var protectedSalesRoles = []string{"admin", "commercial", "commercial_manager"}

func truncateAll(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `DELETE FROM notifications.notification_log`); err != nil {
		return err
	}
	// Detach surviving users from practices/profiles before clearing those tables.
	if _, err := tx.Exec(ctx, `UPDATE identity.users SET practice_id = NULL, active_profile_id = NULL WHERE true`); err != nil {
		return err
	}
	// Drop FK so TRUNCATE … practices CASCADE cannot wipe identity.users via
	// profiles.practice_id → practices and users.active_profile_id → profiles.
	if _, err := tx.Exec(ctx, `
		ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_active_profile_id_fkey`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM practice.team_members`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM identity.profiles`); err != nil {
		return err
	}
	// identity.users is intentionally NOT truncated: admin / commercial / commercial_manager must survive.
	if _, err := tx.Exec(ctx, `TRUNCATE billing.commercial_payout_lines, billing.commercial_payout_runs, billing.commercial_commission_ledger,
		billing.commercial_bonus_awards,
		billing.addon_entitlements, sales.prospects,
		billing.payout_lines, billing.payout_runs, billing.commission_ledger, billing.commission_tiers,
		billing.commission_settings,
		billing.stripe_events, billing.pet_entitlements, billing.stripe_customers,
		billing.stripe_prices, billing.stripe_products,
		identity.email_verification_tokens, identity.password_reset_tokens,
		notifications.client_preferences, notifications.device_tokens,
		discovery.email_sends, discovery.email_journey, discovery.progress,
		ops.support_ticket_replies, ops.support_tickets,
		ops.product_digest_sends, ops.product_digests,
		visits.visits, care.competitions, care.professional_contacts, care.reminders,
		notifications.notification_preferences, messaging.messages, messaging.threads, messaging.vet_availability,
		heartrate.sessions, pets.weight_readings, pets.dossier_events, pets.pets,
		practice.vet_schedule_slots, practice.vet_schedule, practice.vet_vacations,
		practice.client_import_rows, practice.client_import_jobs,
		practice.client_vet_link_requests, practice.invitations, practice.app_invite_codes,
		practice.commercial_referrals,
		practice.practice_clients, practice.practices CASCADE`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		ALTER TABLE identity.users
		ADD CONSTRAINT users_active_profile_id_fkey
		FOREIGN KEY (active_profile_id) REFERENCES identity.profiles(id) ON DELETE SET NULL`); err != nil {
		return err
	}
	// Drop ephemeral smoke commercials (role protected, but email pattern is disposable).
	if _, err := tx.Exec(ctx, `
		DELETE FROM identity.users
		WHERE email LIKE 'smoke-comm+%@petsfollow.test' AND role = 'commercial'`); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		DELETE FROM identity.users
		WHERE role <> ALL($1::text[])`, protectedSalesRoles)
	return err
}

func seedCommercial(ctx context.Context, tx pgx.Tx) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordCommercial), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var managerID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at,
			payout_iban, payout_bic, payout_account_holder, must_change_password,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES (
			$1, 'commercial.manager@petsfollow.test', $2, 'Bérénice Manager', 'commercial_manager', NULL, NOW(),
			'BE68539007547034', 'GEBABEBB', 'Bérénice Manager', false,
			50.6326, 5.5797, 'Liège', '4000'
		)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = 'commercial_manager',
			email_verified_at = COALESCE(identity.users.email_verified_at, NOW()),
			payout_iban = EXCLUDED.payout_iban,
			payout_bic = EXCLUDED.payout_bic,
			payout_account_holder = EXCLUDED.payout_account_holder,
			must_change_password = false,
			base_lat = COALESCE(identity.users.base_lat, EXCLUDED.base_lat),
			base_lng = COALESCE(identity.users.base_lng, EXCLUDED.base_lng),
			base_city = COALESCE(NULLIF(identity.users.base_city, ''), EXCLUDED.base_city),
			base_postal_code = COALESCE(NULLIF(identity.users.base_postal_code, ''), EXCLUDED.base_postal_code)
		RETURNING id::text`,
		uuid.NewString(), string(hash)).Scan(&managerID); err != nil {
		return err
	}
	var commercialID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at,
			payout_iban, payout_bic, payout_account_holder, manager_user_id, sponsor_user_id, must_change_password,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES (
			$1, 'commercial.demo@petsfollow.test', $2, 'Camille Vente', 'commercial', NULL, NOW(),
			'BE68539007547034', 'GEBABEBB', 'Camille Vente', $3::uuid, $3::uuid, false,
			50.8503, 4.3517, 'Bruxelles', '1000'
		)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = 'commercial',
			email_verified_at = COALESCE(identity.users.email_verified_at, NOW()),
			payout_iban = EXCLUDED.payout_iban,
			payout_bic = EXCLUDED.payout_bic,
			payout_account_holder = EXCLUDED.payout_account_holder,
			-- COALESCE kept for upsert symmetry; demo emails are force-linked just below.
			manager_user_id = COALESCE(identity.users.manager_user_id, EXCLUDED.manager_user_id),
			sponsor_user_id = COALESCE(identity.users.sponsor_user_id, EXCLUDED.sponsor_user_id),
			must_change_password = false,
			base_lat = COALESCE(identity.users.base_lat, EXCLUDED.base_lat),
			base_lng = COALESCE(identity.users.base_lng, EXCLUDED.base_lng),
			base_city = COALESCE(NULLIF(identity.users.base_city, ''), EXCLUDED.base_city),
			base_postal_code = COALESCE(NULLIF(identity.users.base_postal_code, ''), EXCLUDED.base_postal_code)
		RETURNING id::text`,
		uuid.NewString(), string(hash), managerID).Scan(&commercialID); err != nil {
		return err
	}
	var commercial2ID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at,
			payout_iban, payout_bic, payout_account_holder, manager_user_id, sponsor_user_id, must_change_password,
			base_lat, base_lng, base_city, base_postal_code
		) VALUES (
			$1, 'commercial.demo2@petsfollow.test', $2, 'Alex Vente', 'commercial', NULL, NOW(),
			'BE68539007547034', 'GEBABEBB', 'Alex Vente', $3::uuid, $3::uuid, false,
			50.6292, 3.0573, 'Lille', '59000'
		)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = 'commercial',
			email_verified_at = COALESCE(identity.users.email_verified_at, NOW()),
			payout_iban = EXCLUDED.payout_iban,
			payout_bic = EXCLUDED.payout_bic,
			payout_account_holder = EXCLUDED.payout_account_holder,
			manager_user_id = COALESCE(identity.users.manager_user_id, EXCLUDED.manager_user_id),
			sponsor_user_id = COALESCE(identity.users.sponsor_user_id, EXCLUDED.sponsor_user_id),
			must_change_password = false,
			base_lat = COALESCE(identity.users.base_lat, EXCLUDED.base_lat),
			base_lng = COALESCE(identity.users.base_lng, EXCLUDED.base_lng),
			base_city = COALESCE(NULLIF(identity.users.base_city, ''), EXCLUDED.base_city),
			base_postal_code = COALESCE(NULLIF(identity.users.base_postal_code, ''), EXCLUDED.base_postal_code)
		RETURNING id::text`,
		uuid.NewString(), string(hash), managerID).Scan(&commercial2ID); err != nil {
		return err
	}
	// Force-link known demo reps to seed manager so overview always shows the 2-rep team
	// after re-seed (COALESCE above preserves staging reassignments for non-demo emails only).
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET manager_user_id = $1,
		  sponsor_user_id = CASE
		    WHEN sponsor_user_id IS NULL OR sponsor_user_id IS NOT DISTINCT FROM manager_user_id THEN $1
		    ELSE sponsor_user_id
		  END
		WHERE email IN ('commercial.demo@petsfollow.test', 'commercial.demo2@petsfollow.test')`,
		managerID); err != nil {
		return err
	}
	// vet.demo → Camille ; vet.parc → Alex (deux portefeuilles distincts).
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id = $1
		WHERE email = 'vet.demo@petsfollow.test' AND role = 'vet'`, commercialID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id = $1
		WHERE email = 'vet.parc@petsfollow.test' AND role = 'vet'`, commercial2ID); err != nil {
		return err
	}
	if err := seedProspects(ctx, tx, commercialID, demo1Prospects); err != nil {
		return err
	}
	if err := seedProspects(ctx, tx, commercial2ID, demo2Prospects); err != nil {
		return err
	}
	return seedSalesBranches(ctx, tx, managerID, commercialID, commercial2ID)
}

func seedSalesBranches(ctx context.Context, tx pgx.Tx, managerID, commercialID, commercial2ID string) error {
	var bruxellesID, nordID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO sales.branches (id, name, code, external_mlm_id)
		VALUES ($1, 'Bruxelles', 'BRU', NULL)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id::text`, uuid.NewString()).Scan(&bruxellesID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO sales.branches (id, name, code, external_mlm_id)
		VALUES ($1, 'Nord', 'NORD', NULL)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id::text`, uuid.NewString()).Scan(&nordID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users
		SET branch_id = $1,
		    sponsor_user_id = COALESCE(sponsor_user_id, manager_user_id)
		WHERE id = $2`, bruxellesID, commercialID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users
		SET branch_id = $1,
		    sponsor_user_id = COALESCE(sponsor_user_id, manager_user_id)
		WHERE id = $2`, nordID, commercial2ID); err != nil {
		return err
	}
	// Manager sees both branches; attach to Bruxelles as primary for seed.
	_, err := tx.Exec(ctx, `
		UPDATE identity.users SET branch_id = COALESCE(branch_id, $1)
		WHERE id = $2`, bruxellesID, managerID)
	return err
}

type seedProspect struct {
	practiceName, contactName, contactEmail, contactPhone, city, notes, status string
	ageDays                                                                    int
}

var demo1Prospects = []seedProspect{
	{"Clinique des Alpes", "Dr Sarah Alpes", "contact@alpes-vet.test", "0450112233", "Annecy", "Intéressée par le suivi cardiaque.", "qualified", 12},
	{"Cabinet du Vieux Port", "Dr Marc Port", "marc@vieuxport-vet.test", "0491223344", "Marseille", "Premier contact salon pro.", "contacted", 5},
	{"Vétérinaire Océan", "Dr Léa Océan", "lea@ocean-vet.test", "0240334455", "Nantes", "Demande de démo.", "new", 1},
	{"Centre Animalier Bordeaux", "Dr Hugo Giron", "hugo@bordeaux-vet.test", "0556445566", "Bordeaux", "A signé, onboarding en cours.", "converted", 30},
	{"Clinique Petite Patte", "Dr Nina Petit", "nina@petitepatte.test", "0388556677", "Strasbourg", "Pas de budget cette année.", "lost", 45},
}

var demo2Prospects = []seedProspect{
	{"Cabinet Flandres Vet", "Dr Inès Flandres", "ines@flandres-vet.test", "0320112233", "Lille", "Relance module CR IA.", "qualified", 10},
	{"Clinique Grand Place", "Dr Tom Place", "tom@grandplace-vet.test", "0321223344", "Mons", "RDV démo planifié.", "contacted", 4},
	{"Vet & Co Tournai", "Dr Sara Tour", "sara@vetco-tournai.test", "0690334455", "Tournai", "Lead salon.", "new", 2},
	{"Centre Équin Ardenne", "Dr Luc Ardenne", "luc@equin-ardenne.test", "0612445566", "Namur", "Converti — onboarding.", "converted", 28},
	{"Cabinet du Canal", "Dr Eva Canal", "eva@canal-vet.test", "02-5556677", "Bruxelles", "Pas intéressée cette année.", "lost", 40},
}

func seedProspects(ctx context.Context, tx pgx.Tx, commercialID string, prospects []seedProspect) error {
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
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at, must_change_password)
		VALUES ($1, 'admin.demo@petsfollow.test', $2, 'Admin Ops', 'admin', NULL, NOW(), false)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			full_name = EXCLUDED.full_name,
			role = 'admin',
			email_verified_at = COALESCE(identity.users.email_verified_at, NOW()),
			must_change_password = false`,
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
		INSERT INTO practice.practices (
			id, name, phone, contact_email, address_line1, address_line2, city, postal_code, website, profile_completed_at,
			company_legal_name, vat_number, company_number, legal_form, billing_same_as_practice,
			payout_iban, payout_bic, payout_account_holder
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, CASE WHEN $10 THEN NULL ELSE NOW() END,
			$11, $12, $13, $14, TRUE, $15, $16, $17
		)`,
		practiceID, p.name, p.phone, p.vetEmail, p.address, p.addressLine2, p.city, p.postalCode, p.website, p.incompleteProfile,
		payoutLegalName(p), payoutVAT(p), payoutCompanyNumber(p), payoutLegalForm(p),
		payoutIBAN(p), "GEBABEBB", payoutHolder(p)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, email_verified_at)
		VALUES ($1, $2, $3, $4, 'vet', $5, CASE WHEN $6 THEN NULL ELSE NOW() END)`,
		vetID, p.vetEmail, string(vetHash), p.vetName, practiceID, p.pendingEmailVerify); err != nil {
		return err
	}
	// Demo: CR IA trial so Flutter/Nuxt dictation works out of the box (admin can still manage).
	if !p.incompleteProfile {
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.ai_cr_modules (
				practice_id, status, activated_at, trial_ends_at, activated_by_user_id,
				price_plan, baseline_minutes_per_cr, hourly_cost_cents, updated_at
			) VALUES ($1, 'trial', NOW(), NOW() + INTERVAL '90 days', $2, 'monthly_39', 10, 8000, NOW())
			ON CONFLICT (practice_id) DO NOTHING`, practiceID, vetID); err != nil {
			return err
		}
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

	hasSubscription := false
	for _, pet := range c.pets {
		petKey := c.email + "/" + pet.name
		if err := seedPet(ctx, tx, reg, clientID, petKey, pet); err != nil {
			return fmt.Errorf("pet %q: %w", pet.name, err)
		}
		if pet.billingMode == billing.ModeSubscription {
			hasSubscription = true
		}
	}
	// Portail Stripe (Gérer mon abonnement) exige un customer — même convention que le mock checkout.
	if hasSubscription {
		if _, err := tx.Exec(ctx, `
			INSERT INTO billing.stripe_customers (user_id, stripe_customer_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id`,
			clientID, billing.MockCustomerID(clientID)); err != nil {
			return err
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
	for _, w := range pet.weights {
		if err := insertWeightReading(ctx, tx, petID, clientID, reg.practiceID, w); err != nil {
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
		// Backdate start for multi-month plans; monthly (30d) must not expire at seed time.
		from := now
		if plan.DurationDays > 30 {
			from = now.Add(-30 * 24 * time.Hour)
		} else if plan.DurationDays > 1 {
			from = now.Add(-24 * time.Hour)
		}
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
	var endedAt, validatedAt, vetSeenAt *time.Time
	switch hr.status {
	case kernel.SessionValidated, kernel.SessionPendingValidation:
		end := startedAt.Add(time.Duration(hr.duration) * time.Second)
		endedAt = &end
		if hr.status == kernel.SessionValidated {
			validatedAt = &end
			if !hr.unread {
				vetSeenAt = &end
			}
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
		INSERT INTO heartrate.sessions (id, pet_id, owner_user_id, practice_id, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at, vet_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		uuid.NewString(), petID, ownerID, practiceID, hr.status, hr.tapCount, hr.duration, hr.bpm, hr.isAlert, startedAt, endedAt, validatedAt, vetSeenAt)
	return err
}

func insertWeightReading(ctx context.Context, tx pgx.Tx, petID, ownerID, practiceID string, w weightReadingDef) error {
	recordedAt := time.Now().Add(w.age)
	var comment *string
	if strings.TrimSpace(w.comment) != "" {
		c := strings.TrimSpace(w.comment)
		comment = &c
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO pets.weight_readings (
			id, pet_id, owner_user_id, author_user_id, practice_id, weight_kg, comment, recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.NewString(), petID, ownerID, ownerID, practiceID, w.kg, comment, recordedAt)
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

func seedCarePros(ctx context.Context, pool *pgxpool.Pool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordCarePro), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	type careProSeed struct {
		email, name, specialty string
	}
	pros := []careProSeed{
		{"farrier.demo@petsfollow.test", "Marc Ferrier", string(kernel.SpecialtyFarrier)},
		{"vetlight.demo@petsfollow.test", "Dr Léa Light", string(kernel.SpecialtyVetLight)},
	}
	var farrierID, vetLightID string
	for _, p := range pros {
		var uid string
		err := pool.QueryRow(ctx, `
			INSERT INTO identity.users (
				id, email, password_hash, full_name, role, practice_id,
				email_verified_at, professional_specialty, preferred_locale
			) VALUES ($1, $2, $3, $4, 'care_pro', NULL, NOW(), $5, 'fr')
			ON CONFLICT (email) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				full_name = EXCLUDED.full_name,
				role = 'care_pro',
				professional_specialty = EXCLUDED.professional_specialty,
				preferred_locale = COALESCE(identity.users.preferred_locale, EXCLUDED.preferred_locale),
				email_verified_at = COALESCE(identity.users.email_verified_at, NOW())
			RETURNING id::text`,
			uuid.NewString(), p.email, string(hash), p.name, p.specialty).Scan(&uid)
		if err != nil {
			return fmt.Errorf("care_pro %s: %w", p.email, err)
		}
		if p.specialty == string(kernel.SpecialtyFarrier) {
			farrierID = uid
		} else {
			vetLightID = uid
		}
	}

	var spiritID, ownerID, practiceID string
	err = pool.QueryRow(ctx, `
		SELECT p.id::text, p.owner_user_id::text, COALESCE(p.practice_id::text, '')
		FROM pets.pets p
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE u.email = 'client.demo@petsfollow.test' AND p.name = 'Spirit'
		LIMIT 1`).Scan(&spiritID, &ownerID, &practiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("seedCarePros: Spirit introuvable — care_pro créés sans grant ni visite")
			return nil
		}
		return err
	}

	grant := func(granteeID, perm string) error {
		_, err := pool.Exec(ctx, `
			INSERT INTO pets.pet_access (id, pet_id, grantee_user_id, permission, granted_by_user_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (pet_id, grantee_user_id) DO UPDATE SET
				permission = EXCLUDED.permission,
				granted_by_user_id = EXCLUDED.granted_by_user_id`,
			uuid.NewString(), spiritID, granteeID, perm, ownerID)
		return err
	}
	if err := grant(farrierID, "write_notes"); err != nil {
		return err
	}
	if err := grant(vetLightID, "write_notes"); err != nil {
		return err
	}

	if practiceID == "" {
		return nil
	}
	var n int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.visits
		WHERE pet_id = $1 AND notes LIKE '%démo care_pro%'`, spiritID).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		_, err = pool.Exec(ctx, `
			UPDATE visits.visits
			SET scheduled_at = date_trunc('day', NOW()) + INTERVAL '10 hours'
			WHERE pet_id = $1 AND notes LIKE '%démo care_pro%'
			  AND status NOT IN ('done', 'cancelled')`, spiritID)
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO visits.visits (
			id, pet_id, practice_id, scheduled_at, status, notes, source,
			pending_action_by, duration_minutes, address_text
		) VALUES (
			$1, $2, $3, date_trunc('day', NOW()) + INTERVAL '10 hours', 'confirmed',
			'Ferrage Spirit — démo care_pro', 'vet', NULL, 30, 'Écurie VetPlus Demo'
		)`,
		uuid.NewString(), spiritID, practiceID)
	return err
}

func logSummary() {
	log.Println("--- Comptes démo petsFollow ---")
	log.Printf("Admin  : admin.demo@petsfollow.test / %s", passwordAdmin)
	log.Printf("Manager: commercial.manager@petsfollow.test / %s", passwordCommercial)
	log.Printf("Commerc: commercial.demo@petsfollow.test / %s (vet.demo assigné, 5 prospects, rattaché manager)", passwordCommercial)
	log.Printf("Commerc: commercial.demo2@petsfollow.test / %s (vet.parc assigné, 5 prospects Nord, rattaché manager)", passwordCommercial)
	log.Printf("Vétos  : *@petsfollow.test / %s", passwordVet)
	log.Println("  vet.demo@        — VetPlus (profil complet, messages non lus, BPM pending)")
	log.Println("  vet.parc@        — Clinique du Parc (alerte Chouchou)")
	log.Println("  vet.lyon@        — Lyon (indisponible, Nico pending payment)")
	log.Println("  vet.onboarding@  — profil cabinet à compléter (onboarding)")
	log.Println("  vet.unverified@  — email non confirmé (login bloqué)")
	log.Println("  vet.reset@       — token démo reset mot de passe")
	log.Printf("Clients: *@petsfollow.test / %s", passwordClient)
	log.Println("  client.demo@     — 6 animaux · mix monthly/annual/triennial")
	log.Println("  client.vide@     — sans animal (kanban)")
	log.Println("  client.marie@    — Mimi + Chouchou · client.paul@ — Max")
	log.Println("  client.julie@    — Oscar · client.thomas@ — Luna + Nico (pending)")
	log.Printf("Care pro: *@petsfollow.test / %s (Flutter pro light)", passwordCarePro)
	log.Println("  farrier.demo@    — maréchal · write_notes sur Spirit + visite ferrage")
	log.Println("  vetlight.demo@   — véto light · write_notes sur Spirit")
	log.Printf("Confirm email : http://localhost:3002/confirm-email?token=%s", demoEmailConfirmToken)
	log.Printf("Reset password: http://localhost:3002/reset-password?token=%s", demoPasswordResetToken)
}

func seedProfilesTeamModules(ctx context.Context, pool *pgxpool.Pool, st *store.Store) error {
	rows, err := pool.Query(ctx, `SELECT id::text FROM identity.users`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := st.EnsureUserProfiles(ctx, id); err != nil {
			return fmt.Errorf("profiles %s: %w", id, err)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// VetPlus équipe : référence + collègue + assistant + secrétaire
	var practiceID, vetDemoID string
	err = pool.QueryRow(ctx, `
		SELECT p.id::text, u.id::text FROM practice.practices p
		JOIN identity.users u ON u.practice_id = p.id AND u.role = 'vet' AND u.email = 'vet.demo@petsfollow.test'
		LIMIT 1`).Scan(&practiceID, &vetDemoID)
	if err == nil {
		_ = st.EnsureReferenceTeamMembership(ctx, practiceID, vetDemoID)
		type member struct {
			email, name string
			role        store.TeamRole
		}
		for _, m := range []member{
			{"vet.colleague@petsfollow.test", "Dr Collègue VetPlus", store.TeamRoleVet},
			{"vet.assist@petsfollow.test", "Camille Assistante", store.TeamRoleAssistant},
			{"secretary.demo@petsfollow.test", "Sophie Secrétariat", store.TeamRoleSecretary},
		} {
			if _, err2 := st.InviteTeamMember(ctx, practiceID, vetDemoID, store.InviteTeamMemberInput{
				Email: m.email, FullName: m.name, Password: passwordVet, TeamRole: m.role,
			}); err2 != nil {
				log.Printf("seed team member %s: %v", m.email, err2)
			}
		}
	}

	// Modules ON pour client.demo
	var clientDemoID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email='client.demo@petsfollow.test'`).Scan(&clientDemoID); err == nil {
		_, _ = st.UpdateFeatureModules(ctx, clientDemoID, store.FeatureModules{
			ModuleCarePlus: true, ModuleHorse: true, ModuleKennel: true, ModuleFamily: true,
		})
	}

	// Horse contacts / competitions on Spirit
	var spiritID string
	err = pool.QueryRow(ctx, `
		SELECT p.id::text FROM pets.pets p
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE u.email='client.demo@petsfollow.test' AND p.name='Spirit'`).Scan(&spiritID)
	if err == nil {
		var ownerID string
		_ = pool.QueryRow(ctx, `SELECT owner_user_id::text FROM pets.pets WHERE id=$1`, spiritID).Scan(&ownerID)
		_, _ = pool.Exec(ctx, `
			INSERT INTO care.professional_contacts (id, pet_id, owner_user_id, role, full_name, phone, notes)
			SELECT $1, $2, $3, 'farrier', 'Marc Ferrier', '+32470000001', 'Démo seed'
			WHERE NOT EXISTS (SELECT 1 FROM care.professional_contacts WHERE pet_id=$2 AND full_name='Marc Ferrier')`,
			uuid.NewString(), spiritID, ownerID)
		_, _ = pool.Exec(ctx, `
			INSERT INTO care.competitions (id, pet_id, owner_user_id, event_date, title, location, result)
			SELECT $1, $2, $3, CURRENT_DATE - 30, 'CSO régional', 'Wavre', 'Clear round'
			WHERE NOT EXISTS (SELECT 1 FROM care.competitions WHERE pet_id=$2 AND title='CSO régional')`,
			uuid.NewString(), spiritID, ownerID)
		_, _ = pool.Exec(ctx, `
			UPDATE visits.visits SET lat = 50.8503, lng = 4.3517, address_text = COALESCE(NULLIF(address_text,''), 'Écurie démo — Bruxelles')
			WHERE pet_id = $1 AND notes LIKE '%démo care_pro%'`, spiritID)
	}

	log.Println("Équipe VetPlus : vet.colleague@ / vet.assist@ / secretary.demo@ (mdp véto)")
	log.Println("Modules UI ON : client.demo (care+/horse/kennel/family)")
	return nil
}
