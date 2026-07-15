package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"golang.org/x/crypto/bcrypt"
)

type PracticeProfile struct {
	PracticeID             string     `json:"practiceId"`
	PracticeName           string     `json:"practiceName"`
	Phone                  string     `json:"phone"`
	ContactEmail           string     `json:"contactEmail"`
	AddressLine1           string     `json:"addressLine1"`
	AddressLine2           string     `json:"addressLine2"`
	City                   string     `json:"city"`
	PostalCode             string     `json:"postalCode"`
	Website                string     `json:"website"`
	ProfileCompletedAt     *time.Time `json:"profileCompletedAt,omitempty"`
	VetFullName            string     `json:"vetFullName"`
	VetEmail               string     `json:"vetEmail"`
	HeartRateDurationsSec  []int      `json:"heartrateDurationsSec"`
}

type RegisterVetInput struct {
	Email            string
	Password         string
	FullName         string
	PracticeName     string
	PreferredLocale  string
	AutoReplyDefault string
}

type RegisterVetResult struct {
	UserID string
	Token  string
}

func (s *Store) GetPracticeProfile(ctx context.Context, practiceID, vetUserID string) (PracticeProfile, error) {
	var p PracticeProfile
	var completedAt *time.Time
	var durations []int32
	err := s.pool.QueryRow(ctx, `
		SELECT pr.id::text, pr.name, COALESCE(pr.phone,''), COALESCE(pr.contact_email,''),
			COALESCE(pr.address_line1,''), COALESCE(pr.address_line2,''), COALESCE(pr.city,''),
			COALESCE(pr.postal_code,''), COALESCE(pr.website,''), pr.profile_completed_at,
			u.full_name, u.email, pr.heartrate_durations_sec
		FROM practice.practices pr
		JOIN identity.users u ON u.id = $2 AND u.practice_id = pr.id
		WHERE pr.id = $1`, practiceID, vetUserID).Scan(
		&p.PracticeID, &p.PracticeName, &p.Phone, &p.ContactEmail,
		&p.AddressLine1, &p.AddressLine2, &p.City, &p.PostalCode, &p.Website, &completedAt,
		&p.VetFullName, &p.VetEmail, &durations,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PracticeProfile{}, ErrNotFound
	}
	p.ProfileCompletedAt = completedAt
	p.HeartRateDurationsSec = int32SliceToInts(durations)
	return p, err
}

func int32SliceToInts(in []int32) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

func (s *Store) GetPracticeHeartRateDurations(ctx context.Context, practiceID string) ([]int, error) {
	var durations []int32
	err := s.pool.QueryRow(ctx, `
		SELECT heartrate_durations_sec FROM practice.practices WHERE id = $1`, practiceID).Scan(&durations)
	if errors.Is(err, pgx.ErrNoRows) {
		return []int{60}, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(durations) == 0 {
		return []int{60}, nil
	}
	return int32SliceToInts(durations), nil
}

func (s *Store) UpdatePracticeProfile(ctx context.Context, practiceID, vetUserID string, p PracticeProfile, markComplete bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET full_name = $2 WHERE id = $1 AND practice_id = $3`,
		vetUserID, p.VetFullName, practiceID); err != nil {
		return err
	}

	durations := p.HeartRateDurationsSec
	if len(durations) == 0 {
		durations = []int{60}
	}
	q := `
		UPDATE practice.practices
		SET name = $2, phone = $3, contact_email = $4, address_line1 = $5, address_line2 = $6,
			city = $7, postal_code = $8, website = $9, heartrate_durations_sec = $10`
	args := []any{practiceID, p.PracticeName, p.Phone, p.ContactEmail, p.AddressLine1, p.AddressLine2, p.City, p.PostalCode, p.Website, durations}
	if markComplete {
		q += `, profile_completed_at = COALESCE(profile_completed_at, NOW())`
	}
	q += ` WHERE id = $1`
	if _, err := tx.Exec(ctx, q, args...); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) IsProfileComplete(ctx context.Context, practiceID string) (bool, error) {
	var complete bool
	err := s.pool.QueryRow(ctx, `
		SELECT profile_completed_at IS NOT NULL
			AND phone <> '' AND address_line1 <> '' AND city <> '' AND postal_code <> '' AND contact_email <> ''
		FROM practice.practices WHERE id = $1`, practiceID).Scan(&complete)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return complete, err
}

func (s *Store) RegisterVet(ctx context.Context, in RegisterVetInput) (RegisterVetResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterVetResult{}, err
	}

	practiceID := uuid.NewString()
	userID := uuid.NewString()
	token := uuid.NewString()
	expires := time.Now().Add(48 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RegisterVetResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO practice.practices (id, name, contact_email) VALUES ($1, $2, $3)`,
		practiceID, in.PracticeName, in.Email); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, preferred_locale)
		VALUES ($1, $2, $3, $4, 'vet', $5, $6)`,
		userID, in.Email, string(hash), in.FullName, practiceID, i18n.NormalizeLocale(in.PreferredLocale)); err != nil {
		return RegisterVetResult{}, err
	}
	autoReply := in.AutoReplyDefault
	if autoReply == "" {
		autoReply = "Je suis indisponible, je reviens vers vous rapidement."
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, 'available', $3)`,
		userID, practiceID, autoReply); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, true, true)`, userID); err != nil {
		return RegisterVetResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), userID, token, expires); err != nil {
		return RegisterVetResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RegisterVetResult{}, err
	}
	return RegisterVetResult{UserID: userID, Token: token}, nil
}

func (s *Store) ConfirmEmail(ctx context.Context, token string) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id::text, expires_at, used_at FROM identity.email_verification_tokens WHERE token = $1`, token).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if usedAt != nil {
		return User{}, errors.New("token already used")
	}
	if time.Now().After(expiresAt) {
		return User{}, errors.New("token expired")
	}

	if _, err := tx.Exec(ctx, `UPDATE identity.users SET email_verified_at = NOW() WHERE id = $1`, userID); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.email_verification_tokens SET used_at = NOW() WHERE token = $1`, token); err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Store) GetUserMe(ctx context.Context, userID string) (map[string]any, error) {
	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"userId":           u.ID,
		"email":            u.Email,
		"role":             u.Role,
		"fullName":         u.FullName,
		"avatarUrl":        u.AvatarURL,
		"emailVerified":    u.EmailVerifiedAt != nil,
		"authProvider":     u.AuthProvider,
		"googleLinked":     u.GoogleSub != "",
		"twoFactorEnabled": u.TOTPEnabled,
		"preferredLocale":  u.PreferredLocale,
	}
	if u.PracticeID != "" {
		out["practiceId"] = u.PracticeID
		var practiceName string
		_ = s.pool.QueryRow(ctx, `SELECT name FROM practice.practices WHERE id = $1`, u.PracticeID).Scan(&practiceName)
		out["practiceName"] = practiceName
		complete, _ := s.IsProfileComplete(ctx, u.PracticeID)
		out["profileComplete"] = complete
	}
	return out, nil
}
