CREATE SCHEMA IF NOT EXISTS billing;

ALTER TABLE pets.pets
    ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'pending_payment'
        CHECK (payment_status IN ('pending_payment', 'active', 'expired'));

CREATE TABLE IF NOT EXISTS billing.stripe_customers (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    stripe_customer_id TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.pet_entitlements (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL UNIQUE REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id),
    plan_code TEXT NOT NULL CHECK (plan_code IN ('annual', 'triennial', 'quinquennial')),
    billing_mode TEXT NOT NULL CHECK (billing_mode IN ('one_time', 'subscription')),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'past_due', 'expired', 'cancelled')),
    amount_cents INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'eur',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    stripe_checkout_session_id TEXT,
    stripe_payment_intent_id TEXT,
    stripe_subscription_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pet_entitlements_owner ON billing.pet_entitlements(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_pet_entitlements_status ON billing.pet_entitlements(status);
CREATE INDEX IF NOT EXISTS idx_pet_entitlements_created ON billing.pet_entitlements(created_at);

CREATE TABLE IF NOT EXISTS billing.stripe_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
