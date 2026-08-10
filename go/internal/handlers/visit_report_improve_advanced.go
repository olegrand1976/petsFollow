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
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const improveHubTTL = 5 * time.Minute

// improveRunHub buffers SSE frames for browser clients (single-instance OK for V1 —
// after process restart clients fall back to DB replay via streamImproveRunFromDB).
type improveRunHub struct {
	mu      sync.Mutex
	runs    map[string]*improveRunChannel
	cancels map[string]context.CancelFunc
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

func newImproveRunHub() *improveRunHub {
	return &improveRunHub{
		runs:    map[string]*improveRunChannel{},
		cancels: map[string]context.CancelFunc{},
	}
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
	defer h.mu.Unlock()
	h.cancels[runID] = cancel
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

func (h *improveRunHub) publish(runID, typ string, data any) {
	b, _ := json.Marshal(data)
	h.mu.Lock()
	ch := h.runs[runID]
	h.mu.Unlock()
	if ch == nil {
		return
	}
	ch.mu.Lock()
	id := fmt.Sprintf("%d", len(ch.events)+1)
	ch.events = append(ch.events, sseFrame{ID: id, Type: typ, Data: b})
	wait := ch.wait
	ch.wait = make(chan struct{})
	ch.mu.Unlock()
	close(wait)
}

func (h *improveRunHub) complete(runID string) {
	h.mu.Lock()
	ch := h.runs[runID]
	h.mu.Unlock()
	if ch == nil {
		return
	}
	ch.mu.Lock()
	ch.done = true
	ch.completed = time.Now()
	wait := ch.wait
	ch.wait = make(chan struct{})
	ch.mu.Unlock()
	close(wait)
}

func (h *improveRunHub) get(runID string) *improveRunChannel {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweepLocked(time.Now())
	return h.runs[runID]
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
	a.improveHub.publish(run.ID, "step", map[string]any{
		"agent": "orchestrator", "label": "Starting multi-agent run", "at": time.Now().UTC().Format(time.RFC3339),
	})

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
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, nil, "crewai_submit_failed", int(time.Since(start).Milliseconds()))
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": "crewai_submit_failed"})
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		return
	}
	_ = a.store.UpdateImproveRunCrewTask(bg, runID, out.TaskID)

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
	failed := false
	errorCode := ""
	cancelled := false
	for ev := range events {
		if a.improveRunWasCancelled(bg, runID) {
			cancelled = true
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
					Markdown  string `json:"markdown"`
					Citations any    `json:"citations"`
				} `json:"result"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			finalMarkdown = strings.TrimSpace(payload.Report)
			if finalMarkdown == "" {
				finalMarkdown = strings.TrimSpace(payload.Result.Markdown)
			}
			citations = payload.Result.Citations
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
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, citations, errorCode, latency)
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
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
		_ = a.store.CompleteImproveRun(bg, runID, store.ImproveRunFailed, citations, "persist_failed", latency)
		a.improveHub.publish(runID, "error", map[string]any{"errorCode": "persist_failed"})
		a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageError)
		return
	}
	if err := a.store.CompleteImproveRun(bg, runID, store.ImproveRunCompleted, citations, "", latency); err != nil {
		// Already cancelled / terminal elsewhere.
		return
	}
	a.trackAiCrUsage(practiceID, userID, visitID, store.AiCrUsageImproveAdvanced)
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
			b, _ := json.Marshal(steps[i])
			fmt.Fprintf(w, "id: %d\nevent: step\ndata: %s\n\n", i+1, string(b))
			flusher.Flush()
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
