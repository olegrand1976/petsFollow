package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Fenêtre du rappel : début fixe (ne pas prévenir un RDV imminent), fin réglable
// par VISIT_REMINDER_LOOKAHEAD_HOURS. Avec le cron quotidien 17:00 Europe/Brussels
// et 30h de fenêtre, chaque RDV du lendemain est couvert exactement une fois ;
// l'index unique sms_log_reminder_once rend tout run manqué/rejoué inoffensif.
const visitReminderLeadMin = time.Hour

const visitReminderBatch = 500

// internalRunVisitReminders — job cron : rappel J-1 des RDV confirmés (push + SMS).
// Protégé par le header X-Visit-Reminders-Secret (env VISIT_REMINDERS_SECRET).
func (a *API) internalRunVisitReminders(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Visit-Reminders-Secret", a.cfg.VisitRemindersSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	lookahead := time.Duration(a.cfg.VisitReminderLookaheadHours) * time.Hour
	if lookahead <= visitReminderLeadMin {
		lookahead = 30 * time.Hour
	}
	now := time.Now()
	from := now.Add(visitReminderLeadMin)
	to := now.Add(lookahead)
	candidates, err := a.store.ListVisitsForReminder(r.Context(), from, to, visitReminderBatch)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var sent, dryRun, skipped, errored, pushed int
	for _, c := range candidates {
		// Push : gratuit, gate prefs.Visits appliquée dans notifyClientPush.
		a.pushVisitReminder(c.OwnerUserID, c.VisitID, c.PetID, c.PetName)
		pushed++

		switch a.sendVisitReminderSMS(r.Context(), c) {
		case smsStatusSent:
			sent++
		case smsStatusDryRun:
			dryRun++
		case smsStatusSkipped:
			skipped++
		case smsStatusError:
			errored++
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"candidates": len(candidates),
		"pushSent":   pushed,
		"smsSent":    sent,
		"smsDryRun":  dryRun,
		"smsSkipped": skipped,
		"smsErrors":  errored,
		"windowFrom": from,
		"windowTo":   to,
	})
}

// sendVisitReminderSMS réserve puis envoie le rappel SMS. Le claim (index unique
// sur visit_id + scheduled_for) garantit au-plus-une-fois même en concurrence.
// Retourne le statut journalisé, "" si le module est off ou le claim perdu.
func (a *API) sendVisitReminderSMS(ctx context.Context, c store.VisitReminderCandidate) string {
	if !a.cfg.SMSEnabled || a.sms == nil {
		return ""
	}
	// Les skips (opt-out, téléphone absent/invalide) sont journalisés sans claim
	// (scheduled_for NULL) : ils ne consomment pas l'index unique.
	rcpt, ok := a.resolveSMSRecipient(ctx, smsKindVisitReminder, c.OwnerUserID, c.VisitID, nil)
	if !ok {
		return smsStatusSkipped
	}
	logID, claimed, err := a.store.ClaimVisitReminderSms(ctx, c.VisitID, c.ScheduledAt, c.OwnerUserID, rcpt.Phone, rcpt.Locale)
	if err != nil {
		log.Printf("sms: claim reminder visit %s: %v", c.VisitID, err)
		return smsStatusError
	}
	if !claimed {
		return "" // déjà traité par un autre run
	}
	res, sendErr := a.sms.Send(ctx, rcpt.Phone, smsBody(smsKindVisitReminder, rcpt.Locale, c.PetName, c.ScheduledAt))
	status, providerID, errMsg := smsStatusSent, res.ProviderMessageID, ""
	switch {
	case sendErr != nil:
		status, providerID, errMsg = smsStatusError, "", truncateRunes(sendErr.Error(), 200)
		log.Printf("sms: send reminder visit %s: %v", c.VisitID, sendErr)
	case res.DryRun:
		status = smsStatusDryRun
	}
	if err := a.store.FinalizeSmsLog(ctx, logID, status, providerID, errMsg); err != nil {
		log.Printf("sms: finalize reminder visit %s: %v", c.VisitID, err)
	}
	return status
}
