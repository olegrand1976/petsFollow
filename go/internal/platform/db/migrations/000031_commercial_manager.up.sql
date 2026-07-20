-- Commercial manager role, team hierarchy, shared directory prospects, CRM tracking fields.

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('vet', 'client', 'admin', 'commercial', 'commercial_manager'));

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS manager_user_id UUID REFERENCES identity.users(id);

CREATE INDEX IF NOT EXISTS idx_users_manager
    ON identity.users(manager_user_id)
    WHERE manager_user_id IS NOT NULL;

-- Shared BCE / directory pool: unassigned prospects.
ALTER TABLE sales.prospects
    ALTER COLUMN commercial_user_id DROP NOT NULL;

ALTER TABLE sales.prospects
    DROP CONSTRAINT IF EXISTS prospects_source_check;

ALTER TABLE sales.prospects
    ADD CONSTRAINT prospects_source_check
    CHECK (source IN ('commercial', 'vet_referral', 'directory'));

ALTER TABLE sales.prospects
    ADD COLUMN IF NOT EXISTS first_contacted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_contacted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS appointment_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS appointment_outcome TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS lost_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS converted_vet_user_id UUID REFERENCES identity.users(id);

ALTER TABLE sales.prospects
    DROP CONSTRAINT IF EXISTS prospects_appointment_outcome_check;

ALTER TABLE sales.prospects
    ADD CONSTRAINT prospects_appointment_outcome_check
    CHECK (appointment_outcome IN ('', 'scheduled', 'done', 'no_show', 'cancelled'));

CREATE INDEX IF NOT EXISTS idx_prospects_appointment_at
    ON sales.prospects(appointment_at)
    WHERE appointment_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_prospects_directory
    ON sales.prospects(source)
    WHERE source = 'directory';
