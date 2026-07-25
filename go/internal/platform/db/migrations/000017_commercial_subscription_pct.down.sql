ALTER TABLE billing.commercial_commission_ledger
    DROP CONSTRAINT IF EXISTS commercial_commission_ledger_source_type_check;

UPDATE billing.commercial_commission_ledger
SET source_type = 'subscription_mirror'
WHERE source_type = 'subscription_pct';

ALTER TABLE billing.commercial_commission_ledger
    ADD CONSTRAINT commercial_commission_ledger_source_type_check
    CHECK (source_type IN ('subscription_mirror', 'addon_pct'));
