package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerCommissionRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/vet/commissions", a.vetCommissions)
		pr.Get("/admin/commissions/runs", a.adminListCommissionRuns)
		pr.Get("/admin/commissions/periods/{period}", a.adminCommissionPeriod)
		pr.Post("/admin/commissions/periods/{period}/close", a.adminCloseCommissionPeriod)
		pr.Post("/admin/commissions/periods/{period}/mark-paid", a.adminMarkCommissionPaid)
		pr.Put("/admin/commissions/tiers", a.adminPutCommissionTiers)
		pr.Get("/admin/commissions/settings", a.adminGetCommissionSettings)
		pr.Put("/admin/commissions/settings", a.adminPutCommissionSettings)

		pr.Get("/admin/commercial-commissions/runs", a.adminListCommercialCommissionRuns)
		pr.Get("/admin/commercial-commissions/periods/{period}", a.adminCommercialCommissionPeriod)
		pr.Post("/admin/commercial-commissions/periods/{period}/close", a.adminCloseCommercialCommissionPeriod)
		pr.Post("/admin/commercial-commissions/periods/{period}/mark-paid", a.adminMarkCommercialCommissionPaid)
	})
}

func requirePeriodYM(w http.ResponseWriter, r *http.Request) (string, bool) {
	period := chi.URLParam(r, "period")
	if !store.ValidPeriodYM(period) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_period")
		return "", false
	}
	return period, true
}

func (a *API) vetCommissions(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	_ = a.store.EnsureDefaultCommissionTiers(r.Context())
	summary, err := a.store.VetCommissionSummary(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, summary)
}

func (a *API) adminListCommissionRuns(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	_ = a.store.EnsureDefaultCommissionTiers(r.Context())
	_ = a.store.EnsureCommissionSettings(r.Context())
	runs, err := a.store.ListPayoutRuns(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	period := store.PeriodYM(time.Now())
	preview, _ := a.store.PreviewPeriodCommissions(r.Context(), period)
	previewTotal := 0
	for _, l := range preview {
		previewTotal += l.AmountCents
	}
	tiers, _ := a.store.ListCommissionTiers(r.Context())
	if tiers == nil {
		tiers = []store.CommissionTier{}
	}
	commercialRateBps, _ := a.store.GetCommercialRateBps(r.Context())
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"runs":                runs,
		"currentPeriodYm":     period,
		"currentPreviewCents": previewTotal,
		"tiers":               tiers,
		"commercialRateBps":   commercialRateBps,
	})
}

func (a *API) adminCommissionPeriod(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	detail, err := a.store.AdminCommissionPeriodDetail(r.Context(), period)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) adminCloseCommissionPeriod(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	run, err := a.store.ClosePayoutRun(r.Context(), period)
	if err != nil {
		if errors.Is(err, store.ErrPayoutNotOpen) {
			writeErr(w, r, http.StatusConflict, "conflict", "payout_not_open")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, run)
}

type markPaidReq struct {
	Note string `json:"note"`
}

func (a *API) adminMarkCommissionPaid(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	var req markPaidReq
	_ = httpx.DecodeJSON(r, &req)
	run, err := a.store.MarkPayoutRunPaid(r.Context(), period, req.Note)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "run_not_found")
			return
		}
		if errors.Is(err, store.ErrPayoutNotClosed) {
			writeErr(w, r, http.StatusConflict, "conflict", "payout_not_closed")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, run)
}

type putTiersReq struct {
	Tiers []store.CommissionTier `json:"tiers"`
}

func (a *API) adminPutCommissionTiers(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req putTiersReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.ReplaceCommissionTiers(r.Context(), req.Tiers); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	tiers, err := a.store.ListCommissionTiers(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"tiers": tiers})
}

func (a *API) adminGetCommissionSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	_ = a.store.EnsureCommissionSettings(r.Context())
	bps, err := a.store.GetCommercialRateBps(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"commercialRateBps": bps})
}

type putSettingsReq struct {
	CommercialRateBps int `json:"commercialRateBps"`
}

func (a *API) adminPutCommissionSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req putSettingsReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.SetCommercialRateBps(r.Context(), req.CommercialRateBps); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_rate_bps")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"commercialRateBps": req.CommercialRateBps})
}

func (a *API) adminListCommercialCommissionRuns(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	_ = a.store.EnsureCommissionSettings(r.Context())
	runs, err := a.store.ListCommercialPayoutRuns(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	period := store.PeriodYM(time.Now())
	preview, _ := a.store.PreviewCommercialPeriodCommissions(r.Context(), period)
	previewTotal := 0
	for _, l := range preview {
		previewTotal += l.AmountCents
	}
	rateBps, _ := a.store.GetCommercialRateBps(r.Context())
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"runs":                runs,
		"currentPeriodYm":     period,
		"currentPreviewCents": previewTotal,
		"commercialRateBps":   rateBps,
	})
}

func (a *API) adminCommercialCommissionPeriod(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	detail, err := a.store.AdminCommercialCommissionPeriodDetail(r.Context(), period)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, detail)
}

func (a *API) adminCloseCommercialCommissionPeriod(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	run, err := a.store.CloseCommercialPayoutRun(r.Context(), period)
	if err != nil {
		if errors.Is(err, store.ErrPayoutNotOpen) {
			writeErr(w, r, http.StatusConflict, "conflict", "payout_not_open")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, run)
}

func (a *API) adminMarkCommercialCommissionPaid(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	period, ok := requirePeriodYM(w, r)
	if !ok {
		return
	}
	var req markPaidReq
	_ = httpx.DecodeJSON(r, &req)
	run, err := a.store.MarkCommercialPayoutRunPaid(r.Context(), period, req.Note)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "run_not_found")
			return
		}
		if errors.Is(err, store.ErrPayoutNotClosed) {
			writeErr(w, r, http.StatusConflict, "conflict", "payout_not_closed")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, run)
}
