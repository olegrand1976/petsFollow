package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var hexColorRE = regexp.MustCompile(`(?i)^#[0-9a-f]{6}$`)

type VisitType struct {
	ID              string `json:"id"`
	PracticeID      string `json:"practiceId"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"durationMinutes"`
	Color           string `json:"color"`
	IsActive        bool   `json:"isActive"`
	SortOrder       int    `json:"sortOrder"`
	// Tarif de l'acte (HTVA, centimes) et taux de TVA, utilisés pour pré-remplir
	// une facture en fin de consultation. 0 = non tarifé : aucune ligne proposée.
	PriceExclCents int     `json:"priceExclCents"`
	VATPercent     float64 `json:"vatPercent"`
}

type VisitTypeInput struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	DurationMinutes int      `json:"durationMinutes"`
	Color           string   `json:"color"`
	IsActive        *bool    `json:"isActive,omitempty"`
	SortOrder       int      `json:"sortOrder"`
	// Tarif absent du payload = inchangé (0 / 21 % à la création), et non remis à
	// zéro : l'écran Agenda masque ces champs quand la facturation est coupée, et
	// tout client qui ignore le tarif effacerait sinon celui du cabinet en
	// enregistrant un simple renommage de type.
	PriceExclCents *int     `json:"priceExclCents,omitempty"`
	VATPercent     *float64 `json:"vatPercent,omitempty"`
}

func NormalizeVisitDuration(minutes int) (int, error) {
	if minutes < 5 || minutes > 480 {
		return 0, fmt.Errorf("%w: invalid_duration", ErrValidation)
	}
	return minutes, nil
}

// validateVisitTypePricing borne le tarif comme la facturation borne ses lignes
// (`invoicing.ValidateLines`) : un prix ou une TVA hors bornes ici produirait une
// facture refusée plus tard, avec une erreur incompréhensible pour le véto.
// Une valeur absente n'est pas validée : elle laisse le tarif en place.
func validateVisitTypePricing(priceExclCents *int, vatPercent *float64) error {
	if priceExclCents != nil && (*priceExclCents < 0 || *priceExclCents > 100_000_000) {
		return fmt.Errorf("%w: invalid_price", ErrValidation)
	}
	if vatPercent != nil {
		vat := *vatPercent
		if math.IsNaN(vat) || math.IsInf(vat, 0) || vat < 0 || vat > 100 {
			return fmt.Errorf("%w: invalid_vat_percent", ErrValidation)
		}
	}
	return nil
}

func normalizeVisitTypeColor(color string) (string, error) {
	c := strings.TrimSpace(color)
	if !hexColorRE.MatchString(c) {
		return "", fmt.Errorf("%w: invalid_color", ErrValidation)
	}
	return strings.ToUpper(c), nil
}

func (s *Store) ListVisitTypes(ctx context.Context, practiceID string, activeOnly bool) ([]VisitType, error) {
	q := `
		SELECT id::text, practice_id::text, name, duration_minutes, color, is_active, sort_order,
		       price_excl_cents, vat_percent
		FROM practice.visit_types
		WHERE practice_id = $1`
	if activeOnly {
		q += ` AND is_active = true`
	}
	q += ` ORDER BY sort_order ASC, name ASC`
	rows, err := s.pool.Query(ctx, q, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisitType
	for rows.Next() {
		var vt VisitType
		if err := rows.Scan(&vt.ID, &vt.PracticeID, &vt.Name, &vt.DurationMinutes, &vt.Color, &vt.IsActive, &vt.SortOrder,
			&vt.PriceExclCents, &vt.VATPercent); err != nil {
			return nil, err
		}
		out = append(out, vt)
	}
	if out == nil {
		out = []VisitType{}
	}
	return out, rows.Err()
}

func (s *Store) GetVisitType(ctx context.Context, practiceID, typeID string) (VisitType, error) {
	var vt VisitType
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, name, duration_minutes, color, is_active, sort_order,
		       price_excl_cents, vat_percent
		FROM practice.visit_types
		WHERE id = $1 AND practice_id = $2`, typeID, practiceID,
	).Scan(&vt.ID, &vt.PracticeID, &vt.Name, &vt.DurationMinutes, &vt.Color, &vt.IsActive, &vt.SortOrder,
		&vt.PriceExclCents, &vt.VATPercent)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitType{}, ErrNotFound
	}
	return vt, err
}

// PutVisitTypes replaces the practice visit-type catalogue (upsert by id, delete missing).
func (s *Store) PutVisitTypes(ctx context.Context, practiceID string, items []VisitTypeInput) ([]VisitType, error) {
	if items == nil {
		items = []VisitTypeInput{}
	}
	seenNames := map[string]struct{}{}
	for i := range items {
		name := strings.TrimSpace(items[i].Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name_required", ErrValidation)
		}
		key := strings.ToLower(name)
		if _, ok := seenNames[key]; ok {
			return nil, fmt.Errorf("%w: duplicate_name", ErrValidation)
		}
		seenNames[key] = struct{}{}
		dur, err := NormalizeVisitDuration(items[i].DurationMinutes)
		if err != nil {
			return nil, err
		}
		color, err := normalizeVisitTypeColor(items[i].Color)
		if err != nil {
			return nil, err
		}
		if err := validateVisitTypePricing(items[i].PriceExclCents, items[i].VATPercent); err != nil {
			return nil, err
		}
		items[i].Name = name
		items[i].DurationMinutes = dur
		items[i].Color = color
		if items[i].IsActive == nil {
			active := true
			items[i].IsActive = &active
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	existing := map[string]struct{}{}
	erows, err := tx.Query(ctx, `SELECT id::text FROM practice.visit_types WHERE practice_id = $1`, practiceID)
	if err != nil {
		return nil, err
	}
	for erows.Next() {
		var id string
		if err := erows.Scan(&id); err != nil {
			erows.Close()
			return nil, err
		}
		existing[id] = struct{}{}
	}
	erows.Close()
	if err := erows.Err(); err != nil {
		return nil, err
	}

	keepIDs := make([]string, 0, len(items))
	for i, it := range items {
		id := strings.TrimSpace(it.ID)
		if id != "" {
			if _, ok := existing[id]; !ok {
				return nil, fmt.Errorf("%w: unknown_visit_type", ErrValidation)
			}
		} else {
			id = uuid.NewString()
			items[i].ID = id
		}
		keepIDs = append(keepIDs, id)
	}

	// Free unique names before upsert (replace-all with new IDs).
	if len(keepIDs) == 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM practice.visit_types WHERE practice_id = $1`, practiceID); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			DELETE FROM practice.visit_types
			WHERE practice_id = $1 AND NOT (id = ANY($2::uuid[]))`, practiceID, keepIDs); err != nil {
			return nil, err
		}
	}

	for i := range items {
		it := items[i]
		active := true
		if it.IsActive != nil {
			active = *it.IsActive
		}
		// COALESCE et non EXCLUDED sur le tarif : un payload sans prix laisse
		// celui déjà enregistré (cf. VisitTypeInput).
		_, err := tx.Exec(ctx, `
			INSERT INTO practice.visit_types (
				id, practice_id, name, duration_minutes, color, is_active, sort_order,
				price_excl_cents, vat_percent
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8::int, 0), COALESCE($9::numeric, 21))
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				duration_minutes = EXCLUDED.duration_minutes,
				color = EXCLUDED.color,
				is_active = EXCLUDED.is_active,
				sort_order = EXCLUDED.sort_order,
				price_excl_cents = COALESCE($8::int, practice.visit_types.price_excl_cents),
				vat_percent = COALESCE($9::numeric, practice.visit_types.vat_percent)
			WHERE practice.visit_types.practice_id = $2`,
			it.ID, practiceID, it.Name, it.DurationMinutes, it.Color, active, it.SortOrder,
			it.PriceExclCents, it.VATPercent,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.ListVisitTypes(ctx, practiceID, false)
}
