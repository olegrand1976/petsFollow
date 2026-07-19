package handlers

import (
	"context"
	"log"
	"time"
	"unicode/utf8"

	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type clientPushKind string

const (
	pushKindMessages clientPushKind = "messages"
	pushKindVisits   clientPushKind = "visits"
)

func prefEnabled(p store.ClientNotificationPrefs, kind clientPushKind) bool {
	switch kind {
	case pushKindMessages:
		return p.Messages
	case pushKindVisits:
		return p.Visits
	default:
		return false
	}
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

// notifyClientPushAsync checks prefs, loads device tokens and sends an FCM push.
// Failures are logged only — callers must not fail the HTTP request.
func (a *API) notifyClientPushAsync(clientUserID string, kind clientPushKind, title, body string, data map[string]string) {
	if clientUserID == "" || a.pusher == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		if err := a.notifyClientPush(ctx, clientUserID, kind, title, body, data); err != nil {
			log.Printf("fcm: notify client %s: %v", clientUserID, err)
		}
	}()
}

func (a *API) notifyClientPush(ctx context.Context, clientUserID string, kind clientPushKind, title, body string, data map[string]string) error {
	prefs, err := a.store.GetClientNotificationPrefs(ctx, clientUserID)
	if err != nil {
		return err
	}
	if !prefEnabled(prefs, kind) {
		return nil
	}
	tokens, err := a.store.ListDeviceTokens(ctx, clientUserID)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}
	tokenStrs := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t.Token != "" {
			tokenStrs = append(tokenStrs, t.Token)
		}
	}
	if len(tokenStrs) == 0 {
		return nil
	}
	invalid, err := a.pusher.Send(ctx, tokenStrs, title, body, data)
	for _, tok := range invalid {
		_ = a.store.DeleteDeviceToken(ctx, clientUserID, tok)
	}
	return err
}

func (a *API) clientLocale(ctx context.Context, userID string) string {
	locale, err := a.store.GetUserPreferredLocale(ctx, userID)
	if err != nil || locale == "" {
		return "fr"
	}
	return i18n.NormalizeLocale(locale)
}

func (a *API) pushNewMessage(clientUserID, threadID, preview string) {
	locale := a.clientLocale(context.Background(), clientUserID)
	preview = truncateRunes(preview, 120)
	if preview == "" {
		preview = "…"
	}
	vars := map[string]string{"preview": preview}
	title := i18n.T(locale, "push.new_message_title", nil)
	body := i18n.T(locale, "push.new_message_body", vars)
	a.notifyClientPushAsync(clientUserID, pushKindMessages, title, body, map[string]string{
		"type":     "message",
		"threadId": threadID,
	})
}

func (a *API) pushVisitConfirmed(clientUserID, visitID, petID, petName string) {
	locale := a.clientLocale(context.Background(), clientUserID)
	if petName == "" {
		petName = "…"
	}
	vars := map[string]string{"petName": petName}
	title := i18n.T(locale, "push.visit_confirmed_title", nil)
	body := i18n.T(locale, "push.visit_confirmed_body", vars)
	a.notifyClientPushAsync(clientUserID, pushKindVisits, title, body, map[string]string{
		"type":    "visit_confirmed",
		"visitId": visitID,
		"petId":   petID,
	})
}
