package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerAFMPSImportRoutes(r chi.Router) {
	r.Post("/admin/afmps-imports", a.adminCreateAFMPSImport)
	r.Get("/admin/afmps-imports", a.adminListAFMPSImports)
	r.Get("/admin/afmps-imports/{id}", a.adminGetAFMPSImport)
	r.Delete("/admin/afmps-imports/{id}", a.adminDeleteAFMPSImport)
	r.Post("/admin/afmps-imports/{id}/mark-reviewed", a.adminMarkAFMPSReviewed)
	r.Patch("/admin/afmps-imports/{id}/rows/{rowId}", a.adminPatchAFMPSRow)
	r.Post("/admin/afmps-imports/{id}/commit", a.adminCommitAFMPSImport)
}

func (a *API) requirePharmacyAdmin(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
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

func (a *API) adminCreateAFMPSImport(w http.ResponseWriter, r *http.Request) {
	admin, ok := a.requirePharmacyAdmin(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(pharmacy.MaxAFMPSCSVBytes + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, pharmacy.MaxAFMPSCSVBytes+1))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_read_failed")
		return
	}
	if len(data) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "empty_file")
		return
	}
	if len(data) > pharmacy.MaxAFMPSCSVBytes {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_too_large")
		return
	}

	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		filename = "afmps.csv"
	}
	rows, report, parseErr := pharmacy.ParseAndValidateAFMPSCSV(bytes.NewReader(data), filename)
	job, err := a.store.CreateAFMPSImportFromParsed(r.Context(), admin.UserID, filename, rows, report)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "persist_failed")
		return
	}
	status := http.StatusCreated
	if report.Blocked || parseErr != nil {
		status = http.StatusUnprocessableEntity
	}
	httpx.WriteData(w, status, map[string]any{
		"job":    job,
		"report": report,
	})
}

func (a *API) adminListAFMPSImports(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	items, err := a.store.ListAFMPSImportJobs(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) adminGetAFMPSImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	collision := r.URL.Query().Get("collision")
	if limit <= 0 {
		limit = 200
	}
	detail, err := a.store.GetAFMPSImportDetail(r.Context(), chi.URLParam(r, "id"), limit, offset, collision)
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

func (a *API) adminDeleteAFMPSImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	err := a.store.DeleteAFMPSImportJob(r.Context(), chi.URLParam(r, "id"))
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
	httpx.WriteData(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *API) adminMarkAFMPSReviewed(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	job, err := a.store.MarkAFMPSImportReviewed(r.Context(), chi.URLParam(r, "id"))
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
	httpx.WriteData(w, http.StatusOK, job)
}

func (a *API) adminPatchAFMPSRow(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	var body struct {
		Exclude *bool  `json:"exclude"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	row, err := a.store.PatchAFMPSImportRow(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "rowId"), body.Exclude, body.Status)
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

func (a *API) adminCommitAFMPSImport(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requirePharmacyAdmin(w, r); !ok {
		return
	}
	var body struct {
		Confirm           string `json:"confirm"`
		DeactivateMissing bool   `json:"deactivateMissing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if !pharmacy.ConfirmAFMPSPhrase(body.Confirm) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "confirm_required")
		return
	}
	result, err := a.store.CommitAFMPSImport(r.Context(), chi.URLParam(r, "id"), body.DeactivateMissing)
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
	detail, _ := a.store.GetAFMPSImportDetail(r.Context(), chi.URLParam(r, "id"), 50, 0, "all")
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"result": result,
		"job":    detail.Job,
	})
}
