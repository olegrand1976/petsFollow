-- MLM org prep: sales branches + sponsor chain + rank (no multi-level commission engine yet).

CREATE TABLE IF NOT EXISTS sales.branches (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    external_mlm_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sales_branches_code_unique UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_sales_branches_code ON sales.branches(code);

ALTER TABLE identity.users
    ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES sales.branches(id),
    ADD COLUMN IF NOT EXISTS sponsor_user_id UUID REFERENCES identity.users(id),
    ADD COLUMN IF NOT EXISTS sales_rank INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_users_branch
    ON identity.users(branch_id)
    WHERE branch_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_sponsor
    ON identity.users(sponsor_user_id)
    WHERE sponsor_user_id IS NOT NULL;

-- Align sponsor with existing manager (depth-1 upline) when unset.
UPDATE identity.users
SET sponsor_user_id = manager_user_id
WHERE manager_user_id IS NOT NULL
  AND sponsor_user_id IS NULL
  AND role = 'commercial';

-- Monthly targets for branch / rep coaching (SFM now, MLM volume later).
CREATE TABLE IF NOT EXISTS sales.commercial_quotas (
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    period_ym TEXT NOT NULL,
    target_activations INT NOT NULL DEFAULT 0,
    target_earned_cents INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, period_ym)
);
