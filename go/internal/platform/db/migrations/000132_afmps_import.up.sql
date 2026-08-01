-- AFMPS pack CSV → staging → pharmacy.ref_medications (triple contrôle)
CREATE TABLE IF NOT EXISTS pharmacy.afmps_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by_admin_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'text/csv',
    file_bytes INT NOT NULL DEFAULT 0 CHECK (file_bytes >= 0),
    checksum_sha256 TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'validated'
        CHECK (status IN (
            'validated', 'blocked', 'reviewed', 'committing', 'completed', 'failed'
        )),
    report_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    row_count INT NOT NULL DEFAULT 0 CHECK (row_count >= 0),
    ready_count INT NOT NULL DEFAULT 0 CHECK (ready_count >= 0),
    error_count INT NOT NULL DEFAULT 0 CHECK (error_count >= 0),
    insert_count INT NOT NULL DEFAULT 0 CHECK (insert_count >= 0),
    update_count INT NOT NULL DEFAULT 0 CHECK (update_count >= 0),
    unchanged_count INT NOT NULL DEFAULT 0 CHECK (unchanged_count >= 0),
    deactivate_preview INT NOT NULL DEFAULT 0 CHECK (deactivate_preview >= 0),
    upserted_count INT NOT NULL DEFAULT 0 CHECK (upserted_count >= 0),
    deactivated_count INT NOT NULL DEFAULT 0 CHECK (deactivated_count >= 0),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_afmps_jobs_created
    ON pharmacy.afmps_import_jobs (created_at DESC);

CREATE TABLE IF NOT EXISTS pharmacy.afmps_import_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES pharmacy.afmps_import_jobs(id) ON DELETE CASCADE,
    row_number INT NOT NULL CHECK (row_number > 0),
    source_line INT NOT NULL DEFAULT 0,
    cnk TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    atc_code TEXT NOT NULL DEFAULT '',
    pharmaceutical_form TEXT NOT NULL DEFAULT '',
    pack_size TEXT NOT NULL DEFAULT '',
    amm_number TEXT NOT NULL DEFAULT '',
    is_antibiotic BOOLEAN NOT NULL DEFAULT false,
    collision TEXT NOT NULL DEFAULT 'insert'
        CHECK (collision IN ('insert', 'update', 'unchanged', 'error')),
    afmps_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'ready'
        CHECK (status IN ('ready', 'error', 'excluded', 'upserted')),
    error_code TEXT,
    error_message TEXT,
    UNIQUE (job_id, row_number)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_afmps_rows_job
    ON pharmacy.afmps_import_rows (job_id, row_number);

CREATE INDEX IF NOT EXISTS idx_pharmacy_afmps_rows_cnk
    ON pharmacy.afmps_import_rows (job_id, cnk);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.afmps_import_jobs TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.afmps_import_rows TO petsfollow_app;
  END IF;
END $$;
