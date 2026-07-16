ALTER TABLE sales.prospects DROP CONSTRAINT IF EXISTS prospects_source_check;
DROP INDEX IF EXISTS sales.idx_prospects_referring_vet;
ALTER TABLE sales.prospects DROP COLUMN IF EXISTS referring_vet_user_id;
ALTER TABLE sales.prospects DROP COLUMN IF EXISTS source;

ALTER TABLE billing.commercial_commission_ledger
    DROP CONSTRAINT IF EXISTS commercial_commission_ledger_source_type_check;

UPDATE billing.commercial_commission_ledger
SET source_type = 'subscription_flat'
WHERE source_type = 'subscription_mirror';

ALTER TABLE billing.commercial_commission_ledger
    ADD CONSTRAINT commercial_commission_ledger_source_type_check
    CHECK (source_type IN ('subscription_flat', 'addon_pct'));
