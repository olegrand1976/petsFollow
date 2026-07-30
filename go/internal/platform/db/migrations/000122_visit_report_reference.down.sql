DROP INDEX IF EXISTS visits.visit_reports_is_reference_idx;
DROP INDEX IF EXISTS visits.visit_reports_practice_reference_idx;
ALTER TABLE visits.visit_reports DROP COLUMN IF EXISTS is_reference;
