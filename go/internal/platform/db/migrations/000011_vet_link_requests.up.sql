CREATE TABLE IF NOT EXISTS practice.client_vet_link_requests (
    id UUID PRIMARY KEY,
    client_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    practice_id UUID NOT NULL REFERENCES practice.practices(id),
    vet_user_id UUID NOT NULL REFERENCES identity.users(id),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (client_user_id, practice_id)
);

CREATE INDEX IF NOT EXISTS idx_vet_link_req_vet_pending
    ON practice.client_vet_link_requests (vet_user_id) WHERE status = 'pending';
