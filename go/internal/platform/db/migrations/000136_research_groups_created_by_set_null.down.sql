-- Revert to RESTRICT (may fail if orphan created_by NULLs exist).
UPDATE research.groups
SET created_by = (
  SELECT m.user_id FROM research.group_members m
  WHERE m.group_id = research.groups.id AND m.member_role = 'owner'
  ORDER BY m.joined_at ASC
  LIMIT 1
)
WHERE created_by IS NULL;

ALTER TABLE research.groups
    DROP CONSTRAINT IF EXISTS groups_created_by_fkey;

ALTER TABLE research.groups
    ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE research.groups
    ADD CONSTRAINT groups_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES identity.users(id) ON DELETE RESTRICT;
