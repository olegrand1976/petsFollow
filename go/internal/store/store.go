package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
)

type User struct {
	ID                 string
	Email              string
	PasswordHash       string
	FullName           string
	Role               kernel.Role
	PracticeID         string
	EmailVerifiedAt    *time.Time
	GoogleSub          string
	AuthProvider       string
	TOTPSecret         string
	TOTPEnabled        bool
	PreferredLocale    string
	AvatarURL          string
	MustChangePassword bool
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
	HeartrateDurationsSec []int `json:"heartrateDurationsSec,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Entitlement   *Entitlement `json:"entitlement,omitempty"`
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
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	PetCount  int    `json:"petCount"`
}

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	var passwordHash *string
	err := row.Scan(
		&u.ID, &u.Email, &passwordHash, &u.FullName, &u.Role, &u.PracticeID, &u.EmailVerifiedAt,
		&u.GoogleSub, &u.AuthProvider, &u.TOTPSecret, &u.TOTPEnabled, &u.PreferredLocale, &u.AvatarURL,
		&u.MustChangePassword,
	)
	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	return u, err
}

const userSelectCols = `
	id::text, email, password_hash, full_name, role, COALESCE(practice_id::text,''), email_verified_at,
	COALESCE(google_sub,''), COALESCE(auth_provider,'password'), COALESCE(totp_secret,''), totp_enabled,
	COALESCE(preferred_locale,'fr'), COALESCE(avatar_url,''), must_change_password`

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

func (s *Store) ListClientsByPractice(ctx context.Context, practiceID string) ([]ClientSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.email, u.full_name, COALESCE(u.avatar_url,''), COUNT(p.id)::int
		FROM practice.practice_clients pc
		JOIN identity.users u ON u.id = pc.client_user_id
		LEFT JOIN pets.pets p ON p.owner_user_id = u.id AND p.practice_id = pc.practice_id
		WHERE pc.practice_id = $1
		GROUP BY u.id, u.email, u.full_name, u.avatar_url
		ORDER BY u.full_name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ClientSummary, 0)
	for rows.Next() {
		var c ClientSummary
		if err := rows.Scan(&c.UserID, &c.Email, &c.FullName, &c.AvatarURL, &c.PetCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreatePet(ctx context.Context, p Pet) (Pet, error) {
	p.ID = uuid.NewString()
	if p.PaymentStatus == "" {
		p.PaymentStatus = "pending_payment"
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, birth_date, weight_kg, photo_url, payment_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at`, p.ID, p.PracticeID, p.OwnerUserID, p.Name, p.Species, p.Breed, p.BirthDate, p.WeightKg, p.PhotoURL, p.PaymentStatus).Scan(&p.CreatedAt)
	return p, err
}

func (s *Store) UpdatePet(ctx context.Context, p Pet) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE pets.pets SET name=$2, species=$3, breed=$4, birth_date=$5, weight_kg=$6, photo_url=$7, updated_at=NOW()
		WHERE id=$1 AND owner_user_id=$8`, p.ID, p.Name, p.Species, p.Breed, p.BirthDate, p.WeightKg, p.PhotoURL, p.OwnerUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetPet(ctx context.Context, id string) (Pet, error) {
	var p Pet
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, created_at
		FROM pets.pets WHERE id=$1`, id).Scan(
		&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed, &p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Pet{}, ErrNotFound
	}
	if err == nil {
		if ent, e := s.GetEntitlementByPetID(ctx, id); e == nil {
			p.Entitlement = &ent
		}
		if durations, e := s.GetPracticeHeartRateDurations(ctx, p.PracticeID); e == nil {
			p.HeartrateDurationsSec = durations
		}
	}
	return p, err
}

func (s *Store) ListPetsByOwner(ctx context.Context, ownerID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, created_at
		FROM pets.pets WHERE owner_user_id=$1 ORDER BY name`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPetsWithEntitlements(ctx, s, rows)
}

func (s *Store) ListPetsByClientForVet(ctx context.Context, practiceID, clientID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, owner_user_id::text, name, species, COALESCE(breed,''),
			birth_date, weight_kg, COALESCE(photo_url,''), payment_status, created_at
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
		if err := rows.Scan(&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed, &p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanPetsWithEntitlements(ctx context.Context, s *Store, rows pgx.Rows) ([]Pet, error) {
	pets, err := scanPets(rows)
	if err != nil {
		return nil, err
	}
	for i := range pets {
		if ent, e := s.GetEntitlementByPetID(ctx, pets[i].ID); e == nil {
			pets[i].Entitlement = &ent
		}
		if durations, e := s.GetPracticeHeartRateDurations(ctx, pets[i].PracticeID); e == nil {
			pets[i].HeartrateDurationsSec = durations
		}
	}
	return pets, nil
}

func (s *Store) GetHeartRateSession(ctx context.Context, sessionID, ownerID string) (HeartRateSession, error) {
	var sess HeartRateSession
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, practice_id::text, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at
		FROM heartrate.sessions WHERE id=$1 AND owner_user_id=$2`,
		sessionID, ownerID).Scan(
		&sess.ID, &sess.PetID, &sess.OwnerUserID, &sess.PracticeID, &sess.Status, &sess.TapCount, &sess.DurationSec, &sess.BPM, &sess.IsAlert, &sess.StartedAt, &sess.EndedAt, &sess.ValidatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	return sess, err
}

func (s *Store) StartHeartRateSession(ctx context.Context, petID, ownerID, practiceID string, durationSec int) (HeartRateSession, error) {
	sess := HeartRateSession{
		ID: uuid.NewString(), PetID: petID, OwnerUserID: ownerID, PracticeID: practiceID,
		Status: kernel.SessionInProgress, DurationSec: durationSec, StartedAt: time.Now(),
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO heartrate.sessions (id, pet_id, owner_user_id, practice_id, status, duration_sec, started_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING started_at`,
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
		RETURNING id::text, pet_id::text, owner_user_id::text, practice_id::text, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at`,
		sessionID, kernel.SessionPendingValidation, tapCount, bpmVal, isAlert, ownerID).Scan(
		&sess.ID, &sess.PetID, &sess.OwnerUserID, &sess.PracticeID, &sess.Status, &sess.TapCount, &sess.DurationSec, &sess.BPM, &sess.IsAlert, &sess.StartedAt, &sess.EndedAt, &sess.ValidatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	return sess, err
}

func (s *Store) ValidateHeartRateSession(ctx context.Context, sessionID, ownerID string) (HeartRateSession, error) {
	var sess HeartRateSession
	err := s.pool.QueryRow(ctx, `
		UPDATE heartrate.sessions SET status='validated', validated_at=NOW()
		WHERE id=$1 AND owner_user_id=$2 AND status='pending_validation'
		RETURNING id::text, pet_id::text, owner_user_id::text, practice_id::text, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at`,
		sessionID, ownerID).Scan(
		&sess.ID, &sess.PetID, &sess.OwnerUserID, &sess.PracticeID, &sess.Status, &sess.TapCount, &sess.DurationSec, &sess.BPM, &sess.IsAlert, &sess.StartedAt, &sess.EndedAt, &sess.ValidatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HeartRateSession{}, ErrNotFound
	}
	return sess, err
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
	q := `SELECT id::text, pet_id::text, owner_user_id::text, practice_id::text, status, tap_count, duration_sec, bpm, is_alert, started_at, ended_at, validated_at
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
		if err := rows.Scan(&sess.ID, &sess.PetID, &sess.OwnerUserID, &sess.PracticeID, &sess.Status, &sess.TapCount, &sess.DurationSec, &sess.BPM, &sess.IsAlert, &sess.StartedAt, &sess.EndedAt, &sess.ValidatedAt); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) GetOrCreateThread(ctx context.Context, practiceID, clientID, vetID string) (Thread, error) {
	var t Thread
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
		FROM messaging.threads WHERE practice_id=$1 AND client_user_id=$2`, practiceID, clientID).Scan(
		&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, err
	}
	t = Thread{ID: uuid.NewString(), PracticeID: practiceID, ClientUserID: clientID, VetUserID: vetID}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id) VALUES ($1,$2,$3,$4)`,
		t.ID, t.PracticeID, t.ClientUserID, t.VetUserID)
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
		SELECT id::text, practice_id::text, client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
		FROM messaging.threads WHERE id=$1`, threadID).Scan(&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, ErrNotFound
	}
	return t, err
}

func (s *Store) ListThreadsForVet(ctx context.Context, vetID string) ([]Thread, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, client_user_id::text, vet_user_id::text, COALESCE(pet_id::text,'')
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

func (s *Store) PetTimeline(ctx context.Context, petID string, vetView bool) ([]TimelineItem, error) {
	hrFilter := ""
	if vetView {
		hrFilter = " AND status='validated'"
	} else {
		hrFilter = " AND status IN ('pending_validation','validated')"
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, 'heartrate', 'Relevé cardiaque', CONCAT('BPM: ', COALESCE(bpm::text,'?')), started_at,
			jsonb_build_object('bpm', bpm, 'status', status, 'is_alert', is_alert)
		FROM heartrate.sessions WHERE pet_id=$1`+hrFilter+`
		UNION ALL
		SELECT id::text, 'event', event_type, content, created_at, '{}'::jsonb
		FROM pets.dossier_events WHERE pet_id=$1
		UNION ALL
		SELECT m.id::text, 'message', 'Message', m.body, m.created_at, '{}'::jsonb
		FROM messaging.messages m
		JOIN messaging.threads t ON t.id = m.thread_id
		WHERE t.pet_id=$1 OR EXISTS (SELECT 1 FROM pets.pets p WHERE p.id=$1 AND p.owner_user_id=t.client_user_id)
		UNION ALL
		SELECT id::text, 'care', title, type, updated_at, jsonb_build_object('status', status, 'due_at', due_at)
		FROM care.reminders WHERE pet_id=$1 AND status='done'
		UNION ALL
		SELECT id::text, 'visit', 'Visite', COALESCE(notes,''), COALESCE(scheduled_at, created_at),
			jsonb_build_object('status', status, 'source', source)
		FROM visits.visits WHERE pet_id=$1 AND status='done'
		ORDER BY 4 DESC`, petID)
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
	return out, rows.Err()
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

func (s *Store) EmailPrefs(ctx context.Context, vetID string) (onMessage, onHeartRate bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT email_on_message, email_on_heartrate FROM notifications.notification_preferences WHERE vet_user_id=$1`, vetID).
		Scan(&onMessage, &onHeartRate)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, true, nil
	}
	return onMessage, onHeartRate, err
}
