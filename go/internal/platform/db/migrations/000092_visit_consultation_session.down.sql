DROP INDEX IF EXISTS visits.idx_visits_consultation_session;
ALTER TABLE visits.visits DROP COLUMN IF EXISTS consultation_session;

