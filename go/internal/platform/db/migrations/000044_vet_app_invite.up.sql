-- Durable per-vet invite code for app download QR / referral attribution.
CREATE TABLE IF NOT EXISTS practice.vet_app_invite_codes (
    vet_user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vet_app_invite_codes_code_unique UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_vet_app_invite_codes_practice
    ON practice.vet_app_invite_codes (practice_id);
