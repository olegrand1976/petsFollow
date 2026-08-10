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
	ErrorCode   string           `json:"errorCode,omitempty"`
	LatencyMs   int              `json:"latencyMs"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
}

func (s *Store) CreateImproveRun(ctx context.Context, visitID, reportID, practiceID, userID string) (ImproveRun, error) {
	id := uuid.NewString()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO rag.improve_runs (id, visit_id, report_id, practice_id, user_id, status)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, 'queued')
		RETURNING id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		          status::text, crew_task_id, steps, citations, error_code, latency_ms,
		          created_at, updated_at, completed_at`,
		id, visitID, reportID, practiceID, userID,
	)
	return scanImproveRun(row)
}

func (s *Store) GetImproveRun(ctx context.Context, runID string) (ImproveRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		       status::text, crew_task_id, steps, citations, error_code, latency_ms,
		       created_at, updated_at, completed_at
		FROM rag.improve_runs WHERE id = $1::uuid`, runID)
	return scanImproveRun(row)
}

func (s *Store) ActiveImproveRunForVisit(ctx context.Context, visitID string) (ImproveRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		       status::text, crew_task_id, steps, citations, error_code, latency_ms,
		       created_at, updated_at, completed_at
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

func (s *Store) CompleteImproveRun(ctx context.Context, runID string, status ImproveRunStatus, citations any, errorCode string, latencyMs int) error {
	cit := []byte("[]")
	if citations != nil {
		if b, err := json.Marshal(citations); err == nil {
			cit = b
		}
	}
	ct, err := s.pool.Exec(ctx, `
		UPDATE rag.improve_runs
		SET status = $2::rag.improve_run_status,
		    citations = $3::jsonb,
		    error_code = $4,
		    latency_ms = $5,
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE id = $1::uuid AND status IN ('queued', 'running')`,
		runID, string(status), string(cit), errorCode, latencyMs)
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
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, visit_id::text, report_id::text, practice_id::text, user_id::text,
		       status::text, crew_task_id, steps, citations, error_code, latency_ms,
		       created_at, updated_at, completed_at
		FROM rag.improve_runs
		WHERE report_id = $1::uuid AND status = 'completed'
		ORDER BY completed_at DESC NULLS LAST, created_at DESC
		LIMIT 1`, reportID)
	return scanImproveRun(row)
}

type improveRunScanner interface {
	Scan(dest ...any) error
}

func scanImproveRun(row improveRunScanner) (ImproveRun, error) {
	var r ImproveRun
	var status string
	err := row.Scan(
		&r.ID, &r.VisitID, &r.ReportID, &r.PracticeID, &r.UserID,
		&status, &r.CrewTaskID, &r.Steps, &r.Citations, &r.ErrorCode, &r.LatencyMs,
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
	return r, nil
}
