-- RAG knowledge base for AI CR advanced (pgvector).
-- Hybrid corpus: platform (admin) + practice uploads (pending until admin approve).

CREATE EXTENSION IF NOT EXISTS vector;

CREATE SCHEMA IF NOT EXISTS rag;

DO $$ BEGIN
    CREATE TYPE rag.document_scope AS ENUM ('platform', 'practice');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE rag.document_status AS ENUM (
        'pending',
        'approved',
        'rejected',
        'indexing',
        'ready',
        'failed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS rag.documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope rag.document_scope NOT NULL,
    practice_id UUID REFERENCES practice.practices (id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    filename TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT 'application/pdf',
    content_sha256 TEXT NOT NULL DEFAULT '',
    source_object_key TEXT NOT NULL DEFAULT '',
    byte_size INT NOT NULL DEFAULT 0,
    status rag.document_status NOT NULL DEFAULT 'pending',
    chunk_count INT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    reject_reason TEXT NOT NULL DEFAULT '',
    uploaded_by UUID REFERENCES identity.users (id) ON DELETE SET NULL,
    reviewed_by UUID REFERENCES identity.users (id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rag_documents_scope_practice_chk CHECK (
        (scope = 'platform' AND practice_id IS NULL)
        OR (scope = 'practice' AND practice_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS rag_documents_status_idx ON rag.documents (status);
CREATE INDEX IF NOT EXISTS rag_documents_practice_idx ON rag.documents (practice_id) WHERE practice_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS rag_documents_scope_status_idx ON rag.documents (scope, status);
CREATE UNIQUE INDEX IF NOT EXISTS rag_documents_sha_platform_uidx
    ON rag.documents (content_sha256)
    WHERE scope = 'platform' AND content_sha256 <> '' AND status NOT IN ('rejected');
CREATE UNIQUE INDEX IF NOT EXISTS rag_documents_sha_practice_uidx
    ON rag.documents (practice_id, content_sha256)
    WHERE scope = 'practice' AND content_sha256 <> '' AND status NOT IN ('rejected');

-- text-embedding-004 / embedding-001 → 768 dimensions
CREATE TABLE IF NOT EXISTS rag.chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES rag.documents (id) ON DELETE CASCADE,
    ordinal INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL DEFAULT '',
    embedding vector(768),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rag_chunks_ordinal_chk CHECK (ordinal >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS rag_chunks_doc_ordinal_uidx ON rag.chunks (document_id, ordinal);
CREATE INDEX IF NOT EXISTS rag_chunks_embedding_hnsw_idx
    ON rag.chunks USING hnsw (embedding vector_cosine_ops);

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'petsfollow_app') THEN
    GRANT USAGE ON SCHEMA rag TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON rag.documents TO petsfollow_app;
    GRANT SELECT, INSERT, UPDATE, DELETE ON rag.chunks TO petsfollow_app;
  END IF;
END $$;
