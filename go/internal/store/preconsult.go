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
	// Informative AI assessment (does not override answers.urgency).
	AIUrgency    string     `json:"aiUrgency,omitempty"`
	AISummary    string     `json:"aiSummary,omitempty"`
	AIAssessedAt *time.Time `json:"aiAssessedAt,omitempty"`
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

// ContainsHTMLMarkup rejects angle brackets to keep answers plain text (anti-injection).
func ContainsHTMLMarkup(s string) bool {
	return strings.ContainsAny(s, "<>")
}

func ValidatePreconsultAnswers(a PreconsultAnswers) error {
	a.ChiefComplaint = strings.TrimSpace(a.ChiefComplaint)
	a.Comment = strings.TrimSpace(a.Comment)
	if a.ChiefComplaint == "" {
		return errors.New("chief_complaint_required")
	}
	if utf8.RuneCountInString(a.ChiefComplaint) > 500 {
		return errors.New("chief_complaint_too_long")
	}
	if a.Comment != "" && utf8.RuneCountInString(a.Comment) > 2000 {
		return errors.New("comment_too_long")
	}
	if ContainsHTMLMarkup(a.ChiefComplaint) || ContainsHTMLMarkup(a.Comment) {
		return errors.New("html_not_allowed")
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
		&in.AIUrgency, &in.AISummary, &in.AIAssessedAt,
	)
	if err != nil {
		return PreconsultIntake{}, err
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &in.Answers)
	}
	return in, nil
}

const preconsultCols = `id::text, visit_id::text, status, answers, submitted_at, created_at, updated_at,
		COALESCE(ai_urgency,''), COALESCE(ai_summary,''), ai_assessed_at`

const preconsultSelect = `
	SELECT ` + preconsultCols + `
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
		RETURNING `+preconsultCols,
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
		RETURNING `+preconsultCols,
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
	if after, ok := strings.CutPrefix(msg, "validation:"); ok {
		return after, true
	}
	return "", false
}

// PreconsultMeta is status (+ optional Pro alert) for calendar enrichment.
type PreconsultMeta struct {
	Status string
	Alert  string // "urgent" when submitted and AI red or declared high
}

// MapPreconsultMeta returns visitID → status/alert for a practice visit list.
func (s *Store) MapPreconsultMeta(ctx context.Context, visitIDs []string) (map[string]PreconsultMeta, error) {
	out := map[string]PreconsultMeta{}
	if len(visitIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT visit_id::text, status,
			COALESCE(ai_urgency, ''),
			COALESCE(answers->>'urgency', '')
		FROM visits.preconsult_intakes
		WHERE visit_id = ANY($1::uuid[])`, visitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, status, aiUrgency, declared string
		if err := rows.Scan(&id, &status, &aiUrgency, &declared); err != nil {
			return nil, err
		}
		meta := PreconsultMeta{Status: status}
		if status == PreconsultSubmitted && (aiUrgency == "red" || declared == "high") {
			meta.Alert = "urgent"
		}
		out[id] = meta
	}
	return out, rows.Err()
}

// MapPreconsultStatuses returns visitID → status for a practice visit list.
func (s *Store) MapPreconsultStatuses(ctx context.Context, visitIDs []string) (map[string]string, error) {
	meta, err := s.MapPreconsultMeta(ctx, visitIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(meta))
	for id, m := range meta {
		out[id] = m.Status
	}
	return out, nil
}

const maxPreconsultAISummaryRunes = 800

// SavePreconsultAIAssessment persists the informative Gemini urgency judgment.
// Does not bump updated_at (avoids re-queueing research ETL on AI-only writes).
func (s *Store) SavePreconsultAIAssessment(ctx context.Context, visitID, urgency, summary string) error {
	urgency = strings.ToLower(strings.TrimSpace(urgency))
	switch urgency {
	case "green", "orange", "red":
	default:
		return errors.New("invalid_ai_urgency")
	}
	summary = strings.TrimSpace(summary)
	if runes := []rune(summary); len(runes) > maxPreconsultAISummaryRunes {
		summary = string(runes[:maxPreconsultAISummaryRunes])
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE visits.preconsult_intakes
		SET ai_urgency = $2, ai_summary = $3, ai_assessed_at = NOW()
		WHERE visit_id = $1`, visitID, urgency, summary)
	return err
}

// AttachPreconsultStatuses fills Visit.PreconsultStatus / PreconsultAlert when an intake exists.
func (s *Store) AttachPreconsultStatuses(ctx context.Context, visits []Visit) error {
	if len(visits) == 0 {
		return nil
	}
	ids := make([]string, len(visits))
	for i, v := range visits {
		ids[i] = v.ID
	}
	m, err := s.MapPreconsultMeta(ctx, ids)
	if err != nil {
		return err
	}
	for i := range visits {
		if meta, ok := m[visits[i].ID]; ok {
			visits[i].PreconsultStatus = meta.Status
			visits[i].PreconsultAlert = meta.Alert
		}
	}
	return nil
}

const preconsultTokenTTL = 30 * 24 * time.Hour

// HasPreconsultInviteIssued reports whether a public invite token was already created for this visit.
func (s *Store) HasPreconsultInviteIssued(ctx context.Context, visitID string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.preconsult_tokens WHERE visit_id = $1`, visitID).Scan(&n)
	return n > 0, err
}

// IssuePreconsultToken creates an opaque public token for a visit (replaces unused active tokens).
func (s *Store) IssuePreconsultToken(ctx context.Context, visitID string) (string, error) {
	token := strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", "")
	id := uuid.NewString()
	expires := time.Now().UTC().Add(preconsultTokenTTL)
	_, err := s.pool.Exec(ctx, `
		DELETE FROM visits.preconsult_tokens WHERE visit_id = $1 AND used_at IS NULL`, visitID)
	if err != nil {
		return "", err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO visits.preconsult_tokens (id, visit_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`, id, visitID, token, expires)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ResolvePreconsultToken returns visitID for a valid (non-expired) token.
func (s *Store) ResolvePreconsultToken(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrNotFound
	}
	var visitID string
	err := s.pool.QueryRow(ctx, `
		SELECT visit_id::text FROM visits.preconsult_tokens
		WHERE token = $1 AND expires_at > NOW()`, token).Scan(&visitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return visitID, err
}

// MarkPreconsultTokenUsed stamps used_at when the public form is submitted.
func (s *Store) MarkPreconsultTokenUsed(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE visits.preconsult_tokens SET used_at = NOW()
		WHERE token = $1 AND used_at IS NULL`, token)
	return err
}

// PublicPreconsultContext is safe non-PII context for the public form.
type PublicPreconsultContext struct {
	VisitID      string             `json:"visitId"`
	PetName      string             `json:"petName"`
	PracticeName string             `json:"practiceName"`
	ScheduledAt  *time.Time         `json:"scheduledAt,omitempty"`
	Status       string             `json:"status"`
	Answers      *PreconsultAnswers `json:"answers,omitempty"`
	InviteURL    string             `json:"inviteUrl,omitempty"`
	DownloadURL  string             `json:"downloadUrl,omitempty"`
}

func (s *Store) GetPublicPreconsultContext(ctx context.Context, visitID string) (PublicPreconsultContext, error) {
	var out PublicPreconsultContext
	out.VisitID = visitID
	err := s.pool.QueryRow(ctx, `
		SELECT p.name, COALESCE(pr.name, ''), v.scheduled_at
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		LEFT JOIN practice.practices pr ON pr.id = v.practice_id
		WHERE v.id = $1`, visitID).Scan(&out.PetName, &out.PracticeName, &out.ScheduledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicPreconsultContext{}, ErrNotFound
	}
	if err != nil {
		return PublicPreconsultContext{}, err
	}
	in, err := s.GetPreconsultByVisit(ctx, visitID)
	if errors.Is(err, ErrNotFound) {
		out.Status = PreconsultPending
		return out, nil
	}
	if err != nil {
		return PublicPreconsultContext{}, err
	}
	out.Status = in.Status
	if in.Status == PreconsultSubmitted {
		ans := in.Answers
		out.Answers = &ans
	}
	return out, nil
}
