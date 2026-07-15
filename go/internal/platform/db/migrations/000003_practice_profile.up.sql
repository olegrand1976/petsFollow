ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_email TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS address_line1 TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS address_line2 TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS postal_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS website TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS profile_completed_at TIMESTAMPTZ;

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS identity.email_verification_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_verification_token ON identity.email_verification_tokens(token);

-- Comptes existants : considérés vérifiés et profilés
UPDATE identity.users SET email_verified_at = COALESCE(email_verified_at, created_at) WHERE email_verified_at IS NULL;
UPDATE practice.practices SET profile_completed_at = COALESCE(profile_completed_at, created_at) WHERE profile_completed_at IS NULL;
