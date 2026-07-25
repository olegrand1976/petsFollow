DELETE FROM billing.commission_ledger WHERE source_kind = 'care_pro';

DROP INDEX IF EXISTS billing.uq_commission_ledger_care_pro;
DROP INDEX IF EXISTS billing.uq_commission_ledger_pet_plan;

ALTER TABLE billing.commission_ledger
  ADD CONSTRAINT commission_ledger_entitlement_id_key UNIQUE (entitlement_id);

ALTER TABLE billing.commission_ledger
  DROP CONSTRAINT IF EXISTS commission_ledger_source_kind_check;

ALTER TABLE billing.commission_ledger
  ADD CONSTRAINT commission_ledger_source_kind_check
  CHECK (source_kind IN ('pet_plan', 'addon'));
