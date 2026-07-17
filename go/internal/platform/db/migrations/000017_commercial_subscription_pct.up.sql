-- Align commercial ledger source_type with flat % model (subscription_pct).

ALTER TABLE billing.commercial_commission_ledger
    DROP CONSTRAINT IF EXISTS commercial_commission_ledger_source_type_check;

UPDATE billing.commercial_commission_ledger
SET source_type = 'subscription_pct'
WHERE source_type IN ('subscription_mirror', 'subscription_flat');

ALTER TABLE billing.commercial_commission_ledger
    ADD CONSTRAINT commercial_commission_ledger_source_type_check
    CHECK (source_type IN ('subscription_pct', 'addon_pct'));
