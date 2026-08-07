package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type TeamRole string

const (
	TeamRoleReferenceVet TeamRole = "reference_vet"
	TeamRoleVet          TeamRole = "vet"
	TeamRoleAssistant    TeamRole = "vet_assistant"
	TeamRoleSecretary    TeamRole = "secretary"
)

// DefaultTeamPermissions returns the product default matrix for a team role.
func DefaultTeamPermissions(role TeamRole) map[string]bool {
	full := map[string]bool{
		"clients.read": true, "clients.write": true,
		"pets.read": true, "pets.write_clinical": true,
		"heartrate.validate": true, "messaging": true,
		"calendar.manage": true, "consultations.history.read": true,
		"care.manage": true,
		"shares.read": true, "shares.manage": true,
		"pharmacy.read": true, "pharmacy.write": true,
		"practice.settings": true,
		"team.manage":       true, "commissions.view": true,
	}
	switch role {
	case TeamRoleReferenceVet:
		return full
	case TeamRoleVet:
		full["practice.settings"] = false
		full["team.manage"] = false
		full["commissions.view"] = false
		return full
	case TeamRoleAssistant:
		return map[string]bool{
			"clients.read": true, "clients.write": true,
			"pets.read": true, "pets.write_clinical": true,
			"heartrate.validate": false, "messaging": true,
			"calendar.manage": true, "consultations.history.read": true,
			"care.manage": true,
			"shares.read": true, "shares.manage": false,
			"pharmacy.read": true, "pharmacy.write": true,
			"practice.settings": false,
			"team.manage":       false, "commissions.view": false,
		}
	case TeamRoleSecretary:
		return map[string]bool{
			"clients.read": true, "clients.write": true,
			"pets.read": true, "pets.write_clinical": false,
			"heartrate.validate": false, "messaging": true,
			"calendar.manage": true, "consultations.history.read": false,
			"care.manage": false,
			"shares.read": true, "shares.manage": false,
			"pharmacy.read": true, "pharmacy.write": false,
			"practice.settings": false,
			"team.manage":       false, "commissions.view": false,
		}
	default:
		return map[string]bool{}
	}
}

type TeamMember struct {
	ID                string          `json:"id"`
	PracticeID        string          `json:"practiceId"`
	UserID            string          `json:"userId"`
	ProfileID         string          `json:"profileId,omitempty"`
	TeamRole          TeamRole        `json:"teamRole"`
	Permissions       map[string]bool `json:"permissions"`
	Status            string          `json:"status"`
	Email             string          `json:"email"`
	FullName          string          `json:"fullName"`
	DefaultSiteID     string          `json:"defaultSiteId,omitempty"`
	IncludeInCalendar bool            `json:"includeInCalendar"`
	CreatedAt         time.Time       `json:"createdAt"`
}

// hardDeniedCapabilities cannot be elevated via override for non-reference roles.
func hardDeniedCapabilities(role TeamRole) map[string]bool {
	switch role {
	case TeamRoleSecretary:
		return map[string]bool{
			"pets.write_clinical": true, "pharmacy.write": true,
			"heartrate.validate": true,
			"shares.manage":      true, "practice.settings": true,
			"team.manage": true, "commissions.view": true,
		}
	case TeamRoleAssistant:
		return map[string]bool{
			"heartrate.validate": true, "shares.manage": true,
			"practice.settings": true, "team.manage": true, "commissions.view": true,
		}
	case TeamRoleVet:
		return map[string]bool{
			"practice.settings": true, "team.manage": true, "commissions.view": true,
		}
	default:
		return nil
	}
}

func mergeTeamPermissions(role TeamRole, override map[string]bool) map[string]bool {
	out := DefaultTeamPermissions(role)
	denied := hardDeniedCapabilities(role)
	_, pharmacyWriteInOverride := override["pharmacy.write"]
	for k, v := range override {
		if v && denied[k] {
			continue
		}
		out[k] = v
	}
	// Compat overrides pré-pharmacy.* : un toggle pets.write_clinical seul
	// pilotait aussi stock/DAF — on réplique tant que pharmacy.write n’est pas explicite.
	if !pharmacyWriteInOverride {
		if v, ok := override["pets.write_clinical"]; ok {
			if v && denied["pharmacy.write"] {
				out["pharmacy.write"] = false
			} else {
				out["pharmacy.write"] = v
			}
		}
	}
	// Write/manage imply the matching read capability.
	if out["shares.manage"] {
		out["shares.read"] = true
	}
	if out["pharmacy.write"] {
		out["pharmacy.read"] = true
	}
	return out
}

// ListClinicalStaffUserIDs returns active team members with pets.write_clinical.
func (s *Store) ListClinicalStaffUserIDs(ctx context.Context, practiceID string) ([]string, error) {
	members, err := s.ListTeamMembers(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	var out []string
	seen := map[string]bool{}
	for _, m := range members {
		if m.Status != "active" {
			continue
		}
		if !m.Permissions["pets.write_clinical"] {
			continue
		}
		if m.UserID == "" || seen[m.UserID] {
			continue
		}
		seen[m.UserID] = true
		out = append(out, m.UserID)
	}
	// Legacy solo reference vet may not appear in team_members filters — include practice ref.
	if ref, err := s.PracticeReferenceVetUserID(ctx, practiceID); err == nil && ref != "" && !seen[ref] {
		out = append(out, ref)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (s *Store) ListTeamMembers(ctx context.Context, practiceID string) ([]TeamMember, error) {
	// Hide multi-switch demo accounts whose home profile is not practice staff
	// (admin/commercial attached to VetPlus for switch demos only).
	rows, err := s.pool.Query(ctx, `
		SELECT tm.id::text, tm.practice_id::text, tm.user_id::text, COALESCE(tm.profile_id::text,''),
			tm.team_role, tm.permissions, tm.status, u.email, u.full_name, tm.created_at,
			COALESCE(tm.default_site_id::text,''), COALESCE(tm.include_in_calendar, true)
		FROM practice.team_members tm
		JOIN identity.users u ON u.id = tm.user_id
		WHERE tm.practice_id = $1 AND tm.status <> 'revoked'
		  AND (
			SELECT p.role FROM identity.profiles p
			WHERE p.user_id = tm.user_id AND p.role <> 'client'
			ORDER BY p.created_at ASC, p.id ASC
			LIMIT 1
		  ) IN ('vet', 'vet_assistant', 'secretary')
		ORDER BY CASE tm.team_role
			WHEN 'reference_vet' THEN 0 WHEN 'vet' THEN 1 WHEN 'vet_assistant' THEN 2 ELSE 3 END,
			tm.created_at`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TeamMember
	for rows.Next() {
		m, err := scanTeamMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func scanTeamMember(row pgx.Row) (TeamMember, error) {
	var m TeamMember
	var raw []byte
	err := row.Scan(&m.ID, &m.PracticeID, &m.UserID, &m.ProfileID, &m.TeamRole, &raw, &m.Status, &m.Email, &m.FullName, &m.CreatedAt, &m.DefaultSiteID, &m.IncludeInCalendar)
	if err != nil {
		return TeamMember{}, err
	}
	var override map[string]bool
	if len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &override)
	}
	m.Permissions = mergeTeamPermissions(m.TeamRole, override)
	return m, nil
}

func (s *Store) IsReferenceVet(ctx context.Context, practiceID, userID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.team_members
			WHERE practice_id=$1 AND user_id=$2 AND team_role='reference_vet' AND status='active'
		) OR EXISTS(
			SELECT 1 FROM practice.practices WHERE id=$1 AND reference_vet_user_id=$2
		)`, practiceID, userID).Scan(&ok)
	return ok, err
}

// PracticeReferenceVetUserID returns practices.reference_vet_user_id (empty if unset).
func (s *Store) PracticeReferenceVetUserID(ctx context.Context, practiceID string) (string, error) {
	var ref string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(reference_vet_user_id::text,'') FROM practice.practices WHERE id=$1`, practiceID,
	).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return ref, err
}

// HasActivePracticeStaffAccess reports whether userID may act as practice staff.
// Mirrors TeamPermission legacy: active membership → true; any revoked row → false;
// else legacy solo vet (role=vet + practice_id match) → true.
func (s *Store) HasActivePracticeStaffAccess(ctx context.Context, practiceID, userID string) (bool, error) {
	var active bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.team_members
			WHERE practice_id=$1 AND user_id=$2 AND status='active'
		)`, practiceID, userID).Scan(&active)
	if err != nil {
		return false, err
	}
	if active {
		return true, nil
	}
	var anyRow bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.team_members
			WHERE practice_id=$1 AND user_id=$2
		)`, practiceID, userID).Scan(&anyRow); err != nil {
		return false, err
	}
	if anyRow {
		return false, nil
	}
	var isVet bool
	err = s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM identity.users WHERE id=$1 AND practice_id=$2 AND role='vet')`,
		userID, practiceID).Scan(&isVet)
	return isVet, err
}

func (s *Store) TeamPermission(ctx context.Context, practiceID, userID, capability string) (bool, error) {
	perms, err := s.PracticePermissions(ctx, practiceID, userID)
	if err != nil {
		return false, err
	}
	return perms[capability], nil
}

// PracticePermissions returns the effective capability map for a practice staff member.
// Empty map when the user has no active staff access on the practice.
func (s *Store) PracticePermissions(ctx context.Context, practiceID, userID string) (map[string]bool, error) {
	var role TeamRole
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT team_role, permissions FROM practice.team_members
		WHERE practice_id=$1 AND user_id=$2 AND status='active'`, practiceID, userID,
	).Scan(&role, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		ok, aerr := s.HasActivePracticeStaffAccess(ctx, practiceID, userID)
		if aerr != nil {
			return nil, aerr
		}
		if !ok {
			return map[string]bool{}, nil
		}
		// Legacy solo vet without any team_members row
		return DefaultTeamPermissions(TeamRoleReferenceVet), nil
	}
	if err != nil {
		return nil, err
	}
	var override map[string]bool
	if len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &override)
	}
	return mergeTeamPermissions(role, override), nil
}

func teamRoleToKernel(r TeamRole) kernel.Role {
	switch r {
	case TeamRoleAssistant:
		return kernel.RoleVetAssistant
	case TeamRoleSecretary:
		return kernel.RoleSecretary
	default:
		return kernel.RoleVet
	}
}

type InviteTeamMemberInput struct {
	Email       string
	FullName    string
	Password    string // temp password if creating
	TeamRole    TeamRole
	Permissions map[string]bool // optional override
}

func (s *Store) InviteTeamMember(ctx context.Context, practiceID, invitedBy string, in InviteTeamMemberInput) (TeamMember, error) {
	ok, err := s.IsReferenceVet(ctx, practiceID, invitedBy)
	if err != nil {
		return TeamMember{}, err
	}
	if !ok {
		return TeamMember{}, ErrForbidden
	}
	role := in.TeamRole
	if role == "" || role == TeamRoleReferenceVet {
		return TeamMember{}, ErrValidation
	}
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" {
		return TeamMember{}, ErrValidation
	}

	profileRole := teamRoleToKernel(role)

	u, err := s.GetUserByEmail(ctx, email)
	userID := ""
	created := false
	if errors.Is(err, ErrNotFound) {
		if strings.TrimSpace(in.Password) == "" || strings.TrimSpace(in.FullName) == "" {
			return TeamMember{}, ErrValidation
		}
		hash, herr := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if herr != nil {
			return TeamMember{}, herr
		}
		userID = uuid.NewString()
		_, err = s.pool.Exec(ctx, `
			INSERT INTO identity.users (
				id, email, password_hash, full_name, role, practice_id,
				email_verified_at, must_change_password
			) VALUES ($1, $2, $3, $4, $5, $6, NOW(), true)`,
			userID, email, string(hash), in.FullName, string(profileRole), practiceID)
		if err != nil {
			return TeamMember{}, err
		}
		_ = s.EnsureUserProfiles(ctx, userID)
		created = true
	} else if err != nil {
		return TeamMember{}, err
	} else {
		userID = u.ID
		// Demo team accounts: re-seed resets password + must_change for deterministic e2e/smoke.
		if strings.HasSuffix(email, "@petsfollow.test") && strings.TrimSpace(in.Password) != "" {
			hash, herr := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
			if herr != nil {
				return TeamMember{}, herr
			}
			if _, uerr := s.pool.Exec(ctx, `
				UPDATE identity.users
				SET password_hash = $2, must_change_password = true, full_name = COALESCE(NULLIF($3,''), full_name)
				WHERE id = $1`,
				userID, string(hash), strings.TrimSpace(in.FullName)); uerr != nil {
				return TeamMember{}, uerr
			}
		}
		if _, err := s.AttachProfile(ctx, invitedBy, userID, AttachProfileInput{
			Role: profileRole, PracticeID: practiceID,
		}); err != nil && !errors.Is(err, ErrProfileExists) {
			return TeamMember{}, err
		}
		_ = s.EnsureUserProfiles(ctx, userID)
	}

	profileID := mustProfileID(ctx, s, userID, profileRole)
	var permJSON any
	if in.Permissions != nil {
		b, _ := json.Marshal(in.Permissions)
		permJSON = b
	}
	memberID := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO practice.team_members (id, practice_id, user_id, profile_id, team_role, permissions, status, invited_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, 'active', $7)
		ON CONFLICT (practice_id, user_id) DO UPDATE SET
			profile_id = EXCLUDED.profile_id,
			team_role = EXCLUDED.team_role,
			permissions = EXCLUDED.permissions,
			status = 'active',
			invited_by_user_id = EXCLUDED.invited_by_user_id`,
		memberID, practiceID, userID, nullIfEmpty(profileID), string(role), permJSON, invitedBy)
	if err != nil {
		return TeamMember{}, err
	}
	// Activate staff profile only after team membership exists (SwitchProfile gates on it).
	if created && profileID != "" {
		_, _ = s.SwitchProfile(ctx, userID, profileID)
	}

	members, err := s.ListTeamMembers(ctx, practiceID)
	if err != nil {
		return TeamMember{}, err
	}
	for _, m := range members {
		if m.UserID == userID {
			return m, nil
		}
	}
	return TeamMember{}, ErrNotFound
}

func mustProfileID(ctx context.Context, s *Store, userID string, role kernel.Role) string {
	var id string
	_ = s.pool.QueryRow(ctx, `SELECT id::text FROM identity.profiles WHERE user_id=$1 AND role=$2`, userID, string(role)).Scan(&id)
	return id
}

func (s *Store) UpdateTeamMember(ctx context.Context, practiceID, actorUserID, memberID string, teamRole *TeamRole, permissions map[string]bool, defaultSiteID *string, includeInCalendar *bool) (TeamMember, error) {
	ok, err := s.IsReferenceVet(ctx, practiceID, actorUserID)
	if err != nil {
		return TeamMember{}, err
	}
	if !ok {
		return TeamMember{}, ErrForbidden
	}
	var cur TeamMember
	var raw []byte
	err = s.pool.QueryRow(ctx, `
		SELECT tm.id::text, tm.practice_id::text, tm.user_id::text, COALESCE(tm.profile_id::text,''),
			tm.team_role, tm.permissions, tm.status, u.email, u.full_name, tm.created_at,
			COALESCE(tm.default_site_id::text,''), COALESCE(tm.include_in_calendar, true)
		FROM practice.team_members tm
		JOIN identity.users u ON u.id = tm.user_id
		WHERE tm.id=$1 AND tm.practice_id=$2`, memberID, practiceID,
	).Scan(&cur.ID, &cur.PracticeID, &cur.UserID, &cur.ProfileID, &cur.TeamRole, &raw, &cur.Status, &cur.Email, &cur.FullName, &cur.CreatedAt, &cur.DefaultSiteID, &cur.IncludeInCalendar)
	if errors.Is(err, pgx.ErrNoRows) {
		return TeamMember{}, ErrNotFound
	}
	if err != nil {
		return TeamMember{}, err
	}
	if cur.TeamRole == TeamRoleReferenceVet {
		// Reference vet role/permissions are immutable; site + calendar visibility may still be set.
		if teamRole != nil || permissions != nil {
			return TeamMember{}, ErrForbidden
		}
		if defaultSiteID == nil && includeInCalendar == nil {
			return TeamMember{}, ErrForbidden
		}
	}
	newRole := cur.TeamRole
	if teamRole != nil && *teamRole != "" && *teamRole != TeamRoleReferenceVet {
		newRole = *teamRole
	}
	var permJSON any
	if permissions != nil {
		b, _ := json.Marshal(permissions)
		permJSON = b
	} else if len(raw) > 0 {
		permJSON = raw
	}
	profileID := cur.ProfileID
	if newRole != cur.TeamRole {
		kr := teamRoleToKernel(newRole)
		newProfileID, perr := s.ensureProfile(ctx, cur.UserID, kr, practiceID, "")
		if perr != nil {
			return TeamMember{}, perr
		}
		profileID = newProfileID
		if cur.ProfileID != "" {
			_, _ = s.pool.Exec(ctx, `
				UPDATE identity.users SET role=$2, active_profile_id=$3
				WHERE id=$1 AND active_profile_id=$4`,
				cur.UserID, string(kr), profileID, cur.ProfileID)
		} else {
			_, _ = s.pool.Exec(ctx, `
				UPDATE identity.users SET role=$2, active_profile_id=$3
				WHERE id=$1 AND practice_id=$4 AND role IN ('vet','vet_assistant','secretary')`,
				cur.UserID, string(kr), profileID, practiceID)
		}
	}
	nextDefaultSite := cur.DefaultSiteID
	if defaultSiteID != nil {
		sid := strings.TrimSpace(*defaultSiteID)
		if sid == "" {
			nextDefaultSite = ""
		} else {
			resolved, rerr := s.ResolveSiteID(ctx, practiceID, sid, false)
			if rerr != nil {
				return TeamMember{}, fmt.Errorf("%w: invalid_site", ErrValidation)
			}
			nextDefaultSite = resolved
		}
	}
	nextInclude := cur.IncludeInCalendar
	if includeInCalendar != nil {
		nextInclude = *includeInCalendar
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE practice.team_members SET team_role=$3, permissions=$4, profile_id=$5, default_site_id=$6, include_in_calendar=$7
		WHERE id=$1 AND practice_id=$2`,
		memberID, practiceID, string(newRole), permJSON, nullIfEmpty(profileID), nullUUIDArg(nextDefaultSite), nextInclude)
	if err != nil {
		return TeamMember{}, err
	}
	members, _ := s.ListTeamMembers(ctx, practiceID)
	for _, m := range members {
		if m.ID == memberID {
			return m, nil
		}
	}
	return TeamMember{}, ErrNotFound
}

func (s *Store) RevokeTeamMember(ctx context.Context, practiceID, actorUserID, memberID string) error {
	ok, err := s.IsReferenceVet(ctx, practiceID, actorUserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	var userID string
	err = s.pool.QueryRow(ctx, `
		SELECT user_id::text FROM practice.team_members
		WHERE id=$1 AND practice_id=$2 AND team_role <> 'reference_vet'`, memberID, practiceID,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.team_members SET status='revoked'
		WHERE id=$1 AND practice_id=$2 AND team_role <> 'reference_vet'`, memberID, practiceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	// Drop practice binding so JWT/practice gates fail even before token expiry.
	_, _ = s.pool.Exec(ctx, `
		UPDATE identity.users SET practice_id=NULL
		WHERE id=$1 AND practice_id=$2 AND role IN ('vet','vet_assistant','secretary')`,
		userID, practiceID)
	_, _ = s.pool.Exec(ctx, `
		UPDATE identity.profiles SET practice_id=NULL
		WHERE user_id=$1 AND practice_id=$2 AND role IN ('vet','vet_assistant','secretary')`,
		userID, practiceID)
	// Prefer personal client profile when available.
	var clientProfileID string
	_ = s.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.profiles WHERE user_id=$1 AND role='client' LIMIT 1`, userID,
	).Scan(&clientProfileID)
	if clientProfileID != "" {
		_, _ = s.SwitchProfile(ctx, userID, clientProfileID)
	}
	return nil
}

// EnsureReferenceTeamMembership ensures the practice owner is reference_vet in team_members.
func (s *Store) EnsureReferenceTeamMembership(ctx context.Context, practiceID, vetUserID string) error {
	_, _ = s.pool.Exec(ctx, `
		UPDATE practice.practices SET reference_vet_user_id=$2 WHERE id=$1 AND reference_vet_user_id IS NULL`,
		practiceID, vetUserID)
	_ = s.EnsureUserProfiles(ctx, vetUserID)
	profileID := mustProfileID(ctx, s, vetUserID, kernel.RoleVet)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.team_members (id, practice_id, user_id, profile_id, team_role, status)
		VALUES ($1, $2, $3, $4, 'reference_vet', 'active')
		ON CONFLICT (practice_id, user_id) DO UPDATE SET
			team_role = 'reference_vet', status = 'active', profile_id = EXCLUDED.profile_id`,
		uuid.NewString(), practiceID, vetUserID, nullIfEmpty(profileID))
	return err
}
