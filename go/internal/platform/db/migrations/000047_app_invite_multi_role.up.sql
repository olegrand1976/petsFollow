-- Generalize vet invite codes to multi-role app invites + commercial referrals.

CREATE TABLE IF NOT EXISTS practice.app_invite_codes (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    practice_id UUID REFERENCES practice.practices(id) ON DELETE SET NULL,
    code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT app_invite_codes_code_unique UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_app_invite_codes_practice
    ON practice.app_invite_codes (practice_id);

-- Migrate existing vet invite codes if present.
INSERT INTO practice.app_invite_codes (user_id, role, practice_id, code, created_at)
SELECT vet_user_id, 'vet', practice_id, code, created_at
FROM practice.vet_app_invite_codes
ON CONFLICT (user_id) DO NOTHING;

DROP TABLE IF EXISTS practice.vet_app_invite_codes;

CREATE TABLE IF NOT EXISTS practice.commercial_referrals (
    client_user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    commercial_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    invite_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commercial_referrals_commercial
    ON practice.commercial_referrals (commercial_user_id);
