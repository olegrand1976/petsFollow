DROP INDEX IF EXISTS visits.idx_visits_visit_type;
ALTER TABLE visits.visits DROP COLUMN IF EXISTS visit_type_id;

ALTER TABLE visits.visits DROP CONSTRAINT IF EXISTS visits_duration_minutes_check;
ALTER TABLE visits.visits
    ADD CONSTRAINT visits_duration_minutes_check
    CHECK (duration_minutes IS NULL OR duration_minutes IN (15, 30, 60));

DROP TABLE IF EXISTS practice.visit_types;
