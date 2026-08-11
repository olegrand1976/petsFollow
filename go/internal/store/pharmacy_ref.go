package store

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RefMedication is a national dictionary entry (CNK / AFMPS).
type RefMedication struct {
	ID                 string          `json:"id"`
	CNK                string          `json:"cnk"`
	Name               string          `json:"name"`
	ATCCode            string          `json:"atcCode,omitempty"`
	PharmaceuticalForm string          `json:"pharmaceuticalForm,omitempty"`
	PackSize           string          `json:"packSize,omitempty"`
	AMMNumber          string          `json:"ammNumber,omitempty"`
	IsAntibiotic       bool            `json:"isAntibiotic"`
	IsActive           bool            `json:"isActive"`
	WithdrawalMeatDays *int            `json:"withdrawalMeatDays,omitempty"`
	WithdrawalMilkDays *int            `json:"withdrawalMilkDays,omitempty"`
	WithdrawalEggsDays *int            `json:"withdrawalEggsDays,omitempty"`
	FoodChainBanned    bool            `json:"foodChainBanned,omitempty"`
	AFMPSMeta          json.RawMessage `json:"afmpsMeta,omitempty"`
	UpdatedAt          string          `json:"updatedAt,omitempty"`
}

// RefMedicationUpsert is one row for ImportCNK / UpsertRefMedication.
type RefMedicationUpsert struct {
	CNK                string
	Name               string
	ATCCode            string
	PharmaceuticalForm string
	PackSize           string
	AMMNumber          string
	IsAntibiotic       bool
	IsActive           bool
	AFMPSMeta          json.RawMessage
}

// NormalizeMedicationName lowercases and strips diacritics for search.
func NormalizeMedicationName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	return strings.ToLower(out)
}

// SearchRefMedications autocomplete (CNK + name). Min 2 runes. Limit capped at 50.
func (s *Store) SearchRefMedications(ctx context.Context, q string, limit int) ([]RefMedication, error) {
	q = strings.TrimSpace(q)
	if utf8.RuneCountInString(q) < 2 {
		return []RefMedication{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	normQ := NormalizeMedicationName(q)
	pattern := "%" + escapeILIKE(normQ) + "%"
	cnkPattern := "%" + escapeILIKE(strings.ToLower(q)) + "%"

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, cnk, name,
		       COALESCE(atc_code, ''), COALESCE(pharmaceutical_form, ''), COALESCE(pack_size, ''),
		       COALESCE(amm_number, ''),
		       is_antibiotic, is_active,
		       withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days, food_chain_banned
		FROM pharmacy.ref_medications
		WHERE is_active
		  AND (
		    lower(cnk) LIKE $1 ESCAPE '\'
		    OR name_normalized LIKE $2 ESCAPE '\'
		    OR ($3 <> '' AND name_normalized % $3)
		  )
		ORDER BY
		  CASE WHEN lower(cnk) LIKE $4 ESCAPE '\' THEN 0 ELSE 1 END,
		  CASE WHEN name_normalized LIKE $4 ESCAPE '\' THEN 0 ELSE 1 END,
		  similarity(name_normalized, $3) DESC NULLS LAST,
		  name ASC
		LIMIT $5`,
		cnkPattern, pattern, normQ, escapeILIKE(normQ)+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RefMedication, 0, limit)
	for rows.Next() {
		var m RefMedication
		var atc, form, pack, amm string
		if err := rows.Scan(&m.ID, &m.CNK, &m.Name, &atc, &form, &pack, &amm, &m.IsAntibiotic, &m.IsActive,
			&m.WithdrawalMeatDays, &m.WithdrawalMilkDays, &m.WithdrawalEggsDays, &m.FoodChainBanned); err != nil {
			return nil, err
		}
		m.ATCCode = atc
		m.PharmaceuticalForm = form
		m.PackSize = pack
		m.AMMNumber = amm
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetRefMedicationByCNK returns an active national dictionary row by exact CNK.
func (s *Store) GetRefMedicationByCNK(ctx context.Context, cnk string) (RefMedication, error) {
	cnk = strings.TrimSpace(cnk)
	if cnk == "" {
		return RefMedication{}, ErrNotFound
	}
	var m RefMedication
	var atc, form, pack, amm string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, cnk, name,
		       COALESCE(atc_code, ''), COALESCE(pharmaceutical_form, ''), COALESCE(pack_size, ''),
		       COALESCE(amm_number, ''),
		       is_antibiotic, is_active,
		       withdrawal_meat_days, withdrawal_milk_days, withdrawal_eggs_days, food_chain_banned
		FROM pharmacy.ref_medications
		WHERE is_active AND cnk = $1
		LIMIT 1`, cnk).Scan(
		&m.ID, &m.CNK, &m.Name, &atc, &form, &pack, &amm, &m.IsAntibiotic, &m.IsActive,
		&m.WithdrawalMeatDays, &m.WithdrawalMilkDays, &m.WithdrawalEggsDays, &m.FoodChainBanned,
	)
	if err == pgx.ErrNoRows {
		return RefMedication{}, ErrNotFound
	}
	if err != nil {
		return RefMedication{}, err
	}
	m.ATCCode = atc
	m.PharmaceuticalForm = form
	m.PackSize = pack
	m.AMMNumber = amm
	return m, nil
}

// UpsertRefMedication inserts or updates by CNK. name_normalized is computed in Go
// (unaccent may be unavailable in some environments; keep app-side normalize consistent).
func (s *Store) UpsertRefMedication(ctx context.Context, row RefMedicationUpsert) (string, error) {
	name := strings.TrimSpace(row.Name)
	cnk := strings.TrimSpace(row.CNK)
	if cnk == "" || name == "" {
		return "", ErrValidation
	}
	norm := NormalizeMedicationName(name)
	meta := row.AFMPSMeta
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	id := uuid.NewString()
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pharmacy.ref_medications (
			id, cnk, name, name_normalized, atc_code, pharmaceutical_form, pack_size,
			amm_number, is_antibiotic, is_active, afmps_meta, updated_at
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
			COALESCE($8, ''), $9, $10, $11::jsonb, now()
		)
		ON CONFLICT (cnk) DO UPDATE SET
			name = EXCLUDED.name,
			name_normalized = EXCLUDED.name_normalized,
			atc_code = EXCLUDED.atc_code,
			pharmaceutical_form = EXCLUDED.pharmaceutical_form,
			pack_size = EXCLUDED.pack_size,
			amm_number = CASE
				WHEN EXCLUDED.amm_number <> '' THEN EXCLUDED.amm_number
				ELSE pharmacy.ref_medications.amm_number
			END,
			is_antibiotic = EXCLUDED.is_antibiotic,
			is_active = EXCLUDED.is_active,
			afmps_meta = EXCLUDED.afmps_meta,
			updated_at = now()
		RETURNING id::text`,
		id, cnk, name, norm,
		strings.TrimSpace(row.ATCCode),
		strings.TrimSpace(row.PharmaceuticalForm),
		strings.TrimSpace(row.PackSize),
		strings.TrimSpace(row.AMMNumber),
		row.IsAntibiotic, row.IsActive, string(meta),
	).Scan(&id)
	return id, err
}

// DeactivateMissingRefMedications soft-disables CNKs not in keep set (optional import flag).
func (s *Store) DeactivateMissingRefMedications(ctx context.Context, keepCNKs []string) (int64, error) {
	if len(keepCNKs) == 0 {
		return 0, nil
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.ref_medications
		SET is_active = false, updated_at = now()
		WHERE is_active AND NOT (cnk = ANY($1))`, keepCNKs)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
