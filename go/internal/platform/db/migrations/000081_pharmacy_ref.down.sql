DROP INDEX IF EXISTS pharmacy.idx_pharmacy_ref_med_active_cnk;
DROP INDEX IF EXISTS pharmacy.idx_pharmacy_ref_med_antibiotic;
DROP INDEX IF EXISTS pharmacy.idx_pharmacy_ref_med_name_trgm;
DROP TABLE IF EXISTS pharmacy.ref_medications;
DROP SCHEMA IF EXISTS pharmacy CASCADE;
-- pg_trgm left installed (shared extension).
