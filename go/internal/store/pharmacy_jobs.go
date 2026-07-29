package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// PharmacyJobAudit is one worker attempt trail row.
type PharmacyJobAudit struct {
	ID           string          `json:"id"`
	PracticeID   string          `json:"practiceId"`
	JobType      string          `json:"jobType"`
	EntityID     string          `json:"entityId"`
	Attempt      int             `json:"attempt"`
	Status       string          `json:"status"`
	RequestJSON  json.RawMessage `json:"requestJson,omitempty"`
	ResponseJSON json.RawMessage `json:"responseJson,omitempty"`
	Error        string          `json:"error,omitempty"`
	CreatedAt    string          `json:"createdAt"`
}

// InsertPharmacyJobAudit writes a started/success/failed row.
func (s *Store) InsertPharmacyJobAudit(ctx context.Context, practiceID, jobType, entityID string, attempt int, status string, req, resp json.RawMessage, errMsg string) error {
	if attempt < 1 {
		attempt = 1
	}
	if len(req) == 0 {
		req = json.RawMessage(`{}`)
	}
	id := uuid.NewString()
	var respPtr *string
	if len(resp) > 0 && string(resp) != "null" {
		s := string(resp)
		respPtr = &s
	}
	var errArg any
	if errMsg != "" {
		errArg = errMsg
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pharmacy.job_audit (
			id, practice_id, job_type, entity_id, attempt, status, request_json, response_json, error
		) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb,$9)`,
		id, practiceID, jobType, entityID, attempt, status, string(req), respPtr, errArg,
	)
	return err
}

// NextPharmacyJobAttempt returns max(attempt)+1 for job_type+entity (or 1).
func (s *Store) NextPharmacyJobAttempt(ctx context.Context, jobType, entityID string) (int, error) {
	var n *int
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(attempt) FROM pharmacy.job_audit
		WHERE job_type = $1 AND entity_id = $2`, jobType, entityID).Scan(&n)
	if err != nil {
		return 1, err
	}
	if n == nil {
		return 1, nil
	}
	return *n + 1, nil
}

// ListPharmacyJobAudits returns recent audits for an entity.
func (s *Store) ListPharmacyJobAudits(ctx context.Context, practiceID, jobType, entityID string, limit int) ([]PharmacyJobAudit, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, job_type, entity_id::text, attempt, status,
		       request_json, response_json, COALESCE(error, ''), created_at::text
		FROM pharmacy.job_audit
		WHERE practice_id = $1 AND job_type = $2 AND entity_id = $3
		ORDER BY created_at DESC
		LIMIT $4`, practiceID, jobType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PharmacyJobAudit
	for rows.Next() {
		var a PharmacyJobAudit
		var req, resp []byte
		if err := rows.Scan(&a.ID, &a.PracticeID, &a.JobType, &a.EntityID, &a.Attempt, &a.Status, &req, &resp, &a.Error, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.RequestJSON = req
		a.ResponseJSON = resp
		out = append(out, a)
	}
	return out, rows.Err()
}

// UpdateDAFVamregStatus sets vamreg_status with allowed transitions:
// pending|failed → sent|failed|pending ; sent is terminal for workers (retry handler blocks).
func (s *Store) UpdateDAFVamregStatus(ctx context.Context, practiceID, dafID, status string) error {
	switch status {
	case "sent", "failed", "pending":
		// ok
	default:
		return ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.daf_documents
		SET vamreg_status = $3, updated_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'finalized' AND has_antibiotic = true
		  AND vamreg_status IN ('pending', 'failed')`,
		practiceID, dafID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pharmacy.ErrDAFNotFound
	}
	return nil
}

// BuildVamregDeclareRequest loads finalized antibiotic lines for declaration.
func (s *Store) BuildVamregDeclareRequest(ctx context.Context, practiceID, dafID string) (pharmacy.VamregDeclareRequest, error) {
	doc, err := s.GetDAF(ctx, practiceID, dafID)
	if err != nil {
		return pharmacy.VamregDeclareRequest{}, err
	}
	if doc.Status != "finalized" {
		return pharmacy.VamregDeclareRequest{}, pharmacy.ErrDAFNotFinalized
	}
	if !doc.HasAntibiotic {
		return pharmacy.VamregDeclareRequest{}, pharmacy.ErrDAFNotAntibiotic
	}
	req := pharmacy.VamregDeclareRequest{
		DAFID:      doc.ID,
		DAFNumber:  doc.DisplayNumber,
		PracticeID: practiceID,
	}
	for _, it := range doc.Items {
		if !it.IsAntibiotic {
			continue
		}
		payload := it.VamregPayload
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		req.Lines = append(req.Lines, pharmacy.VamregLine{
			MedicationCNK:  it.MedicationCNK,
			MedicationName: it.MedicationName,
			AMMNumber:      it.AMMNumber,
			Qty:            it.Qty,
			Unit:           it.Unit,
			Payload:        payload,
		})
	}
	if len(req.Lines) == 0 {
		return pharmacy.VamregDeclareRequest{}, pharmacy.ErrDAFVAMRegIncomplete
	}
	return req, nil
}

// GetDAFVamregGate returns status fields needed before enqueue/retry.
func (s *Store) GetDAFVamregGate(ctx context.Context, practiceID, dafID string) (status, vamreg string, hasAB bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT status, vamreg_status, has_antibiotic
		FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2`, practiceID, dafID).Scan(&status, &vamreg, &hasAB)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, pharmacy.ErrDAFNotFound
	}
	return status, vamreg, hasAB, err
}
