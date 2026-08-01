package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// VisitReportExplanation is a cached owner-facing CR vulgarization.
type VisitReportExplanation struct {
	ID          string          `json:"id"`
	VisitID     string          `json:"visitId"`
	Locale      string          `json:"locale"`
	SourceHash  string          `json:"sourceHash"`
	PayloadJSON json.RawMessage `json:"payload"`
	Model       string          `json:"model"`
	RefreshedAt *time.Time      `json:"refreshedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

// GetVisitReportExplanation returns a cached explanation or ErrNotFound.
func (s *Store) GetVisitReportExplanation(ctx context.Context, visitID, locale string) (VisitReportExplanation, error) {
	var e VisitReportExplanation
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, locale, source_hash, payload_json, COALESCE(model,''),
			refreshed_at, created_at
		FROM visits.visit_report_explanations
		WHERE visit_id = $1::uuid AND locale = $2`, visitID, locale,
	).Scan(&e.ID, &e.VisitID, &e.Locale, &e.SourceHash, &e.PayloadJSON, &e.Model, &e.RefreshedAt, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return VisitReportExplanation{}, ErrNotFound
	}
	return e, err
}

// UpsertVisitReportExplanation stores or replaces the cached explanation for visit+locale.
func (s *Store) UpsertVisitReportExplanation(ctx context.Context, visitID, locale, sourceHash, model string, payload json.RawMessage, markRefresh bool) (VisitReportExplanation, error) {
	id := uuid.NewString()
	var e VisitReportExplanation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO visits.visit_report_explanations (id, visit_id, locale, source_hash, payload_json, model, refreshed_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5::jsonb, $6, CASE WHEN $7 THEN NOW() ELSE NULL END)
		ON CONFLICT (visit_id, locale) DO UPDATE SET
			source_hash = EXCLUDED.source_hash,
			payload_json = EXCLUDED.payload_json,
			model = EXCLUDED.model,
			refreshed_at = CASE WHEN $7 THEN NOW() ELSE visits.visit_report_explanations.refreshed_at END
		RETURNING id::text, visit_id::text, locale, source_hash, payload_json, COALESCE(model,''), refreshed_at, created_at`,
		id, visitID, locale, sourceHash, string(payload), model, markRefresh,
	).Scan(&e.ID, &e.VisitID, &e.Locale, &e.SourceHash, &e.PayloadJSON, &e.Model, &e.RefreshedAt, &e.CreatedAt)
	return e, err
}

// ClientAITriageSession is one triage conversation.
type ClientAITriageSession struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	PetID      string    `json:"petId,omitempty"`
	PracticeID string    `json:"practiceId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ClientAITriageMessage is one turn in a triage session.
type ClientAITriageMessage struct {
	ID                string          `json:"id"`
	SessionID         string          `json:"sessionId"`
	Role              string          `json:"role"`
	Body              string          `json:"body"`
	Level             string          `json:"level,omitempty"`
	WatchSigns        json.RawMessage `json:"watchSigns,omitempty"`
	RecommendedAction string          `json:"recommendedAction,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
}

// CreateClientAITriageSession opens a new triage session for a client.
func (s *Store) CreateClientAITriageSession(ctx context.Context, userID, petID, practiceID string) (ClientAITriageSession, error) {
	id := uuid.NewString()
	var sess ClientAITriageSession
	var petArg, practiceArg any
	if petID != "" {
		petArg = petID
	}
	if practiceID != "" {
		practiceArg = practiceID
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO client_ai.triage_sessions (id, user_id, pet_id, practice_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		RETURNING id::text, user_id::text, COALESCE(pet_id::text,''), COALESCE(practice_id::text,''), created_at, updated_at`,
		id, userID, petArg, practiceArg,
	).Scan(&sess.ID, &sess.UserID, &sess.PetID, &sess.PracticeID, &sess.CreatedAt, &sess.UpdatedAt)
	return sess, err
}

// GetClientAITriageSession loads a session owned by userID.
func (s *Store) GetClientAITriageSession(ctx context.Context, sessionID, userID string) (ClientAITriageSession, error) {
	var sess ClientAITriageSession
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, COALESCE(pet_id::text,''), COALESCE(practice_id::text,''), created_at, updated_at
		FROM client_ai.triage_sessions
		WHERE id = $1::uuid AND user_id = $2::uuid`, sessionID, userID,
	).Scan(&sess.ID, &sess.UserID, &sess.PetID, &sess.PracticeID, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientAITriageSession{}, ErrNotFound
	}
	return sess, err
}

// ListClientAITriageMessages returns messages oldest-first.
func (s *Store) ListClientAITriageMessages(ctx context.Context, sessionID string) ([]ClientAITriageMessage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, session_id::text, role, body, COALESCE(level,''), watch_signs,
			COALESCE(recommended_action,''), created_at
		FROM client_ai.triage_messages
		WHERE session_id = $1::uuid
		ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClientAITriageMessage
	for rows.Next() {
		var m ClientAITriageMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Body, &m.Level, &m.WatchSigns, &m.RecommendedAction, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []ClientAITriageMessage{}
	}
	return out, rows.Err()
}

// CountClientAITriageMessages returns the number of messages in a session.
func (s *Store) CountClientAITriageMessages(ctx context.Context, sessionID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM client_ai.triage_messages WHERE session_id = $1::uuid`, sessionID).Scan(&n)
	return n, err
}

func insertTriageMessageTx(ctx context.Context, tx pgx.Tx, sessionID, role, body, level string, watchSigns []string, recommendedAction string) (ClientAITriageMessage, error) {
	id := uuid.NewString()
	signsJSON := []byte("[]")
	if watchSigns != nil {
		b, err := json.Marshal(watchSigns)
		if err != nil {
			return ClientAITriageMessage{}, err
		}
		signsJSON = b
	}
	var levelNS *string
	if level != "" {
		levelNS = &level
	}
	var m ClientAITriageMessage
	err := tx.QueryRow(ctx, `
		INSERT INTO client_ai.triage_messages (id, session_id, role, body, level, watch_signs, recommended_action)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::jsonb, $7)
		RETURNING id::text, session_id::text, role, body, COALESCE(level,''), watch_signs,
			COALESCE(recommended_action,''), created_at`,
		id, sessionID, role, body, levelNS, string(signsJSON), recommendedAction,
	).Scan(&m.ID, &m.SessionID, &m.Role, &m.Body, &m.Level, &m.WatchSigns, &m.RecommendedAction, &m.CreatedAt)
	return m, err
}

// InsertClientAITriageTurn persists user + assistant messages atomically (no orphan user row on AI failure).
func (s *Store) InsertClientAITriageTurn(ctx context.Context, sessionID, userBody, assistantBody, level string, watchSigns []string, recommendedAction string) (userMsg, asstMsg ClientAITriageMessage, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClientAITriageMessage{}, ClientAITriageMessage{}, err
	}
	defer tx.Rollback(ctx)

	userMsg, err = insertTriageMessageTx(ctx, tx, sessionID, "user", userBody, "", nil, "")
	if err != nil {
		return ClientAITriageMessage{}, ClientAITriageMessage{}, err
	}
	asstMsg, err = insertTriageMessageTx(ctx, tx, sessionID, "assistant", assistantBody, level, watchSigns, recommendedAction)
	if err != nil {
		return ClientAITriageMessage{}, ClientAITriageMessage{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE client_ai.triage_sessions SET updated_at = NOW() WHERE id = $1::uuid`, sessionID); err != nil {
		return ClientAITriageMessage{}, ClientAITriageMessage{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ClientAITriageMessage{}, ClientAITriageMessage{}, err
	}
	return userMsg, asstMsg, nil
}

// InsertClientAIUsageEvent records an aggregated usage counter (no PHI body).
func (s *Store) InsertClientAIUsageEvent(ctx context.Context, userID, kind string, meta map[string]any) {
	if meta == nil {
		meta = map[string]any{}
	}
	b, _ := json.Marshal(meta)
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO client_ai.usage_events (id, user_id, kind, meta)
		VALUES ($1::uuid, $2::uuid, $3, $4::jsonb)`,
		uuid.NewString(), userID, kind, string(b))
}
