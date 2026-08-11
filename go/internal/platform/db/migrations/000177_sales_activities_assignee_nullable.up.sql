-- CRM activities: allow unassigning on Pro tombstone (RGPD).
ALTER TABLE sales.activities
    ALTER COLUMN assignee_user_id DROP NOT NULL;
