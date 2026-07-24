-- Commercial sales force: role, vet assignment, CRM prospects, addons and commissions.

-- 1. Allow the "commercial" role.
ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check CHECK (role IN ('vet', 'client', 'admin', 'commercial'));

-- 2. Assign a vet to a commercial (force de vente).
ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS assigned_commercial_id UUID REFERENCES identity.users(id);

CREATE INDEX IF NOT EXISTS idx_users_assigned_commercial
    ON identity.users(assigned_commercial_id) WHERE assigned_commercial_id IS NOT NULL;

-- 3. Sales CRM prospects (commercial-owned + vet referrals).
CREATE SCHEMA IF NOT EXISTS sales;

CREATE TABLE IF NOT EXISTS sales.prospects (
    id UUID PRIMARY KEY,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    practice_name TEXT NOT NULL,
    contact_name TEXT NOT NULL DEFAULT '',
    contact_email TEXT NOT NULL DEFAULT '',
    contact_phone TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'commercial'
        CHECK (source IN ('commercial', 'vet_referral')),
    referring_vet_user_id UUID REFERENCES identity.users(id),
    status TEXT NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'contacted', 'qualified', 'converted', 'lost')),
    status_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prospects_commercial ON sales.prospects(commercial_user_id);
CREATE INDEX IF NOT EXISTS idx_prospects_status ON sales.prospects(status);

-- 4. Paid addons (family / care_plus / horse) per owner.
CREATE TABLE IF NOT EXISTS billing.addon_entitlements (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    addon_code TEXT NOT NULL CHECK (addon_code IN ('family', 'care_plus', 'horse')),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'expired', 'cancelled')),
    amount_cents INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'eur',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    stripe_checkout_session_id TEXT,
    stripe_payment_intent_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_addon_entitlements_owner ON billing.addon_entitlements(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_addon_entitlements_status ON billing.addon_entitlements(status);

-- 5. Commercial commission ledger (mirror vet subscription + 15% addons).
CREATE TABLE IF NOT EXISTS billing.commercial_commission_ledger (
    id UUID PRIMARY KEY,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id),
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    client_user_id UUID NOT NULL REFERENCES identity.users(id),
    source_type TEXT NOT NULL CHECK (source_type IN ('subscription_mirror', 'addon_pct')),
    source_id UUID NOT NULL,
    base_amount_cents INT NOT NULL,
    rate_bps INT NOT NULL,
    commission_cents INT NOT NULL,
    period_ym TEXT NOT NULL,
    accrued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_type, source_id)
);

CREATE INDEX IF NOT EXISTS idx_commercial_ledger_commercial ON billing.commercial_commission_ledger(commercial_user_id);
CREATE INDEX IF NOT EXISTS idx_commercial_ledger_period ON billing.commercial_commission_ledger(period_ym);
CREATE INDEX IF NOT EXISTS idx_commercial_ledger_commercial_period ON billing.commercial_commission_ledger(commercial_user_id, period_ym);

-- 6. Commercial payout runs / lines.
CREATE TABLE IF NOT EXISTS billing.commercial_payout_runs (
    id UUID PRIMARY KEY,
    period_ym TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closed', 'paid')),
    closed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.commercial_payout_lines (
    id UUID PRIMARY KEY,
    run_id UUID NOT NULL REFERENCES billing.commercial_payout_runs(id) ON DELETE CASCADE,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id),
    ledger_count INT NOT NULL DEFAULT 0,
    amount_cents INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid')),
    UNIQUE (run_id, commercial_user_id)
);

CREATE INDEX IF NOT EXISTS idx_commercial_payout_lines_run ON billing.commercial_payout_lines(run_id);

-- 7. Grant the sales schema to the runtime role when present.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA sales TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA sales TO petsfollow_app;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA sales TO petsfollow_app;
    ALTER DEFAULT PRIVILEGES IN SCHEMA sales GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON
        billing.addon_entitlements,
        billing.commercial_commission_ledger,
        billing.commercial_payout_runs,
        billing.commercial_payout_lines
        TO petsfollow_app;
  END IF;
END $$;
