package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/crewai"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/redisx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const improveHubTTL = 5 * time.Minute

// crewWarmupBudget bounds the wait for the CrewAI Cloud Run instance
// (minScale=0 → cold start) before submitting the task. Var: shortened in
// integration tests via TestSetCrewWarmupBudget.
var crewWarmupBudget = 90 * time.Second

// improveRunHub buffers SSE frames for browser clients.
// With Redis: frames are always PUBLISHed (cross-instance live SSE); in-proc
// buffer remains for same-instance subscribers and CancelFunc map.
type improveRunHub struct {
	mu      sync.Mutex
	runs    map[string]*improveRunChannel
	cancels map[string]context.CancelFunc
	redis   *redisx.Client
}

type improveRunChannel struct {
	mu        sync.Mutex
	events    []sseFrame
	done      bool
	wait      chan struct{}
	openedAt  time.Time
	completed time.Time
}

type sseFrame struct {
	ID   string
	Type string
	Data json.RawMessage
}

type improveHubWire struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
	Done bool            `json:"done,omitempty"`
}

func newImproveRunHub() *improveRunHub {
	return &improveRunHub{
		runs:    map[string]*improveRunChannel{},
		cancels: map[string]context.CancelFunc{},
	}
}

func (h *improveRunHub) setRedis(c *redisx.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.redis = c
}

func (h *improveRunHub) redisClient() *redisx.Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.redis
}

func (h *improveRunHub) hubChannel(runID string) string {
	return "improve-hub:" + runID
}

func (h *improveRunHub) cancelChannel(runID string) string {
	return "improve-cancel:" + runID
}

func (h *improveRunHub) sweepLocked(now time.Time) {
	for id, ch := range h.runs {
		ch.mu.Lock()
		done := ch.done
		opened := ch.openedAt
		completed := ch.completed
		ch.mu.Unlock()
		age := now.Sub(opened)
		if done && !completed.IsZero() && now.Sub(completed) > improveHubTTL {
			delete(h.runs, id)
			delete(h.cancels, id)
			continue
		}
		if age > improveHubTTL*2 {
			delete(h.runs, id)
			delete(h.cancels, id)
		}
	}
}

func (h *improveRunHub) open(runID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweepLocked(time.Now())
	h.runs[runID] = &improveRunChannel{wait: make(chan struct{}), openedAt: time.Now()}
}

func (h *improveRunHub) registerCancel(runID string, cancel context.CancelFunc) {
	h.mu.Lock()
	h.cancels[runID] = cancel
	rdb := h.redis
	h.mu.Unlock()
	if rdb == nil {
		return
	}
	// Cross-instance cancel signal → invoke local CancelFunc.
	go func() {
		ctx, stop := context.WithTimeout(context.Background(), improveHubTTL)
		defer stop()
		ch := rdb.Subscribe(ctx, h.cancelChannel(runID))
		if ch == nil {
			return
		}
		select {
		case <-ctx.Done():
		case _, ok := <-ch:
			if ok {
				h.cancel(runID)
			}
		}
	}()
}

func (h *improveRunHub) clearCancel(runID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.cancels, runID)
}

func (h *improveRunHub) cancel(runID string) {
	h.mu.Lock()
	c := h.cancels[runID]
	delete(h.cancels, runID)
	h.mu.Unlock()
	if c != nil {
		c()
	}
}

func (h *improveRunHub) publishCancelSignal(runID string) {
	rdb := h.redisClient()
	if rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rdb.Publish(ctx, h.cancelChannel(runID), "1")
}

func (h *improveRunHub) publish(runID, typ string, data any) {
	b, _ := json.Marshal(data)
	h.mu.Lock()
	ch := h.runs[runID]
	h.mu.Unlock()
	id := "0"
	if ch != nil {
		ch.mu.Lock()
		id = fmt.Sprintf("%d", len(ch.events)+1)
		ch.events = append(ch.events, sseFrame{ID: id, Type: typ, Data: b})
		wait := ch.wait
		ch.wait = make(chan struct{})
		ch.mu.Unlock()
		close(wait)
	}
	// Always fan-out to Redis when available (even if local channel missing).
	if rdb := h.redisClient(); rdb != nil {
		wire, _ := json.Marshal(improveHubWire{ID: id, Type: typ, Data: b})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = rdb.Publish(ctx, h.hubChannel(runID), string(wire))
		cancel()
	}
}

func (h *improveRunHub) complete(runID string) {
	h.mu.Lock()
	ch := h.runs[runID]
	h.mu.Unlock()
	if ch != nil {
		ch.mu.Lock()
		ch.done = true
		ch.completed = time.Now()
		wait := ch.wait
		ch.wait = make(chan struct{})
		ch.mu.Unlock()
		close(wait)
	}
	if rdb := h.redisClient(); rdb != nil {
		wire, _ := json.Marshal(improveHubWire{Type: "ping", Data: json.RawMessage(`{"done":true}`), Done: true})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = rdb.Publish(ctx, h.hubChannel(runID), string(wire))
		cancel()
	}
}

func (h *improveRunHub) get(runID string) *improveRunChannel {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweepLocked(time.Now())
	return h.runs[runID]
}

func (h *improveRunHub) hasRedis() bool {
	return h.redisClient() != nil
}

func (a *API) improveVisitReportAdvanced(w http.ResponseWriter, r *http.Request) {
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	var improveReq visitReportImproveReq
	if err := httpx.DecodeJSON(r, &improveReq); err != nil && !errors.Is(err, io.EOF) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	localeHint, ok := gemini.NormalizeVisitReportTargetLocale(improveReq.TargetLocale)
	if !ok {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_target_locale")
		return
	}
	if !a.requireAiCrEntitlement(w, r, visit.PracticeID) {
		return
	}
	if a.crewai == nil || !a.crewai.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "crewai_not_configured")
		return
	}
	if _, err := a.store.ActiveImproveRunForVisit(r.Context(), visitID); err == nil {
		writeErr(w, r, http.StatusConflict, "conflict", "run_in_progress")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	report, err := a.store.EnsureVisitReport(r.Context(), visitID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if report.Status == "final" {
		writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
		return
	}
	source := gemini.NormalizeVisitReportText(improveReq.SourceText)
	if strings.TrimSpace(source) == "" {
		source = gemini.NormalizeVisitReportText(report.BodyText)
	}
	if strings.TrimSpace(source) == "" {
		source = gemini.NormalizeVisitReportText(report.TranscriptText)
	}
	if strings.TrimSpace(source) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	country := "BE"
	if contact, cerr := a.store.GetPracticeContact(r.Context(), visit.PracticeID); cerr == nil && contact.CountryCode != "" {
		country = store.NormalizeCountryCode(contact.CountryCode)
	}
	run, err := a.store.CreateImproveRun(r.Context(), visitID, report.ID, visit.PracticeID, id.UserID)
	if err != nil {
		// unique active run race
		writeErr(w, r, http.StatusConflict, "conflict", "run_in_progress")
		return
	}
	a.improveHub.open(run.ID)

	runCtx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	a.improveHub.registerCancel(run.ID, cancel)
	go func() {
		defer cancel()
		defer a.improveHub.clearCancel(run.ID)
		a.runImproveAdvanced(runCtx, run.ID, visitID, report.ID, visit.PracticeID, id.UserID, source, localeHint, country)
	}()

	httpx.WriteData(w, http.StatusAccepted, map[string]any{
		"runId":  run.ID,
		"status": "queued",
	})
}

func (a *API) runImproveAdvanced(
	ctx context.Context,
	runID, visitID, reportID, practiceID, userID, source, localeHint, country string,
) {
	start := time.Now()
	defer a.improveHub.complete(runID)
	bg := context.Background()

	// minScale=0 — make sure the orchestrator instance is up before submitting.
	if err := a.waitCrewReady(ctx, bg, runID); err != nil {
		if a.improveRunWasCancelled(bg, runID) {
			return
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			_ = a.store.CancelImproveRun(bg, runID)
			return
		}
		latency := int(time.Since(start).Milliseconds())
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, nil, nil, "crewai_unavailable", latency)
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": "crewai_unavailable"})
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		fmt.Printf("improve-advanced run_id=%s practice_id=%s status=failed error_code=crewai_unavailable latency_ms=%d err=%v\n",
			runID, practiceID, latency, err)
		return
	}

	out, err := a.crewai.SubmitTask(ctx, crewai.SubmitRequest{
		Workflow:      crewai.WorkflowCRImprove,
		Tenant:        crewai.TenantPetsFollow,
		CorrelationID: runID,
		Async:         true,
		Payload: map[string]any{
			"sourceText":   source,
			"targetLocale": localeHint,
			"practiceId":   practiceID,
			"countryCode":  country,
		},
	})
	if err != nil || out == nil || out.TaskID == "" {
		if a.improveRunWasCancelled(bg, runID) {
			return
		}
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, nil, nil, "crewai_submit_failed", int(time.Since(start).Milliseconds()))
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": "crewai_submit_failed"})
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		fmt.Printf("improve-advanced run_id=%s practice_id=%s status=failed error_code=crewai_submit_failed latency_ms=%d\n",
			runID, practiceID, int(time.Since(start).Milliseconds()))
		return
	}
	_ = a.store.UpdateImproveRunCrewTask(bg, runID, out.TaskID)
	a.publishImproveStatus(bg, runID, "running")

	events := make(chan crewai.StreamEvent, 16)
	var streamErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamErr = a.crewai.StreamEvents(ctx, out.TaskID, events)
		close(events)
	}()

	var finalMarkdown string
	var citations any
	var metrics any
	failed := false
	errorCode := ""
	cancelled := false
	for ev := range events {
		if a.improveRunWasCancelled(bg, runID) {
			cancelled = true
			a.improveHub.cancel(runID)
			break
		}
		switch ev.Type {
		case "step":
			var step map[string]any
			_ = json.Unmarshal(ev.Data, &step)
			_ = a.store.AppendImproveRunStep(bg, runID, step)
			a.improveHub.publish(runID, "step", step)
		case "thought", "warning":
			a.improveHub.publish(runID, ev.Type, json.RawMessage(ev.Data))
		case "final":
			var payload struct {
				Report string `json:"report"`
				Result struct {
					Markdown    string `json:"markdown"`
					Citations   any    `json:"citations"`
					RagHitCount int    `json:"ragHitCount"`
				} `json:"result"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			finalMarkdown = strings.TrimSpace(payload.Report)
			if finalMarkdown == "" {
				finalMarkdown = strings.TrimSpace(payload.Result.Markdown)
			}
			citations = payload.Result.Citations
			if payload.Result.RagHitCount > 0 {
				metrics = map[string]any{"ragHitCount": payload.Result.RagHitCount}
			}
			a.improveHub.publish(runID, "final", map[string]any{"report": finalMarkdown})
		case "error":
			failed = true
			var payload struct {
				ErrorCode string `json:"errorCode"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			errorCode = payload.ErrorCode
			if errorCode == "" {
				errorCode = "crewai_failed"
			}
			a.improveHub.publish(runID, "error", map[string]any{"errorCode": errorCode})
		case "ping":
			// keepalive
		}
	}
	if cancelled {
		for range events {
			// drain until StreamEvents exits after ctx cancel
		}
	}
	<-done

	if cancelled || a.improveRunWasCancelled(bg, runID) {
		return
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(streamErr, context.Canceled) {
		_ = a.store.CancelImproveRun(bg, runID)
		return
	}
	if streamErr != nil && finalMarkdown == "" && !failed {
		failed = true
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(streamErr, context.DeadlineExceeded) {
			errorCode = "timeout"
		} else {
			errorCode = "crewai_stream_failed"
		}
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": errorCode})
	}

	latency := int(time.Since(start).Milliseconds())
	if failed || finalMarkdown == "" {
		if errorCode == "" {
			errorCode = "empty_result"
		}
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, citations, metrics, errorCode, latency)
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		fmt.Printf("improve-advanced run_id=%s practice_id=%s status=failed error_code=%s latency_ms=%d\n",
			runID, practiceID, errorCode, latency)
		return
	}

	// Cancel may have won the race after stream finished — never overwrite report.
	if a.improveRunWasCancelled(bg, runID) {
		return
	}

	improved := gemini.NormalizeVisitReportText(finalMarkdown)
	if _, err := a.store.UpdateVisitReportImprovedIfRunActive(bg, reportID, runID, improved); err != nil {
		if a.improveRunWasCancelled(bg, runID) {
			return
		}
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, citations, metrics, "persist_failed", latency)
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": "persist_failed"})
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		fmt.Printf("improve-advanced run_id=%s practice_id=%s status=failed error_code=persist_failed latency_ms=%d\n",
			runID, practiceID, latency)
		return
	}
	if err := a.store.CompleteImproveRun(bg, runID, store.ImproveRunCompleted, citations, metrics, "", latency); err != nil {
		// Already cancelled / terminal elsewhere.
		return
	}
	ragHits := 0
	if m, ok := metrics.(map[string]any); ok {
		if n, ok := m["ragHitCount"].(int); ok {
			ragHits = n
		}
	}
	fmt.Printf("improve-advanced run_id=%s practice_id=%s status=completed latency_ms=%d rag_hit_count=%d\n",
		runID, practiceID, latency, ragHits)
	a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageImproveAdvanced)
}

// publishImproveStatus persists a control step (with stable `state`/`label`) and
// fans out SSE `status` + `step` so live clients and DB/Redis late joiners stay
// aligned. Label = state code (locale-neutral) — the UI translates via `status`.
func (a *API) publishImproveStatus(bg context.Context, runID, state string) {
	step := map[string]any{
		"agent": "orchestrator",
		"label": state,
		"state": state,
		"at":    time.Now().UTC().Format(time.RFC3339),
	}
	_ = a.store.AppendImproveRunStep(bg, runID, step)
	a.improveHub.publish(runID, "status", map[string]any{"state": state})
	a.improveHub.publish(runID, "step", step)
}

// waitCrewReady wakes the CrewAI Cloud Run instance (minScale=0) and streams
// the wait as SSE: `status` events (crew_warming / crew_ready — translated by
// the UI) + persisted `step` frames so late joiners / DB fallback replay them.
func (a *API) waitCrewReady(ctx, bg context.Context, runID string) error {
	err := a.crewai.WaitReady(ctx, crewWarmupBudget, func() {
		a.publishImproveStatus(bg, runID, "crew_warming")
	})
	if err != nil {
		return err
	}
	a.publishImproveStatus(bg, runID, "crew_ready")
	return nil
}

// writeImprovePersistedStep replays a stored step and, when it carries a control
// `state`, also emits the matching `status` event (DB/Redis late-join path).
func writeImprovePersistedStep(w http.ResponseWriter, flusher http.Flusher, id string, step map[string]any) {
	if state, ok := step["state"].(string); ok {
		switch state {
		case "crew_warming", "crew_ready", "running":
			b, _ := json.Marshal(map[string]any{"state": state})
			fmt.Fprintf(w, "event: status\ndata: %s\n\n", string(b))
			flusher.Flush()
		}
	}
	b, _ := json.Marshal(step)
	fmt.Fprintf(w, "id: %s\nevent: step\ndata: %s\n\n", id, string(b))
	flusher.Flush()
}

func (a *API) improveRunWasCancelled(ctx context.Context, runID string) bool {
	cur, err := a.store.GetImproveRun(ctx, runID)
	return err == nil && cur.Status == store.ImproveRunCancelled
}

func (a *API) improveVisitReportAdvancedEvents(w http.ResponseWriter, r *http.Request) {
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	runID := chi.URLParam(r, "runID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	run, err := a.store.GetImproveRun(r.Context(), runID)
	if err != nil || run.VisitID != visitID {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	ch := a.improveHub.get(runID)
	if ch == nil {
		if a.improveHub.hasRedis() {
			a.streamImproveRunFromRedis(w, r, run)
			return
		}
		// Process restart / late join — synthesize from DB steps + terminal status.
		a.streamImproveRunFromDB(w, r, run)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, r, http.StatusInternalServerError, "internal", "sse_unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	lastID := r.Header.Get("Last-Event-ID")
	idx := 0
	if lastID != "" {
		ch.mu.Lock()
		for i, ev := range ch.events {
			if ev.ID == lastID {
				idx = i + 1
				break
			}
		}
		ch.mu.Unlock()
	}

	ctx := r.Context()
	for {
		ch.mu.Lock()
		for idx < len(ch.events) {
			ev := ch.events[idx]
			idx++
			ch.mu.Unlock()
			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", ev.ID, ev.Type, string(ev.Data))
			flusher.Flush()
			ch.mu.Lock()
		}
		done := ch.done
		wait := ch.wait
		ch.mu.Unlock()
		if done {
			fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
			flusher.Flush()
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-wait:
		case <-time.After(15 * time.Second):
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}

func (a *API) streamImproveRunFromRedis(w http.ResponseWriter, r *http.Request, run store.ImproveRun) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, r, http.StatusInternalServerError, "internal", "sse_unsupported")
		return
	}
	rdb := a.improveHub.redisClient()
	if rdb == nil {
		a.streamImproveRunFromDB(w, r, run)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Replay steps already persisted, then live Redis frames until done.
	var steps []map[string]any
	_ = json.Unmarshal(run.Steps, &steps)
	for i, step := range steps {
		writeImprovePersistedStep(w, flusher, fmt.Sprintf("db-%d", i+1), step)
	}
	switch run.Status {
	case store.ImproveRunCompleted, store.ImproveRunFailed, store.ImproveRunCancelled:
		code := "ok"
		if run.Status == store.ImproveRunFailed {
			code = run.ErrorCode
			if code == "" {
				code = "failed"
			}
		} else if run.Status == store.ImproveRunCancelled {
			code = "cancelled"
		}
		if run.Status == store.ImproveRunCompleted {
			fmt.Fprintf(w, "event: final\ndata: {\"report\":\"\"}\n\n")
		} else {
			fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":%q}\n\n", code)
		}
		fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
		flusher.Flush()
		return
	}

	ctx := r.Context()
	sub := rdb.Subscribe(ctx, a.improveHub.hubChannel(run.ID))
	if sub == nil {
		a.streamImproveRunFromDB(w, r, run)
		return
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub:
			if !ok {
				a.pollImproveRunTerminal(w, flusher, run.ID)
				return
			}
			var wire improveHubWire
			if err := json.Unmarshal([]byte(msg), &wire); err != nil {
				continue
			}
			if wire.Done || wire.Type == "ping" {
				data := wire.Data
				if len(data) == 0 {
					data = json.RawMessage(`{"done":true}`)
				}
				fmt.Fprintf(w, "event: ping\ndata: %s\n\n", string(data))
				flusher.Flush()
				if wire.Done || strings.Contains(string(data), `"done":true`) {
					return
				}
				continue
			}
			id := wire.ID
			if id == "" {
				id = "r"
			}
			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", id, wire.Type, string(wire.Data))
			flusher.Flush()
			if wire.Type == "final" || wire.Type == "error" {
				fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
				flusher.Flush()
				return
			}
		case <-ticker.C:
			cur, err := a.store.GetImproveRun(ctx, run.ID)
			if err != nil {
				continue
			}
			if cur.Status == store.ImproveRunCompleted || cur.Status == store.ImproveRunFailed || cur.Status == store.ImproveRunCancelled {
				a.emitTerminalFromRun(w, flusher, cur)
				return
			}
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}

func (a *API) pollImproveRunTerminal(w http.ResponseWriter, flusher http.Flusher, runID string) {
	cur, err := a.store.GetImproveRun(context.Background(), runID)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":\"run_missing\"}\n\n")
		fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
		flusher.Flush()
		return
	}
	a.emitTerminalFromRun(w, flusher, cur)
}

func (a *API) emitTerminalFromRun(w http.ResponseWriter, flusher http.Flusher, run store.ImproveRun) {
	switch run.Status {
	case store.ImproveRunCompleted:
		fmt.Fprintf(w, "event: final\ndata: {\"report\":\"\"}\n\n")
	case store.ImproveRunCancelled:
		fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":\"cancelled\"}\n\n")
	default:
		code := run.ErrorCode
		if code == "" {
			code = "failed"
		}
		fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":%q}\n\n", code)
	}
	fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
	flusher.Flush()
}

func (a *API) streamImproveRunFromDB(w http.ResponseWriter, r *http.Request, run store.ImproveRun) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, r, http.StatusInternalServerError, "internal", "sse_unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	sent := 0
	deadline := time.Now().Add(4 * time.Minute)
	for {
		var steps []map[string]any
		_ = json.Unmarshal(run.Steps, &steps)
		for i := sent; i < len(steps); i++ {
			writeImprovePersistedStep(w, flusher, fmt.Sprintf("%d", i+1), steps[i])
		}
		sent = len(steps)

		switch run.Status {
		case store.ImproveRunCompleted:
			reportText := ""
			if rep, err := a.store.GetVisitReportByID(ctx, run.ReportID); err == nil {
				reportText = strings.TrimSpace(rep.ImprovedText)
			}
			b, _ := json.Marshal(map[string]any{"report": reportText, "status": "completed"})
			fmt.Fprintf(w, "event: final\ndata: %s\n\n", string(b))
			fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
			flusher.Flush()
			return
		case store.ImproveRunFailed, store.ImproveRunCancelled:
			code := run.ErrorCode
			if run.Status == store.ImproveRunCancelled && code == "" {
				code = "cancelled"
			}
			b, _ := json.Marshal(map[string]any{"errorCode": code})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(b))
			fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
			flusher.Flush()
			return
		}

		if time.Now().After(deadline) {
			fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":\"timeout\"}\n\n")
			fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
			flusher.Flush()
			return
		}
		fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
		flusher.Flush()

		select {
		case <-ctx.Done():
			return
		case <-time.After(750 * time.Millisecond):
		}
		next, err := a.store.GetImproveRun(ctx, run.ID)
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: {\"errorCode\":\"run_missing\"}\n\n")
			fmt.Fprintf(w, "event: ping\ndata: {\"done\":true}\n\n")
			flusher.Flush()
			return
		}
		run = next
	}
}

func (a *API) cancelVisitReportAdvanced(w http.ResponseWriter, r *http.Request) {
	if !a.requireAiCrAdvancedEnabled(w, r) {
		return
	}
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	runID := chi.URLParam(r, "runID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	run, err := a.store.GetImproveRun(r.Context(), runID)
	if err != nil || run.VisitID != visitID {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if err := a.store.CancelImproveRun(r.Context(), runID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "conflict", "run_not_cancellable")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Stop the worker before/while it may still be streaming — prevents late persist.
	a.improveHub.cancel(runID)
	a.improveHub.publishCancelSignal(runID)
	if crewID := strings.TrimSpace(run.CrewTaskID); crewID != "" && a.crewai != nil && a.crewai.Configured() {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		if err := a.crewai.CancelTask(ctx, crewID); err != nil {
			fmt.Printf("improve-advanced cancel crew task %s: %v\n", crewID, err)
		}
		cancel()
	}
	a.improveHub.publish(runID, "error", map[string]any{"errorCode": "cancelled"})
	a.improveHub.complete(runID)
	httpx.WriteData(w, http.StatusOK, map[string]any{"runId": runID, "status": "cancelled"})
}

// TestSetAiCrAdvancedEnabled toggles the advanced CR flag (integration tests).
func (a *API) TestSetAiCrAdvancedEnabled(v bool) {
	a.cfg.AiCrAdvancedEnabled = v
}

// TestSetCrewAI injects a CrewAI client (integration tests).
func (a *API) TestSetCrewAI(c *crewai.Client) {
	a.crewai = c
}

// TestSetCrewWarmupBudget shortens the warm-up wait (integration tests).
func (a *API) TestSetCrewWarmupBudget(d time.Duration) {
	crewWarmupBudget = d
}
