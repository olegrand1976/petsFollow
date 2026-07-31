DROP TABLE IF EXISTS sales.commercial_quotas;

UPDATE identity.users
SET sponsor_user_id = NULL,
    branch_id = NULL,
    sales_rank = 0
WHERE role IN ('commercial', 'commercial_manager');

DROP INDEX IF EXISTS identity.idx_users_sponsor;
DROP INDEX IF EXISTS identity.idx_users_branch;

ALTER TABLE identity.users
    DROP COLUMN IF EXISTS sales_rank,
    DROP COLUMN IF EXISTS sponsor_user_id,
    DROP COLUMN IF EXISTS branch_id;

DROP TABLE IF EXISTS sales.branches;
