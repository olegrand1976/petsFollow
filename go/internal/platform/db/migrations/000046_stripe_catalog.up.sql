-- Stripe product/price catalogue (admin CSV import + checkout lookup).
-- No environment-specific Stripe IDs here — seed via `make seed` or admin CSV import.

CREATE TABLE IF NOT EXISTS billing.stripe_products (
    stripe_product_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    tax_code TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    metadata_plan_slug TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.stripe_prices (
    stripe_price_id TEXT PRIMARY KEY,
    stripe_product_id TEXT NOT NULL REFERENCES billing.stripe_products(stripe_product_id) ON DELETE CASCADE,
    amount_cents INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'eur',
    interval TEXT NOT NULL DEFAULT '',
    interval_count INT NOT NULL DEFAULT 0,
    billing_scheme TEXT NOT NULL DEFAULT '',
    tax_behavior TEXT NOT NULL DEFAULT '',
    plan_code TEXT CHECK (plan_code IS NULL OR plan_code IN ('monthly', 'annual', 'triennial', 'quinquennial')),
    billing_mode TEXT CHECK (billing_mode IS NULL OR billing_mode IN ('one_time', 'subscription')),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT stripe_prices_plan_mode_pair CHECK (
        (plan_code IS NULL AND billing_mode IS NULL)
        OR (plan_code IS NOT NULL AND billing_mode IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_stripe_prices_active_plan_mode
    ON billing.stripe_prices (plan_code, billing_mode)
    WHERE active AND plan_code IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_stripe_prices_product ON billing.stripe_prices(stripe_product_id);
