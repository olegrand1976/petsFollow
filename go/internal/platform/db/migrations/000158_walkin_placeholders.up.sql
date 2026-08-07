-- Walk-in placeholders: one immutable "Nouveau client" + "Nouvel animal" per practice.

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS is_walkin_placeholder BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE pets.pets
    ADD COLUMN IF NOT EXISTS is_walkin_placeholder BOOLEAN NOT NULL DEFAULT false;

-- At most one walk-in client per practice (users.practice_id is set for desk-provisioned clients).
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_one_walkin_per_practice
    ON identity.users (practice_id)
    WHERE is_walkin_placeholder AND role = 'client' AND practice_id IS NOT NULL;

-- At most one walk-in pet per practice.
CREATE UNIQUE INDEX IF NOT EXISTS idx_pets_one_walkin_per_practice
    ON pets.pets (practice_id)
    WHERE is_walkin_placeholder AND practice_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_walkin_placeholder
    ON identity.users (is_walkin_placeholder)
    WHERE is_walkin_placeholder;

CREATE INDEX IF NOT EXISTS idx_pets_walkin_placeholder
    ON pets.pets (is_walkin_placeholder)
    WHERE is_walkin_placeholder;
