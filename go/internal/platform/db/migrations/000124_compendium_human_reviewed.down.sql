DROP INDEX IF EXISTS pharmacy.idx_pharmacy_compendium_rows_reviewed;
ALTER TABLE pharmacy.compendium_import_rows
  DROP COLUMN IF EXISTS human_reviewed;
