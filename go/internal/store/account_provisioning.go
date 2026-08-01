package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type CreateClientInput struct {
	Email                  string
	Password               string
	FullName               string
	FirstName              string
	LastName               string
	Locale                 string
	ContactPhone           string
	Address                string
	NationalRegistryNumber string
	SkipJourney            bool
}

func composeClientNames(in *CreateClientInput) {
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.FullName = strings.TrimSpace(in.FullName)
	if in.FirstName != "" || in.LastName != "" {
		in.FullName = strings.TrimSpace(in.FirstName + " " + in.LastName)
		return
	}
	if in.FullName == "" {
		return
	}
	parts := strings.SplitN(in.FullName, " ", 2)
	in.FirstName = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		in.LastName = strings.TrimSpace(parts[1])
	}
}

// practiceLinkForStaff resolves practice_id for practice staff and the vet_user_id
// to attach on practice_clients / messaging threads (reference vet when set).
func (s *Store) practiceLinkForStaff(ctx context.Context, staffUserID string) (practiceID, threadVetID string, err error) {
	var role kernel.Role
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,''), role FROM identity.users WHERE id=$1`, staffUserID,
	).Scan(&practiceID, &role)
	if errors.Is(err, pgx.ErrNoRows) || practiceID == "" || !kernel.IsPracticeStaff(role) {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", err
	}
	threadVetID = staffUserID
	var ref string
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(reference_vet_user_id::text,'') FROM practice.practices WHERE id=$1`, practiceID,
	).Scan(&ref)
	if ref != "" {
		threadVetID = ref
	} else if role != kernel.RoleVet {
		// Prefer any active vet on the practice for thread ownership.
		_ = s.pool.QueryRow(ctx, `
			SELECT id::text FROM identity.users
			WHERE practice_id=$1 AND role='vet' ORDER BY created_at LIMIT 1`, practiceID,
		).Scan(&threadVetID)
		if threadVetID == "" {
			threadVetID = staffUserID
		}
	}
	return practiceID, threadVetID, nil
}

type VetOption struct {
	UserID       string `json:"userId"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	PracticeID   string `json:"practiceId,omitempty"`
	PracticeName string `json:"practiceName"`
}

func (s *Store) CreateClientForVet(ctx context.Context, vetUserID string, in CreateClientInput) (string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	composeClientNames(&in)
	if in.Password == "" {
		pw, err := randomPassword(24)
		if err != nil {
			return "", err
		}
		in.Password = pw
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	practiceID, threadVetID, err := s.practiceLinkForStaff(ctx, vetUserID)
	if err != nil {
		return "", err
	}

	clientID := uuid.NewString()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, first_name, last_name, role, practice_id,
			email_verified_at, preferred_locale, must_change_password, contact_phone,
			address, national_registry_number
		) VALUES ($1, $2, $3, $4, $5, $6, 'client', $7, NOW(), $8, true, $9, $10, $11)`,
		clientID, in.Email, string(hash), in.FullName, in.FirstName, in.LastName, practiceID,
		i18n.NormalizeLocale(in.Locale), strings.TrimSpace(in.ContactPhone),
		strings.TrimSpace(in.Address), strings.TrimSpace(in.NationalRegistryNumber)); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), practiceID, clientID, threadVetID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1, $2, $3, $4, NULL)`,
		uuid.NewString(), practiceID, clientID, threadVetID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	// Best-effort: start in-app discovery + email loyalty journey.
	// SkipJourney: mark completed so BackfillEmailJourneys never enrolls the import later.
	if in.SkipJourney {
		_ = s.MarkEmailJourneySkipped(ctx, clientID)
	} else {
		_ = s.EnrollEmailJourney(ctx, clientID, time.Now().UTC())
	}
	var assignedComm string
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`, threadVetID).Scan(&assignedComm)
	_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
		EventType:        FiliationEventPracticeClientLinked,
		CommercialUserID: assignedComm,
		VetUserID:        threadVetID,
		ClientUserID:     clientID,
		PracticeID:       practiceID,
		ActorUserID:      threadVetID,
		Meta:             map[string]any{"source": "create_client_for_vet"},
	})
	return clientID, nil
}

// CreateClientStandalone creates a client with no practice / vet link.
// The client can later request a vet link from the pets app.
func (s *Store) CreateClientStandalone(ctx context.Context, in CreateClientInput) (string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	composeClientNames(&in)
	if in.Password == "" {
		pw, err := randomPassword(24)
		if err != nil {
			return "", err
		}
		in.Password = pw
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	clientID := uuid.NewString()
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, first_name, last_name, role, practice_id,
			email_verified_at, preferred_locale, must_change_password, contact_phone,
			address, national_registry_number
		) VALUES ($1, $2, $3, $4, $5, $6, 'client', NULL, NOW(), $7, true, $8, $9, $10)`,
		clientID, in.Email, string(hash), in.FullName, in.FirstName, in.LastName,
		i18n.NormalizeLocale(in.Locale), strings.TrimSpace(in.ContactPhone),
		strings.TrimSpace(in.Address), strings.TrimSpace(in.NationalRegistryNumber)); err != nil {
		return "", err
	}
	if in.SkipJourney {
		_ = s.MarkEmailJourneySkipped(ctx, clientID)
	} else {
		_ = s.EnrollEmailJourney(ctx, clientID, time.Now().UTC())
	}
	return clientID, nil
}

func (s *Store) CommercialOwnsVet(ctx context.Context, commercialUserID, vetUserID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM identity.users
			WHERE id=$1 AND role='vet' AND assigned_commercial_id=$2
		)`, vetUserID, commercialUserID).Scan(&ok)
	return ok, err
}

func (s *Store) CreateVetAsAdmin(ctx context.Context, in EncodeVetInput, assignedCommercialID string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	practiceID := uuid.NewString()
	userID := uuid.NewString()
	contactEmail := in.ContactEmail
	if contactEmail == "" {
		contactEmail = in.Email
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practices (id, name, phone, contact_email, address_line1, city, postal_code, profile_completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		practiceID, in.PracticeName, in.Phone, contactEmail, in.AddressLine1, in.City, in.PostalCode); err != nil {
		return "", err
	}

	var commercialArg any
	if assignedCommercialID != "" {
		commercialArg = assignedCommercialID
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id,
			email_verified_at, assigned_commercial_id, preferred_locale, must_change_password
		) VALUES ($1, $2, $3, $4, 'vet', $5, NOW(), $6, $7, true)`,
		userID, in.Email, string(hash), in.FullName, practiceID, commercialArg, i18n.NormalizeLocale(in.PreferredLocale)); err != nil {
		return "", err
	}
	autoReply := in.AutoReplyDefault
	if autoReply == "" {
		autoReply = "Je suis indisponible, je reviens vers vous rapidement."
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.vet_availability (vet_user_id, practice_id, status, auto_reply)
		VALUES ($1, $2, 'available', $3)`, userID, practiceID, autoReply); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate)
		VALUES ($1, true, true)`, userID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	if assignedCommercialID != "" {
		_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
			EventType:        FiliationEventVetAssigned,
			CommercialUserID: assignedCommercialID,
			VetUserID:        userID,
			PracticeID:       practiceID,
			ActorUserID:      assignedCommercialID,
			Meta:             map[string]any{"source": "admin_create_vet"},
		})
	}
	return userID, nil
}

func (s *Store) ListVetsForAdmin(ctx context.Context) ([]VetOption, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email, COALESCE(u.practice_id::text,''), COALESCE(pr.name,'')
		FROM identity.users u
		LEFT JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE u.role='vet'
		ORDER BY u.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]VetOption, 0)
	for rows.Next() {
		var v VetOption
		if err := rows.Scan(&v.UserID, &v.FullName, &v.Email, &v.PracticeID, &v.PracticeName); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
