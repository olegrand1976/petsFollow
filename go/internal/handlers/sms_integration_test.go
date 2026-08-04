package handlers_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/sms"
)

// fakeSMSSender enregistre les envois au lieu d'appeler Telnyx.
type fakeSMSSender struct {
	mu    sync.Mutex
	sent  []fakeSMS
	fail  error
	msgID string
}

type fakeSMS struct {
	To   string
	Text string
}

func (f *fakeSMSSender) Send(_ context.Context, to, text string) (sms.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return sms.Result{}, f.fail
	}
	f.sent = append(f.sent, fakeSMS{To: to, Text: text})
	id := f.msgID
	if id == "" {
		id = "msg_fake"
	}
	return sms.Result{ProviderMessageID: id}, nil
}

func (f *fakeSMSSender) calls() []fakeSMS {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]fakeSMS, len(f.sent))
	copy(out, f.sent)
	return out
}

// setClientPhone pose un contact_phone déterministe pour rendre le test hermétique.
// Le numéro d'un client est renseigné par le cabinet (PATCH /clients/{id}) : un
// client ne peut pas modifier le sien (contact_phone_role côté PATCH /me).
func setClientPhone(t *testing.T, api *testAPI, staffTok, clientUserID, phone string) {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientUserID, staffTok, map[string]any{
		"contactPhone": phone,
	})
	if code != http.StatusOK {
		t.Fatalf("set phone %d %#v", code, env)
	}
}

// clientUserID lit l'id du compte connecté (GET /me).
func clientUserID(t *testing.T, api *testAPI, tok string) string {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	id, _ := dataMap(t, env)["userId"].(string)
	if id == "" {
		t.Fatalf("no user id in /me: %#v", dataMap(t, env))
	}
	return id
}

const demoClientPhone = "0470 00 00 01"

func setSMSPref(t *testing.T, api *testAPI, clientTok string, enabled bool) {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me/notification-preferences", clientTok, map[string]any{
		"sms": enabled,
	})
	if code != http.StatusOK {
		t.Fatalf("set sms pref %d %#v", code, env)
	}
	if got, ok := dataMap(t, env)["sms"].(bool); !ok || got != enabled {
		t.Fatalf("sms pref want %v got %#v", enabled, dataMap(t, env)["sms"])
	}
}

// waitSmsLog attend une ligne sms_log (les hooks événementiels sont asynchrones).
func waitSmsLog(t *testing.T, api *testAPI, visitID, kind string) (string, string, string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var status, errMsg, phone string
		err := api.pool.QueryRow(context.Background(), `
			SELECT status, error, to_phone FROM notifications.sms_log
			WHERE visit_id = $1 AND kind = $2
			ORDER BY created_at DESC LIMIT 1`, visitID, kind).Scan(&status, &errMsg, &phone)
		if err == nil {
			return status, errMsg, phone
		}
		time.Sleep(75 * time.Millisecond)
	}
	t.Fatalf("no sms_log row for visit %s kind %s", visitID, kind)
	return "", "", ""
}

func countSmsLog(t *testing.T, api *testAPI, visitID, kind string) int {
	t.Helper()
	var n int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT count(*) FROM notifications.sms_log WHERE visit_id = $1 AND kind = $2`,
		visitID, kind).Scan(&n); err != nil {
		t.Fatalf("count sms_log: %v", err)
	}
	return n
}

// createConfirmedVisit crée une visite déjà confirmée (chemin confirmDirect).
// silentConfirm=false pour déclencher les hooks de notification.
func createConfirmedVisit(t *testing.T, api *testAPI, staffTok, petID string) string {
	t.Helper()
	for i := 0; i < 8; i++ {
		slot := time.Now().UTC().Add(time.Duration(26+i*3)*time.Hour + 13*time.Minute).Truncate(time.Minute).Format(time.RFC3339)
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", staffTok, map[string]any{
			"scheduledAt":     slot,
			"durationMinutes": 30,
			"confirmDirect":   true,
		})
		if code == http.StatusCreated {
			id, _ := dataMap(t, env)["id"].(string)
			return id
		}
		if code != http.StatusConflict && code != http.StatusBadRequest {
			t.Fatalf("create visit %d %#v", code, env)
		}
	}
	t.Skip("no free slot for a confirmed visit after retries")
	return ""
}

func firstClientPetID(t *testing.T, api *testAPI, clientTok string) string {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets for demo client")
	}
	id, _ := pets[0].(map[string]any)["id"].(string)
	return id
}

func TestVisitConfirmedSendsSMS(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{msgID: "msg_confirmed"}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)

	petID := firstClientPetID(t, api, clientTok)
	visitID := createConfirmedVisit(t, api, vetTok, petID)

	status, errMsg, phone := waitSmsLog(t, api, visitID, "visit_confirmed")
	if status != "sent" {
		t.Fatalf("status want sent got %q (err=%q)", status, errMsg)
	}
	if phone != "+32470000001" {
		t.Fatalf("phone want E.164 got %q", phone)
	}
	calls := fake.calls()
	if len(calls) == 0 {
		t.Fatal("sender not called")
	}
	if calls[0].To != "+32470000001" || calls[0].Text == "" {
		t.Fatalf("unexpected send: %+v", calls[0])
	}
}

func TestVisitConfirmedRespectsSMSOptOut(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, false)
	t.Cleanup(func() { setSMSPref(t, api, clientTok, true) })

	petID := firstClientPetID(t, api, clientTok)
	visitID := createConfirmedVisit(t, api, vetTok, petID)

	status, errMsg, _ := waitSmsLog(t, api, visitID, "visit_confirmed")
	if status != "skipped" || errMsg != "pref_opt_out" {
		t.Fatalf("want skipped/pref_opt_out got %q/%q", status, errMsg)
	}
	if calls := fake.calls(); len(calls) != 0 {
		t.Fatalf("sender must not be called on opt-out, got %d", len(calls))
	}
}

func TestVisitConfirmedSkipsInvalidPhone(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), "pas un numero")
	setSMSPref(t, api, clientTok, true)
	t.Cleanup(func() { setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone) })

	petID := firstClientPetID(t, api, clientTok)
	visitID := createConfirmedVisit(t, api, vetTok, petID)

	status, errMsg, _ := waitSmsLog(t, api, visitID, "visit_confirmed")
	if status != "skipped" || errMsg != "invalid_phone" {
		t.Fatalf("want skipped/invalid_phone got %q/%q", status, errMsg)
	}
	if calls := fake.calls(); len(calls) != 0 {
		t.Fatalf("sender must not be called on invalid phone, got %d", len(calls))
	}
}

func TestVisitRescheduleDirectSendsSMS(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{msgID: "msg_resched"}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)

	petID := firstClientPetID(t, api, clientTok)
	visitID := createConfirmedVisit(t, api, vetTok, petID)

	var moved bool
	for i := 0; i < 8; i++ {
		next := time.Now().UTC().Add(time.Duration(60+i*3)*time.Hour + 21*time.Minute).Truncate(time.Minute).Format(time.RFC3339)
		code, _ := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"action":              "reschedule_direct",
			"proposedScheduledAt": next,
		})
		if code == http.StatusOK {
			moved = true
			break
		}
	}
	if !moved {
		t.Skip("no free slot for reschedule after retries")
	}

	status, errMsg, _ := waitSmsLog(t, api, visitID, "visit_reschedule")
	if status != "sent" {
		t.Fatalf("status want sent got %q (err=%q)", status, errMsg)
	}
}

func TestSMSDisabledWritesNoLog(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{}
	api.api.TestSetSMSEnabled(false)
	api.api.TestReplaceSMSSender(fake)
	t.Cleanup(func() { api.api.TestSetSMSEnabled(true) })

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)

	petID := firstClientPetID(t, api, clientTok)
	visitID := createConfirmedVisit(t, api, vetTok, petID)

	// Laisser le temps au hook asynchrone de ne rien faire.
	time.Sleep(400 * time.Millisecond)
	if n := countSmsLog(t, api, visitID, "visit_confirmed"); n != 0 {
		t.Fatalf("module off must not log, got %d rows", n)
	}
	if calls := fake.calls(); len(calls) != 0 {
		t.Fatalf("module off must not send, got %d", len(calls))
	}
}
