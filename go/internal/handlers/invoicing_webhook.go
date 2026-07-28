package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing/billit"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func (a *API) registerInvoicingWebhookRoutes(r chi.Router) {
	if a.invoicing == nil || !a.cfg.BillitEnabled {
		return
	}
	rl := a.billitWebhookRL
	if rl == nil {
		rl = httpx.NewRateLimiter(120, time.Minute)
	}
	r.With(rl.Middleware).Post("/invoicing/webhooks/billit", a.billitWebhook)
}

func (a *API) billitWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_body")
		return
	}
	sig := r.Header.Get("X-Billit-Signature")
	if sig == "" {
		sig = r.Header.Get("X-Signature")
	}
	if !billit.VerifyWebhookSignature(a.cfg.BillitWebhookSecret, payload, sig) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_signature")
		return
	}
	orderID, eventType, status, peppolStatus, err := billit.ParseWebhook(payload)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_webhook")
		return
	}
	dedupeKey := billit.WebhookDedupeKey(payload)
	// Empty event_type: dedupe key already uniquely identifies the delivery (evt id or body hash).
	eventID, dup, err := a.invoicing.RecordWebhook(r.Context(), "", dedupeKey, payload)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	applyErr := a.invoicing.ApplyWebhookStatus(r.Context(), orderID, eventType, status, peppolStatus)
	if applyErr != nil {
		if !dup {
			// Allow Billit redelivery: drop idempotency row. If Forget fails, the
			// duplicate path below still re-applies on the next attempt.
			_ = a.invoicing.ForgetWebhook(r.Context(), eventID)
		}
		if errors.Is(applyErr, invoicing.ErrDocNotFound) {
			writeErr(w, r, http.StatusServiceUnavailable, "temporary", "document_not_ready")
			return
		}
		writeErr(w, r, http.StatusServiceUnavailable, "temporary", "webhook_apply_failed")
		return
	}
	if !dup {
		_ = a.invoicing.MarkWebhookProcessed(r.Context(), eventID, nil)
	}
	statusLabel := "ok"
	if dup {
		// Exact redelivery (or Forget failed after a prior apply error): apply is
		// idempotent (FOR UPDATE + usage only on first delivered).
		statusLabel = "duplicate"
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": statusLabel})
}
