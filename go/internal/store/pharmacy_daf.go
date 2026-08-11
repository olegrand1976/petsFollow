package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

type DAFDocument struct {
	ID                   string    `json:"id"`
	PracticeID           string    `json:"practiceId"`
	DAFYear              int       `json:"dafYear"`
	DAFNumber            *int64    `json:"dafNumber,omitempty"`
	DisplayNumber        string    `json:"displayNumber,omitempty"`
	Status               string    `json:"status"`
	ClientUserID         string    `json:"clientUserId,omitempty"`
	PetID                string    `json:"petId,omitempty"`
	VisitID              string    `json:"visitId,omitempty"`
	PrescriberUserID     string    `json:"prescriberUserId"`
	PrescriberName       string    `json:"prescriberName,omitempty"`
	ClientName           string    `json:"clientName,omitempty"`
	PetName              string    `json:"petName,omitempty"`
	PracticeName         string    `json:"practiceName,omitempty"`
	Notes                string    `json:"notes,omitempty"`
	IssuedAt             string    `json:"issuedAt,omitempty"`
	FinalizedAt          string    `json:"finalizedAt,omitempty"`
	CancelledAt          string    `json:"cancelledAt,omitempty"`
	CancelReason         string    `json:"cancelReason,omitempty"`
	PDFObjectKey         string    `json:"pdfObjectKey,omitempty"`
	PDFSHA256            string    `json:"pdfSha256,omitempty"`
	HasAntibiotic        bool      `json:"hasAntibiotic"`
	VamregStatus         string    `json:"vamregStatus"`
	InvoicesExportStatus string    `json:"invoicesExportStatus"`
	Items                []DAFItem `json:"items,omitempty"`
	CreatedAt            string    `json:"createdAt,omitempty"`
}

type DAFItem struct {
	ID                 string          `json:"id"`
	DAFID              string          `json:"dafId"`
	MedicationID       string          `json:"medicationId"`
	MedicationCNK      string          `json:"medicationCnk,omitempty"`
	MedicationName     string          `json:"medicationName,omitempty"`
	BatchID            string          `json:"batchId,omitempty"`
	LotNumber          string          `json:"lotNumber,omitempty"`
	ExpiresOn          string          `json:"expiresOn,omitempty"`
	DepositID          string          `json:"depositId,omitempty"`
	Qty                float64         `json:"qty"`
	Unit               string          `json:"unit"`
	AMMNumber          string          `json:"ammNumber"`
	IsAntibiotic       bool            `json:"isAntibiotic"`
	WithdrawalMeatDays *int            `json:"withdrawalMeatDays,omitempty"`
	WithdrawalMilkDays *int            `json:"withdrawalMilkDays,omitempty"`
	WithdrawalEggsDays *int            `json:"withdrawalEggsDays,omitempty"`
	VamregPayload      json.RawMessage `json:"vamregPayload,omitempty"`
	SortOrder          int             `json:"sortOrder"`
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
	var atc, form, pack, amm string
	var meta []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, cnk, name,
		       COALESCE(atc_code, ''), COALESCE(pharmaceutical_form, ''), COALESCE(pack_size, ''),
		       COALESCE(amm_number, ''),
		       is_antibiotic, is_active,
		       withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days, food_chain_banned,
		       COALESCE(afmps_meta, '{}'::jsonb)
		FROM pharmacy.ref_medications WHERE id = $1`, id).Scan(
		&m.ID, &m.CNK, &m.Name, &atc, &form, &pack, &amm, &m.IsAntibiotic, &m.IsActive,
		&m.WithdrawalMeatDays, &m.WithdrawalMilkDays, &m.WithdrawalEggsDays, &m.FoodChainBanned,
		&meta,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefMedication{}, ErrNotFound
	}
	if err != nil {
		return RefMedication{}, err
	}
	m.ATCCode, m.PharmaceuticalForm, m.PackSize, m.AMMNumber = atc, form, pack, amm
	applyAFMPSMeta(&m, meta)
	return m, nil
}

// GetRefMedicationForPractice returns national ref merged with practice overlay (withdrawal / ban).
func (s *Store) GetRefMedicationForPractice(ctx context.Context, practiceID, medicationID string) (RefMedication, error) {
	if strings.TrimSpace(practiceID) == "" {
		return s.GetRefMedication(ctx, medicationID)
	}
	var m RefMedication
	var atc, form, pack, amm string
	var meta []byte
	err := s.pool.QueryRow(ctx, `
		SELECT m.id::text, m.cnk, m.name,
		       COALESCE(m.atc_code, ''), COALESCE(m.pharmaceutical_form, ''), COALESCE(m.pack_size, ''),
		       COALESCE(m.amm_number, ''),
		       m.is_antibiotic, m.is_active,
		       CASE WHEN o.medication_id IS NOT NULL THEN o.withdrawal_meat_days ELSE m.withdrawal_meat_days END,
		       CASE WHEN o.medication_id IS NOT NULL THEN o.withdrawal_milk_days ELSE m.withdrawal_milk_days END,
		       CASE WHEN o.medication_id IS NOT NULL THEN o.withdrawal_eggs_days ELSE m.withdrawal_eggs_days END,
		       CASE WHEN o.medication_id IS NOT NULL THEN o.food_chain_banned ELSE m.food_chain_banned END,
		       COALESCE(m.afmps_meta, '{}'::jsonb)
		FROM pharmacy.ref_medications m
		LEFT JOIN pharmacy.medication_practice_attrs o
		  ON o.medication_id = m.id AND o.practice_id = $2::uuid
		WHERE m.id = $1`, medicationID, practiceID).Scan(
		&m.ID, &m.CNK, &m.Name, &atc, &form, &pack, &amm, &m.IsAntibiotic, &m.IsActive,
		&m.WithdrawalMeatDays, &m.WithdrawalMilkDays, &m.WithdrawalEggsDays, &m.FoodChainBanned,
		&meta,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefMedication{}, ErrNotFound
	}
	if err != nil {
		return RefMedication{}, err
	}
	m.ATCCode, m.PharmaceuticalForm, m.PackSize, m.AMMNumber = atc, form, pack, amm
	applyAFMPSMeta(&m, meta)
	return m, nil
}

// UpsertPracticeMedicationWithdrawal sets practice-scoped withdrawal / food-chain ban (never mutates national ref).
func (s *Store) UpsertPracticeMedicationWithdrawal(ctx context.Context, practiceID, medicationID string, meat, milk, eggs *int, banned *bool, setMeat, setMilk, setEggs, setBanned bool) (RefMedication, error) {
	if strings.TrimSpace(practiceID) == "" || strings.TrimSpace(medicationID) == "" {
		return RefMedication{}, ErrValidation
	}
	for _, d := range []*int{meat, milk, eggs} {
		if d != nil && *d < 0 {
			return RefMedication{}, ErrValidation
		}
	}
	var active bool
	err := s.pool.QueryRow(ctx, `
		SELECT is_active FROM pharmacy.ref_medications WHERE id = $1`, medicationID).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return RefMedication{}, ErrNotFound
	}
	if err != nil {
		return RefMedication{}, err
	}
	if !active {
		return RefMedication{}, ErrNotFound
	}
	bannedVal := false
	if banned != nil {
		bannedVal = *banned
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO pharmacy.medication_practice_attrs (
			practice_id, medication_id, withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days,
			food_chain_banned, updated_at
		) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, now())
		ON CONFLICT (practice_id, medication_id) DO UPDATE SET
			withdrawal_meat_days = CASE WHEN $7 THEN EXCLUDED.withdrawal_meat_days ELSE pharmacy.medication_practice_attrs.withdrawal_meat_days END,
			withdrawal_milk_days = CASE WHEN $8 THEN EXCLUDED.withdrawal_milk_days ELSE pharmacy.medication_practice_attrs.withdrawal_milk_days END,
			withdrawal_eggs_days = CASE WHEN $9 THEN EXCLUDED.withdrawal_eggs_days ELSE pharmacy.medication_practice_attrs.withdrawal_eggs_days END,
			food_chain_banned = CASE WHEN $10 THEN EXCLUDED.food_chain_banned ELSE pharmacy.medication_practice_attrs.food_chain_banned END,
			updated_at = now()`,
		practiceID, medicationID, meat, milk, eggs, bannedVal, setMeat, setMilk, setEggs, setBanned)
	if err != nil {
		return RefMedication{}, err
	}
	return s.GetRefMedicationForPractice(ctx, practiceID, medicationID)
}

func (s *Store) CreateDAFDraft(ctx context.Context, practiceID, prescriberID string, clientUserID, petID, visitID, notes string, items []DAFItemInput) (DAFDocument, error) {
	if len(items) == 0 {
		return DAFDocument{}, pharmacy.ErrDAFEmpty
	}
	if err := s.validateDAFSpeciesApplicable(ctx, practiceID, petID); err != nil {
		return DAFDocument{}, err
	}
	visitID = strings.TrimSpace(visitID)
	// One draft per visit: collide → replace existing draft items.
	if visitID != "" {
		if existing, err := s.GetDraftDAFByVisit(ctx, practiceID, visitID); err == nil {
			return s.ReplaceDAFDraftItems(ctx, practiceID, existing.ID, clientUserID, petID, visitID, notes, items)
		} else if !errors.Is(err, pharmacy.ErrDAFNotFound) {
			return DAFDocument{}, err
		}
	}
	year := pharmacy.BrusselsToday(time.Now()).Year()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DAFDocument{}, err
	}
	defer tx.Rollback(ctx)

	// Sous quel régime réglementaire ce DAF est-il émis ? Snapshot, comme les ordonnances.
	country, cerr := s.GetPracticeCountryCode(ctx, practiceID)
	if cerr != nil && !errors.Is(cerr, ErrNotFound) {
		return DAFDocument{}, cerr
	}

	docID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.daf_documents (
			id, practice_id, daf_year, status, client_user_id, pet_id, visit_id, prescriber_user_id, notes, country_code
		) VALUES ($1,$2,$3,'draft',NULLIF($4,'')::uuid,NULLIF($5,'')::uuid,NULLIF($6,'')::uuid,$7,NULLIF($8,''),$9)`,
		docID, practiceID, year, clientUserID, petID, visitID, prescriberID, strings.TrimSpace(notes), country,
	)
	if err != nil {
		if visitID != "" && isUniqueViolation(err) {
			tx.Rollback(ctx)
			existing, gerr := s.GetDraftDAFByVisit(ctx, practiceID, visitID)
			if gerr != nil {
				return DAFDocument{}, err
			}
			return s.ReplaceDAFDraftItems(ctx, practiceID, existing.ID, clientUserID, petID, visitID, notes, items)
		}
		return DAFDocument{}, err
	}
	for i, it := range items {
		if err := s.insertDAFItemTx(ctx, tx, practiceID, docID, it, i); err != nil {
			return DAFDocument{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, docID)
}

func (s *Store) insertDAFItemTx(ctx context.Context, tx pgx.Tx, practiceID, dafID string, it DAFItemInput, sort int) error {
	if it.Qty <= 0 || strings.TrimSpace(it.MedicationID) == "" {
		return ErrValidation
	}
	if it.Unit == "" {
		it.Unit = "unit"
	}
	if strings.TrimSpace(it.DepositID) != "" {
		if err := s.assertDepositInPracticeTx(ctx, tx, practiceID, it.DepositID); err != nil {
			return err
		}
	}
	med, err := s.GetRefMedicationForPractice(ctx, practiceID, it.MedicationID)
	if err != nil {
		return err
	}
	payload := it.VamregPayload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO pharmacy.daf_items (
			id, daf_id, medication_id, deposit_id, qty, unit, amm_number, is_antibiotic, vamreg_payload, sort_order,
			withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days
		) VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13)`,
		uuid.NewString(), dafID, it.MedicationID, it.DepositID, it.Qty, it.Unit,
		strings.TrimSpace(it.AMMNumber), med.IsAntibiotic, string(payload), sort,
		med.WithdrawalMeatDays, med.WithdrawalMilkDays, med.WithdrawalEggsDays,
	)
	return err
}

func (s *Store) ReplaceDAFDraftItems(ctx context.Context, practiceID, dafID string, clientUserID, petID, visitID, notes string, items []DAFItemInput) (DAFDocument, error) {
	if len(items) == 0 {
		return DAFDocument{}, pharmacy.ErrDAFEmpty
	}
	if err := s.validateDAFSpeciesApplicable(ctx, practiceID, petID); err != nil {
		return DAFDocument{}, err
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
			visit_id = NULLIF($5,'')::uuid,
			notes = NULLIF($6,''),
			updated_at = now()
		WHERE practice_id = $1 AND id = $2`, practiceID, dafID, clientUserID, petID, visitID, strings.TrimSpace(notes),
	); err != nil {
		return DAFDocument{}, err
	}
	for i, it := range items {
		if err := s.insertDAFItemTx(ctx, tx, practiceID, dafID, it, i); err != nil {
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
		       COALESCE(d.client_user_id::text,''), COALESCE(d.pet_id::text,''), COALESCE(d.visit_id::text,''),
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
			&d.ClientUserID, &d.PetID, &d.VisitID, &d.PrescriberUserID, &d.PrescriberName,
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
		       COALESCE(d.client_user_id::text,''), COALESCE(d.pet_id::text,''), COALESCE(d.visit_id::text,''),
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
		&d.ClientUserID, &d.PetID, &d.VisitID, &d.PrescriberUserID, &d.PrescriberName,
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

// ListFinalizedDAFByPet returns finalized DAF documents for a pet (practice-scoped).
func (s *Store) ListFinalizedDAFByPet(ctx context.Context, practiceID, petID string) ([]DAFDocument, error) {
	petID = strings.TrimSpace(petID)
	if practiceID == "" || petID == "" {
		return nil, ErrValidation
	}
	rows, err := s.pool.Query(ctx, `
		SELECT d.id::text, d.practice_id::text, d.daf_year, d.daf_number, d.status,
		       COALESCE(d.client_user_id::text,''), COALESCE(d.pet_id::text,''), COALESCE(d.visit_id::text,''),
		       d.prescriber_user_id::text, COALESCE(u.full_name,''),
		       COALESCE(d.notes,''), COALESCE(d.issued_at::text,''), COALESCE(d.finalized_at::text,''),
		       d.has_antibiotic, d.vamreg_status, d.invoices_export_status, d.created_at::text,
		       COALESCE(d.pdf_object_key,'')
		FROM pharmacy.daf_documents d
		JOIN identity.users u ON u.id = d.prescriber_user_id
		WHERE d.practice_id = $1 AND d.pet_id = $2::uuid AND d.status = 'finalized'
		ORDER BY d.finalized_at DESC NULLS LAST, d.created_at DESC
		LIMIT 50`, practiceID, petID)
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
			&d.ClientUserID, &d.PetID, &d.VisitID, &d.PrescriberUserID, &d.PrescriberName,
			&d.Notes, &d.IssuedAt, &d.FinalizedAt, &d.HasAntibiotic, &d.VamregStatus,
			&d.InvoicesExportStatus, &d.CreatedAt, &d.PDFObjectKey,
		); err != nil {
			return nil, err
		}
		d.DAFNumber = num
		if num != nil {
			d.DisplayNumber = pharmacy.FormatDAFNumber(d.DAFYear, *num)
		}
		items, err := s.listDAFItems(ctx, d.ID)
		if err != nil {
			return nil, err
		}
		d.Items = items
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListStaleDraftDAFByVisits returns draft DAFs linked to the given visits whose
// updated_at is older than minAge (idle drafts, not freshly edited).
func (s *Store) ListStaleDraftDAFByVisits(ctx context.Context, practiceID string, visitIDs []string, minAge time.Duration) (map[string]string, error) {
	out := map[string]string{}
	if practiceID == "" || len(visitIDs) == 0 {
		return out, nil
	}
	cutoff := time.Now().Add(-minAge)
	rows, err := s.pool.Query(ctx, `
		SELECT visit_id::text, id::text
		FROM pharmacy.daf_documents
		WHERE practice_id = $1
		  AND status = 'draft'
		  AND visit_id = ANY($2::uuid[])
		  AND updated_at < $3`, practiceID, visitIDs, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var visitID, dafID string
		if err := rows.Scan(&visitID, &dafID); err != nil {
			return nil, err
		}
		out[visitID] = dafID
	}
	return out, rows.Err()
}

// GetDraftDAFByVisit returns the unique draft DAF linked to a visit, if any.
func (s *Store) GetDraftDAFByVisit(ctx context.Context, practiceID, visitID string) (DAFDocument, error) {
	visitID = strings.TrimSpace(visitID)
	if practiceID == "" || visitID == "" {
		return DAFDocument{}, ErrValidation
	}
	var dafID string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND visit_id = $2::uuid AND status = 'draft'
		ORDER BY created_at DESC
		LIMIT 1`, practiceID, visitID).Scan(&dafID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, dafID)
}

// GetFinalizedDAFByVisit returns the most recent finalized DAF of a visit, if any.
// Sert au pré-remplissage d'une facture en fin de consultation : seul un DAF
// finalisé a consommé du stock, donc est facturable.
func (s *Store) GetFinalizedDAFByVisit(ctx context.Context, practiceID, visitID string) (DAFDocument, error) {
	visitID = strings.TrimSpace(visitID)
	if practiceID == "" || visitID == "" {
		return DAFDocument{}, ErrValidation
	}
	var dafID string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND visit_id = $2::uuid AND status = 'finalized'
		ORDER BY finalized_at DESC NULLS LAST, created_at DESC
		LIMIT 1`, practiceID, visitID).Scan(&dafID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	return s.GetDAF(ctx, practiceID, dafID)
}

// UpsertDraftDAFForVisit creates or replaces the draft DAF for a consultation visit.
func (s *Store) UpsertDraftDAFForVisit(
	ctx context.Context,
	practiceID, prescriberID, clientUserID, petID, visitID, notes string,
	items []DAFItemInput,
) (DAFDocument, error) {
	visitID = strings.TrimSpace(visitID)
	if visitID == "" {
		return DAFDocument{}, ErrValidation
	}
	return s.CreateDAFDraft(ctx, practiceID, prescriberID, clientUserID, petID, visitID, notes, items)
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
		       i.is_antibiotic, i.vamreg_payload, i.sort_order,
		       i.withdrawal_meat_days, i.withdrawal_milk_days, i.withdrawal_eggs_days
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
			&it.WithdrawalMeatDays, &it.WithdrawalMilkDays, &it.WithdrawalEggsDays,
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
	todStr := tod.Format("2006-01-02")
	settings, err := s.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	allowExpired := !settings.BlockExpiredOnDAF
	var out []FEFOPreviewLine
	for _, it := range items {
		if it.Qty <= 0 {
			return nil, ErrValidation
		}
		if strings.TrimSpace(it.DepositID) != "" {
			if err := s.assertDepositInPractice(ctx, practiceID, it.DepositID); err != nil {
				return nil, err
			}
		}
		med, err := s.GetRefMedication(ctx, it.MedicationID)
		if err != nil {
			return nil, err
		}
		rows, err := s.pool.Query(ctx, `
			SELECT id::text, lot_number, expires_on, qty_on_hand::float8
			FROM pharmacy.medication_batches
			WHERE practice_id = $1 AND medication_id = $2
			  AND status = 'active' AND qty_on_hand > 0
			  AND ($3::bool OR expires_on >= $4::date)
			  AND ($5 = '' OR deposit_id::text = $5)
			ORDER BY expires_on ASC, created_at ASC`,
			practiceID, it.MedicationID, allowExpired, todStr, it.DepositID)
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
				return nil, pharmacy.StockErrForMed(it.MedicationID, pharmacy.ErrStockUnavailableValidLots)
			}
			return nil, pharmacy.StockErrForMed(it.MedicationID, pharmacy.ErrStockInsufficient)
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
	var petID string
	err = tx.QueryRow(ctx, `
		SELECT status, daf_year, COALESCE(pet_id::text,'') FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, dafID).Scan(&status, &year, &petID)
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

	// Le patient a pu changer d'espèce ou sortir de la chaîne alimentaire depuis la
	// création du brouillon : re-vérifier avant d'allouer un numéro inviolable.
	if err := s.validateDAFSpeciesApplicable(ctx, practiceID, petID); err != nil {
		return DAFDocument{}, err
	}

	if err := s.validateDAFFoodChain(ctx, practiceID, petID, draftItems); err != nil {
		return DAFDocument{}, err
	}

	settings, err := s.GetPharmacySettings(ctx, practiceID)
	if err != nil {
		return DAFDocument{}, err
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
		allocs, err := s.allocateFEFOTx(ctx, tx, practiceID, req.MedicationID, req.DepositID, req.Qty, today, settings.BlockExpiredOnDAF)
		if err != nil {
			return DAFDocument{}, err
		}
		for _, al := range allocs {
			itemID := uuid.NewString()
			if _, err := tx.Exec(ctx, `
				INSERT INTO pharmacy.daf_items (
					id, daf_id, medication_id, batch_id, deposit_id, qty, unit, amm_number,
					is_antibiotic, vamreg_payload, sort_order,
					withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days
				) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10::jsonb,$11,$12,$13,$14)`,
				itemID, dafID, req.MedicationID, al.BatchID, req.DepositID, al.Qty, req.Unit,
				req.AMMNumber, req.IsAntibiotic, string(payload), sort,
				req.WithdrawalMeatDays, req.WithdrawalMilkDays, req.WithdrawalEggsDays,
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

	var status, vamregStatus string
	var hasAB bool
	err = tx.QueryRow(ctx, `
		SELECT status, COALESCE(vamreg_status, 'n/a'), has_antibiotic
		FROM pharmacy.daf_documents
		WHERE practice_id = $1 AND id = $2 FOR UPDATE`, practiceID, dafID).Scan(&status, &vamregStatus, &hasAB)
	if errors.Is(err, pgx.ErrNoRows) {
		return DAFDocument{}, pharmacy.ErrDAFNotFound
	}
	if err != nil {
		return DAFDocument{}, err
	}
	if status != "finalized" {
		return DAFDocument{}, pharmacy.ErrDAFNotFinalized
	}
	// No VAMReg retract yet — block cancel while pending/sent to avoid orphan authority records
	// or races with an in-flight declaration worker.
	if hasAB && (vamregStatus == "sent" || vamregStatus == "pending") {
		if vamregStatus == "sent" {
			return DAFDocument{}, pharmacy.ErrDAFVAMRegAlreadySent
		}
		return DAFDocument{}, pharmacy.ErrDAFVAMRegInFlight
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

func (s *Store) validateDAFFoodChain(ctx context.Context, practiceID, petID string, items []DAFItem) error {
	if strings.TrimSpace(petID) == "" {
		return nil
	}
	status, err := s.GetPetFoodChainStatus(ctx, petID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	switch status {
	case "", "companion":
		return nil
	case "food_producing", "excluded_from_food_chain":
		// continue
	default:
		return nil
	}
	for _, it := range items {
		med, err := s.GetRefMedicationForPractice(ctx, practiceID, it.MedicationID)
		if err != nil {
			return err
		}
		if med.FoodChainBanned && status == "food_producing" {
			return pharmacy.ErrFoodChainBannedMedication
		}
		if status == "food_producing" {
			if it.WithdrawalMeatDays == nil || it.WithdrawalMilkDays == nil || it.WithdrawalEggsDays == nil {
				return pharmacy.ErrFoodChainWithdrawalRequired
			}
		}
	}
	return nil
}

// validateDAFSpeciesApplicable refuse un DAF quand l'espèce du patient n'est pas
// productrice de denrées alimentaires dans le pays de la clinique.
//
// Réglementation belge (AR du 21/07/2016) : le DAF vise les animaux producteurs de
// denrées. Les équidés en relèvent sauf exclusion définitive de la chaîne alimentaire
// via leur passeport — d'où la valeur `per_animal`, qui délègue l'arbitrage au statut
// individuel du patient.
//
// Un DAF sans patient lié n'est pas gaté : l'espèce est alors indéterminable.
func (s *Store) validateDAFSpeciesApplicable(ctx context.Context, practiceID, petID string) error {
	if strings.TrimSpace(petID) == "" {
		return nil
	}
	var species, foodChainStatus string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(species,''), COALESCE(food_chain_status,'companion')
		FROM pets.pets WHERE id = $1`, petID).Scan(&species, &foodChainStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	country, err := s.GetPracticeCountryCode(ctx, practiceID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	rule, err := s.ResolveSpeciesRule(ctx, species, country)
	if err != nil {
		return err
	}
	switch rule.DAFRequired {
	case DAFRequiredAlways:
		return nil
	case DAFRequiredPerAnimal:
		if foodChainStatus == "excluded_from_food_chain" {
			return pharmacy.ErrDAFSpeciesNotApplicable
		}
		return nil
	default: // DAFRequiredNever
		return pharmacy.ErrDAFSpeciesNotApplicable
	}
}

// GetPetFoodChainStatus returns companion | food_producing | excluded_from_food_chain.
func (s *Store) GetPetFoodChainStatus(ctx context.Context, petID string) (string, error) {
	var st string
	err := s.pool.QueryRow(ctx, `
		SELECT food_chain_status FROM pets.pets WHERE id = $1`, petID).Scan(&st)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return st, err
}

// GetPetDomicileLocation returns the housing / stable location text.
func (s *Store) GetPetDomicileLocation(ctx context.Context, petID string) (string, error) {
	var loc string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(domicile_location,'') FROM pets.pets WHERE id = $1`, petID).Scan(&loc)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return loc, err
}

const maxPetDomicileRunes = 500

func clipDomicileLocation(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= maxPetDomicileRunes {
		return s
	}
	return string([]rune(s)[:maxPetDomicileRunes])
}

// SetPetRegulatoryFields atomically updates food-chain status and/or domicile (vet practice).
// At least one of status / domicile must be non-nil.
func (s *Store) SetPetRegulatoryFields(ctx context.Context, practiceID, petID string, status, domicile *string) error {
	if status == nil && domicile == nil {
		return ErrValidation
	}
	var stVal any
	if status != nil {
		v := strings.TrimSpace(*status)
		switch v {
		case "companion", "food_producing", "excluded_from_food_chain":
			stVal = v
		default:
			return ErrValidation
		}
	}
	var domVal any
	if domicile != nil {
		domVal = clipDomicileLocation(*domicile)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE pets.pets SET
			food_chain_status = COALESCE($3::text, food_chain_status),
			domicile_location = COALESCE($4::text, domicile_location),
			updated_at = NOW()
		WHERE id = $2 AND practice_id = $1`, practiceID, petID, stVal, domVal)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPetFoodChainStatus updates the animal food-chain classification (vet practice).
func (s *Store) SetPetFoodChainStatus(ctx context.Context, practiceID, petID, status string) error {
	return s.SetPetRegulatoryFields(ctx, practiceID, petID, &status, nil)
}

// SetPetDomicileLocation updates housing location for a practice pet (vet).
func (s *Store) SetPetDomicileLocation(ctx context.Context, practiceID, petID, location string) error {
	return s.SetPetRegulatoryFields(ctx, practiceID, petID, nil, &location)
}
