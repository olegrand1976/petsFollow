-- pharmacy.job_audit — VAMReg / invoices.connect worker trail (Sprint 4)
CREATE TABLE IF NOT EXISTS pharmacy.job_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    job_type TEXT NOT NULL CHECK (job_type IN ('vamreg', 'invoices_connect')),
    entity_id UUID NOT NULL,
    attempt INT NOT NULL DEFAULT 1 CHECK (attempt > 0),
    status TEXT NOT NULL CHECK (status IN ('started', 'success', 'failed')),
    request_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_json JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pharmacy_job_audit_lookup
    ON pharmacy.job_audit (job_type, entity_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_pharmacy_job_audit_practice
    ON pharmacy.job_audit (practice_id, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pharmacy.job_audit TO petsfollow_app;
  END IF;
END $$;
