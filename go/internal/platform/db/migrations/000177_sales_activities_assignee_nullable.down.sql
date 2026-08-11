-- Reverse: restore NOT NULL only when no null assignees remain.
UPDATE sales.activities
SET assignee_user_id = COALESCE(
    assignee_user_id,
    created_by,
    (SELECT id FROM identity.users WHERE role = 'admin' ORDER BY created_at LIMIT 1)
)
WHERE assignee_user_id IS NULL;

ALTER TABLE sales.activities
    ALTER COLUMN assignee_user_id SET NOT NULL;
