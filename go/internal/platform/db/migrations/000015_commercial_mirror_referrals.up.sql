-- Align commercial commission source with mirror model + vet referral prospects.

-- 1. subscription_flat → subscription_mirror
ALTER TABLE billing.commercial_commission_ledger
    DROP CONSTRAINT IF EXISTS commercial_commission_ledger_source_type_check;

UPDATE billing.commercial_commission_ledger
SET source_type = 'subscription_mirror'
WHERE source_type = 'subscription_flat';

ALTER TABLE billing.commercial_commission_ledger
    ADD CONSTRAINT commercial_commission_ledger_source_type_check
    CHECK (source_type IN ('subscription_mirror', 'addon_pct'));

-- 2. Prospect source (commercial CRM or vet referral)
ALTER TABLE sales.prospects
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'commercial';

ALTER TABLE sales.prospects
    DROP CONSTRAINT IF EXISTS prospects_source_check;

ALTER TABLE sales.prospects
    ADD CONSTRAINT prospects_source_check
    CHECK (source IN ('commercial', 'vet_referral'));

ALTER TABLE sales.prospects
    ADD COLUMN IF NOT EXISTS referring_vet_user_id UUID REFERENCES identity.users(id);

CREATE INDEX IF NOT EXISTS idx_prospects_referring_vet
    ON sales.prospects(referring_vet_user_id)
    WHERE referring_vet_user_id IS NOT NULL;
