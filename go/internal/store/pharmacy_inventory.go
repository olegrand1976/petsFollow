package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// InventorySession is an annual / periodic stock count.
type InventorySession struct {
	ID         string          `json:"id"`
	PracticeID string          `json:"practiceId"`
	DepositID  string          `json:"depositId,omitempty"`
	Status     string          `json:"status"`
	Notes      string          `json:"notes,omitempty"`
	CreatedBy  string          `json:"createdBy,omitempty"`
	ClosedBy   string          `json:"closedBy,omitempty"`
	ClosedAt   string          `json:"closedAt,omitempty"`
	CreatedAt  string          `json:"createdAt,omitempty"`
	Lines      []InventoryLine `json:"lines,omitempty"`
	LineCount  int             `json:"lineCount,omitempty"`
	CountedN   int             `json:"countedCount,omitempty"`
	VarianceN  int             `json:"varianceCount,omitempty"`
}

// InventoryLine is one batch snapshot in a session.
type InventoryLine struct {
	ID             string   `json:"id"`
	BatchID        string   `json:"batchId"`
	MedicationID   string   `json:"medicationId"`
	MedicationCNK  string   `json:"medicationCnk,omitempty"`
	MedicationName string   `json:"medicationName,omitempty"`
	DepositID      string   `json:"depositId,omitempty"`
	LotNumber      string   `json:"lotNumber"`
	ExpiresOn      string   `json:"expiresOn"`
	BatchStatus    string   `json:"batchStatus"`
	SystemQty      float64  `json:"systemQty"`
	CountedQty     *float64 `json:"countedQty,omitempty"`
	Delta          *float64 `json:"delta,omitempty"`
	Adjusted       bool     `json:"adjusted"`
	SortOrder      int      `json:"sortOrder"`
}

// InventoryCountInput is a counted qty for one inventory line (batch close).
type InventoryCountInput struct {
	LineID     string  `json:"lineId"`
	CountedQty float64 `json:"countedQty"`
}

// StartInventorySession snapshots active/quarantine batches into an open session.
func (s *Store) StartInventorySession(ctx context.Context, practiceID, userID, depositID, notes string) (InventorySession, error) {
	depositID = strings.TrimSpace(depositID)
	if depositID != "" {
		if _, err := uuid.Parse(depositID); err != nil {
			return InventorySession{}, ErrValidation
		}
		if err := s.assertDepositInPractice(ctx, practiceID, depositID); err != nil {
			return InventorySession{}, err
		}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return InventorySession{}, err
	}
	defer tx.Rollback(ctx)

	sessionID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.inventory_sessions (id, practice_id, deposit_id, status, notes, created_by)
		VALUES ($1,$2,NULLIF($3,'')::uuid,'open',$4,NULLIF($5,'')::uuid)`,
		sessionID, practiceID, depositID, strings.TrimSpace(notes), userID)
	if err != nil {
		if isUniqueViolation(err) {
			return InventorySession{}, ErrConflict
		}
		return InventorySession{}, err
	}

	q := `
		SELECT b.id::text, b.medication_id::text, COALESCE(b.deposit_id::text,''),
		       b.lot_number, b.expires_on::text, b.status, b.qty_on_hand::float8
		FROM pharmacy.medication_batches b
		WHERE b.practice_id = $1
		  AND b.status IN ('active', 'quarantine')
		  AND ($2 = '' OR b.deposit_id::text = $2)
		ORDER BY b.expires_on, b.lot_number`
	rows, err := tx.Query(ctx, q, practiceID, depositID)
	if err != nil {
		return InventorySession{}, err
	}
	type snap struct {
		batchID, medID, depID, lot, exp, status string
		qty                                     float64
	}
	var snaps []snap
	for rows.Next() {
		var row snap
		if err := rows.Scan(&row.batchID, &row.medID, &row.depID, &row.lot, &row.exp, &row.status, &row.qty); err != nil {
			rows.Close()
			return InventorySession{}, err
		}
		snaps = append(snaps, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return InventorySession{}, err
	}
	for i, sn := range snaps {
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.inventory_lines (
				id, session_id, batch_id, medication_id, deposit_id, lot_number, expires_on,
				batch_status, system_qty, sort_order
			) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7::date,$8,$9,$10)`,
			uuid.NewString(), sessionID, sn.batchID, sn.medID, sn.depID, sn.lot, sn.exp, sn.status, sn.qty, i); err != nil {
			return InventorySession{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return InventorySession{}, err
	}
	return s.GetInventorySession(ctx, practiceID, sessionID)
}

// GetInventorySession loads session + lines.
func (s *Store) GetInventorySession(ctx context.Context, practiceID, sessionID string) (InventorySession, error) {
	var sess InventorySession
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, COALESCE(deposit_id::text,''), status, notes,
		       COALESCE(created_by::text,''), COALESCE(closed_by::text,''),
		       COALESCE(closed_at::text,''), created_at::text
		FROM pharmacy.inventory_sessions
		WHERE practice_id = $1 AND id = $2`, practiceID, sessionID).
		Scan(&sess.ID, &sess.PracticeID, &sess.DepositID, &sess.Status, &sess.Notes,
			&sess.CreatedBy, &sess.ClosedBy, &sess.ClosedAt, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return InventorySession{}, ErrNotFound
	}
	if err != nil {
		return InventorySession{}, err
	}
	lines, err := s.listInventoryLines(ctx, sessionID)
	if err != nil {
		return InventorySession{}, err
	}
	sess.Lines = lines
	sess.LineCount = len(lines)
	for _, ln := range lines {
		if ln.CountedQty != nil {
			sess.CountedN++
			if ln.Delta != nil && *ln.Delta != 0 {
				sess.VarianceN++
			}
		}
	}
	return sess, nil
}

func (s *Store) listInventoryLines(ctx context.Context, sessionID string) ([]InventoryLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT l.id::text, l.batch_id::text, l.medication_id::text, m.cnk, m.name,
		       COALESCE(l.deposit_id::text,''), l.lot_number, l.expires_on::text, l.batch_status,
		       l.system_qty::float8, l.counted_qty, l.adjusted, l.sort_order
		FROM pharmacy.inventory_lines l
		JOIN pharmacy.ref_medications m ON m.id = l.medication_id
		WHERE l.session_id = $1
		ORDER BY l.sort_order`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventoryLine{}
	for rows.Next() {
		var ln InventoryLine
		var counted *float64
		if err := rows.Scan(&ln.ID, &ln.BatchID, &ln.MedicationID, &ln.MedicationCNK, &ln.MedicationName,
			&ln.DepositID, &ln.LotNumber, &ln.ExpiresOn, &ln.BatchStatus,
			&ln.SystemQty, &counted, &ln.Adjusted, &ln.SortOrder); err != nil {
			return nil, err
		}
		ln.CountedQty = counted
		if counted != nil {
			d := *counted - ln.SystemQty
			ln.Delta = &d
		}
		out = append(out, ln)
	}
	return out, rows.Err()
}

// ListInventorySessions lists recent sessions (no lines).
func (s *Store) ListInventorySessions(ctx context.Context, practiceID string, limit int) ([]InventorySession, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.practice_id::text, COALESCE(s.deposit_id::text,''), s.status, s.notes,
		       COALESCE(s.created_by::text,''), COALESCE(s.closed_at::text,''), s.created_at::text,
		       (SELECT COUNT(*)::int FROM pharmacy.inventory_lines l WHERE l.session_id = s.id),
		       (SELECT COUNT(*)::int FROM pharmacy.inventory_lines l WHERE l.session_id = s.id AND l.counted_qty IS NOT NULL)
		FROM pharmacy.inventory_sessions s
		WHERE s.practice_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2`, practiceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InventorySession{}
	for rows.Next() {
		var sess InventorySession
		if err := rows.Scan(&sess.ID, &sess.PracticeID, &sess.DepositID, &sess.Status, &sess.Notes,
			&sess.CreatedBy, &sess.ClosedAt, &sess.CreatedAt, &sess.LineCount, &sess.CountedN); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// SetInventoryCountedQty sets counted qty on an open session line.
func (s *Store) SetInventoryCountedQty(ctx context.Context, practiceID, sessionID, lineID string, counted float64) (InventoryLine, error) {
	if counted < 0 {
		return InventoryLine{}, ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.inventory_lines l
		SET counted_qty = $4
		FROM pharmacy.inventory_sessions s
		WHERE l.id = $3 AND l.session_id = s.id
		  AND s.id = $2 AND s.practice_id = $1 AND s.status = 'open'`,
		practiceID, sessionID, lineID, counted)
	if err != nil {
		return InventoryLine{}, err
	}
	if tag.RowsAffected() == 0 {
		return InventoryLine{}, ErrNotFound
	}
	lines, err := s.listInventoryLines(ctx, sessionID)
	if err != nil {
		return InventoryLine{}, err
	}
	for _, ln := range lines {
		if ln.ID == lineID {
			return ln, nil
		}
	}
	return InventoryLine{}, ErrNotFound
}

// CloseInventorySession applies counted vs current on-hand deltas and marks session closed.
// Optional counts are written in the same transaction before validation.
// Every line must have a counted_qty (ErrInventoryIncomplete otherwise).
// Delta = counted - qty_on_hand (locked), not counted - snapshot system_qty.
func (s *Store) CloseInventorySession(ctx context.Context, practiceID, sessionID, userID string, counts []InventoryCountInput, settings PharmacySettings, today time.Time) (InventorySession, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return InventorySession{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.inventory_sessions
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, sessionID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return InventorySession{}, ErrNotFound
	}
	if err != nil {
		return InventorySession{}, err
	}
	if status != "open" {
		return InventorySession{}, ErrConflict
	}

	for _, c := range counts {
		lineID := strings.TrimSpace(c.LineID)
		if lineID == "" || c.CountedQty < 0 {
			return InventorySession{}, ErrValidation
		}
		tag, err := tx.Exec(ctx, `
			UPDATE pharmacy.inventory_lines l
			SET counted_qty = $3
			FROM pharmacy.inventory_sessions s
			WHERE l.id = $2 AND l.session_id = s.id
			  AND s.id = $1 AND s.practice_id = $4 AND s.status = 'open'`,
			sessionID, lineID, c.CountedQty, practiceID)
		if err != nil {
			return InventorySession{}, err
		}
		if tag.RowsAffected() == 0 {
			return InventorySession{}, ErrNotFound
		}
	}

	rows, err := tx.Query(ctx, `
		SELECT id::text, batch_id::text, counted_qty
		FROM pharmacy.inventory_lines
		WHERE session_id = $1
		ORDER BY sort_order`, sessionID)
	if err != nil {
		return InventorySession{}, err
	}
	type lineAdj struct {
		lineID, batchID string
		counted         float64
	}
	var lines []lineAdj
	for rows.Next() {
		var lineID, batchID string
		var counted *float64
		if err := rows.Scan(&lineID, &batchID, &counted); err != nil {
			rows.Close()
			return InventorySession{}, err
		}
		if counted == nil {
			rows.Close()
			return InventorySession{}, ErrInventoryIncomplete
		}
		lines = append(lines, lineAdj{lineID: lineID, batchID: batchID, counted: *counted})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return InventorySession{}, err
	}

	for _, ln := range lines {
		var current float64
		err := tx.QueryRow(ctx, `
			SELECT qty_on_hand::float8 FROM pharmacy.medication_batches
			WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, ln.batchID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return InventorySession{}, pharmacy.ErrBatchNotFound
		}
		if err != nil {
			return InventorySession{}, err
		}
		delta := ln.counted - current
		if delta == 0 {
			continue
		}
		detail := fmt.Sprintf("inventory:%s", sessionID)
		if err := s.adjustBatchQtyTx(ctx, tx, practiceID, ln.batchID, userID, delta, detail, sessionID, settings, today); err != nil {
			return InventorySession{}, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.inventory_lines SET adjusted = TRUE WHERE id = $1`, ln.lineID); err != nil {
			return InventorySession{}, err
		}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE pharmacy.inventory_sessions
		SET status = 'closed', closed_by = NULLIF($3,'')::uuid, closed_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'open'`,
		practiceID, sessionID, userID)
	if err != nil {
		return InventorySession{}, err
	}
	if tag.RowsAffected() == 0 {
		return InventorySession{}, ErrConflict
	}
	if err := tx.Commit(ctx); err != nil {
		return InventorySession{}, err
	}
	return s.GetInventorySession(ctx, practiceID, sessionID)
}

// CancelInventorySession abandons an open session without stock changes.
func (s *Store) CancelInventorySession(ctx context.Context, practiceID, sessionID, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.inventory_sessions
		SET status = 'cancelled', closed_by = NULLIF($3,'')::uuid, closed_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'open'`,
		practiceID, sessionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FormatInventoryCSV exports session lines for the annual register.
func FormatInventoryCSV(sess InventorySession) string {
	var b strings.Builder
	b.WriteString("cnk;name;lot;expires_on;batch_status;system_qty;counted_qty;delta;adjusted\n")
	for _, ln := range sess.Lines {
		counted := ""
		delta := ""
		if ln.CountedQty != nil {
			counted = fmt.Sprintf("%g", *ln.CountedQty)
		}
		if ln.Delta != nil {
			delta = fmt.Sprintf("%g", *ln.Delta)
		}
		b.WriteString(fmt.Sprintf("%s;%s;%s;%s;%s;%g;%s;%s;%t\n",
			ln.MedicationCNK, csvEscape(ln.MedicationName), csvEscape(ln.LotNumber),
			ln.ExpiresOn, ln.BatchStatus, ln.SystemQty, counted, delta, ln.Adjusted))
	}
	return b.String()
}
