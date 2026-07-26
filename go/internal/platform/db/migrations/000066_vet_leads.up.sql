-- Client-suggested veterinarians not yet on petsFollow (email + phone lead).

CREATE TABLE IF NOT EXISTS practice.vet_leads (
    id UUID PRIMARY KEY,
    client_user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    phone TEXT NOT NULL,
    full_name TEXT NOT NULL DEFAULT '',
    practice_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'contacted', 'converted', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vet_leads_client ON practice.vet_leads(client_user_id);
CREATE INDEX IF NOT EXISTS idx_vet_leads_email ON practice.vet_leads(lower(email));
CREATE INDEX IF NOT EXISTS idx_vet_leads_status ON practice.vet_leads(status);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON practice.vet_leads TO petsfollow_app;
  END IF;
END $$;
