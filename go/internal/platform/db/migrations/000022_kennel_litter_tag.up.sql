-- Kennel addon + optional litter tag for quick encode / roster.
ALTER TABLE billing.addon_entitlements
  DROP CONSTRAINT IF EXISTS addon_entitlements_addon_code_check;

ALTER TABLE billing.addon_entitlements
  ADD CONSTRAINT addon_entitlements_addon_code_check
  CHECK (addon_code IN ('family', 'care_plus', 'horse', 'kennel'));

ALTER TABLE pets.pets
  ADD COLUMN IF NOT EXISTS litter_tag TEXT NOT NULL DEFAULT '';

-- Vet commission on Family/Kennel addons (flat 5%), separate from progressive pet ranks.
ALTER TABLE billing.commission_ledger
  ALTER COLUMN pet_id DROP NOT NULL;

ALTER TABLE billing.commission_ledger
  ALTER COLUMN entitlement_id DROP NOT NULL;

ALTER TABLE billing.commission_ledger
  ADD COLUMN IF NOT EXISTS addon_entitlement_id UUID UNIQUE
    REFERENCES billing.addon_entitlements(id) ON DELETE CASCADE;

ALTER TABLE billing.commission_ledger
  ADD COLUMN IF NOT EXISTS source_kind TEXT NOT NULL DEFAULT 'pet_plan';

ALTER TABLE billing.commission_ledger
  DROP CONSTRAINT IF EXISTS commission_ledger_source_kind_check;

ALTER TABLE billing.commission_ledger
  ADD CONSTRAINT commission_ledger_source_kind_check
  CHECK (source_kind IN ('pet_plan', 'addon'));
