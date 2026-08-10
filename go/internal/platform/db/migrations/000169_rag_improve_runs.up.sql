-- Phase 3: audit trail for Améliorer IA avancé (CrewAI multi-agent runs).

DO $$ BEGIN
    CREATE TYPE rag.improve_run_status AS ENUM (
        'queued',
        'running',
        'completed',
        'failed',
        'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS rag.improve_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits.visits (id) ON DELETE CASCADE,
    report_id UUID NOT NULL REFERENCES visits.visit_reports (id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users (id) ON DELETE CASCADE,
    status rag.improve_run_status NOT NULL DEFAULT 'queued',
    crew_task_id TEXT NOT NULL DEFAULT '',
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    citations JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_code TEXT NOT NULL DEFAULT '',
    latency_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS rag_improve_runs_visit_idx ON rag.improve_runs (visit_id, created_at DESC);
CREATE INDEX IF NOT EXISTS rag_improve_runs_practice_idx ON rag.improve_runs (practice_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS rag_improve_runs_active_visit_uidx
    ON rag.improve_runs (visit_id)
    WHERE status IN ('queued', 'running');

ALTER TABLE practice.ai_cr_usage_events
    DROP CONSTRAINT IF EXISTS ai_cr_usage_events_kind_check;
ALTER TABLE practice.ai_cr_usage_events
    ADD CONSTRAINT ai_cr_usage_events_kind_check
    CHECK (kind IN ('transcribe', 'improve', 'improve_advanced', 'finalize', 'error'));

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON rag.improve_runs TO petsfollow_app;
  END IF;
END $$;
