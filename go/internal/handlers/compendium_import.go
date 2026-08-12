package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const maxCompendiumPDFBytes = 20 << 20 // 20 MiB

// testCompendiumExtract overrides Gemini extract in integration tests.
var testCompendiumExtract func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error)

// compendiumExtractRunning prevents double goroutines for the same job on one API process.
var compendiumExtractRunning sync.Map

// TestSetCompendiumExtract installs a mock extractor (integration tests only).
func TestSetCompendiumExtract(fn func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error)) {
	testCompendiumExtract = fn
}

// TestClearCompendiumExtract clears the mock extractor.
func TestClearCompendiumExtract() { testCompendiumExtract = nil }

func (a *API) registerCompendiumImportRoutes(r chi.Router) {
	r.Post("/admin/compendium-imports", a.adminCreateCompendiumImport)
	r.Get("/admin/compendium-imports", a.adminListCompendiumImports)
	r.Get("/admin/compendium-imports/{id}", a.adminGetCompendiumImport)
	r.Get("/admin/compendium-imports/{id}/pdf", a.adminGetCompendiumImportPDF)
	r.Delete("/admin/compendium-imports/{id}", a.adminDeleteCompendiumImport)
	r.Post("/admin/compendium-imports/{id}/extract", a.adminStartCompendiumExtract)
	r.Patch("/admin/compendium-imports/{id}/rows/{rowId}", a.adminPatchCompendiumRow)
	r.Post("/admin/compendium-imports/{id}/rows/{rowId}/lookup-cnk", a.adminLookupCompendiumRowCNK)
	r.Post("/admin/compendium-imports/{id}/confirm-ready", a.adminConfirmCompendiumReady)
	r.Post("/admin/compendium-imports/{id}/commit", a.adminCommitCompendiumImport)
}

func (a *API) requireCompendiumAdmin(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	admin, ok := a.requireAdmin(w, r)
	if !ok {
		return authx.Identity{}, false
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
		return authx.Identity{}, false
	}
	return admin, true
}

func scrubCompendiumJob(j *store.CompendiumImportJob) {
	if j != nil {
		j.PDFObjectKey = ""
	}
}

func scrubCompendiumDetail(d *store.CompendiumImportDetail) {
	if d == nil {
		return
	}
	scrubCompendiumJob(&d.Job)
}

func (a *API) adminCreateCompendiumImport(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireCompendiumAdmin(w, r)
	if !ok {
		return
	}
	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "media_not_configured")
		return
	}
	if err := r.ParseMultipartForm(maxCompendiumPDFBytes + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_multipart")
		return
	}
	pageStart, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("pageStart")))
	pageEnd, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("pageEnd")))
	if pageStart < 1 || pageEnd < pageStart {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_page_range")
		return
	}
	if pageEnd-pageStart >= 200 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "page_range_too_large")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxCompendiumPDFBytes+1))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_read_failed")
		return
	}
	if len(data) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "empty_file")
		return
	}
	if len(data) > maxCompendiumPDFBytes {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_too_large")
		return
	}
	if !pharmacy.LooksLikePDF(data) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "unsupported_format")
		return
	}
	totalPages, err := pharmacy.PageCount(data)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_pdf")
		return
	}
	// Heuristic page count can under-count; only reject when clearly impossible.
	if totalPages > 0 && pageStart > totalPages+50 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "page_start_out_of_range")
		return
	}

	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		filename = "compendium.pdf"
	}
	jobID := uuid.NewString()
	key := fmt.Sprintf("compendium-imports/%s.pdf", jobID)
	if _, err := a.media.Upload(r.Context(), key, bytes.NewReader(data), int64(len(data)), "application/pdf"); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "media_upload_failed")
		return
	}

	job, err := a.store.CreateCompendiumImportJob(r.Context(), store.CreateCompendiumImportInput{
		ID:               jobID,
		CreatedByAdminID: admin.UserID,
		Filename:         filename,
		ContentType:      "application/pdf",
		PageStart:        pageStart,
		PageEnd:          pageEnd,
		PDFObjectKey:     key,
	})
	if err != nil {
		_ = a.media.Delete(r.Context(), key)
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	scrubCompendiumJob(&job)
	httpx.WriteData(w, http.StatusCreated, job)
}

func (a *API) adminListCompendiumImports(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	items, err := a.store.ListCompendiumImportJobs(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	for i := range items {
		scrubCompendiumJob(&items[i])
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) adminGetCompendiumImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	detail, err := a.store.GetCompendiumImportDetail(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	scrubCompendiumDetail(&detail)
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) adminGetCompendiumImportPDF(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	job, err := a.store.GetCompendiumImportJob(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.media == nil || strings.TrimSpace(job.PDFObjectKey) == "" {
		writeErr(w, r, http.StatusNotFound, "not_found", "pdf_missing")
		return
	}
	rc, ct, err := a.media.Open(r.Context(), job.PDFObjectKey)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pdf_missing")
		return
	}
	defer rc.Close()
	if ct == "" {
		ct = "application/pdf"
	}
	fname := filepath.Base(strings.TrimSpace(job.Filename))
	if fname == "" || fname == "." || fname == "/" {
		fname = "compendium.pdf"
	}
	if !strings.HasSuffix(strings.ToLower(fname), ".pdf") {
		fname += ".pdf"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename=%q`, fname))
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = io.Copy(w, rc)
}

func (a *API) adminDeleteCompendiumImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	key, err := a.store.DeleteCompendiumImportJob(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.media != nil && strings.TrimSpace(key) != "" {
		if err := a.media.Delete(r.Context(), key); err != nil {
			log.Printf("compendium delete %s: media cleanup %s: %v", id, key, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminStartCompendiumExtract(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	job, err := a.store.GetCompendiumImportJob(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// uploaded | failed | extracting (stale recovery only — see BeginCompendiumExtract)
	if job.Status != "uploaded" && job.Status != "failed" && job.Status != "extracting" {
		writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
		return
	}
	if testCompendiumExtract == nil && (a.gemini == nil || !a.gemini.Configured()) {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "gemini_not_configured")
		return
	}
	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "media_not_configured")
		return
	}

	chunks := pharmacy.ChunkPageRanges(job.PageStart, job.PageEnd, pharmacy.CompendiumPagesPerChunk)
	if len(chunks) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_page_range")
		return
	}
	forceRestart := r.URL.Query().Get("restart") == "1" || r.URL.Query().Get("restart") == "true"
	if _, loaded := compendiumExtractRunning.LoadOrStore(id, struct{}{}); loaded {
		writeErr(w, r, http.StatusConflict, "conflict", "extract_already_running")
		return
	}
	begin, err := a.store.BeginCompendiumExtract(r.Context(), id, len(chunks), forceRestart)
	if err != nil {
		compendiumExtractRunning.Delete(id)
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	go a.runCompendiumExtract(id, begin.ResumeFromChunk)

	detail, err := a.store.GetCompendiumImportDetail(r.Context(), id)
	if err != nil {
		httpx.WriteData(w, http.StatusAccepted, map[string]any{"status": "extracting", "resumeFromChunk": begin.ResumeFromChunk})
		return
	}
	scrubCompendiumDetail(&detail)
	httpx.WriteData(w, http.StatusAccepted, detail)
}

func (a *API) runCompendiumExtract(jobID string, resumeFromChunk int) {
	defer compendiumExtractRunning.Delete(jobID)

	job, err := a.store.GetCompendiumImportJob(context.Background(), jobID)
	if err != nil {
		return
	}
	chunks := pharmacy.ChunkPageRanges(job.PageStart, job.PageEnd, pharmacy.CompendiumPagesPerChunk)
	if len(chunks) == 0 {
		_ = a.store.FailCompendiumImportJob(context.Background(), jobID, "invalid_page_range")
		return
	}
	if resumeFromChunk < 0 {
		resumeFromChunk = 0
	}
	if resumeFromChunk > len(chunks) {
		resumeFromChunk = len(chunks)
	}
	// Budget: media HTTP timeout is 5 min/chunk — keep headroom for trim + CNK match + persist.
	// Max UI range is 200 pages → ≤34 chunks → ~3 h worst case.
	remaining := len(chunks) - resumeFromChunk
	timeout := min(10*time.Minute+time.Duration(remaining)*5*time.Minute, 3*time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// Persist failure even if the run ctx already timed out / cancelled.
	failCtx := context.WithoutCancel(ctx)

	rc, _, err := a.media.Open(ctx, job.PDFObjectKey)
	if err != nil {
		log.Printf("compendium extract %s: pdf open: %v", jobID, err)
		_ = a.store.FailCompendiumImportJob(failCtx, jobID, "pdf_open_failed")
		return
	}
	pdfBytes, err := io.ReadAll(io.LimitReader(rc, maxCompendiumPDFBytes+1))
	_ = rc.Close()
	if err != nil || len(pdfBytes) == 0 || len(pdfBytes) > maxCompendiumPDFBytes {
		log.Printf("compendium extract %s: pdf read err=%v len=%d", jobID, err, len(pdfBytes))
		_ = a.store.FailCompendiumImportJob(failCtx, jobID, "pdf_read_failed")
		return
	}

	extractor := &pharmacy.CompendiumExtractor{Gemini: a.gemini}
	for i := resumeFromChunk; i < len(chunks); i++ {
		ch := chunks[i]
		start, end := ch[0], ch[1]
		var chunkMeds []pharmacy.ExtractedMedication
		if testCompendiumExtract != nil {
			chunkMeds, err = testCompendiumExtract(ctx, pdfBytes, start, end)
		} else {
			chunkPDF, trimErr := pharmacy.ExtractPDFPages(pdfBytes, start, end)
			if trimErr != nil {
				log.Printf("compendium extract %s: trim p%d-%d: %v", jobID, start, end, trimErr)
				_ = a.store.FailCompendiumImportJob(failCtx, jobID, "trim_failed")
				return
			}
			chunkMeds, err = extractor.ExtractChunk(ctx, chunkPDF, start, end)
		}
		if err != nil {
			log.Printf("compendium extract %s: gemini chunk %d/%d p%d-%d: %v", jobID, i+1, len(chunks), start, end, err)
			msg := fmt.Sprintf("extract_failed chunk %d/%d p%d-%d: %v", i+1, len(chunks), start, end, err)
			if len(msg) > 400 {
				msg = msg[:400] + "…"
			}
			_ = a.store.FailCompendiumImportJob(failCtx, jobID, msg)
			return
		}
		if len(chunkMeds) == 0 {
			log.Printf("compendium extract %s: empty chunk %d/%d p%d-%d (advancing)", jobID, i+1, len(chunks), start, end)
		}
		// Gemini / PDF overlap can emit the same CNK (or nameless-CNK name) twice in a chunk.
		chunkMeds = pharmacy.DedupExtractedMedications(chunkMeds)
		// Fallback sourcePage = chunk start (not job.PageStart) so mid-PDF rows keep a useful hint.
		inserts := a.compendiumRowsFromMeds(ctx, jobID, start, chunkMeds)
		if err := a.store.AppendCompendiumExtractRows(ctx, jobID, inserts); err != nil {
			log.Printf("compendium extract %s: persist chunk %d: %v", jobID, i+1, err)
			_ = a.store.FailCompendiumImportJob(failCtx, jobID, "persist_failed")
			return
		}
		_ = a.store.SetCompendiumExtractProgress(ctx, jobID, i+1)
	}

	if err := a.store.FinalizeCompendiumExtract(ctx, jobID); err != nil {
		log.Printf("compendium extract %s: finalize: %v", jobID, err)
		_ = a.store.FailCompendiumImportJob(failCtx, jobID, "persist_failed")
	}
}

func (a *API) compendiumRowsFromMeds(ctx context.Context, jobID string, fallbackPage int, meds []pharmacy.ExtractedMedication) []store.CompendiumRowInsert {
	all := make([]store.CompendiumRowInsert, 0, len(meds))
	for _, m := range meds {
		match := pharmacy.CNKMatchResult{}
		if strings.TrimSpace(m.CNK) == "" && strings.TrimSpace(m.Name) != "" {
			hits, searchErr := a.store.SearchRefMedications(ctx, m.Name, 10)
			if searchErr != nil {
				log.Printf("compendium extract %s: cnk search: %v", jobID, searchErr)
			} else {
				refs := make([]pharmacy.RefMedMatchInput, 0, len(hits))
				for _, h := range hits {
					refs = append(refs, pharmacy.RefMedMatchInput{
						CNK:                h.CNK,
						Name:               h.Name,
						PharmaceuticalForm: h.PharmaceuticalForm,
						PackSize:           h.PackSize,
						Manufacturer:       manufacturerFromAFMPSMeta(h.AFMPSMeta),
					})
				}
				match = pharmacy.SuggestCNK(m, refs)
			}
		}
		classified, status, code, msg := pharmacy.ClassifyAfterMatch(m, match)
		raw, _ := json.Marshal(classified)
		candsRaw, _ := json.Marshal(match.Candidates)
		if len(match.Candidates) == 0 {
			candsRaw = []byte("[]")
		}
		sp := classified.SourcePage
		if sp == nil {
			p := fallbackPage
			sp = &p
		}
		var scorePtr *float64
		if match.Score > 0 || match.SuggestedCNK != "" {
			sc := match.Score
			scorePtr = &sc
		}
		all = append(all, store.CompendiumRowInsert{
			SourcePage:         sp,
			CNK:                classified.CNK,
			Name:               classified.Name,
			Manufacturer:       classified.Manufacturer,
			ActiveSubstance:    classified.ActiveSubstance,
			Strength:           classified.Strength,
			ATCCode:            classified.ATCCode,
			PharmaceuticalForm: classified.PharmaceuticalForm,
			PackSize:           classified.PackSize,
			IsAntibiotic:       classified.IsAntibiotic,
			SuggestedCNK:       match.SuggestedCNK,
			MatchScore:         scorePtr,
			MatchCandidates:    candsRaw,
			RawJSON:            raw,
			Status:             status,
			ErrorCode:          code,
			ErrorMessage:       msg,
		})
	}
	return all
}

func manufacturerFromAFMPSMeta(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var meta map[string]any
	if err := json.Unmarshal(raw, &meta); err != nil {
		return ""
	}
	for _, k := range []string{"manufacturer", "lab", "labo", "holder", "titulaire"} {
		if v, ok := meta[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (a *API) adminPatchCompendiumRow(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	var body store.PatchCompendiumRowInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	row, err := a.store.PatchCompendiumImportRow(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "rowId"), body)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "row_locked")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, row)
}

func (a *API) adminLookupCompendiumRowCNK(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	jobID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")

	var body struct {
		CNK *string `json:"cnk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}

	verifyCNK := ""
	if body.CNK != nil {
		verifyCNK = strings.TrimSpace(*body.CNK)
	}

	result, err := a.store.LookupCompendiumImportRowCNK(r.Context(), jobID, rowID, verifyCNK)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "row_locked")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	payload := map[string]any{"row": result.Row}
	if result.CNKFound != nil {
		payload["cnkFound"] = *result.CNKFound
	}
	httpx.WriteData(w, http.StatusOK, payload)
}

func (a *API) adminConfirmCompendiumReady(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	n, err := a.store.ConfirmCompendiumPendingRows(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	detail, _ := a.store.GetCompendiumImportDetail(r.Context(), chi.URLParam(r, "id"))
	scrubCompendiumDetail(&detail)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"confirmed": n,
		"job":       detail.Job,
		"rows":      detail.Rows,
	})
}

func (a *API) adminCommitCompendiumImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireCompendiumAdmin(w, r); !ok {
		return
	}
	result, err := a.store.CommitCompendiumImport(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	detail, _ := a.store.GetCompendiumImportDetail(r.Context(), chi.URLParam(r, "id"))
	scrubCompendiumDetail(&detail)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"result": result,
		"job":    detail.Job,
	})
}
