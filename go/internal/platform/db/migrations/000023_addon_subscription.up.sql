-- Annual recurring Stripe subscriptions for household addons.
ALTER TABLE billing.addon_entitlements
  ADD COLUMN IF NOT EXISTS stripe_subscription_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_addon_entitlements_stripe_sub
  ON billing.addon_entitlements (stripe_subscription_id)
  WHERE stripe_subscription_id IS NOT NULL AND stripe_subscription_id <> '';

ALTER TABLE billing.addon_entitlements
  DROP CONSTRAINT IF EXISTS addon_entitlements_status_check;

ALTER TABLE billing.addon_entitlements
  ADD CONSTRAINT addon_entitlements_status_check
  CHECK (status IN ('pending', 'active', 'past_due', 'expired', 'cancelled'));
