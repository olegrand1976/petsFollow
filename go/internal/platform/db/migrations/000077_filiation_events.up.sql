-- Attribution / filiation audit trail (append-only).

CREATE TABLE IF NOT EXISTS practice.filiation_events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL
        CHECK (event_type IN (
            'vet_assigned',
            'vet_unassigned',
            'client_referral',
            'practice_client_linked'
        )),
    commercial_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    vet_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    client_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    practice_id UUID REFERENCES practice.practices(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES identity.users(id) ON DELETE SET NULL,
    invite_code TEXT,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS filiation_events_created_idx
    ON practice.filiation_events (created_at DESC);

CREATE INDEX IF NOT EXISTS filiation_events_commercial_idx
    ON practice.filiation_events (commercial_user_id, created_at DESC)
    WHERE commercial_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS filiation_events_client_idx
    ON practice.filiation_events (client_user_id, created_at DESC)
    WHERE client_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS filiation_events_vet_idx
    ON practice.filiation_events (vet_user_id, created_at DESC)
    WHERE vet_user_id IS NOT NULL;
