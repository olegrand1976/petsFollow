-- Client→client sponsorship (QR parrainage) — first-wins, no client commission.

CREATE TABLE IF NOT EXISTS practice.client_referrals (
    referred_client_user_id UUID PRIMARY KEY
        REFERENCES identity.users(id) ON DELETE CASCADE,
    sponsor_client_user_id UUID NOT NULL
        REFERENCES identity.users(id) ON DELETE CASCADE,
    invite_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT client_referrals_no_self CHECK (referred_client_user_id <> sponsor_client_user_id)
);

CREATE INDEX IF NOT EXISTS client_referrals_sponsor_idx
    ON practice.client_referrals (sponsor_client_user_id);
