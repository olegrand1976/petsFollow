package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ImproveRunStatus string

const (
	ImproveRunQueued    ImproveRunStatus = "queued"
	ImproveRunRunning   ImproveRunStatus = "running"
	ImproveRunCompleted ImproveRunStatus = "completed"
	ImproveRunFailed    ImproveRunStatus = "failed"
	ImproveRunCancelled ImproveRunStatus = "cancelled"
)

type ImproveRun struct {
	ID          string           `json:"id"`
	VisitID     string           `json:"visitId"`
	ReportID    string           `json:"reportId"`
	PracticeID  string           `json:"practiceId"`
	UserID      string           `json:"userId"`
	Status      ImproveRunStatus `json:"status"`
	CrewTaskID  string           `json:"crewTaskId,omitempty"`
	Steps       json.RawMessage  `json:"steps"`
	Citations   json.RawMessage  `json:"citations"`
	Metrics     json.RawMessage  `json:"metrics,omitempty"`
	ErrorCode   string           `json:"errorCode,omitempty"`
	LatencyMs   int              `json:"latencyMs"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
}

const improveRunSelect = `
		SELECT id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		       status::text, crew_task_id, steps, citations, COALESCE(metrics, '{}'::jsonb), error_code, latency_ms,
		       created_at, updated_at, completed_at`

func (s *Store) CreateImproveRun(ctx context.Context, visitID, reportID, practiceID, userID string) (ImproveRun, error) {
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO rag.improve_runs (id, visit_id, report_id, practice_id, user_id, status)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, 'queued')
		RETURNING id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		          status::text, crew_task_id, steps, citations, COALESCE(metrics, '{}'::jsonb), error_code, latency_ms,
		          created_at, updated_at, completed_at`,
		id, visitID, reportID, practiceID, userID,
	)
	return scanImproveRun(row)
}

func (s *Store) GetImproveRun(ctx context.Context, runID string) (ImproveRun, error) {
	row := s.pool.QueryRow(ctx, improveRunSelect+`
		FROM rag.improve_runs WHERE id = $1::uuid`, runID)
	return scanImproveRun(row)
}

func (s *Store) ActiveImproveRunForVisit(ctx context.Context, visitID string) (ImproveRun, error) {
	row := s.pool.QueryRow(ctx, improveRunSelect+`
		FROM rag.improve_runs
		WHERE visit_id = $1::uuid AND status IN ('queued', 'running')
		ORDER BY created_at DESC LIMIT 1`, visitID)
	return scanImproveRun(row)
}

func (s *Store) UpdateImproveRunCrewTask(ctx context.Context, runID, crewTaskID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE rag.improve_runs
		SET crew_task_id = $2, status = 'running', updated_at = NOW()
		WHERE id = $1::uuid AND status IN ('queued', 'running')`, runID, crewTaskID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AppendImproveRunStep(ctx context.Context, runID string, step any) error {
	arr, err := json.Marshal([]any{step})
	if err != nil {
		return err
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE rag.improve_runs
		SET steps = steps || $2::jsonb, updated_at = NOW()
		WHERE id = $1::uuid`, runID, string(arr))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CompleteImproveRun(ctx context.Context, runID string, status ImproveRunStatus, citations any, metrics any, errorCode string, latencyMs int) error {
	cit := []byte("[]")
	if citations != nil {
		if b, err := json.Marshal(citations); err == nil {
			cit = b
		}
	}
	met := []byte("{}")
	if metrics != nil {
		if b, err := json.Marshal(metrics); err == nil {
			met = b
		}
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE rag.improve_runs
		SET status = $2::rag.improve_run_status,
		    citations = $3::jsonb,
		    metrics = $4::jsonb,
		    error_code = $5,
		    latency_ms = $6,
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE id = $1::uuid AND status IN ('queued', 'running')`,
		runID, string(status), string(cit), string(met), errorCode, latencyMs)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CancelImproveRun(ctx context.Context, runID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE rag.improve_runs
		SET status = 'cancelled', updated_at = NOW(), completed_at = NOW()
		WHERE id = $1::uuid AND status IN ('queued', 'running')`, runID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// LatestCompletedImproveRunForReport returns the newest completed run for a report (citations polish).
func (s *Store) LatestCompletedImproveRunForReport(ctx context.Context, reportID string) (ImproveRun, error) {
	row := s.pool.QueryRow(ctx, improveRunSelect+`
		FROM rag.improve_runs
		WHERE report_id = $1::uuid AND status = 'completed'
		ORDER BY completed_at DESC NULLS LAST, created_at DESC
		LIMIT 1`, reportID)
	return scanImproveRun(row)
}

// ImproveRunStats aggregates advanced improve runs for admin ops.
type ImproveRunStats struct {
	Days           int            `json:"days"`
	Total          int            `json:"total"`
	ByStatus       map[string]int `json:"byStatus"`
	LatencyP50Ms   float64        `json:"latencyP50Ms"`
	LatencyP95Ms   float64        `json:"latencyP95Ms"`
	CancelRate     float64        `json:"cancelRate"`
	ErrorRate      float64        `json:"errorRate"`
	AvgRagHitCount float64        `json:"avgRagHitCount"`
}

func (s *Store) ImproveRunStats(ctx context.Context, days int) (ImproveRunStats, error) {
	if days < 1 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	out := ImproveRunStats{Days: days, ByStatus: map[string]int{}}
	rows, err := s.pool.Query(ctx, `
		SELECT status::text, COUNT(*)::int
		FROM rag.improve_runs
		WHERE created_at >= NOW() - ($1::int * INTERVAL '1 day')
		GROUP BY status`, days)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return out, err
		}
		out.ByStatus[st] = n
		out.Total += n
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY latency_ms), 0),
		       COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms), 0),
		       COALESCE(AVG(NULLIF((metrics->>'ragHitCount')::float, 0)), 0)
		FROM rag.improve_runs
		WHERE created_at >= NOW() - ($1::int * INTERVAL '1 day')
		  AND status IN ('completed', 'failed', 'cancelled')
		  AND latency_ms > 0`, days).Scan(&out.LatencyP50Ms, &out.LatencyP95Ms, &out.AvgRagHitCount)
	if out.Total > 0 {
		out.CancelRate = float64(out.ByStatus["cancelled"]) / float64(out.Total)
		out.ErrorRate = float64(out.ByStatus["failed"]) / float64(out.Total)
	}
	return out, nil
}

type improveRunScanner interface {
	Scan(dest ...any) error
}

func scanImproveRun(row improveRunScanner) (ImproveRun, error) {
	var r ImproveRun
	var status string
	err := row.Scan(
		&r.ID, &r.VisitID, &r.ReportID, &r.PracticeID, &r.UserID,
		&status, &r.CrewTaskID, &r.Steps, &r.Citations, &r.Metrics, &r.ErrorCode, &r.LatencyMs,
		&r.CreatedAt, &r.UpdatedAt, &r.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ImproveRun{}, ErrNotFound
		}
		return ImproveRun{}, err
	}
	r.Status = ImproveRunStatus(status)
	if len(r.Steps) == 0 {
		r.Steps = json.RawMessage("[]")
	}
	if len(r.Citations) == 0 {
		r.Citations = json.RawMessage("[]")
	}
	if len(r.Metrics) == 0 {
		r.Metrics = json.RawMessage("{}")
	}
	return r, nil
}
