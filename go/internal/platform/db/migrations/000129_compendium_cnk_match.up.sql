-- Compendium staging: AFMPS CNK match fields (Vetcompendium PDF has no CNK).
ALTER TABLE pharmacy.compendium_import_rows
    ADD COLUMN IF NOT EXISTS manufacturer TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS active_substance TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS suggested_cnk TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS match_score DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS match_candidates JSONB NOT NULL DEFAULT '[]'::jsonb;
