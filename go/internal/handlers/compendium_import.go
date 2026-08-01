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
	r.Delete("/admin/compendium-imports/{id}", a.adminDeleteCompendiumImport)
	r.Post("/admin/compendium-imports/{id}/extract", a.adminStartCompendiumExtract)
	r.Patch("/admin/compendium-imports/{id}/rows/{rowId}", a.adminPatchCompendiumRow)
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
	// uploaded | failed | extracting (recovery if process died mid-job)
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

	// One Gemini call for the whole page range (full PDF + prompt bounds).
	if err := a.store.MarkCompendiumExtracting(r.Context(), id, 1); err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "invalid_status")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	go a.runCompendiumExtract(id)

	detail, err := a.store.GetCompendiumImportDetail(r.Context(), id)
	if err != nil {
		httpx.WriteData(w, http.StatusAccepted, map[string]any{"status": "extracting"})
		return
	}
	scrubCompendiumDetail(&detail)
	httpx.WriteData(w, http.StatusAccepted, detail)
}

func (a *API) runCompendiumExtract(jobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	job, err := a.store.GetCompendiumImportJob(ctx, jobID)
	if err != nil {
		return
	}
	rc, _, err := a.media.Open(ctx, job.PDFObjectKey)
	if err != nil {
		log.Printf("compendium extract %s: pdf open: %v", jobID, err)
		_ = a.store.FailCompendiumImportJob(ctx, jobID, "pdf_open_failed")
		return
	}
	pdfBytes, err := io.ReadAll(io.LimitReader(rc, maxCompendiumPDFBytes+1))
	_ = rc.Close()
	if err != nil || len(pdfBytes) == 0 || len(pdfBytes) > maxCompendiumPDFBytes {
		log.Printf("compendium extract %s: pdf read err=%v len=%d", jobID, err, len(pdfBytes))
		_ = a.store.FailCompendiumImportJob(ctx, jobID, "pdf_read_failed")
		return
	}

	extractor := &pharmacy.CompendiumExtractor{Gemini: a.gemini}
	var meds []pharmacy.ExtractedMedication
	if testCompendiumExtract != nil {
		meds, err = testCompendiumExtract(ctx, pdfBytes, job.PageStart, job.PageEnd)
	} else {
		meds, err = extractor.ExtractChunk(ctx, pdfBytes, job.PageStart, job.PageEnd)
	}
	if err != nil {
		log.Printf("compendium extract %s: gemini: %v", jobID, err)
		_ = a.store.FailCompendiumImportJob(ctx, jobID, "extract_failed")
		return
	}
	_ = a.store.SetCompendiumExtractProgress(ctx, jobID, 1)

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
			p := job.PageStart
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

	if err := a.store.ReplaceCompendiumExtractRows(ctx, jobID, all); err != nil {
		log.Printf("compendium extract %s: persist: %v", jobID, err)
		_ = a.store.FailCompendiumImportJob(ctx, jobID, "persist_failed")
	}
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
