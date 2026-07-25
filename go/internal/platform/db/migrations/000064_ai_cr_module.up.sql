-- Module CR IA (VetPro) : entitlement trial 90j, usage, feedback, friction alerts.

CREATE TABLE IF NOT EXISTS practice.ai_cr_modules (
    practice_id UUID PRIMARY KEY REFERENCES practice.practices(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'trial'
        CHECK (status IN ('trial', 'active', 'expired', 'disabled')),
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trial_ends_at TIMESTAMPTZ NOT NULL,
    activated_by_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    converted_at TIMESTAMPTZ,
    price_plan TEXT NOT NULL DEFAULT 'monthly_39'
        CHECK (price_plan IN ('monthly_39', 'annual_390')),
    baseline_minutes_per_cr INT NOT NULL DEFAULT 10 CHECK (baseline_minutes_per_cr > 0 AND baseline_minutes_per_cr <= 120),
    hourly_cost_cents INT NOT NULL DEFAULT 8000 CHECK (hourly_cost_cents >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS practice.ai_cr_usage_events (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    visit_id UUID REFERENCES visits.visits(id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK (kind IN ('transcribe', 'improve', 'finalize', 'error')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_cr_usage_practice_created
    ON practice.ai_cr_usage_events (practice_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_cr_usage_practice_kind
    ON practice.ai_cr_usage_events (practice_id, kind, created_at DESC);

CREATE TABLE IF NOT EXISTS practice.ai_cr_feedback (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    nps INT NOT NULL CHECK (nps >= 0 AND nps <= 10),
    comment TEXT NOT NULL DEFAULT '',
    friction_tags TEXT[] NOT NULL DEFAULT '{}',
    source TEXT NOT NULL DEFAULT 'in_app'
        CHECK (source IN ('j14', 'j45', 'j75', 'in_app')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_cr_feedback_practice
    ON practice.ai_cr_feedback (practice_id, created_at DESC);

CREATE TABLE IF NOT EXISTS practice.ai_cr_friction_alerts (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    commercial_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    signal TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_cr_friction_practice_signal
    ON practice.ai_cr_friction_alerts (practice_id, signal, created_at DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.ai_cr_modules TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.ai_cr_usage_events TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.ai_cr_feedback TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.ai_cr_friction_alerts TO petsfollow_app;
  END IF;
END $$;
