-- Configurable appointment (RDV) types per practice: name, duration, calendar color.

CREATE TABLE IF NOT EXISTS practice.visit_types (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    duration_minutes INT NOT NULL
        CHECK (duration_minutes >= 5 AND duration_minutes <= 480),
    color TEXT NOT NULL
        CHECK (color ~ '^#[0-9A-Fa-f]{6}$'),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (practice_id, name)
);

CREATE INDEX IF NOT EXISTS idx_visit_types_practice
    ON practice.visit_types (practice_id, sort_order, name);

ALTER TABLE visits.visits
    DROP CONSTRAINT IF EXISTS visits_duration_minutes_check;

ALTER TABLE visits.visits
    ADD CONSTRAINT visits_duration_minutes_check
    CHECK (duration_minutes IS NULL OR (duration_minutes >= 5 AND duration_minutes <= 480));

ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS visit_type_id UUID NULL
        REFERENCES practice.visit_types(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_visits_visit_type
    ON visits.visits (visit_type_id)
    WHERE visit_type_id IS NOT NULL;

GRANT SELECT, INSERT, UPDATE, DELETE ON practice.visit_types TO CURRENT_USER;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.visit_types TO petsfollow_app;
  END IF;
END $$;
