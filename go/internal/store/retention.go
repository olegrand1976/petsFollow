package store

import (
	"context"
	"time"
)

// TouchLastLogin met à jour la date de dernière connexion (rétention RGPD).
// Best-effort : une erreur ici ne doit jamais bloquer le login.
func (s *Store) TouchLastLogin(ctx context.Context, userID string) {
	_, _ = s.pool.Exec(ctx,
		`UPDATE identity.users SET last_login_at = NOW() WHERE id = $1`, userID)
}

// InactiveAccount — compte candidat à la purge de rétention.
type InactiveAccount struct {
	ID    string
	Role  string
	Email string
}

// ListInactiveAccounts liste les comptes (hors admin, hors comptes déjà anonymisés)
// sans connexion depuis cutoff — purge « 3 ans d'inactivité » des textes légaux.
func (s *Store) ListInactiveAccounts(ctx context.Context, cutoff time.Time, limit int) ([]InactiveAccount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, role, email FROM identity.users
		WHERE role <> 'admin'
		  AND email NOT LIKE '%'||$3
		  AND COALESCE(last_login_at, created_at) < $1
		ORDER BY COALESCE(last_login_at, created_at)
		LIMIT $2`, cutoff, limit, tombstoneEmailSuffix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InactiveAccount
	for rows.Next() {
		var a InactiveAccount
		if err := rows.Scan(&a.ID, &a.Role, &a.Email); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
