package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ClientVetSummary struct {
	PracticeID   string `json:"practiceId"`
	PracticeName string `json:"practiceName"`
	VetUserID    string `json:"vetUserId"`
	VetFullName  string `json:"vetFullName"`
	VetEmail     string `json:"vetEmail"`
	LinkedAt     string `json:"linkedAt,omitempty"`
}

func (s *Store) ListClientVets(ctx context.Context, clientUserID string) ([]ClientVetSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pc.practice_id::text, pr.name, pc.vet_user_id::text, u.full_name, u.email, pc.created_at::text
		FROM practice.practice_clients pc
		JOIN practice.practices pr ON pr.id = pc.practice_id
		JOIN identity.users u ON u.id = pc.vet_user_id
		WHERE pc.client_user_id = $1
		ORDER BY pr.name`, clientUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClientVetSummary
	for rows.Next() {
		var v ClientVetSummary
		if err := rows.Scan(&v.PracticeID, &v.PracticeName, &v.VetUserID, &v.VetFullName, &v.VetEmail, &v.LinkedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) UpsertPracticeClient(ctx context.Context, practiceID, clientUserID, vetUserID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO UPDATE SET vet_user_id = EXCLUDED.vet_user_id`,
		uuid.NewString(), practiceID, clientUserID, vetUserID)
	return err
}

func (s *Store) ClientIsMemberOfPractice(ctx context.Context, clientUserID, practiceID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.practice_clients
			WHERE client_user_id = $1 AND practice_id = $2
		)`, clientUserID, practiceID).Scan(&exists)
	return exists, err
}

// VetInviteResult is returned after a client attempts to link a vet by email.
type VetInviteResult struct {
	Found        bool   `json:"found"`
	Status       string `json:"status"` // pending | not_found
	PracticeName string `json:"practiceName,omitempty"`
	VetFullName  string `json:"vetFullName,omitempty"`
}

// InviteClientToVetByEmail creates a pending link request (no auto-membership).
// Unknown emails return Found=false (client UX needs an explicit not_found).
func (s *Store) InviteClientToVetByEmail(ctx context.Context, clientUserID, vetEmail string) (VetInviteResult, error) {
	email := strings.ToLower(strings.TrimSpace(vetEmail))
	if email == "" {
		return VetInviteResult{Found: false, Status: "not_found"}, nil
	}
	var vetID, practiceID, vetName, practiceName string
	err := s.pool.QueryRow(ctx, `
		SELECT u.id::text, u.practice_id::text, u.full_name, pr.name
		FROM identity.users u
		JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE lower(u.email) = $1 AND u.role = 'vet' AND u.practice_id IS NOT NULL`, email).
		Scan(&vetID, &practiceID, &vetName, &practiceName)
	if errors.Is(err, pgx.ErrNoRows) {
		return VetInviteResult{Found: false, Status: "not_found"}, nil
	}
	if err != nil {
		return VetInviteResult{}, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO practice.client_vet_link_requests (id, client_user_id, practice_id, vet_user_id, status)
		VALUES ($1, $2, $3, $4, 'pending')
		ON CONFLICT (client_user_id, practice_id) DO UPDATE SET
			vet_user_id = EXCLUDED.vet_user_id,
			status = 'pending',
			updated_at = NOW()`,
		uuid.NewString(), clientUserID, practiceID, vetID)
	if err != nil {
		return VetInviteResult{}, err
	}
	return VetInviteResult{
		Found:        true,
		Status:       "pending",
		PracticeName: practiceName,
		VetFullName:  vetName,
	}, nil
}

type VetLinkRequest struct {
	ID           string    `json:"id"`
	ClientUserID string    `json:"clientUserId"`
	ClientName   string    `json:"clientName"`
	ClientEmail  string    `json:"clientEmail"`
	PracticeID   string    `json:"practiceId"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (s *Store) ListPendingVetLinkRequests(ctx context.Context, vetUserID string) ([]VetLinkRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id::text, r.client_user_id::text, u.full_name, u.email, r.practice_id::text, r.status, r.created_at
		FROM practice.client_vet_link_requests r
		JOIN identity.users u ON u.id = r.client_user_id
		WHERE r.vet_user_id = $1 AND r.status = 'pending'
		ORDER BY r.created_at DESC`, vetUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VetLinkRequest
	for rows.Next() {
		var v VetLinkRequest
		if err := rows.Scan(&v.ID, &v.ClientUserID, &v.ClientName, &v.ClientEmail, &v.PracticeID, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) AcceptVetLinkRequest(ctx context.Context, requestID, vetUserID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var clientID, practiceID, reqVetID, status string
	err = tx.QueryRow(ctx, `
		SELECT client_user_id::text, practice_id::text, vet_user_id::text, status
		FROM practice.client_vet_link_requests WHERE id = $1 FOR UPDATE`, requestID,
	).Scan(&clientID, &practiceID, &reqVetID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if reqVetID != vetUserID || status != "pending" {
		return ErrForbidden
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (practice_id, client_user_id) DO UPDATE SET vet_user_id = EXCLUDED.vet_user_id`,
		uuid.NewString(), practiceID, clientID, vetUserID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE practice.client_vet_link_requests SET status = 'accepted', updated_at = NOW() WHERE id = $1`,
		requestID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) RejectVetLinkRequest(ctx context.Context, requestID, vetUserID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE practice.client_vet_link_requests SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND vet_user_id = $2 AND status = 'pending'`, requestID, vetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetPetPrimaryPractice(ctx context.Context, petID, ownerUserID, practiceID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var ownerID, currentPractice string
	err = tx.QueryRow(ctx, `
		SELECT owner_user_id::text, practice_id::text FROM pets.pets WHERE id = $1 FOR UPDATE`, petID,
	).Scan(&ownerID, &currentPractice)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if ownerID != ownerUserID {
		return ErrForbidden
	}
	var member bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.practice_clients
			WHERE client_user_id = $1 AND practice_id = $2
		)`, ownerUserID, practiceID).Scan(&member); err != nil {
		return err
	}
	if !member {
		return ErrForbidden
	}
	if currentPractice != practiceID {
		var status string
		err := tx.QueryRow(ctx, `
			SELECT status FROM billing.pet_entitlements WHERE pet_id = $1`, petID).Scan(&status)
		if err == nil {
			ent := Entitlement{Status: status}
			if ent.AllowsAccess() {
				return ErrForbidden
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	ct, err := tx.Exec(ctx, `
		UPDATE pets.pets SET practice_id = $2, updated_at = NOW()
		WHERE id = $1 AND owner_user_id = $3`, petID, practiceID, ownerUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}
