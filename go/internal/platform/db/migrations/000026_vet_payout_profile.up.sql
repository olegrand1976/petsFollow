-- Vet practice payout / company profile for commission sheets
ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS company_legal_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS vat_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS company_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS legal_form TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS billing_same_as_practice BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS billing_address_line1 TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS billing_address_line2 TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS billing_postal_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS billing_city TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payout_iban TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payout_bic TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payout_account_holder TEXT NOT NULL DEFAULT '';

-- Finer payout line statuses
ALTER TABLE billing.payout_lines DROP CONSTRAINT IF EXISTS payout_lines_status_check;
UPDATE billing.payout_lines SET status = 'missing_info' WHERE status = 'pending';
ALTER TABLE billing.payout_lines
    ADD CONSTRAINT payout_lines_status_check
    CHECK (status IN ('accruing', 'missing_info', 'ready_to_pay', 'paid'));

-- Run can be partially paid
ALTER TABLE billing.payout_runs DROP CONSTRAINT IF EXISTS payout_runs_status_check;
ALTER TABLE billing.payout_runs
    ADD CONSTRAINT payout_runs_status_check
    CHECK (status IN ('open', 'closed', 'partially_paid', 'paid'));
