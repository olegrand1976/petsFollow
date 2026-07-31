DROP INDEX IF EXISTS visits.idx_visits_deleted_at;
ALTER TABLE visits.visits DROP COLUMN IF EXISTS deleted_at;
