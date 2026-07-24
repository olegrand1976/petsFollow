-- Multi-profils care_pro, ACL partage, GPS visites, CR visite.

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('vet', 'client', 'admin', 'commercial', 'commercial_manager', 'care_pro'));

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS professional_specialty TEXT NULL;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_specialty_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_specialty_check
    CHECK (
        professional_specialty IS NULL
        OR professional_specialty IN (
            'vet_light', 'farrier', 'physio', 'behaviorist', 'groomer', 'breeder'
        )
    );

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_care_pro_specialty;
ALTER TABLE identity.users
    ADD CONSTRAINT users_care_pro_specialty
    CHECK (
        (role = 'care_pro' AND professional_specialty IS NOT NULL)
        OR (role <> 'care_pro' AND professional_specialty IS NULL)
    );

CREATE TABLE IF NOT EXISTS practice.client_access (
    id UUID PRIMARY KEY,
    client_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    grantee_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read'
        CHECK (permission IN ('read', 'write_notes', 'full')),
    granted_by_user_id UUID NOT NULL REFERENCES identity.users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (client_user_id, grantee_user_id)
);

CREATE INDEX IF NOT EXISTS idx_client_access_grantee
    ON practice.client_access (grantee_user_id);

CREATE TABLE IF NOT EXISTS pets.pet_access (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    grantee_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'read'
        CHECK (permission IN ('read', 'write_notes', 'full')),
    granted_by_user_id UUID NOT NULL REFERENCES identity.users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (pet_id, grantee_user_id)
);

CREATE INDEX IF NOT EXISTS idx_pet_access_grantee
    ON pets.pet_access (grantee_user_id);
CREATE INDEX IF NOT EXISTS idx_pet_access_pet
    ON pets.pet_access (pet_id);

ALTER TABLE visits.visits
    ADD COLUMN IF NOT EXISTS address_text TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS lng DOUBLE PRECISION;

CREATE TABLE IF NOT EXISTS visits.visit_reports (
    id UUID PRIMARY KEY,
    visit_id UUID NOT NULL REFERENCES visits.visits(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'final')),
    body_text TEXT NOT NULL DEFAULT '',
    audio_url TEXT NOT NULL DEFAULT '',
    audio_object_key TEXT NOT NULL DEFAULT '',
    transcript_text TEXT NOT NULL DEFAULT '',
    improved_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_visit_reports_visit_author
    ON visits.visit_reports (visit_id, author_user_id);

CREATE INDEX IF NOT EXISTS idx_visit_reports_visit
    ON visits.visit_reports (visit_id);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.client_access TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON pets.pet_access TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON visits.visit_reports TO petsfollow_app;
  END IF;
END $$;
