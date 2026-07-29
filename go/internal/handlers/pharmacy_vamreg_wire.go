package handlers

import (
	"context"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// VamregEnqueuer schedules VAMReg declaration (Asynq or inline dry-run path).
type VamregEnqueuer interface {
	EnqueueDeclare(ctx context.Context, practiceID, dafID string) error
}

// inlineVamregEnqueue runs ProcessDeclare in-process with a detached timeout so
// chi request cancel cannot erase failure audits (default when workers off).
type inlineVamregEnqueue struct {
	decl    *pharmacy.VamregDeclarer
	timeout time.Duration
}

func (s inlineVamregEnqueue) EnqueueDeclare(ctx context.Context, practiceID, dafID string) error {
	if s.decl == nil {
		return nil
	}
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 25 * time.Second
	}
	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	return s.decl.ProcessDeclare(runCtx, practiceID, dafID)
}

// VamregDeclarer returns the canonical declarer (shared with Asynq workers).
func (a *API) VamregDeclarer() *pharmacy.VamregDeclarer { return a.vamreg }

// SetVamregEnqueuer replaces the inline enqueuer with Asynq (or test double).
func (a *API) SetVamregEnqueuer(q VamregEnqueuer) {
	if q != nil {
		a.vamregQ = q
	}
}
