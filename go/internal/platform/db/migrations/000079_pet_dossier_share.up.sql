ALTER TABLE identity.users
  ADD COLUMN IF NOT EXISTS contact_phone TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS pets.dossier_share_tokens (
    id UUID PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    recipient_email TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT 'fr',
    commercial_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    commercial_name TEXT NOT NULL DEFAULT '',
    commercial_phone TEXT NOT NULL DEFAULT '',
    register_url TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    first_downloaded_at TIMESTAMPTZ,
    download_count INT NOT NULL DEFAULT 0,
    object_key TEXT NOT NULL DEFAULT '',
    cached_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_dossier_share_tokens_token ON pets.dossier_share_tokens(token);
CREATE INDEX IF NOT EXISTS idx_dossier_share_tokens_pet ON pets.dossier_share_tokens(pet_id);
CREATE INDEX IF NOT EXISTS idx_dossier_share_tokens_owner ON pets.dossier_share_tokens(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_dossier_share_tokens_expires ON pets.dossier_share_tokens(expires_at);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pets.dossier_share_tokens TO petsfollow_app;
  END IF;
END $$;
