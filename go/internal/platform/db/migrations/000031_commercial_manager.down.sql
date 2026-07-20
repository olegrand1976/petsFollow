DROP INDEX IF EXISTS sales.idx_prospects_directory;
DROP INDEX IF EXISTS sales.idx_prospects_appointment_at;

ALTER TABLE sales.prospects
    DROP CONSTRAINT IF EXISTS prospects_appointment_outcome_check;

ALTER TABLE sales.prospects
    DROP COLUMN IF EXISTS converted_vet_user_id,
    DROP COLUMN IF EXISTS lost_reason,
    DROP COLUMN IF EXISTS appointment_outcome,
    DROP COLUMN IF EXISTS appointment_at,
    DROP COLUMN IF EXISTS last_contacted_at,
    DROP COLUMN IF EXISTS first_contacted_at;

-- Re-assign orphan directory rows before restoring NOT NULL.
UPDATE sales.prospects p
SET commercial_user_id = (
    SELECT u.id FROM identity.users u WHERE u.role = 'commercial' ORDER BY u.created_at LIMIT 1
)
WHERE p.commercial_user_id IS NULL;

UPDATE sales.prospects SET source = 'commercial' WHERE source = 'directory';

ALTER TABLE sales.prospects
    DROP CONSTRAINT IF EXISTS prospects_source_check;

ALTER TABLE sales.prospects
    ADD CONSTRAINT prospects_source_check
    CHECK (source IN ('commercial', 'vet_referral'));

ALTER TABLE sales.prospects
    ALTER COLUMN commercial_user_id SET NOT NULL;

DROP INDEX IF EXISTS identity.idx_users_manager;

ALTER TABLE identity.users
    DROP COLUMN IF EXISTS manager_user_id;

ALTER TABLE identity.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE identity.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('vet', 'client', 'admin', 'commercial'));
