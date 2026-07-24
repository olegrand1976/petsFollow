ALTER TABLE billing.addon_entitlements
  DROP CONSTRAINT IF EXISTS addon_entitlements_status_check;

ALTER TABLE billing.addon_entitlements
  ADD CONSTRAINT addon_entitlements_status_check
  CHECK (status IN ('pending', 'active', 'expired', 'cancelled'));

DROP INDEX IF EXISTS billing.idx_addon_entitlements_stripe_sub;

ALTER TABLE billing.addon_entitlements
  DROP COLUMN IF EXISTS stripe_subscription_id;
