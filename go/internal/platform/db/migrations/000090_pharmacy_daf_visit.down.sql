DROP INDEX IF EXISTS pharmacy.idx_pharmacy_daf_visit;
ALTER TABLE pharmacy.daf_documents DROP COLUMN IF EXISTS visit_id;
