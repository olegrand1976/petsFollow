package store

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// MaxHeartRateCommentLen is the max rune length stored on a validated session.
const MaxHeartRateCommentLen = 500

var (
	ErrNotFound             = errors.New("not found")
	ErrValidation           = errors.New("validation")
	ErrForbidden            = errors.New("forbidden")
	ErrConflict             = errors.New("conflict")
	ErrInventoryIncomplete  = errors.New("inventory_incomplete")
	ErrDiagnosticsTooLarge  = errors.New("diagnostics too large")
	ErrSelfReferral         = errors.New("self referral")
	ErrAdvisoryLockBusy     = errors.New("advisory lock busy")
)

// StagingSeedLockKey is the session advisory lock for admin/CLI staging re-seed.
const StagingSeedLockKey int64 = 0x70667365656401 // "pfseed\x01"

type User struct {
	ID                    string
	Email                 string
	PasswordHash          string
	FullName              string
	Role                  kernel.Role
	PracticeID            string
	EmailVerifiedAt       *time.Time
	GoogleSub             string
	AuthProvider          string
	TOTPSecret            string
	TOTPEnabled           bool
	PreferredLocale       string
	AvatarURL             string
	MustChangePassword    bool
	ProfessionalSpecialty string
	ContactPhone          string
	TermsAcceptedAt       *time.Time
}

type Practice struct {
	ID   string
	Name string
}

type Pet struct {
	ID            string    `json:"id"`
	PracticeID    string    `json:"practiceId"`
	OwnerUserID   string    `json:"ownerUserId"`
	Name          string    `json:"name"`
	Species       string    `json:"species"`
	Breed         string    `json:"breed"`
	BirthDate     *time.Time `json:"birthDate,omitempty"`
	WeightKg      *float64  `json:"weightKg,omitempty"`
	PhotoURL      string    `json:"photoUrl"`
	PaymentStatus string    `json:"paymentStatus"`
	LitterTag     string    `json:"litterTag,omitempty"`
	MicrochipNumber        string `json:"microchipNumber,omitempty"`
	HealthBookNumber       string `json:"healthBookNumber,omitempty"`
	HealthBookPDFURL       string `json:"healthBookPdfUrl,omitempty"` // legacy; never a public media URL
	HealthBookPDFAttached  bool   `json:"healthBookPdfAttached,omitempty"`
	HealthBookPDFObjectKey string `json:"-"`
	// FoodChainStatus: companion | food_producing | excluded_from_food_chain (DAF / médicaments).
	FoodChainStatus string `json:"foodChainStatus,omitempty"`
	// DomicileLocation: écurie / lieu de détention (équidés, rente, camélidés).
	DomicileLocation string `json:"domicileLocation,omitempty"`
	// Lifecycle status dates (adopted / sold / deceased) — DATE, all species.
	AdoptedAt  *time.Time `json:"adoptedAt,omitempty"`
	SoldAt     *time.Time `json:"soldAt,omitempty"`
	DeceasedAt *time.Time `json:"deceasedAt,omitempty"`
	HeartrateDurationsSec []int `json:"heartrateDurationsSec,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Entitlement   *Entitlement `json:"entitlement,omitempty"`
	// Permission is set on list responses for care_pro / shared client access (read | write_notes | full).
	Permission string `json:"permission,omitempty"`
}

type HeartRateSession struct {
	ID          string               `json:"id"`
	PetID       string               `json:"petId"`
	OwnerUserID string               `json:"ownerUserId"`
	PracticeID  string               `json:"practiceId"`
	Status      kernel.SessionStatus `json:"status"`
	TapCount    int                  `json:"tapCount"`
	DurationSec int                  `json:"durationSec"`
	BPM         *int                 `json:"bpm,omitempty"`
	IsAlert     bool                 `json:"isAlert"`
	StartedAt   time.Time            `json:"startedAt"`
	EndedAt     *time.Time           `json:"endedAt,omitempty"`
	ValidatedAt *time.Time           `json:"validatedAt,omitempty"`
	VetSeenAt   *time.Time           `json:"vetSeenAt,omitempty"`
	Comment     *string              `json:"comment,omitempty"`
	IsNew       bool                 `json:"isNew"`
}

const heartRateSelectCols = `
	id::text, pet_id::text, owner_user_id::text, practice_id::text, status, tap_count, duration_sec,
	bpm, is_alert, started_at, ended_at, validated_at, vet_seen_at, comment`

func heartRateScanDest(sess *HeartRateSession) []any {
	return []any{
		&sess.ID, &sess.PetID, &sess.OwnerUserID, &sess.PracticeID, &sess.Status,
		&sess.TapCount, &sess.DurationSec, &sess.BPM, &sess.IsAlert,
		&sess.StartedAt, &sess.EndedAt, &sess.ValidatedAt, &sess.VetSeenAt, &sess.Comment,
	}
}

// NormalizeHeartRateComment trims, truncates to MaxHeartRateCommentLen, and returns nil when empty.
func NormalizeHeartRateComment(raw *string) *string {
	if raw == nil {
		return nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil
	}
	if utf8.RuneCountInString(s) > MaxHeartRateCommentLen {
		s = string([]rune(s)[:MaxHeartRateCommentLen])
	}
	return &s
}

func decorateHeartRateSession(sess *HeartRateSession) {
	sess.IsNew = sess.Status == kernel.SessionValidated && sess.VetSeenAt == nil
}

type Thread struct {
	ID           string `json:"id"`
	PracticeID   string `json:"practiceId"`
	ClientUserID string `json:"clientUserId"`
	VetUserID    string `json:"vetUserId"`
	PetID        string `json:"petId"`
}

type Message struct {
	ID           string     `json:"id"`
	ThreadID     string     `json:"threadId"`
	SenderUserID string     `json:"senderUserId"`
	Body         string     `json:"body"`
	MediaURL     string     `json:"mediaUrl,omitempty"`
	MediaType    string     `json:"mediaType,omitempty"`
	ReadAt       *time.Time `json:"readAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type DossierEvent struct {
	ID           string    `json:"id"`
	PetID        string    `json:"petId"`
	AuthorUserID string    `json:"authorUserId"`
	EventType    string    `json:"eventType"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
}

type TimelineItem struct {
	ID        string              `json:"id"`
	Type      kernel.TimelineType `json:"type"`
	Title     string              `json:"title"`
	Body      string              `json:"body"`
	CreatedAt time.Time           `json:"createdAt"`
	Meta      map[string]any      `json:"meta,omitempty"`
}

type ClientSummary struct {
	UserID                 string `json:"userId"`
	Email                  string `json:"email"`
	FullName               string `json:"fullName"`
	FirstName              string `json:"firstName,omitempty"`
	LastName               string `json:"lastName,omitempty"`
	AvatarURL              string `json:"avatarUrl,omitempty"`
	ContactPhone           string `json:"contactPhone,omitempty"`
	Address                string `json:"address,omitempty"`
	NationalRegistryNumber string `json:"nationalRegistryNumber,omitempty"`
	PetCount               int    `json:"petCount"`
}

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Pool exposes the underlying pgx pool (admin staging seed, ops tooling).
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	var passwordHash *string
	err := row.Scan(
		&u.ID, &u.Email, &passwordHash, &u.FullName, &u.Role, &u.PracticeID, &u.EmailVerifiedAt,
		&u.GoogleSub, &u.AuthProvider, &u.TOTPSecret, &u.TOTPEnabled, &u.PreferredLocale, &u.AvatarURL,
		&u.MustChangePassword, &u.ProfessionalSpecialty, &u.ContactPhone, &u.TermsAcceptedAt,
	)
	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	return u, err
}

const userSelectCols = `
	id::text, email, password_hash, full_name, role, COALESCE(practice_id::text,''), email_verified_at,
	COALESCE(google_sub,''), COALESCE(auth_provider,'password'), COALESCE(totp_secret,''), totp_enabled,
	COALESCE(preferred_locale,'fr'), COALESCE(avatar_url,''), must_change_password,
	COALESCE(professional_specialty,''), COALESCE(contact_phone,''), terms_accepted_at`

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := scanUser(s.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+` FROM identity.users WHERE lower(email)=$1`, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+` FROM identity.users WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

const clientSummarySelect = `
		u.id::text, u.email, u.full_name,
		COALESCE(u.first_name,''), COALESCE(u.last_name,''),
		COALESCE(u.avatar_url,''), COALESCE(u.contact_phone,''),
		COALESCE(u.address,''), COALESCE(u.national_registry_number,''),
		COUNT(p.id)::int`

const clientSummaryGroupBy = `
		u.id, u.email, u.full_name, u.first_name, u.last_name,
		u.avatar_url, u.contact_phone, u.address, u.national_registry_number`

func scanClientSummary(scan func(dest ...any) error) (ClientSummary, error) {
	var c ClientSummary
	err := scan(
		&c.UserID, &c.Email, &c.FullName, &c.FirstName, &c.LastName,
		&c.AvatarURL, &c.ContactPhone, &c.Address, &c.NationalRegistryNumber, &c.PetCount,
	)
	return c, err
}

func (s *Store) ListClientsByPractice(ctx context.Context, practiceID string) ([]ClientSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+clientSummarySelect+`
		FROM practice.practice_clients pc
		JOIN identity.users u ON u.id = pc.client_user_id
		LEFT JOIN pets.pets p ON p.owner_user_id = u.id AND p.practice_id = pc.practice_id
		WHERE pc.practice_id = $1
		GROUP BY `+clientSummaryGroupBy+`
		ORDER BY u.full_name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ClientSummary, 0)
	for rows.Next() {
		c, err := scanClientSummary(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) UpdatePet(ctx context.Context, p Pet) error {
	foodChain := p.FoodChainStatus
	if foodChain == "" {
		foodChain = "companion"
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE pets.pets SET name=$2, species=$3, breed=$4, birth_date=$5, weight_kg=$6, photo_url=$7, litter_tag=$8,
			microchip_number=$9, health_book_number=$10, domicile_location=$11, food_chain_status=$12,
			adopted_at=$13, sold_at=$14, deceased_at=$15, updated_at=NOW()
		WHERE id=$1 AND owner_user_id=$16`,
		p.ID, p.Name, p.Species, p.Breed, p.BirthDate, p.WeightKg, p.PhotoURL, p.LitterTag,
		p.MicrochipNumber, p.HealthBookNumber, p.DomicileLocation, foodChain,
		p.AdoptedAt, p.SoldAt, p.DeceasedAt, p.OwnerUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPetLifecycleDates updates adopted/sold/deceased dates for a practice pet (vet clinical write).
func (s *Store) SetPetLifecycleDates(ctx context.Context, practiceID, petID string, adoptedAt, soldAt, deceasedAt *time.Time) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE pets.pets
		SET adopted_at = $3, sold_at = $4, deceased_at = $5, updated_at = NOW()
		WHERE id = $1 AND practice_id = $2`,
		petID, practiceID, adoptedAt, soldAt, deceasedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdatePetHealthBookPDF(ctx context.Context, petID, url, objectKey string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE pets.pets SET health_book_pdf_url=$2, health_book_pdf_object_key=$3, updated_at=NOW()
		WHERE id=$1`, petID, url, objectKey)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ClearPetHealthBookPDF(ctx context.Context, petID string) (previousObjectKey string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(health_book_pdf_object_key,'') FROM pets.pets WHERE id=$1`, petID).Scan(&previousObjectKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE pets.pets SET health_book_pdf_url='', health_book_pdf_object_key='', updated_at=NOW()
		WHERE id=$1`, petID)
	if err != nil {
		return "", err
	}
	if ct.RowsAffected() == 0 {
		return "", ErrNotFound
	}
	return previousObjectKey, nil
}

func (s *Store) GetPet(ctx context.Context, id string) (Pet, error) {
	var p Pet
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, COALESCE(practice_id::text,''), owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, COALESCE(litter_tag,''),
			COALESCE(microchip_number,''), COALESCE(health_book_number,''),
			COALESCE(health_book_pdf_url,''), COALESCE(health_book_pdf_object_key,''),
			COALESCE(food_chain_status,'companion'), COALESCE(domicile_location,''),
			adopted_at, sold_at, deceased_at, created_at
		FROM pets.pets WHERE id=$1`, id).Scan(
		&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed, &p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.LitterTag,
		&p.MicrochipNumber, &p.HealthBookNumber, &p.HealthBookPDFURL, &p.HealthBookPDFObjectKey,
		&p.FoodChainStatus, &p.DomicileLocation, &p.AdoptedAt, &p.SoldAt, &p.DeceasedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Pet{}, ErrNotFound
	}
	if err == nil {
		decoratePetHealthBook(&p)
		if ent, e := s.GetEntitlementByPetID(ctx, id); e == nil {
			p.Entitlement = &ent
		}
		if p.PracticeID != "" {
			if durations, e := s.GetPracticeHeartRateDurations(ctx, p.PracticeID); e == nil {
				p.HeartrateDurationsSec = durations
			}
		}
	}
	return p, err
}

func (s *Store) ListPetsByOwner(ctx context.Context, ownerID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(practice_id::text,''), owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, COALESCE(litter_tag,''),
			COALESCE(microchip_number,''), COALESCE(health_book_number,''),
			COALESCE(health_book_pdf_url,''), COALESCE(health_book_pdf_object_key,''),
			COALESCE(food_chain_status,'companion'), COALESCE(domicile_location,''),
			adopted_at, sold_at, deceased_at, created_at
		FROM pets.pets WHERE owner_user_id=$1 ORDER BY name`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPetsWithEntitlements(ctx, s, rows)
}

// ListPetsAccessibleToClient returns owned pets plus those granted via pet_access or client_access.
func (s *Store) ListPetsAccessibleToClient(ctx context.Context, clientUserID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, COALESCE(p.practice_id::text,''), p.owner_user_id::text, p.name, p.species,
			COALESCE(p.breed,''), p.birth_date, p.weight_kg, COALESCE(p.photo_url,''),
			COALESCE(p.payment_status,''), COALESCE(p.litter_tag,''),
			COALESCE(p.microchip_number,''), COALESCE(p.health_book_number,''),
			COALESCE(p.health_book_pdf_url,''), COALESCE(p.health_book_pdf_object_key,''),
			COALESCE(p.food_chain_status,'companion'), COALESCE(p.domicile_location,''),
			p.adopted_at, p.sold_at, p.deceased_at, p.created_at,
			COALESCE((
				SELECT CASE
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 3 THEN 'full'
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 2 THEN 'write_notes'
					ELSE 'read'
				END
				FROM (
					SELECT 'full'::text AS permission WHERE p.owner_user_id=$1
					UNION ALL
					SELECT pa.permission FROM pets.pet_access pa
					WHERE pa.pet_id=p.id AND pa.grantee_user_id=$1
						AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
					UNION ALL
					SELECT ca.permission FROM practice.client_access ca
					WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
						AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
				) x
			), 'read') AS permission
		FROM pets.pets p
		WHERE p.owner_user_id=$1
		OR EXISTS (
			SELECT 1 FROM pets.pet_access pa
			WHERE pa.pet_id=p.id AND pa.grantee_user_id=$1
				AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
		)
		OR EXISTS (
			SELECT 1 FROM practice.client_access ca
			WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
				AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
		)
		ORDER BY p.name`, clientUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Pet
	for rows.Next() {
		var p Pet
		if err := rows.Scan(
			&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed,
			&p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.LitterTag,
			&p.MicrochipNumber, &p.HealthBookNumber, &p.HealthBookPDFURL, &p.HealthBookPDFObjectKey,
			&p.FoodChainStatus, &p.DomicileLocation, &p.AdoptedAt, &p.SoldAt, &p.DeceasedAt, &p.CreatedAt,
			&p.Permission,
		); err != nil {
			return nil, err
		}
		decoratePetHealthBook(&p)
		out = append(out, p)
	}
	if out == nil {
		out = []Pet{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	durCache := map[string][]int{}
	for i := range out {
		if ent, e := s.GetEntitlementByPetID(ctx, out[i].ID); e == nil {
			out[i].Entitlement = &ent
		}
		attachPracticeHeartRateDurations(ctx, s, &out[i], durCache)
	}
	return out, nil
}

func (s *Store) ListPetsByClientForVet(ctx context.Context, practiceID, clientID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(practice_id::text,''), owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, COALESCE(litter_tag,''),
			COALESCE(microchip_number,''), COALESCE(health_book_number,''),
			COALESCE(health_book_pdf_url,''), COALESCE(health_book_pdf_object_key,''),
			COALESCE(food_chain_status,'companion'), COALESCE(domicile_location,''),
			adopted_at, sold_at, deceased_at, created_at
		FROM pets.pets WHERE practice_id=$1 AND owner_user_id=$2 ORDER BY name`, practiceID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPetsWithEntitlements(ctx, s, rows)
}

func scanPets(rows pgx.Rows) ([]Pet, error) {
	var out []Pet
	for rows.Next() {
		var p Pet
		if err := rows.Scan(
			&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed, &p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.LitterTag,
			&p.MicrochipNumber, &p.HealthBookNumber, &p.HealthBookPDFURL, &p.HealthBookPDFObjectKey,
			&p.FoodChainStatus, &p.DomicileLocation, &p.AdoptedAt, &p.SoldAt, &p.DeceasedAt, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		decoratePetHealthBook(&p)
		out = append(out, p)
	}
	return out, rows.Err()
}

// decoratePetHealthBook exposes attachment via flag only — never a public media URL (PHI).
func decoratePetHealthBook(p *Pet) {
	if p == nil {
		return
	}
	attached := strings.TrimSpace(p.HealthBookPDFObjectKey) != ""
	p.HealthBookPDFAttached = attached
	p.HealthBookPDFURL = ""
}

func scanPetsWithEntitlements(ctx context.Context, s *Store, rows pgx.Rows) ([]Pet, error) {
	pets, err := scanPets(rows)
	if err != nil {
		return nil, err
	}
	durCache := map[string][]int{}
	for i := range pets {
		if ent, e := s.GetEntitlementByPetID(ctx, pets[i].ID); e == nil {
			pets[i].Entitlement = &ent
		}
		attachPracticeHeartRateDurations(ctx, s, &pets[i], durCache)
	}
	return pets, nil
}

// attachPracticeHeartRateDurations fills HeartrateDurationsSec, caching by practice ID
// so list endpoints do one lookup per cabinet rather than per pet.
func attachPracticeHeartRateDurations(ctx context.Context, s *Store, p *Pet, cache map[string][]int) {
	if p.PracticeID == "" {
		return
	}
	if d, ok := cache[p.PracticeID]; ok {
		p.HeartrateDurationsSec = d
		return
	}
	durations, err := s.GetPracticeHeartRateDurations(ctx, p.PracticeID)
	if err != nil {
		return
	}
	cache[p.PracticeID] = durations
	p.HeartrateDurationsSec = durations
}

func (s *Store) GetHeartRateSession(ctx context.Context, sessionID, ownerID string) (HeartRateSession, error) {
	var sess HeartRateSession
	err := s.pool.QueryRow(ctx, `
		SELECT `+heartRateSelectCols+`
		FROM heartrate.sessions WHERE id=$1 AND owner_user_id=$2`,
		sessionID, ownerID).Scan(heartRateScanDest(&sess)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	if err != nil {
		return HeartRateSession{}, err
	}
	decorateHeartRateSession(&sess)
	return sess, nil
}

func (s *Store) StartHeartRateSession(ctx context.Context, petID, ownerID, practiceID string, durationSec int) (HeartRateSession, error) {
	sess := HeartRateSession{
		ID: uuid.NewString(), PetID: petID, OwnerUserID: ownerID, PracticeID: practiceID,
		Status: kernel.SessionInProgress, DurationSec: durationSec, StartedAt: time.Now(),
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO heartrate.sessions (id, pet_id, owner_user_id, practice_id, status, duration_sec, started_at)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7) RETURNING started_at`,
		sess.ID, sess.PetID, sess.OwnerUserID, sess.PracticeID, sess.Status, sess.DurationSec, sess.StartedAt).Scan(&sess.StartedAt)
	return sess, err
}

func (s *Store) CompleteHeartRateSession(ctx context.Context, sessionID, ownerID string, tapCount int, bpm int, isAlert bool) (HeartRateSession, error) {
	var sess HeartRateSession
	var bpmVal *int
	bpmVal = &bpm
	err := s.pool.QueryRow(ctx, `
		UPDATE heartrate.sessions SET status=$2, tap_count=$3, bpm=$4, is_alert=$5, ended_at=NOW()
		WHERE id=$1 AND owner_user_id=$6 AND status='in_progress'
		RETURNING `+heartRateSelectCols,
		sessionID, kernel.SessionPendingValidation, tapCount, bpmVal, isAlert, ownerID).Scan(heartRateScanDest(&sess)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	if err != nil {
		return HeartRateSession{}, err
	}
	decorateHeartRateSession(&sess)
	return sess, nil
}

func (s *Store) ValidateHeartRateSession(ctx context.Context, sessionID, ownerID string, comment *string) (HeartRateSession, error) {
	comment = NormalizeHeartRateComment(comment)
	var sess HeartRateSession
	err := s.pool.QueryRow(ctx, `
		UPDATE heartrate.sessions
		SET status='validated', validated_at=NOW(), vet_seen_at=NULL, comment=$3
		WHERE id=$1 AND owner_user_id=$2 AND status='pending_validation'
		RETURNING `+heartRateSelectCols,
		sessionID, ownerID, comment).Scan(heartRateScanDest(&sess)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	if err != nil {
		return HeartRateSession{}, err
	}
	decorateHeartRateSession(&sess)
	return sess, nil
}

func (s *Store) CancelHeartRateSession(ctx context.Context, sessionID, ownerID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE heartrate.sessions SET status='cancelled', ended_at=NOW()
		WHERE id=$1 AND owner_user_id=$2 AND status IN ('in_progress','pending_validation')`, sessionID, ownerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListHeartRateSessions(ctx context.Context, petID string, vetView bool) ([]HeartRateSession, error) {
	q := `SELECT ` + heartRateSelectCols + `
		FROM heartrate.sessions WHERE pet_id=$1`
	if vetView {
		q += ` AND status='validated'`
	} else {
		q += ` AND status IN ('pending_validation','validated')`
	}
	q += ` ORDER BY started_at DESC`
	rows, err := s.pool.Query(ctx, q, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HeartRateSession
	for rows.Next() {
		var sess HeartRateSession
		if err := rows.Scan(heartRateScanDest(&sess)...); err != nil {
			return nil, err
		}
		decorateHeartRateSession(&sess)
		out = append(out, sess)
	}
	return out, rows.Err()
}

// GetHeartRateAlertDelta returns the species rise threshold. ok=false when the
// species is not monitored (e.g. other).
func (s *Store) GetHeartRateAlertDelta(ctx context.Context, species string) (int, bool, error) {
	species = strings.TrimSpace(strings.ToLower(species))
	if !kernel.SupportsHeartRateControl(species) {
		return 0, false, nil
	}
	var delta int
	err := s.pool.QueryRow(ctx, `
		SELECT delta_bpm FROM heartrate.species_alert_deltas WHERE species=$1`, species,
	).Scan(&delta)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return delta, true, nil
}

// LastValidatedBPM returns the most recent validated BPM for a pet, if any.
func (s *Store) LastValidatedBPM(ctx context.Context, petID string) (*int, error) {
	var bpm int
	err := s.pool.QueryRow(ctx, `
		SELECT bpm FROM heartrate.sessions
		WHERE pet_id=$1 AND status='validated' AND bpm IS NOT NULL
		ORDER BY COALESCE(validated_at, ended_at, started_at) DESC
		LIMIT 1`, petID).Scan(&bpm)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bpm, nil
}

// MarkPetHeartRateSessionsSeen sets vet_seen_at for all validated unread sessions of a pet in a practice.
func (s *Store) MarkPetHeartRateSessionsSeen(ctx context.Context, petID, practiceID string) (int64, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE heartrate.sessions
		SET vet_seen_at = NOW()
		WHERE pet_id = $1
		  AND practice_id = $2
		  AND status = 'validated'
		  AND vet_seen_at IS NULL`, petID, practiceID)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (s *Store) GetOrCreateThread(ctx context.Context, practiceID, clientID, vetID string) (Thread, error) {
	return s.GetOrCreateThreadForPet(ctx, practiceID, clientID, vetID, "")
}

// GetOrCreateThreadForPet returns the (practice, client, pet) thread.
// Empty petID keeps the legacy general thread (pet_id IS NULL).
func (s *Store) GetOrCreateThreadForPet(ctx context.Context, practiceID, clientID, vetID, petID string) (Thread, error) {
	petID = strings.TrimSpace(petID)
	var t Thread
	var err error
	if petID == "" {
		err = s.pool.QueryRow(ctx, `
			SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
			FROM messaging.threads
			WHERE practice_id=$1 AND client_user_id=$2 AND pet_id IS NULL`,
			practiceID, clientID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
			FROM messaging.threads
			WHERE practice_id=$1 AND client_user_id=$2 AND pet_id=$3::uuid`,
			practiceID, clientID, petID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	}
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, err
	}
	t = Thread{ID: uuid.NewString(), PracticeID: practiceID, ClientUserID: clientID, VetUserID: vetID, PetID: petID}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid)`,
		t.ID, t.PracticeID, t.ClientUserID, t.VetUserID, t.PetID)
	return t, err
}

// GetOrCreateCareProThread returns a person-scoped thread (practice_id IS NULL)
// between a care_pro and a client. Distinct from cabinet practice threads.
func (s *Store) GetOrCreateCareProThread(ctx context.Context, careProID, clientID, petID string) (Thread, error) {
	petID = strings.TrimSpace(petID)
	var t Thread
	var err error
	if petID == "" {
		err = s.pool.QueryRow(ctx, `
			SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
			FROM messaging.threads
			WHERE practice_id IS NULL AND vet_user_id=$1 AND client_user_id=$2 AND pet_id IS NULL`,
			careProID, clientID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
			FROM messaging.threads
			WHERE practice_id IS NULL AND vet_user_id=$1 AND client_user_id=$2 AND pet_id=$3::uuid`,
			careProID, clientID, petID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	}
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, err
	}
	t = Thread{ID: uuid.NewString(), ClientUserID: clientID, VetUserID: careProID, PetID: petID}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1,NULL,$2,$3,NULLIF($4,'')::uuid)`,
		t.ID, t.ClientUserID, t.VetUserID, t.PetID)
	return t, err
}

func (s *Store) AddMessage(ctx context.Context, threadID, senderID, body string) (Message, error) {
	return s.AddMessageMedia(ctx, threadID, senderID, body, "", "")
}

func (s *Store) AddMessageMedia(ctx context.Context, threadID, senderID, body, mediaURL, mediaType string) (Message, error) {
	m := Message{
		ID:           uuid.NewString(),
		ThreadID:     threadID,
		SenderUserID: senderID,
		Body:         body,
		MediaURL:     mediaURL,
		MediaType:    mediaType,
		CreatedAt:    time.Now(),
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO messaging.messages (id, thread_id, sender_user_id, body, media_url, media_type, created_at)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),$7) RETURNING created_at`,
		m.ID, m.ThreadID, m.SenderUserID, m.Body, m.MediaURL, m.MediaType, m.CreatedAt).Scan(&m.CreatedAt)
	return m, err
}

func (s *Store) ListMessages(ctx context.Context, threadID string) ([]Message, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, thread_id::text, sender_user_id::text, body,
			COALESCE(media_url,''), COALESCE(media_type,''), read_at, created_at
		FROM messaging.messages WHERE thread_id=$1 ORDER BY created_at ASC`, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ThreadID, &m.SenderUserID, &m.Body, &m.MediaURL, &m.MediaType, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// MarkThreadRead sets read_at on messages sent by the other participant.
func (s *Store) MarkThreadRead(ctx context.Context, threadID, readerUserID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE messaging.messages
		SET read_at = NOW()
		WHERE thread_id = $1
		  AND sender_user_id <> $2
		  AND read_at IS NULL`, threadID, readerUserID)
	return err
}

// MarkAllUnreadForUser marks all unread inbound messages as read for a participant.
func (s *Store) MarkAllUnreadForUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE messaging.messages m
		SET read_at = NOW()
		FROM messaging.threads t
		WHERE m.thread_id = t.id
		  AND (t.vet_user_id = $1 OR t.client_user_id = $1)
		  AND m.sender_user_id <> $1
		  AND m.read_at IS NULL`, userID)
	return err
}

// MarkAllUnreadForPractice marks unread client messages as read for all practice threads.
func (s *Store) MarkAllUnreadForPractice(ctx context.Context, practiceID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE messaging.messages m
		SET read_at = NOW()
		FROM messaging.threads t
		WHERE m.thread_id = t.id
		  AND t.practice_id = $1
		  AND m.sender_user_id = t.client_user_id
		  AND m.read_at IS NULL`, practiceID)
	return err
}

func (s *Store) GetVetAvailability(ctx context.Context, vetID string) (kernel.AvailabilityStatus, string, error) {
	var status kernel.AvailabilityStatus
	var autoReply string
	err := s.pool.QueryRow(ctx, `
		SELECT status, COALESCE(auto_reply,'') FROM messaging.vet_availability WHERE vet_user_id=$1`, vetID).Scan(&status, &autoReply)
	if errors.Is(err, pgx.ErrNoRows) {
		return kernel.AvailabilityAvailable, "", nil
	}
	return status, autoReply, err
}

func (s *Store) SetVetAvailability(ctx context.Context, vetID, practiceID string, status kernel.AvailabilityStatus, autoReply string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply, updated_at)
		VALUES ($1,$2,$3,$4,NOW())
		ON CONFLICT (vet_user_id) DO UPDATE SET status=$3, auto_reply=$4, updated_at=NOW()`,
		vetID, practiceID, status, autoReply)
	return err
}

func (s *Store) GetThreadByID(ctx context.Context, threadID string) (Thread, error) {
	var t Thread
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
		FROM messaging.threads WHERE id=$1`, threadID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, ErrNotFound
	}
	return t, err
}

func (s *Store) ListThreadsForVet(ctx context.Context, vetID string) ([]Thread, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(practice_id::text,''), client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
		FROM messaging.threads WHERE vet_user_id=$1 ORDER BY created_at DESC`, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// PetTimelineFiltered builds the pet timeline.
// includeMessages: messaging bodies for this pet's thread.
// redactVisitNotes: empty visit notes / CR excerpts (no write_notes / no pets.write_clinical for staff).
// Visit report meta (hasReport) is included for staff/care_pro (vetView) when a non-empty CR exists;
// CR body excerpts only when !redactVisitNotes — never for client timeline or public dossier share.
// Vet timeline includes done visits, plus confirmed visits that already have a non-empty CR
// (walk-in save without Terminer / finalize).
func (s *Store) PetTimelineFiltered(ctx context.Context, petID string, vetView, includeMessages, redactVisitNotes bool) ([]TimelineItem, error) {
	hrFilter := ""
	if vetView {
		hrFilter = " AND status='validated'"
	} else {
		hrFilter = " AND status IN ('pending_validation','validated')"
	}

	var visitBranch string
	if vetView {
		// Prefer non-empty CR excerpt over visit notes; redact body when ACL forbids clinical text.
		visitBody := "''"
		if !redactVisitNotes {
			visitBody = `
			CASE
				WHEN r.id IS NOT NULL THEN left(r.body_text, 280)
				WHEN btrim(COALESCE(v.notes, '')) <> '' THEN COALESCE(v.notes, '')
				ELSE ''
			END`
		}
		visitBranch = `
		SELECT v.id::text, 'visit', 'Visite', ` + visitBody + `,
			COALESCE(v.scheduled_at, v.created_at),
			jsonb_build_object(
				'status', v.status,
				'source', v.source,
				'visitId', v.id::text,
				'hasReport', (r.id IS NOT NULL),
				'reportStatus', COALESCE(r.status, '')
			)
		FROM visits.visits v
		LEFT JOIN LATERAL (
			SELECT id, status, body_text
			FROM visits.visit_reports
			WHERE visit_id = v.id
				AND btrim(COALESCE(body_text, '')) <> ''
			ORDER BY CASE status WHEN 'final' THEN 0 ELSE 1 END, updated_at DESC
			LIMIT 1
		) r ON true
		WHERE v.pet_id=$1
			AND v.deleted_at IS NULL
			AND (
				v.status = 'done'
				OR (v.status = 'confirmed' AND r.id IS NOT NULL)
			)`
	} else {
		// Client / dossier: hasReport only for finalized CR (opens client consultation).
		// reportStatus may be draft|final (existence signal only); never leak draft body.
		// Include confirmed visits that already have a persisted CR so draft/final CTAs appear
		// before Terminer / even if finalize auto-done was skipped.
		visitBody := "COALESCE(notes,'')"
		if redactVisitNotes {
			visitBody = "''"
		}
		visitBranch = `
		SELECT v.id::text, 'visit', 'Visite', ` + visitBody + `,
			COALESCE(v.scheduled_at, v.created_at),
			jsonb_build_object(
				'status', v.status,
				'source', v.source,
				'visitId', v.id::text,
				'hasReport', (r.id IS NOT NULL AND r.status = 'final'),
				'reportStatus', COALESCE(r.status, '')
			)
		FROM visits.visits v
		LEFT JOIN LATERAL (
			SELECT id, status
			FROM visits.visit_reports r
			WHERE r.visit_id = v.id
			  AND ` + sqlVisitReportIsPersisted + `
			ORDER BY CASE status WHEN 'final' THEN 0 ELSE 1 END, updated_at DESC
			LIMIT 1
		) r ON true
		WHERE v.pet_id=$1 AND v.deleted_at IS NULL
			AND (
				v.status = 'done'
				OR (v.status = 'confirmed' AND r.id IS NOT NULL)
			)`
	}
	q := `
		SELECT id::text, 'heartrate', 'Relevé cardiaque',
			CASE
				WHEN comment IS NOT NULL AND btrim(comment) <> ''
					THEN CONCAT('BPM: ', COALESCE(bpm::text,'?'), ' — ', comment)
				ELSE CONCAT('BPM: ', COALESCE(bpm::text,'?'))
			END,
			started_at,
			jsonb_build_object('bpm', bpm, 'status', status, 'is_alert', is_alert, 'comment', comment)
		FROM heartrate.sessions WHERE pet_id=$1` + hrFilter + `
		UNION ALL
		SELECT id::text, 'weight', 'Poids',
			CASE
				WHEN comment IS NOT NULL AND btrim(comment) <> ''
					THEN CONCAT(weight_kg::text, ' kg — ', comment)
				ELSE CONCAT(weight_kg::text, ' kg')
			END,
			recorded_at,
			jsonb_build_object('weight_kg', weight_kg, 'comment', comment)
		FROM pets.weight_readings WHERE pet_id=$1
		UNION ALL
		SELECT id::text, 'blood_pressure', 'Tension',
			CASE
				WHEN comment IS NOT NULL AND btrim(comment) <> ''
					THEN CONCAT(systolic_mmhg::text, '/', diastolic_mmhg::text, ' — ', comment)
				ELSE CONCAT(systolic_mmhg::text, '/', diastolic_mmhg::text, ' · ', initcap(method))
			END,
			recorded_at,
			jsonb_build_object(
				'systolic_mmhg', systolic_mmhg,
				'diastolic_mmhg', diastolic_mmhg,
				'mean_mmhg', mean_mmhg,
				'method', method,
				'site', site,
				'comment', comment
			)
		FROM pets.blood_pressure_readings WHERE pet_id=$1
		UNION ALL
		SELECT p.id::text, 'lab_panel',
			CASE WHEN btrim(p.lab_name) <> '' THEN CONCAT('Prise de sang — ', p.lab_name) ELSE 'Prise de sang' END,
			CASE
				WHEN COALESCE(ab.cnt, 0) > 0
					THEN CONCAT(ab.cnt::text, ' valeur(s) hors norme')
				ELSE CONCAT(COALESCE(rc.cnt, 0)::text, ' analyte(s)')
			END,
			p.collected_at,
			jsonb_build_object(
				'lab_name', p.lab_name,
				'abnormal_count', COALESCE(ab.cnt, 0),
				'result_count', COALESCE(rc.cnt, 0),
				'document_id', p.document_id
			)
		FROM labs.panels p
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int AS cnt FROM labs.panel_results r
			WHERE r.panel_id = p.id AND r.flag IN ('low','high')
		) ab ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int AS cnt FROM labs.panel_results r WHERE r.panel_id = p.id
		) rc ON true
		WHERE p.pet_id=$1
		UNION ALL
		SELECT id::text, 'event', event_type, content, created_at, '{}'::jsonb
		FROM pets.dossier_events WHERE pet_id=$1
		UNION ALL
		SELECT id::text, 'care', title, type, updated_at, jsonb_build_object('status', status, 'due_at', due_at)
		FROM care.reminders WHERE pet_id=$1 AND status='done'
		UNION ALL` + visitBranch
	if includeMessages {
		q += `
		UNION ALL
		SELECT m.id::text, 'message', 'Message', m.body, m.created_at, '{}'::jsonb
		FROM messaging.messages m
		JOIN messaging.threads t ON t.id = m.thread_id
		WHERE t.pet_id=$1`
	}
	q += `
		ORDER BY 5 DESC`
	rows, err := s.pool.Query(ctx, q, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TimelineItem
	for rows.Next() {
		var item TimelineItem
		var typeStr string
		var meta map[string]any
		if err := rows.Scan(&item.ID, &typeStr, &item.Title, &item.Body, &item.CreatedAt, &meta); err != nil {
			return nil, err
		}
		item.Type = kernel.TimelineType(typeStr)
		item.Meta = meta
		out = append(out, item)
	}
	if out == nil {
		out = []TimelineItem{}
	}
	return out, rows.Err()
}

func (s *Store) InsertDossierEvent(ctx context.Context, petID, authorUserID, eventType, content string) error {
	petID = strings.TrimSpace(petID)
	authorUserID = strings.TrimSpace(authorUserID)
	eventType = strings.TrimSpace(eventType)
	content = strings.TrimSpace(content)
	if petID == "" || authorUserID == "" || eventType == "" || content == "" {
		return ErrValidation
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pets.dossier_events (id, pet_id, author_user_id, event_type, content)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5)`,
		uuid.NewString(), petID, authorUserID, eventType, content)
	return err
}

func (s *Store) LogNotification(ctx context.Context, vetID, kind string, payload map[string]any) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications.notification_log (id, vet_user_id, kind, payload)
		VALUES ($1,$2,$3,$4)`, uuid.NewString(), vetID, kind, payload)
	return err
}

func (s *Store) GetVetForClient(ctx context.Context, clientID, practiceID string) (string, error) {
	var vetID string
	err := s.pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM practice.practice_clients
		WHERE client_user_id=$1 AND practice_id=$2`, clientID, practiceID).Scan(&vetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return vetID, err
}

type VetEmailPrefs struct {
	OnMessage      bool
	OnHeartRate    bool
	OnVisitRequest bool
}

func (s *Store) EmailPrefs(ctx context.Context, vetID string) (VetEmailPrefs, error) {
	var p VetEmailPrefs
	err := s.pool.QueryRow(ctx, `
		SELECT email_on_message, email_on_heartrate, COALESCE(email_on_visit_request, TRUE)
		FROM notifications.notification_preferences WHERE vet_user_id=$1`, vetID).
		Scan(&p.OnMessage, &p.OnHeartRate, &p.OnVisitRequest)
	if errors.Is(err, pgx.ErrNoRows) {
		return VetEmailPrefs{OnMessage: true, OnHeartRate: true, OnVisitRequest: true}, nil
	}
	return p, err
}
