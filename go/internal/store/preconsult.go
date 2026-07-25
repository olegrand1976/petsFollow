package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	PreconsultPending   = "pending"
	PreconsultSubmitted = "submitted"
	PreconsultSkipped   = "skipped"
)

// PreconsultAnswers is the structured client intake before a confirmed visit.
type PreconsultAnswers struct {
	ChiefComplaint string `json:"chiefComplaint"`
	Duration       string `json:"duration"`
	Behavior       string `json:"behavior"`
	Appetite       string `json:"appetite"`
	Thirst         string `json:"thirst"`
	Elimination    string `json:"elimination"`
	Urgency        string `json:"urgency"`
	Comment        string `json:"comment,omitempty"`
}

type PreconsultIntake struct {
	ID          string            `json:"id"`
	VisitID     string            `json:"visitId"`
	Status      string            `json:"status"`
	Answers     PreconsultAnswers `json:"answers"`
	SubmittedAt *time.Time        `json:"submittedAt,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	// Optional denormalized fields for client UI.
	PetID   string `json:"petId,omitempty"`
	PetName string `json:"petName,omitempty"`
}

var (
	validPreconsultDuration = map[string]bool{
		"today": true, "few_days": true, "week": true, "weeks": true, "months": true, "unknown": true,
	}
	validPreconsultBehavior = map[string]bool{
		"normal": true, "lethargic": true, "restless": true, "aggressive": true,
		"anxious": true, "other": true, "unknown": true,
	}
	validPreconsultScale = map[string]bool{
		"normal": true, "decreased": true, "increased": true, "unknown": true,
	}
	validPreconsultUrgency = map[string]bool{
		"low": true, "medium": true, "high": true,
	}
)

func ValidatePreconsultAnswers(a PreconsultAnswers) error {
	a.ChiefComplaint = strings.TrimSpace(a.ChiefComplaint)
	if a.ChiefComplaint == "" {
		return errors.New("chief_complaint_required")
	}
	if utf8.RuneCountInString(a.ChiefComplaint) > 500 {
		return errors.New("chief_complaint_too_long")
	}
	if a.Comment != "" && utf8.RuneCountInString(a.Comment) > 2000 {
		return errors.New("comment_too_long")
	}
	if !validPreconsultDuration[a.Duration] {
		return errors.New("invalid_duration")
	}
	if !validPreconsultBehavior[a.Behavior] {
		return errors.New("invalid_behavior")
	}
	if !validPreconsultScale[a.Appetite] {
		return errors.New("invalid_appetite")
	}
	if !validPreconsultScale[a.Thirst] {
		return errors.New("invalid_thirst")
	}
	if !validPreconsultScale[a.Elimination] {
		return errors.New("invalid_elimination")
	}
	if !validPreconsultUrgency[a.Urgency] {
		return errors.New("invalid_urgency")
	}
	return nil
}

func scanPreconsult(row pgx.Row) (PreconsultIntake, error) {
	var in PreconsultIntake
	var raw []byte
	err := row.Scan(
		&in.ID, &in.VisitID, &in.Status, &raw, &in.SubmittedAt, &in.CreatedAt, &in.UpdatedAt,
	)
	if err != nil {
		return PreconsultIntake{}, err
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &in.Answers)
	}
	return in, nil
}

const preconsultSelect = `
	SELECT id::text, visit_id::text, status, answers, submitted_at, created_at, updated_at
	FROM visits.preconsult_intakes`

// EnsurePreconsultPending creates a pending intake for a confirmed visit.
// Returns created=true when a new row was inserted.
func (s *Store) EnsurePreconsultPending(ctx context.Context, visitID string) (PreconsultIntake, bool, error) {
	existing, gerr := s.GetPreconsultByVisit(ctx, visitID)
	if gerr == nil {
		return existing, false, nil
	}
	if !errors.Is(gerr, ErrNotFound) {
		return PreconsultIntake{}, false, gerr
	}
	id := uuid.NewString()
	in, err := scanPreconsult(s.pool.QueryRow(ctx, `
		INSERT INTO visits.preconsult_intakes (id, visit_id, status, answers)
		VALUES ($1, $2, 'pending', '{}'::jsonb)
		RETURNING id::text, visit_id::text, status, answers, submitted_at, created_at, updated_at`,
		id, visitID))
	if err != nil {
		// Race: another writer won.
		existing, gerr = s.GetPreconsultByVisit(ctx, visitID)
		if gerr == nil {
			return existing, false, nil
		}
		return PreconsultIntake{}, false, err
	}
	return in, true, nil
}

func (s *Store) GetPreconsultByVisit(ctx context.Context, visitID string) (PreconsultIntake, error) {
	in, err := scanPreconsult(s.pool.QueryRow(ctx, preconsultSelect+` WHERE visit_id = $1`, visitID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PreconsultIntake{}, ErrNotFound
	}
	return in, err
}

func (s *Store) SubmitPreconsult(ctx context.Context, visitID string, answers PreconsultAnswers) (PreconsultIntake, error) {
	if err := ValidatePreconsultAnswers(answers); err != nil {
		return PreconsultIntake{}, fmtValidation(err.Error())
	}
	answers.ChiefComplaint = strings.TrimSpace(answers.ChiefComplaint)
	answers.Comment = strings.TrimSpace(answers.Comment)
	raw, err := json.Marshal(answers)
	if err != nil {
		return PreconsultIntake{}, err
	}
	in, err := scanPreconsult(s.pool.QueryRow(ctx, `
		UPDATE visits.preconsult_intakes
		SET status = 'submitted', answers = $2::jsonb, submitted_at = NOW(), updated_at = NOW()
		WHERE visit_id = $1 AND status = 'pending'
		RETURNING id::text, visit_id::text, status, answers, submitted_at, created_at, updated_at`,
		visitID, raw))
	if errors.Is(err, pgx.ErrNoRows) {
		cur, gerr := s.GetPreconsultByVisit(ctx, visitID)
		if gerr != nil {
			return PreconsultIntake{}, gerr
		}
		if cur.Status == PreconsultSubmitted {
			return PreconsultIntake{}, ErrConflict
		}
		return PreconsultIntake{}, ErrNotFound
	}
	return in, err
}

func fmtValidation(code string) error {
	return errors.New("validation:" + code)
}

func IsPreconsultValidation(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "validation:") {
		return strings.TrimPrefix(msg, "validation:"), true
	}
	return "", false
}

// MapPreconsultStatuses returns visitID → status for a practice visit list.
func (s *Store) MapPreconsultStatuses(ctx context.Context, visitIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(visitIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT visit_id::text, status FROM visits.preconsult_intakes
		WHERE visit_id = ANY($1::uuid[])`, visitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			return nil, err
		}
		out[id] = status
	}
	return out, rows.Err()
}

// AttachPreconsultStatuses fills Visit.PreconsultStatus when an intake exists.
func (s *Store) AttachPreconsultStatuses(ctx context.Context, visits []Visit) error {
	if len(visits) == 0 {
		return nil
	}
	ids := make([]string, len(visits))
	for i, v := range visits {
		ids[i] = v.ID
	}
	m, err := s.MapPreconsultStatuses(ctx, ids)
	if err != nil {
		return err
	}
	for i := range visits {
		if st, ok := m[visits[i].ID]; ok {
			visits[i].PreconsultStatus = st
		}
	}
	return nil
}
