package store

import (
	"context"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// VamregRefKind values stored in pharmacy.vamreg_ref_codes.kind.
const (
	VamregRefKindTargetSpecies      = "target_species"
	VamregRefKindIndication         = "indication"
	VamregRefKindPharmaceuticalForm = "pharmaceutical_form"
)

// VamregRefCode is a synced AFMPS readonly list entry.
type VamregRefCode struct {
	Kind       string    `json:"kind"`
	Code       string    `json:"code"`
	LabelEN    string    `json:"labelEn"`
	LabelNL    string    `json:"labelNl"`
	LabelFR    string    `json:"labelFr"`
	Deprecated bool      `json:"deprecated"`
	SyncedAt   time.Time `json:"syncedAt"`
}

// VamregRefSyncResult summarizes a sync (dry-run or applied).
type VamregRefSyncResult struct {
	DryRun   bool                       `json:"dryRun"`
	Fetched  map[string]int             `json:"fetched"`
	Upserted map[string]int             `json:"upserted,omitempty"`
	Sample   map[string][]VamregRefCode `json:"sample,omitempty"`
}

// UpsertVamregRefCodes replaces codes for the given kind with the provided rows (full refresh per kind).
func (s *Store) UpsertVamregRefCodes(ctx context.Context, kind string, rows []pharmacy.VamregCodeLabel) (int, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return 0, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.vamreg_ref_codes WHERE kind = $1`, kind); err != nil {
		return 0, err
	}
	n := 0
	now := time.Now().UTC()
	for _, r := range rows {
		code := strings.TrimSpace(r.Code)
		if code == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.vamreg_ref_codes (kind, code, label_en, label_nl, label_fr, deprecated, synced_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			kind, code, strings.TrimSpace(r.En), strings.TrimSpace(r.Nl), strings.TrimSpace(r.Fr), r.Deprecated, now,
		); err != nil {
			return 0, err
		}
		n++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return n, nil
}

// ListVamregRefCodes returns codes for a kind (active first).
func (s *Store) ListVamregRefCodes(ctx context.Context, kind string, includeDeprecated bool, limit int) ([]VamregRefCode, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, ErrValidation
	}
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, `
		SELECT kind, code, label_en, label_nl, label_fr, deprecated, synced_at
		FROM pharmacy.vamreg_ref_codes
		WHERE kind = $1 AND ($2 OR NOT deprecated)
		ORDER BY deprecated ASC, code ASC
		LIMIT $3`, kind, includeDeprecated, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VamregRefCode
	for rows.Next() {
		var r VamregRefCode
		if err := rows.Scan(&r.Kind, &r.Code, &r.LabelEN, &r.LabelNL, &r.LabelFR, &r.Deprecated, &r.SyncedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
