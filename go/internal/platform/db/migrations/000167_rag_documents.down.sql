DROP INDEX IF EXISTS rag_chunks_embedding_hnsw_idx;
DROP INDEX IF EXISTS rag_chunks_doc_ordinal_uidx;
DROP TABLE IF EXISTS rag.chunks;

DROP INDEX IF EXISTS rag_documents_sha_practice_uidx;
DROP INDEX IF EXISTS rag_documents_sha_platform_uidx;
DROP INDEX IF EXISTS rag_documents_scope_status_idx;
DROP INDEX IF EXISTS rag_documents_practice_idx;
DROP INDEX IF EXISTS rag_documents_status_idx;
DROP TABLE IF EXISTS rag.documents;

DROP TYPE IF EXISTS rag.document_status;
DROP TYPE IF EXISTS rag.document_scope;
DROP SCHEMA IF EXISTS rag;
-- Extension vector intentionally kept (may be used elsewhere).
