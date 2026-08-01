package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	pr.Get("/pets/{petID}/pacs/studies/{studyID}/comments", a.listPetPacsStudyComments)
	pr.Post("/pets/{petID}/pacs/studies/{studyID}/comments", a.createPetPacsStudyComment)
	pr.Get("/pacs/studies/{orthancStudyID}", a.getPacsStudy)
	pr.Get("/pacs/series/{orthancSeriesID}", a.getPacsSeries)
	pr.Get("/pacs/instances/{instanceID}/file", a.getPacsInstanceFile)
	pr.Get("/pacs/instances/{instanceID}/metadata", a.getPacsInstanceMetadata)
	pr.Get("/pacs/instances/{instanceID}/frames/{frame}/preview", a.getPacsInstancePreview)
}

func (a *API) registerAdminPacsRoutes(pr chi.Router) {
	pr.Get("/admin/pacs/logs", a.adminPacsLogs)
	pr.Get("/admin/pacs/metrics", a.adminPacsMetrics)
	pr.Post("/admin/pacs/wake", a.adminPacsWake)
	pr.Post("/admin/pacs/prune-orphans", a.adminPacsPruneOrphans)
	pr.Get("/admin/pacs/playground-pets", a.adminPacsPlaygroundPets)
	pr.Get("/admin/pacs/playground-clients", a.adminPacsPlaygroundClients)
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

// TestClearFailNextPetStudyInsert disarms the fail-next insert hook (test cleanup).
func TestClearFailNextPetStudyInsert(a *API) {
	a.failNextPetStudyInsert = false
}

// TestPurgeOrthancStudies exposes retention Orthanc delete for integration tests.
func TestPurgeOrthancStudies(a *API, studyIDs []string) {
	a.purgeOrthancStudies(context.Background(), studyIDs)
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
	a.beginPacsWake(w, r)
}

func (a *API) adminPacsWake(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	a.beginPacsWake(w, r)
}

func (a *API) beginPacsWake(w http.ResponseWriter, r *http.Request) {
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

// adminPacsPruneOrphans deletes imaging.pet_studies rows whose Orthanc study is gone
// (e.g. Orthanc DB reset while Postgres links remain). Does not touch Orthanc itself.
// Query dryRun=1 lists candidates without deleting.
func (a *API) adminPacsPruneOrphans(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return
	}
	st := a.resolvePacsStatus(r.Context(), false)
	if st.State != "ready" {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_not_ready")
		return
	}
	dryRun := r.URL.Query().Get("dryRun") == "1" || strings.EqualFold(r.URL.Query().Get("dryRun"), "true")

	const listCap = 10000
	aliveIDs, err := client.listStudyIDs(r.Context())
	if err != nil {
		a.appendPacsLog(r.Context(), "error", "prune_orphan", "list Orthanc studies failed", err.Error())
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	alive := make(map[string]struct{}, len(aliveIDs))
	for _, id := range aliveIDs {
		alive[id] = struct{}{}
	}

	rows, err := a.store.ListAllPetStudies(r.Context(), listCap)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	truncated := len(rows) >= listCap

	type prunedRow struct {
		ID             string `json:"id"`
		OrthancStudyID string `json:"orthancStudyId"`
		Description    string `json:"description"`
	}
	pruned := make([]prunedRow, 0)
	kept := 0
	skipped := 0
	for _, row := range rows {
		if _, ok := alive[row.OrthancStudyID]; ok {
			kept++
			continue
		}
		candidate := prunedRow{
			ID:             row.ID,
			OrthancStudyID: row.OrthancStudyID,
			Description:    row.Description,
		}
		if dryRun {
			pruned = append(pruned, candidate)
			continue
		}
		if derr := a.store.DeletePetStudyByID(r.Context(), row.ID); derr != nil {
			a.appendPacsLog(r.Context(), "error", "prune_orphan", "delete failed", derr.Error())
			skipped++
			continue
		}
		pruned = append(pruned, candidate)
	}
	a.appendPacsLog(r.Context(), "info", "prune_orphan",
		fmt.Sprintf("dryRun=%v pruned=%d kept=%d skipped=%d truncated=%v orthanc=%d",
			dryRun, len(pruned), kept, skipped, truncated, len(aliveIDs)), "")
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"pruned":    pruned,
		"kept":      kept,
		"skipped":   skipped,
		"dryRun":    dryRun,
		"truncated": truncated,
	})
}

// adminPacsPlaygroundClients lists seed clients (*@petsfollow.test) that own pets — step 1 of the PACS picker.
func (a *API) adminPacsPlaygroundClients(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	items, err := a.store.ListPlaygroundPacsClients(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.PlaygroundPacsClient{}
	}
	preferred := ""
	const demoEmail = "client.demo@petsfollow.test"
	for _, c := range items {
		if c.Email == demoEmail {
			preferred = demoEmail
			break
		}
	}
	if preferred == "" && len(items) > 0 {
		preferred = items[0].Email
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"preferredEmail": preferred,
		"clients":        items,
	})
}

// adminPacsPlaygroundPets lists pets for the admin PACS playground (default: client.demo seed).
// ownerEmail is optional and restricted to *.petsfollow.test to avoid arbitrary PHI lookup.
func (a *API) adminPacsPlaygroundPets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.requirePacsEnabled(w, r) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("ownerEmail")))
	if email == "" {
		email = "client.demo@petsfollow.test"
	}
	if !strings.HasSuffix(email, "@petsfollow.test") {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "playground_email_restricted")
		return
	}
	u, err := a.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "owner_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pets, err := a.store.ListPetsByOwner(r.Context(), u.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	type item struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Species    string `json:"species"`
		PracticeID string `json:"practiceId"`
	}
	out := make([]item, 0, len(pets))
	for _, p := range pets {
		out = append(out, item{ID: p.ID, Name: p.Name, Species: p.Species, PracticeID: p.PracticeID})
	}
	preferred := ""
	if len(out) > 0 {
		preferred = out[0].ID
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"ownerEmail":     email,
		"preferredPetId": preferred,
		"pets":           out,
	})
}

func (a *API) listPetPacsStudies(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePacsClinicalAccess(w, r, "pets.read")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermRead)
	if !ok {
		return
	}
	practiceID := a.resolvePacsPracticeID(id, pet)
	items, err := a.store.ListPetStudies(r.Context(), petID, practiceID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.PetStudy{}
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) requirePetPacsStudyRow(w http.ResponseWriter, r *http.Request, petID, studyRowID string, id authx.Identity, pet store.Pet) (store.PetStudy, bool) {
	if !isUUID(studyRowID) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return store.PetStudy{}, false
	}
	practiceID := a.resolvePacsPracticeID(id, pet)
	st, err := a.store.GetPetStudyForPet(r.Context(), studyRowID, petID, practiceID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return store.PetStudy{}, false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return store.PetStudy{}, false
	}
	return st, true
}

func (a *API) listPetPacsStudyComments(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePacsClinicalAccess(w, r, "pets.read")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	studyID := chi.URLParam(r, "studyID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermRead)
	if !ok {
		return
	}
	st, ok := a.requirePetPacsStudyRow(w, r, petID, studyID, id, pet)
	if !ok {
		return
	}
	items, err := a.store.ListPetStudyComments(r.Context(), st.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if items == nil {
		items = []store.PetStudyComment{}
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) createPetPacsStudyComment(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePacsClinicalAccess(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	studyID := chi.URLParam(r, "studyID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
	if !ok {
		return
	}
	st, ok := a.requirePetPacsStudyRow(w, r, petID, studyID, id, pet)
	if !ok {
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "invalid_json")
		return
	}
	comment, err := a.store.InsertPetStudyComment(r.Context(), st.ID, id.UserID, body.Body)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "body_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, comment)
}

func (a *API) uploadPetPacsStudy(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	id, ok := a.requirePacsClinicalAccess(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
	if !ok {
		return
	}
	practiceID := a.resolvePacsPracticeID(id, pet)
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
	studyID := strings.TrimSpace(up.ParentStudy)
	seriesID := strings.TrimSpace(up.ParentSeries)
	instanceID := strings.TrimSpace(up.ID)
	// Orthanc 1.12 may omit ParentStudy on upload response — resolve via ParentSeries.
	if (studyID == "" || validateOrthancID(studyID) != nil) && seriesID != "" && validateOrthancID(seriesID) == nil {
		if resolved, resErr := client.seriesParentStudy(r.Context(), seriesID); resErr == nil {
			studyID = strings.TrimSpace(resolved)
		}
	}
	if studyID == "" || validateOrthancID(studyID) != nil {
		a.appendPacsLog(r.Context(), "error", "upload", "orthanc parent study missing", instanceID)
		if instanceID != "" && validateOrthancID(instanceID) == nil {
			_ = client.deleteInstance(r.Context(), instanceID)
		}
		writeErr(w, r, http.StatusBadGateway, "pacs_upload_failed", "pacs_upload_failed")
		return
	}
	if seriesID != "" && validateOrthancID(seriesID) != nil {
		seriesID = ""
	}
	if instanceID != "" && validateOrthancID(instanceID) != nil {
		instanceID = ""
	}
	studyAlreadyLinked := false
	if _, linkErr := a.store.FindPetStudyByOrthancStudy(r.Context(), practiceID, studyID); linkErr == nil {
		studyAlreadyLinked = true
	} else if !errors.Is(linkErr, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	studyUID := ""
	modality := ""
	desc := strings.TrimSpace(r.FormValue("description"))
	if len([]rune(desc)) > 500 {
		desc = string([]rune(desc)[:500])
	}
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
	var row store.PetStudy
	if a.failNextPetStudyInsert {
		a.failNextPetStudyInsert = false
		err = errors.New("test_forced_pet_study_insert")
	} else {
		row, err = a.store.CreatePetStudy(r.Context(), store.CreatePetStudyInput{
			PetID:            petID,
			PracticeID:       practiceID,
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
		// Prefer instance delete: never wipe a ParentStudy that already holds other PHI.
		if instanceID != "" {
			_ = client.deleteInstance(r.Context(), instanceID)
			a.appendPacsLog(r.Context(), "warn", "upload", "db insert failed — orthanc instance deleted", instanceID)
		} else if studyID != "" && !studyAlreadyLinked {
			_ = client.deleteStudy(r.Context(), studyID)
			a.appendPacsLog(r.Context(), "warn", "upload", "db insert failed — orthanc study deleted", studyID)
		}
		if errors.Is(err, store.ErrPacsStudyOtherPractice) {
			writeErr(w, r, http.StatusConflict, "conflict", "pacs_study_other_practice")
			return
		}
		if errors.Is(err, store.ErrPacsStudyOtherPet) || errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "pacs_study_other_pet")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.appendPacsLog(r.Context(), "info", "upload", "dicom uploaded", studyID)
	httpx.WriteData(w, http.StatusCreated, row)
}

// requirePacsClinicalAccess allows practice staff with the given capability, or global admin
// (ops playground on /admin/pacs — no practice switch required).
func (a *API) requirePacsClinicalAccess(w http.ResponseWriter, r *http.Request, capability string) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return authx.Identity{}, false
	}
	if id.Role == kernel.RoleAdmin {
		return id, true
	}
	return a.requirePracticePerm(w, r, capability)
}

func (a *API) resolvePacsPracticeID(id authx.Identity, pet store.Pet) string {
	if strings.TrimSpace(id.PracticeID) != "" {
		return id.PracticeID
	}
	return pet.PracticeID
}

func (a *API) findPetStudyForPacs(ctx context.Context, id authx.Identity, orthancStudyID string) (store.PetStudy, error) {
	if id.Role == kernel.RoleAdmin {
		return a.store.FindPetStudyByOrthancStudyID(ctx, orthancStudyID)
	}
	return a.store.FindPetStudyByOrthancStudy(ctx, id.PracticeID, orthancStudyID)
}

func (a *API) requirePacsVetRead(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, ok := a.requirePacsClinicalAccess(w, r, "pets.read")
	if !ok {
		return authx.Identity{}, false
	}
	if id.Role == kernel.RoleAdmin {
		return id, true
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
	if _, err := a.findPetStudyForPacs(r.Context(), id, orthancStudyID); err != nil {
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
		var se *orthancStatusError
		if errors.As(err, &se) && se.Status == http.StatusNotFound {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
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
	client := a.orthanc()
	if client == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "pacs_unavailable", "orthanc_url_missing")
		return
	}
	parentStudy, err := client.seriesParentStudy(r.Context(), seriesID)
	if err != nil || parentStudy == "" {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if _, err := a.findPetStudyForPacs(r.Context(), id, parentStudy); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
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
	if _, err := a.findPetStudyForPacs(r.Context(), id, parentStudy); err != nil {
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
		a.appendPacsLog(r.Context(), "error", "instance_file", "orthanc file failed", err.Error())
		if orthancBlobUnavailable(err) {
			writeErr(w, r, http.StatusBadGateway, "pacs_instance_unavailable", "pacs_instance_unavailable")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (a *API) getPacsInstanceMetadata(w http.ResponseWriter, r *http.Request) {
	if !a.requirePacsEnabled(w, r) {
		return
	}
	instanceID := chi.URLParam(r, "instanceID")
	client, ok := a.authorizePacsInstance(w, r, instanceID)
	if !ok {
		return
	}
	tags, err := client.getInstanceSimplifiedTags(r.Context(), instanceID)
	if err != nil {
		var se *orthancStatusError
		if errors.As(err, &se) && (se.Status == http.StatusNotFound || se.Status == http.StatusBadRequest) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	httpx.WriteData(w, http.StatusOK, buildPacsInstanceMetadata(instanceID, tags))
}

type pacsInstanceMetadata struct {
	InstanceID       string         `json:"instanceId"`
	PixelSpacingMm   []float64      `json:"pixelSpacingMm,omitempty"`
	WindowCenter     *float64       `json:"windowCenter,omitempty"`
	WindowWidth      *float64       `json:"windowWidth,omitempty"`
	NumberOfFrames   int            `json:"numberOfFrames,omitempty"`
	Modality         string         `json:"modality,omitempty"`
	Rows             int            `json:"rows,omitempty"`
	Columns          int            `json:"columns,omitempty"`
	Tags             map[string]any `json:"tags,omitempty"`
}

func buildPacsInstanceMetadata(instanceID string, tags map[string]any) pacsInstanceMetadata {
	out := pacsInstanceMetadata{
		InstanceID: instanceID,
		Tags:       tags,
		Modality:   tagString(tags, "Modality"),
	}
	out.PixelSpacingMm = parseSpacingTag(tags, "PixelSpacing")
	if len(out.PixelSpacingMm) == 0 {
		out.PixelSpacingMm = parseSpacingTag(tags, "ImagerPixelSpacing")
	}
	if len(out.PixelSpacingMm) == 0 {
		out.PixelSpacingMm = parseSpacingTag(tags, "NominalScannedPixelSpacing")
	}
	if v, ok := tagFloat(tags, "WindowCenter"); ok {
		out.WindowCenter = &v
	}
	if v, ok := tagFloat(tags, "WindowWidth"); ok {
		out.WindowWidth = &v
	}
	if n, ok := tagInt(tags, "NumberOfFrames"); ok && n > 0 {
		out.NumberOfFrames = n
	}
	if n, ok := tagInt(tags, "Rows"); ok {
		out.Rows = n
	}
	if n, ok := tagInt(tags, "Columns"); ok {
		out.Columns = n
	}
	return out
}

func tagString(tags map[string]any, key string) string {
	v, _ := tags[key].(string)
	return strings.TrimSpace(v)
}

func tagFloat(tags map[string]any, key string) (float64, bool) {
	switch v := tags[key].(type) {
	case float64:
		return v, true
	case string:
		parts := strings.FieldsFunc(v, func(r rune) bool { return r == '\\' || r == ',' })
		if len(parts) == 0 {
			return 0, false
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func tagInt(tags map[string]any, key string) (int, bool) {
	switch v := tags[key].(type) {
	case float64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil
	default:
		return 0, false
	}
}

func parseSpacingTag(tags map[string]any, key string) []float64 {
	raw, ok := tags[key]
	if !ok || raw == nil {
		return nil
	}
	var s string
	switch v := raw.(type) {
	case string:
		s = v
	case float64:
		return []float64{v, v}
	default:
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '\\' || r == ',' })
	var out []float64
	for _, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil || f <= 0 {
			return nil
		}
		out = append(out, f)
	}
	if len(out) == 1 {
		out = append(out, out[0])
	}
	if len(out) < 2 {
		return nil
	}
	return out[:2]
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
	// Clamp OOR via DICOM tags before Orthanc — staging Orthanc often returns 500/502
	// for a bad frame index (GCS/plugin path), which would otherwise look like blob miss.
	if frame > 0 {
		if tags, tagErr := client.getInstanceSimplifiedTags(r.Context(), instanceID); tagErr == nil {
			nFrames := 1
			if n, ok := tagInt(tags, "NumberOfFrames"); ok && n > 0 {
				nFrames = n
			}
			if frame >= nFrames {
				writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
				return
			}
		}
	}
	data, ct, err := client.getInstanceFramesPreview(r.Context(), instanceID, frame)
	if err != nil {
		a.appendPacsLog(r.Context(), "error", "instance_preview", "orthanc preview failed", err.Error())
		// Frame hors plage (400/404, frame>0) → 404 pour clamp UI.
		if orthancPreviewOutOfRange(err, frame) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		// Index Orthanc OK mais bytes absents (souvent GCS) → message dédié.
		if orthancBlobUnavailable(err) {
			writeErr(w, r, http.StatusBadGateway, "pacs_instance_unavailable", "pacs_instance_unavailable")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "pacs_error", "pacs_error")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
