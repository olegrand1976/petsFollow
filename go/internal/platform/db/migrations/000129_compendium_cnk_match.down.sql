ALTER TABLE pharmacy.compendium_import_rows
    DROP COLUMN IF EXISTS match_candidates,
    DROP COLUMN IF EXISTS match_score,
    DROP COLUMN IF EXISTS suggested_cnk,
    DROP COLUMN IF EXISTS active_substance,
    DROP COLUMN IF EXISTS manufacturer;
