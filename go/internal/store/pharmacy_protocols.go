package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ClinicalProtocolLine is one medication line in a protocol JSONB array.
type ClinicalProtocolLine struct {
	MedicationID string  `json:"medicationId"`
	Qty          float64 `json:"qty"`
	AMMNumber    string  `json:"ammNumber,omitempty"`
	Unit         string  `json:"unit,omitempty"`
}

type ClinicalProtocol struct {
	ID          string                 `json:"id"`
	PracticeID  string                 `json:"practiceId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Lines       []ClinicalProtocolLine `json:"lines"`
	SortOrder   int                    `json:"sortOrder"`
}

func (s *Store) ListClinicalProtocols(ctx context.Context, practiceID string) ([]ClinicalProtocol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, name, COALESCE(description,''), lines, sort_order
		FROM pharmacy.clinical_protocols
		WHERE practice_id = $1
		ORDER BY sort_order ASC, name ASC
		LIMIT 100`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ClinicalProtocol, 0)
	for rows.Next() {
		var p ClinicalProtocol
		var raw json.RawMessage
		if err := rows.Scan(&p.ID, &p.PracticeID, &p.Name, &p.Description, &raw, &p.SortOrder); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &p.Lines)
		if p.Lines == nil {
			p.Lines = []ClinicalProtocolLine{}
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpsertClinicalProtocol(ctx context.Context, practiceID, name, description string, lines []ClinicalProtocolLine, sortOrder int) (ClinicalProtocol, error) {
	name = strings.TrimSpace(name)
	if practiceID == "" || name == "" || utf8.RuneCountInString(name) > 200 || len(lines) == 0 {
		return ClinicalProtocol{}, ErrValidation
	}
	raw, err := json.Marshal(lines)
	if err != nil {
		return ClinicalProtocol{}, err
	}
	var id string
	err = s.pool.QueryRow(ctx, `
		SELECT id::text FROM pharmacy.clinical_protocols
		WHERE practice_id = $1::uuid AND lower(name) = lower($2)
		LIMIT 1`, practiceID, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO pharmacy.clinical_protocols (id, practice_id, name, description, lines, sort_order)
			VALUES ($1, $2::uuid, $3, $4, $5::jsonb, $6)
			RETURNING id::text`,
			uuid.NewString(), practiceID, name, strings.TrimSpace(description), string(raw), sortOrder,
		).Scan(&id)
	} else if err == nil {
		_, err = s.pool.Exec(ctx, `
			UPDATE pharmacy.clinical_protocols
			SET description = $3, lines = $4::jsonb, sort_order = $5, updated_at = now()
			WHERE id = $1::uuid AND practice_id = $2::uuid`,
			id, practiceID, strings.TrimSpace(description), string(raw), sortOrder)
	}
	if err != nil {
		return ClinicalProtocol{}, err
	}
	return ClinicalProtocol{
		ID: id, PracticeID: practiceID, Name: name,
		Description: strings.TrimSpace(description), Lines: lines, SortOrder: sortOrder,
	}, nil
}
