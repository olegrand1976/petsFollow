ALTER TABLE billing.commission_ledger
  DROP CONSTRAINT IF EXISTS commission_ledger_source_kind_check;

ALTER TABLE billing.commission_ledger
  DROP COLUMN IF EXISTS source_kind;

ALTER TABLE billing.commission_ledger
  DROP COLUMN IF EXISTS addon_entitlement_id;

-- pet_id / entitlement_id NOT NULL restore only if no null rows remain.
DELETE FROM billing.commission_ledger WHERE pet_id IS NULL OR entitlement_id IS NULL;
ALTER TABLE billing.commission_ledger ALTER COLUMN pet_id SET NOT NULL;
ALTER TABLE billing.commission_ledger ALTER COLUMN entitlement_id SET NOT NULL;

ALTER TABLE pets.pets DROP COLUMN IF EXISTS litter_tag;

ALTER TABLE billing.addon_entitlements
  DROP CONSTRAINT IF EXISTS addon_entitlements_addon_code_check;

ALTER TABLE billing.addon_entitlements
  ADD CONSTRAINT addon_entitlements_addon_code_check
  CHECK (addon_code IN ('family', 'care_plus', 'horse'));
