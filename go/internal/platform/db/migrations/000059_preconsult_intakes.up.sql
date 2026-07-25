CREATE TABLE IF NOT EXISTS visits.preconsult_intakes (
    id UUID PRIMARY KEY,
    visit_id UUID NOT NULL UNIQUE REFERENCES visits.visits(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'submitted', 'skipped')),
    answers JSONB NOT NULL DEFAULT '{}'::jsonb,
    submitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_preconsult_visit ON visits.preconsult_intakes(visit_id);
CREATE INDEX IF NOT EXISTS idx_preconsult_status ON visits.preconsult_intakes(status)
    WHERE status = 'pending';

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON visits.preconsult_intakes TO petsfollow_app;
  END IF;
END $$;
