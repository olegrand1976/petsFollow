CREATE TABLE IF NOT EXISTS pharmacy.vamreg_ref_codes (
    kind TEXT NOT NULL,
    code TEXT NOT NULL,
    label_en TEXT NOT NULL DEFAULT '',
    label_nl TEXT NOT NULL DEFAULT '',
    label_fr TEXT NOT NULL DEFAULT '',
    deprecated BOOLEAN NOT NULL DEFAULT false,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (kind, code)
);

CREATE INDEX IF NOT EXISTS idx_vamreg_ref_codes_kind
    ON pharmacy.vamreg_ref_codes (kind)
    WHERE NOT deprecated;

DO $$ BEGIN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.vamreg_ref_codes TO petsfollow_app;
EXCEPTION WHEN undefined_object THEN NULL;
END $$;
