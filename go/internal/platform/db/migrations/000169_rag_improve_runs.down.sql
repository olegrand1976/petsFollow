DROP TABLE IF EXISTS rag.improve_runs;
DROP TYPE IF EXISTS rag.improve_run_status;

ALTER TABLE practice.ai_cr_usage_events
    DROP CONSTRAINT IF EXISTS ai_cr_usage_events_kind_check;
ALTER TABLE practice.ai_cr_usage_events
    ADD CONSTRAINT ai_cr_usage_events_kind_check
    CHECK (kind IN ('transcribe', 'improve', 'finalize', 'error'));
