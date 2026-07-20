CREATE TABLE IF NOT EXISTS practice.client_import_jobs (
    id UUID PRIMARY KEY,
    vet_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    created_by_admin_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT '',
    source_format TEXT NOT NULL CHECK (source_format IN ('csv', 'xlsx')),
    status TEXT NOT NULL DEFAULT 'uploaded'
        CHECK (status IN (
            'uploaded', 'mapping_ready', 'preview_ready', 'approved',
            'importing', 'completed', 'failed', 'cancelled'
        )),
    headers JSONB NOT NULL DEFAULT '[]'::jsonb,
    sample_rows JSONB NOT NULL DEFAULT '[]'::jsonb,
    column_mapping JSONB,
    gemini_raw JSONB,
    row_count INT NOT NULL DEFAULT 0,
    ok_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    created_count INT NOT NULL DEFAULT 0,
    credentials_cipher BYTEA,
    credentials_token_hash TEXT,
    credentials_expires_at TIMESTAMPTZ,
    credentials_downloaded_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_import_jobs_created
    ON practice.client_import_jobs (created_at DESC);

CREATE TABLE IF NOT EXISTS practice.client_import_rows (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES practice.client_import_jobs(id) ON DELETE CASCADE,
    row_number INT NOT NULL,
    raw JSONB NOT NULL DEFAULT '{}'::jsonb,
    email TEXT,
    full_name TEXT,
    locale TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'excluded', 'ready', 'created', 'error')),
    error_code TEXT,
    error_message TEXT,
    created_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    UNIQUE (job_id, row_number)
);

CREATE INDEX IF NOT EXISTS idx_client_import_rows_job
    ON practice.client_import_rows (job_id, row_number);
