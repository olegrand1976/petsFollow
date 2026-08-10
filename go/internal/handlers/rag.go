package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/rag"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const (
	maxRAGUploadBytes = 20 << 20 // 20 MiB
	maxRAGTitleRunes  = 200
	ragIndexTimeout   = 5 * time.Minute
)

func (a *API) requireAiCrAdvancedEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.AiCrAdvancedEnabled {
		writeErr(w, r, http.StatusNotFound, "ai_cr_advanced_disabled", "ai_cr_advanced_disabled")
		return false
	}
	return true
}

func (a *API) registerRAGAdminRoutes(pr chi.Router) {
	pr.Get("/admin/rag/documents", a.adminListRAGDocuments)
	pr.Post("/admin/rag/documents", a.adminUploadRAGDocument)
	pr.Post("/admin/rag/documents/{id}/approve", a.adminApproveRAGDocument)
	pr.Post("/admin/rag/documents/{id}/reject", a.adminRejectRAGDocument)
	pr.Delete("/admin/rag/documents/{id}", a.adminDeleteRAGDocument)
	pr.Get("/admin/rag/documents/{id}/download", a.adminDownloadRAGDocument)
	pr.Post("/admin/rag/reindex", a.adminRAGReindex)
}

func (a *API) registerRAGPracticeRoutes(pr chi.Router) {
	pr.Get("/practices/me/rag/documents", a.practiceListRAGDocuments)
	pr.Post("/practices/me/rag/documents", a.practiceUploadRAGDocument)
	pr.Delete("/practices/me/rag/documents/{id}", a.practiceDeleteRAGDocument)
	pr.Get("/practices/me/rag/documents/{id}/download", a.practiceDownloadRAGDocument)
}

func (a *API) ragIndexer() *rag.Indexer {
	return &rag.Indexer{
		Store:    a.store,
		Media:    a.media,
		Embedder: a.effectiveRAGEmbedder(),
		Extract:  a.effectiveRAGPDFExtract(),
	}
}

func (a *API) effectiveRAGEmbedder() gemini.Embedder {
	if a.ragEmbedder != nil {
		return a.ragEmbedder
	}
	if a.gemini == nil || !a.gemini.Configured() {
		return nil
	}
	return a.gemini
}

func (a *API) effectiveRAGPDFExtract() rag.PDFExtractor {
	if a.ragPDFExtract != nil {
		return a.ragPDFExtract
	}
	if a.gemini == nil || !a.gemini.Configured() {
		return nil
	}
	return a.gemini
}

// startRAGIndexAsync indexes in the background (avoids HTTP/LB timeouts on large PDFs).
func (a *API) startRAGIndexAsync(docID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), ragIndexTimeout)
		defer cancel()
		if _, err := a.ragIndexer().IndexDocument(ctx, docID); err != nil {
			log.Printf("rag: async index %s: %v", docID, err)
		}
	}()
}

// queueRAGIndex marks indexing then kicks the background worker. Returns the updated doc.
func (a *API) queueRAGIndex(ctx context.Context, docID string) (store.RAGDocument, error) {
	if a.effectiveRAGEmbedder() == nil {
		return a.store.SetRAGDocumentStatus(ctx, docID, store.RAGStatusFailed, "gemini_not_configured", 0)
	}
	doc, err := a.store.SetRAGDocumentStatus(ctx, docID, store.RAGStatusIndexing, "", -1)
	if err != nil {
		return store.RAGDocument{}, err
	}
	a.startRAGIndexAsync(docID)
	return doc, nil
}

func (a *API) adminListRAGDocuments(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	statusRaw := strings.TrimSpace(r.URL.Query().Get("status"))
	status, okStatus := store.ParseRAGDocumentStatusFilter(statusRaw)
	if !okStatus {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	docs, err := a.store.ListAdminRAGDocuments(r.Context(), status, 200)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if docs == nil {
		docs = []store.RAGDocument{}
	}
	httpx.WriteData(w, http.StatusOK, docs)
}

func (a *API) adminUploadRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	doc, err := a.ingestRAGUpload(r, store.RAGScopePlatform, "", id.UserID, store.RAGStatusApproved)
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	queued, qerr := a.queueRAGIndex(r.Context(), doc.ID)
	if qerr != nil {
		a.writeRAGErr(w, r, qerr)
		return
	}
	if queued.Status == store.RAGStatusFailed {
		httpx.WriteData(w, http.StatusServiceUnavailable, queued)
		return
	}
	httpx.WriteData(w, http.StatusAccepted, queued)
}

func (a *API) adminApproveRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	docID := chi.URLParam(r, "id")
	doc, err := a.store.ApproveRAGDocument(r.Context(), docID, id.UserID)
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	queued, qerr := a.queueRAGIndex(r.Context(), doc.ID)
	if qerr != nil {
		a.writeRAGErr(w, r, qerr)
		return
	}
	if queued.Status == store.RAGStatusFailed {
		httpx.WriteData(w, http.StatusServiceUnavailable, queued)
		return
	}
	httpx.WriteData(w, http.StatusAccepted, queued)
}

func (a *API) adminRejectRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = httpx.DecodeJSON(r, &body)
	doc, err := a.store.RejectRAGDocument(r.Context(), chi.URLParam(r, "id"), id.UserID, body.Reason)
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	if a.media != nil && doc.SourceObjectKey != "" {
		_ = a.media.Delete(r.Context(), doc.SourceObjectKey)
		if cleared, cerr := a.store.ClearRAGDocumentSourceKey(r.Context(), doc.ID); cerr == nil {
			doc = cleared
		}
	}
	httpx.WriteData(w, http.StatusOK, doc)
}

func (a *API) adminDeleteRAGDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	doc, err := a.store.DeleteRAGDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	if a.media != nil && doc.SourceObjectKey != "" {
		_ = a.media.Delete(r.Context(), doc.SourceObjectKey)
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"id": doc.ID})
}

func (a *API) adminDownloadRAGDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	doc, err := a.store.GetRAGDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	a.streamRAGSource(w, r, doc)
}

func (a *API) adminRAGReindex(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	a.runRAGReindex(w, r)
}

func (a *API) practiceDownloadRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticeRAGAccess(w, r)
	if !ok {
		return
	}
	doc, err := a.store.GetRAGDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	if doc.Scope != store.RAGScopePractice || doc.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	a.streamRAGSource(w, r, doc)
}

func (a *API) practiceListRAGDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticeRAGAccess(w, r)
	if !ok {
		return
	}
	docs, err := a.store.ListRAGDocuments(r.Context(), store.ListRAGDocumentsFilter{
		PracticeID: id.PracticeID,
		Limit:      100,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if docs == nil {
		docs = []store.RAGDocument{}
	}
	httpx.WriteData(w, http.StatusOK, docs)
}

func (a *API) practiceUploadRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticeRAGAccess(w, r)
	if !ok {
		return
	}
	doc, err := a.ingestRAGUpload(r, store.RAGScopePractice, id.PracticeID, id.UserID, store.RAGStatusPending)
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) practiceDeleteRAGDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticeRAGAccess(w, r)
	if !ok {
		return
	}
	doc, err := a.store.GetRAGDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	if doc.Scope != store.RAGScopePractice || doc.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if doc.Status == store.RAGStatusReady || doc.Status == store.RAGStatusIndexing {
		writeErr(w, r, http.StatusConflict, "conflict", "rag_locked")
		return
	}
	deleted, err := a.store.DeleteRAGDocument(r.Context(), doc.ID)
	if err != nil {
		a.writeRAGErr(w, r, err)
		return
	}
	if a.media != nil && deleted.SourceObjectKey != "" {
		_ = a.media.Delete(r.Context(), deleted.SourceObjectKey)
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"id": deleted.ID})
}

func (a *API) requirePracticeRAGAccess(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return authx.Identity{}, false
	}
	return a.requirePracticePerm(w, r, "practice.settings")
}

func (a *API) ingestRAGUpload(r *http.Request, scope store.RAGDocumentScope, practiceID, userID string, status store.RAGDocumentStatus) (store.RAGDocument, error) {
	if a.media == nil {
		return store.RAGDocument{}, fmt.Errorf("media_not_configured")
	}
	if err := r.ParseMultipartForm(maxRAGUploadBytes + (1 << 20)); err != nil {
		return store.RAGDocument{}, errBadRequest("invalid_multipart")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return store.RAGDocument{}, errBadRequest("file_required")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxRAGUploadBytes+1))
	if err != nil {
		return store.RAGDocument{}, err
	}
	if len(data) == 0 {
		return store.RAGDocument{}, errBadRequest("empty_file")
	}
	if len(data) > maxRAGUploadBytes {
		return store.RAGDocument{}, errBadRequest("file_too_large")
	}

	filename := path.Base(strings.TrimSpace(header.Filename))
	if filename == "" || filename == "." || filename == "/" {
		filename = "document"
	}
	mime := strings.ToLower(strings.TrimSpace(strings.Split(header.Header.Get("Content-Type"), ";")[0]))
	if mime == "" || mime == "application/octet-stream" {
		mime = rag.DetectMimeFromFilename(filename)
	}
	if !rag.IsAllowedMime(mime) {
		return store.RAGDocument{}, errBadRequest("rag_unsupported_format")
	}
	switch mime {
	case "application/pdf":
		if !rag.LooksLikePDF(data) {
			return store.RAGDocument{}, errBadRequest("rag_unsupported_format")
		}
	case "text/plain", "text/markdown":
		if !rag.ValidUTF8Text(data) {
			return store.RAGDocument{}, errBadRequest("rag_unsupported_format")
		}
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = strings.TrimSuffix(filename, path.Ext(filename))
	}
	if title == "" {
		title = filename
	}
	if utf8.RuneCountInString(title) > maxRAGTitleRunes {
		runes := []rune(title)
		title = string(runes[:maxRAGTitleRunes])
	}
	if utf8.RuneCountInString(filename) > maxRAGTitleRunes {
		runes := []rune(filename)
		filename = string(runes[:maxRAGTitleRunes])
	}

	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	docID := uuid.NewString()
	ext := ".bin"
	switch mime {
	case "application/pdf":
		ext = ".pdf"
	case "text/plain":
		ext = ".txt"
	case "text/markdown":
		ext = ".md"
	}
	entity := "platform"
	if scope == store.RAGScopePractice {
		entity = practiceID
	}
	key := media.ObjectKey("rag-docs", entity+"/"+docID, ext)
	if _, err := a.media.Upload(r.Context(), key, bytes.NewReader(data), int64(len(data)), mime); err != nil {
		return store.RAGDocument{}, err
	}

	doc, err := a.store.CreateRAGDocument(r.Context(), store.CreateRAGDocumentInput{
		Scope:           scope,
		PracticeID:      practiceID,
		Title:           title,
		Filename:        filename,
		MimeType:        mime,
		ContentSHA256:   sha,
		SourceObjectKey: key,
		ByteSize:        len(data),
		Status:          status,
		UploadedBy:      userID,
	})
	if err != nil {
		_ = a.media.Delete(r.Context(), key)
		return store.RAGDocument{}, err
	}
	return doc, nil
}

func (a *API) streamRAGSource(w http.ResponseWriter, r *http.Request, doc store.RAGDocument) {
	if a.media == nil || doc.SourceObjectKey == "" {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	rc, ct, err := a.media.Open(r.Context(), doc.SourceObjectKey)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	defer rc.Close()
	if ct == "" {
		ct = doc.MimeType
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", doc.Filename))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func (a *API) runRAGReindex(w http.ResponseWriter, r *http.Request) {
	docs, err := a.store.ListRAGDocumentsForReindex(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Detach from client cancel so a browser/LB disconnect does not abort mid-batch.
	ctx, cancel := context.WithTimeout(context.Background(), ragIndexTimeout)
	defer cancel()
	okN, failN := 0, 0
	ix := a.ragIndexer()
	for _, d := range docs {
		if _, err := ix.ForceIndexDocument(ctx, d.ID); err != nil {
			failN++
			continue
		}
		okN++
	}
	httpx.WriteData(w, http.StatusOK, map[string]int{"indexed": okN, "failed": failN, "considered": len(docs)})
}

func (a *API) internalRAGReindex(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Rag-Reindex-Secret", a.cfg.RagReindexSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	a.runRAGReindex(w, r)
}

func (a *API) internalRAGSearch(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Rag-Reindex-Secret", a.cfg.RagReindexSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	var body struct {
		Query      string `json:"query"`
		PracticeID string `json:"practiceId"`
		Limit      int    `json:"limit"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || strings.TrimSpace(body.Query) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	emb := a.effectiveRAGEmbedder()
	if emb == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "gemini_not_configured")
		return
	}
	vecs, err := emb.EmbedTexts(r.Context(), []string{body.Query})
	if err != nil || len(vecs) == 0 {
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "internal")
		return
	}
	hits, err := a.store.SearchRAGChunks(r.Context(), body.PracticeID, vecs[0], body.Limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if hits == nil {
		hits = []store.RAGChunkHit{}
	}
	httpx.WriteData(w, http.StatusOK, hits)
}

type ragBadRequest string

func errBadRequest(code string) error { return ragBadRequest(code) }

func (e ragBadRequest) Error() string { return string(e) }

func (a *API) writeRAGErr(w http.ResponseWriter, r *http.Request, err error) {
	var br ragBadRequest
	switch {
	case errors.As(err, &br):
		writeErr(w, r, http.StatusBadRequest, "bad_request", string(br))
	case errors.Is(err, store.ErrRAGDuplicate):
		writeErr(w, r, http.StatusConflict, "conflict", "rag_duplicate")
	case errors.Is(err, store.ErrRAGNotPending):
		writeErr(w, r, http.StatusConflict, "conflict", "rag_not_pending")
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case strings.Contains(err.Error(), "media_not_configured"):
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "media_not_configured")
	case strings.Contains(err.Error(), "gemini_not_configured"):
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "gemini_not_configured")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}

// TestSetRAGEmbedder injects a fake embedder (integration tests).
func (a *API) TestSetRAGEmbedder(e gemini.Embedder) {
	a.ragEmbedder = e
}

// TestSetRAGPDFExtract injects a fake PDF extractor (integration tests).
func (a *API) TestSetRAGPDFExtract(e ragPDFExtractor) {
	a.ragPDFExtract = e
}
