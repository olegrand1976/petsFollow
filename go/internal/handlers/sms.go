package handlers

import (
	"context"
	"log"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/sms"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type smsKind string

const (
	smsKindVisitConfirmed  smsKind = "visit_confirmed"
	smsKindVisitReminder   smsKind = "visit_reminder"
	smsKindVisitReschedule smsKind = "visit_reschedule"
)

// Statuts et raisons de skip journalisés dans notifications.sms_log.
const (
	smsStatusSent    = "sent"
	smsStatusDryRun  = "dry_run"
	smsStatusSkipped = "skipped"
	smsStatusError   = "error"

	smsSkipPrefOptOut   = "pref_opt_out"
	smsSkipNoPhone      = "no_phone"
	smsSkipInvalidPhone = "invalid_phone"
)

// smsRecipient rassemble ce qu'il faut pour envoyer : gates prefs passées,
// téléphone E.164 et locale résolue.
type smsRecipient struct {
	Phone  string
	Locale string
}

// resolveSMSRecipient applique le double gate (topic visits + canal sms), lit le
// téléphone et le normalise en E.164. Un skip journalise sa raison et renvoie ok=false.
func (a *API) resolveSMSRecipient(ctx context.Context, kind smsKind, userID, visitID string, scheduledFor *time.Time) (smsRecipient, bool) {
	logSkip := func(reason, phone string) {
		if _, err := a.store.InsertSmsLog(ctx, store.SmsLogEntry{
			UserID: userID, VisitID: visitID, Kind: string(kind), ToPhone: phone,
			Locale: "fr", ScheduledFor: scheduledFor, Status: smsStatusSkipped, Error: reason,
		}); err != nil {
			log.Printf("sms: log skip %s user %s: %v", reason, userID, err)
		}
	}
	prefs, err := a.store.GetClientNotificationPrefs(ctx, userID)
	if err != nil {
		log.Printf("sms: prefs user %s: %v", userID, err)
		return smsRecipient{}, false
	}
	if !prefs.Visits || !prefs.SMS {
		logSkip(smsSkipPrefOptOut, "")
		return smsRecipient{}, false
	}
	client, err := a.store.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("sms: user %s: %v", userID, err)
		return smsRecipient{}, false
	}
	if client.ContactPhone == "" {
		logSkip(smsSkipNoPhone, "")
		return smsRecipient{}, false
	}
	phone, ok := sms.NormalizeE164(client.ContactPhone, a.cfg.SMSDefaultRegion)
	if !ok {
		// Numéro brut conservé (tronqué) pour diagnostiquer les formats rejetés.
		logSkip(smsSkipInvalidPhone, truncateRunes(client.ContactPhone, 40))
		return smsRecipient{}, false
	}
	return smsRecipient{Phone: phone, Locale: a.clientLocale(ctx, userID)}, true
}

// smsBody rend le gabarit localisé du kind.
func smsBody(kind smsKind, locale, petName string, when time.Time) string {
	if petName == "" {
		petName = "…"
	}
	return i18n.T(locale, "sms."+string(kind), map[string]string{
		"petName": petName,
		"when":    formatWhenBrussels(when),
	})
}

// formatWhenBrussels formate un créneau comme les emails/alertes visites.
func formatWhenBrussels(t time.Time) string {
	loc, _ := time.LoadLocation("Europe/Brussels")
	if loc == nil {
		loc = time.Local
	}
	return t.In(loc).Format("02/01/2006 15:04")
}

// sendVisitSMS envoie un SMS événementiel (confirmation / reprogrammation) et
// journalise l'issue. Synchrone : les wrappers async l'appellent en goroutine.
func (a *API) sendVisitSMS(ctx context.Context, kind smsKind, userID, visitID, petName string, when time.Time) {
	if !a.cfg.SMSEnabled || a.sms == nil || userID == "" {
		return
	}
	rcpt, ok := a.resolveSMSRecipient(ctx, kind, userID, visitID, nil)
	if !ok {
		return
	}
	res, err := a.sms.Send(ctx, rcpt.Phone, smsBody(kind, rcpt.Locale, petName, when))
	entry := store.SmsLogEntry{
		UserID: userID, VisitID: visitID, Kind: string(kind),
		ToPhone: rcpt.Phone, Locale: rcpt.Locale,
	}
	switch {
	case err != nil:
		entry.Status = smsStatusError
		entry.Error = truncateRunes(err.Error(), 200)
		log.Printf("sms: send %s user %s: %v", kind, userID, err)
	case res.DryRun:
		entry.Status = smsStatusDryRun
		entry.ProviderMessageID = res.ProviderMessageID
	default:
		entry.Status = smsStatusSent
		entry.ProviderMessageID = res.ProviderMessageID
	}
	if _, logErr := a.store.InsertSmsLog(ctx, entry); logErr != nil {
		log.Printf("sms: log %s user %s: %v", kind, userID, logErr)
	}
}

// smsVisitConfirmed / smsVisitReschedule — wrappers async (jamais bloquants
// pour la requête HTTP), miroirs de notifyClientPushAsync.
// smsVisitConfirmed annonce le créneau ferme : scheduled_at fait foi sur une
// visite confirmée (proposed_scheduled_at est purgé par le confirm/reschedule direct).
func (a *API) smsVisitConfirmed(pet store.Pet, visit store.Visit) {
	if visit.ScheduledAt == nil {
		return
	}
	a.smsVisitAsync(smsKindVisitConfirmed, pet, visit, *visit.ScheduledAt)
}

// smsVisitReschedule annonce le créneau proposé (ProposeReschedule) ou, à défaut,
// le nouveau créneau imposé (RescheduleVisitDirect purge proposed_scheduled_at).
func (a *API) smsVisitReschedule(pet store.Pet, visit store.Visit) {
	when := visit.ProposedScheduledAt
	if when == nil {
		when = visit.ScheduledAt
	}
	if when == nil {
		return
	}
	a.smsVisitAsync(smsKindVisitReschedule, pet, visit, *when)
}

func (a *API) smsVisitAsync(kind smsKind, pet store.Pet, visit store.Visit, when time.Time) {
	if !a.cfg.SMSEnabled || a.sms == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		a.sendVisitSMS(ctx, kind, pet.OwnerUserID, visit.ID, pet.Name, when)
	}()
}
