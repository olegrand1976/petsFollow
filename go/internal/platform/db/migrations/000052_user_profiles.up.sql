-- Multi-profil: plusieurs rôles par user + profil actif.

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN (
        'vet', 'client', 'admin', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    ));

CREATE TABLE IF NOT EXISTS identity.profiles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN (
        'vet', 'client', 'admin', 'commercial', 'commercial_manager',
        'care_pro', 'vet_assistant', 'secretary'
    )),
    practice_id UUID REFERENCES practice.practices(id) ON DELETE SET NULL,
    professional_specialty TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, role)
);

CREATE INDEX IF NOT EXISTS idx_profiles_user ON identity.profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_practice ON identity.profiles(practice_id)
    WHERE practice_id IS NOT NULL;

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS active_profile_id UUID REFERENCES identity.profiles(id) ON DELETE SET NULL;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Backfill one profile per existing user from current role/practice/specialty.
INSERT INTO identity.profiles (id, user_id, role, practice_id, professional_specialty, created_at)
SELECT gen_random_uuid(), u.id, u.role, u.practice_id, NULLIF(u.professional_specialty, ''), u.created_at
FROM identity.users u
WHERE NOT EXISTS (
    SELECT 1 FROM identity.profiles p WHERE p.user_id = u.id AND p.role = u.role
);

UPDATE identity.users u
SET active_profile_id = p.id
FROM identity.profiles p
WHERE p.user_id = u.id AND p.role = u.role AND u.active_profile_id IS NULL;
