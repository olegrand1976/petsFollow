-- Pet dossier documents (PDF / images) uploaded by vet or client.

CREATE TABLE IF NOT EXISTS pets.documents (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    uploaded_by_user_id UUID NOT NULL REFERENCES identity.users(id),
    title TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    file_url TEXT NOT NULL,
    object_key TEXT NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pets_documents_pet_id
    ON pets.documents (pet_id, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON pets.documents TO petsfollow_app;
  END IF;
END $$;
