package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// liveCtl is a JSON control message sent to the browser alongside binary audio frames.
type liveCtl struct {
	Type            string `json:"type"` // ready | transcript | interrupted | ended | error
	Role            string `json:"role,omitempty"`
	Text            string `json:"text,omitempty"`
	Outcome         string `json:"outcome,omitempty"`
	AppointmentSlot string `json:"appointmentSlot,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Code            string `json:"code,omitempty"`
}

// commercialPitchSimLive upgrades to WebSocket and proxies full-duplex audio
// between the browser and Gemini Live for a pitch simulation.
// Auth via ?token= (WS API has no Authorization header from browsers).
func (a *API) commercialPitchSimLive(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	id, err := a.tokens.Parse(token)
	if err != nil || !kernel.IsSalesForce(id.Role) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
		return
	}
	simID := chi.URLParam(r, "id")
	sim, err := a.store.GetPitchSimulation(r.Context(), simID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if sim.Outcome != "in_progress" {
		writeErr(w, r, http.StatusConflict, "conflict", "already_ended")
		return
	}
	if a.cfg.GeminiAPIKey == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
		return
	}
	remaining := store.PitchMaxCallDuration - time.Since(sim.CreatedAt)
	if remaining <= 0 {
		writeErr(w, r, http.StatusConflict, "conflict", "call_timeout")
		return
	}

	vetPrompt, err := a.store.GetAgentPromptVersion(r.Context(), derefStr(sim.VetPromptVersionID))
	if err != nil {
		vetPrompt, err = a.store.GetCurrentAgentPrompt(r.Context(), "vet_live")
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "vet_prompt_missing")
			return
		}
	}

	client, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	client.SetReadLimit(1 << 20)

	// Session context bounded by the server-side hard cap (8 min minus elapsed).
	ctx, cancel := context.WithTimeout(context.Background(), remaining)
	defer cancel()

	live, err := gemini.DialLive(ctx, a.cfg.GeminiAPIKey, gemini.LiveSetup{
		Model:        a.cfg.GeminiLiveModel,
		SystemPrompt: gemini.BuildVetLivePrompt(vetPrompt.ContentJSON, sim.InterestLevel),
		VoiceName:    sim.VoiceName,
		LanguageCode: "fr-FR",
	})
	if err != nil {
		_ = sendCtl(ctx, client, liveCtl{Type: "error", Code: "gemini_live_unavailable"})
		client.Close(websocket.StatusInternalError, "gemini_live_unavailable")
		return
	}
	defer live.Close()

	sess := &liveSimSession{api: a, sim: sim, client: client, live: live}
	sess.loadTranscript()
	sess.run(ctx, cancel)
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// liveSimSession accumulates transcripts and pumps audio both ways.
type liveSimSession struct {
	api    *API
	sim    store.PitchSimulation
	client *websocket.Conn
	live   *gemini.LiveSession

	// Gemini n'accepte du realtimeInput qu'après setupComplete.
	setupOK atomic.Bool

	mu       sync.Mutex
	segments []map[string]string
	curRole  string
	curText  strings.Builder

	ended     bool
	outcome   string
	slot      string
	endReason string
}

func (s *liveSimSession) loadTranscript() {
	_ = json.Unmarshal(s.sim.TranscriptJSON, &s.segments)
	if s.segments == nil {
		s.segments = []map[string]string{}
	}
	// Le "Allo ?" seedé à la création est re-transcrit par Gemini Live — on l'enlève.
	if len(s.segments) == 1 && s.segments[0]["role"] == "vet" && s.segments[0]["text"] == "Allo ?" {
		s.segments = s.segments[:0]
	}
}

func (s *liveSimSession) run(ctx context.Context, cancel context.CancelFunc) {
	// Client → Gemini pump: binary frames are PCM16 16 kHz audio,
	// text frames are JSON control ({"type":"end"} for manual hang up).
	go func() {
		defer cancel()
		for {
			typ, data, err := s.client.Read(ctx)
			if err != nil {
				return
			}
			switch typ {
			case websocket.MessageBinary:
				if !s.setupOK.Load() {
					continue
				}
				if err := s.live.SendAudioChunk(ctx, data); err != nil {
					return
				}
			case websocket.MessageText:
				var msg struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				if json.Unmarshal(data, &msg) != nil {
					continue
				}
				switch msg.Type {
				case "end":
					// Raccrochage manuel : annule le contexte (≠ deadline) → sim reste
					// in_progress, le client déclenche POST /finalize.
					return
				case "text":
					// Degraded input: typed line forwarded as a user turn.
					if strings.TrimSpace(msg.Text) != "" {
						s.appendDelta("commercial", msg.Text)
						_ = s.live.SendUserText(ctx, msg.Text)
					}
				}
			}
		}
	}()

	// Gemini → client pump.
	setupDone := false
	for {
		ev, err := s.live.Recv(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("pitch live %s: gemini recv: %v", s.sim.ID, err)
				if !setupDone {
					// Session refusée par Gemini (modèle/quota) : le client bascule
					// immédiatement en mode tour-par-tour.
					_ = sendCtl(ctx, s.client, liveCtl{Type: "error", Code: "gemini_live_unavailable"})
				}
			}
			break
		}
		if ev.SetupComplete && !setupDone {
			setupDone = true
			s.setupOK.Store(true)
			// Trigger the vet's spoken opening before opening the mic gate client-side.
			_ = s.live.SendUserText(ctx, "(Le téléphone sonne. Tu décroches et tu réponds.)")
			_ = sendCtl(ctx, s.client, liveCtl{Type: "ready"})
			continue
		}
		if ev.Interrupted {
			_ = sendCtl(ctx, s.client, liveCtl{Type: "interrupted"})
		}
		if ev.InputTranscript != "" {
			s.appendDelta("commercial", ev.InputTranscript)
			_ = sendCtl(ctx, s.client, liveCtl{Type: "transcript", Role: "commercial", Text: ev.InputTranscript})
		}
		if ev.OutputTranscript != "" {
			s.appendDelta("vet", ev.OutputTranscript)
			_ = sendCtl(ctx, s.client, liveCtl{Type: "transcript", Role: "vet", Text: ev.OutputTranscript})
		}
		for _, chunk := range ev.AudioChunks {
			wctx, wcancel := context.WithTimeout(ctx, 10*time.Second)
			err := s.client.Write(wctx, websocket.MessageBinary, chunk)
			wcancel()
			if err != nil {
				break
			}
		}
		if ev.TurnComplete {
			s.flushSegment()
			s.persistTranscript()
		}
		if len(ev.ToolCalls) > 0 {
			s.handleToolCalls(ctx, ev.ToolCalls)
		}
		if s.ended || ev.GoAway {
			break
		}
	}

	s.flushSegment()
	// Seul le dépassement du cap 8 min (deadline) vaut timeout ; une annulation
	// (déconnexion client, raccrochage manuel) laisse la simulation in_progress.
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded) && !s.ended

	// Persist final state in a fresh context (session ctx may be done).
	pctx, pcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pcancel()
	s.mu.Lock()
	tr, _ := json.Marshal(s.segments)
	s.mu.Unlock()
	s.sim.TranscriptJSON = tr

	switch {
	case s.ended:
		now := time.Now().UTC()
		s.sim.Outcome = s.outcome
		s.sim.AppointmentSlot = s.slot
		s.sim.EndedAt = &now
		s.sim.DurationSec = int(now.Sub(s.sim.CreatedAt).Seconds())
		_ = s.api.store.FinalizePitchSimulation(pctx, s.sim)
		_ = sendCtl(pctx, s.client, liveCtl{Type: "ended", Outcome: s.outcome, AppointmentSlot: s.slot, Reason: s.endReason})
	case timedOut:
		now := time.Now().UTC()
		s.sim.Outcome = "timeout"
		s.sim.EndedAt = &now
		s.sim.DurationSec = int(store.PitchMaxCallDuration.Seconds())
		_ = s.api.store.FinalizePitchSimulation(pctx, s.sim)
		_ = sendCtl(pctx, s.client, liveCtl{Type: "ended", Outcome: "timeout"})
	default:
		// Manual end or disconnect: keep in_progress, client triggers POST /finalize.
		_ = s.api.store.UpdatePitchSimulationTranscript(pctx, s.sim.ID, s.sim.UserID, tr)
	}
	s.client.Close(websocket.StatusNormalClosure, "done")
}

func (s *liveSimSession) handleToolCalls(ctx context.Context, calls []gemini.LiveToolCall) {
	for _, call := range calls {
		var args struct {
			Slot   string `json:"slot"`
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(call.Args, &args)
		switch call.Name {
		case "book_appointment":
			s.ended = true
			s.outcome = "appointment"
			s.slot = args.Slot
			if s.slot == "" {
				s.slot = "Créneau démo proposé"
			}
		case "hang_up_not_interested":
			s.ended = true
			s.outcome = "hangup"
			s.endReason = args.Reason
		default:
			continue
		}
		_ = s.live.SendToolResponse(ctx, call, map[string]any{"ok": true})
	}
}

func (s *liveSimSession) appendDelta(role, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.curRole != role && s.curText.Len() > 0 {
		s.flushLocked()
	}
	s.curRole = role
	s.curText.WriteString(text)
}

func (s *liveSimSession) flushSegment() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushLocked()
}

func (s *liveSimSession) flushLocked() {
	txt := strings.TrimSpace(s.curText.String())
	if txt != "" {
		s.segments = append(s.segments, map[string]string{"role": s.curRole, "text": txt})
	}
	s.curText.Reset()
}

func (s *liveSimSession) persistTranscript() {
	s.mu.Lock()
	tr, err := json.Marshal(s.segments)
	s.mu.Unlock()
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.api.store.UpdatePitchSimulationTranscript(ctx, s.sim.ID, s.sim.UserID, tr)
}

func sendCtl(ctx context.Context, conn *websocket.Conn, msg liveCtl) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return conn.Write(wctx, websocket.MessageText, data)
}
