package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testVisitRemindersSecret = "test-visit-reminders-secret"

// decodeData déballe l'enveloppe { data: ... } des endpoints internes (appelés
// sans token, donc hors du helper doAuthJSON).
func decodeData(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	if len(raw) == 0 {
		return nil
	}
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		return map[string]any{"raw": string(raw)}
	}
	if d, ok := env["data"].(map[string]any); ok {
		return d
	}
	return env
}

func runVisitReminders(t *testing.T, api *testAPI, secret string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/visit-reminders/run", nil)
	if secret != "" {
		req.Header.Set("X-Visit-Reminders-Secret", secret)
	}
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	return rec.Code, decodeData(t, rec.Body.Bytes())
}

// insertConfirmedVisitFixture pose une visite confirmée à un créneau donné.
// Le job de rappel lit la base : une fixture évite la validation d'agenda
// (créneaux libres, horaires d'ouverture) qui rendrait le test non déterministe.
func insertConfirmedVisitFixture(t *testing.T, api *testAPI, petID string, at time.Time) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := api.pool.QueryRow(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, site_id, scheduled_at, status, source, consultation_session)
		SELECT gen_random_uuid(), p.id, p.practice_id,
			(SELECT s.id FROM practice.sites s WHERE s.practice_id = p.practice_id AND s.is_primary LIMIT 1),
			$2, 'confirmed', 'vet', FALSE
		FROM pets.pets p WHERE p.id = $1
		RETURNING id::text`, petID, at).Scan(&id)
	if err != nil {
		t.Fatalf("insert visit fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM notifications.sms_log WHERE visit_id = $1`, id)
		_, _ = api.pool.Exec(ctx, `DELETE FROM visits.visits WHERE id = $1`, id)
	})
	return id
}

func TestVisitRemindersRequiresSecret(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	if code, _ := runVisitReminders(t, api, ""); code != http.StatusUnauthorized {
		t.Fatalf("no secret: want 401 got %d", code)
	}
	if code, _ := runVisitReminders(t, api, "wrong-secret"); code != http.StatusUnauthorized {
		t.Fatalf("wrong secret: want 401 got %d", code)
	}
}

func TestVisitRemindersSendsOnceThenIdempotent(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{msgID: "msg_reminder"}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)

	petID := firstClientPetID(t, api, clientTok)
	visitID := insertConfirmedVisitFixture(t, api, petID, time.Now().Add(20*time.Hour))

	// 1er run : la visite est dans la fenêtre now+1h..now+30h, le rappel part.
	code, data := runVisitReminders(t, api, testVisitRemindersSecret)
	if code != http.StatusOK {
		t.Fatalf("run 1 %d %#v", code, data)
	}
	if got := smsLogReminderCount(t, api, visitID); got != 1 {
		t.Fatalf("run 1: want 1 reminder row, got %d (resp=%#v)", got, data)
	}
	sentAfterFirst := len(fake.calls())
	if sentAfterFirst == 0 {
		t.Fatal("run 1: sender not called")
	}
	if status, _ := reminderRowStatus(t, api, visitID); status != "sent" {
		t.Fatalf("run 1: status want sent got %q", status)
	}

	// 2e run : le claim (index unique visit_id+scheduled_for) bloque le doublon.
	code, data = runVisitReminders(t, api, testVisitRemindersSecret)
	if code != http.StatusOK {
		t.Fatalf("run 2 %d %#v", code, data)
	}
	if got := smsLogReminderCount(t, api, visitID); got != 1 {
		t.Fatalf("run 2: reminder must stay unique, got %d rows", got)
	}
	if len(fake.calls()) != sentAfterFirst {
		t.Fatalf("run 2: sender called again (%d -> %d)", sentAfterFirst, len(fake.calls()))
	}
}

// Un créneau déplacé doit pouvoir être rappelé à nouveau : la clé d'idempotence
// est (visite, créneau), pas la visite seule.
func TestVisitRemindersAllowNewSlotAfterReschedule(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{msgID: "msg_reminder"}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)

	petID := firstClientPetID(t, api, clientTok)
	visitID := insertConfirmedVisitFixture(t, api, petID, time.Now().Add(20*time.Hour))

	if code, data := runVisitReminders(t, api, testVisitRemindersSecret); code != http.StatusOK {
		t.Fatalf("run 1 %d %#v", code, data)
	}
	if got := smsLogReminderCount(t, api, visitID); got != 1 {
		t.Fatalf("run 1: want 1 got %d", got)
	}

	// Déplacement du créneau (toujours dans la fenêtre).
	if _, err := api.pool.Exec(context.Background(),
		`UPDATE visits.visits SET scheduled_at = $2 WHERE id = $1`,
		visitID, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("reschedule fixture: %v", err)
	}

	if code, data := runVisitReminders(t, api, testVisitRemindersSecret); code != http.StatusOK {
		t.Fatalf("run 2 %d %#v", code, data)
	}
	if got := smsLogReminderCount(t, api, visitID); got != 2 {
		t.Fatalf("new slot must allow a new reminder: want 2 rows got %d", got)
	}
}

func TestVisitRemindersSkipsOptOutWithoutClaiming(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, false)
	t.Cleanup(func() { setSMSPref(t, api, clientTok, true) })

	petID := firstClientPetID(t, api, clientTok)
	visitID := insertConfirmedVisitFixture(t, api, petID, time.Now().Add(20*time.Hour))

	code, data := runVisitReminders(t, api, testVisitRemindersSecret)
	if code != http.StatusOK {
		t.Fatalf("run %d %#v", code, data)
	}
	if calls := fake.calls(); len(calls) != 0 {
		t.Fatalf("opt-out must not send, got %d", len(calls))
	}
	// Skip journalisé sans scheduled_for : il ne consomme pas l'index unique.
	var status, errMsg string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT status, error FROM notifications.sms_log
		WHERE visit_id = $1 AND kind = 'visit_reminder' AND scheduled_for IS NULL
		ORDER BY created_at DESC LIMIT 1`, visitID).Scan(&status, &errMsg); err != nil {
		t.Fatalf("expected a skipped reminder row: %v", err)
	}
	if status != "skipped" || errMsg != "pref_opt_out" {
		t.Fatalf("want skipped/pref_opt_out got %q/%q", status, errMsg)
	}
}

// Les sessions walk-in et les visites annulées ne déclenchent aucun rappel.
func TestVisitRemindersExcludesWalkInAndCancelled(t *testing.T) {
	api := newTestAPI(t)
	fake := &fakeSMSSender{}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)
	petID := firstClientPetID(t, api, clientTok)

	walkIn := insertConfirmedVisitFixture(t, api, petID, time.Now().Add(21*time.Hour))
	cancelled := insertConfirmedVisitFixture(t, api, petID, time.Now().Add(22*time.Hour))
	ctx := context.Background()
	if _, err := api.pool.Exec(ctx,
		`UPDATE visits.visits SET consultation_session = TRUE WHERE id = $1`, walkIn); err != nil {
		t.Fatalf("mark walk-in: %v", err)
	}
	if _, err := api.pool.Exec(ctx,
		`UPDATE visits.visits SET status = 'cancelled' WHERE id = $1`, cancelled); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if code, data := runVisitReminders(t, api, testVisitRemindersSecret); code != http.StatusOK {
		t.Fatalf("run %d %#v", code, data)
	}
	if got := smsLogReminderCount(t, api, walkIn); got != 0 {
		t.Fatalf("walk-in must not be reminded, got %d", got)
	}
	if got := smsLogReminderCount(t, api, cancelled); got != 0 {
		t.Fatalf("cancelled must not be reminded, got %d", got)
	}
}

// smsLogReminderCount compte les rappels réellement réservés (scheduled_for posé).
func smsLogReminderCount(t *testing.T, api *testAPI, visitID string) int {
	t.Helper()
	var n int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT count(*) FROM notifications.sms_log
		WHERE visit_id = $1 AND kind = 'visit_reminder' AND scheduled_for IS NOT NULL`,
		visitID).Scan(&n); err != nil {
		t.Fatalf("count reminders: %v", err)
	}
	return n
}

func reminderRowStatus(t *testing.T, api *testAPI, visitID string) (string, string) {
	t.Helper()
	var status, providerID string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT status, provider_message_id FROM notifications.sms_log
		WHERE visit_id = $1 AND kind = 'visit_reminder' AND scheduled_for IS NOT NULL
		ORDER BY created_at DESC LIMIT 1`, visitID).Scan(&status, &providerID); err != nil {
		t.Fatalf("reminder row: %v", err)
	}
	return status, providerID
}
