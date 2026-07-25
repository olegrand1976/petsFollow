CREATE TABLE IF NOT EXISTS identity.password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_token ON identity.password_reset_tokens(token);
CREATE INDEX IF NOT EXISTS idx_password_reset_user ON identity.password_reset_tokens(user_id);
