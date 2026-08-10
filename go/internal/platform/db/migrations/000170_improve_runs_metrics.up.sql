-- Improve-run metrics (ragHitCount, future token usage) for admin stats / ops.

ALTER TABLE rag.improve_runs
    ADD COLUMN IF NOT EXISTS metrics JSONB NOT NULL DEFAULT '{}'::jsonb;
