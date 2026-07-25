ALTER TABLE billing.payout_runs DROP CONSTRAINT IF EXISTS payout_runs_status_check;
UPDATE billing.payout_runs SET status = 'closed' WHERE status = 'partially_paid';
ALTER TABLE billing.payout_runs
    ADD CONSTRAINT payout_runs_status_check
    CHECK (status IN ('open', 'closed', 'paid'));

ALTER TABLE billing.payout_lines DROP CONSTRAINT IF EXISTS payout_lines_status_check;
UPDATE billing.payout_lines SET status = 'pending' WHERE status IN ('accruing', 'missing_info', 'ready_to_pay');
ALTER TABLE billing.payout_lines
    ADD CONSTRAINT payout_lines_status_check
    CHECK (status IN ('pending', 'paid'));

ALTER TABLE practice.practices
    DROP COLUMN IF EXISTS company_legal_name,
    DROP COLUMN IF EXISTS vat_number,
    DROP COLUMN IF EXISTS company_number,
    DROP COLUMN IF EXISTS legal_form,
    DROP COLUMN IF EXISTS billing_same_as_practice,
    DROP COLUMN IF EXISTS billing_address_line1,
    DROP COLUMN IF EXISTS billing_address_line2,
    DROP COLUMN IF EXISTS billing_postal_code,
    DROP COLUMN IF EXISTS billing_city,
    DROP COLUMN IF EXISTS payout_iban,
    DROP COLUMN IF EXISTS payout_bic,
    DROP COLUMN IF EXISTS payout_account_holder;
