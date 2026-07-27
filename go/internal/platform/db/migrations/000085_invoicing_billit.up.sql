-- Facturation Billit reseller (schema invoicing).
CREATE SCHEMA IF NOT EXISTS invoicing;

CREATE TABLE IF NOT EXISTS invoicing.practice_connections (
    practice_id UUID PRIMARY KEY REFERENCES practice.practices(id) ON DELETE CASCADE,
    billit_party_id TEXT,
    status TEXT NOT NULL DEFAULT 'disconnected'
        CHECK (status IN (
            'disconnected',
            'pending_registration',
            'pending_kyc',
            'active',
            'suspended',
            'error'
        )),
    invoice_to_partner BOOLEAN NOT NULL DEFAULT true,
    partner_listed_at TIMESTAMPTZ,
    docs_included_monthly INT NOT NULL DEFAULT 50,
    api_secret_ref TEXT,
    last_error TEXT,
    connected_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoicing.connect_states (
    state TEXT PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES identity.users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_invoicing_connect_states_practice
    ON invoicing.connect_states (practice_id);

CREATE TABLE IF NOT EXISTS invoicing.documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('invoice', 'credit_note', 'proforma')),
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN (
            'draft', 'issued', 'sending', 'delivered', 'rejected', 'cancelled'
        )),
    number TEXT,
    billit_order_id TEXT,
    idempotency_key TEXT NOT NULL,
    counterparty_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    currency TEXT NOT NULL DEFAULT 'EUR',
    total_excl_cents BIGINT NOT NULL DEFAULT 0,
    total_vat_cents BIGINT NOT NULL DEFAULT 0,
    total_incl_cents BIGINT NOT NULL DEFAULT 0,
    related_document_id UUID REFERENCES invoicing.documents(id),
    peppol_status TEXT,
    sent_at TIMESTAMPTZ,
    created_by UUID REFERENCES identity.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT invoicing_documents_idempotency UNIQUE (practice_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_invoicing_documents_practice_created
    ON invoicing.documents (practice_id, created_at DESC);

CREATE TABLE IF NOT EXISTS invoicing.document_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES invoicing.documents(id) ON DELETE CASCADE,
    position INT NOT NULL,
    description TEXT NOT NULL,
    quantity NUMERIC(12, 3) NOT NULL,
    unit_price_excl_cents BIGINT NOT NULL,
    vat_percent NUMERIC(5, 2) NOT NULL,
    meta_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_invoicing_document_lines_doc
    ON invoicing.document_lines (document_id, position);

CREATE TABLE IF NOT EXISTS invoicing.webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    provider TEXT NOT NULL DEFAULT 'billit',
    event_type TEXT,
    external_id TEXT,
    payload JSONB NOT NULL,
    processed_at TIMESTAMPTZ,
    error TEXT
);

CREATE TABLE IF NOT EXISTS invoicing.usage_monthly (
    practice_id UUID NOT NULL REFERENCES practice.practices(id) ON DELETE CASCADE,
    yyyymm INT NOT NULL,
    doc_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (practice_id, yyyymm)
);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA invoicing TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA invoicing TO petsfollow_app;
  END IF;
END $$;
