package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MedicationPrice is practice sell/purchase pricing for a CNK.
type MedicationPrice struct {
	PracticeID         string  `json:"practiceId"`
	MedicationID       string  `json:"medicationId"`
	MedicationCNK      string  `json:"medicationCnk,omitempty"`
	MedicationName     string  `json:"medicationName,omitempty"`
	PurchasePriceCents int     `json:"purchasePriceCents"`
	SellPriceCents     int     `json:"sellPriceCents"`
	VATPercent         float64 `json:"vatPercent"`
	Currency           string  `json:"currency"`
	UpdatedAt          string  `json:"updatedAt,omitempty"`
}

// ReorderThreshold is min qty alert for a medication (optional deposit).
type ReorderThreshold struct {
	ID           string  `json:"id"`
	PracticeID   string  `json:"practiceId"`
	MedicationID string  `json:"medicationId"`
	DepositID    string  `json:"depositId,omitempty"`
	MinQty       float64 `json:"minQty"`
	UpdatedAt    string  `json:"updatedAt,omitempty"`
}

// ReorderAlert is a medication below threshold.
type ReorderAlert struct {
	MedicationID   string  `json:"medicationId"`
	MedicationCNK  string  `json:"medicationCnk"`
	MedicationName string  `json:"medicationName"`
	DepositID      string  `json:"depositId,omitempty"`
	MinQty         float64 `json:"minQty"`
	OnHandQty      float64 `json:"onHandQty"`
}

// UpsertMedicationPrice sets purchase/sell prices for a practice medication.
func (s *Store) UpsertMedicationPrice(ctx context.Context, practiceID, medicationID, userID string, purchaseCents, sellCents int, vat float64) (MedicationPrice, error) {
	medicationID = strings.TrimSpace(medicationID)
	if medicationID == "" || purchaseCents < 0 || sellCents < 0 {
		return MedicationPrice{}, ErrValidation
	}
	if vat < 0 || vat > 100 {
		return MedicationPrice{}, ErrValidation
	}
	if vat == 0 {
		vat = 21
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pharmacy.ref_medications WHERE id = $1 AND is_active)`, medicationID).Scan(&exists); err != nil {
		return MedicationPrice{}, err
	}
	if !exists {
		return MedicationPrice{}, ErrNotFound
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pharmacy.medication_prices (
			practice_id, medication_id, purchase_price_cents, sell_price_cents, vat_percent, updated_by
		) VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::uuid)
		ON CONFLICT (practice_id, medication_id) DO UPDATE SET
			purchase_price_cents = EXCLUDED.purchase_price_cents,
			sell_price_cents = EXCLUDED.sell_price_cents,
			vat_percent = EXCLUDED.vat_percent,
			updated_by = EXCLUDED.updated_by,
			updated_at = now()`,
		practiceID, medicationID, purchaseCents, sellCents, vat, userID)
	if err != nil {
		return MedicationPrice{}, err
	}
	return s.GetMedicationPrice(ctx, practiceID, medicationID)
}

// GetMedicationPrice returns price or zero values if unset.
func (s *Store) GetMedicationPrice(ctx context.Context, practiceID, medicationID string) (MedicationPrice, error) {
	var p MedicationPrice
	err := s.pool.QueryRow(ctx, `
		SELECT p.practice_id::text, p.medication_id::text, COALESCE(m.cnk,''), COALESCE(m.name,''),
		       p.purchase_price_cents, p.sell_price_cents, p.vat_percent::float8, p.currency, p.updated_at::text
		FROM pharmacy.medication_prices p
		JOIN pharmacy.ref_medications m ON m.id = p.medication_id
		WHERE p.practice_id = $1 AND p.medication_id = $2`, practiceID, medicationID).Scan(
		&p.PracticeID, &p.MedicationID, &p.MedicationCNK, &p.MedicationName,
		&p.PurchasePriceCents, &p.SellPriceCents, &p.VATPercent, &p.Currency, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return MedicationPrice{
			PracticeID: practiceID, MedicationID: medicationID, VATPercent: 21, Currency: "EUR",
		}, nil
	}
	return p, err
}

// ListMedicationPricesByIDs returns the practice prices for the given medications,
// keyed by medication ID. Une clé absente = **pas** de tarif catalogue, ce que
// GetMedicationPrice ne sait pas dire (il renvoie un prix à 0 / TVA 21 par défaut).
//
// Le préremplissage d'une facture ne porte que sur les quelques lignes d'un DAF :
// charger tout le catalogue tarifé d'un cabinet pour en retenir trois ne tient pas
// dès qu'il tarife son stock complet.
func (s *Store) ListMedicationPricesByIDs(ctx context.Context, practiceID string, medicationIDs []string) (map[string]MedicationPrice, error) {
	out := map[string]MedicationPrice{}
	if practiceID == "" || len(medicationIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.medication_id::text, COALESCE(m.cnk,''), COALESCE(m.name,''),
		       p.purchase_price_cents, p.sell_price_cents, p.vat_percent::float8, p.currency
		FROM pharmacy.medication_prices p
		JOIN pharmacy.ref_medications m ON m.id = p.medication_id
		WHERE p.practice_id = $1 AND p.medication_id = ANY($2::uuid[])`, practiceID, medicationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p := MedicationPrice{PracticeID: practiceID}
		if err := rows.Scan(&p.MedicationID, &p.MedicationCNK, &p.MedicationName,
			&p.PurchasePriceCents, &p.SellPriceCents, &p.VATPercent, &p.Currency); err != nil {
			return nil, err
		}
		out[p.MedicationID] = p
	}
	return out, rows.Err()
}

// ListMedicationPrices lists all priced medications for a practice.
func (s *Store) ListMedicationPrices(ctx context.Context, practiceID string) ([]MedicationPrice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.practice_id::text, p.medication_id::text, m.cnk, m.name,
		       p.purchase_price_cents, p.sell_price_cents, p.vat_percent::float8, p.currency, p.updated_at::text
		FROM pharmacy.medication_prices p
		JOIN pharmacy.ref_medications m ON m.id = p.medication_id
		WHERE p.practice_id = $1
		ORDER BY m.name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MedicationPrice
	for rows.Next() {
		var p MedicationPrice
		if err := rows.Scan(&p.PracticeID, &p.MedicationID, &p.MedicationCNK, &p.MedicationName,
			&p.PurchasePriceCents, &p.SellPriceCents, &p.VATPercent, &p.Currency, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertReorderThreshold sets min qty (depositID empty = practice-wide).
func (s *Store) UpsertReorderThreshold(ctx context.Context, practiceID, medicationID, depositID string, minQty float64) (ReorderThreshold, error) {
	medicationID = strings.TrimSpace(medicationID)
	if medicationID == "" || minQty < 0 {
		return ReorderThreshold{}, ErrValidation
	}
	depositID = strings.TrimSpace(depositID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ReorderThreshold{}, err
	}
	defer tx.Rollback(ctx)
	if depositID != "" {
		if err := s.assertDepositInPracticeTx(ctx, tx, practiceID, depositID); err != nil {
			return ReorderThreshold{}, err
		}
	}
	if depositID == "" {
		_, _ = tx.Exec(ctx, `
			DELETE FROM pharmacy.reorder_thresholds
			WHERE practice_id = $1 AND medication_id = $2 AND deposit_id IS NULL`, practiceID, medicationID)
	} else {
		_, _ = tx.Exec(ctx, `
			DELETE FROM pharmacy.reorder_thresholds
			WHERE practice_id = $1 AND medication_id = $2 AND deposit_id = $3::uuid`, practiceID, medicationID, depositID)
	}
	id := uuid.NewString()
	var out ReorderThreshold
	err = tx.QueryRow(ctx, `
		INSERT INTO pharmacy.reorder_thresholds (id, practice_id, medication_id, deposit_id, min_qty)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5)
		RETURNING id::text, practice_id::text, medication_id::text, COALESCE(deposit_id::text,''), min_qty::float8, updated_at::text`,
		id, practiceID, medicationID, depositID, minQty,
	).Scan(&out.ID, &out.PracticeID, &out.MedicationID, &out.DepositID, &out.MinQty, &out.UpdatedAt)
	if err != nil {
		return ReorderThreshold{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ReorderThreshold{}, err
	}
	return out, nil
}

// ListReorderAlerts returns medications whose active on-hand qty is below threshold.
func (s *Store) ListReorderAlerts(ctx context.Context, practiceID string) ([]ReorderAlert, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.medication_id::text, m.cnk, m.name,
		       COALESCE(r.deposit_id::text, ''), r.min_qty::float8,
		       COALESCE((
		         SELECT SUM(b.qty_on_hand)::float8
		         FROM pharmacy.medication_batches b
		         WHERE b.practice_id = r.practice_id
		           AND b.medication_id = r.medication_id
		           AND b.status = 'active'
		           AND b.qty_on_hand > 0
		           AND (r.deposit_id IS NULL OR b.deposit_id = r.deposit_id)
		       ), 0)::float8 AS on_hand
		FROM pharmacy.reorder_thresholds r
		JOIN pharmacy.ref_medications m ON m.id = r.medication_id
		WHERE r.practice_id = $1
		ORDER BY m.name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReorderAlert
	for rows.Next() {
		var a ReorderAlert
		if err := rows.Scan(&a.MedicationID, &a.MedicationCNK, &a.MedicationName, &a.DepositID, &a.MinQty, &a.OnHandQty); err != nil {
			return nil, err
		}
		if a.OnHandQty < a.MinQty {
			out = append(out, a)
		}
	}
	return out, rows.Err()
}
