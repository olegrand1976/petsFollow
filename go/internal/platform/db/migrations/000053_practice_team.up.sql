-- Équipe cabinet : véto de référence + membres + permissions.

ALTER TABLE practice.practices
    ADD COLUMN IF NOT EXISTS reference_vet_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL;

UPDATE practice.practices p
SET reference_vet_user_id = u.id
FROM identity.users u
WHERE u.practice_id = p.id AND u.role = 'vet' AND p.reference_vet_user_id IS NULL;

CREATE TABLE IF NOT EXISTS practice.team_members (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    profile_id UUID REFERENCES identity.profiles(id) ON DELETE SET NULL,
    team_role TEXT NOT NULL CHECK (team_role IN ('reference_vet', 'vet', 'vet_assistant', 'secretary')),
    permissions JSONB,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'invited', 'revoked')),
    invited_by_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (practice_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_practice ON practice.team_members(practice_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user ON practice.team_members(user_id);

-- Seed reference_vet membership for existing practices.
INSERT INTO practice.team_members (id, practice_id, user_id, profile_id, team_role, permissions, status, created_at)
SELECT gen_random_uuid(), p.id, p.reference_vet_user_id, pr.id, 'reference_vet', NULL, 'active', NOW()
FROM practice.practices p
JOIN identity.profiles pr ON pr.user_id = p.reference_vet_user_id AND pr.role = 'vet'
WHERE p.reference_vet_user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM practice.team_members tm
    WHERE tm.practice_id = p.id AND tm.user_id = p.reference_vet_user_id
  );
