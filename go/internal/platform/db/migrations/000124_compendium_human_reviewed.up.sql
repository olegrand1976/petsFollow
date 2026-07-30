-- Track human control separately from AI classification (reviewPct).
ALTER TABLE pharmacy.compendium_import_rows
  ADD COLUMN IF NOT EXISTS human_reviewed BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_pharmacy_compendium_rows_reviewed
  ON pharmacy.compendium_import_rows (job_id)
  WHERE human_reviewed;
