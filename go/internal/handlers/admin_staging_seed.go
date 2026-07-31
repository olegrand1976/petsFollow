package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// StagingSeedConfirmPhrase must be sent as JSON {"confirm":"..."} to run seed.
const StagingSeedConfirmPhrase = "RESET STAGING"

const stagingSeedTimeout = 15 * time.Minute

func (a *API) adminStagingSeedStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"enabled": a.cfg.AdminStagingSeedEnabled,
	})
}

type stagingSeedReq struct {
	Confirm string `json:"confirm"`
}

func (a *API) adminStagingSeed(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if !a.cfg.AdminStagingSeedEnabled {
		writeErr(w, r, http.StatusForbidden, "forbidden", "staging_seed_disabled")
		return
	}
	var req stagingSeedReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.Confirm) != StagingSeedConfirmPhrase {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "confirm_required")
		return
	}

	// Detach from request cancel (LB / client abort) so truncate+seed can finish cleanly.
	seedCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), stagingSeedTimeout)
	defer cancel()

	err := a.store.TryWithAdvisoryLock(seedCtx, store.StagingSeedLockKey, func(ctx context.Context) error {
		return seed.Run(ctx, a.store.Pool())
	})
	if errors.Is(err, store.ErrAdvisoryLockBusy) {
		writeErr(w, r, http.StatusConflict, "conflict", "seed_in_progress")
		return
	}
	if err != nil {
		log.Printf("admin staging seed: %v", err)
		writeErr(w, r, http.StatusInternalServerError, "internal", "seed_failed")
		return
	}

	notified, _ := seed.NotifyStaff(seedCtx, a.store, a.notifier, a.cfg.ProPublicSiteURL, a.cfg.OpsNotifyEmail)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"ok":       true,
		"notified": notified,
	})
}
