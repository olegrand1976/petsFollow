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

// MassOptions tunes volumes for seed-mass (zero = defaults).
type MassOptions struct {
	PracticeCount      int // default 20
	ClientsPerPractice int // default 15 → ~300 clients
	OrphanClients      int // default 25 (pas de lien cabinet)
	ExtraDemoClients   int // default 36 (répartis VetPlus / Parc / Lyon)
	CareProCount       int // default 8 (farrier / vet_light)
}

func (o MassOptions) withDefaults() MassOptions {
	if o.PracticeCount <= 0 {
		o.PracticeCount = 20
	}
	if o.ClientsPerPractice <= 0 {
		o.ClientsPerPractice = 15
	}
	if o.OrphanClients <= 0 {
		o.OrphanClients = 25
	}
	if o.ExtraDemoClients <= 0 {
		o.ExtraDemoClients = 36
	}
	if o.CareProCount <= 0 {
		o.CareProCount = 8
	}
	return o
}

type massCity struct {
	name, postal string
	lat, lng     float64
	countryHint  string // FR / BE
}

var massCities = []massCity{
	{"Bruxelles", "1000", 50.8503, 4.3517, "BE"},
	{"Liège", "4000", 50.6326, 5.5797, "BE"},
	{"Gand", "9000", 51.0543, 3.7174, "BE"},
	{"Anvers", "2000", 51.2194, 4.4025, "BE"},
	{"Namur", "5000", 50.4674, 4.8719, "BE"},
	{"Charleroi", "6000", 50.4108, 4.4446, "BE"},
	{"Mons", "7000", 50.4542, 3.9523, "BE"},
	{"Bruges", "8000", 51.2093, 3.2247, "BE"},
	{"Lille", "59000", 50.6292, 3.0573, "FR"},
	{"Lyon", "69001", 45.7640, 4.8357, "FR"},
	{"Nantes", "44000", 47.2184, -1.5536, "FR"},
	{"Bordeaux", "33000", 44.8378, -0.5792, "FR"},
	{"Toulouse", "31000", 43.6047, 1.4442, "FR"},
	{"Strasbourg", "67000", 48.5734, 7.7521, "FR"},
	{"Rennes", "35000", 48.1173, -1.6778, "FR"},
	{"Nice", "06000", 43.7102, 7.2620, "FR"},
	{"Rouen", "76000", 49.4431, 1.0993, "FR"},
	{"Dijon", "21000", 47.3220, 5.0415, "FR"},
	{"Amiens", "80000", 49.8941, 2.2958, "FR"},
	{"Reims", "51100", 49.2583, 4.0317, "FR"},
}

var massFirstNames = []string{
	"Alice", "Bruno", "Clara", "David", "Emma", "Farid", "Grace", "Hugo",
	"Inès", "Jules", "Karen", "Louis", "Maya", "Noah", "Olga", "Paul",
	"Quinn", "Rita", "Samir", "Tina", "Ulysse", "Vera", "Wade", "Yara", "Zoé",
}

var massLastNames = []string{
	"Martin", "Bernard", "Dubois", "Laurent", "Petit", "Moreau", "Lefebvre",
	"Garcia", "Roux", "Fournier", "Girard", "Bonnet", "Dupont", "Lambert",
	"Fontaine", "Rousseau", "Vincent", "Muller", "Lefevre", "Faure",
}

var massPetNames = []string{
	"Rex", "Luna", "Max", "Bella", "Charlie", "Nala", "Rocky", "Milo",
	"Coco", "Oscar", "Ruby", "Simba", "Daisy", "Thor", "Lily", "Buddy",
	"Moka", "Pixel", "Shadow", "Kiwi", "Spirit", "Athena", "Jazz", "Pepper",
}

var massSpecies = []struct {
	species, breed string
	weight         float64
}{
	{"dog", "Labrador", 28},
	{"dog", "Berger", 32},
	{"dog", "Caniche", 8},
	{"cat", "Européen", 4.5},
	{"cat", "Siamois", 4},
	{"cat", "Maine Coon", 6.5},
	{"horse", "Selle Français", 520},
	{"horse", "Poney", 280},
	{"bird", "Perruche", 0.04},
	{"rabbit", "Nain", 1.5},
}

// RunMass densifie la DB après `make seed` (additif, emails mass.*@petsfollow.test).
// Conserve les commerciaux démo ; rattache les nouveaux vétos à Camille / Alex.
func RunMass(ctx context.Context, pool *pgxpool.Pool) error {
	return RunMassWithOptions(ctx, pool, MassOptions{})
}

// RunMassWithOptions same as RunMass with tunable volumes (tests).
func RunMassWithOptions(ctx context.Context, pool *pgxpool.Pool, opts MassOptions) error {
	opts = opts.withDefaults()
	if err := refuseSeedUnlessSeedableEnv("seed-mass"); err != nil {
		return err
	}

	var marker int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.users WHERE email = 'mass.vet.01@petsfollow.test'`).Scan(&marker); err != nil {
		return err
	}
	if marker > 0 {
		log.Println("seed-mass: déjà présent (mass.vet.01@) — skip. Relancer `make seed` puis `make seed-mass` pour régénérer.")
		return nil
	}

	var camilleID, alexID string
	err := pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users
		WHERE email = 'commercial.demo@petsfollow.test' AND role = 'commercial'`).Scan(&camilleID)
	if err != nil {
		return fmt.Errorf("seed-mass: commercial.demo requis (lancer make seed d'abord): %w", err)
	}
	err = pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users
		WHERE email = 'commercial.demo2@petsfollow.test' AND role = 'commercial'`).Scan(&alexID)
	if err != nil {
		return fmt.Errorf("seed-mass: commercial.demo2 requis (lancer make seed d'abord): %w", err)
	}

	vetHash, err := bcrypt.GenerateFromPassword([]byte(passwordVet), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	clientHash, err := bcrypt.GenerateFromPassword([]byte(passwordClient), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	careHash, err := bcrypt.GenerateFromPassword([]byte(passwordCarePro), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	type massPractice struct {
		practiceID, vetID, vetEmail, commercialID string
		clientIDs                                 []string
		petIDs                                    []string
	}
	practices := make([]massPractice, 0, opts.PracticeCount)
	stats := struct {
		vets, clients, pets, orphans, extras, carePros, links int
	}{}

	for i := 1; i <= opts.PracticeCount; i++ {
		city := massCities[(i-1)%len(massCities)]
		vetEmail := fmt.Sprintf("mass.vet.%02d@petsfollow.test", i)
		vetName := fmt.Sprintf("Dr %s %s", massFirstNames[(i*3)%len(massFirstNames)], massLastNames[i%len(massLastNames)])
		practiceName := fmt.Sprintf("Cabinet Mass %02d — %s", i, city.name)
		commercialID := camilleID
		if i > opts.PracticeCount/2 {
			commercialID = alexID
		}

		practiceID := uuid.NewString()
		vetID := uuid.NewString()
		phone := fmt.Sprintf("0%d%08d", 2+(i%7), 10000000+i*137)
		addr := fmt.Sprintf("%d rue des Animaux", 10+i)

		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.practices (
				id, name, phone, contact_email, address_line1, city, postal_code, website, profile_completed_at,
				company_legal_name, vat_number, company_number, legal_form, billing_same_as_practice,
				payout_iban, payout_bic, payout_account_holder
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, NOW(),
				$9, 'BE0123456789', '0123.456.789', 'srl', TRUE,
				'BE68539007547034', 'GEBABEBB', $10
			)`,
			practiceID, practiceName, phone, vetEmail, addr, city.name, city.postal,
			fmt.Sprintf("https://mass%02d.petsfollow.test", i),
			practiceName+" SRL", vetName); err != nil {
			return fmt.Errorf("mass practice %d: %w", i, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO identity.users (
				id, email, password_hash, full_name, role, practice_id, email_verified_at, assigned_commercial_id
			) VALUES ($1, $2, $3, $4, 'vet', $5, NOW(), $6::uuid)`,
			vetID, vetEmail, string(vetHash), vetName, practiceID, commercialID); err != nil {
			return fmt.Errorf("mass vet %d: %w", i, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
			VALUES ($1, true, true)`, vetID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
			VALUES ($1, $2, 'available', NULL)`, vetID, practiceID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.ai_cr_modules (
				practice_id, status, activated_at, trial_ends_at, activated_by_user_id,
				price_plan, baseline_minutes_per_cr, hourly_cost_cents, updated_at
			) VALUES ($1, 'trial', NOW(), NOW() + INTERVAL '90 days', $2, 'monthly_39', 10, 8000, NOW())
			ON CONFLICT (practice_id) DO NOTHING`, practiceID, vetID); err != nil {
			return err
		}

		mp := massPractice{
			practiceID: practiceID, vetID: vetID, vetEmail: vetEmail, commercialID: commercialID,
		}
		for j := 1; j <= opts.ClientsPerPractice; j++ {
			clientIdx := (i-1)*opts.ClientsPerPractice + j
			clientID, petIDs, err := insertMassClient(ctx, tx, massClientInsert{
				index:       clientIdx,
				email:       fmt.Sprintf("mass.client.%03d@petsfollow.test", clientIdx),
				practiceID:  practiceID,
				vetID:       vetID,
				clientHash:  string(clientHash),
				linkPractice: true,
				config:      massClientConfig(clientIdx),
			})
			if err != nil {
				return fmt.Errorf("mass client %d: %w", clientIdx, err)
			}
			mp.clientIDs = append(mp.clientIDs, clientID)
			mp.petIDs = append(mp.petIDs, petIDs...)
			stats.clients++
			stats.pets += len(petIDs)
			stats.links++
		}
		practices = append(practices, mp)
		stats.vets++
	}

	// Orphans — clients sans cabinet (onboarding / discovery).
	orphanStart := opts.PracticeCount*opts.ClientsPerPractice + 1
	for j := 0; j < opts.OrphanClients; j++ {
		idx := orphanStart + j
		_, petIDs, err := insertMassClient(ctx, tx, massClientInsert{
			index:        idx,
			email:        fmt.Sprintf("mass.client.%03d@petsfollow.test", idx),
			clientHash:   string(clientHash),
			linkPractice: false,
			config:       massClientConfig(idx),
		})
		if err != nil {
			return fmt.Errorf("mass orphan %d: %w", idx, err)
		}
		stats.clients++
		stats.orphans++
		stats.pets += len(petIDs)
	}

	// Clients supplémentaires sur cabinets démo P0 (densité filiation / kanban).
	demoTargets := []struct {
		vetEmail string
	}{
		{"vet.demo@petsfollow.test"},
		{"vet.parc@petsfollow.test"},
		{"vet.lyon@petsfollow.test"},
	}
	extraStart := orphanStart + opts.OrphanClients
	for e := 0; e < opts.ExtraDemoClients; e++ {
		target := demoTargets[e%len(demoTargets)]
		var practiceID, vetID string
		err := tx.QueryRow(ctx, `
			SELECT practice_id::text, id::text FROM identity.users
			WHERE email = $1 AND role = 'vet'`, target.vetEmail).Scan(&practiceID, &vetID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		idx := extraStart + e
		_, petIDs, err := insertMassClient(ctx, tx, massClientInsert{
			index:        idx,
			email:        fmt.Sprintf("mass.client.%03d@petsfollow.test", idx),
			practiceID:   practiceID,
			vetID:        vetID,
			clientHash:   string(clientHash),
			linkPractice: true,
			config:       massClientConfig(idx),
		})
		if err != nil {
			return fmt.Errorf("mass extra %d: %w", idx, err)
		}
		stats.clients++
		stats.extras++
		stats.pets += len(petIDs)
		stats.links++
	}

	// Liens secondaires multi-cabinet (~1 practice sur 3).
	for i, mp := range practices {
		if i%3 != 0 || len(mp.clientIDs) == 0 {
			continue
		}
		other := practices[(i+3)%len(practices)]
		if other.practiceID == mp.practiceID {
			continue
		}
		clientID := mp.clientIDs[0]
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
			uuid.NewString(), other.practiceID, clientID, other.vetID); err != nil {
			return err
		}
		stats.links++
	}

	// Pending link requests (quelques orphelins → cabinets mass).
	for j := 0; j < 8 && j < opts.OrphanClients && len(practices) > 0; j++ {
		idx := orphanStart + j
		email := fmt.Sprintf("mass.client.%03d@petsfollow.test", idx)
		var clientID string
		if err := tx.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, email).Scan(&clientID); err != nil {
			continue
		}
		mp := practices[j%len(practices)]
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.client_vet_link_requests (id, client_user_id, practice_id, vet_user_id, status)
			VALUES ($1, $2, $3, $4, 'pending')
			ON CONFLICT (client_user_id, practice_id) DO UPDATE SET
				vet_user_id = EXCLUDED.vet_user_id,
				status = 'pending',
				updated_at = NOW()`,
			uuid.NewString(), clientID, mp.practiceID, mp.vetID); err != nil {
			return err
		}
	}

	// Care pros farrier / vet_light + grants.
	careProIDs := make([]string, 0, opts.CareProCount)
	for i := 1; i <= opts.CareProCount; i++ {
		spec := string(kernel.SpecialtyFarrier)
		if i%2 == 0 {
			spec = string(kernel.SpecialtyVetLight)
		}
		email := fmt.Sprintf("mass.care.%02d@petsfollow.test", i)
		name := fmt.Sprintf("%s %s", massFirstNames[(i*5)%len(massFirstNames)], "Care")
		var uid string
		if err := tx.QueryRow(ctx, `
			INSERT INTO identity.users (
				id, email, password_hash, full_name, role, practice_id,
				email_verified_at, professional_specialty, preferred_locale
			) VALUES ($1, $2, $3, $4, 'care_pro', NULL, NOW(), $5, 'fr')
			RETURNING id::text`,
			uuid.NewString(), email, string(careHash), name, spec).Scan(&uid); err != nil {
			return fmt.Errorf("mass care_pro %d: %w", i, err)
		}
		careProIDs = append(careProIDs, uid)
		stats.carePros++
	}

	// Grants : pet_access + client_access (commissions care_pro) sur pets mass.
	allPetOwner := make([]struct{ petID, ownerID string }, 0)
	rows, err := tx.Query(ctx, `
		SELECT p.id::text, p.owner_user_id::text
		FROM pets.pets p
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE u.email LIKE 'mass.client.%'
		ORDER BY p.created_at
		LIMIT 120`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var petID, ownerID string
		if err := rows.Scan(&petID, &ownerID); err != nil {
			rows.Close()
			return err
		}
		allPetOwner = append(allPetOwner, struct{ petID, ownerID string }{petID, ownerID})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for i, cp := range careProIDs {
		for k := 0; k < 3 && len(allPetOwner) > 0; k++ {
			po := allPetOwner[(i*3+k)%len(allPetOwner)]
			if _, err := tx.Exec(ctx, `
				INSERT INTO pets.pet_access (id, pet_id, grantee_user_id, permission, granted_by_user_id)
				VALUES ($1, $2, $3, 'write_notes', $4)
				ON CONFLICT (pet_id, grantee_user_id) DO UPDATE SET permission = EXCLUDED.permission`,
				uuid.NewString(), po.petID, cp, po.ownerID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO practice.client_access (id, client_user_id, grantee_user_id, permission, granted_by_user_id)
				VALUES ($1, $2, $3, 'write_notes', $4)
				ON CONFLICT (client_user_id, grantee_user_id) DO UPDATE SET permission = EXCLUDED.permission`,
				uuid.NewString(), po.ownerID, cp, po.ownerID); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	st := store.New(pool)

	// Profiles pour tous les users mass.
	profRows, err := pool.Query(ctx, `
		SELECT id::text FROM identity.users
		WHERE email LIKE 'mass.%@petsfollow.test'`)
	if err != nil {
		return err
	}
	defer profRows.Close()
	for profRows.Next() {
		var id string
		if err := profRows.Scan(&id); err != nil {
			return err
		}
		if err := st.EnsureUserProfiles(ctx, id); err != nil {
			return fmt.Errorf("mass profiles %s: %w", id, err)
		}
	}
	if err := profRows.Err(); err != nil {
		return err
	}

	// Filiation audit (échantillon) — assignations véto + liens clients.
	for i, mp := range practices {
		_ = st.RecordFiliationEvent(ctx, store.FiliationEventInput{
			EventType:        store.FiliationEventVetAssigned,
			CommercialUserID: mp.commercialID,
			VetUserID:        mp.vetID,
			PracticeID:       mp.practiceID,
			ActorUserID:      mp.commercialID,
			Meta:             map[string]any{"source": "seed-mass", "vet": mp.vetEmail},
		})
		if i%2 == 0 && len(mp.clientIDs) > 0 {
			st.RecordPracticeClientLinkedEvent(ctx, mp.practiceID, mp.clientIDs[0], mp.vetID, mp.vetID, map[string]any{
				"source": "seed-mass",
			})
		}
	}

	// Schedules pour cabinets mass.
	year := time.Now().Year()
	slots := []store.ScheduleSlot{
		{Weekday: 1, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 1, StartTime: "14:00", EndTime: "18:00"},
		{Weekday: 2, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 3, StartTime: "09:00", EndTime: "12:00"},
		{Weekday: 4, StartTime: "14:00", EndTime: "18:00"},
		{Weekday: 5, StartTime: "09:00", EndTime: "12:00"},
	}
	for _, mp := range practices {
		if _, err := st.PutVetSchedule(ctx, mp.practiceID, true, 30, &year, slots); err != nil {
			return fmt.Errorf("mass schedule %s: %w", mp.vetEmail, err)
		}
	}

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

	// Mots de passe démo volontairement absents des logs — voir AGENTS.md.
	log.Printf("--- seed-mass OK (mots de passe : AGENTS.md) ---")
	log.Printf("  vétos mass     : %d (emails mass.vet.NN@)", stats.vets)
	log.Printf("  clients mass   : %d (dont %d orphelins, %d sur cabinets démo)", stats.clients, stats.orphans, stats.extras)
	log.Printf("  animaux        : %d", stats.pets)
	log.Printf("  care_pros mass : %d (mass.care.NN@)", stats.carePros)
	log.Printf("  liens practice : ~%d", stats.links)
	log.Printf("  commerciaux    : inchangés (Camille / Alex / Bérénice) — nouveaux vétos rattachés 50/50")
	log.Printf("  commissions    : AccrueAll* (véto + commercial + care_pro via client_access)")
	return nil
}

type massClientCfg struct {
	petCount      int
	pendingPay    bool
	locale        string
	forceEmpty    bool
	forceMultiMax int // if >0, pet count = this
}

func massClientConfig(index int) massClientCfg {
	cfg := massClientCfg{locale: "fr", petCount: 1}
	switch index % 20 {
	case 0, 1, 2:
		cfg.petCount = 1 // monthly via planForPet
	case 3, 4:
		cfg.petCount = 1
	case 5, 6:
		cfg.petCount = 1
	case 7, 8, 9, 10:
		cfg.petCount = 2 + (index % 3) // 2–4
	case 11, 12:
		cfg.forceEmpty = true
	case 13, 14:
		cfg.petCount = 1
		cfg.pendingPay = true
	case 15:
		cfg.forceMultiMax = 4
	case 16:
		cfg.locale = "nl"
		cfg.petCount = 2
	case 17:
		cfg.locale = "en"
		cfg.petCount = 1
	case 18:
		cfg.locale = "es"
		cfg.petCount = 2
	case 19:
		cfg.petCount = 3
	}
	if cfg.forceMultiMax > 0 {
		cfg.petCount = cfg.forceMultiMax
	}
	if cfg.forceEmpty {
		cfg.petCount = 0
	}
	return cfg
}

type massClientInsert struct {
	index        int
	email        string
	practiceID   string
	vetID        string
	clientHash   string
	linkPractice bool
	config       massClientCfg
}

func insertMassClient(ctx context.Context, tx pgx.Tx, in massClientInsert) (clientID string, petIDs []string, err error) {
	clientID = uuid.NewString()
	fn := massFirstNames[in.index%len(massFirstNames)]
	ln := massLastNames[(in.index/3)%len(massLastNames)]
	fullName := fmt.Sprintf("%s %s", fn, ln)

	var practiceArg any
	if in.linkPractice && in.practiceID != "" {
		practiceArg = in.practiceID
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, email_verified_at, preferred_locale
		) VALUES ($1, $2, $3, $4, 'client', $5, NOW(), $6)`,
		clientID, in.email, in.clientHash, fullName, practiceArg, in.config.locale); err != nil {
		return "", nil, err
	}

	if in.linkPractice && in.practiceID != "" && in.vetID != "" {
		if _, err = tx.Exec(ctx, `
			INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
			VALUES ($1, $2, $3, $4)`,
			uuid.NewString(), in.practiceID, clientID, in.vetID); err != nil {
			return "", nil, err
		}
	}

	for p := 0; p < in.config.petCount; p++ {
		petID, err := insertMassPet(ctx, tx, massPetInsert{
			index:        in.index*10 + p,
			clientID:     clientID,
			practiceID:   in.practiceID,
			pendingPay:   in.config.pendingPay && p == 0,
			planOverride: planForMassPet(in.index, p),
		})
		if err != nil {
			return "", nil, err
		}
		petIDs = append(petIDs, petID)
	}

	if len(petIDs) > 0 {
		if _, err = tx.Exec(ctx, `
			INSERT INTO billing.stripe_customers (user_id, stripe_customer_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id`,
			clientID, billing.MockCustomerID(clientID)); err != nil {
			return "", nil, err
		}
	}

	// Thread messaging si lié à un cabinet + au moins un pet.
	if in.linkPractice && in.practiceID != "" && in.vetID != "" && len(petIDs) > 0 {
		threadID := uuid.NewString()
		if _, err = tx.Exec(ctx, `
			INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
			VALUES ($1, $2, $3, $4, $5)`,
			threadID, in.practiceID, clientID, in.vetID, petIDs[0]); err != nil {
			return "", nil, err
		}
		// 1 message sur ~3 clients
		if in.index%3 == 0 {
			if _, err = tx.Exec(ctx, `
				INSERT INTO messaging.messages (id, thread_id, sender_user_id, body, read_at, created_at)
				VALUES ($1, $2, $3, $4, NULL, NOW() - INTERVAL '2 days')`,
				uuid.NewString(), threadID, clientID, "Bonjour, petit suivi démo mass seed."); err != nil {
				return "", nil, err
			}
		}
	}

	return clientID, petIDs, nil
}

func planForMassPet(clientIndex, petIndex int) billing.PlanCode {
	switch (clientIndex + petIndex) % 5 {
	case 0, 1:
		return billing.PlanMonthly
	case 2, 3:
		return billing.PlanAnnual
	default:
		return billing.PlanTriennial
	}
}

type massPetInsert struct {
	index        int
	clientID     string
	practiceID   string
	pendingPay   bool
	planOverride billing.PlanCode
}

func insertMassPet(ctx context.Context, tx pgx.Tx, in massPetInsert) (string, error) {
	petID := uuid.NewString()
	spec := massSpecies[in.index%len(massSpecies)]
	name := massPetNames[in.index%len(massPetNames)]
	if in.index/len(massPetNames) > 0 {
		name = fmt.Sprintf("%s-%d", name, in.index/len(massPetNames)+1)
	}

	paymentStatus := "active"
	entStatus := billing.StatusActive
	if in.pendingPay {
		paymentStatus = "pending_payment"
		entStatus = billing.StatusPending
	}

	var practiceArg any
	if strings.TrimSpace(in.practiceID) != "" {
		practiceArg = in.practiceID
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, weight_kg, payment_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		petID, practiceArg, in.clientID, name, spec.species, spec.breed, spec.weight, paymentStatus); err != nil {
		return "", err
	}

	plan, err := billing.GetPlan(in.planOverride)
	if err != nil {
		return "", err
	}
	now := time.Now()
	var validFrom, validUntil *time.Time
	if entStatus.AllowsAccess() || entStatus == billing.StatusPending {
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
	if _, err := tx.Exec(ctx, `
		INSERT INTO billing.pet_entitlements (
			id, pet_id, owner_user_id, plan_code, billing_mode, status, amount_cents, currency, valid_from, valid_until
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'eur', $8, $9)`,
		uuid.NewString(), petID, in.clientID, in.planOverride, billing.ModeSubscription, entStatus,
		plan.AmountCents, validFrom, validUntil); err != nil {
		return "", err
	}

	// Poids + un rappel care sur une partie des pets actifs (practice requis si colonne encore NOT NULL).
	if !in.pendingPay && in.index%4 == 0 && practiceArg != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO pets.weight_readings (
				id, pet_id, owner_user_id, author_user_id, practice_id, weight_kg, comment, recorded_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() - INTERVAL '10 days')`,
			uuid.NewString(), petID, in.clientID, in.clientID, practiceArg, spec.weight, "contrôlée mass"); err != nil {
			return "", err
		}
	}
	if !in.pendingPay && practiceArg != nil && in.index%5 == 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status, updated_at)
			VALUES ($1, $2, $3, 'vaccination', 'Vaccination', NOW() + INTERVAL '20 days', 'pending', NOW())`,
			uuid.NewString(), petID, practiceArg); err != nil {
			return "", err
		}
	}
	return petID, nil
}
