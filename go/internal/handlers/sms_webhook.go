package handlers

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/sms"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Chemins des deux webhooks à coller dans le Messaging Profile Telnyx.
// Même traitement des deux côtés : seul le marquage via_failover diffère, ce qui
// rend visible une défaillance de la primaire (sinon la bascule est silencieuse).
const (
	smsWebhookPath         = "/notifications/webhooks/telnyx"
	smsWebhookFailoverPath = "/notifications/webhooks/telnyx/failover"
)

// smsInboundBodyMax borne le corps conservé d'un SMS entrant (minimisation).
const smsInboundBodyMax = 160

func (a *API) registerSMSWebhookRoutes(r chi.Router) {
	if !a.cfg.SMSEnabled {
		return
	}
	rl := a.smsWebhookRL
	if rl == nil {
		rl = httpx.NewRateLimiter(120, time.Minute)
	}
	r.With(rl.Middleware).Post(smsWebhookPath, a.telnyxWebhook(false))
	r.With(rl.Middleware).Post(smsWebhookFailoverPath, a.telnyxWebhook(true))
}

// telnyxWebhook traite les DLR (SMS sortants) et les SMS entrants (dont STOP).
// Répond 2xx dès que l'événement est enregistré : un non-2xx déclencherait les
// retries Telnyx puis la bascule sur l'URL de failover.
func (a *API) telnyxWebhook(viaFailover bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_body")
			return
		}
		if err := sms.VerifyWebhookSignature(
			a.cfg.TelnyxPublicKey,
			r.Header.Get(sms.SignatureHeader),
			r.Header.Get(sms.TimestampHeader),
			body, time.Now(),
		); err != nil {
			log.Printf("sms webhook: signature rejected (failover=%v): %v", viaFailover, err)
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_signature")
			return
		}
		ev, err := sms.ParseWebhook(body)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_webhook")
			return
		}
		if viaFailover {
			// Signal opérationnel : la primaire n'a pas répondu 2xx.
			log.Printf("sms webhook: event %s received via FAILOVER url", ev.EventType)
		}

		switch ev.EventType {
		case sms.EventMessageSent, sms.EventMessageFinalized:
			a.applySmsDeliveryReport(r, ev)
		case sms.EventMessageReceived:
			a.handleSmsInbound(r, ev, viaFailover)
		default:
			// Événement non exploité (ex. message.queued) : acquitté sans traitement.
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (a *API) applySmsDeliveryReport(r *http.Request, ev sms.WebhookEvent) {
	var deliveredAt *time.Time
	if ev.Status == "delivered" {
		at := ev.OccurredAt
		if at.IsZero() {
			at = time.Now()
		}
		deliveredAt = &at
	}
	errMsg := ""
	if sms.DeliveryFailed(ev.Status) {
		errMsg = truncateRunes(ev.ErrorCode, 200)
	}
	matched, err := a.store.ApplySmsDeliveryReport(r.Context(), ev.MessageID, ev.Status, errMsg, deliveredAt)
	if err != nil {
		log.Printf("sms webhook: apply DLR %s: %v", ev.MessageID, err)
		return
	}
	if !matched {
		log.Printf("sms webhook: DLR for unknown message %s (status=%s)", ev.MessageID, ev.Status)
	}
}

// handleSmsInbound journalise l'entrant et applique STOP/START sur la pref canal.
// Un numéro qui ne correspond pas exactement à un compte est conservé sans
// rattachement : on ne coupe jamais le canal d'un client sur une correspondance douteuse.
func (a *API) handleSmsInbound(r *http.Request, ev sms.WebhookEvent, viaFailover bool) {
	ctx := r.Context()
	command := sms.InboundCommand(ev.Text)
	userID, err := a.store.FindUserIDByContactPhone(ctx, ev.FromPhone)
	if err != nil {
		log.Printf("sms webhook: match phone: %v", err)
	}
	inserted, err := a.store.InsertSmsInbound(ctx, store.SmsInboundEntry{
		UserID:            userID,
		FromPhone:         ev.FromPhone,
		ToPhone:           ev.ToPhone,
		Body:              truncateRunes(ev.Text, smsInboundBodyMax),
		Command:           command,
		ProviderMessageID: ev.MessageID,
		ViaFailover:       viaFailover,
	})
	if err != nil {
		log.Printf("sms webhook: log inbound: %v", err)
		return
	}
	if !inserted {
		return // relivraison déjà traitée
	}
	if userID == "" {
		if command != "other" {
			log.Printf("sms webhook: %s from unmatched number — channel left untouched", command)
		}
		return
	}
	switch command {
	case "stop":
		if err := a.store.SetClientSMSPref(ctx, userID, false); err != nil {
			log.Printf("sms webhook: opt-out user %s: %v", userID, err)
		}
	case "start":
		if err := a.store.SetClientSMSPref(ctx, userID, true); err != nil {
			log.Printf("sms webhook: opt-in user %s: %v", userID, err)
		}
	}
}
