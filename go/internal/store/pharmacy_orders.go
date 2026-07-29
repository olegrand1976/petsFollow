package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// Supplier is a practice wholesaler / pharmacy supplier contact.
type Supplier struct {
	ID         string `json:"id"`
	PracticeID string `json:"practiceId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone,omitempty"`
	Notes      string `json:"notes,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// PurchaseOrder is a draft/sent reorder request.
type PurchaseOrder struct {
	ID           string              `json:"id"`
	PracticeID   string              `json:"practiceId"`
	SupplierID   string              `json:"supplierId,omitempty"`
	SupplierName string              `json:"supplierName,omitempty"`
	Status       string              `json:"status"`
	ToEmail      string              `json:"toEmail"`
	Notes        string              `json:"notes,omitempty"`
	CreatedBy    string              `json:"createdBy,omitempty"`
	SentAt       string              `json:"sentAt,omitempty"`
	CreatedAt    string              `json:"createdAt,omitempty"`
	Items        []PurchaseOrderItem `json:"items,omitempty"`
}

// PurchaseOrderItem is one line on a purchase order.
type PurchaseOrderItem struct {
	ID             string  `json:"id"`
	MedicationID   string  `json:"medicationId"`
	MedicationCNK  string  `json:"medicationCnk,omitempty"`
	MedicationName string  `json:"medicationName,omitempty"`
	Qty            float64 `json:"qty"`
	Unit           string  `json:"unit"`
	SortOrder      int     `json:"sortOrder"`
}

// PurchaseOrderItemInput creates a PO line.
type PurchaseOrderItemInput struct {
	MedicationID string  `json:"medicationId"`
	Qty          float64 `json:"qty"`
	Unit         string  `json:"unit"`
}

// DeliveryNote is a received supplier delivery note (BL).
type DeliveryNote struct {
	ID           string             `json:"id"`
	PracticeID   string             `json:"practiceId"`
	SupplierID   string             `json:"supplierId,omitempty"`
	SupplierName string             `json:"supplierName"`
	NoteNumber   string             `json:"noteNumber"`
	Notes        string             `json:"notes,omitempty"`
	CreatedBy    string             `json:"createdBy,omitempty"`
	NotifiedAt   string             `json:"notifiedAt,omitempty"`
	CreatedAt    string             `json:"createdAt,omitempty"`
	Items        []DeliveryNoteItem `json:"items,omitempty"`
	SoftWarnings int                `json:"softWarnings,omitempty"`
}

// DeliveryNoteItem is one received line.
type DeliveryNoteItem struct {
	ID             string  `json:"id"`
	MedicationID   string  `json:"medicationId"`
	MedicationCNK  string  `json:"medicationCnk,omitempty"`
	MedicationName string  `json:"medicationName,omitempty"`
	BatchID        string  `json:"batchId,omitempty"`
	DepositID      string  `json:"depositId,omitempty"`
	LotNumber      string  `json:"lotNumber"`
	ExpiresOn      string  `json:"expiresOn"`
	Qty            float64 `json:"qty"`
	Unit           string  `json:"unit"`
}

// DeliveryNoteItemInput is a BL line to receive into stock.
type DeliveryNoteItemInput struct {
	MedicationID string  `json:"medicationId"`
	DepositID    string  `json:"depositId"`
	LotNumber    string  `json:"lotNumber"`
	ExpiresOn    string  `json:"expiresOn"`
	Qty          float64 `json:"qty"`
	Unit         string  `json:"unit"`
}

// UpsertSupplier creates or updates a supplier by id (empty id = create).
func (s *Store) UpsertSupplier(ctx context.Context, practiceID, id, name, email, phone, notes string) (Supplier, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" {
		return Supplier{}, ErrValidation
	}
	if id == "" {
		id = uuid.NewString()
		_, err := s.pool.Exec(ctx, `
			INSERT INTO pharmacy.suppliers (id, practice_id, name, email, phone, notes)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			id, practiceID, name, email, strings.TrimSpace(phone), strings.TrimSpace(notes))
		if err != nil {
			return Supplier{}, err
		}
	} else {
		tag, err := s.pool.Exec(ctx, `
			UPDATE pharmacy.suppliers
			SET name = $3, email = $4, phone = $5, notes = $6, updated_at = now()
			WHERE practice_id = $1 AND id = $2`,
			practiceID, id, name, email, strings.TrimSpace(phone), strings.TrimSpace(notes))
		if err != nil {
			return Supplier{}, err
		}
		if tag.RowsAffected() == 0 {
			return Supplier{}, ErrNotFound
		}
	}
	return s.GetSupplier(ctx, practiceID, id)
}

// GetSupplier loads one supplier.
func (s *Store) GetSupplier(ctx context.Context, practiceID, id string) (Supplier, error) {
	var out Supplier
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, name, email, phone, notes, created_at::text
		FROM pharmacy.suppliers WHERE practice_id = $1 AND id = $2`, practiceID, id).
		Scan(&out.ID, &out.PracticeID, &out.Name, &out.Email, &out.Phone, &out.Notes, &out.CreatedAt)
	if err == pgx.ErrNoRows {
		return Supplier{}, ErrNotFound
	}
	return out, err
}

// ListSuppliers lists practice suppliers.
func (s *Store) ListSuppliers(ctx context.Context, practiceID string) ([]Supplier, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, name, email, phone, notes, created_at::text
		FROM pharmacy.suppliers WHERE practice_id = $1 ORDER BY name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Supplier{}
	for rows.Next() {
		var su Supplier
		if err := rows.Scan(&su.ID, &su.PracticeID, &su.Name, &su.Email, &su.Phone, &su.Notes, &su.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, su)
	}
	return out, rows.Err()
}

// CreatePurchaseOrderFromAlerts builds a draft PO from current reorder alerts (qty = deficit).
func (s *Store) CreatePurchaseOrderFromAlerts(ctx context.Context, practiceID, userID, supplierID, notes string) (PurchaseOrder, error) {
	alerts, err := s.ListReorderAlerts(ctx, practiceID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	if len(alerts) == 0 {
		return PurchaseOrder{}, ErrValidation
	}
	items := make([]PurchaseOrderItemInput, 0, len(alerts))
	for _, a := range alerts {
		need := a.MinQty - a.OnHandQty
		if need <= 0 {
			need = a.MinQty
		}
		items = append(items, PurchaseOrderItemInput{
			MedicationID: a.MedicationID,
			Qty:          need,
			Unit:         "unit",
		})
	}
	return s.CreatePurchaseOrder(ctx, practiceID, userID, supplierID, notes, items)
}

// CreatePurchaseOrder inserts a draft order with lines.
func (s *Store) CreatePurchaseOrder(ctx context.Context, practiceID, userID, supplierID, notes string, items []PurchaseOrderItemInput) (PurchaseOrder, error) {
	if len(items) == 0 {
		return PurchaseOrder{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PurchaseOrder{}, err
	}
	defer tx.Rollback(ctx)

	orderID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.purchase_orders (id, practice_id, supplier_id, status, notes, created_by)
		VALUES ($1,$2,NULLIF($3,'')::uuid,'draft',$4,NULLIF($5,'')::uuid)`,
		orderID, practiceID, strings.TrimSpace(supplierID), strings.TrimSpace(notes), userID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	for i, it := range items {
		medID := strings.TrimSpace(it.MedicationID)
		if medID == "" || it.Qty <= 0 {
			return PurchaseOrder{}, ErrValidation
		}
		if err := s.requireActiveMedicationTx(ctx, tx, medID); err != nil {
			return PurchaseOrder{}, err
		}
		unit := it.Unit
		if unit == "" {
			unit = "unit"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.purchase_order_items (id, order_id, medication_id, qty, unit, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			uuid.NewString(), orderID, medID, it.Qty, unit, i); err != nil {
			return PurchaseOrder{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return PurchaseOrder{}, err
	}
	return s.GetPurchaseOrder(ctx, practiceID, orderID)
}

// GetPurchaseOrder loads order + items.
func (s *Store) GetPurchaseOrder(ctx context.Context, practiceID, orderID string) (PurchaseOrder, error) {
	var o PurchaseOrder
	err := s.pool.QueryRow(ctx, `
		SELECT o.id::text, o.practice_id::text, COALESCE(o.supplier_id::text,''), COALESCE(su.name,''),
		       o.status, o.to_email, o.notes, COALESCE(o.created_by::text,''),
		       COALESCE(o.sent_at::text,''), o.created_at::text
		FROM pharmacy.purchase_orders o
		LEFT JOIN pharmacy.suppliers su ON su.id = o.supplier_id
		WHERE o.practice_id = $1 AND o.id = $2`, practiceID, orderID).
		Scan(&o.ID, &o.PracticeID, &o.SupplierID, &o.SupplierName, &o.Status, &o.ToEmail, &o.Notes, &o.CreatedBy, &o.SentAt, &o.CreatedAt)
	if err == pgx.ErrNoRows {
		return PurchaseOrder{}, ErrNotFound
	}
	if err != nil {
		return PurchaseOrder{}, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT i.id::text, i.medication_id::text, m.cnk, m.name, i.qty::float8, i.unit, i.sort_order
		FROM pharmacy.purchase_order_items i
		JOIN pharmacy.ref_medications m ON m.id = i.medication_id
		WHERE i.order_id = $1 ORDER BY i.sort_order`, orderID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	defer rows.Close()
	o.Items = []PurchaseOrderItem{}
	for rows.Next() {
		var it PurchaseOrderItem
		if err := rows.Scan(&it.ID, &it.MedicationID, &it.MedicationCNK, &it.MedicationName, &it.Qty, &it.Unit, &it.SortOrder); err != nil {
			return PurchaseOrder{}, err
		}
		o.Items = append(o.Items, it)
	}
	return o, rows.Err()
}

// ListPurchaseOrders lists recent orders (no items).
func (s *Store) ListPurchaseOrders(ctx context.Context, practiceID string, limit int) ([]PurchaseOrder, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT o.id::text, o.practice_id::text, COALESCE(o.supplier_id::text,''), COALESCE(su.name,''),
		       o.status, o.to_email, o.notes, COALESCE(o.sent_at::text,''), o.created_at::text
		FROM pharmacy.purchase_orders o
		LEFT JOIN pharmacy.suppliers su ON su.id = o.supplier_id
		WHERE o.practice_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2`, practiceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PurchaseOrder{}
	for rows.Next() {
		var o PurchaseOrder
		if err := rows.Scan(&o.ID, &o.PracticeID, &o.SupplierID, &o.SupplierName, &o.Status, &o.ToEmail, &o.Notes, &o.SentAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// MarkPurchaseOrderSent sets status sent + to_email + sent_at.
func (s *Store) MarkPurchaseOrderSent(ctx context.Context, practiceID, orderID, toEmail, supplierID string) (PurchaseOrder, error) {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" || !strings.Contains(toEmail, "@") {
		return PurchaseOrder{}, ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.purchase_orders
		SET status = 'sent', to_email = $3, supplier_id = COALESCE(NULLIF($4,'')::uuid, supplier_id),
		    sent_at = now(), updated_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'draft'`,
		practiceID, orderID, toEmail, strings.TrimSpace(supplierID))
	if err != nil {
		return PurchaseOrder{}, err
	}
	if tag.RowsAffected() == 0 {
		return PurchaseOrder{}, ErrNotFound
	}
	return s.GetPurchaseOrder(ctx, practiceID, orderID)
}

// SendPurchaseOrderLocked claims the draft (short FOR UPDATE), releases the TX,
// runs sendFn (SMTP) outside any row lock, then marks sent. Concurrent sends on
// the same order use try-advisory-lock (busy → ErrConflict, no pool wait).
func (s *Store) SendPurchaseOrderLocked(ctx context.Context, practiceID, orderID, toEmail, supplierID string, sendFn func(PurchaseOrder) error) (PurchaseOrder, error) {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" || !strings.Contains(toEmail, "@") {
		return PurchaseOrder{}, ErrValidation
	}

	var lockKey int64
	if err := s.pool.QueryRow(ctx, `SELECT hashtext($1)::bigint`, "pharmacy-po-send:"+orderID).Scan(&lockKey); err != nil {
		return PurchaseOrder{}, err
	}

	var out PurchaseOrder
	err := s.TryWithAdvisoryLock(ctx, lockKey, func(ctx context.Context) error {
		order, err := s.claimPurchaseOrderDraft(ctx, practiceID, orderID)
		if err != nil {
			return err
		}
		if sendFn != nil {
			if err := sendFn(order); err != nil {
				return err
			}
		}
		marked, err := s.MarkPurchaseOrderSent(ctx, practiceID, orderID, toEmail, supplierID)
		if err != nil {
			// SMTP already succeeded: tolerate concurrent mark / race → return current sent row.
			cur, getErr := s.GetPurchaseOrder(ctx, practiceID, orderID)
			if getErr == nil && cur.Status == "sent" {
				out = cur
				return nil
			}
			log.Printf("pharmacy: mark purchase order sent after SMTP failed practice=%s order=%s: %v", practiceID, orderID, err)
			return err
		}
		out = marked
		return nil
	})
	if errors.Is(err, ErrAdvisoryLockBusy) {
		return PurchaseOrder{}, ErrConflict
	}
	if err != nil {
		return PurchaseOrder{}, err
	}
	return out, nil
}

// claimPurchaseOrderDraft FOR UPDATE-checks draft status then commits (releases the row lock).
func (s *Store) claimPurchaseOrderDraft(ctx context.Context, practiceID, orderID string) (PurchaseOrder, error) {
	order, err := s.GetPurchaseOrder(ctx, practiceID, orderID)
	if err != nil {
		return PurchaseOrder{}, err
	}
	if order.Status != "draft" {
		return PurchaseOrder{}, ErrNotFound
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PurchaseOrder{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.purchase_orders
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, orderID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return PurchaseOrder{}, ErrNotFound
	}
	if err != nil {
		return PurchaseOrder{}, err
	}
	if status != "draft" {
		return PurchaseOrder{}, ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return PurchaseOrder{}, err
	}
	return order, nil
}

// FormatPurchaseOrderCSV builds a simple CSV for supplier email.
func FormatPurchaseOrderCSV(o PurchaseOrder) string {
	var b strings.Builder
	b.WriteString("cnk;name;qty;unit\n")
	for _, it := range o.Items {
		b.WriteString(fmt.Sprintf("%s;%s;%g;%s\n", it.MedicationCNK, csvEscape(it.MedicationName), it.Qty, it.Unit))
	}
	return b.String()
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ";\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// ReceiveDeliveryNote creates a BL and receives each line into stock (FEFO receipt rules).
// Empty noteNumber → server-generated MANUAL-<uuid8> (idempotent client retries still unique).
func (s *Store) ReceiveDeliveryNote(ctx context.Context, practiceID, userID string, noteNumber, supplierID, supplierName, notes string, lines []DeliveryNoteItemInput, today time.Time) (DeliveryNote, error) {
	noteNumber = strings.TrimSpace(noteNumber)
	if noteNumber == "" {
		noteNumber = "MANUAL-" + uuid.NewString()[:8]
	}
	if len(lines) == 0 {
		return DeliveryNote{}, ErrValidation
	}
	settings, err := s.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		return DeliveryNote{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DeliveryNote{}, err
	}
	defer tx.Rollback(ctx)

	dnID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.delivery_notes (
			id, practice_id, supplier_id, supplier_name, note_number, notes, created_by
		) VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,NULLIF($7,'')::uuid)`,
		dnID, practiceID, strings.TrimSpace(supplierID), strings.TrimSpace(supplierName),
		noteNumber, strings.TrimSpace(notes), userID)
	if err != nil {
		if isUniqueViolationConstraint(err, "delivery_notes_practice_id_note_number") {
			return DeliveryNote{}, ErrConflict
		}
		return DeliveryNote{}, err
	}

	softWarnings := 0
	for i, line := range lines {
		medID := strings.TrimSpace(line.MedicationID)
		if err := s.requireActiveMedicationTx(ctx, tx, medID); err != nil {
			return DeliveryNote{}, err
		}
		in := ReceiptInput{
			PracticeID:   practiceID,
			DepositID:    line.DepositID,
			MedicationID: medID,
			LotNumber:    line.LotNumber,
			Qty:          line.Qty,
			Unit:         line.Unit,
			CreatedBy:    userID,
		}
		exp, err := time.Parse("2006-01-02", strings.TrimSpace(line.ExpiresOn))
		if err != nil {
			return DeliveryNote{}, ErrValidation
		}
		in.ExpiresOn = exp
		batch, soft, err := s.receiveMedicationBatchTx(ctx, tx, in, settings, today, dnID)
		if err != nil {
			return DeliveryNote{}, err
		}
		if soft {
			softWarnings++
		}
		unit := line.Unit
		if unit == "" {
			unit = "unit"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.delivery_note_items (
				id, delivery_note_id, medication_id, batch_id, deposit_id, lot_number, expires_on, qty, unit, sort_order
			) VALUES ($1,$2,$3,$4::uuid,$5::uuid,$6,$7::date,$8,$9,$10)`,
			uuid.NewString(), dnID, line.MedicationID, batch.ID, batch.DepositID,
			strings.TrimSpace(line.LotNumber), exp.Format("2006-01-02"), line.Qty, unit, i); err != nil {
			return DeliveryNote{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return DeliveryNote{}, err
	}
	dn, err := s.GetDeliveryNote(ctx, practiceID, dnID)
	if err != nil {
		return DeliveryNote{}, err
	}
	dn.SoftWarnings = softWarnings
	return dn, nil
}

// MarkDeliveryNoteNotified stamps notified_at.
func (s *Store) MarkDeliveryNoteNotified(ctx context.Context, practiceID, dnID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.delivery_notes SET notified_at = now()
		WHERE practice_id = $1 AND id = $2`, practiceID, dnID)
	return err
}

// GetDeliveryNote loads BL + items.
func (s *Store) GetDeliveryNote(ctx context.Context, practiceID, dnID string) (DeliveryNote, error) {
	var dn DeliveryNote
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, COALESCE(supplier_id::text,''), supplier_name, note_number, notes,
		       COALESCE(created_by::text,''), COALESCE(notified_at::text,''), created_at::text
		FROM pharmacy.delivery_notes WHERE practice_id = $1 AND id = $2`, practiceID, dnID).
		Scan(&dn.ID, &dn.PracticeID, &dn.SupplierID, &dn.SupplierName, &dn.NoteNumber, &dn.Notes, &dn.CreatedBy, &dn.NotifiedAt, &dn.CreatedAt)
	if err == pgx.ErrNoRows {
		return DeliveryNote{}, ErrNotFound
	}
	if err != nil {
		return DeliveryNote{}, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT i.id::text, i.medication_id::text, m.cnk, m.name, COALESCE(i.batch_id::text,''),
		       COALESCE(i.deposit_id::text,''), i.lot_number, i.expires_on::text, i.qty::float8, i.unit
		FROM pharmacy.delivery_note_items i
		JOIN pharmacy.ref_medications m ON m.id = i.medication_id
		WHERE i.delivery_note_id = $1 ORDER BY i.sort_order`, dnID)
	if err != nil {
		return DeliveryNote{}, err
	}
	defer rows.Close()
	dn.Items = []DeliveryNoteItem{}
	for rows.Next() {
		var it DeliveryNoteItem
		if err := rows.Scan(&it.ID, &it.MedicationID, &it.MedicationCNK, &it.MedicationName, &it.BatchID,
			&it.DepositID, &it.LotNumber, &it.ExpiresOn, &it.Qty, &it.Unit); err != nil {
			return DeliveryNote{}, err
		}
		dn.Items = append(dn.Items, it)
	}
	return dn, rows.Err()
}

func (s *Store) requireActiveMedicationTx(ctx context.Context, tx pgx.Tx, medicationID string) error {
	if strings.TrimSpace(medicationID) == "" {
		return ErrValidation
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pharmacy.ref_medications WHERE id = $1 AND is_active)`, medicationID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolationConstraint(err error, constraintSubstr string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, constraintSubstr)
}

// receiveMedicationBatchTx mirrors ReceiveMedicationBatch inside an open transaction.
func (s *Store) receiveMedicationBatchTx(ctx context.Context, tx pgx.Tx, in ReceiptInput, settings PharmacySettings, today time.Time, deliveryNoteID string) (MedicationBatch, bool, error) {
	if in.Qty <= 0 || strings.TrimSpace(in.LotNumber) == "" || in.MedicationID == "" {
		return MedicationBatch{}, false, ErrValidation
	}
	if in.Unit == "" {
		in.Unit = "unit"
	}
	exp := time.Date(in.ExpiresOn.Year(), in.ExpiresOn.Month(), in.ExpiresOn.Day(), 0, 0, 0, 0, time.UTC)
	tod := pharmacy.BrusselsToday(today)
	softWarn := false
	if exp.Before(tod) {
		if !settings.AllowExpiredReceipt {
			return MedicationBatch{}, false, pharmacy.ErrInvalidExpiryOnReceipt
		}
	} else {
		days := int(exp.Sub(tod).Hours() / 24)
		if days <= settings.ReceiptWarnDays {
			softWarn = true
		}
	}
	if in.DepositID == "" {
		dep, err := s.EnsureDefaultDeposit(ctx, in.PracticeID)
		if err != nil {
			return MedicationBatch{}, false, err
		}
		in.DepositID = dep.ID
	} else if err := s.assertDepositInPracticeTx(ctx, tx, in.PracticeID, in.DepositID); err != nil {
		return MedicationBatch{}, false, err
	}

	batchID := uuid.NewString()
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO pharmacy.medication_batches (
			id, practice_id, deposit_id, medication_id, lot_number, expires_on, qty_on_hand, unit, status
		) VALUES ($1,$2,$3,$4,$5,$6::date,$7,$8,'active')
		ON CONFLICT (practice_id, deposit_id, medication_id, lot_number, expires_on)
		DO UPDATE SET
			qty_on_hand = pharmacy.medication_batches.qty_on_hand + EXCLUDED.qty_on_hand,
			updated_at = now()
		WHERE pharmacy.medication_batches.status <> 'wasted'
		RETURNING id::text`,
		batchID, in.PracticeID, in.DepositID, in.MedicationID, strings.TrimSpace(in.LotNumber),
		exp.Format("2006-01-02"), in.Qty, in.Unit,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return MedicationBatch{}, false, pharmacy.ErrBatchWasted
	}
	if err != nil {
		return MedicationBatch{}, false, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, created_by, delivery_note_id)
		VALUES ($1,$2,$3,$4,'receipt',$5,NULLIF($6,'')::uuid)`,
		uuid.NewString(), in.PracticeID, id, in.Qty, nullIfEmpty(in.CreatedBy), deliveryNoteID,
	); err != nil {
		return MedicationBatch{}, false, err
	}
	b := MedicationBatch{ID: id, DepositID: in.DepositID, MedicationID: in.MedicationID, LotNumber: in.LotNumber, QtyOnHand: in.Qty, Unit: in.Unit}
	return b, softWarn, nil
}
