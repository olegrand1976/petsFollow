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
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

type DAFDocument struct {
	ID                    string    `json:"id"`
	PracticeID            string    `json:"practiceId"`
	DAFYear               int       `json:"dafYear"`
	DAFNumber             *int64    `json:"dafNumber,omitempty"`
	DisplayNumber         string    `json:"displayNumber,omitempty"`
	Status                string    `json:"status"`
	ClientUserID          string    `json:"clientUserId,omitempty"`
	PetID                 string    `json:"petId,omitempty"`
	PrescriberUserID      string    `json:"prescriberUserId"`
	PrescriberName        string    `json:"prescriberName,omitempty"`
	ClientName            string    `json:"clientName,omitempty"`
	PetName               string    `json:"petName,omitempty"`
	PracticeName          string    `json:"practiceName,omitempty"`
	Notes                 string    `json:"notes,omitempty"`
	IssuedAt              string    `json:"issuedAt,omitempty"`
	FinalizedAt           string    `json:"finalizedAt,omitempty"`
	CancelledAt           string    `json:"cancelledAt,omitempty"`
	CancelReason          string    `json:"cancelReason,omitempty"`
	PDFObjectKey          string    `json:"pdfObjectKey,omitempty"`
	PDFSHA256             string    `json:"pdfSha256,omitempty"`
	HasAntibiotic          bool      `json:"hasAntibiotic"`
	VamregStatus          string    `json:"vamregStatus"`
	InvoicesExportStatus  string    `json:"invoicesExportStatus"`
	Items                 []DAFItem `json:"items,omitempty"`
	CreatedAt             string    `json:"createdAt,omitempty"`
}

type DAFItem struct {
	ID             string          `json:"id"`
	DAFID          string          `json:"dafId"`
	MedicationID   string          `json:"medicationId"`
	MedicationCNK  string          `json:"medicationCnk,omitempty"`
	MedicationName string          `json:"medicationName,omitempty"`
	BatchID        string          `json:"batchId,omitempty"`
	LotNumber      string          `json:"lotNumber,omitempty"`
	ExpiresOn      string          `json:"expiresOn,omitempty"`
	DepositID      string          `json:"depositId,omitempty"`
	Qty            float64         `json:"qty"`
	Unit           string          `json:"unit"`
	AMMNumber      string          `json:"ammNumber"`
	IsAntibiotic   bool            `json:"isAntibiotic"`
	VamregPayload  json.RawMessage `json:"vamregPayload,omitempty"`
	SortOrder      int             `json:"sortOrder"`
}

type DAFItemInput struct {
	MedicationID  string          `json:"medicationId"`
	DepositID     string          `json:"depositId"`
	Qty           float64         `json:"qty"`
	Unit          string          `json:"unit"`
	AMMNumber     string          `json:"ammNumber"`
	VamregPayload json.RawMessage `json:"vamregPayload"`
}

type FEFOPreviewLine struct {
	MedicationID   string  `json:"medicationId"`
	MedicationName string  `json:"medicationName,omitempty"`
	BatchID        string  `json:"batchId"`
	LotNumber      string  `json:"lotNumber"`
	ExpiresOn      string  `json:"expiresOn"`
	Qty            float64 `json:"qty"`
}

func (s *Store) GetRefMedication(ctx context.Context, id string) (RefMedication, error) {
	var m RefMedication
	var atc, form, pack string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, cnk, name,
		       COALESCE(atc_code, ''), COALESCE(pharmaceutical_form, ''), COALESCE(pack_size, ''),
		       is_antibiotic, is_active
		FROM pharmacy.ref_medications WHERE id = $1`, id).Scan(
		&m.ID, &m.CNK, &m.Name, &atc, &form, &pack, &m.IsAntibiotic, &m.IsActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefMedication{}, ErrNotFound
	}
	m.ATCCode, m.PharmaceuticalForm, m.PackSize = atc, form, pack
	return m, err
}

func (s *Store) CreateDAFDraft(ctx context.Context, practiceID, prescriberID string, clientUserID, petID, notes string, items []DAFItemInput) (DAFDocument, error) {
	if len(items) == 0 {
		return DAFDocument{}, pharmacy.ErrDAFEmpty
	}
	year := pharmacy.BrusselsToday(time.Now()).Year()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DAFDocument{}, err
	}
	defer tx.Rollback(ctx)

	docID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.daf_documents (
			id, practice_id, daf_year, status, client_user_id, pet_id, prescriber_user_id, notes
		) VALUES ($1,$2,$3,'draft',NULLIF($4,'')::uuid,NULLIF($5,'')::uuid,$6,NULLIF($7,''))`,
		docID, practiceID, year, clientUserID, petID, prescriberID, strings.TrimSpace(notes),
	)
	if err != nil {
		return DAFDocument{}, err
	}
	for i, it := range items {
		if err := s.insertDAFItemTx(ctx, tx, docID, it, i); err != nil {
			return DAFDocument{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, docID)
}

func (s *Store) insertDAFItemTx(ctx context.Context, tx pgx.Tx, dafID string, it DAFItemInput, sort int) error {
	if it.Qty <= 0 || strings.TrimSpace(it.MedicationID) == "" {
		return ErrValidation
	}
	if it.Unit == "" {
		it.Unit = "unit"
	}
	med, err := s.GetRefMedication(ctx, it.MedicationID)
	if err != nil {
		return err
	}
	payload := it.VamregPayload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.daf_items (
			id, daf_id, medication_id, deposit_id, qty, unit, amm_number, is_antibiotic, vamreg_payload, sort_order
		) VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9::jsonb,$10)`,
		uuid.NewString(), dafID, it.MedicationID, it.DepositID, it.Qty, it.Unit,
		strings.TrimSpace(it.AMMNumber), med.IsAntibiotic, string(payload), sort,
	)
	return err
}

func (s *Store) ReplaceDAFDraftItems(ctx context.Context, practiceID, dafID string, clientUserID, petID, notes string, items []DAFItemInput) (DAFDocument, error) {
	if len(items) == 0 {
		return DAFDocument{}, pharmacy.ErrDAFEmpty
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DAFDocument{}, err
	}
	defer tx.Rollback(ctx)
	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, dafID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	if status != "draft" {
		return DAFDocument{}, pharmacy.ErrDAFNotDraft
	}
	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.daf_items WHERE daf_id = $1`, dafID); err != nil {
		return DAFDocument{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.daf_documents SET
			client_user_id = NULLIF($3,'')::uuid,
			pet_id = NULLIF($4,'')::uuid,
			notes = NULLIF($5,''),
			updated_at = now()
		WHERE practice_id = $1 AND id = $2`, practiceID, dafID, clientUserID, petID, strings.TrimSpace(notes),
	); err != nil {
		return DAFDocument{}, err
	}
	for i, it := range items {
		if err := s.insertDAFItemTx(ctx, tx, dafID, it, i); err != nil {
			return DAFDocument{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, dafID)
}

func (s *Store) ListDAF(ctx context.Context, practiceID, status string, year int) ([]DAFDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id::text, d.practice_id::text, d.daf_year, d.daf_number, d.status,
		       COALESCE(d.client_user_id::text,''), COALESCE(d.pet_id::text,''),
		       d.prescriber_user_id::text, COALESCE(u.full_name,''),
		       COALESCE(d.notes,''), COALESCE(d.issued_at::text,''), COALESCE(d.finalized_at::text,''),
		       d.has_antibiotic, d.vamreg_status, d.invoices_export_status, d.created_at::text,
		       COALESCE(d.pdf_object_key,'')
		FROM pharmacy.daf_documents d
		JOIN identity.users u ON u.id = d.prescriber_user_id
		WHERE d.practice_id = $1
		  AND ($2 = '' OR d.status = $2)
		  AND ($3 = 0 OR d.daf_year = $3)
		ORDER BY d.created_at DESC
		LIMIT 100`, practiceID, status, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DAFDocument
	for rows.Next() {
		var d DAFDocument
		var num *int64
		if err := rows.Scan(
			&d.ID, &d.PracticeID, &d.DAFYear, &num, &d.Status,
			&d.ClientUserID, &d.PetID, &d.PrescriberUserID, &d.PrescriberName,
			&d.Notes, &d.IssuedAt, &d.FinalizedAt, &d.HasAntibiotic, &d.VamregStatus,
			&d.InvoicesExportStatus, &d.CreatedAt, &d.PDFObjectKey,
		); err != nil {
			return nil, err
		}
		d.DAFNumber = num
		if num != nil {
			d.DisplayNumber = pharmacy.FormatDAFNumber(d.DAFYear, *num)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) GetDAF(ctx context.Context, practiceID, dafID string) (DAFDocument, error) {
	var d DAFDocument
	var num *int64
	err := s.pool.QueryRow(ctx, `
		SELECT d.id::text, d.practice_id::text, d.daf_year, d.daf_number, d.status,
		       COALESCE(d.client_user_id::text,''), COALESCE(d.pet_id::text,''),
		       d.prescriber_user_id::text, COALESCE(u.full_name,''),
		       COALESCE(c.full_name,''), COALESCE(p.name,''), COALESCE(pr.name,''),
		       COALESCE(d.notes,''), COALESCE(d.issued_at::text,''), COALESCE(d.finalized_at::text,''),
		       COALESCE(d.cancelled_at::text,''), COALESCE(d.cancel_reason,''),
		       COALESCE(d.pdf_object_key,''), COALESCE(d.pdf_sha256,''),
		       d.has_antibiotic, d.vamreg_status, d.invoices_export_status, d.created_at::text
		FROM pharmacy.daf_documents d
		JOIN identity.users u ON u.id = d.prescriber_user_id
		LEFT JOIN identity.users c ON c.id = d.client_user_id
		LEFT JOIN pets.pets p ON p.id = d.pet_id
		JOIN practice.practices pr ON pr.id = d.practice_id
		WHERE d.practice_id = $1 AND d.id = $2`, practiceID, dafID).Scan(
		&d.ID, &d.PracticeID, &d.DAFYear, &num, &d.Status,
		&d.ClientUserID, &d.PetID, &d.PrescriberUserID, &d.PrescriberName,
		&d.ClientName, &d.PetName, &d.PracticeName,
		&d.Notes, &d.IssuedAt, &d.FinalizedAt, &d.CancelledAt, &d.CancelReason,
		&d.PDFObjectKey, &d.PDFSHA256, &d.HasAntibiotic, &d.VamregStatus,
		&d.InvoicesExportStatus, &d.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	d.DAFNumber = num
	if num != nil {
		d.DisplayNumber = pharmacy.FormatDAFNumber(d.DAFYear, *num)
	}
	items, err := s.listDAFItems(ctx, dafID)
	if err != nil {
		return DAFDocument{}, err
	}
	d.Items = items
	return d, nil
}

func (s *Store) listDAFItems(ctx context.Context, dafID string) ([]DAFItem, error) {
	return s.listDAFItemsQuerier(ctx, s.pool, dafID)
}

func (s *Store) listDAFItemsTx(ctx context.Context, tx pgx.Tx, dafID string) ([]DAFItem, error) {
	return s.listDAFItemsQuerier(ctx, tx, dafID)
}

type dafItemsQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (s *Store) listDAFItemsQuerier(ctx context.Context, q dafItemsQuerier, dafID string) ([]DAFItem, error) {
	rows, err := q.Query(ctx, `
		SELECT i.id::text, i.daf_id::text, i.medication_id::text, m.cnk, m.name,
		       COALESCE(i.batch_id::text,''), COALESCE(b.lot_number,''), COALESCE(b.expires_on::text,''),
		       COALESCE(i.deposit_id::text,''), i.qty::float8, i.unit, COALESCE(i.amm_number,''),
		       i.is_antibiotic, i.vamreg_payload, i.sort_order
		FROM pharmacy.daf_items i
		JOIN pharmacy.ref_medications m ON m.id = i.medication_id
		LEFT JOIN pharmacy.medication_batches b ON b.id = i.batch_id
		WHERE i.daf_id = $1
		ORDER BY i.sort_order, i.created_at`, dafID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DAFItem
	for rows.Next() {
		var it DAFItem
		var payload []byte
		if err := rows.Scan(
			&it.ID, &it.DAFID, &it.MedicationID, &it.MedicationCNK, &it.MedicationName,
			&it.BatchID, &it.LotNumber, &it.ExpiresOn, &it.DepositID, &it.Qty, &it.Unit,
			&it.AMMNumber, &it.IsAntibiotic, &payload, &it.SortOrder,
		); err != nil {
			return nil, err
		}
		it.VamregPayload = json.RawMessage(payload)
		out = append(out, it)
	}
	return out, rows.Err()
}

// PreviewFEFOAllocation simulates FEFO without mutating stock.
func (s *Store) PreviewFEFOAllocation(ctx context.Context, practiceID string, items []DAFItemInput, today time.Time) ([]FEFOPreviewLine, error) {
	tod := pharmacy.BrusselsToday(today)
	var out []FEFOPreviewLine
	for _, it := range items {
		if it.Qty <= 0 {
			return nil, ErrValidation
		}
		med, err := s.GetRefMedication(ctx, it.MedicationID)
		if err != nil {
			return nil, err
		}
		rows, err := s.pool.Query(ctx, `
			SELECT id::text, lot_number, expires_on, qty_on_hand::float8
			FROM pharmacy.medication_batches
			WHERE practice_id = $1 AND medication_id = $2
			  AND status = 'active' AND qty_on_hand > 0 AND expires_on >= $3::date
			  AND ($4 = '' OR deposit_id::text = $4)
			ORDER BY expires_on ASC, created_at ASC`,
			practiceID, it.MedicationID, tod.Format("2006-01-02"), it.DepositID)
		if err != nil {
			return nil, err
		}
		remaining := it.Qty
		var total float64
		for rows.Next() {
			var id, lot string
			var exp time.Time
			var qty float64
			if err := rows.Scan(&id, &lot, &exp, &qty); err != nil {
				rows.Close()
				return nil, err
			}
			total += qty
			if remaining <= 0 {
				continue
			}
			take := qty
			if take > remaining {
				take = remaining
			}
			out = append(out, FEFOPreviewLine{
				MedicationID: it.MedicationID, MedicationName: med.Name,
				BatchID: id, LotNumber: lot, ExpiresOn: exp.Format("2006-01-02"), Qty: take,
			})
			remaining -= take
		}
		rows.Close()
		if remaining > 1e-9 {
			if total == 0 {
				return nil, pharmacy.ErrStockUnavailableValidLots
			}
			return nil, pharmacy.ErrStockInsufficient
		}
	}
	return out, nil
}

func (s *Store) FinalizeDAF(ctx context.Context, practiceID, dafID, userID string, today time.Time) (DAFDocument, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DAFDocument{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	var year int
	err = tx.QueryRow(ctx, `
		SELECT status, daf_year FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, dafID).Scan(&status, &year)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	if status != "draft" {
		return DAFDocument{}, pharmacy.ErrDAFNotDraft
	}

	draftItems, err := s.listDAFItems(ctx, dafID)
	if err != nil {
		return DAFDocument{}, err
	}
	if len(draftItems) == 0 {
		return DAFDocument{}, pharmacy.ErrDAFEmpty
	}

	hasAB := false
	for _, it := range draftItems {
		if strings.TrimSpace(it.AMMNumber) == "" {
			return DAFDocument{}, pharmacy.ErrDAFAMMRequired
		}
		if it.IsAntibiotic {
			hasAB = true
			if err := pharmacy.ValidateVamregPayload(it.VamregPayload); err != nil {
				return DAFDocument{}, err
			}
		}
	}

	// Allocate gapless number (table counter + FOR UPDATE — not a Postgres SEQUENCE).
	if _, err := tx.Exec(ctx, `
		INSERT INTO pharmacy.daf_sequences (practice_id, daf_year, next_number)
		VALUES ($1,$2,1)
		ON CONFLICT (practice_id, daf_year) DO NOTHING`, practiceID, year); err != nil {
		return DAFDocument{}, err
	}
	var next int64
	err = tx.QueryRow(ctx, `
		SELECT next_number FROM pharmacy.daf_sequences
		WHERE practice_id = $1 AND daf_year = $2 FOR UPDATE`, practiceID, year).Scan(&next)
	if err != nil {
		return DAFDocument{}, err
	}
	dafNumber := next
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.daf_sequences SET next_number = $3
		WHERE practice_id = $1 AND daf_year = $2`, practiceID, year, next+1); err != nil {
		return DAFDocument{}, err
	}

	// Rebuild items from FEFO allocations
	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.daf_items WHERE daf_id = $1`, dafID); err != nil {
		return DAFDocument{}, err
	}
	sort := 0
	for _, req := range draftItems {
		payload := req.VamregPayload
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		allocs, err := s.allocateFEFOTx(ctx, tx, practiceID, req.MedicationID, req.DepositID, req.Qty, today)
		if err != nil {
			return DAFDocument{}, err
		}
		for _, al := range allocs {
			itemID := uuid.NewString()
			if _, err := tx.Exec(ctx, `
				INSERT INTO pharmacy.daf_items (
					id, daf_id, medication_id, batch_id, deposit_id, qty, unit, amm_number,
					is_antibiotic, vamreg_payload, sort_order
				) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10::jsonb,$11)`,
				itemID, dafID, req.MedicationID, al.BatchID, req.DepositID, al.Qty, req.Unit,
				req.AMMNumber, req.IsAntibiotic, string(payload), sort,
			); err != nil {
				return DAFDocument{}, err
			}
			if err := s.insertDAFStockMovement(ctx, tx, practiceID, al.BatchID, dafID, itemID, userID, "daf", "", -al.Qty); err != nil {
				return DAFDocument{}, err
			}
			sort++
		}
	}

	vamreg := "n/a"
	inv := "n/a"
	if hasAB {
		vamreg = "pending"
	}

	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.daf_documents SET
			status = 'finalized', daf_number = $3, issued_at = now(), finalized_at = now(),
			has_antibiotic = $4, vamreg_status = $5, invoices_export_status = $6, updated_at = now()
		WHERE practice_id = $1 AND id = $2`,
		practiceID, dafID, dafNumber, hasAB, vamreg, inv,
	); err != nil {
		return DAFDocument{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, dafID)
}

func (s *Store) CancelDAF(ctx context.Context, practiceID, dafID, userID, reason string, restock bool) (DAFDocument, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DAFDocument{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, dafID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	if status != "finalized" {
		return DAFDocument{}, pharmacy.ErrDAFNotFinalized
	}
	items, err := s.listDAFItemsTx(ctx, tx, dafID)
	if err != nil {
		return DAFDocument{}, err
	}
	for _, it := range items {
		if it.BatchID == "" {
			continue
		}
		delta := 0.0
		if restock {
			delta = it.Qty
			if _, err := tx.Exec(ctx, `
				UPDATE pharmacy.medication_batches
				SET qty_on_hand = qty_on_hand + $3, updated_at = now()
				WHERE practice_id = $1 AND id = $2`, practiceID, it.BatchID, it.Qty); err != nil {
				return DAFDocument{}, err
			}
		}
		detail := strings.TrimSpace(reason)
		if !restock {
			if detail != "" {
				detail = detail + " "
			}
			detail += "no_restock"
		}
		if err := s.insertDAFStockMovement(ctx, tx, practiceID, it.BatchID, dafID, it.ID, userID, "daf_cancel", detail, delta); err != nil {
			return DAFDocument{}, err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.daf_documents SET
			status = 'cancelled', cancelled_at = now(), cancel_reason = $3, updated_at = now()
		WHERE practice_id = $1 AND id = $2`, practiceID, dafID, strings.TrimSpace(reason),
	); err != nil {
		return DAFDocument{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, dafID)
}

func (s *Store) SetDAFPDFMeta(ctx context.Context, practiceID, dafID, objectKey, sha256 string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.daf_documents
		SET pdf_object_key = $3, pdf_sha256 = $4, updated_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'finalized'
		  AND (pdf_object_key IS NULL OR pdf_object_key = '')`,
		practiceID, dafID, objectKey, sha256)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Already has PDF or wrong status
		var existing string
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(pdf_object_key,'') FROM pharmacy.daf_documents
			WHERE practice_id = $1 AND id = $2`, practiceID, dafID).Scan(&existing)
		if existing != "" {
			return pharmacy.ErrDAFAlreadyHasPDF
		}
		return fmt.Errorf("pdf meta update failed")
	}
	return nil
}
