package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type streamCtl struct {
	Type            string `json:"type"` // ready | delta | turn_complete | interrupted | ended | error
	Delta           string `json:"delta,omitempty"`
	Reply           string `json:"reply,omitempty"`
	Action          string `json:"action,omitempty"`
	Outcome         string `json:"outcome,omitempty"`
	AppointmentSlot string `json:"appointmentSlot,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Ended           bool   `json:"ended,omitempty"`
	Interrupted     bool   `json:"interrupted,omitempty"`
	Code            string `json:"code,omitempty"`
	Text            string `json:"text,omitempty"`
}

// commercialPitchSimStream: text token streaming fallback when Live audio is unavailable.
func (a *API) commercialPitchSimStream(w http.ResponseWriter, r *http.Request) {
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
	if a.gemini == nil || !a.gemini.Configured() {
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

	ctx, cancel := context.WithTimeout(context.Background(), remaining)
	defer cancel()

	sess := &textStreamSession{
		api:    a,
		sim:    sim,
		client: client,
		system: gemini.BuildVetStreamPrompt(vetPrompt.ContentJSON, sim.InterestLevel),
	}
	sess.loadTranscript()
	_ = sess.send(ctx, streamCtl{Type: "ready"})
	sess.run(ctx, cancel)
}

type textStreamSession struct {
	api    *API
	sim    store.PitchSimulation
	client *websocket.Conn
	system string

	mu         sync.Mutex
	writeMu    sync.Mutex
	history    []map[string]string
	turnCancel context.CancelFunc
	turnBusy   bool
}

func (s *textStreamSession) loadTranscript() {
	_ = json.Unmarshal(s.sim.TranscriptJSON, &s.history)
	if s.history == nil {
		s.history = []map[string]string{}
	}
}

func (s *textStreamSession) send(ctx context.Context, msg streamCtl) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return sendStreamCtl(ctx, s.client, msg)
}

func (s *textStreamSession) run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer s.client.Close(websocket.StatusNormalClosure, "done")

	for {
		_, data, err := s.client.Read(ctx)
		if err != nil {
			return
		}
		var msg struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "end":
			return
		case "interrupt":
			s.mu.Lock()
			if s.turnCancel != nil {
				s.turnCancel()
			}
			s.mu.Unlock()
			_ = s.send(ctx, streamCtl{Type: "interrupted"})
		case "user":
			text := strings.TrimSpace(msg.Text)
			if text == "" {
				continue
			}
			s.mu.Lock()
			simSnap := s.sim
			busy := s.turnBusy
			s.mu.Unlock()
			if busy {
				continue
			}
			if s.api.store.PitchCallTimedOut(simSnap) {
				_ = s.send(ctx, streamCtl{Type: "ended", Outcome: "timeout", Ended: true})
				cancel()
				return
			}
			// Goroutine : laisse le Read loop traiter interrupt pendant le stream.
			go s.handleUserTurn(ctx, cancel, text)
		}
	}
}

func (s *textStreamSession) handleUserTurn(parent context.Context, sessionCancel context.CancelFunc, userText string) {
	s.mu.Lock()
	if s.turnBusy {
		s.mu.Unlock()
		return
	}
	if s.turnCancel != nil {
		s.turnCancel()
	}
	turnCtx, cancel := context.WithCancel(parent)
	s.turnCancel = cancel
	s.turnBusy = true
	s.history = append(s.history, map[string]string{"role": "commercial", "text": userText})
	histCopy := append([]map[string]string(nil), s.history...)
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		s.turnBusy = false
		s.mu.Unlock()
	}()

	turn, err := s.api.gemini.VetTurnStream(turnCtx, s.system, histCopy, userText, func(delta string) error {
		if turnCtx.Err() != nil {
			return turnCtx.Err()
		}
		return s.send(parent, streamCtl{Type: "delta", Delta: delta})
	})
	if turn == nil {
		_ = s.send(parent, streamCtl{Type: "error", Code: "gemini_failed", Text: errString(err)})
		return
	}

	interrupted := turnCtx.Err() != nil

	s.mu.Lock()
	s.history = append(s.history, map[string]string{"role": "vet", "text": turn.Reply})
	tr, _ := json.Marshal(s.history)
	s.sim.TranscriptJSON = tr
	ended, outcome, slot := false, "in_progress", ""
	if !interrupted {
		ended, outcome, slot = applyVetTurnOutcome(&s.sim, turn)
	}
	simCopy := s.sim
	s.mu.Unlock()

	if ended {
		if err := s.api.store.FinalizePitchSimulation(parent, simCopy); err != nil {
			_ = s.send(parent, streamCtl{Type: "error", Code: "persist_failed"})
			return
		}
		s.mu.Lock()
		s.sim = simCopy
		s.mu.Unlock()
	} else {
		_ = s.api.store.UpdatePitchSimulationTranscript(parent, simCopy.ID, simCopy.UserID, tr)
	}

	_ = s.send(parent, streamCtl{
		Type:            "turn_complete",
		Reply:           turn.Reply,
		Action:          turn.Action,
		AppointmentSlot: slot,
		Reason:          turn.Reason,
		Ended:           ended,
		Outcome:         outcome,
		Interrupted:     interrupted,
	})
	if ended {
		_ = s.send(parent, streamCtl{
			Type: "ended", Outcome: outcome, AppointmentSlot: slot, Ended: true,
		})
		sessionCancel()
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func sendStreamCtl(ctx context.Context, conn *websocket.Conn, msg streamCtl) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return conn.Write(wctx, websocket.MessageText, data)
}

// applyVetTurnOutcome mutates sim for book/hangup; returns ended, outcome, slot.
func applyVetTurnOutcome(sim *store.PitchSimulation, turn *gemini.VetTurnResult) (ended bool, outcome, slot string) {
	outcome = "in_progress"
	slot = ""
	if turn == nil {
		return false, outcome, slot
	}
	switch turn.Action {
	case "book_appointment":
		ended = true
		outcome = "appointment"
		slot = turn.AppointmentSlot
		if slot == "" {
			slot = "Créneau démo proposé"
		}
	case "hang_up_not_interested":
		ended = true
		outcome = "hangup"
	}
	if ended {
		now := time.Now().UTC()
		sim.Outcome = outcome
		sim.AppointmentSlot = slot
		sim.EndedAt = &now
		sim.DurationSec = int(now.Sub(sim.CreatedAt).Seconds())
	}
	return ended, outcome, slot
}
