ALTER TABLE identity.users
    DROP COLUMN IF EXISTS totp_enabled,
    DROP COLUMN IF EXISTS totp_secret,
    DROP COLUMN IF EXISTS auth_provider,
    DROP COLUMN IF EXISTS google_sub;

ALTER TABLE identity.users
    ALTER COLUMN password_hash SET NOT NULL;

DROP INDEX IF EXISTS identity.idx_users_google_sub;
