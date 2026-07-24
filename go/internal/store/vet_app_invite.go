package store

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const inviteCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// VetAppInvite is a durable referral code for a vet (QR / email / deep link).
type VetAppInvite struct {
	Code         string `json:"code"`
	VetUserID    string `json:"vetUserId"`
	PracticeID   string `json:"practiceId"`
	PracticeName string `json:"practiceName"`
	VetFullName  string `json:"vetFullName"`
}

// ClaimVetAppInviteResult is returned after auto-linking a client via invite code.
type ClaimVetAppInviteResult struct {
	Status       string `json:"status"` // linked | already_linked
	PracticeName string `json:"practiceName"`
	VetFullName  string `json:"vetFullName"`
	PracticeID   string `json:"practiceId"`
	VetUserID    string `json:"vetUserId"`
}

func NormalizeInviteCode(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func generateInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = inviteCodeAlphabet[int(b[i])%len(inviteCodeAlphabet)]
	}
	return string(b), nil
}

// EnsureVetAppInviteCode returns the durable invite code for a vet, creating one if needed.
func (s *Store) EnsureVetAppInviteCode(ctx context.Context, vetUserID string) (VetAppInvite, error) {
	var inv VetAppInvite
	err := s.pool.QueryRow(ctx, `
		SELECT c.code, c.vet_user_id::text, c.practice_id::text, pr.name, u.full_name
		FROM practice.vet_app_invite_codes c
		JOIN practice.practices pr ON pr.id = c.practice_id
		JOIN identity.users u ON u.id = c.vet_user_id
		WHERE c.vet_user_id = $1`, vetUserID).
		Scan(&inv.Code, &inv.VetUserID, &inv.PracticeID, &inv.PracticeName, &inv.VetFullName)
	if err == nil {
		return inv, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return VetAppInvite{}, err
	}

	var practiceID, vetName, practiceName string
	err = s.pool.QueryRow(ctx, `
		SELECT u.practice_id::text, u.full_name, pr.name
		FROM identity.users u
		JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE u.id = $1 AND u.role = 'vet' AND u.practice_id IS NOT NULL`, vetUserID).
		Scan(&practiceID, &vetName, &practiceName)
	if errors.Is(err, pgx.ErrNoRows) {
		return VetAppInvite{}, ErrNotFound
	}
	if err != nil {
		return VetAppInvite{}, err
	}

	for attempt := 0; attempt < 8; attempt++ {
		code, genErr := generateInviteCode()
		if genErr != nil {
			return VetAppInvite{}, genErr
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO practice.vet_app_invite_codes (vet_user_id, practice_id, code)
			VALUES ($1, $2, $3)
			ON CONFLICT (vet_user_id) DO NOTHING`, vetUserID, practiceID, code)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return VetAppInvite{}, err
		}
		return s.EnsureVetAppInviteCode(ctx, vetUserID)
	}
	return VetAppInvite{}, ErrConflict
}

// GetVetAppInviteByCode resolves a public invite code.
func (s *Store) GetVetAppInviteByCode(ctx context.Context, code string) (VetAppInvite, error) {
	code = NormalizeInviteCode(code)
	if code == "" {
		return VetAppInvite{}, ErrNotFound
	}
	var inv VetAppInvite
	err := s.pool.QueryRow(ctx, `
		SELECT c.code, c.vet_user_id::text, c.practice_id::text, pr.name, u.full_name
		FROM practice.vet_app_invite_codes c
		JOIN practice.practices pr ON pr.id = c.practice_id
		JOIN identity.users u ON u.id = c.vet_user_id
		WHERE c.code = $1`, code).
		Scan(&inv.Code, &inv.VetUserID, &inv.PracticeID, &inv.PracticeName, &inv.VetFullName)
	if errors.Is(err, pgx.ErrNoRows) {
		return VetAppInvite{}, ErrNotFound
	}
	if err != nil {
		return VetAppInvite{}, err
	}
	return inv, nil
}

// ClaimVetAppInvite auto-links a client to the inviting vet's practice (idempotent).
func (s *Store) ClaimVetAppInvite(ctx context.Context, clientUserID, code string) (ClaimVetAppInviteResult, error) {
	inv, err := s.GetVetAppInviteByCode(ctx, code)
	if err != nil {
		return ClaimVetAppInviteResult{}, err
	}
	client, err := s.GetUserByID(ctx, clientUserID)
	if err != nil {
		return ClaimVetAppInviteResult{}, err
	}
	if client.Role != kernel.RoleClient {
		return ClaimVetAppInviteResult{}, ErrValidation
	}

	already, err := s.ClientIsMemberOfPractice(ctx, clientUserID, inv.PracticeID)
	if err != nil {
		return ClaimVetAppInviteResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClaimVetAppInviteResult{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO UPDATE SET vet_user_id = EXCLUDED.vet_user_id`,
		uuid.NewString(), inv.PracticeID, clientUserID, inv.VetUserID); err != nil {
		return ClaimVetAppInviteResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		SELECT $1, $2, $3, $4, NULL
		WHERE NOT EXISTS (
			SELECT 1 FROM messaging.threads
			WHERE practice_id=$2 AND client_user_id=$3 AND vet_user_id=$4 AND pet_id IS NULL
		)`, uuid.NewString(), inv.PracticeID, clientUserID, inv.VetUserID); err != nil {
		return ClaimVetAppInviteResult{}, err
	}
	_, _ = tx.Exec(ctx, `
		UPDATE practice.client_vet_link_requests
		SET status = 'accepted', updated_at = NOW()
		WHERE client_user_id = $1 AND practice_id = $2 AND status = 'pending'`,
		clientUserID, inv.PracticeID)

	if err := tx.Commit(ctx); err != nil {
		return ClaimVetAppInviteResult{}, err
	}

	status := "linked"
	if already {
		status = "already_linked"
	}
	return ClaimVetAppInviteResult{
		Status:       status,
		PracticeName: inv.PracticeName,
		VetFullName:  inv.VetFullName,
		PracticeID:   inv.PracticeID,
		VetUserID:    inv.VetUserID,
	}, nil
}
