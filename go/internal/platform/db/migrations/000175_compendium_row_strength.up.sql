-- Persist extracted strength for Compendium review UI (was only in raw_json).
ALTER TABLE pharmacy.compendium_import_rows
    ADD COLUMN IF NOT EXISTS strength TEXT NOT NULL DEFAULT '';

UPDATE pharmacy.compendium_import_rows
SET strength = COALESCE(NULLIF(TRIM(raw_json->>'strength'), ''), strength)
WHERE strength = ''
  AND raw_json ? 'strength';
