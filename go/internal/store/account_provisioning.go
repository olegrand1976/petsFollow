package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"golang.org/x/crypto/bcrypt"
)

type CreateClientInput struct {
	Email       string
	Password    string
	FullName    string
	Locale      string
	SkipJourney bool
}

type VetOption struct {
	UserID       string `json:"userId"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	PracticeName string `json:"practiceName"`
}

func (s *Store) CreateClientForVet(ctx context.Context, vetUserID string, in CreateClientInput) (string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var practiceID string
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,'') FROM identity.users
		WHERE id=$1 AND role='vet'`, vetUserID).Scan(&practiceID)
	if errors.Is(err, pgx.ErrNoRows) || practiceID == "" {
		return "", ErrNotFound
	}
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
			id, email, password_hash, full_name, role, practice_id,
			email_verified_at, preferred_locale, must_change_password
		) VALUES ($1, $2, $3, $4, 'client', $5, NOW(), $6, true)`,
		clientID, in.Email, string(hash), in.FullName, practiceID, i18n.NormalizeLocale(in.Locale)); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), practiceID, clientID, vetUserID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		VALUES ($1, $2, $3, $4, NULL)`,
		uuid.NewString(), practiceID, clientID, vetUserID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	// Best-effort: start in-app discovery + email loyalty journey (skipped for bulk imports).
	if !in.SkipJourney {
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
	return userID, nil
}

func (s *Store) ListVetsForAdmin(ctx context.Context) ([]VetOption, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email, COALESCE(pr.name,'')
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
		if err := rows.Scan(&v.UserID, &v.FullName, &v.Email, &v.PracticeName); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
