package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type RegisterClientInput struct {
	Email           string
	Password        string
	FullName        string
	PreferredLocale string
	// TermsAccepted horodate le consentement CGU/privacy (RGPD art. 7).
	TermsAccepted bool
}

type RegisterClientResult struct {
	UserID string
	Token  string
}

type RegisterCareProInput struct {
	Email           string
	Password        string
	FullName        string
	Specialty       kernel.ProfessionalSpecialty
	PreferredLocale string
}

type RegisterCareProResult struct {
	UserID string
	Token  string
}

func (s *Store) RegisterClient(ctx context.Context, in RegisterClientInput) (RegisterClientResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterClientResult{}, err
	}
	userID := uuid.NewString()
	token := uuid.NewString()
	expires := time.Now().Add(48 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RegisterClientResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (id, email, password_hash, full_name, role, practice_id, preferred_locale, terms_accepted_at)
		VALUES ($1, $2, $3, $4, 'client', NULL, $5, CASE WHEN $6 THEN NOW() END)`,
		userID, in.Email, string(hash), in.FullName, i18n.NormalizeLocale(in.PreferredLocale), in.TermsAccepted); err != nil {
		return RegisterClientResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), userID, token, expires); err != nil {
		return RegisterClientResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RegisterClientResult{}, err
	}
	_ = s.EnrollEmailJourney(ctx, userID, time.Now().UTC())
	return RegisterClientResult{UserID: userID, Token: token}, nil
}

func (s *Store) RegisterCarePro(ctx context.Context, in RegisterCareProInput) (RegisterCareProResult, error) {
	if !kernel.ValidSpecialty(in.Specialty) {
		return RegisterCareProResult{}, ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterCareProResult{}, err
	}
	userID := uuid.NewString()
	token := uuid.NewString()
	expires := time.Now().Add(48 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RegisterCareProResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id, preferred_locale, professional_specialty
		) VALUES ($1, $2, $3, $4, 'care_pro', NULL, $5, $6)`,
		userID, in.Email, string(hash), in.FullName,
		i18n.NormalizeLocale(in.PreferredLocale), string(in.Specialty)); err != nil {
		return RegisterCareProResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity.email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), userID, token, expires); err != nil {
		return RegisterCareProResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RegisterCareProResult{}, err
	}
	return RegisterCareProResult{UserID: userID, Token: token}, nil
}

// CreateCareProAsAdmin provisions a verified care_pro (must change password on first login).
func (s *Store) CreateCareProAsAdmin(ctx context.Context, email, password, fullName string, specialty kernel.ProfessionalSpecialty, locale string) (string, error) {
	if !kernel.ValidSpecialty(specialty) {
		return "", ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	userID := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO identity.users (
			id, email, password_hash, full_name, role, practice_id,
			email_verified_at, preferred_locale, professional_specialty, must_change_password
		) VALUES ($1, $2, $3, $4, 'care_pro', NULL, NOW(), $5, $6, true)`,
		userID, email, string(hash), fullName, i18n.NormalizeLocale(locale), string(specialty))
	if err != nil {
		return "", err
	}
	return userID, nil
}
