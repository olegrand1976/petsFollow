package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const PitchMaxCallDuration = 8 * time.Minute

var (
	ErrPitchScriptNotFound = errors.New("pitch_script_not_found")
	ErrPitchSimNotFound    = errors.New("pitch_sim_not_found")
	ErrPitchFeedbackExists = errors.New("pitch_feedback_exists")
	ErrPitchFeedbackLocked = errors.New("pitch_feedback_locked")
	ErrAgentPromptNotFound  = errors.New("agent_prompt_not_found")
	ErrPitchScriptForbidden = errors.New("pitch_script_forbidden")
)

type PitchScript struct {
	ID                  string          `json:"id"`
	Slug                string          `json:"slug"`
	Title               string          `json:"title"`
	Audience            string          `json:"audience"`
	OwnerUserID         *string         `json:"ownerUserId,omitempty"`
	ParentScriptID      *string         `json:"parentScriptId,omitempty"`
	StepsJSON           json.RawMessage `json:"steps"`
	ExampleDialogueJSON json.RawMessage `json:"exampleDialogue"`
	CoachHints          string          `json:"coachHints"`
	Locale              string          `json:"locale"`
	IsActive            bool            `json:"isActive"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type AgentPromptVersion struct {
	ID            string          `json:"id"`
	AgentKind     string          `json:"agentKind"`
	Version       int             `json:"version"`
	ContentJSON   json.RawMessage `json:"content"`
	Changelog     string          `json:"changelog"`
	Source        string          `json:"source"`
	CreatedBy     *string         `json:"createdBy,omitempty"`
	AnalyzerRunID *string         `json:"analyzerRunId,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	IsCurrent     bool            `json:"isCurrent,omitempty"`
}

type PitchSimulation struct {
	ID                   string          `json:"id"`
	UserID               string          `json:"userId"`
	ScriptID             string          `json:"scriptId"`
	InterestLevel        string          `json:"interestLevel"`
	VoiceName            string          `json:"voiceName"`
	VetPromptVersionID   *string         `json:"vetPromptVersionId,omitempty"`
	CoachPromptVersionID *string         `json:"coachPromptVersionId,omitempty"`
	Outcome              string          `json:"outcome"`
	AppointmentSlot      string          `json:"appointmentSlot,omitempty"`
	DurationSec          int             `json:"durationSec"`
	EndedAt              *time.Time      `json:"endedAt,omitempty"`
	TranscriptJSON       json.RawMessage `json:"transcript"`
	CoachFeedbackJSON    json.RawMessage `json:"coachFeedback,omitempty"`
	AIScore              *float64        `json:"aiScore,omitempty"`
	UserScore            *float64        `json:"userScore,omitempty"`
	AudioObjectKey       string          `json:"audioObjectKey,omitempty"`
	AudioURL             string          `json:"audioUrl,omitempty"`
	IsTop5               bool            `json:"isTop5"`
	FeedbackSkipped      bool            `json:"feedbackSkipped,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	HasFeedback          bool            `json:"hasFeedback,omitempty"`
}

type PitchSimFeedback struct {
	ID                   string          `json:"id"`
	SimulationID         string          `json:"simulationId"`
	UserID               string          `json:"userId"`
	VetRealism           int             `json:"vetRealism"`
	CoachUsefulness      int             `json:"coachUsefulness"`
	DifficultyFelt       string          `json:"difficultyFelt"`
	Comment              string          `json:"comment"`
	Flags                json.RawMessage `json:"flags"`
	AnalyzerProcessedAt  *time.Time      `json:"analyzerProcessedAt,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
}

type PitchAnalyzerRun struct {
	ID              string          `json:"id"`
	StartedAt       time.Time       `json:"startedAt"`
	FinishedAt      *time.Time      `json:"finishedAt,omitempty"`
	FeedbackCount   int             `json:"feedbackCount"`
	Status          string          `json:"status"`
	InputSummaryJSON json.RawMessage `json:"inputSummary"`
	OutputJSON      json.RawMessage `json:"output"`
	VetVersionID    *string         `json:"vetVersionId,omitempty"`
	CoachVersionID  *string         `json:"coachVersionId,omitempty"`
}

func scanPitchScript(row pgx.Row) (PitchScript, error) {
	var s PitchScript
	var owner, parent *string
	err := row.Scan(
		&s.ID, &s.Slug, &s.Title, &s.Audience, &owner, &parent,
		&s.StepsJSON, &s.ExampleDialogueJSON, &s.CoachHints, &s.Locale, &s.IsActive,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return s, err
	}
	s.OwnerUserID = owner
	s.ParentScriptID = parent
	return s, nil
}

const pitchScriptCols = `id, slug, title, audience, owner_user_id::text, parent_script_id::text,
	steps_json, example_dialogue_json, coach_hints, locale, is_active, created_at, updated_at`

func (s *Store) ListPitchScriptsForUser(ctx context.Context, userID string) ([]PitchScript, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+pitchScriptCols+`
		FROM sales.pitch_scripts
		WHERE is_active = true AND (owner_user_id IS NULL OR owner_user_id = $1)
		ORDER BY owner_user_id NULLS FIRST, title`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PitchScript, 0)
	for rows.Next() {
		sc, err := scanPitchScript(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Store) ListAdminPitchScripts(ctx context.Context) ([]PitchScript, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+pitchScriptCols+`
		FROM sales.pitch_scripts
		WHERE owner_user_id IS NULL
		ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PitchScript, 0)
	for rows.Next() {
		sc, err := scanPitchScript(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Store) GetPitchScript(ctx context.Context, id string) (PitchScript, error) {
	sc, err := scanPitchScript(s.pool.QueryRow(ctx, `
		SELECT `+pitchScriptCols+` FROM sales.pitch_scripts WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return PitchScript{}, ErrPitchScriptNotFound
	}
	return sc, err
}

// CanAccessPitchScript: admin defaults (owner NULL) or own personalization.
func (s *Store) CanAccessPitchScript(ctx context.Context, scriptID, userID string) error {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sales.pitch_scripts
			WHERE id=$1 AND is_active=true
			  AND (owner_user_id IS NULL OR owner_user_id=$2)
		)`, scriptID, userID).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPitchScriptForbidden
	}
	return nil
}

func (s *Store) PitchCallTimedOut(sim PitchSimulation) bool {
	return time.Since(sim.CreatedAt) >= PitchMaxCallDuration
}

func (s *Store) CreatePitchScript(ctx context.Context, sc PitchScript) (PitchScript, error) {
	if sc.ID == "" {
		sc.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	sc.CreatedAt = now
	sc.UpdatedAt = now
	if sc.StepsJSON == nil {
		sc.StepsJSON = json.RawMessage("[]")
	}
	if sc.ExampleDialogueJSON == nil {
		sc.ExampleDialogueJSON = json.RawMessage("[]")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sales.pitch_scripts (
			id, slug, title, audience, owner_user_id, parent_script_id,
			steps_json, example_dialogue_json, coach_hints, locale, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		sc.ID, sc.Slug, sc.Title, sc.Audience, sc.OwnerUserID, sc.ParentScriptID,
		sc.StepsJSON, sc.ExampleDialogueJSON, sc.CoachHints, sc.Locale, sc.IsActive, sc.CreatedAt, sc.UpdatedAt,
	)
	return sc, err
}

func (s *Store) UpdatePitchScript(ctx context.Context, sc PitchScript) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_scripts SET
			title=$2, steps_json=$3, example_dialogue_json=$4, coach_hints=$5,
			is_active=$6, updated_at=NOW()
		WHERE id=$1`,
		sc.ID, sc.Title, sc.StepsJSON, sc.ExampleDialogueJSON, sc.CoachHints, sc.IsActive,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrPitchScriptNotFound
	}
	return nil
}

func (s *Store) PersonalizePitchScript(ctx context.Context, parentID, userID string) (PitchScript, error) {
	parent, err := s.GetPitchScript(ctx, parentID)
	if err != nil {
		return PitchScript{}, err
	}
	if parent.OwnerUserID != nil {
		return PitchScript{}, fmt.Errorf("cannot_personalize_non_admin_script")
	}
	// Return existing copy if any
	var existingID string
	err = s.pool.QueryRow(ctx, `
		SELECT id::text FROM sales.pitch_scripts
		WHERE parent_script_id=$1 AND owner_user_id=$2 LIMIT 1`, parentID, userID).Scan(&existingID)
	if err == nil {
		return s.GetPitchScript(ctx, existingID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PitchScript{}, err
	}
	copy := parent
	copy.ID = uuid.NewString()
	copy.Slug = parent.Slug + "-" + userID[:8]
	uid := userID
	pid := parentID
	copy.OwnerUserID = &uid
	copy.ParentScriptID = &pid
	return s.CreatePitchScript(ctx, copy)
}

func (s *Store) DeletePitchScript(ctx context.Context, id, ownerUserID string) error {
	ct, err := s.pool.Exec(ctx, `
		DELETE FROM sales.pitch_scripts WHERE id=$1 AND owner_user_id=$2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrPitchScriptNotFound
	}
	return nil
}

func (s *Store) GetCurrentAgentPrompt(ctx context.Context, kind string) (AgentPromptVersion, error) {
	var v AgentPromptVersion
	err := s.pool.QueryRow(ctx, `
		SELECT v.id::text, v.agent_kind, v.version, v.content_json, v.changelog, v.source,
			v.created_by::text, v.analyzer_run_id::text, v.created_at
		FROM sales.agent_prompt_current c
		JOIN sales.agent_prompt_versions v ON v.id = c.version_id
		WHERE c.agent_kind=$1`, kind).Scan(
		&v.ID, &v.AgentKind, &v.Version, &v.ContentJSON, &v.Changelog, &v.Source,
		&v.CreatedBy, &v.AnalyzerRunID, &v.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrAgentPromptNotFound
	}
	v.IsCurrent = true
	return v, err
}

func (s *Store) ListAgentPromptVersions(ctx context.Context, kind string) ([]AgentPromptVersion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.agent_kind, v.version, v.content_json, v.changelog, v.source,
			v.created_by::text, v.analyzer_run_id::text, v.created_at,
			(c.version_id IS NOT NULL) AS is_current
		FROM sales.agent_prompt_versions v
		LEFT JOIN sales.agent_prompt_current c ON c.version_id = v.id AND c.agent_kind = v.agent_kind
		WHERE v.agent_kind=$1
		ORDER BY v.version DESC`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AgentPromptVersion
	for rows.Next() {
		var v AgentPromptVersion
		if err := rows.Scan(
			&v.ID, &v.AgentKind, &v.Version, &v.ContentJSON, &v.Changelog, &v.Source,
			&v.CreatedBy, &v.AnalyzerRunID, &v.CreatedAt, &v.IsCurrent,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GetAgentPromptVersion(ctx context.Context, id string) (AgentPromptVersion, error) {
	var v AgentPromptVersion
	err := s.pool.QueryRow(ctx, `
		SELECT v.id::text, v.agent_kind, v.version, v.content_json, v.changelog, v.source,
			v.created_by::text, v.analyzer_run_id::text, v.created_at,
			(c.version_id IS NOT NULL)
		FROM sales.agent_prompt_versions v
		LEFT JOIN sales.agent_prompt_current c ON c.version_id = v.id
		WHERE v.id=$1`, id).Scan(
		&v.ID, &v.AgentKind, &v.Version, &v.ContentJSON, &v.Changelog, &v.Source,
		&v.CreatedBy, &v.AnalyzerRunID, &v.CreatedAt, &v.IsCurrent,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrAgentPromptNotFound
	}
	return v, err
}

func (s *Store) CreateAgentPromptVersion(ctx context.Context, kind string, content json.RawMessage, changelog, source string, createdBy *string, analyzerRunID *string, setCurrent bool) (AgentPromptVersion, error) {
	var next int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1 FROM sales.agent_prompt_versions WHERE agent_kind=$1`, kind).Scan(&next)
	if err != nil {
		return AgentPromptVersion{}, err
	}
	v := AgentPromptVersion{
		ID:            uuid.NewString(),
		AgentKind:     kind,
		Version:       next,
		ContentJSON:   content,
		Changelog:     changelog,
		Source:        source,
		CreatedBy:     createdBy,
		AnalyzerRunID: analyzerRunID,
		CreatedAt:     time.Now().UTC(),
		IsCurrent:     setCurrent,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return v, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO sales.agent_prompt_versions (
			id, agent_kind, version, content_json, changelog, source, created_by, analyzer_run_id, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		v.ID, v.AgentKind, v.Version, v.ContentJSON, v.Changelog, v.Source, v.CreatedBy, v.AnalyzerRunID, v.CreatedAt,
	)
	if err != nil {
		return v, err
	}
	if setCurrent {
		_, err = tx.Exec(ctx, `
			INSERT INTO sales.agent_prompt_current (agent_kind, version_id)
			VALUES ($1,$2)
			ON CONFLICT (agent_kind) DO UPDATE SET version_id=EXCLUDED.version_id`, kind, v.ID)
		if err != nil {
			return v, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return v, err
	}
	return v, nil
}

func (s *Store) RestoreAgentPromptVersion(ctx context.Context, versionID, adminUserID string) (AgentPromptVersion, error) {
	src, err := s.GetAgentPromptVersion(ctx, versionID)
	if err != nil {
		return AgentPromptVersion{}, err
	}
	by := adminUserID
	return s.CreateAgentPromptVersion(ctx, src.AgentKind, src.ContentJSON,
		fmt.Sprintf("Rollback vers v%d", src.Version), "admin", &by, nil, true)
}

func (s *Store) CreatePitchSimulation(ctx context.Context, sim PitchSimulation) (PitchSimulation, error) {
	if sim.ID == "" {
		sim.ID = uuid.NewString()
	}
	if sim.TranscriptJSON == nil {
		sim.TranscriptJSON = json.RawMessage("[]")
	}
	if sim.Outcome == "" {
		sim.Outcome = "in_progress"
	}
	sim.CreatedAt = time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sales.pitch_simulations (
			id, user_id, script_id, interest_level, voice_name,
			vet_prompt_version_id, coach_prompt_version_id, outcome,
			appointment_slot, duration_sec, ended_at, transcript_json,
			coach_feedback_json, ai_score, user_score, audio_object_key, is_top5, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		sim.ID, sim.UserID, sim.ScriptID, sim.InterestLevel, sim.VoiceName,
		sim.VetPromptVersionID, sim.CoachPromptVersionID, sim.Outcome,
		sim.AppointmentSlot, sim.DurationSec, sim.EndedAt, sim.TranscriptJSON,
		nilIfEmptyJSON(sim.CoachFeedbackJSON), sim.AIScore, sim.UserScore, sim.AudioObjectKey, sim.IsTop5, sim.CreatedAt,
	)
	return sim, err
}

func nilIfEmptyJSON(r json.RawMessage) any {
	if len(r) == 0 || string(r) == "null" {
		return nil
	}
	return []byte(r)
}

func (s *Store) FinalizePitchSimulation(ctx context.Context, sim PitchSimulation) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_simulations SET
			outcome=$2,
			appointment_slot=COALESCE(NULLIF($3,''), appointment_slot),
			duration_sec=CASE WHEN $4 > 0 THEN $4 ELSE duration_sec END,
			ended_at=COALESCE($5, ended_at),
			transcript_json=CASE WHEN $6::jsonb = '[]'::jsonb AND transcript_json <> '[]'::jsonb THEN transcript_json ELSE $6::jsonb END,
			coach_feedback_json=COALESCE($7, coach_feedback_json),
			ai_score=COALESCE($8, ai_score),
			audio_object_key=COALESCE(NULLIF($9,''), audio_object_key),
			coach_prompt_version_id=COALESCE($10, coach_prompt_version_id)
		WHERE id=$1 AND user_id=$11`,
		sim.ID, sim.Outcome, sim.AppointmentSlot, sim.DurationSec, sim.EndedAt,
		sim.TranscriptJSON, nilIfEmptyJSON(sim.CoachFeedbackJSON), sim.AIScore,
		sim.AudioObjectKey, sim.CoachPromptVersionID, sim.UserID,
	)
	return err
}

func (s *Store) SetPitchSimulationAudio(ctx context.Context, id, userID, objectKey string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_simulations SET audio_object_key=$3
		WHERE id=$1 AND user_id=$2`, id, userID, objectKey)
	return err
}

func (s *Store) UpdatePitchSimulationTranscript(ctx context.Context, id, userID string, transcript json.RawMessage) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_simulations SET transcript_json=$3
		WHERE id=$1 AND user_id=$2`, id, userID, transcript)
	return err
}

func (s *Store) SetPitchSimulationUserScore(ctx context.Context, id, userID string, score float64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_simulations SET user_score=$3 WHERE id=$1 AND user_id=$2`,
		id, userID, score)
	if err != nil {
		return err
	}
	return s.RecalcPitchTop5(ctx, userID)
}

func (s *Store) RecalcPitchTop5(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `UPDATE sales.pitch_simulations SET is_top5=false WHERE user_id=$1`, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE sales.pitch_simulations SET is_top5=true
		WHERE id IN (
			SELECT id FROM sales.pitch_simulations
			WHERE user_id=$1 AND outcome <> 'in_progress'
			ORDER BY COALESCE(user_score, ai_score, 0) DESC, created_at DESC
			LIMIT 5
		)`, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) GetPitchSimulation(ctx context.Context, id, userID string) (PitchSimulation, error) {
	sim, err := s.scanPitchSim(s.pool.QueryRow(ctx, `
		SELECT s.id::text, s.user_id::text, s.script_id::text, s.interest_level, s.voice_name,
			s.vet_prompt_version_id::text, s.coach_prompt_version_id::text, s.outcome, s.appointment_slot,
			s.duration_sec, s.ended_at, s.transcript_json, s.coach_feedback_json,
			s.ai_score, s.user_score, s.audio_object_key, s.is_top5, s.feedback_skipped, s.created_at,
			EXISTS(SELECT 1 FROM sales.pitch_sim_feedback f WHERE f.simulation_id=s.id)
		FROM sales.pitch_simulations s
		WHERE s.id=$1 AND s.user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PitchSimulation{}, ErrPitchSimNotFound
	}
	return sim, err
}

func (s *Store) scanPitchSim(row pgx.Row) (PitchSimulation, error) {
	var sim PitchSimulation
	var vetVID, coachVID *string
	var ended *time.Time
	var coachFB []byte
	err := row.Scan(
		&sim.ID, &sim.UserID, &sim.ScriptID, &sim.InterestLevel, &sim.VoiceName,
		&vetVID, &coachVID, &sim.Outcome, &sim.AppointmentSlot,
		&sim.DurationSec, &ended, &sim.TranscriptJSON, &coachFB,
		&sim.AIScore, &sim.UserScore, &sim.AudioObjectKey, &sim.IsTop5, &sim.FeedbackSkipped, &sim.CreatedAt,
		&sim.HasFeedback,
	)
	if err != nil {
		return sim, err
	}
	sim.VetPromptVersionID = vetVID
	sim.CoachPromptVersionID = coachVID
	sim.EndedAt = ended
	if len(coachFB) > 0 {
		sim.CoachFeedbackJSON = coachFB
	}
	return sim, nil
}

func (s *Store) ListPitchSimulations(ctx context.Context, userID string, includeExpiredAudio bool) ([]PitchSimulation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.user_id::text, s.script_id::text, s.interest_level, s.voice_name,
			s.vet_prompt_version_id::text, s.coach_prompt_version_id::text, s.outcome, s.appointment_slot,
			s.duration_sec, s.ended_at, s.transcript_json, s.coach_feedback_json,
			s.ai_score, s.user_score, s.audio_object_key, s.is_top5, s.feedback_skipped, s.created_at,
			EXISTS(SELECT 1 FROM sales.pitch_sim_feedback f WHERE f.simulation_id=s.id)
		FROM sales.pitch_simulations s
		WHERE s.user_id=$1
		  AND s.outcome <> 'in_progress'
		  AND (s.is_top5 = true OR s.created_at > NOW() - INTERVAL '30 days' OR $2)
		ORDER BY s.is_top5 DESC, COALESCE(s.user_score, s.ai_score, 0) DESC, s.created_at DESC
		LIMIT 100`, userID, includeExpiredAudio)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PitchSimulation
	for rows.Next() {
		sim, err := s.scanPitchSim(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sim)
	}
	return out, rows.Err()
}

func (s *Store) ListTeamPitchSimulations(ctx context.Context, managerID string) ([]PitchSimulation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.user_id::text, s.script_id::text, s.interest_level, s.voice_name,
			s.vet_prompt_version_id::text, s.coach_prompt_version_id::text, s.outcome, s.appointment_slot,
			s.duration_sec, s.ended_at, s.transcript_json, s.coach_feedback_json,
			s.ai_score, s.user_score, s.audio_object_key, s.is_top5, s.feedback_skipped, s.created_at,
			EXISTS(SELECT 1 FROM sales.pitch_sim_feedback f WHERE f.simulation_id=s.id)
		FROM sales.pitch_simulations s
		JOIN identity.users u ON u.id = s.user_id
		WHERE (u.manager_user_id=$1 OR s.user_id=$1)
		  AND s.outcome <> 'in_progress'
		ORDER BY s.created_at DESC
		LIMIT 200`, managerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PitchSimulation
	for rows.Next() {
		sim, err := s.scanPitchSim(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sim)
	}
	return out, rows.Err()
}

func (s *Store) UpsertPitchSimFeedback(ctx context.Context, fb PitchSimFeedback) (PitchSimFeedback, error) {
	var existingID string
	var updatedAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, updated_at FROM sales.pitch_sim_feedback WHERE simulation_id=$1`,
		fb.SimulationID).Scan(&existingID, &updatedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fb, err
	}
	if existingID != "" {
		if time.Since(updatedAt) > 24*time.Hour {
			return fb, ErrPitchFeedbackLocked
		}
		fb.ID = existingID
		_, err = s.pool.Exec(ctx, `
			UPDATE sales.pitch_sim_feedback SET
				vet_realism=$2, coach_usefulness=$3, difficulty_felt=$4,
				comment=$5, flags=$6, updated_at=NOW()
			WHERE id=$1`,
			fb.ID, fb.VetRealism, fb.CoachUsefulness, fb.DifficultyFelt, fb.Comment, fb.Flags,
		)
		fb.UpdatedAt = time.Now().UTC()
		return fb, err
	}
	fb.ID = uuid.NewString()
	fb.CreatedAt = time.Now().UTC()
	fb.UpdatedAt = fb.CreatedAt
	if fb.Flags == nil {
		fb.Flags = json.RawMessage("[]")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO sales.pitch_sim_feedback (
			id, simulation_id, user_id, vet_realism, coach_usefulness,
			difficulty_felt, comment, flags, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		fb.ID, fb.SimulationID, fb.UserID, fb.VetRealism, fb.CoachUsefulness,
		fb.DifficultyFelt, fb.Comment, fb.Flags, fb.CreatedAt, fb.UpdatedAt,
	)
	return fb, err
}

func (s *Store) ListUnprocessedPitchFeedback(ctx context.Context, limit int) ([]PitchSimFeedback, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, simulation_id::text, user_id::text, vet_realism, coach_usefulness,
			difficulty_felt, comment, flags, analyzer_processed_at, created_at, updated_at
		FROM sales.pitch_sim_feedback
		WHERE analyzer_processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PitchSimFeedback
	for rows.Next() {
		var fb PitchSimFeedback
		if err := rows.Scan(
			&fb.ID, &fb.SimulationID, &fb.UserID, &fb.VetRealism, &fb.CoachUsefulness,
			&fb.DifficultyFelt, &fb.Comment, &fb.Flags, &fb.AnalyzerProcessedAt, &fb.CreatedAt, &fb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, fb)
	}
	return out, rows.Err()
}

func (s *Store) MarkPitchFeedbackProcessed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_sim_feedback SET analyzer_processed_at=NOW()
		WHERE id = ANY($1::uuid[])`, ids)
	return err
}

func (s *Store) CreateAnalyzerRun(ctx context.Context, run PitchAnalyzerRun) (PitchAnalyzerRun, error) {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	run.StartedAt = time.Now().UTC()
	if run.InputSummaryJSON == nil {
		run.InputSummaryJSON = json.RawMessage("{}")
	}
	if run.OutputJSON == nil {
		run.OutputJSON = json.RawMessage("{}")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sales.pitch_analyzer_runs (
			id, started_at, finished_at, feedback_count, status,
			input_summary_json, output_json, vet_version_id, coach_version_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		run.ID, run.StartedAt, run.FinishedAt, run.FeedbackCount, run.Status,
		run.InputSummaryJSON, run.OutputJSON, run.VetVersionID, run.CoachVersionID,
	)
	return run, err
}

func (s *Store) FinishAnalyzerRun(ctx context.Context, run PitchAnalyzerRun) error {
	now := time.Now().UTC()
	run.FinishedAt = &now
	_, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_analyzer_runs SET
			finished_at=$2, feedback_count=$3, status=$4,
			input_summary_json=$5, output_json=$6,
			vet_version_id=$7, coach_version_id=$8
		WHERE id=$1`,
		run.ID, run.FinishedAt, run.FeedbackCount, run.Status,
		run.InputSummaryJSON, run.OutputJSON, run.VetVersionID, run.CoachVersionID,
	)
	return err
}

func (s *Store) ListAnalyzerRuns(ctx context.Context, limit int) ([]PitchAnalyzerRun, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, started_at, finished_at, feedback_count, status,
			input_summary_json, output_json, vet_version_id::text, coach_version_id::text
		FROM sales.pitch_analyzer_runs
		ORDER BY started_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PitchAnalyzerRun
	for rows.Next() {
		var r PitchAnalyzerRun
		if err := rows.Scan(
			&r.ID, &r.StartedAt, &r.FinishedAt, &r.FeedbackCount, &r.Status,
			&r.InputSummaryJSON, &r.OutputJSON, &r.VetVersionID, &r.CoachVersionID,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListRecentPitchFeedback(ctx context.Context, limit int) ([]PitchSimFeedback, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, simulation_id::text, user_id::text, vet_realism, coach_usefulness,
			difficulty_felt, comment, flags, analyzer_processed_at, created_at, updated_at
		FROM sales.pitch_sim_feedback
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PitchSimFeedback
	for rows.Next() {
		var fb PitchSimFeedback
		if err := rows.Scan(
			&fb.ID, &fb.SimulationID, &fb.UserID, &fb.VetRealism, &fb.CoachUsefulness,
			&fb.DifficultyFelt, &fb.Comment, &fb.Flags, &fb.AnalyzerProcessedAt, &fb.CreatedAt, &fb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, fb)
	}
	return out, rows.Err()
}

func (s *Store) CountFeedbackSkipsToday(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sales.pitch_simulations s
		WHERE s.user_id=$1
		  AND s.feedback_skipped = true
		  AND s.ended_at IS NOT NULL
		  AND s.ended_at::date = (NOW() AT TIME ZONE 'UTC')::date`, userID).Scan(&n)
	return n, err
}

func (s *Store) MarkPitchFeedbackSkipped(ctx context.Context, simID, userID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.pitch_simulations SET feedback_skipped=true
		WHERE id=$1 AND user_id=$2 AND outcome <> 'in_progress'`, simID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrPitchSimNotFound
	}
	return nil
}
