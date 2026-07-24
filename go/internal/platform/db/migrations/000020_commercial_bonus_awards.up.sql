-- SPIFF awards for commercials (Ramp cabinet + Mix triennial). Manual payout V1.

CREATE TABLE IF NOT EXISTS billing.commercial_bonus_awards (
    id UUID PRIMARY KEY,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id),
    bonus_code TEXT NOT NULL
        CHECK (bonus_code IN ('commercial_ramp', 'commercial_mix')),
    amount_cents INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'earned'
        CHECK (status IN ('earned', 'paid')),
    period_ym TEXT,
    vet_user_id UUID REFERENCES identity.users(id),
    progress INT NOT NULL DEFAULT 0,
    target INT NOT NULL DEFAULT 0,
    dedupe_key TEXT NOT NULL UNIQUE,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    paid_by_admin_id UUID REFERENCES identity.users(id),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT commercial_bonus_ramp_fields CHECK (
        bonus_code <> 'commercial_ramp' OR vet_user_id IS NOT NULL
    ),
    CONSTRAINT commercial_bonus_mix_fields CHECK (
        bonus_code <> 'commercial_mix' OR period_ym IS NOT NULL
    )
);

CREATE INDEX IF NOT EXISTS idx_commercial_bonus_awards_commercial
    ON billing.commercial_bonus_awards(commercial_user_id);
CREATE INDEX IF NOT EXISTS idx_commercial_bonus_awards_status
    ON billing.commercial_bonus_awards(status);
CREATE INDEX IF NOT EXISTS idx_commercial_bonus_awards_period
    ON billing.commercial_bonus_awards(period_ym)
    WHERE period_ym IS NOT NULL;
