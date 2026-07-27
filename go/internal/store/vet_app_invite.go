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
	Kind         string `json:"kind"`   // vet | care_pro | commercial | client
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
	case kernel.RoleVet, kernel.RoleCarePro, kernel.RoleCommercial, kernel.RoleCommercialManager, kernel.RoleClient:
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
	case string(kernel.RoleClient):
		return s.claimClientInvite(ctx, clientUserID, inv)
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

	stamped, err := linkClientToPracticeTx(ctx, tx, clientUserID, inv.PracticeID, inv.UserID)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ClaimAppInviteResult{}, err
	}
	seedCareForStampedPets(ctx, s, clientUserID, inv.PracticeID, stamped)

	// Honour first membership: report the vet actually stored on practice_clients.
	linkedVet := inv.UserID
	_ = s.pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM practice.practice_clients
		WHERE practice_id=$1 AND client_user_id=$2`,
		inv.PracticeID, clientUserID).Scan(&linkedVet)

	status := "linked"
	if already {
		status = "already_linked"
	} else {
		var assignedComm string
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`, linkedVet).Scan(&assignedComm)
		_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
			EventType:        FiliationEventPracticeClientLinked,
			CommercialUserID: assignedComm,
			VetUserID:        linkedVet,
			ClientUserID:     clientUserID,
			PracticeID:       inv.PracticeID,
			ActorUserID:      clientUserID,
			InviteCode:       inv.Code,
			Meta:             map[string]any{"source": "claim_vet_invite"},
		})
	}
	return ClaimAppInviteResult{
		Status:       status,
		Kind:         "vet",
		PracticeName: inv.PracticeName,
		DisplayName:  inv.DisplayName,
		PracticeID:   inv.PracticeID,
		InviterID:    linkedVet,
	}, nil
}

// linkClientToPracticeTx inserts practice_clients (first-wins), stamps primary practice_id
// when still NULL, stamps orphan pets, ensures messaging thread, accepts pending link requests.
func linkClientToPracticeTx(ctx context.Context, tx pgx.Tx, clientUserID, practiceID, vetUserID string) ([]stampedOrphanPet, error) {
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO NOTHING`,
		uuid.NewString(), practiceID, clientUserID, vetUserID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users
		SET practice_id = $2::uuid
		WHERE id = $1 AND role = 'client' AND practice_id IS NULL`,
		clientUserID, practiceID); err != nil {
		return nil, err
	}
	stamped, err := stampOrphanPetsTx(ctx, tx, clientUserID, practiceID)
	if err != nil {
		return nil, err
	}
	// One general thread per (practice, client) when pet_id IS NULL.
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		SELECT $1, $2, $3, $4, NULL
		WHERE NOT EXISTS (
			SELECT 1 FROM messaging.threads
			WHERE practice_id=$2 AND client_user_id=$3 AND pet_id IS NULL
		)`,
		uuid.NewString(), practiceID, clientUserID, vetUserID); err != nil {
		return nil, err
	}
	_, _ = tx.Exec(ctx, `
		UPDATE practice.client_vet_link_requests
		SET status = 'accepted', updated_at = NOW()
		WHERE client_user_id = $1 AND practice_id = $2 AND status = 'pending'`,
		clientUserID, practiceID)
	return stamped, nil
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
	var existingCode string
	err := s.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text, COALESCE(invite_code,'')
		FROM practice.commercial_referrals
		WHERE client_user_id=$1`, clientUserID).Scan(&existingCommercial, &existingCode)
	if err == nil {
		// Same promoter: backfill invite_code if a nearby-link landed first without the code.
		if existingCommercial == inv.UserID && existingCode == "" && inv.Code != "" {
			_, _ = s.pool.Exec(ctx, `
				UPDATE practice.commercial_referrals
				SET invite_code = $2, updated_at = NOW()
				WHERE client_user_id = $1 AND commercial_user_id = $3
				  AND COALESCE(invite_code,'') = ''`,
				clientUserID, inv.Code, inv.UserID)
		}
		return ClaimAppInviteResult{
			Status:      "already_linked",
			Kind:        "commercial",
			DisplayName: inv.DisplayName,
			InviterID:   existingCommercial,
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
	// Re-read after insert to honour first-wins under concurrency (never claim success for the loser).
	var linked string
	var savedCode string
	err = s.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text, COALESCE(invite_code,'')
		FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientUserID).Scan(&linked, &savedCode)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	if linked != inv.UserID {
		return ClaimAppInviteResult{
			Status:      "already_linked",
			Kind:        "commercial",
			DisplayName: inv.DisplayName,
			InviterID:   linked,
		}, nil
	}
	if savedCode == "" && inv.Code != "" {
		_, _ = s.pool.Exec(ctx, `
			UPDATE practice.commercial_referrals
			SET invite_code = COALESCE(NULLIF(invite_code,''), $2), updated_at = NOW()
			WHERE client_user_id = $1 AND commercial_user_id = $3`,
			clientUserID, inv.Code, inv.UserID)
	}
	_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
		EventType:        FiliationEventClientReferral,
		CommercialUserID: inv.UserID,
		ClientUserID:     clientUserID,
		ActorUserID:      clientUserID,
		InviteCode:       inv.Code,
		Meta:             map[string]any{"source": "claim_commercial_invite"},
	})
	return ClaimAppInviteResult{
		Status:      "referred",
		Kind:        "commercial",
		DisplayName: inv.DisplayName,
		InviterID:   inv.UserID,
	}, nil
}

func (s *Store) claimClientInvite(ctx context.Context, clientUserID string, inv AppInvite) (ClaimAppInviteResult, error) {
	if clientUserID == inv.UserID {
		return ClaimAppInviteResult{}, ErrSelfReferral
	}

	// 1) First-wins client→client sponsorship.
	var existingSponsor string
	err := s.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text FROM practice.client_referrals
		WHERE referred_client_user_id=$1`, clientUserID).Scan(&existingSponsor)
	alreadySponsored := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ClaimAppInviteResult{}, err
	}
	if !alreadySponsored {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO practice.client_referrals (referred_client_user_id, sponsor_client_user_id, invite_code, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (referred_client_user_id) DO NOTHING`,
			clientUserID, inv.UserID, inv.Code)
		if err != nil {
			return ClaimAppInviteResult{}, err
		}
	}
	var linkedSponsor string
	var savedCode string
	err = s.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text, COALESCE(invite_code,'')
		FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		clientUserID).Scan(&linkedSponsor, &savedCode)
	if err != nil {
		return ClaimAppInviteResult{}, err
	}
	if linkedSponsor == inv.UserID && savedCode == "" && inv.Code != "" {
		_, _ = s.pool.Exec(ctx, `
			UPDATE practice.client_referrals
			SET invite_code = COALESCE(NULLIF(invite_code,''), $2), updated_at = NOW()
			WHERE referred_client_user_id = $1 AND sponsor_client_user_id = $3`,
			clientUserID, inv.Code, inv.UserID)
	}

	status := "referred"
	if alreadySponsored || linkedSponsor != inv.UserID {
		status = "already_linked"
	}

	// Side-effects (commercial inherit + cabinet) only when this invite's sponsor owns the referral.
	if linkedSponsor == inv.UserID {
		if err := s.inheritCommercialFromSponsor(ctx, clientUserID, inv.UserID, inv.Code); err != nil {
			return ClaimAppInviteResult{}, err
		}

		practiceID, vetUserID, practiceName, err := s.resolveSponsorPractice(ctx, inv.UserID)
		if err != nil {
			return ClaimAppInviteResult{}, err
		}
		linkedCabinet := false
		if practiceID != "" && vetUserID != "" {
			tx, txErr := s.pool.Begin(ctx)
			if txErr != nil {
				return ClaimAppInviteResult{}, txErr
			}
			// Serialize against concurrent vet/client claims; re-check free inside tx (no multi-cabinet).
			var lockedID string
			if lockErr := tx.QueryRow(ctx, `
				SELECT id::text FROM identity.users WHERE id=$1 AND role='client' FOR UPDATE`,
				clientUserID).Scan(&lockedID); lockErr != nil {
				_ = tx.Rollback(ctx)
				return ClaimAppInviteResult{}, lockErr
			}
			var attached bool
			if scanErr := tx.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM practice.practice_clients WHERE client_user_id=$1)
				    OR EXISTS(
						SELECT 1 FROM identity.users
						WHERE id=$1 AND role='client' AND practice_id IS NOT NULL
					)`, clientUserID).Scan(&attached); scanErr != nil {
				_ = tx.Rollback(ctx)
				return ClaimAppInviteResult{}, scanErr
			}
			if attached {
				_ = tx.Rollback(ctx)
			} else {
				stamped, linkErr := linkClientToPracticeTx(ctx, tx, clientUserID, practiceID, vetUserID)
				if linkErr != nil {
					_ = tx.Rollback(ctx)
					return ClaimAppInviteResult{}, linkErr
				}
				if commitErr := tx.Commit(ctx); commitErr != nil {
					return ClaimAppInviteResult{}, commitErr
				}
				seedCareForStampedPets(ctx, s, clientUserID, practiceID, stamped)
				linkedCabinet = true
			}
		}

		if linkedCabinet {
			status = "linked"
			s.RecordPracticeClientLinkedEvent(ctx, practiceID, clientUserID, vetUserID, clientUserID, map[string]any{
				"source": "claim_client_invite",
			})
		}

		result := ClaimAppInviteResult{
			Status:      status,
			Kind:        "client",
			DisplayName: inv.DisplayName,
			InviterID:   linkedSponsor,
		}
		if linkedCabinet {
			result.PracticeID = practiceID
			result.PracticeName = practiceName
		}
		return result, nil
	}

	return ClaimAppInviteResult{
		Status:      status,
		Kind:        "client",
		DisplayName: inv.DisplayName,
		InviterID:   linkedSponsor,
	}, nil
}

// resolveSponsorPractice returns the sponsor's newest practice membership (vet referent).
func (s *Store) resolveSponsorPractice(ctx context.Context, sponsorUserID string) (practiceID, vetUserID, practiceName string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT pc.practice_id::text, pc.vet_user_id::text, COALESCE(pr.name,'')
		FROM practice.practice_clients pc
		LEFT JOIN practice.practices pr ON pr.id = pc.practice_id
		WHERE pc.client_user_id=$1
		ORDER BY pc.created_at DESC NULLS LAST, pc.practice_id
		LIMIT 1`, sponsorUserID).Scan(&practiceID, &vetUserID, &practiceName)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", nil
	}
	return practiceID, vetUserID, practiceName, err
}

// inheritCommercialFromSponsor seeds commercial_referrals for the filleul (first-wins).
// Order: sponsor's commercial_referrals, else assigned_commercial_id of sponsor's referent vet.
func (s *Store) inheritCommercialFromSponsor(ctx context.Context, filleulID, sponsorID, inviteCode string) error {
	var existing string
	err := s.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals
		WHERE client_user_id=$1`, filleulID).Scan(&existing)
	if err == nil {
		return nil // first-wins — already attributed
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	commercialID := ""
	err = s.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals
		WHERE client_user_id=$1`, sponsorID).Scan(&commercialID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if commercialID == "" {
		err = s.pool.QueryRow(ctx, `
			SELECT u.assigned_commercial_id::text
			FROM practice.practice_clients pc
			JOIN identity.users u ON u.id = pc.vet_user_id
			WHERE pc.client_user_id=$1 AND u.assigned_commercial_id IS NOT NULL
			ORDER BY pc.created_at DESC NULLS LAST, pc.practice_id
			LIMIT 1`, sponsorID).Scan(&commercialID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			commercialID = ""
		}
	}
	if commercialID == "" {
		return nil
	}

	tag, err := s.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (client_user_id) DO NOTHING`,
		filleulID, commercialID, inviteCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
			EventType:        FiliationEventClientReferral,
			CommercialUserID: commercialID,
			ClientUserID:     filleulID,
			ActorUserID:      sponsorID,
			InviteCode:       inviteCode,
			Meta:             map[string]any{"source": "sponsor_inherit"},
		})
	}
	return nil
}
