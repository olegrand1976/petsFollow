-- Flux A opt-in: only practices with saas_billing_enabled receive C1 drafts / live SaaS invoices.
ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS saas_billing_enabled BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN practice.practices.saas_billing_enabled IS
    'Flux A: cabinet opted-in for LL-IT-SC SaaS Peppol invoices (cron C1 + admin draft).';
