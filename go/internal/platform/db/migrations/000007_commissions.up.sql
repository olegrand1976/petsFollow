-- Progressive vet commissions (% of first pet payment)
CREATE TABLE IF NOT EXISTS billing.commission_tiers (
    id UUID PRIMARY KEY,
    min_clients INT NOT NULL,
    max_clients INT,
    rate_bps INT NOT NULL CHECK (rate_bps >= 0 AND rate_bps <= 1500),
    CONSTRAINT commission_tiers_range CHECK (max_clients IS NULL OR max_clients >= min_clients)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commission_tiers_min ON billing.commission_tiers(min_clients);

CREATE TABLE IF NOT EXISTS billing.commission_ledger (
    id UUID PRIMARY KEY,
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    client_user_id UUID NOT NULL REFERENCES identity.users(id),
    pet_id UUID NOT NULL REFERENCES pets.pets(id),
    entitlement_id UUID NOT NULL UNIQUE REFERENCES billing.pet_entitlements(id) ON DELETE CASCADE,
    base_amount_cents INT NOT NULL,
    rate_bps INT NOT NULL,
    commission_cents INT NOT NULL,
    period_ym TEXT NOT NULL,
    accrued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commission_ledger_vet ON billing.commission_ledger(vet_user_id);
CREATE INDEX IF NOT EXISTS idx_commission_ledger_period ON billing.commission_ledger(period_ym);
CREATE INDEX IF NOT EXISTS idx_commission_ledger_vet_period ON billing.commission_ledger(vet_user_id, period_ym);

CREATE TABLE IF NOT EXISTS billing.payout_runs (
    id UUID PRIMARY KEY,
    period_ym TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closed', 'paid')),
    closed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.payout_lines (
    id UUID PRIMARY KEY,
    run_id UUID NOT NULL REFERENCES billing.payout_runs(id) ON DELETE CASCADE,
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    eligible_clients INT NOT NULL DEFAULT 0,
    ledger_count INT NOT NULL DEFAULT 0,
    amount_cents INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid')),
    UNIQUE (run_id, vet_user_id)
);

CREATE INDEX IF NOT EXISTS idx_payout_lines_run ON billing.payout_lines(run_id);
