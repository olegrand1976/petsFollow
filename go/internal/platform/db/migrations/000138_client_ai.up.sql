-- Client AI (tag dev): CR vulgarization cache + triage 24/7 sessions.

CREATE SCHEMA IF NOT EXISTS client_ai;

CREATE TABLE IF NOT EXISTS visits.visit_report_explanations (
    id UUID PRIMARY KEY,
    visit_id UUID NOT NULL REFERENCES visits.visits(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    payload_json JSONB NOT NULL,
    model TEXT NOT NULL DEFAULT '',
    refreshed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (visit_id, locale)
);

CREATE INDEX IF NOT EXISTS idx_visit_report_explanations_visit
    ON visits.visit_report_explanations (visit_id);

CREATE TABLE IF NOT EXISTS client_ai.triage_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    pet_id UUID REFERENCES pets.pets(id) ON DELETE SET NULL,
    practice_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_ai_triage_sessions_user
    ON client_ai.triage_sessions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS client_ai.triage_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES client_ai.triage_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    body TEXT NOT NULL DEFAULT '',
    level TEXT CHECK (level IS NULL OR level IN ('green', 'orange', 'red')),
    watch_signs JSONB NOT NULL DEFAULT '[]'::jsonb,
    recommended_action TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_ai_triage_messages_session
    ON client_ai.triage_messages (session_id, created_at ASC);

CREATE TABLE IF NOT EXISTS client_ai.usage_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_ai_usage_user
    ON client_ai.usage_events (user_id, created_at DESC);
