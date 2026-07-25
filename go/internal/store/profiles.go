package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

var (
	ErrCannotModifyOwnProfiles = errors.New("cannot_modify_own_profiles")
	ErrProfileExists           = errors.New("profile_exists")
)

type Profile struct {
	ID                     string      `json:"id"`
	UserID                 string      `json:"userId"`
	Role                   kernel.Role `json:"role"`
	PracticeID             string      `json:"practiceId,omitempty"`
	ProfessionalSpecialty  string      `json:"professionalSpecialty,omitempty"`
	CreatedAt              time.Time   `json:"createdAt"`
	Active                 bool        `json:"active"`
}

// EnsureUserProfiles creates the primary profile from users.role and, for pro roles, a personal client profile.
func (s *Store) EnsureUserProfiles(ctx context.Context, userID string) error {
	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	proID, err := s.ensureProfile(ctx, userID, u.Role, u.PracticeID, u.ProfessionalSpecialty)
	if err != nil {
		return err
	}
	if kernel.IsProRole(u.Role) {
		if _, err := s.ensureProfile(ctx, userID, kernel.RoleClient, "", ""); err != nil {
			return err
		}
	}
	var active string
	_ = s.pool.QueryRow(ctx, `SELECT COALESCE(active_profile_id::text,'') FROM identity.users WHERE id=$1`, userID).Scan(&active)
	if active == "" {
		_, err = s.pool.Exec(ctx, `UPDATE identity.users SET active_profile_id=$2 WHERE id=$1`, userID, proID)
		return err
	}
	return nil
}

func (s *Store) ensureProfile(ctx context.Context, userID string, role kernel.Role, practiceID, specialty string) (string, error) {
	var existing string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.profiles WHERE user_id=$1 AND role=$2`, userID, string(role),
	).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	id := uuid.NewString()
	var practice any
	if practiceID != "" {
		practice = practiceID
	}
	var spec any
	if specialty != "" {
		spec = specialty
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO identity.profiles (id, user_id, role, practice_id, professional_specialty)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, role) DO NOTHING`,
		id, userID, string(role), practice, spec)
	if err != nil {
		return "", err
	}
	// Re-read in case of race
	err = s.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.profiles WHERE user_id=$1 AND role=$2`, userID, string(role),
	).Scan(&existing)
	return existing, err
}

func (s *Store) ListProfiles(ctx context.Context, userID string) ([]Profile, error) {
	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = s.EnsureUserProfiles(ctx, userID)

	var activeID string
	_ = s.pool.QueryRow(ctx, `SELECT COALESCE(active_profile_id::text,'') FROM identity.users WHERE id=$1`, userID).Scan(&activeID)

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, user_id::text, role, COALESCE(practice_id::text,''), COALESCE(professional_specialty,''), created_at
		FROM identity.profiles WHERE user_id=$1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.UserID, &p.Role, &p.PracticeID, &p.ProfessionalSpecialty, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Active = p.ID == activeID || (activeID == "" && p.Role == u.Role)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetActiveProfile(ctx context.Context, userID string) (Profile, error) {
	profiles, err := s.ListProfiles(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	for _, p := range profiles {
		if p.Active {
			return p, nil
		}
	}
	if len(profiles) == 0 {
		return Profile{}, ErrNotFound
	}
	return profiles[0], nil
}

// SwitchProfile activates a profile owned by the user and syncs users.role / practice_id / specialty.
func (s *Store) SwitchProfile(ctx context.Context, userID, profileID string) (Profile, error) {
	var p Profile
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, role, COALESCE(practice_id::text,''), COALESCE(professional_specialty,''), created_at
		FROM identity.profiles WHERE id=$1 AND user_id=$2`, profileID, userID,
	).Scan(&p.ID, &p.UserID, &p.Role, &p.PracticeID, &p.ProfessionalSpecialty, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, err
	}
	if kernel.IsPracticeStaff(p.Role) && p.PracticeID != "" {
		ok, aerr := s.HasActivePracticeStaffAccess(ctx, p.PracticeID, userID)
		if aerr != nil {
			return Profile{}, aerr
		}
		if !ok {
			return Profile{}, ErrForbidden
		}
	}
	var practice any
	if p.PracticeID != "" {
		practice = p.PracticeID
	}
	var spec any
	if p.ProfessionalSpecialty != "" {
		spec = p.ProfessionalSpecialty
	} else {
		spec = ""
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE identity.users SET
			active_profile_id = $2,
			role = $3,
			practice_id = $4,
			professional_specialty = NULLIF($5::text, '')
		WHERE id = $1`,
		userID, p.ID, string(p.Role), practice, spec)
	if err != nil {
		return Profile{}, err
	}
	p.Active = true
	return p, nil
}

type AttachProfileInput struct {
	Role      kernel.Role
	Specialty string
	PracticeID string
}

// AttachProfile adds a profile to another user (admin/commercial). Actor cannot attach to self.
func (s *Store) AttachProfile(ctx context.Context, actorUserID, targetUserID string, in AttachProfileInput) (Profile, error) {
	if actorUserID == targetUserID {
		return Profile{}, ErrCannotModifyOwnProfiles
	}
	if !kernel.ValidRole(in.Role) || in.Role == kernel.RoleAdmin {
		return Profile{}, ErrValidation
	}
	if in.Role == kernel.RoleCarePro {
		if !kernel.ValidSpecialty(kernel.ProfessionalSpecialty(in.Specialty)) {
			return Profile{}, ErrValidation
		}
	}
	if kernel.IsPracticeStaff(in.Role) && strings.TrimSpace(in.PracticeID) == "" {
		return Profile{}, ErrValidation
	}

	var exists string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.profiles WHERE user_id=$1 AND role=$2`, targetUserID, string(in.Role),
	).Scan(&exists)
	if err == nil {
		return Profile{}, ErrProfileExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, err
	}

	id, err := s.ensureProfile(ctx, targetUserID, in.Role, in.PracticeID, in.Specialty)
	if err != nil {
		return Profile{}, err
	}

	// If attaching practice staff, optionally add team membership (non-reference).
	if kernel.IsPracticeStaff(in.Role) && in.PracticeID != "" {
		teamRole := "vet"
		switch in.Role {
		case kernel.RoleVetAssistant:
			teamRole = "vet_assistant"
		case kernel.RoleSecretary:
			teamRole = "secretary"
		case kernel.RoleVet:
			teamRole = "vet"
		}
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO practice.team_members (id, practice_id, user_id, profile_id, team_role, status, invited_by_user_id)
			VALUES ($1, $2, $3, $4, $5, 'active', $6)
			ON CONFLICT (practice_id, user_id) DO UPDATE SET
				profile_id = EXCLUDED.profile_id,
				team_role = EXCLUDED.team_role,
				status = 'active'`,
			uuid.NewString(), in.PracticeID, targetUserID, id, teamRole, actorUserID)
	}

	var p Profile
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, role, COALESCE(practice_id::text,''), COALESCE(professional_specialty,''), created_at
		FROM identity.profiles WHERE id=$1`, id,
	).Scan(&p.ID, &p.UserID, &p.Role, &p.PracticeID, &p.ProfessionalSpecialty, &p.CreatedAt)
	return p, err
}
