-- petsFollow schema bootstrap
CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS practice;
CREATE SCHEMA IF NOT EXISTS pets;
CREATE SCHEMA IF NOT EXISTS heartrate;
CREATE SCHEMA IF NOT EXISTS messaging;
CREATE SCHEMA IF NOT EXISTS notifications;

CREATE TABLE IF NOT EXISTS identity.users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('vet', 'client', 'admin')),
    practice_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS practice.practices (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS practice.practice_clients (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    client_user_id UUID NOT NULL REFERENCES identity.users(id),
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (practice_id, client_user_id)
);

CREATE TABLE IF NOT EXISTS practice.invitations (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    code TEXT NOT NULL UNIQUE,
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pets.pets (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    owner_user_id UUID NOT NULL REFERENCES identity.users(id),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    breed TEXT,
    birth_date DATE,
    weight_kg NUMERIC(5,2),
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pets.dossier_events (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES identity.users(id),
    event_type TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS heartrate.sessions (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets.pets(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES identity.users(id),
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    status TEXT NOT NULL CHECK (status IN ('in_progress', 'pending_validation', 'validated', 'cancelled')),
    tap_count INT NOT NULL DEFAULT 0,
    duration_sec INT NOT NULL DEFAULT 60,
    bpm INT,
    is_alert BOOLEAN NOT NULL DEFAULT FALSE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    validated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS messaging.threads (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    client_user_id UUID NOT NULL REFERENCES identity.users(id),
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    pet_id UUID REFERENCES pets.pets(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (practice_id, client_user_id)
);

CREATE TABLE IF NOT EXISTS messaging.messages (
    id UUID PRIMARY KEY,
    thread_id UUID NOT NULL REFERENCES messaging.threads(id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL REFERENCES identity.users(id),
    body TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messaging.vet_availability (
    vet_user_id UUID PRIMARY KEY REFERENCES identity.users(id),
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    status TEXT NOT NULL CHECK (status IN ('available', 'unavailable')),
    auto_reply TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications.notification_preferences (
    vet_user_id UUID PRIMARY KEY REFERENCES identity.users(id),
    email_on_message BOOLEAN NOT NULL DEFAULT TRUE,
    email_on_heartrate BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS notifications.notification_log (
    id UUID PRIMARY KEY,
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    kind TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pets_owner ON pets.pets(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_pets_practice ON pets.pets(practice_id);
CREATE INDEX IF NOT EXISTS idx_heartrate_pet ON heartrate.sessions(pet_id);
CREATE INDEX IF NOT EXISTS idx_messages_thread ON messaging.messages(thread_id);
