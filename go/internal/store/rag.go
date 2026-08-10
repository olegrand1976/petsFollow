package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
)

var (
	ErrRAGDuplicate     = errors.New("rag_duplicate")
	ErrRAGInvalidStatus = errors.New("rag_invalid_status")
	ErrRAGNotPending    = errors.New("rag_not_pending")
)

type RAGDocumentScope string

const (
	RAGScopePlatform RAGDocumentScope = "platform"
	RAGScopePractice RAGDocumentScope = "practice"
)

type RAGDocumentStatus string

const (
	RAGStatusPending  RAGDocumentStatus = "pending"
	RAGStatusApproved RAGDocumentStatus = "approved"
	RAGStatusRejected RAGDocumentStatus = "rejected"
	RAGStatusIndexing RAGDocumentStatus = "indexing"
	RAGStatusReady    RAGDocumentStatus = "ready"
	RAGStatusFailed   RAGDocumentStatus = "failed"
)

type RAGDocument struct {
	ID              string            `json:"id"`
	Scope           RAGDocumentScope  `json:"scope"`
	PracticeID      string            `json:"practiceId,omitempty"`
	Title           string            `json:"title"`
	Filename        string            `json:"filename"`
	MimeType        string            `json:"mimeType"`
	ContentSHA256   string            `json:"contentSha256,omitempty"`
	SourceObjectKey string            `json:"-"`
	ByteSize        int               `json:"byteSize"`
	Status          RAGDocumentStatus `json:"status"`
	ChunkCount      int               `json:"chunkCount"`
	ErrorMessage    string            `json:"errorMessage,omitempty"`
	RejectReason    string            `json:"rejectReason,omitempty"`
	UploadedBy      string            `json:"uploadedBy,omitempty"`
	ReviewedBy      string            `json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time        `json:"reviewedAt,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type RAGChunkHit struct {
	ChunkID    string           `json:"chunkId"`
	DocumentID string           `json:"documentId"`
	Title      string           `json:"title"`
	Scope      RAGDocumentScope `json:"scope"`
	PracticeID string           `json:"practiceId,omitempty"`
	Ordinal    int              `json:"ordinal"`
	Content    string           `json:"content"`
	Score      float64          `json:"score"`
	Metadata   json.RawMessage  `json:"metadata,omitempty"`
}

type CreateRAGDocumentInput struct {
	Scope           RAGDocumentScope
	PracticeID      string
	Title           string
	Filename        string
	MimeType        string
	ContentSHA256   string
	SourceObjectKey string
	ByteSize        int
	Status          RAGDocumentStatus
	UploadedBy      string
}

const ragDocumentCols = `
	id::text, scope::text, COALESCE(practice_id::text,''), title, filename, mime_type,
	content_sha256, source_object_key, byte_size, status::text, chunk_count,
	error_message, reject_reason, COALESCE(uploaded_by::text,''), COALESCE(reviewed_by::text,''),
	reviewed_at, created_at, updated_at`

func scanRAGDocument(row pgx.Row) (RAGDocument, error) {
	var d RAGDocument
	var reviewedAt *time.Time
	err := row.Scan(
		&d.ID, &d.Scope, &d.PracticeID, &d.Title, &d.Filename, &d.MimeType,
		&d.ContentSHA256, &d.SourceObjectKey, &d.ByteSize, &d.Status, &d.ChunkCount,
		&d.ErrorMessage, &d.RejectReason, &d.UploadedBy, &d.ReviewedBy,
		&reviewedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return RAGDocument{}, err
	}
	d.ReviewedAt = reviewedAt
	return d, nil
}

func (s *Store) CreateRAGDocument(ctx context.Context, in CreateRAGDocumentInput) (RAGDocument, error) {
	if in.Status == "" {
		in.Status = RAGStatusPending
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO rag.documents (
			scope, practice_id, title, filename, mime_type, content_sha256,
			source_object_key, byte_size, status, uploaded_by
		) VALUES (
			$1::rag.document_scope,
			NULLIF($2,'')::uuid,
			$3, $4, $5, $6, $7, $8,
			$9::rag.document_status,
			NULLIF($10,'')::uuid
		)
		RETURNING `+ragDocumentCols, in.Scope, in.PracticeID, in.Title, in.Filename, in.MimeType,
		in.ContentSHA256, in.SourceObjectKey, in.ByteSize, in.Status, in.UploadedBy)
	d, err := scanRAGDocument(row)
	if err != nil {
		if isUniqueViolation(err) {
			return RAGDocument{}, ErrRAGDuplicate
		}
		return RAGDocument{}, err
	}
	return d, nil
}

func (s *Store) GetRAGDocument(ctx context.Context, id string) (RAGDocument, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+ragDocumentCols+` FROM rag.documents WHERE id = $1::uuid`, id)
	d, err := scanRAGDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RAGDocument{}, ErrNotFound
	}
	return d, err
}

type ListRAGDocumentsFilter struct {
	Scope      RAGDocumentScope
	PracticeID string
	Status     RAGDocumentStatus
	Limit      int
}

func (s *Store) ListRAGDocuments(ctx context.Context, f ListRAGDocumentsFilter) ([]RAGDocument, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := `SELECT ` + ragDocumentCols + ` FROM rag.documents WHERE 1=1`
	args := []any{}
	n := 1
	if f.Scope != "" {
		q += fmt.Sprintf(` AND scope = $%d::rag.document_scope`, n)
		args = append(args, f.Scope)
		n++
	}
	if f.PracticeID != "" {
		q += fmt.Sprintf(` AND practice_id = $%d::uuid`, n)
		args = append(args, f.PracticeID)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(` AND status = $%d::rag.document_status`, n)
		args = append(args, f.Status)
		n++
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RAGDocument
	for rows.Next() {
		d, err := scanRAGDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListAdminRAGDocuments returns platform docs + practice docs awaiting moderation / all statuses.
func (s *Store) ListAdminRAGDocuments(ctx context.Context, status RAGDocumentStatus, limit int) ([]RAGDocument, error) {
	return s.ListRAGDocuments(ctx, ListRAGDocumentsFilter{Status: status, Limit: limit})
}

func (s *Store) ApproveRAGDocument(ctx context.Context, id, reviewerID string) (RAGDocument, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rag.documents
		SET status = 'approved',
		    reviewed_by = NULLIF($2,'')::uuid,
		    reviewed_at = now(),
		    reject_reason = '',
		    error_message = '',
		    updated_at = now()
		WHERE id = $1::uuid AND status = 'pending'
		RETURNING `+ragDocumentCols, id, reviewerID)
	d, err := scanRAGDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RAGDocument{}, ErrRAGNotPending
	}
	return d, err
}

func (s *Store) RejectRAGDocument(ctx context.Context, id, reviewerID, reason string) (RAGDocument, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rag.documents
		SET status = 'rejected',
		    reviewed_by = NULLIF($2,'')::uuid,
		    reviewed_at = now(),
		    reject_reason = $3,
		    updated_at = now()
		WHERE id = $1::uuid AND status = 'pending'
		RETURNING `+ragDocumentCols, id, reviewerID, strings.TrimSpace(reason))
	d, err := scanRAGDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RAGDocument{}, ErrRAGNotPending
	}
	return d, err
}

func (s *Store) SetRAGDocumentStatus(ctx context.Context, id string, status RAGDocumentStatus, errMsg string, chunkCount int) (RAGDocument, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rag.documents
		SET status = $2::rag.document_status,
		    error_message = COALESCE($3,''),
		    chunk_count = CASE WHEN $4 >= 0 THEN $4 ELSE chunk_count END,
		    updated_at = now()
		WHERE id = $1::uuid
		RETURNING `+ragDocumentCols, id, status, errMsg, chunkCount)
	d, err := scanRAGDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RAGDocument{}, ErrNotFound
	}
	return d, err
}

// ClearRAGDocumentSourceKey blanks the object key after media purge (e.g. reject).
func (s *Store) ClearRAGDocumentSourceKey(ctx context.Context, id string) (RAGDocument, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE rag.documents
		SET source_object_key = '', byte_size = 0, updated_at = now()
		WHERE id = $1::uuid
		RETURNING `+ragDocumentCols, id)
	d, err := scanRAGDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RAGDocument{}, ErrNotFound
	}
	return d, err
}

// ParseRAGDocumentStatusFilter allowlists list filters. Empty → all statuses.
func ParseRAGDocumentStatusFilter(raw string) (RAGDocumentStatus, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", true
	}
	switch RAGDocumentStatus(v) {
	case RAGStatusPending, RAGStatusApproved, RAGStatusRejected,
		RAGStatusIndexing, RAGStatusReady, RAGStatusFailed:
		return RAGDocumentStatus(v), true
	default:
		return "", false
	}
}

func (s *Store) DeleteRAGDocument(ctx context.Context, id string) (RAGDocument, error) {
	d, err := s.GetRAGDocument(ctx, id)
	if err != nil {
		return RAGDocument{}, err
	}
	_, err = s.pool.Exec(ctx, `DELETE FROM rag.documents WHERE id = $1::uuid`, id)
	return d, err
}

type RAGChunkInsert struct {
	Ordinal   int
	Content   string
	Embedding []float32
	Metadata  map[string]any
}

func formatVectorLiteral(v []float32) string {
	if len(v) == 0 {
		v = make([]float32, gemini.EmbeddingDims)
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%g", f))
	}
	b.WriteByte(']')
	return b.String()
}

// ReplaceRAGChunks deletes existing chunks and inserts the new set in a transaction.
func (s *Store) ReplaceRAGChunks(ctx context.Context, documentID string, chunks []RAGChunkInsert) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM rag.chunks WHERE document_id = $1::uuid`, documentID); err != nil {
		return err
	}
	for _, c := range chunks {
		meta, _ := json.Marshal(c.Metadata)
		if meta == nil {
			meta = []byte("{}")
		}
		if len(c.Embedding) != gemini.EmbeddingDims {
			return fmt.Errorf("rag_embedding_dims: got %d want %d", len(c.Embedding), gemini.EmbeddingDims)
		}
		vec := formatVectorLiteral(c.Embedding)
		if _, err := tx.Exec(ctx, `
			INSERT INTO rag.chunks (document_id, ordinal, content, embedding, metadata)
			VALUES ($1::uuid, $2, $3, $4::vector, $5::jsonb)`,
			documentID, c.Ordinal, c.Content, vec, meta); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SearchRAGChunks returns top-k ready chunks for platform + optional practice scope.
func (s *Store) SearchRAGChunks(ctx context.Context, practiceID string, queryEmbedding []float32, limit int) ([]RAGChunkHit, error) {
	if limit <= 0 || limit > 50 {
		limit = 8
	}
	if len(queryEmbedding) != gemini.EmbeddingDims {
		return nil, fmt.Errorf("rag_embedding_dims: got %d want %d", len(queryEmbedding), gemini.EmbeddingDims)
	}
	vec := formatVectorLiteral(queryEmbedding)
	rows, err := s.pool.Query(ctx, `
		SELECT c.id::text, c.document_id::text, d.title, d.scope::text, COALESCE(d.practice_id::text,''),
		       c.ordinal, c.content, 1 - (c.embedding <=> $1::vector) AS score, c.metadata
		FROM rag.chunks c
		JOIN rag.documents d ON d.id = c.document_id
		WHERE d.status = 'ready'
		  AND c.embedding IS NOT NULL
		  AND (
		    d.scope = 'platform'
		    OR ($2 <> '' AND d.scope = 'practice' AND d.practice_id = $2::uuid)
		  )
		ORDER BY c.embedding <=> $1::vector
		LIMIT $3`, vec, practiceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RAGChunkHit
	for rows.Next() {
		var h RAGChunkHit
		var meta []byte
		if err := rows.Scan(&h.ChunkID, &h.DocumentID, &h.Title, &h.Scope, &h.PracticeID,
			&h.Ordinal, &h.Content, &h.Score, &meta); err != nil {
			return nil, err
		}
		h.Metadata = meta
		out = append(out, h)
	}
	return out, rows.Err()
}

// ListRAGDocumentsForReindex returns docs eligible for admin/ops reindex (batch ≤ limit).
// Includes ready (force rebuild), failed/approved/indexing (repair), not pending/rejected.
func (s *Store) ListRAGDocumentsForReindex(ctx context.Context, limit int) ([]RAGDocument, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+ragDocumentCols+`
		FROM rag.documents
		WHERE status IN ('approved', 'failed', 'indexing', 'ready')
		ORDER BY
			CASE status
				WHEN 'failed' THEN 0
				WHEN 'approved' THEN 1
				WHEN 'indexing' THEN 2
				ELSE 3
			END,
			updated_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RAGDocument
	for rows.Next() {
		d, err := scanRAGDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListRAGDocumentsUploadedBy returns document metadata for RGPD export.
func (s *Store) ListRAGDocumentsUploadedBy(ctx context.Context, userID string) ([]RAGDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+ragDocumentCols+`
		FROM rag.documents WHERE uploaded_by = $1::uuid
		ORDER BY created_at DESC LIMIT 500`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RAGDocument
	for rows.Next() {
		d, err := scanRAGDocument(rows)
		if err != nil {
			return nil, err
		}
		d.SourceObjectKey = ""
		d.ContentSHA256 = ""
		out = append(out, d)
	}
	return out, rows.Err()
}
