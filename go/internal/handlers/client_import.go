package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/spreadsheet"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerClientImportRoutes(r chi.Router) {
	r.Post("/admin/client-imports", a.adminCreateClientImport)
	r.Get("/admin/client-imports", a.adminListClientImports)
	r.Get("/admin/client-imports/{id}", a.adminGetClientImport)
	r.Post("/admin/client-imports/{id}/suggest-mapping", a.adminSuggestClientImportMapping)
	r.Put("/admin/client-imports/{id}/mapping", a.adminPutClientImportMapping)
	r.Patch("/admin/client-imports/{id}/rows/{rowId}", a.adminPatchClientImportRow)
	r.Post("/admin/client-imports/{id}/commit", a.adminCommitClientImport)
	r.Get("/admin/client-imports/{id}/credentials", a.adminDownloadClientImportCredentials)
}

func (a *API) adminCreateClientImport(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(spreadsheet.MaxFileBytes + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_multipart")
		return
	}
	vetUserID := strings.TrimSpace(r.FormValue("vetUserId"))
	if vetUserID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "vet_user_id_required")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, spreadsheet.MaxFileBytes+1))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_read_failed")
		return
	}
	if len(data) > spreadsheet.MaxFileBytes {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_too_large")
		return
	}

	filename := filepath.Base(header.Filename)
	contentType := header.Header.Get("Content-Type")
	format := spreadsheet.DetectFormat(filename, contentType)
	if format == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "unsupported_format")
		return
	}
	parsed, err := spreadsheet.Parse(data, format)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	detail, err := a.store.CreateClientImportJob(r.Context(), store.CreateClientImportInput{
		VetUserID:        vetUserID,
		CreatedByAdminID: admin.UserID,
		Filename:         filename,
		ContentType:      contentType,
		SourceFormat:     parsed.Format,
		Headers:          parsed.Headers,
		SampleRows:       parsed.SampleRows,
		Rows:             parsed.Rows,
	})
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "vet_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, detail)
}

func (a *API) adminListClientImports(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	items, err := a.store.ListClientImportJobs(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) adminGetClientImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	detail, err := a.store.GetClientImportJobDetail(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "import_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) adminSuggestClientImportMapping(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	job, err := a.store.GetClientImportJob(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "import_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
		return
	}

	var sample []map[string]string
	_ = json.Unmarshal(job.SampleRows, &sample)
	sug, err := a.gemini.SuggestColumnMapping(r.Context(), job.Headers, sample)
	if err != nil {
		if strings.Contains(err.Error(), "gemini_not_configured") {
			writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "gemini_failed", "gemini_failed")
		return
	}

	mapping := store.ColumnMapping{
		Email:    sug.Email,
		FullName: sug.FullName,
		Locale:   sug.Locale,
	}
	if err := a.store.SaveClientImportMappingSuggestion(r.Context(), id, mapping, sug.Raw); err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "import_status_conflict")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	detail, err := a.store.GetClientImportJobDetail(r.Context(), id)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"job":        detail.Job,
		"suggestion": sug,
		"rows":       detail.Rows,
	})
}

type putMappingReq struct {
	Email    *string `json:"email"`
	FullName *string `json:"fullName"`
	Locale   *string `json:"locale"`
}

func (a *API) adminPutClientImportMapping(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	var req putMappingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Email == nil || strings.TrimSpace(*req.Email) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "email_column_required")
		return
	}
	if req.FullName == nil || strings.TrimSpace(*req.FullName) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "full_name_column_required")
		return
	}
	mapping := store.ColumnMapping{
		Email:    req.Email,
		FullName: req.FullName,
		Locale:   req.Locale,
	}
	detail, err := a.store.ApplyClientImportMapping(r.Context(), id, mapping)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "import_not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "import_status_conflict")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

type patchImportRowReq struct {
	Excluded *bool   `json:"excluded"`
	Email    *string `json:"email"`
	FullName *string `json:"fullName"`
	Locale   *string `json:"locale"`
}

func (a *API) adminPatchClientImportRow(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	jobID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")
	var req patchImportRowReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	row, err := a.store.PatchClientImportRow(r.Context(), jobID, rowID, store.PatchClientImportRowInput{
		Excluded: req.Excluded,
		Email:    req.Email,
		FullName: req.FullName,
		Locale:   req.Locale,
	})
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "row_not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "import_status_conflict")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, row)
}

func (a *API) adminCommitClientImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	result, err := a.store.CommitClientImport(r.Context(), id, a.cfg.JWTSigningKey)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "import_not_found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeErr(w, r, http.StatusConflict, "conflict", "import_status_conflict")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
}

func (a *API) adminDownloadClientImportCredentials(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	csvBytes, err := a.store.DownloadClientImportCredentials(r.Context(), id, token, a.cfg.JWTSigningKey)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "credentials_unavailable")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="client-import-credentials.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}

// Ensure gemini package type referenced for interface satisfaction in tests.
var _ gemini.Mapper = (*gemini.Client)(nil)
