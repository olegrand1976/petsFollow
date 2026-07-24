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

// AppInvite is a durable referral code for vet / care_pro / commercial.
type AppInvite struct {
	Code         string `json:"code"`
	UserID       string `json:"userId"`
	Role         string `json:"role"`
	PracticeID   string `json:"practiceId,omitempty"`
	PracticeName string `json:"practiceName,omitempty"`
	DisplayName  string `json:"displayName"`
	Specialty    string `json:"specialty,omitempty"`
}

// ClaimAppInviteResult is returned after applying an invite code for a client.
type ClaimAppInviteResult struct {
	Status       string `json:"status"` // linked | already_linked | referred | granted
	Kind         string `json:"kind"`   // vet | care_pro | commercial
	PracticeName string `json:"practiceName,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	PracticeID   string `json:"practiceId,omitempty"`
	InviterID    string `json:"inviterId,omitempty"`
}

// Backward-compatible aliases used by existing call sites.
type VetAppInvite = AppInvite
type ClaimVetAppInviteResult = ClaimAppInviteResult

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

func canIssueAppInvite(role kernel.Role) bool {
	switch role {
	case kernel.RoleVet, kernel.RoleCarePro, kernel.RoleCommercial, kernel.RoleCommercialManager:
		return true
	default:
		return false
	}
}

func (s *Store) scanAppInvite(row pgx.Row) (AppInvite, error) {
	var inv AppInvite
	err := row.Scan(&inv.Code, &inv.UserID, &inv.Role, &inv.PracticeID, &inv.PracticeName, &inv.DisplayName, &inv.Specialty)
	return inv, err
}

const appInviteSelect = `
	SELECT c.code, c.user_id::text, c.role, COALESCE(c.practice_id::text,''), COALESCE(pr.name,''), u.full_name,
		COALESCE(u.professional_specialty,'')
	FROM practice.app_invite_codes c
	JOIN identity.users u ON u.id = c.user_id
	LEFT JOIN practice.practices pr ON pr.id = c.practice_id`

// EnsureAppInviteCode returns the durable invite code for an eligible user.
func (s *Store) EnsureAppInviteCode(ctx context.Context, userID string) (AppInvite, error) {
	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return AppInvite{}, err
	}
	if !canIssueAppInvite(u.Role) {
		return AppInvite{}, ErrForbidden
	}
	if u.Role == kernel.RoleVet && u.PracticeID == "" {
		return AppInvite{}, ErrNotFound
	}

	inv, err := s.scanAppInvite(s.pool.QueryRow(ctx, appInviteSelect+` WHERE c.user_id = $1`, userID))
	if err == nil {
		// Keep practice_id in sync if the vet moved cabinets.
		if u.Role == kernel.RoleVet && u.PracticeID != "" && inv.PracticeID != u.PracticeID {
			if _, updErr := s.pool.Exec(ctx, `
				UPDATE practice.app_invite_codes SET practice_id = $2 WHERE user_id = $1`,
				userID, u.PracticeID); updErr != nil {
				return AppInvite{}, updErr
			}
			return s.scanAppInvite(s.pool.QueryRow(ctx, appInviteSelect+` WHERE c.user_id = $1`, userID))
		}
		return inv, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return AppInvite{}, err
	}

	var practicePtr any
	if u.PracticeID != "" && u.Role == kernel.RoleVet {
		practicePtr = u.PracticeID
	} else {
		practicePtr = nil
	}

	for attempt := 0; attempt < 8; attempt++ {
		code, genErr := generateInviteCode()
		if genErr != nil {
			return AppInvite{}, genErr
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO practice.app_invite_codes (user_id, role, practice_id, code)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO NOTHING`,
			userID, string(u.Role), practicePtr, code)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return AppInvite{}, err
		}
		return s.EnsureAppInviteCode(ctx, userID)
	}
	return AppInvite{}, ErrConflict
}

// EnsureVetAppInviteCode is kept for call sites that expect vet-only creation.
func (s *Store) EnsureVetAppInviteCode(ctx context.Context, vetUserID string) (AppInvite, error) {
	inv, err := s.EnsureAppInviteCode(ctx, vetUserID)
	if err != nil {
		return AppInvite{}, err
	}
	if inv.Role != string(kernel.RoleVet) {
		return AppInvite{}, ErrForbidden
	}
	return inv, nil
}

// GetAppInviteByCode resolves a public invite code.
func (s *Store) GetAppInviteByCode(ctx context.Context, code string) (AppInvite, error) {
	code = NormalizeInviteCode(code)
	if code == "" {
		return AppInvite{}, ErrNotFound
	}
	inv, err := s.scanAppInvite(s.pool.QueryRow(ctx, appInviteSelect+` WHERE c.code = $1`, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return AppInvite{}, ErrNotFound
	}
	if err != nil {
		return AppInvite{}, err
	}
	return inv, nil
}

// GetVetAppInviteByCode resolves invite (any role) — public landing uses display fields.
func (s *Store) GetVetAppInviteByCode(ctx context.Context, code string) (AppInvite, error) {
	return s.GetAppInviteByCode(ctx, code)
}

// ClaimAppInvite applies role-specific linking for a client invite code.
func (s *Store) ClaimAppInvite(ctx context.Context, clientUserID, code string) (ClaimAppInviteResult, error) {
	inv, err := s.GetAppInviteByCode(ctx, code)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	client, err := s.GetUserByID(ctx, clientUserID)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	if client.Role != kernel.RoleClient {
		return ClaimAppInviteResult{}, ErrValidation
	}

	switch inv.Role {
	case string(kernel.RoleVet):
		return s.claimVetInvite(ctx, clientUserID, inv)
	case string(kernel.RoleCarePro):
		return s.claimCareProInvite(ctx, clientUserID, inv)
	case string(kernel.RoleCommercial), string(kernel.RoleCommercialManager):
		return s.claimCommercialInvite(ctx, clientUserID, inv)
	default:
		return ClaimAppInviteResult{}, ErrValidation
	}
}

// ClaimVetAppInvite keeps the previous name for handlers/tryClaim.
func (s *Store) ClaimVetAppInvite(ctx context.Context, clientUserID, code string) (ClaimAppInviteResult, error) {
	return s.ClaimAppInvite(ctx, clientUserID, code)
}

func (s *Store) claimVetInvite(ctx context.Context, clientUserID string, inv AppInvite) (ClaimAppInviteResult, error) {
	if inv.PracticeID == "" {
		return ClaimAppInviteResult{}, ErrNotFound
	}
	already, err := s.ClientIsMemberOfPractice(ctx, clientUserID, inv.PracticeID)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	defer tx.Rollback(ctx)

	// First membership wins — do not reassign vet_user_id on re-scan.
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
		uuid.NewString(), inv.PracticeID, clientUserID, inv.UserID); err != nil {
		return ClaimAppInviteResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		SELECT $1, $2, $3, $4, NULL
		WHERE NOT EXISTS (
			SELECT 1 FROM messaging.threads
			WHERE practice_id=$2 AND client_user_id=$3 AND vet_user_id=$4 AND pet_id IS NULL
		)`, uuid.NewString(), inv.PracticeID, clientUserID, inv.UserID); err != nil {
		return ClaimAppInviteResult{}, err
	}
	_, _ = tx.Exec(ctx, `
		UPDATE practice.client_vet_link_requests
		SET status = 'accepted', updated_at = NOW()
		WHERE client_user_id = $1 AND practice_id = $2 AND status = 'pending'`,
		clientUserID, inv.PracticeID)

	if err := tx.Commit(ctx); err != nil {
		return ClaimAppInviteResult{}, err
	}

	status := "linked"
	if already {
		status = "already_linked"
	}
	return ClaimAppInviteResult{
		Status:       status,
		Kind:         "vet",
		PracticeName: inv.PracticeName,
		DisplayName:  inv.DisplayName,
		PracticeID:   inv.PracticeID,
		InviterID:    inv.UserID,
	}, nil
}

func (s *Store) claimCareProInvite(ctx context.Context, clientUserID string, inv AppInvite) (ClaimAppInviteResult, error) {
	var already bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.client_access
			WHERE client_user_id=$1 AND grantee_user_id=$2
				AND (expires_at IS NULL OR expires_at > NOW())
		)`, clientUserID, inv.UserID).Scan(&already)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	if already {
		return ClaimAppInviteResult{
			Status:      "already_linked",
			Kind:        "care_pro",
			DisplayName: inv.DisplayName,
			InviterID:   inv.UserID,
		}, nil
	}
	if _, err := s.GrantClientAccess(ctx, clientUserID, inv.UserID, inv.UserID, string(PermWriteNotes), nil); err != nil {
		return ClaimAppInviteResult{}, err
	}
	return ClaimAppInviteResult{
		Status:      "granted",
		Kind:        "care_pro",
		DisplayName: inv.DisplayName,
		InviterID:   inv.UserID,
	}, nil
}

func (s *Store) claimCommercialInvite(ctx context.Context, clientUserID string, inv AppInvite) (ClaimAppInviteResult, error) {
	// First referral wins — do not overwrite an existing commercial attribution.
	var existingCommercial string
	err := s.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals
		WHERE client_user_id=$1`, clientUserID).Scan(&existingCommercial)
	if err == nil {
		status := "already_linked"
		if existingCommercial == inv.UserID {
			status = "already_linked"
		}
		return ClaimAppInviteResult{
			Status:      status,
			Kind:        "commercial",
			DisplayName: inv.DisplayName,
			InviterID:   inv.UserID,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ClaimAppInviteResult{}, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (client_user_id) DO NOTHING`,
		clientUserID, inv.UserID, inv.Code)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	return ClaimAppInviteResult{
		Status:      "referred",
		Kind:        "commercial",
		DisplayName: inv.DisplayName,
		InviterID:   inv.UserID,
	}, nil
}
