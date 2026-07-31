-- Cleanup for DBs that applied the first 000081 draft (FTS index + unused unaccent).
DROP INDEX IF EXISTS pharmacy.idx_pharmacy_ref_med_fts;
