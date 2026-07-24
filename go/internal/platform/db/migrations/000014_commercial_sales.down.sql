DROP TABLE IF EXISTS billing.commercial_payout_lines;
DROP TABLE IF EXISTS billing.commercial_payout_runs;
DROP TABLE IF EXISTS billing.commercial_commission_ledger;
DROP TABLE IF EXISTS billing.addon_entitlements;
DROP TABLE IF EXISTS sales.prospects;
DROP SCHEMA IF EXISTS sales;

ALTER TABLE identity.users DROP COLUMN IF EXISTS assigned_commercial_id;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check CHECK (role IN ('vet', 'client', 'admin'));
