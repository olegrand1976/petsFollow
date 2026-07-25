-- Ledger care_pro : autorise plusieurs lignes par entitlement (véto + care_pros),
-- dédupliquées par (entitlement, bénéficiaire) selon la source.
ALTER TABLE billing.commission_ledger
  DROP CONSTRAINT IF EXISTS commission_ledger_source_kind_check;

ALTER TABLE billing.commission_ledger
  ADD CONSTRAINT commission_ledger_source_kind_check
  CHECK (source_kind IN ('pet_plan', 'addon', 'care_pro'));

-- L'UNIQUE de 000007 empêchait une 2e ligne (care_pro) sur le même entitlement.
ALTER TABLE billing.commission_ledger
  DROP CONSTRAINT IF EXISTS commission_ledger_entitlement_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_commission_ledger_pet_plan
  ON billing.commission_ledger(entitlement_id)
  WHERE source_kind = 'pet_plan';

CREATE UNIQUE INDEX IF NOT EXISTS uq_commission_ledger_care_pro
  ON billing.commission_ledger(entitlement_id, vet_user_id)
  WHERE source_kind = 'care_pro';
