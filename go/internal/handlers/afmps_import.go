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
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
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
	if int64(len(data)) > pharmacy.MaxAFMPSCSVBytes {
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

// POST /internal/afmps-import/run — monthly gate-1 only (no commit / deactivate-missing).
func (a *API) internalAfmpsImportRun(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Afmps-Import-Secret", a.cfg.AfmpsImportSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if !a.cfg.PharmacyEnabled {
		writeErr(w, r, http.StatusNotFound, "pharmacy_disabled", "not_found")
		return
	}

	objectKey := strings.TrimSpace(a.cfg.AfmpsImportObjectKey)
	if objectKey == "" {
		objectKey = pharmacy.DefaultAFMPSImportObjectKey
	}
	// Bucket médias = allUsers objectViewer : une clé prévisible fuit le pack licencié.
	if pharmacy.IsPredictableAFMPSImportObjectKey(objectKey) && !a.cfg.DevSeedEnabled {
		a.notifyAfmpsImportOps(r.Context(), "predictable_object_key",
			fmt.Sprintf("AFMPS_IMPORT_OBJECT_KEY=%q est prévisible sur le bucket public — poser une clé opaque (ex. afmps-imports/<uuid>.csv) puis redéployer.", objectKey),
			"/admin/afmps-imports", "")
		writeErr(w, r, http.StatusServiceUnavailable, "afmps_object_key_predictable", "afmps_object_key_predictable")
		return
	}

	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "media_unavailable", "media_unavailable")
		return
	}

	filename := filepath.Base(objectKey)
	if filename == "" || filename == "." {
		filename = "pack.csv"
	}

	open, err := a.store.HasOpenAFMPSImportJob(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if open {
		a.notifyAfmpsImportOps(r.Context(), "pending_job",
			"Un import AFMPS validated/reviewed/blocked est encore ouvert — finir gates 2–3 (ou supprimer un blocked) avant un nouveau run.",
			"/admin/afmps-imports", "")
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"skipped": true,
			"reason":  "pending_job",
		})
		return
	}

	rc, _, err := a.media.Open(r.Context(), objectKey)
	if err != nil {
		if media.IsNotExist(err) {
			a.notifyAfmpsImportOps(r.Context(), "object_missing",
				fmt.Sprintf("Objet CSV AFMPS introuvable : %s — déposer le pack licencié avant le prochain run.", objectKey),
				"/admin/afmps-imports", "")
			httpx.WriteData(w, http.StatusOK, map[string]any{
				"skipped":   true,
				"reason":    "object_missing",
				"objectKey": objectKey,
			})
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "object_open_failed")
		return
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, pharmacy.MaxAFMPSCSVBytes+1))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "file_read_failed")
		return
	}
	if len(data) == 0 {
		a.notifyAfmpsImportOps(r.Context(), "empty_file",
			fmt.Sprintf("CSV AFMPS vide : %s", objectKey),
			"/admin/afmps-imports", "")
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"skipped":   true,
			"reason":    "empty_file",
			"objectKey": objectKey,
		})
		return
	}
	if int64(len(data)) > pharmacy.MaxAFMPSCSVBytes {
		a.notifyAfmpsImportOps(r.Context(), "file_too_large",
			fmt.Sprintf("CSV AFMPS trop volumineux (>%d MiB) : %s", pharmacy.MaxAFMPSCSVBytes>>20, objectKey),
			"/admin/afmps-imports", "")
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"skipped":   true,
			"reason":    "file_too_large",
			"objectKey": objectKey,
		})
		return
	}

	// Hash before full parse: monthly re-runs with the same pack must not burn CPU/timeout.
	checksum := pharmacy.AFMPSChecksum(data)
	if prev, err := a.store.LatestCompletedAFMPSChecksum(r.Context()); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	} else if prev != "" && prev == checksum {
		a.notifyAfmpsImportOps(r.Context(), "unchanged_checksum",
			fmt.Sprintf("Checksum AFMPS identique au dernier commit (%s) — rien à faire.", prev[:min(12, len(prev))]),
			"/admin/afmps-imports", "")
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"skipped":        true,
			"reason":         "unchanged_checksum",
			"checksumSha256": checksum,
			"objectKey":      objectKey,
		})
		return
	}

	rows, report, parseErr := pharmacy.ParseAndValidateAFMPSCSV(bytes.NewReader(data), filename)

	job, err := a.store.CreateAFMPSImportFromParsed(r.Context(), "", filename, rows, report)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "persist_failed")
		return
	}

	route := "/admin/afmps-imports/" + job.ID
	detail := fmt.Sprintf(
		"Gate 1 AFMPS terminée (status=%s, ready=%d, insert=%d, update=%d, unchanged=%d, blocked=%v).",
		job.Status, job.ReadyCount, job.InsertCount, job.UpdateCount, job.UnchangedCount, report.Blocked,
	)
	if parseErr != nil {
		detail += " parseErr=" + parseErr.Error()
	}
	if report.BlockReason != "" {
		detail += " blockReason=" + report.BlockReason
	}
	kind := "gate1_ready"
	if report.Blocked {
		kind = "gate1_blocked"
	}
	a.notifyAfmpsImportOps(r.Context(), kind, detail, route, job.ID)

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"skipped":   false,
		"objectKey": objectKey,
		"job":       job,
		"report":    report,
	})
}

// notifyAfmpsImportOps creates a system support ticket + ops email (soft-fail).
// Deduped via ops.auth_alerts (kind afmps_<kind>, window ~25d / job id).
func (a *API) notifyAfmpsImportOps(ctx context.Context, kind, detail, route, jobID string) {
	if a.store == nil {
		return
	}
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return
	}
	alertKind := "afmps_" + kind
	fingerprint := strings.TrimSpace(jobID)
	if fingerprint == "" {
		fingerprint = time.Now().UTC().Format("2006-01")
	}
	const dedupWindow = 25 * 24 * time.Hour
	if exists, err := a.store.RecentAuthAlertExists(ctx, alertKind, fingerprint, dedupWindow); err != nil {
		log.Printf("afmps_import notify dedup: %v", err)
	} else if exists {
		return
	}

	subject := "[AFMPS] Import catalogue : " + kind
	if len([]rune(subject)) > store.MaxSupportSubjectLen {
		subject = string([]rune(subject)[:store.MaxSupportSubjectLen])
	}
	msg := strings.TrimSpace(detail)
	if msg == "" {
		msg = kind
	}
	if route == "" {
		route = "/admin/afmps-imports"
	}
	diag, _ := json.Marshal(map[string]any{
		"kind":   kind,
		"jobId":  jobID,
		"source": "afmps-import-cron",
		"at":     time.Now().UTC().Format(time.RFC3339),
	})
	ticketID := ""
	if ticket, err := a.store.CreateSupportTicket(ctx, store.CreateSupportTicketInput{
		Source:      store.SupportSourceSystem,
		Subject:     subject,
		Message:     msg,
		Diagnostics: diag,
		Route:       route,
	}); err != nil {
		log.Printf("afmps_import notify ticket: %v", err)
	} else {
		ticketID = ticket.ID
	}
	if err := a.store.InsertAuthAlert(ctx, alertKind, fingerprint, msg, ticketID); err != nil {
		log.Printf("afmps_import notify alert: %v", err)
	}

	to := strings.TrimSpace(a.cfg.OpsNotifyEmail)
	if to == "" {
		to = strings.TrimSpace(a.cfg.SupportInboxEmail)
	}
	if to == "" || a.notifier == nil {
		return
	}
	adminURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + route
	body := fmt.Sprintf(
		"<p><strong>[AFMPS]</strong> Synchro catalogue mensuelle</p>"+
			"<ul><li><strong>Kind</strong> : %s</li><li><strong>Detail</strong> : %s</li></ul>"+
			"<p><a href=\"%s\">Ouvrir dans l'admin</a></p>",
		htmlEsc(kind), htmlEsc(msg), htmlEsc(adminURL),
	)
	go func(toAddr, subj, html string) {
		_ = a.notifier.SendVetAlert(toAddr, subj, html)
	}(to, subject, body)
}
