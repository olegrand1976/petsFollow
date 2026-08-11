package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

type PharmacySettings struct {
	PracticeID              string    `json:"practiceId"`
	WarnSoonDays            int       `json:"warnSoonDays"`
	WarnReturnDays          int       `json:"warnReturnDays"`
	WarnCriticalDays        int       `json:"warnCriticalDays"`
	ReceiptWarnDays         int       `json:"receiptWarnDays"`
	AllowExpiredReceipt     bool      `json:"allowExpiredReceipt"`
	BlockExpiredOnDAF       bool      `json:"blockExpiredOnDaf"`
	BlockExpiredOnAdjustOut bool      `json:"blockExpiredOnAdjustOut"`
	AutoQuarantineExpired   bool      `json:"autoQuarantineExpired"`
	ExpiryDigestEnabled     bool      `json:"expiryDigestEnabled"`
	ExpiryDigestWeekday     int       `json:"expiryDigestWeekday"`
	NotifyOnAutoQuarantine  bool      `json:"notifyOnAutoQuarantine"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type MedicationDeposit struct {
	ID         string `json:"id"`
	PracticeID string `json:"practiceId"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	IsDefault  bool   `json:"isDefault"`
}

type MedicationBatch struct {
	ID               string  `json:"id"`
	PracticeID       string  `json:"practiceId"`
	DepositID        string  `json:"depositId"`
	DepositCode      string  `json:"depositCode,omitempty"`
	DepositName      string  `json:"depositName,omitempty"`
	MedicationID     string  `json:"medicationId"`
	MedicationCNK    string  `json:"medicationCnk,omitempty"`
	MedicationName   string  `json:"medicationName,omitempty"`
	LotNumber        string  `json:"lotNumber"`
	ExpiresOn        string  `json:"expiresOn"` // YYYY-MM-DD
	QtyOnHand        float64 `json:"qtyOnHand"`
	Unit             string  `json:"unit"`
	Status           string  `json:"status"`
	ExpiryBand       string  `json:"expiryBand,omitempty"`
	IsAntibiotic     bool    `json:"isAntibiotic,omitempty"`
	QuarantineReason string  `json:"quarantineReason,omitempty"`
	WasteReason      string  `json:"wasteReason,omitempty"`
}

type StockMovement struct {
	ID             string  `json:"id"`
	PracticeID     string  `json:"practiceId"`
	BatchID        string  `json:"batchId"`
	Delta          float64 `json:"delta"`
	Reason         string  `json:"reason"`
	ReasonDetail   string  `json:"reasonDetail,omitempty"`
	DafID          string  `json:"dafId,omitempty"`
	DafItemID      string  `json:"dafItemId,omitempty"`
	LotNumber      string  `json:"lotNumber,omitempty"`
	ExpiresOn      string  `json:"expiresOn,omitempty"`
	MedicationCNK  string  `json:"medicationCnk,omitempty"`
	MedicationName string  `json:"medicationName,omitempty"`
	CreatedBy      string  `json:"createdBy,omitempty"`
	CreatedAt      string  `json:"createdAt"`
}

type ListMovementsFilter struct {
	DafID string
	Limit int
}

type ExpirySummary struct {
	OK         int `json:"ok"`
	Soon       int `json:"soon"`
	Return     int `json:"return"`
	Critical   int `json:"critical"`
	Expired    int `json:"expired"`
	Quarantine int `json:"quarantine"`
}

func defaultPharmacySettings(practiceID string) PharmacySettings {
	return PharmacySettings{
		PracticeID:              practiceID,
		WarnSoonDays:            90,
		WarnReturnDays:          60,
		WarnCriticalDays:        30,
		ReceiptWarnDays:         30,
		BlockExpiredOnDAF:       true,
		BlockExpiredOnAdjustOut: true,
		AutoQuarantineExpired:   true,
		ExpiryDigestEnabled:     true,
		ExpiryDigestWeekday:     1,
		NotifyOnAutoQuarantine:  true,
	}
}

func (s *Store) GetPharmacySettings(ctx context.Context, practiceID string) (PharmacySettings, error) {
	var st PharmacySettings
	err := s.pool.QueryRow(ctx, `
		SELECT practice_id::text, warn_soon_days, warn_return_days, warn_critical_days, receipt_warn_days,
		       allow_expired_receipt, block_expired_on_daf, block_expired_on_adjust_out,
		       auto_quarantine_expired, expiry_digest_enabled, expiry_digest_weekday,
		       notify_on_auto_quarantine, updated_at
		FROM pharmacy.practice_settings WHERE practice_id = $1`, practiceID).Scan(
		&st.PracticeID, &st.WarnSoonDays, &st.WarnReturnDays, &st.WarnCriticalDays, &st.ReceiptWarnDays,
		&st.AllowExpiredReceipt, &st.BlockExpiredOnDAF, &st.BlockExpiredOnAdjustOut,
		&st.AutoQuarantineExpired, &st.ExpiryDigestEnabled, &st.ExpiryDigestWeekday,
		&st.NotifyOnAutoQuarantine, &st.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return defaultPharmacySettings(practiceID), nil
	}
	return st, err
}

func (s *Store) UpsertPharmacySettings(ctx context.Context, st PharmacySettings) (PharmacySettings, error) {
	if st.WarnCriticalDays <= 0 || st.WarnReturnDays <= 0 || st.WarnSoonDays <= 0 {
		return PharmacySettings{}, ErrValidation
	}
	if st.WarnCriticalDays > st.WarnReturnDays || st.WarnReturnDays > st.WarnSoonDays {
		return PharmacySettings{}, ErrValidation
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pharmacy.practice_settings (
			practice_id, warn_soon_days, warn_return_days, warn_critical_days, receipt_warn_days,
			allow_expired_receipt, block_expired_on_daf, block_expired_on_adjust_out,
			auto_quarantine_expired, expiry_digest_enabled, expiry_digest_weekday,
			notify_on_auto_quarantine, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, now())
		ON CONFLICT (practice_id) DO UPDATE SET
			warn_soon_days = EXCLUDED.warn_soon_days,
			warn_return_days = EXCLUDED.warn_return_days,
			warn_critical_days = EXCLUDED.warn_critical_days,
			receipt_warn_days = EXCLUDED.receipt_warn_days,
			allow_expired_receipt = EXCLUDED.allow_expired_receipt,
			block_expired_on_daf = EXCLUDED.block_expired_on_daf,
			block_expired_on_adjust_out = EXCLUDED.block_expired_on_adjust_out,
			auto_quarantine_expired = EXCLUDED.auto_quarantine_expired,
			expiry_digest_enabled = EXCLUDED.expiry_digest_enabled,
			expiry_digest_weekday = EXCLUDED.expiry_digest_weekday,
			notify_on_auto_quarantine = EXCLUDED.notify_on_auto_quarantine,
			updated_at = now()
		RETURNING practice_id::text, warn_soon_days, warn_return_days, warn_critical_days, receipt_warn_days,
		          allow_expired_receipt, block_expired_on_daf, block_expired_on_adjust_out,
		          auto_quarantine_expired, expiry_digest_enabled, expiry_digest_weekday,
		          notify_on_auto_quarantine, updated_at`,
		st.PracticeID, st.WarnSoonDays, st.WarnReturnDays, st.WarnCriticalDays, st.ReceiptWarnDays,
		st.AllowExpiredReceipt, st.BlockExpiredOnDAF, st.BlockExpiredOnAdjustOut,
		st.AutoQuarantineExpired, st.ExpiryDigestEnabled, st.ExpiryDigestWeekday,
		st.NotifyOnAutoQuarantine,
	).Scan(
		&st.PracticeID, &st.WarnSoonDays, &st.WarnReturnDays, &st.WarnCriticalDays, &st.ReceiptWarnDays,
		&st.AllowExpiredReceipt, &st.BlockExpiredOnDAF, &st.BlockExpiredOnAdjustOut,
		&st.AutoQuarantineExpired, &st.ExpiryDigestEnabled, &st.ExpiryDigestWeekday,
		&st.NotifyOnAutoQuarantine, &st.UpdatedAt,
	)
	return st, err
}

func (s *Store) ListMedicationDeposits(ctx context.Context, practiceID string) ([]MedicationDeposit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, name, code, is_default
		FROM pharmacy.medication_deposits WHERE practice_id = $1
		ORDER BY is_default DESC, name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MedicationDeposit
	for rows.Next() {
		var d MedicationDeposit
		if err := rows.Scan(&d.ID, &d.PracticeID, &d.Name, &d.Code, &d.IsDefault); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) EnsureDefaultDeposit(ctx context.Context, practiceID string) (MedicationDeposit, error) {
	list, err := s.ListMedicationDeposits(ctx, practiceID)
	if err != nil {
		return MedicationDeposit{}, err
	}
	for _, d := range list {
		if d.IsDefault {
			return d, nil
		}
	}
	if len(list) > 0 {
		return list[0], nil
	}
	return s.CreateMedicationDeposit(ctx, practiceID, "Principal", "MAIN", true)
}

func (s *Store) assertDepositInPractice(ctx context.Context, practiceID, depositID string) error {
	depositID = strings.TrimSpace(depositID)
	if depositID == "" {
		return nil
	}
	if _, err := uuid.Parse(depositID); err != nil {
		return ErrValidation
	}
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pharmacy.medication_deposits WHERE id = $1::uuid AND practice_id = $2::uuid
		)`, depositID, practiceID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrValidation
	}
	return nil
}

func (s *Store) assertDepositInPracticeTx(ctx context.Context, tx pgx.Tx, practiceID, depositID string) error {
	depositID = strings.TrimSpace(depositID)
	if depositID == "" {
		return nil
	}
	if _, err := uuid.Parse(depositID); err != nil {
		return ErrValidation
	}
	var ok bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pharmacy.medication_deposits WHERE id = $1::uuid AND practice_id = $2::uuid
		)`, depositID, practiceID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrValidation
	}
	return nil
}

func (s *Store) CreateMedicationDeposit(ctx context.Context, practiceID, name, code string, isDefault bool) (MedicationDeposit, error) {
	name = strings.TrimSpace(name)
	code = strings.ToUpper(strings.TrimSpace(code))
	if name == "" || code == "" {
		return MedicationDeposit{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MedicationDeposit{}, err
	}
	defer tx.Rollback(ctx)
	if isDefault {
		if _, err := tx.Exec(ctx, `UPDATE pharmacy.medication_deposits SET is_default = false WHERE practice_id = $1`, practiceID); err != nil {
			return MedicationDeposit{}, err
		}
	}
	var d MedicationDeposit
	err = tx.QueryRow(ctx, `
		INSERT INTO pharmacy.medication_deposits (id, practice_id, name, code, is_default)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id::text, practice_id::text, name, code, is_default`,
		uuid.NewString(), practiceID, name, code, isDefault,
	).Scan(&d.ID, &d.PracticeID, &d.Name, &d.Code, &d.IsDefault)
	if err != nil {
		if isUniqueViolationConstraint(err, "medication_deposits_practice_id_code") {
			return MedicationDeposit{}, ErrConflict
		}
		return MedicationDeposit{}, err
	}
	return d, tx.Commit(ctx)
}

type ReceiptInput struct {
	PracticeID   string
	DepositID    string
	MedicationID string
	LotNumber    string
	ExpiresOn    time.Time
	Qty          float64
	Unit         string
	CreatedBy    string
}

func (s *Store) ReceiveMedicationBatch(ctx context.Context, in ReceiptInput, settings PharmacySettings, today time.Time) (MedicationBatch, bool, error) {
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
	} else if err := s.assertDepositInPractice(ctx, in.PracticeID, in.DepositID); err != nil {
		return MedicationBatch{}, false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MedicationBatch{}, false, err
	}
	defer tx.Rollback(ctx)

	batchID := uuid.NewString()
	var id string
	err = tx.QueryRow(ctx, `
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
		INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, created_by)
		VALUES ($1,$2,$3,$4,'receipt',$5)`,
		uuid.NewString(), in.PracticeID, id, in.Qty, nullIfEmpty(in.CreatedBy),
	); err != nil {
		return MedicationBatch{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MedicationBatch{}, false, err
	}
	b, err := s.GetMedicationBatch(ctx, in.PracticeID, id)
	return b, softWarn, err
}

func (s *Store) GetMedicationBatch(ctx context.Context, practiceID, batchID string) (MedicationBatch, error) {
	settings, _ := s.GetPharmacySettings(ctx, practiceID)
	today := pharmacy.BrusselsToday(time.Now())
	var b MedicationBatch
	var exp time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT b.id::text, b.practice_id::text, b.deposit_id::text, d.code, d.name,
		       b.medication_id::text, m.cnk, m.name, b.lot_number, b.expires_on,
		       b.qty_on_hand::float8, b.unit, b.status, m.is_antibiotic,
		       COALESCE(b.quarantine_reason,''), COALESCE(b.waste_reason,'')
		FROM pharmacy.medication_batches b
		JOIN pharmacy.medication_deposits d ON d.id = b.deposit_id
		JOIN pharmacy.ref_medications m ON m.id = b.medication_id
		WHERE b.practice_id = $1 AND b.id = $2`, practiceID, batchID).Scan(
		&b.ID, &b.PracticeID, &b.DepositID, &b.DepositCode, &b.DepositName,
		&b.MedicationID, &b.MedicationCNK, &b.MedicationName, &b.LotNumber, &exp,
		&b.QtyOnHand, &b.Unit, &b.Status, &b.IsAntibiotic, &b.QuarantineReason, &b.WasteReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return MedicationBatch{}, pharmacy.ErrBatchNotFound
	}
	if err != nil {
		return MedicationBatch{}, err
	}
	b.ExpiresOn = exp.Format("2006-01-02")
	th := pharmacy.BandThresholds{SoonDays: settings.WarnSoonDays, ReturnDays: settings.WarnReturnDays, CriticalDays: settings.WarnCriticalDays}
	b.ExpiryBand = string(pharmacy.ClassifyExpiry(exp, today, b.Status, th))
	return b, nil
}

type ListBatchesFilter struct {
	DepositID    string
	MedicationID string
	Band         string // ok|soon|return|critical|expired|quarantine|all
	Status       string
	Q            string
}

func (s *Store) ListMedicationBatches(ctx context.Context, practiceID string, f ListBatchesFilter) ([]MedicationBatch, error) {
	settings, err := s.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	today := pharmacy.BrusselsToday(time.Now())
	rows, err := s.pool.Query(ctx, `
		SELECT b.id::text, b.practice_id::text, b.deposit_id::text, d.code, d.name,
		       b.medication_id::text, m.cnk, m.name, b.lot_number, b.expires_on,
		       b.qty_on_hand::float8, b.unit, b.status, m.is_antibiotic,
		       COALESCE(b.quarantine_reason,''), COALESCE(b.waste_reason,'')
		FROM pharmacy.medication_batches b
		JOIN pharmacy.medication_deposits d ON d.id = b.deposit_id
		JOIN pharmacy.ref_medications m ON m.id = b.medication_id
		WHERE b.practice_id = $1 AND b.qty_on_hand > 0
		  AND ($2 = '' OR b.deposit_id::text = $2)
		  AND ($3 = '' OR b.status = $3)
		  AND ($4 = '' OR m.name_normalized ILIKE '%' || $4 || '%' OR m.cnk ILIKE '%' || $4 || '%' OR b.lot_number ILIKE '%' || $4 || '%')
		  AND ($5 = '' OR b.medication_id::text = $5)
		ORDER BY b.expires_on ASC, m.name ASC`,
		practiceID, f.DepositID, f.Status, strings.ToLower(strings.TrimSpace(f.Q)), strings.TrimSpace(f.MedicationID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	th := pharmacy.BandThresholds{SoonDays: settings.WarnSoonDays, ReturnDays: settings.WarnReturnDays, CriticalDays: settings.WarnCriticalDays}
	var out []MedicationBatch
	for rows.Next() {
		var b MedicationBatch
		var exp time.Time
		if err := rows.Scan(
			&b.ID, &b.PracticeID, &b.DepositID, &b.DepositCode, &b.DepositName,
			&b.MedicationID, &b.MedicationCNK, &b.MedicationName, &b.LotNumber, &exp,
			&b.QtyOnHand, &b.Unit, &b.Status, &b.IsAntibiotic, &b.QuarantineReason, &b.WasteReason,
		); err != nil {
			return nil, err
		}
		b.ExpiresOn = exp.Format("2006-01-02")
		b.ExpiryBand = string(pharmacy.ClassifyExpiry(exp, today, b.Status, th))
		if f.Band != "" && f.Band != "all" && b.ExpiryBand != f.Band {
			continue
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) ExpirySummary(ctx context.Context, practiceID string) (ExpirySummary, error) {
	batches, err := s.ListMedicationBatches(ctx, practiceID, ListBatchesFilter{})
	if err != nil {
		return ExpirySummary{}, err
	}
	var sum ExpirySummary
	for _, b := range batches {
		switch pharmacy.ExpiryBand(b.ExpiryBand) {
		case pharmacy.BandOK:
			sum.OK++
		case pharmacy.BandSoon:
			sum.Soon++
		case pharmacy.BandReturn:
			sum.Return++
		case pharmacy.BandCritical:
			sum.Critical++
		case pharmacy.BandExpired:
			sum.Expired++
		case pharmacy.BandQuarantine:
			sum.Quarantine++
		}
	}
	return sum, nil
}

func (s *Store) AdjustBatchQty(ctx context.Context, practiceID, batchID, userID string, delta float64, detail string, settings PharmacySettings, today time.Time) (MedicationBatch, error) {
	return s.adjustBatchQty(ctx, practiceID, batchID, userID, delta, detail, "", settings, today)
}

func (s *Store) adjustBatchQty(ctx context.Context, practiceID, batchID, userID string, delta float64, detail, inventorySessionID string, settings PharmacySettings, today time.Time) (MedicationBatch, error) {
	if delta == 0 {
		return MedicationBatch{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MedicationBatch{}, err
	}
	defer tx.Rollback(ctx)
	if err := s.adjustBatchQtyTx(ctx, tx, practiceID, batchID, userID, delta, detail, inventorySessionID, settings, today); err != nil {
		return MedicationBatch{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MedicationBatch{}, err
	}
	return s.GetMedicationBatch(ctx, practiceID, batchID)
}

func (s *Store) adjustBatchQtyTx(ctx context.Context, tx pgx.Tx, practiceID, batchID, userID string, delta float64, detail, inventorySessionID string, settings PharmacySettings, today time.Time) error {
	if delta == 0 {
		return ErrValidation
	}
	var status string
	var exp time.Time
	var qty float64
	err := tx.QueryRow(ctx, `
		SELECT status, expires_on, qty_on_hand::float8
		FROM pharmacy.medication_batches
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, batchID).Scan(&status, &exp, &qty)
	if errors.Is(err, pgx.ErrNoRows) {
		return pharmacy.ErrBatchNotFound
	}
	if err != nil {
		return err
	}
	if status == "quarantine" && delta < 0 && strings.TrimSpace(inventorySessionID) == "" {
		return pharmacy.ErrBatchQuarantined
	}
	if status == "wasted" {
		return pharmacy.ErrBatchNotFound
	}
	tod := pharmacy.BrusselsToday(today)
	if delta < 0 && settings.BlockExpiredOnAdjustOut {
		if exp.Before(tod) {
			return pharmacy.ErrBatchExpired
		}
	}
	newQty := qty + delta
	if newQty < 0 {
		return pharmacy.ErrStockInsufficient
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.medication_batches SET qty_on_hand = $3, updated_at = now()
		WHERE practice_id = $1 AND id = $2`, practiceID, batchID, newQty); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, reason_detail, created_by, inventory_session_id)
		VALUES ($1,$2,$3,$4,'adjust',$5,$6,NULLIF($7,'')::uuid)`,
		uuid.NewString(), practiceID, batchID, delta, nullIfEmpty(detail), nullIfEmpty(userID), strings.TrimSpace(inventorySessionID),
	)
	return err
}

func (s *Store) QuarantineBatch(ctx context.Context, practiceID, batchID, userID, reason string) (MedicationBatch, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MedicationBatch{}, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE pharmacy.medication_batches
		SET status = 'quarantine', quarantined_at = now(), quarantine_reason = $3, updated_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'active'`, practiceID, batchID, strings.TrimSpace(reason))
	if err != nil {
		return MedicationBatch{}, err
	}
	if tag.RowsAffected() == 0 {
		return MedicationBatch{}, pharmacy.ErrBatchNotFound
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, reason_detail, created_by)
		VALUES ($1,$2,$3,0,'quarantine',$4,$5)`,
		uuid.NewString(), practiceID, batchID, nullIfEmpty(reason), nullIfEmpty(userID),
	); err != nil {
		return MedicationBatch{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MedicationBatch{}, err
	}
	return s.GetMedicationBatch(ctx, practiceID, batchID)
}

func (s *Store) WasteBatch(ctx context.Context, practiceID, batchID, userID, wasteReason string, qty *float64) (MedicationBatch, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MedicationBatch{}, err
	}
	defer tx.Rollback(ctx)
	var curQty float64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT qty_on_hand::float8, status FROM pharmacy.medication_batches
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, batchID).Scan(&curQty, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return MedicationBatch{}, pharmacy.ErrBatchNotFound
	}
	if err != nil {
		return MedicationBatch{}, err
	}
	delta := -curQty
	newQty := 0.0
	if qty != nil {
		if *qty <= 0 || *qty > curQty {
			return MedicationBatch{}, ErrValidation
		}
		delta = -*qty
		newQty = curQty - *qty
	}
	newStatus := status
	if newQty == 0 {
		newStatus = "wasted"
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.medication_batches
		SET qty_on_hand = $3, status = $4, wasted_at = CASE WHEN $4 = 'wasted' THEN now() ELSE wasted_at END,
		    waste_reason = $5, updated_at = now()
		WHERE practice_id = $1 AND id = $2`,
		practiceID, batchID, newQty, newStatus, strings.TrimSpace(wasteReason),
	); err != nil {
		return MedicationBatch{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, reason_detail, created_by)
		VALUES ($1,$2,$3,$4,'waste',$5,$6)`,
		uuid.NewString(), practiceID, batchID, delta, nullIfEmpty(wasteReason), nullIfEmpty(userID),
	); err != nil {
		return MedicationBatch{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MedicationBatch{}, err
	}
	return s.GetMedicationBatch(ctx, practiceID, batchID)
}

// AllocateFEFO locks valid lots, decrements qty, and writes reason=adjust ledger rows
// (not reason=daf — DAF exits must go through FinalizeDAF with daf_id + daf_item_id).
func (s *Store) AllocateFEFO(ctx context.Context, practiceID, medicationID, depositID string, qty float64, userID string, today time.Time) ([]pharmacy.AllocationLine, error) {
	if qty <= 0 {
		return nil, ErrValidation
	}
	settings, err := s.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	lines, err := s.allocateFEFOTx(ctx, tx, practiceID, medicationID, depositID, qty, today, settings.BlockExpiredOnAdjustOut)
	if err != nil {
		return nil, err
	}
	for _, al := range lines {
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.stock_movements (
				id, practice_id, batch_id, delta, reason, reason_detail, created_by
			) VALUES ($1,$2,$3,$4,'adjust','fefo_allocate',$5)`,
			uuid.NewString(), practiceID, al.BatchID, -al.Qty, nullIfEmpty(userID),
		); err != nil {
			return nil, err
		}
	}
	return lines, tx.Commit(ctx)
}

func (s *Store) allocateFEFOTx(ctx context.Context, tx pgx.Tx, practiceID, medicationID, depositID string, qty float64, today time.Time, blockExpired bool) ([]pharmacy.AllocationLine, error) {
	tod := pharmacy.BrusselsToday(today)
	todStr := tod.Format("2006-01-02")
	allowExpired := !blockExpired
	rows, err := tx.Query(ctx, `
		SELECT id::text, lot_number, expires_on, qty_on_hand::float8
		FROM pharmacy.medication_batches
		WHERE practice_id = $1 AND medication_id = $2
		  AND status = 'active' AND qty_on_hand > 0
		  AND ($3::bool OR expires_on >= $4::date)
		  AND ($5 = '' OR deposit_id::text = $5)
		ORDER BY expires_on ASC, created_at ASC
		FOR UPDATE`, practiceID, medicationID, allowExpired, todStr, depositID)
	if err != nil {
		return nil, err
	}
	type lot struct {
		id, lot string
		exp     time.Time
		qty     float64
	}
	var lots []lot
	for rows.Next() {
		var l lot
		if err := rows.Scan(&l.id, &l.lot, &l.exp, &l.qty); err != nil {
			rows.Close()
			return nil, err
		}
		lots = append(lots, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var total float64
	for _, l := range lots {
		total += l.qty
	}
	if total < qty {
		if blockExpired {
			var invalid float64
			_ = tx.QueryRow(ctx, `
				SELECT COALESCE(SUM(qty_on_hand),0)::float8 FROM pharmacy.medication_batches
				WHERE practice_id = $1 AND medication_id = $2 AND qty_on_hand > 0
				  AND (status <> 'active' OR expires_on < $3::date)`,
				practiceID, medicationID, todStr).Scan(&invalid)
			if invalid > 0 && total == 0 {
				return nil, pharmacy.StockErrForMed(medicationID, pharmacy.ErrStockUnavailableValidLots)
			}
		} else if total == 0 {
			return nil, pharmacy.StockErrForMed(medicationID, pharmacy.ErrStockUnavailableValidLots)
		}
		return nil, pharmacy.StockErrForMed(medicationID, pharmacy.ErrStockInsufficient)
	}

	remaining := qty
	var lines []pharmacy.AllocationLine
	for _, l := range lots {
		if remaining <= 0 {
			break
		}
		take := l.qty
		if take > remaining {
			take = remaining
		}
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.medication_batches SET qty_on_hand = qty_on_hand - $3, updated_at = now()
			WHERE id = $1 AND practice_id = $2 AND qty_on_hand >= $3`, l.id, practiceID, take); err != nil {
			return nil, err
		}
		lines = append(lines, pharmacy.AllocationLine{BatchID: l.id, Qty: take, ExpiresOn: l.exp, LotNumber: l.lot})
		remaining -= take
	}
	if remaining > 1e-9 {
		return nil, pharmacy.ErrStockInsufficient
	}
	return lines, nil
}

func (s *Store) insertDAFStockMovement(ctx context.Context, tx pgx.Tx, practiceID, batchID, dafID, dafItemID, userID, reason, reasonDetail string, delta float64) error {
	if reason != "daf" && reason != "daf_cancel" {
		return ErrValidation
	}
	if strings.TrimSpace(dafID) == "" || strings.TrimSpace(dafItemID) == "" {
		return pharmacy.ErrDAFTraceRequired
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO pharmacy.stock_movements (
			id, practice_id, batch_id, delta, reason, reason_detail, daf_id, daf_item_id, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		uuid.NewString(), practiceID, batchID, delta, reason, nullIfEmpty(reasonDetail),
		dafID, dafItemID, nullIfEmpty(userID),
	)
	return err
}

func (s *Store) AutoQuarantineExpiredBatches(ctx context.Context, practiceID string, today time.Time) (int, []string, error) {
	tod := pharmacy.BrusselsToday(today)
	rows, err := s.pool.Query(ctx, `
		UPDATE pharmacy.medication_batches
		SET status = 'quarantine', quarantined_at = now(),
		    quarantine_reason = COALESCE(NULLIF(quarantine_reason,''), 'expired'),
		    updated_at = now()
		WHERE practice_id = $1 AND status = 'active' AND expires_on < $2::date AND qty_on_hand > 0
		RETURNING id::text`, practiceID, tod.Format("2006-01-02"))
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, nil, err
		}
		ids = append(ids, id)
		_, _ = s.pool.Exec(ctx, `
			INSERT INTO pharmacy.stock_movements (id, practice_id, batch_id, delta, reason, reason_detail)
			VALUES ($1,$2,$3,0,'quarantine','auto_expired')`, uuid.NewString(), practiceID, id)
	}
	return len(ids), ids, rows.Err()
}

func (s *Store) ListPracticesWithPharmacySettings(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT practice_id::text FROM (
			SELECT practice_id FROM pharmacy.practice_settings
			UNION
			SELECT practice_id FROM pharmacy.medication_batches WHERE qty_on_hand > 0
		) t`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) ListStockMovements(ctx context.Context, practiceID string, filter ListMovementsFilter) ([]StockMovement, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT m.id::text, m.practice_id::text, m.batch_id::text, m.delta::float8, m.reason,
		       COALESCE(m.reason_detail,''),
		       COALESCE(m.daf_id::text,''), COALESCE(m.daf_item_id::text,''),
		       COALESCE(b.lot_number,''), COALESCE(b.expires_on::text,''),
		       COALESCE(rm.cnk,''), COALESCE(rm.name,''),
		       COALESCE(m.created_by::text,''), m.created_at::text
		FROM pharmacy.stock_movements m
		LEFT JOIN pharmacy.medication_batches b ON b.id = m.batch_id
		LEFT JOIN pharmacy.ref_medications rm ON rm.id = b.medication_id
		WHERE m.practice_id = $1
		  AND ($2 = '' OR m.daf_id::text = $2)
		ORDER BY m.created_at DESC
		LIMIT $3`, practiceID, strings.TrimSpace(filter.DafID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StockMovement
	for rows.Next() {
		var m StockMovement
		if err := rows.Scan(
			&m.ID, &m.PracticeID, &m.BatchID, &m.Delta, &m.Reason, &m.ReasonDetail,
			&m.DafID, &m.DafItemID, &m.LotNumber, &m.ExpiresOn,
			&m.MedicationCNK, &m.MedicationName, &m.CreatedBy, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// StockMovementsRetentionYears — conservation typique registres (UE 2019/6 / BE).
const StockMovementsRetentionYears = 5

// StockMovementsRetentionStats proves pharmacy movements are kept (not purged by user retention).
type StockMovementsRetentionStats struct {
	PracticeID              string `json:"practiceId"`
	Total                   int    `json:"total"`
	OldestAt                string `json:"oldestAt,omitempty"`
	NewestAt                string `json:"newestAt,omitempty"`
	OlderThanRetentionYears int    `json:"olderThanRetentionYears"`
	RetentionYears          int    `json:"retentionYears"`
	ImmutableAppRole        bool   `json:"immutableAppRole"` // documented: UPDATE/DELETE revoked for petsfollow_app
}

// StockMovementsRetentionStats returns counts for a practice register (preuve rétention 5 ans).
func (s *Store) StockMovementsRetentionStats(ctx context.Context, practiceID string) (StockMovementsRetentionStats, error) {
	out := StockMovementsRetentionStats{
		PracticeID:       practiceID,
		RetentionYears:   StockMovementsRetentionYears,
		ImmutableAppRole: true,
	}
	var oldest, newest *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
		       MIN(created_at),
		       MAX(created_at),
		       COUNT(*) FILTER (WHERE created_at < NOW() - make_interval(years => $2))::int
		FROM pharmacy.stock_movements
		WHERE practice_id = $1`, practiceID, StockMovementsRetentionYears).Scan(
		&out.Total, &oldest, &newest, &out.OlderThanRetentionYears,
	)
	if err != nil {
		return out, err
	}
	if oldest != nil {
		out.OldestAt = oldest.UTC().Format(time.RFC3339)
	}
	if newest != nil {
		out.NewestAt = newest.UTC().Format(time.RFC3339)
	}
	return out, nil
}

func (s *Store) ListPracticeVetEmails(ctx context.Context, practiceID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT email FROM identity.users
		WHERE practice_id = $1 AND role = 'vet' AND COALESCE(email,'') <> ''
		ORDER BY email`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
