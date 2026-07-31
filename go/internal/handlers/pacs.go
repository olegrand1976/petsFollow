package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/redisx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	pacsStatusKey = "pacs:status"
	pacsLogsKey   = "pacs:logs"
	pacsLogMax    = 200

	pacsStateOffline  = "offline"
	pacsStateStarting = "starting"
	pacsStateReady    = "ready"

	pacsTTLReady    = 30 * time.Second
	pacsTTLStarting = 90 * time.Second
	pacsTTLOffline  = 10 * time.Second
)

type pacsStatusPayload struct {
	State     string `json:"state"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	CheckedAt string `json:"checkedAt"`
	Error     string `json:"error,omitempty"`
}

type pacsLogEntry struct {
	At      string `json:"at"`
	Level   string `json:"level"`
	Event   string `json:"event"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type pacsMemCached struct {
	payload   pacsStatusPayload
	expiresAt time.Time
}

var (
	pacsMemMu     sync.Mutex
	pacsMemStatus *pacsMemCached
	pacsMemLogs   []pacsLogEntry
)

func (a *API) registerPacsRoutes(pr chi.Router) {
	pr.Get("/pacs/status", a.pacsStatus)
	pr.Post("/pacs/wake", a.pacsWake)
	pr.Get("/pets/{petID}/pacs/studies", a.listPetPacsStudies)
	pr.Post("/pets/{petID}/pacs/studies", a.uploadPetPacsStudy)
	pr.Get("/pacs/studies/{orthancStudyID}", a.getPacsStudy)
	pr.Get("/pacs/series/{orthancSeriesID}", a.getPacsSeries)
	pr.Get("/pacs/instances/{instanceID}/file", a.getPacsInstanceFile)
	pr.Get("/pacs/instances/{instanceID}/frames/{frame}/preview", a.getPacsInstancePreview)
}

func (a *API) registerAdminPacsRoutes(pr chi.Router) {
	pr.Get("/admin/pacs/logs", a.adminPacsLogs)
	pr.Get("/admin/pacs/metrics", a.adminPacsMetrics)
}

func (a *API) requirePacsEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.PacsEnabled {
		writeErr(w, r, http.StatusNotFound, "pacs_disabled", "pacs_disabled")
		return false
	}
	return true
}

func (a *API) orthanc() *orthancClient {
	if a.orthancClient != nil {
		return a.orthancClient
	}
	return newOrthancClient(a.cfg.PacsOrthancURL, a.cfg.PacsOrthancUser, a.cfg.PacsOrthancPassword, a.cfg.PacsOrthancUseIDToken)
}

func (a *API) SetRedis(c *redisx.Client) {
	a.redis = c
}

func (a *API) SetOrthancClientForTest(c *orthancClient) {
	a.orthancClient = c
}

// TestSetOrthanc wires a mock Orthanc base URL (integration tests).
func TestSetOrthanc(a *API, baseURL string) {
	a.orthancClient = newOrthancClient(baseURL, "", "", false)
	pacsMemMu.Lock()
	pacsMemStatus = nil
	pacsMemLogs = nil
	pacsMemMu.Unlock()
}

// TestArmFailNextPetStudyInsert forces the next CreatePetStudy in upload to fail (compensate path).
func TestArmFailNextPetStudyInsert(a *API) {
	a.failNextPetStudyInsert = true
}

// TestClearFailNextPetStudyInsert clears a pending armed fail (test cleanup).
func TestClearFailNextPetStudyInsert(a *API) {
	a.failNextPetStudyInsert = false
}


func sanitizePacsDetail(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > 120 {
		r := []rune(s)
		s = string(r[:120]) + "…"
	}
	return s
}

func looksLikeDicom(data []byte) bool {
	if len(data) < 132 {
		return false
	}
	return string(data[128:132]) == "DICM"
}

func (a *API) pacsStatus(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	if _, ok := a.requirePacsVetRead(w, r); !ok {
		return
	}
	httpx.WriteData(w, http.StatusOK, a.resolvePacsStatus(r.Context(), false))
}

func (a *API) pacsWake(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	if _, ok := a.requirePacsVetRead(w, r); !ok {
		return
	}
	a.appendPacsLog(r.Context(), "info", "wake", "wake requested", "")
	starting := pacsStatusPayload{
		State:     pacsStateStarting,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	a.cachePacsStatus(r.Context(), starting, pacsTTLStarting)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		_ = a.resolvePacsStatus(ctx, true)
	}()

	httpx.WriteData(w, http.StatusAccepted, starting)
}

func (a *API) resolvePacsStatus(ctx context.Context, force bool) pacsStatusPayload {
	if !force {
		if cached, ok := a.loadPacsStatus(ctx); ok {
			return cached
		}
	}
	client := a.orthanc()
	if client == nil {
		out := pacsStatusPayload{
			State:     pacsStateOffline,
			CheckedAt: time.Now().UTC().Format(time.RFC3339),
			Error:     "orthanc_url_missing",
		}
		a.cachePacsStatus(ctx, out, pacsTTLOffline)
		a.appendPacsLog(ctx, "error", "status", "orthanc url missing", "")
		return out
	}
	latency, err := client.pingSystem(ctx, 2*time.Second)
	now := time.Now().UTC().Format(time.RFC3339)
	if err != nil {
		out := pacsStatusPayload{
			State:     pacsStateOffline,
			LatencyMs: latency,
			CheckedAt: now,
			Error:     "unreachable",
		}
		if prev, ok := a.loadPacsStatus(ctx); ok && prev.State == pacsStateStarting {
			out.State = pacsStateStarting
			a.cachePacsStatus(ctx, out, pacsTTLStarting)
		} else {
			a.cachePacsStatus(ctx, out, pacsTTLOffline)
		}
		a.appendPacsLog(ctx, "warn", "status", "orthanc unreachable", sanitizePacsDetail(err.Error()))
		return out
	}
	out := pacsStatusPayload{
		State:     pacsStateReady,
		LatencyMs: latency,
		CheckedAt: now,
	}
	a.cachePacsStatus(ctx, out, pacsTTLReady)
	a.appendPacsLog(ctx, "info", "status", "orthanc ready", strconv.FormatInt(latency, 10)+"ms")
	return out
}

func (a *API) loadPacsStatus(ctx context.Context) (pacsStatusPayload, bool) {
	if a.redis != nil {
		raw, err := a.redis.Get(ctx, pacsStatusKey)
		if err == nil && raw != "" {
			var p pacsStatusPayload
			if json.Unmarshal([]byte(raw), &p) == nil {
				return p, true
			}
		}
		// Redis configured: miss = expired — do not fall back to sticky memory.
		return pacsStatusPayload{}, false
	}
	pacsMemMu.Lock()
	defer pacsMemMu.Unlock()
	if pacsMemStatus != nil && time.Now().Before(pacsMemStatus.expiresAt) {
		return pacsMemStatus.payload, true
	}
	pacsMemStatus = nil
	return pacsStatusPayload{}, false
}

func (a *API) cachePacsStatus(ctx context.Context, p pacsStatusPayload, ttl time.Duration) {
	raw, _ := json.Marshal(p)
	if a.redis != nil {
		_ = a.redis.Set(ctx, pacsStatusKey, string(raw), ttl)
	}
	pacsMemMu.Lock()
	pacsMemStatus = &pacsMemCached{payload: p, expiresAt: time.Now().Add(ttl)}
	pacsMemMu.Unlock()
}

func (a *API) appendPacsLog(ctx context.Context, level, event, message, detail string) {
	entry := pacsLogEntry{
		At:      time.Now().UTC().Format(time.RFC3339),
		Level:   level,
		Event:   event,
		Message: message,
		Detail:  sanitizePacsDetail(detail),
	}
	raw, _ := json.Marshal(entry)
	if a.redis != nil {
		_ = a.redis.LPush(ctx, pacsLogsKey, string(raw))
		_ = a.redis.LTrim(ctx, pacsLogsKey, 0, pacsLogMax-1)
	}
	pacsMemMu.Lock()
	pacsMemLogs = append([]pacsLogEntry{entry}, pacsMemLogs...)
	if len(pacsMemLogs) > pacsLogMax {
		pacsMemLogs = pacsMemLogs[:pacsLogMax]
	}
	pacsMemMu.Unlock()
}

func (a *API) listPacsLogs(ctx context.Context, limit int) []pacsLogEntry {
	if limit <= 0 || limit > pacsLogMax {
		limit = 100
	}
	if a.redis != nil {
		rows, err := a.redis.LRange(ctx, pacsLogsKey, 0, int64(limit-1))
		if err == nil && len(rows) > 0 {
			out := make([]pacsLogEntry, 0, len(rows))
			for _, row := range rows {
				var e pacsLogEntry
				if json.Unmarshal([]byte(row), &e) == nil {
					out = append(out, e)
				}
			}
			return out
		}
	}
	pacsMemMu.Lock()
	defer pacsMemMu.Unlock()
	if len(pacsMemLogs) == 0 {
		return []pacsLogEntry{}
	}
	if limit > len(pacsMemLogs) {
		limit = len(pacsMemLogs)
	}
	out := make([]pacsLogEntry, limit)
	copy(out, pacsMemLogs[:limit])
	return out
}

func (a *API) adminPacsLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	httpx.WriteData(w, http.StatusOK, a.listPacsLogs(r.Context(), limit))
}

func (a *API) adminPacsMetrics(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	status := a.resolvePacsStatus(r.Context(), false)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":         status,
		"orthancUrlSet":  strings.TrimSpace(a.cfg.PacsOrthancURL) != "",
		"useIdToken":     a.cfg.PacsOrthancUseIDToken,
		"redisConnected": a.redis != nil,
	})
}

func (a *API) listPetPacsStudies(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	items, err := a.store.ListPetStudies(r.Context(), petID, id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.PetStudy{}
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) uploadPetPacsStudy(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes); !ok {
		return
	}
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return
	}
	status := a.resolvePacsStatus(r.Context(), false)
	if status.State != pacsStateReady {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_not_ready", status.State)
		return
	}
	if err := r.ParseMultipartForm(64<<20 + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "multipart_required")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "file_required")
		return
	}
	defer file.Close()
	name := strings.ToLower(hdr.Filename)
	if !strings.HasSuffix(name, ".dcm") && !strings.Contains(hdr.Header.Get("Content-Type"), "dicom") {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "dicom_required")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, 64<<20))
	if err != nil || len(data) == 0 {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "empty_file")
		return
	}
	if !looksLikeDicom(data) {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "dicom_magic_required")
		return
	}
	up, err := client.uploadInstance(r.Context(), data)
	if err != nil {
		a.appendPacsLog(r.Context(), "error", "upload", "orthanc upload failed", err.Error())
		writeErr(w, r, http.StatusBadGateway, "pacs_upload_failed", "pacs_upload_failed")
		return
	}
	studyID := up.ParentStudy
	seriesID := up.ParentSeries
	studyUID := ""
	modality := ""
	desc := strings.TrimSpace(r.FormValue("description"))
	if studyID != "" {
		if meta, err := client.getStudy(r.Context(), studyID); err == nil {
			if mt, ok := meta["MainDicomTags"].(map[string]any); ok {
				if v, ok := mt["StudyInstanceUID"].(string); ok {
					studyUID = v
				}
				if v, ok := mt["StudyDescription"].(string); ok && desc == "" {
					desc = v
				}
				if v, ok := mt["ModalitiesInStudy"].(string); ok {
					modality = v
				}
			}
		}
	}
	var row store.PetStudy
	if a.failNextPetStudyInsert {
		a.failNextPetStudyInsert = false
		err = errors.New("test_forced_pet_study_insert")
	} else {
		row, err = a.store.CreatePetStudy(r.Context(), store.CreatePetStudyInput{
			PetID:            petID,
			PracticeID:       id.PracticeID,
			OrthancStudyID:   studyID,
			StudyInstanceUID: studyUID,
			OrthancSeriesID:  seriesID,
			Description:      desc,
			Modality:         modality,
			UploadedByUserID: id.UserID,
		})
	}
	if err != nil {
		a.appendPacsLog(r.Context(), "error", "upload", "db insert failed", err.Error())
		if studyID != "" {
			_ = client.deleteStudy(r.Context(), studyID)
			a.appendPacsLog(r.Context(), "warn", "upload", "db insert failed — orthanc study deleted", studyID)
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.appendPacsLog(r.Context(), "info", "upload", "dicom uploaded", studyID)
	httpx.WriteData(w, http.StatusCreated, row)
}

func (a *API) requirePacsVetRead(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return authx.Identity{}, false
	}
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) requireOrthancStudyAccess(w http.ResponseWriter, r *http.Request, orthancStudyID string) (authx.Identity, bool) {
	id, ok := a.requirePacsVetRead(w, r)
	if !ok {
		return authx.Identity{}, false
	}
	if err := validateOrthancID(orthancStudyID); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "invalid_orthanc_id")
		return authx.Identity{}, false
	}
	if _, err := a.store.FindPetStudyByOrthancStudy(r.Context(), id.PracticeID, orthancStudyID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return authx.Identity{}, false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) getPacsStudy(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	studyID := chi.URLParam(r, "orthancStudyID")
	if _, ok := a.requireOrthancStudyAccess(w, r, studyID); !ok {
		return
	}
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return
	}
	meta, err := client.getStudy(r.Context(), studyID)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	httpx.WriteData(w, http.StatusOK, meta)
}

func (a *API) getPacsSeries(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePacsVetRead(w, r)
	if !ok {
		return
	}
	seriesID := chi.URLParam(r, "orthancSeriesID")
	if err := validateOrthancID(seriesID); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "invalid_orthanc_id")
		return
	}
	if _, err := a.store.FindPetStudyByOrthancSeries(r.Context(), id.PracticeID, seriesID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return
	}
	meta, err := client.getSeries(r.Context(), seriesID)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	httpx.WriteData(w, http.StatusOK, meta)
}

func (a *API) authorizePacsInstance(w http.ResponseWriter, r *http.Request, instanceID string) (*orthancClient, bool) {
	id, ok := a.requirePacsVetRead(w, r)
	if !ok {
		return nil, false
	}
	if err := validateOrthancID(instanceID); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "invalid_orthanc_id")
		return nil, false
	}
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return nil, false
	}
	parentStudy, err := client.instanceParentStudy(r.Context(), instanceID)
	if err != nil || parentStudy == "" {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return nil, false
	}
	if _, err := a.store.FindPetStudyByOrthancStudy(r.Context(), id.PracticeID, parentStudy); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return nil, false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return nil, false
	}
	return client, true
}

func (a *API) getPacsInstanceFile(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	instanceID := chi.URLParam(r, "instanceID")
	client, ok := a.authorizePacsInstance(w, r, instanceID)
	if !ok {
		return
	}
	data, ct, err := client.getInstanceFile(r.Context(), instanceID)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (a *API) getPacsInstancePreview(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	instanceID := chi.URLParam(r, "instanceID")
	client, ok := a.authorizePacsInstance(w, r, instanceID)
	if !ok {
		return
	}
	frame, _ := strconv.Atoi(chi.URLParam(r, "frame"))
	if frame < 0 {
		frame = 0
	}
	data, ct, err := client.getInstanceFramesPreview(r.Context(), instanceID, frame)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
