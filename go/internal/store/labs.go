package store

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	MaxLabNameLen  = 120
	MaxLabNotesLen = 2000
)

var (
	ErrInvalidLabPanel   = errors.New("invalid lab panel")
	ErrUnknownAnalyte    = errors.New("unknown analyte")
	ErrDuplicateAnalyte  = errors.New("duplicate analyte")
)

// LabAnalyteCatalog — codes V1 stables (labels i18n côté clients).
var LabAnalyteCatalog = map[string]struct{}{
	"crea": {}, "urea": {}, "bun": {},
	"alat": {}, "alt": {}, "asat": {}, "ast": {}, "alp": {}, "ggt": {},
	"hct": {}, "wbc": {}, "plt": {},
	"glucose": {}, "tp": {}, "protein": {},
}

type LabResultFlag string

const (
	LabFlagLow     LabResultFlag = "low"
	LabFlagNormal  LabResultFlag = "normal"
	LabFlagHigh    LabResultFlag = "high"
	LabFlagUnknown LabResultFlag = "unknown"
)

type LabPanelResult struct {
	ID          string        `json:"id"`
	PanelID     string        `json:"panelId"`
	AnalyteCode string        `json:"analyteCode"`
	ValueNum    *float64      `json:"valueNum,omitempty"`
	ValueText   *string       `json:"valueText,omitempty"`
	Unit        string        `json:"unit"`
	RefLow      *float64      `json:"refLow,omitempty"`
	RefHigh     *float64      `json:"refHigh,omitempty"`
	Flag        LabResultFlag `json:"flag"`
}

type LabPanel struct {
	ID           string           `json:"id"`
	PetID        string           `json:"petId"`
	PracticeID   string           `json:"practiceId"`
	AuthorUserID string           `json:"authorUserId"`
	CollectedAt  time.Time        `json:"collectedAt"`
	LabName      string           `json:"labName"`
	Notes        string           `json:"notes"`
	DocumentID   *string          `json:"documentId,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	Results      []LabPanelResult `json:"results,omitempty"`
	AbnormalCount int             `json:"abnormalCount,omitempty"`
}

type LabPanelResultInput struct {
	AnalyteCode string   `json:"analyteCode"`
	ValueNum    *float64 `json:"valueNum"`
	ValueText   *string  `json:"valueText"`
	Unit        string   `json:"unit"`
	RefLow      *float64 `json:"refLow"`
	RefHigh     *float64 `json:"refHigh"`
}

func IsKnownAnalyte(code string) bool {
	_, ok := LabAnalyteCatalog[strings.TrimSpace(strings.ToLower(code))]
	return ok
}

func ComputeLabFlag(value *float64, refLow, refHigh *float64) LabResultFlag {
	if value == nil {
		return LabFlagUnknown
	}
	if refLow != nil && *value < *refLow {
		return LabFlagLow
	}
	if refHigh != nil && *value > *refHigh {
		return LabFlagHigh
	}
	if refLow != nil || refHigh != nil {
		return LabFlagNormal
	}
	return LabFlagUnknown
}

func normalizeLabName(raw string) string {
	s := strings.TrimSpace(raw)
	if utf8.RuneCountInString(s) > MaxLabNameLen {
		s = string([]rune(s)[:MaxLabNameLen])
	}
	return s
}

func normalizeLabNotes(raw string) string {
	s := strings.TrimSpace(raw)
	if utf8.RuneCountInString(s) > MaxLabNotesLen {
		s = string([]rune(s)[:MaxLabNotesLen])
	}
	return s
}

func (s *Store) CreateLabPanel(
	ctx context.Context,
	petID, practiceID, authorID string,
	collectedAt time.Time,
	labName, notes string,
	documentID *string,
	results []LabPanelResultInput,
) (LabPanel, error) {
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}
	labName = normalizeLabName(labName)
	notes = normalizeLabNotes(notes)

	seen := map[string]struct{}{}
	normalized := make([]LabPanelResultInput, 0, len(results))
	for _, in := range results {
		code := strings.TrimSpace(strings.ToLower(in.AnalyteCode))
		if code == "" {
			continue
		}
		if !IsKnownAnalyte(code) {
			return LabPanel{}, ErrUnknownAnalyte
		}
		if _, dup := seen[code]; dup {
			return LabPanel{}, ErrDuplicateAnalyte
		}
		seen[code] = struct{}{}
		in.AnalyteCode = code
		in.Unit = strings.TrimSpace(in.Unit)
		if in.ValueText != nil {
			t := strings.TrimSpace(*in.ValueText)
			if t == "" {
				in.ValueText = nil
			} else {
				in.ValueText = &t
			}
		}
		if in.ValueNum == nil && in.ValueText == nil {
			return LabPanel{}, ErrInvalidLabPanel
		}
		if in.RefLow != nil && in.RefHigh != nil && *in.RefLow > *in.RefHigh {
			return LabPanel{}, ErrInvalidLabPanel
		}
		if utf8.RuneCountInString(in.Unit) > 32 {
			in.Unit = string([]rune(in.Unit)[:32])
		}
		if in.ValueText != nil && utf8.RuneCountInString(*in.ValueText) > 64 {
			t := string([]rune(*in.ValueText)[:64])
			in.ValueText = &t
		}
		normalized = append(normalized, in)
	}
	if len(normalized) == 0 {
		return LabPanel{}, ErrInvalidLabPanel
	}

	panel := LabPanel{
		ID:           uuid.NewString(),
		PetID:        petID,
		PracticeID:   practiceID,
		AuthorUserID: authorID,
		CollectedAt:  collectedAt.UTC(),
		LabName:      labName,
		Notes:        notes,
		DocumentID:   documentID,
		Results:      []LabPanelResult{},
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LabPanel{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO labs.panels (
			id, pet_id, practice_id, author_user_id, collected_at,
			lab_name, notes, document_id, created_at, updated_at
		) VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,NULLIF($8,'')::uuid,NOW(),NOW())
		RETURNING created_at, updated_at`,
		panel.ID, panel.PetID, panel.PracticeID, panel.AuthorUserID, panel.CollectedAt,
		panel.LabName, panel.Notes, nullableStr(documentID),
	).Scan(&panel.CreatedAt, &panel.UpdatedAt)
	if err != nil {
		return LabPanel{}, err
	}

	abnormal := 0
	for _, in := range normalized {
		flag := ComputeLabFlag(in.ValueNum, in.RefLow, in.RefHigh)
		if flag == LabFlagLow || flag == LabFlagHigh {
			abnormal++
		}
		res := LabPanelResult{
			ID:          uuid.NewString(),
			PanelID:     panel.ID,
			AnalyteCode: in.AnalyteCode,
			ValueNum:    in.ValueNum,
			ValueText:   in.ValueText,
			Unit:        in.Unit,
			RefLow:      in.RefLow,
			RefHigh:     in.RefHigh,
			Flag:        flag,
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO labs.panel_results (
				id, panel_id, analyte_code, value_num, value_text, unit, ref_low, ref_high, flag
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			res.ID, res.PanelID, res.AnalyteCode, res.ValueNum, res.ValueText,
			res.Unit, res.RefLow, res.RefHigh, string(res.Flag),
		); err != nil {
			return LabPanel{}, err
		}
		panel.Results = append(panel.Results, res)
	}
	panel.AbnormalCount = abnormal

	if err := tx.Commit(ctx); err != nil {
		return LabPanel{}, err
	}
	return panel, nil
}

func nullableStr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func (s *Store) ListLabPanels(ctx context.Context, petID string) ([]LabPanel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.pet_id::text, COALESCE(p.practice_id::text, ''),
			p.author_user_id::text, p.collected_at, p.lab_name, p.notes,
			p.document_id::text, p.created_at, p.updated_at,
			COALESCE((
				SELECT COUNT(*)::int FROM labs.panel_results r
				WHERE r.panel_id = p.id AND r.flag IN ('low','high')
			), 0)
		FROM labs.panels p
		WHERE p.pet_id = $1
		ORDER BY p.collected_at DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LabPanel
	for rows.Next() {
		var p LabPanel
		var docID *string
		if err := rows.Scan(
			&p.ID, &p.PetID, &p.PracticeID, &p.AuthorUserID, &p.CollectedAt,
			&p.LabName, &p.Notes, &docID, &p.CreatedAt, &p.UpdatedAt, &p.AbnormalCount,
		); err != nil {
			return nil, err
		}
		if docID != nil && *docID != "" {
			p.DocumentID = docID
		}
		out = append(out, p)
	}
	if out == nil {
		out = []LabPanel{}
	}
	return out, rows.Err()
}

func (s *Store) GetLabPanel(ctx context.Context, panelID string) (LabPanel, error) {
	var p LabPanel
	var docID *string
	err := s.pool.QueryRow(ctx, `
		SELECT p.id::text, p.pet_id::text, COALESCE(p.practice_id::text, ''),
			p.author_user_id::text, p.collected_at, p.lab_name, p.notes,
			p.document_id::text, p.created_at, p.updated_at
		FROM labs.panels p WHERE p.id = $1`, panelID).Scan(
		&p.ID, &p.PetID, &p.PracticeID, &p.AuthorUserID, &p.CollectedAt,
		&p.LabName, &p.Notes, &docID, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return LabPanel{}, ErrNotFound
	}
	if err != nil {
		return LabPanel{}, err
	}
	if docID != nil && *docID != "" {
		p.DocumentID = docID
	}
	results, err := s.listLabPanelResults(ctx, p.ID)
	if err != nil {
		return LabPanel{}, err
	}
	p.Results = results
	abnormal := 0
	for _, r := range results {
		if r.Flag == LabFlagLow || r.Flag == LabFlagHigh {
			abnormal++
		}
	}
	p.AbnormalCount = abnormal
	return p, nil
}

func (s *Store) listLabPanelResults(ctx context.Context, panelID string) ([]LabPanelResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, panel_id::text, analyte_code, value_num, value_text,
			unit, ref_low, ref_high, flag
		FROM labs.panel_results
		WHERE panel_id = $1
		ORDER BY analyte_code`, panelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LabPanelResult
	for rows.Next() {
		var r LabPanelResult
		var flag string
		if err := rows.Scan(
			&r.ID, &r.PanelID, &r.AnalyteCode, &r.ValueNum, &r.ValueText,
			&r.Unit, &r.RefLow, &r.RefHigh, &flag,
		); err != nil {
			return nil, err
		}
		r.Flag = LabResultFlag(flag)
		out = append(out, r)
	}
	if out == nil {
		out = []LabPanelResult{}
	}
	return out, rows.Err()
}

func (s *Store) ReplaceLabPanelResults(ctx context.Context, panelID string, results []LabPanelResultInput) (LabPanel, error) {
	panel, err := s.GetLabPanel(ctx, panelID)
	if err != nil {
		return LabPanel{}, err
	}
	seen := map[string]struct{}{}
	normalized := make([]LabPanelResultInput, 0, len(results))
	for _, in := range results {
		code := strings.TrimSpace(strings.ToLower(in.AnalyteCode))
		if code == "" {
			continue
		}
		if !IsKnownAnalyte(code) {
			return LabPanel{}, ErrUnknownAnalyte
		}
		if _, dup := seen[code]; dup {
			return LabPanel{}, ErrDuplicateAnalyte
		}
		seen[code] = struct{}{}
		in.AnalyteCode = code
		in.Unit = strings.TrimSpace(in.Unit)
		if in.ValueText != nil {
			t := strings.TrimSpace(*in.ValueText)
			if t == "" {
				in.ValueText = nil
			} else {
				in.ValueText = &t
			}
		}
		if in.ValueNum == nil && in.ValueText == nil {
			return LabPanel{}, ErrInvalidLabPanel
		}
		if in.RefLow != nil && in.RefHigh != nil && *in.RefLow > *in.RefHigh {
			return LabPanel{}, ErrInvalidLabPanel
		}
		if utf8.RuneCountInString(in.Unit) > 32 {
			in.Unit = string([]rune(in.Unit)[:32])
		}
		if in.ValueText != nil && utf8.RuneCountInString(*in.ValueText) > 64 {
			t := string([]rune(*in.ValueText)[:64])
			in.ValueText = &t
		}
		normalized = append(normalized, in)
	}
	if len(normalized) == 0 {
		return LabPanel{}, ErrInvalidLabPanel
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LabPanel{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM labs.panel_results WHERE panel_id = $1`, panelID); err != nil {
		return LabPanel{}, err
	}
	panel.Results = []LabPanelResult{}
	abnormal := 0
	for _, in := range normalized {
		flag := ComputeLabFlag(in.ValueNum, in.RefLow, in.RefHigh)
		if flag == LabFlagLow || flag == LabFlagHigh {
			abnormal++
		}
		res := LabPanelResult{
			ID:          uuid.NewString(),
			PanelID:     panelID,
			AnalyteCode: in.AnalyteCode,
			ValueNum:    in.ValueNum,
			ValueText:   in.ValueText,
			Unit:        in.Unit,
			RefLow:      in.RefLow,
			RefHigh:     in.RefHigh,
			Flag:        flag,
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO labs.panel_results (
				id, panel_id, analyte_code, value_num, value_text, unit, ref_low, ref_high, flag
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			res.ID, res.PanelID, res.AnalyteCode, res.ValueNum, res.ValueText,
			res.Unit, res.RefLow, res.RefHigh, string(res.Flag),
		); err != nil {
			return LabPanel{}, err
		}
		panel.Results = append(panel.Results, res)
	}
	if _, err := tx.Exec(ctx, `UPDATE labs.panels SET updated_at = NOW() WHERE id = $1`, panelID); err != nil {
		return LabPanel{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LabPanel{}, err
	}
	panel.AbnormalCount = abnormal
	panel.UpdatedAt = time.Now().UTC()
	return panel, nil
}

func (s *Store) DeleteLabPanel(ctx context.Context, panelID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM labs.panels WHERE id = $1`, panelID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListLabAnalyteTrend returns numeric results for one analyte across panels of a pet.
func (s *Store) ListLabAnalyteTrend(ctx context.Context, petID, analyteCode string) ([]map[string]any, error) {
	code := strings.TrimSpace(strings.ToLower(analyteCode))
	if !IsKnownAnalyte(code) {
		return nil, ErrUnknownAnalyte
	}
	rows, err := s.pool.Query(ctx, `
		SELECT r.value_num, r.unit, r.flag, r.ref_low, r.ref_high, p.collected_at, p.id::text
		FROM labs.panel_results r
		JOIN labs.panels p ON p.id = r.panel_id
		WHERE p.pet_id = $1 AND r.analyte_code = $2 AND r.value_num IS NOT NULL
		ORDER BY p.collected_at ASC`, petID, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var value float64
		var unit, flag, panelID string
		var refLow, refHigh *float64
		var at time.Time
		if err := rows.Scan(&value, &unit, &flag, &refLow, &refHigh, &at, &panelID); err != nil {
			return nil, err
		}
		item := map[string]any{
			"valueNum":    value,
			"unit":        unit,
			"flag":        flag,
			"collectedAt": at,
			"panelId":     panelID,
		}
		if refLow != nil {
			item["refLow"] = *refLow
		}
		if refHigh != nil {
			item["refHigh"] = *refHigh
		}
		out = append(out, item)
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out, rows.Err()
}
