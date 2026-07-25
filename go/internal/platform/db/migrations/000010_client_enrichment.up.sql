-- Client enrichment: care, visits, discovery, push prefs

CREATE INDEX IF NOT EXISTS idx_practice_clients_client ON practice.practice_clients(client_user_id);

CREATE SCHEMA IF NOT EXISTS care;

CREATE TABLE IF NOT EXISTS care.reminders (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    type TEXT NOT NULL CHECK (type IN ('vaccination', 'deworming', 'vet_check', 'dental', 'farrier', 'fecal_egg', 'custom')),
    title TEXT NOT NULL,
    due_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_care_reminders_pet ON care.reminders(pet_id);
CREATE INDEX IF NOT EXISTS idx_care_reminders_due ON care.reminders(due_at) WHERE status = 'pending';

CREATE SCHEMA IF NOT EXISTS visits;

CREATE TABLE IF NOT EXISTS visits.visits (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    scheduled_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'requested' CHECK (status IN ('requested', 'confirmed', 'done', 'cancelled')),
    notes TEXT,
    source TEXT NOT NULL CHECK (source IN ('client', 'vet')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visits_pet ON visits.visits(pet_id);
CREATE INDEX IF NOT EXISTS idx_visits_scheduled ON visits.visits(scheduled_at) WHERE status IN ('requested', 'confirmed');

CREATE SCHEMA IF NOT EXISTS discovery;

CREATE TABLE IF NOT EXISTS discovery.progress (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_cards JSONB NOT NULL DEFAULT '[]',
    streak_days INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications.device_tokens (
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, token)
);

CREATE TABLE IF NOT EXISTS notifications.client_preferences (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    hr BOOLEAN NOT NULL DEFAULT TRUE,
    care BOOLEAN NOT NULL DEFAULT TRUE,
    visits BOOLEAN NOT NULL DEFAULT TRUE,
    messages BOOLEAN NOT NULL DEFAULT TRUE,
    discovery BOOLEAN NOT NULL DEFAULT TRUE,
    billing BOOLEAN NOT NULL DEFAULT TRUE
);
