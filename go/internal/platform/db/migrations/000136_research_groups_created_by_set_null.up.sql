-- Allow user purge/anonymisation without blocking on research group ownership.
ALTER TABLE research.groups
    ALTER COLUMN created_by DROP NOT NULL;

ALTER TABLE research.groups
    DROP CONSTRAINT IF EXISTS groups_created_by_fkey;

ALTER TABLE research.groups
    ADD CONSTRAINT groups_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES identity.users(id) ON DELETE SET NULL;
