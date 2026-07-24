package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type AccessPermission string

const (
	PermRead       AccessPermission = "read"
	PermWriteNotes AccessPermission = "write_notes"
	PermFull       AccessPermission = "full"
)

func permissionAtLeast(have, need AccessPermission) bool {
	rank := map[AccessPermission]int{PermRead: 1, PermWriteNotes: 2, PermFull: 3}
	return rank[have] >= rank[need]
}

func maxPermission(a, b AccessPermission) AccessPermission {
	rank := map[AccessPermission]int{PermRead: 1, PermWriteNotes: 2, PermFull: 3, "": 0}
	if rank[a] >= rank[b] {
		return a
	}
	return b
}

// EffectivePetPermission returns the best active grant for grantee on pet (pet_access or client_access).
func (s *Store) EffectivePetPermission(ctx context.Context, pet Pet, granteeUserID string) (AccessPermission, error) {
	var best AccessPermission
	var petPerm string
	err := s.pool.QueryRow(ctx, `
		SELECT permission FROM pets.pet_access
		WHERE pet_id=$1 AND grantee_user_id=$2
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY CASE permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END DESC
		LIMIT 1`, pet.ID, granteeUserID).Scan(&petPerm)
	if err == nil {
		best = AccessPermission(petPerm)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	var clientPerm string
	err = s.pool.QueryRow(ctx, `
		SELECT permission FROM practice.client_access
		WHERE client_user_id=$1 AND grantee_user_id=$2
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY CASE permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END DESC
		LIMIT 1`, pet.OwnerUserID, granteeUserID).Scan(&clientPerm)
	if err == nil {
		best = maxPermission(best, AccessPermission(clientPerm))
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	if best == "" {
		return "", nil
	}
	return best, nil
}

type AccessGrant struct {
	ID              string     `json:"id"`
	GranteeUserID   string     `json:"granteeUserId"`
	GranteeName     string     `json:"granteeName,omitempty"`
	GranteeEmail    string     `json:"granteeEmail,omitempty"`
	Permission      string     `json:"permission"`
	GrantedByUserID string     `json:"grantedByUserId"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// CanAccessPet reports whether the user may see the pet at the required permission level.
func (s *Store) CanAccessPet(ctx context.Context, id kernelIdentity, pet Pet, need AccessPermission) (bool, error) {
	switch id.Role {
	case kernel.RoleClient:
		if pet.OwnerUserID == id.UserID {
			return true, nil
		}
		ok, err := s.hasPetAccess(ctx, pet.ID, id.UserID, need)
		if err != nil || ok {
			return ok, err
		}
		return s.hasClientAccess(ctx, pet.OwnerUserID, id.UserID, need)
	case kernel.RoleVet:
		if id.PracticeID != "" && pet.PracticeID == id.PracticeID {
			return true, nil
		}
		ok, err := s.hasPetAccess(ctx, pet.ID, id.UserID, need)
		if err != nil || ok {
			return ok, err
		}
		return s.hasClientAccess(ctx, pet.OwnerUserID, id.UserID, need)
	case kernel.RoleCarePro:
		ok, err := s.hasPetAccess(ctx, pet.ID, id.UserID, need)
		if err != nil || ok {
			return ok, err
		}
		// client_access on the owner also grants pet access at the same permission level
		return s.hasClientAccess(ctx, pet.OwnerUserID, id.UserID, need)
	case kernel.RoleAdmin:
		return true, nil
	default:
		return false, nil
	}
}

func ValidAccessPermission(p string) bool {
	switch AccessPermission(p) {
	case PermRead, PermWriteNotes, PermFull:
		return true
	default:
		return false
	}
}

type kernelIdentity struct {
	UserID     string
	Role       kernel.Role
	PracticeID string
}

func IdentityOf(userID string, role kernel.Role, practiceID string) kernelIdentity {
	return kernelIdentity{UserID: userID, Role: role, PracticeID: practiceID}
}

func (s *Store) hasPetAccess(ctx context.Context, petID, granteeID string, need AccessPermission) (bool, error) {
	var perm string
	err := s.pool.QueryRow(ctx, `
		SELECT permission FROM pets.pet_access
		WHERE pet_id=$1 AND grantee_user_id=$2
			AND (expires_at IS NULL OR expires_at > NOW())`, petID, granteeID).Scan(&perm)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return permissionAtLeast(AccessPermission(perm), need), nil
}

func (s *Store) hasClientAccess(ctx context.Context, clientUserID, granteeID string, need AccessPermission) (bool, error) {
	var perm string
	err := s.pool.QueryRow(ctx, `
		SELECT permission FROM practice.client_access
		WHERE client_user_id=$1 AND grantee_user_id=$2
			AND (expires_at IS NULL OR expires_at > NOW())`, clientUserID, granteeID).Scan(&perm)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return permissionAtLeast(AccessPermission(perm), need), nil
}

func (s *Store) GrantPetAccess(ctx context.Context, petID, granteeUserID, grantedBy, permission string, expiresAt *time.Time) (AccessGrant, error) {
	if permission == "" {
		permission = string(PermWriteNotes)
	}
	id := uuid.NewString()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pets.pet_access (id, pet_id, grantee_user_id, permission, granted_by_user_id, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (pet_id, grantee_user_id) DO UPDATE
			SET permission = EXCLUDED.permission,
				granted_by_user_id = EXCLUDED.granted_by_user_id,
				expires_at = EXCLUDED.expires_at`,
		id, petID, granteeUserID, permission, grantedBy, expiresAt)
	if err != nil {
		return AccessGrant{}, err
	}
	return s.getPetAccessRow(ctx, petID, granteeUserID)
}

func (s *Store) RevokePetAccess(ctx context.Context, petID, granteeUserID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM pets.pet_access WHERE pet_id=$1 AND grantee_user_id=$2`, petID, granteeUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListPetAccess(ctx context.Context, petID string) ([]AccessGrant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.grantee_user_id::text, COALESCE(u.full_name,''), COALESCE(u.email,''),
			a.permission, a.granted_by_user_id::text, a.expires_at, a.created_at
		FROM pets.pet_access a
		JOIN identity.users u ON u.id = a.grantee_user_id
		WHERE a.pet_id=$1
			AND (a.expires_at IS NULL OR a.expires_at > NOW())
		ORDER BY a.created_at DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccessGrants(rows)
}

func (s *Store) getPetAccessRow(ctx context.Context, petID, granteeUserID string) (AccessGrant, error) {
	var g AccessGrant
	err := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.grantee_user_id::text, COALESCE(u.full_name,''), COALESCE(u.email,''),
			a.permission, a.granted_by_user_id::text, a.expires_at, a.created_at
		FROM pets.pet_access a
		JOIN identity.users u ON u.id = a.grantee_user_id
		WHERE a.pet_id=$1 AND a.grantee_user_id=$2`, petID, granteeUserID,
	).Scan(&g.ID, &g.GranteeUserID, &g.GranteeName, &g.GranteeEmail, &g.Permission, &g.GrantedByUserID, &g.ExpiresAt, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccessGrant{}, ErrNotFound
	}
	return g, err
}

func (s *Store) GrantClientAccess(ctx context.Context, clientUserID, granteeUserID, grantedBy, permission string, expiresAt *time.Time) (AccessGrant, error) {
	if permission == "" {
		permission = string(PermWriteNotes)
	}
	id := uuid.NewString()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.client_access (id, client_user_id, grantee_user_id, permission, granted_by_user_id, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (client_user_id, grantee_user_id) DO UPDATE
			SET permission = EXCLUDED.permission,
				granted_by_user_id = EXCLUDED.granted_by_user_id,
				expires_at = EXCLUDED.expires_at`,
		id, clientUserID, granteeUserID, permission, grantedBy, expiresAt)
	if err != nil {
		return AccessGrant{}, err
	}
	return s.getClientAccessRow(ctx, clientUserID, granteeUserID)
}

func (s *Store) RevokeClientAccess(ctx context.Context, clientUserID, granteeUserID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM practice.client_access WHERE client_user_id=$1 AND grantee_user_id=$2`, clientUserID, granteeUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListClientAccess(ctx context.Context, clientUserID string) ([]AccessGrant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.grantee_user_id::text, COALESCE(u.full_name,''), COALESCE(u.email,''),
			a.permission, a.granted_by_user_id::text, a.expires_at, a.created_at
		FROM practice.client_access a
		JOIN identity.users u ON u.id = a.grantee_user_id
		WHERE a.client_user_id=$1
			AND (a.expires_at IS NULL OR a.expires_at > NOW())
		ORDER BY a.created_at DESC`, clientUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccessGrants(rows)
}

func (s *Store) getClientAccessRow(ctx context.Context, clientUserID, granteeUserID string) (AccessGrant, error) {
	var g AccessGrant
	err := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.grantee_user_id::text, COALESCE(u.full_name,''), COALESCE(u.email,''),
			a.permission, a.granted_by_user_id::text, a.expires_at, a.created_at
		FROM practice.client_access a
		JOIN identity.users u ON u.id = a.grantee_user_id
		WHERE a.client_user_id=$1 AND a.grantee_user_id=$2`, clientUserID, granteeUserID,
	).Scan(&g.ID, &g.GranteeUserID, &g.GranteeName, &g.GranteeEmail, &g.Permission, &g.GrantedByUserID, &g.ExpiresAt, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccessGrant{}, ErrNotFound
	}
	return g, err
}

func scanAccessGrants(rows pgx.Rows) ([]AccessGrant, error) {
	var out []AccessGrant
	for rows.Next() {
		var g AccessGrant
		if err := rows.Scan(&g.ID, &g.GranteeUserID, &g.GranteeName, &g.GranteeEmail, &g.Permission, &g.GrantedByUserID, &g.ExpiresAt, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if out == nil {
		out = []AccessGrant{}
	}
	return out, rows.Err()
}

// LinkExistingClientToVet attaches an existing client account to the vet's practice.
func (s *Store) LinkExistingClientToVet(ctx context.Context, vetUserID, clientUserID string) error {
	var practiceID string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,'') FROM identity.users
		WHERE id=$1 AND role='vet'`, vetUserID).Scan(&practiceID)
	if errors.Is(err, pgx.ErrNoRows) || practiceID == "" {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	client, err := s.GetUserByID(ctx, clientUserID)
	if err != nil {
		return err
	}
	if client.Role != kernel.RoleClient {
		return ErrValidation
	}
	var already bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.practice_clients
			WHERE practice_id=$1 AND client_user_id=$2
		)`, practiceID, clientUserID).Scan(&already); err != nil {
		return err
	}
	if already {
		return ErrConflict
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`,
		uuid.NewString(), practiceID, clientUserID, vetUserID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO messaging.threads (id, practice_id, client_user_id, vet_user_id, pet_id)
		SELECT $1, $2, $3, $4, NULL
		WHERE NOT EXISTS (
			SELECT 1 FROM messaging.threads
			WHERE practice_id=$2 AND client_user_id=$3 AND vet_user_id=$4 AND pet_id IS NULL
		)`, uuid.NewString(), practiceID, clientUserID, vetUserID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// LookupClientConflict returns details for an email that already exists (client preferred).
func (s *Store) LookupClientConflict(ctx context.Context, email, vetUserID string) (map[string]any, error) {
	u, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"exists":      true,
		"userId":      u.ID,
		"displayName": u.FullName,
		"role":        string(u.Role),
		"email":       u.Email,
	}
	if u.Role != kernel.RoleClient {
		out["linkable"] = false
		out["alreadyLinked"] = false
		return out, nil
	}
	out["linkable"] = true
	var practiceID string
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,'') FROM identity.users WHERE id=$1`, vetUserID).Scan(&practiceID)
	already := false
	if practiceID != "" {
		_ = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM practice.practice_clients
				WHERE practice_id=$1 AND client_user_id=$2
			)`, practiceID, u.ID).Scan(&already)
	}
	out["alreadyLinked"] = already
	return out, nil
}

// ListPracticeColleagueVets lists other vets in the same practice (for share picker).
func (s *Store) ListPracticeColleagueVets(ctx context.Context, practiceID, excludeUserID string) ([]map[string]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, full_name, email FROM identity.users
		WHERE practice_id=$1 AND role='vet' AND id<>$2
		ORDER BY full_name`, practiceID, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]string
	for rows.Next() {
		var id, name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			return nil, err
		}
		out = append(out, map[string]string{"userId": id, "fullName": name, "email": email})
	}
	if out == nil {
		out = []map[string]string{}
	}
	return out, rows.Err()
}

// ListCareProAccessiblePets returns pets granted via pet_access or client_access on the owner.
func (s *Store) ListCareProAccessiblePets(ctx context.Context, granteeUserID string) ([]Pet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, COALESCE(p.practice_id::text,''), p.owner_user_id::text, p.name, p.species,
			COALESCE(p.breed,''), p.birth_date, p.weight_kg, COALESCE(p.photo_url,''),
			COALESCE(p.payment_status,''), COALESCE(p.litter_tag,''), p.created_at,
			COALESCE((
				SELECT CASE
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 3 THEN 'full'
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 2 THEN 'write_notes'
					ELSE 'read'
				END
				FROM (
					SELECT pa.permission FROM pets.pet_access pa
					WHERE pa.pet_id=p.id AND pa.grantee_user_id=$1
						AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
					UNION ALL
					SELECT ca.permission FROM practice.client_access ca
					WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
						AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
				) x
			), 'read') AS permission
		FROM pets.pets p
		WHERE EXISTS (
			SELECT 1 FROM pets.pet_access pa
			WHERE pa.pet_id=p.id AND pa.grantee_user_id=$1
				AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
		)
		OR EXISTS (
			SELECT 1 FROM practice.client_access ca
			WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
				AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
		)
		ORDER BY p.name`, granteeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Pet
	for rows.Next() {
		var p Pet
		if err := rows.Scan(
			&p.ID, &p.PracticeID, &p.OwnerUserID, &p.Name, &p.Species, &p.Breed,
			&p.BirthDate, &p.WeightKg, &p.PhotoURL, &p.PaymentStatus, &p.LitterTag, &p.CreatedAt,
			&p.Permission,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []Pet{}
	}
	return out, rows.Err()
}

// ListCareProClients aggregates distinct owners from pet grants + client_access.
func (s *Store) ListCareProClients(ctx context.Context, granteeUserID string) ([]ClientSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.email, u.full_name, COALESCE(u.avatar_url,''),
			(
				SELECT COUNT(*)::int FROM pets.pets p
				WHERE p.owner_user_id=u.id AND (
					EXISTS (
						SELECT 1 FROM practice.client_access ca
						WHERE ca.client_user_id=u.id AND ca.grantee_user_id=$1
							AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
					)
					OR EXISTS (
						SELECT 1 FROM pets.pet_access pa
						WHERE pa.pet_id=p.id AND pa.grantee_user_id=$1
							AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
					)
				)
			) AS pet_count
		FROM identity.users u
		WHERE u.role='client' AND (
			EXISTS (
				SELECT 1 FROM practice.client_access ca
				WHERE ca.client_user_id=u.id AND ca.grantee_user_id=$1
					AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
			)
			OR EXISTS (
				SELECT 1 FROM pets.pet_access pa
				JOIN pets.pets p ON p.id = pa.pet_id
				WHERE p.owner_user_id=u.id AND pa.grantee_user_id=$1
					AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
			)
		)
		ORDER BY u.full_name`, granteeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClientSummary
	for rows.Next() {
		var c ClientSummary
		if err := rows.Scan(&c.UserID, &c.Email, &c.FullName, &c.AvatarURL, &c.PetCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []ClientSummary{}
	}
	return out, rows.Err()
}

// ListCareProVisits lists visits for pets accessible to the care pro (pet_access or client_access).
func (s *Store) ListCareProVisits(ctx context.Context, granteeUserID string) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by,
			COALESCE(v.address_text,''), v.lat, v.lng,
			COALESCE((
				SELECT CASE
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 3 THEN 'full'
					WHEN MAX(CASE x.permission WHEN 'full' THEN 3 WHEN 'write_notes' THEN 2 ELSE 1 END) = 2 THEN 'write_notes'
					ELSE 'read'
				END
				FROM (
					SELECT pa.permission FROM pets.pet_access pa
					WHERE pa.pet_id=v.pet_id AND pa.grantee_user_id=$1
						AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
					UNION ALL
					SELECT ca.permission FROM practice.client_access ca
					WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
						AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
				) x
			), 'read') AS permission
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE EXISTS (
			SELECT 1 FROM pets.pet_access pa
			WHERE pa.pet_id=v.pet_id AND pa.grantee_user_id=$1
				AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
		)
		OR EXISTS (
			SELECT 1 FROM practice.client_access ca
			WHERE ca.client_user_id=p.owner_user_id AND ca.grantee_user_id=$1
				AND (ca.expires_at IS NULL OR ca.expires_at > NOW())
		)
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC
		LIMIT 200`, granteeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Visit
	for rows.Next() {
		var v Visit
		if err := rows.Scan(
			&v.ID, &v.PetID, &v.PracticeID, &v.ScheduledAt, &v.Status, &v.Notes, &v.Source, &v.CreatedAt,
			&v.PetName, &v.ClientName, &v.ClientID,
			&v.DurationMinutes, &v.ProposedScheduledAt, &v.PendingActionBy,
			&v.AddressText, &v.Lat, &v.Lng, &v.Permission,
		); err != nil {
			return nil, err
		}
		if !permissionAtLeast(AccessPermission(v.Permission), PermWriteNotes) {
			v.Notes = ""
		}
		out = append(out, v)
	}
	if out == nil {
		out = []Visit{}
	}
	return out, rows.Err()
}
