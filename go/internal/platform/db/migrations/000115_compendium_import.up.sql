-- Admin Compendium PDF → staging → pharmacy.ref_medications
CREATE TABLE IF NOT EXISTS pharmacy.compendium_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by_admin_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE RESTRICT,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'application/pdf',
    page_start INT NOT NULL CHECK (page_start >= 1),
    page_end INT NOT NULL CHECK (page_end >= page_start),
    status TEXT NOT NULL DEFAULT 'uploaded'
        CHECK (status IN (
            'uploaded', 'extracting', 'extracted', 'committing', 'completed', 'failed', 'cancelled'
        )),
    pdf_object_key TEXT NOT NULL DEFAULT '',
    extract_done INT NOT NULL DEFAULT 0 CHECK (extract_done >= 0),
    extract_total INT NOT NULL DEFAULT 0 CHECK (extract_total >= 0),
    row_count INT NOT NULL DEFAULT 0 CHECK (row_count >= 0),
    ready_count INT NOT NULL DEFAULT 0 CHECK (ready_count >= 0),
    error_count INT NOT NULL DEFAULT 0 CHECK (error_count >= 0),
    reviewed_count INT NOT NULL DEFAULT 0 CHECK (reviewed_count >= 0),
    upserted_count INT NOT NULL DEFAULT 0 CHECK (upserted_count >= 0),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pharmacy_compendium_jobs_page_span CHECK (page_end - page_start < 200)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_compendium_jobs_created
    ON pharmacy.compendium_import_jobs (created_at DESC);

CREATE TABLE IF NOT EXISTS pharmacy.compendium_import_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES pharmacy.compendium_import_jobs(id) ON DELETE CASCADE,
    row_number INT NOT NULL CHECK (row_number > 0),
    source_page INT,
    cnk TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    atc_code TEXT NOT NULL DEFAULT '',
    pharmaceutical_form TEXT NOT NULL DEFAULT '',
    pack_size TEXT NOT NULL DEFAULT '',
    is_antibiotic BOOLEAN NOT NULL DEFAULT false,
    raw_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'ready', 'excluded', 'error', 'upserted')),
    error_code TEXT,
    error_message TEXT,
    UNIQUE (job_id, row_number)
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_compendium_rows_job
    ON pharmacy.compendium_import_rows (job_id, row_number);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.compendium_import_jobs TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.compendium_import_rows TO petsfollow_app;
  END IF;
END $$;
