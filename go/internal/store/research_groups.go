package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type ResearchGroup struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CreatedBy       string    `json:"createdBy"`
	MemberRole      string    `json:"memberRole,omitempty"`
	MemberCount     int       `json:"memberCount"`
	DataroomEnabled bool      `json:"dataroomEnabled"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ResearchGroupMember struct {
	UserID     string    `json:"userId"`
	Email      string    `json:"email"`
	FullName   string    `json:"fullName"`
	MemberRole string    `json:"memberRole"`
	JoinedAt   time.Time `json:"joinedAt"`
}

func (s *Store) ListResearchGroupsForUser(ctx context.Context, userID string) ([]ResearchGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT g.id::text, g.name, g.description, COALESCE(g.created_by::text, ''), m.member_role,
		       (SELECT COUNT(*)::int FROM research.group_members gm WHERE gm.group_id = g.id),
		       g.dataroom_enabled, g.created_at, g.updated_at
		FROM research.groups g
		JOIN research.group_members m ON m.group_id = g.id AND m.user_id = $1::uuid
		ORDER BY g.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchGroup
	for rows.Next() {
		var g ResearchGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.MemberRole,
			&g.MemberCount, &g.DataroomEnabled, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if out == nil {
		out = []ResearchGroup{}
	}
	return out, rows.Err()
}

// ListAllResearchGroups is admin-facing (opt-in network + Data room flags).
func (s *Store) ListAllResearchGroups(ctx context.Context) ([]ResearchGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT g.id::text, g.name, g.description, COALESCE(g.created_by::text, ''), '',
		       (SELECT COUNT(*)::int FROM research.group_members gm WHERE gm.group_id = g.id),
		       g.dataroom_enabled, g.created_at, g.updated_at
		FROM research.groups g
		ORDER BY g.updated_at DESC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchGroup
	for rows.Next() {
		var g ResearchGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.MemberRole,
			&g.MemberCount, &g.DataroomEnabled, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if out == nil {
		out = []ResearchGroup{}
	}
	return out, rows.Err()
}

func (s *Store) CreateResearchGroup(ctx context.Context, userID, name, description string) (ResearchGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ResearchGroup{}, fmt.Errorf("name_required")
	}
	if len(name) > 120 {
		return ResearchGroup{}, fmt.Errorf("name_too_long")
	}
	description = strings.TrimSpace(description)
	if len(description) > 2000 {
		return ResearchGroup{}, fmt.Errorf("description_too_long")
	}
	id := uuid.NewString()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ResearchGroup{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	_, err = tx.Exec(ctx, `
		INSERT INTO research.groups (id, name, description, created_by)
		VALUES ($1::uuid, $2, $3, $4::uuid)`, id, name, description, userID)
	if err != nil {
		return ResearchGroup{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO research.group_members (group_id, user_id, member_role)
		VALUES ($1::uuid, $2::uuid, 'owner')`, id, userID)
	if err != nil {
		return ResearchGroup{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ResearchGroup{}, err
	}
	return s.GetResearchGroupForUser(ctx, userID, id)
}

func (s *Store) GetResearchGroupForUser(ctx context.Context, userID, groupID string) (ResearchGroup, error) {
	var g ResearchGroup
	err := s.pool.QueryRow(ctx, `
		SELECT g.id::text, g.name, g.description, COALESCE(g.created_by::text, ''), m.member_role,
		       (SELECT COUNT(*)::int FROM research.group_members gm WHERE gm.group_id = g.id),
		       g.dataroom_enabled, g.created_at, g.updated_at
		FROM research.groups g
		JOIN research.group_members m ON m.group_id = g.id AND m.user_id = $1::uuid
		WHERE g.id = $2::uuid`, userID, groupID).Scan(
		&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.MemberRole,
		&g.MemberCount, &g.DataroomEnabled, &g.CreatedAt, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResearchGroup{}, ErrNotFound
	}
	return g, err
}

func (s *Store) researchGroupMemberRole(ctx context.Context, userID, groupID string) (string, error) {
	var role string
	err := s.pool.QueryRow(ctx, `
		SELECT member_role FROM research.group_members
		WHERE group_id = $1::uuid AND user_id = $2::uuid`, groupID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (s *Store) UpdateResearchGroup(ctx context.Context, userID, groupID, name, description string) (ResearchGroup, error) {
	role, err := s.researchGroupMemberRole(ctx, userID, groupID)
	if err != nil {
		return ResearchGroup{}, err
	}
	if role != "owner" {
		return ResearchGroup{}, fmt.Errorf("forbidden")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ResearchGroup{}, fmt.Errorf("name_required")
	}
	if len(name) > 120 {
		return ResearchGroup{}, fmt.Errorf("name_too_long")
	}
	description = strings.TrimSpace(description)
	if len(description) > 2000 {
		return ResearchGroup{}, fmt.Errorf("description_too_long")
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE research.groups
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1::uuid`, groupID, name, description)
	if err != nil {
		return ResearchGroup{}, err
	}
	return s.GetResearchGroupForUser(ctx, userID, groupID)
}

func (s *Store) DeleteResearchGroup(ctx context.Context, userID, groupID string) error {
	role, err := s.researchGroupMemberRole(ctx, userID, groupID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return fmt.Errorf("forbidden")
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM research.groups WHERE id = $1::uuid`, groupID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListResearchGroupMembers(ctx context.Context, userID, groupID string) ([]ResearchGroupMember, error) {
	if _, err := s.researchGroupMemberRole(ctx, userID, groupID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT m.user_id::text, COALESCE(u.email,''), COALESCE(u.full_name,''), m.member_role, m.joined_at
		FROM research.group_members m
		JOIN identity.users u ON u.id = m.user_id
		WHERE m.group_id = $1::uuid
		ORDER BY m.member_role ASC, u.full_name ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchGroupMember
	for rows.Next() {
		var m ResearchGroupMember
		if err := rows.Scan(&m.UserID, &m.Email, &m.FullName, &m.MemberRole, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []ResearchGroupMember{}
	}
	return out, rows.Err()
}

// TryAddResearchGroupMemberByEmail adds a research user if found.
// Returns (added=false, nil) when email unknown / not research — anti-enumeration for callers.
func (s *Store) TryAddResearchGroupMemberByEmail(ctx context.Context, actorUserID, groupID, email string) (added bool, err error) {
	role, err := s.researchGroupMemberRole(ctx, actorUserID, groupID)
	if err != nil {
		return false, err
	}
	if role != "owner" {
		return false, fmt.Errorf("forbidden")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, fmt.Errorf("email_required")
	}
	var targetID string
	err = s.pool.QueryRow(ctx, `
		SELECT u.id::text
		FROM identity.users u
		WHERE lower(u.email) = $1
		  AND (
		    u.role = $2
		    OR EXISTS (
		      SELECT 1 FROM identity.profiles p
		      WHERE p.user_id = u.id AND p.role = $2
		    )
		  )`, email, string(kernel.RoleResearch)).Scan(&targetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO research.group_members (group_id, user_id, member_role)
		VALUES ($1::uuid, $2::uuid, 'member')
		ON CONFLICT (group_id, user_id) DO NOTHING`, groupID, targetID)
	if err != nil {
		return false, err
	}
	_, _ = s.pool.Exec(ctx, `UPDATE research.groups SET updated_at = NOW() WHERE id = $1::uuid`, groupID)
	return tag.RowsAffected() > 0, nil
}

func (s *Store) RemoveResearchGroupMember(ctx context.Context, actorUserID, groupID, targetUserID string) error {
	role, err := s.researchGroupMemberRole(ctx, actorUserID, groupID)
	if err != nil {
		return err
	}
	if role != "owner" && actorUserID != targetUserID {
		return fmt.Errorf("forbidden")
	}
	var targetRole string
	err = s.pool.QueryRow(ctx, `
		SELECT member_role FROM research.group_members
		WHERE group_id = $1::uuid AND user_id = $2::uuid`, groupID, targetUserID).Scan(&targetRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		var owners int
		if err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM research.group_members
			WHERE group_id = $1::uuid AND member_role = 'owner'`, groupID).Scan(&owners); err != nil {
			return err
		}
		if owners <= 1 {
			return fmt.Errorf("last_owner")
		}
	}
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM research.group_members
		WHERE group_id = $1::uuid AND user_id = $2::uuid`, groupID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	_, _ = s.pool.Exec(ctx, `UPDATE research.groups SET updated_at = NOW() WHERE id = $1::uuid`, groupID)
	return nil
}

// UserHasResearchDataroomAccess is true when the user is member of a group with dataroom_enabled.
func (s *Store) UserHasResearchDataroomAccess(ctx context.Context, userID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM research.group_members m
			JOIN research.groups g ON g.id = m.group_id
			WHERE m.user_id = $1::uuid AND g.dataroom_enabled = true
		)`, userID).Scan(&ok)
	return ok, err
}

// SetResearchGroupDataroomEnabled is admin-only (enforced in handler).
func (s *Store) SetResearchGroupDataroomEnabled(ctx context.Context, groupID string, enabled bool) (ResearchGroup, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE research.groups
		SET dataroom_enabled = $2, updated_at = NOW()
		WHERE id = $1::uuid`, groupID, enabled)
	if err != nil {
		return ResearchGroup{}, err
	}
	if tag.RowsAffected() == 0 {
		return ResearchGroup{}, ErrNotFound
	}
	var g ResearchGroup
	err = s.pool.QueryRow(ctx, `
		SELECT g.id::text, g.name, g.description, COALESCE(g.created_by::text, ''), '',
		       (SELECT COUNT(*)::int FROM research.group_members gm WHERE gm.group_id = g.id),
		       g.dataroom_enabled, g.created_at, g.updated_at
		FROM research.groups g WHERE g.id = $1::uuid`, groupID).Scan(
		&g.ID, &g.Name, &g.Description, &g.CreatedBy, &g.MemberRole,
		&g.MemberCount, &g.DataroomEnabled, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}
