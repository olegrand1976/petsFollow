package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const maxCompendiumPDFBytes = 20 << 20 // 20 MiB

// CompendiumExtractFunc extracts medications from a PDF page range (tests / Gemini).
type CompendiumExtractFunc func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error)

// TestSetCompendiumExtract installs a per-API mock extractor (integration tests only).
func (a *API) TestSetCompendiumExtract(fn CompendiumExtractFunc) {
	a.compendiumExtract = fn
}

func (a *API) registerCompendiumImportRoutes(r chi.Router) {
	r.Post("/admin/compendium-imports", a.adminCreateCompendiumImport)
	r.Get("/admin/compendium-imports", a.adminListCompendiumImports)
	r.Get("/admin/compendium-imports/{id}", a.adminGetCompendiumImport)
	r.Post("/admin/compendium-imports/{id}/extract", a.adminStartCompendiumExtract)
	r.Patch("/admin/compendium-imports/{id}/rows/{rowId}", a.adminPatchCompendiumRow)
	r.Post("/admin/compendium-imports/{id}/commit", a.adminCommitCompendiumImport)
}

func (a *API) adminCreateCompendiumImport(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
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
	if pageEnd > totalPages {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "page_end_out_of_range")
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
	httpx.WriteData(w, http.StatusCreated, job)
}

func (a *API) adminListCompendiumImports(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
		return
	}
	items, err := a.store.ListCompendiumImportJobs(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) adminGetCompendiumImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
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
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) adminStartCompendiumExtract(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
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
	if job.Status != "uploaded" && job.Status != "failed" {
		writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
		return
	}
	if a.compendiumExtract == nil && (a.gemini == nil || !a.gemini.Configured()) {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "gemini_not_configured")
		return
	}
	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "media_not_configured")
		return
	}

	chunks := pharmacy.ChunkPageRanges(job.PageStart, job.PageEnd, pharmacy.CompendiumPagesPerChunk)
	if err := a.store.MarkCompendiumExtracting(r.Context(), id, len(chunks)); err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = a.store.FailCompendiumImportJob(ctx, id, fmt.Sprintf("panic:%v", rec))
			}
		}()
		a.runCompendiumExtract(id)
	}()

	detail, err := a.store.GetCompendiumImportDetail(r.Context(), id)
	if err != nil {
		httpx.WriteData(w, http.StatusAccepted, map[string]any{"status": "extracting"})
		return
	}
	httpx.WriteData(w, http.StatusAccepted, detail)
}

func (a *API) runCompendiumExtract(jobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	fail := func(msg string) {
		_ = a.store.FailCompendiumImportJob(ctx, jobID, msg)
	}

	job, err := a.store.GetCompendiumImportJob(ctx, jobID)
	if err != nil {
		fail(fmt.Sprintf("job_load:%v", err))
		return
	}
	rc, _, err := a.media.Open(ctx, job.PDFObjectKey)
	if err != nil {
		fail("pdf_open_failed")
		return
	}
	pdfBytes, err := io.ReadAll(io.LimitReader(rc, maxCompendiumPDFBytes+1))
	_ = rc.Close()
	if err != nil || len(pdfBytes) == 0 || len(pdfBytes) > maxCompendiumPDFBytes {
		fail("pdf_read_failed")
		return
	}

	// Trim each Gemini chunk directly from the full PDF (absolute pages) — no pre-trim of the whole range.
	chunks := pharmacy.ChunkPageRanges(job.PageStart, job.PageEnd, pharmacy.CompendiumPagesPerChunk)
	extractor := &pharmacy.CompendiumExtractor{Gemini: a.gemini}
	var allMeds []pharmacy.ExtractedMedication

	for i, rng := range chunks {
		start, end := rng[0], rng[1]
		slice, err := pharmacy.ExtractPageRange(pdfBytes, start, end)
		if err != nil {
			fail(fmt.Sprintf("pdf_chunk:%v", err))
			return
		}
		var meds []pharmacy.ExtractedMedication
		if a.compendiumExtract != nil {
			meds, err = a.compendiumExtract(ctx, slice, start, end)
		} else {
			meds, err = extractor.ExtractChunk(ctx, slice, start, end)
		}
		if err != nil {
			fail(fmt.Sprintf("extract:%v", err))
			return
		}
		allMeds = append(allMeds, meds...)
		_ = a.store.SetCompendiumExtractProgress(ctx, jobID, i+1)
	}

	allMeds = pharmacy.DedupExtractedMedications(allMeds)
	all := make([]store.CompendiumRowInsert, 0, len(allMeds))
	for _, m := range allMeds {
		status, code, msg := pharmacy.ClassifyExtractedRow(m)
		raw, _ := json.Marshal(m)
		sp := m.SourcePage
		if sp == nil {
			p := job.PageStart
			sp = &p
		}
		all = append(all, store.CompendiumRowInsert{
			SourcePage:         sp,
			CNK:                m.CNK,
			Name:               m.Name,
			ATCCode:            m.ATCCode,
			PharmaceuticalForm: m.PharmaceuticalForm,
			PackSize:           m.PackSize,
			IsAntibiotic:       m.IsAntibiotic,
			RawJSON:            raw,
			Status:             status,
			ErrorCode:          code,
			ErrorMessage:       msg,
		})
	}

	if err := a.store.ReplaceCompendiumExtractRows(ctx, jobID, all); err != nil {
		fail(fmt.Sprintf("persist:%v", err))
	}
}

func (a *API) adminPatchCompendiumRow(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
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

func (a *API) adminCommitCompendiumImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "pharmacy_disabled")
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
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"result": result,
		"job":    detail.Job,
	})
}
