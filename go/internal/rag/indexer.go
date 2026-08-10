package rag

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// MediaOpener loads private object bytes (PHI-safe).
type MediaOpener interface {
	Open(ctx context.Context, objectKey string) (rc io.ReadCloser, contentType string, err error)
}

// Indexer chunks + embeds a document that is approved / platform-ready for indexing.
type Indexer struct {
	Store    *store.Store
	Media    MediaOpener
	Embedder gemini.Embedder
	Extract  PDFExtractor
}

// PDFExtractor turns PDF bytes into plain text.
type PDFExtractor interface {
	ExtractPlainTextFromPDF(ctx context.Context, data []byte) (string, error)
}

// IndexDocument loads source, extracts text, embeds chunks, marks ready/failed.
// Ready docs with existing chunks are skipped (idempotent). Use ForceIndexDocument to rebuild.
func (ix *Indexer) IndexDocument(ctx context.Context, docID string) (store.RAGDocument, error) {
	return ix.indexDocument(ctx, docID, false)
}

// ForceIndexDocument re-embeds even when status is ready with chunks (admin/ops reindex).
func (ix *Indexer) ForceIndexDocument(ctx context.Context, docID string) (store.RAGDocument, error) {
	return ix.indexDocument(ctx, docID, true)
}

func (ix *Indexer) indexDocument(ctx context.Context, docID string, force bool) (store.RAGDocument, error) {
	if ix == nil || ix.Store == nil || ix.Embedder == nil {
		return store.RAGDocument{}, fmt.Errorf("rag_indexer_not_configured")
	}
	doc, err := ix.Store.GetRAGDocument(ctx, docID)
	if err != nil {
		return store.RAGDocument{}, err
	}
	switch doc.Status {
	case store.RAGStatusApproved, store.RAGStatusFailed, store.RAGStatusIndexing:
		// proceed
	case store.RAGStatusReady:
		if !force && doc.ChunkCount > 0 {
			return doc, nil
		}
	default:
		return doc, fmt.Errorf("%w: %s", store.ErrRAGInvalidStatus, doc.Status)
	}

	doc, err = ix.Store.SetRAGDocumentStatus(ctx, docID, store.RAGStatusIndexing, "", -1)
	if err != nil {
		return store.RAGDocument{}, err
	}

	fail := func(code string) (store.RAGDocument, error) {
		safe := sanitizeIndexError(code)
		d, uerr := ix.Store.SetRAGDocumentStatus(ctx, docID, store.RAGStatusFailed, safe, 0)
		if uerr != nil {
			return d, uerr
		}
		return d, fmt.Errorf("%s", safe)
	}

	if ix.Media == nil || strings.TrimSpace(doc.SourceObjectKey) == "" {
		return fail("rag_missing_source")
	}
	rc, _, err := ix.Media.Open(ctx, doc.SourceObjectKey)
	if err != nil {
		log.Printf("rag: open source %s: %v", docID, err)
		return fail("rag_open_source")
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, 25<<20+1))
	if err != nil {
		log.Printf("rag: read source %s: %v", docID, err)
		return fail("rag_read_source")
	}
	if len(data) > 25<<20 {
		return fail("rag_source_too_large")
	}

	text, err := extractText(ctx, ix.Extract, doc.MimeType, data)
	if err != nil {
		log.Printf("rag: extract %s: %v", docID, err)
		return fail(stableExtractCode(err))
	}
	parts := Chunk(text, DefaultChunkSize, DefaultChunkOverlap)
	if len(parts) == 0 {
		return fail("rag_empty_text")
	}
	if len(parts) > MaxChunks {
		parts = parts[:MaxChunks]
	}

	vectors, err := ix.Embedder.EmbedTexts(ctx, parts)
	if err != nil {
		log.Printf("rag: embed %s: %v", docID, err)
		return fail("rag_embed_failed")
	}
	if len(vectors) != len(parts) {
		return fail("rag_embed_count_mismatch")
	}

	inserts := make([]store.RAGChunkInsert, len(parts))
	for i, p := range parts {
		inserts[i] = store.RAGChunkInsert{
			Ordinal:   i,
			Content:   p,
			Embedding: vectors[i],
			Metadata: map[string]any{
				"documentId": doc.ID,
				"title":      doc.Title,
				"scope":      string(doc.Scope),
			},
		}
	}
	if err := ix.Store.ReplaceRAGChunks(ctx, doc.ID, inserts); err != nil {
		log.Printf("rag: store chunks %s: %v", docID, err)
		return fail("rag_store_chunks")
	}
	return ix.Store.SetRAGDocumentStatus(ctx, doc.ID, store.RAGStatusReady, "", len(inserts))
}

func extractText(ctx context.Context, extract PDFExtractor, mime string, data []byte) (string, error) {
	mime = strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0]))
	switch mime {
	case "text/plain", "text/markdown":
		if !ValidUTF8Text(data) {
			return "", fmt.Errorf("rag_invalid_text")
		}
		return string(data), nil
	case "application/pdf":
		if !LooksLikePDF(data) {
			return "", fmt.Errorf("rag_invalid_pdf")
		}
		if extract == nil {
			return "", fmt.Errorf("rag_pdf_extractor_missing")
		}
		text, err := extract.ExtractPlainTextFromPDF(ctx, data)
		if err != nil {
			return "", fmt.Errorf("rag_pdf_extract: %w", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return "", fmt.Errorf("rag_empty_text")
		}
		return text, nil
	default:
		return "", fmt.Errorf("rag_unsupported_mime")
	}
}

func stableExtractCode(err error) string {
	if err == nil {
		return "rag_extract_failed"
	}
	msg := err.Error()
	for _, code := range []string{
		"rag_invalid_text", "rag_invalid_pdf", "rag_pdf_extractor_missing",
		"rag_empty_text", "rag_unsupported_mime",
	} {
		if strings.Contains(msg, code) {
			return code
		}
	}
	if strings.Contains(msg, "rag_pdf_extract") || strings.Contains(msg, "gemini") {
		return "rag_pdf_extract_failed"
	}
	return "rag_extract_failed"
}

func sanitizeIndexError(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return "rag_index_failed"
	}
	// Keep only stable snake_case codes in DB / API responses.
	for _, r := range code {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "rag_index_failed"
	}
	if utf8.RuneCountInString(code) > 64 {
		return "rag_index_failed"
	}
	return code
}
