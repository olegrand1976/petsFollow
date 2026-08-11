package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// VamregRefKind values — aliases of pharmacy constants (single source of truth).
const (
	VamregRefKindTargetSpecies      = pharmacy.VamregRefKindTargetSpecies
	VamregRefKindIndication         = pharmacy.VamregRefKindIndication
	VamregRefKindPharmaceuticalForm = pharmacy.VamregRefKindPharmaceuticalForm
)

// ErrVamregRefEmptyList — refuse DELETE+INSERT that would wipe a kind (AFMPS glitch / empty payload).
var ErrVamregRefEmptyList = errors.New("vamreg_ref_empty_list")

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

var vamregRefRequiredKinds = []string{
	VamregRefKindTargetSpecies,
	VamregRefKindIndication,
	VamregRefKindPharmaceuticalForm,
}

// dedupeVamregCodeLabels keeps the first non-empty code (trimmed); drops blanks / duplicates.
func dedupeVamregCodeLabels(rows []pharmacy.VamregCodeLabel) []pharmacy.VamregCodeLabel {
	seen := make(map[string]struct{}, len(rows))
	out := make([]pharmacy.VamregCodeLabel, 0, len(rows))
	for _, r := range rows {
		code := strings.TrimSpace(r.Code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, pharmacy.VamregCodeLabel{
			Code:       code,
			En:         strings.TrimSpace(r.En),
			Nl:         strings.TrimSpace(r.Nl),
			Fr:         strings.TrimSpace(r.Fr),
			Deprecated: r.Deprecated,
		})
	}
	return out
}

// ValidateVamregRefLists ensures every required kind has at least one code after dedupe.
func ValidateVamregRefLists(lists map[string][]pharmacy.VamregCodeLabel) error {
	if lists == nil {
		return ErrVamregRefEmptyList
	}
	for _, kind := range vamregRefRequiredKinds {
		if len(dedupeVamregCodeLabels(lists[kind])) == 0 {
			return ErrVamregRefEmptyList
		}
	}
	return nil
}

// ReplaceVamregRefLists atomically refreshes all required kinds (all-or-nothing).
// Refuses empty lists so a bad AFMPS response cannot wipe pharmacy.vamreg_ref_codes.
func (s *Store) ReplaceVamregRefLists(ctx context.Context, lists map[string][]pharmacy.VamregCodeLabel) (map[string]int, error) {
	if err := ValidateVamregRefLists(lists); err != nil {
		return nil, err
	}
	prepared := make(map[string][]pharmacy.VamregCodeLabel, len(vamregRefRequiredKinds))
	for _, kind := range vamregRefRequiredKinds {
		prepared[kind] = dedupeVamregCodeLabels(lists[kind])
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	upserted := make(map[string]int, len(prepared))
	for _, kind := range vamregRefRequiredKinds {
		rows := prepared[kind]
		if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.vamreg_ref_codes WHERE kind = $1`, kind); err != nil {
			return nil, err
		}
		n := 0
		for _, r := range rows {
			if _, err := tx.Exec(ctx, `
				INSERT INTO pharmacy.vamreg_ref_codes (kind, code, label_en, label_nl, label_fr, deprecated, synced_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				kind, r.Code, r.En, r.Nl, r.Fr, r.Deprecated, now,
			); err != nil {
				return nil, err
			}
			n++
		}
		upserted[kind] = n
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return upserted, nil
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
