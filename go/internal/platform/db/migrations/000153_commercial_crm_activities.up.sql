-- Commercial CRM: prospect timeline events + activities/tasks.

CREATE TABLE IF NOT EXISTS sales.prospect_events (
    id UUID PRIMARY KEY,
    prospect_id UUID NOT NULL REFERENCES sales.prospects(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    kind TEXT NOT NULL
        CHECK (kind IN ('note','call','meeting','email_sent','status_change','task_done','system')),
    body TEXT NOT NULL DEFAULT '',
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_prospect_events_prospect
    ON sales.prospect_events(prospect_id, created_at DESC);

CREATE TABLE IF NOT EXISTS sales.activities (
    id UUID PRIMARY KEY,
    prospect_id UUID NOT NULL REFERENCES sales.prospects(id) ON DELETE CASCADE,
    assignee_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    created_by UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    kind TEXT NOT NULL DEFAULT 'follow_up'
        CHECK (kind IN ('call','email','follow_up','meeting','other')),
    title TEXT NOT NULL,
    due_at TIMESTAMPTZ,
    done_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','done','cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_activities_assignee
    ON sales.activities(assignee_user_id, status, due_at);
CREATE INDEX IF NOT EXISTS idx_sales_activities_prospect
    ON sales.activities(prospect_id, status);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON
        sales.prospect_events, sales.activities
        TO petsfollow_app;
  END IF;
END $$;
