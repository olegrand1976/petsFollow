-- Allow re-upload of the same file after a failed index (retry without admin delete).

DROP INDEX IF EXISTS rag_documents_sha_platform_uidx;
DROP INDEX IF EXISTS rag_documents_sha_practice_uidx;

CREATE UNIQUE INDEX IF NOT EXISTS rag_documents_sha_platform_uidx
    ON rag.documents (content_sha256)
    WHERE scope = 'platform' AND content_sha256 <> '' AND status NOT IN ('rejected', 'failed');

CREATE UNIQUE INDEX IF NOT EXISTS rag_documents_sha_practice_uidx
    ON rag.documents (practice_id, content_sha256)
    WHERE scope = 'practice' AND content_sha256 <> '' AND status NOT IN ('rejected', 'failed');
