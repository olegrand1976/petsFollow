ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_visits_deleted_at
    ON visits.visits (practice_id, deleted_at)
    WHERE deleted_at IS NULL;
