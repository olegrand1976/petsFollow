package handlers_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/sms"
)

const (
	smsWebhookURL         = "/api/v1/notifications/webhooks/telnyx"
	smsWebhookFailoverURL = "/api/v1/notifications/webhooks/telnyx/failover"
)

// telnyxSigner simule la signature Ed25519 du portail Telnyx.
type telnyxSigner struct {
	pubB64 string
	priv   ed25519.PrivateKey
}

func newTelnyxSigner(t *testing.T) telnyxSigner {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return telnyxSigner{pubB64: base64.StdEncoding.EncodeToString(pub), priv: priv}
}

func (s telnyxSigner) post(t *testing.T, api *testAPI, url, body string) int {
	t.Helper()
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(s.priv, append([]byte(ts+"|"), body...)))
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(sms.SignatureHeader, sig)
	req.Header.Set(sms.TimestampHeader, ts)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	return rec.Code
}

func TestTelnyxWebhookRejectsBadSignature(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)

	body := `{"data":{"event_type":"message.finalized","payload":{"id":"x","to":[{"status":"delivered"}]}}}`
	// Signature absente.
	req := httptest.NewRequest(http.MethodPost, smsWebhookURL, strings.NewReader(body))
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned: want 401 got %d", rec.Code)
	}

	// Signature d'une autre clé.
	other := newTelnyxSigner(t)
	if code := other.post(t, api, smsWebhookURL, body); code != http.StatusUnauthorized {
		t.Fatalf("wrong key: want 401 got %d", code)
	}
}

func TestTelnyxWebhookAppliesDeliveryReport(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)
	msgID := fmt.Sprintf("msg_dlr_%d", time.Now().UnixNano())
	fake := &fakeSMSSender{msgID: msgID}
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
		t.Fatalf("reminders %d %#v", code, data)
	}
	if _, gotID := reminderRowStatus(t, api, visitID); gotID != msgID {
		t.Fatalf("provider id want %q got %q", msgID, gotID)
	}

	body := fmt.Sprintf(`{"data":{"event_type":"message.finalized","occurred_at":"%s",
		"payload":{"id":%q,"direction":"outbound","to":[{"phone_number":"+32470000001","status":"delivered"}]}}}`,
		time.Now().UTC().Format(time.RFC3339), msgID)
	if code := signer.post(t, api, smsWebhookURL, body); code != http.StatusOK {
		t.Fatalf("webhook %d", code)
	}

	var deliveryStatus string
	var deliveredAt *time.Time
	if err := api.pool.QueryRow(context.Background(), `
		SELECT delivery_status, delivered_at FROM notifications.sms_log
		WHERE provider_message_id = $1`, msgID).Scan(&deliveryStatus, &deliveredAt); err != nil {
		t.Fatalf("read dlr: %v", err)
	}
	if deliveryStatus != "delivered" {
		t.Fatalf("delivery_status want delivered got %q", deliveryStatus)
	}
	if deliveredAt == nil {
		t.Fatal("delivered_at not set")
	}
}

func TestTelnyxWebhookRecordsFailedDelivery(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)
	msgID := fmt.Sprintf("msg_fail_%d", time.Now().UnixNano())
	fake := &fakeSMSSender{msgID: msgID}
	api.api.TestSetSMSEnabled(true)
	api.api.TestReplaceSMSSender(fake)
	api.api.TestSetVisitRemindersSecret(testVisitRemindersSecret)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	setClientPhone(t, api, vetTok, clientUserID(t, api, clientTok), demoClientPhone)
	setSMSPref(t, api, clientTok, true)
	petID := firstClientPetID(t, api, clientTok)
	insertConfirmedVisitFixture(t, api, petID, time.Now().Add(20*time.Hour))
	if code, _ := runVisitReminders(t, api, testVisitRemindersSecret); code != http.StatusOK {
		t.Fatalf("reminders %d", code)
	}

	body := fmt.Sprintf(`{"data":{"event_type":"message.finalized",
		"payload":{"id":%q,"direction":"outbound","to":[{"phone_number":"+32470000001","status":"delivery_failed"}],
		"errors":[{"code":"40010","detail":"unreachable handset"}]}}}`, msgID)
	if code := signer.post(t, api, smsWebhookURL, body); code != http.StatusOK {
		t.Fatalf("webhook %d", code)
	}

	var status, errText string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT delivery_status, delivery_error FROM notifications.sms_log
		WHERE provider_message_id = $1`, msgID).Scan(&status, &errText); err != nil {
		t.Fatalf("read dlr: %v", err)
	}
	if status != "delivery_failed" || errText == "" {
		t.Fatalf("want failed status + error detail, got %q/%q", status, errText)
	}
}

// Un STOP entrant coupe le canal SMS du client : c'est l'opt-out légal.
func TestTelnyxWebhookInboundStopOptsOut(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)
	api.api.TestSetSMSEnabled(true)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	userID := clientUserID(t, api, clientTok)
	setClientPhone(t, api, vetTok, userID, demoClientPhone)
	setSMSPref(t, api, clientTok, true)
	t.Cleanup(func() { setSMSPref(t, api, clientTok, true) })

	msgID := fmt.Sprintf("in_stop_%d", time.Now().UnixNano())
	body := fmt.Sprintf(`{"data":{"event_type":"message.received",
		"payload":{"id":%q,"direction":"inbound","text":"STOP",
		"from":{"phone_number":"+32470000001"},"to":[{"phone_number":"+32460000000","status":"delivered"}]}}}`, msgID)
	if code := signer.post(t, api, smsWebhookURL, body); code != http.StatusOK {
		t.Fatalf("webhook %d", code)
	}

	var smsPref bool
	if err := api.pool.QueryRow(context.Background(),
		`SELECT sms FROM notifications.client_preferences WHERE user_id = $1`, userID).Scan(&smsPref); err != nil {
		t.Fatalf("read pref: %v", err)
	}
	if smsPref {
		t.Fatal("STOP must disable the sms channel")
	}

	var command, storedBody string
	var viaFailover bool
	if err := api.pool.QueryRow(context.Background(), `
		SELECT command, body, via_failover FROM notifications.sms_inbound
		WHERE provider_message_id = $1`, msgID).Scan(&command, &storedBody, &viaFailover); err != nil {
		t.Fatalf("read inbound: %v", err)
	}
	if command != "stop" || storedBody != "STOP" || viaFailover {
		t.Fatalf("unexpected inbound row: %q/%q/%v", command, storedBody, viaFailover)
	}

	// Relivraison du même événement : idempotente (index unique provider_message_id).
	if code := signer.post(t, api, smsWebhookURL, body); code != http.StatusOK {
		t.Fatalf("redelivery %d", code)
	}
	var n int
	if err := api.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM notifications.sms_inbound WHERE provider_message_id = $1`, msgID).Scan(&n); err != nil {
		t.Fatalf("count inbound: %v", err)
	}
	if n != 1 {
		t.Fatalf("redelivery must not duplicate, got %d rows", n)
	}
}

// Un START rouvre le canal.
func TestTelnyxWebhookInboundStartOptsIn(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)
	api.api.TestSetSMSEnabled(true)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	userID := clientUserID(t, api, clientTok)
	setClientPhone(t, api, vetTok, userID, demoClientPhone)
	setSMSPref(t, api, clientTok, false)
	t.Cleanup(func() { setSMSPref(t, api, clientTok, true) })

	msgID := fmt.Sprintf("in_start_%d", time.Now().UnixNano())
	body := fmt.Sprintf(`{"data":{"event_type":"message.received",
		"payload":{"id":%q,"direction":"inbound","text":"START",
		"from":{"phone_number":"+32470000001"}}}}`, msgID)
	if code := signer.post(t, api, smsWebhookURL, body); code != http.StatusOK {
		t.Fatalf("webhook %d", code)
	}
	var smsPref bool
	if err := api.pool.QueryRow(context.Background(),
		`SELECT sms FROM notifications.client_preferences WHERE user_id = $1`, userID).Scan(&smsPref); err != nil {
		t.Fatalf("read pref: %v", err)
	}
	if !smsPref {
		t.Fatal("START must re-enable the sms channel")
	}
}

// L'URL de failover traite le même événement et marque son origine, ce qui rend
// visible une défaillance de la primaire.
func TestTelnyxWebhookFailoverURLMarksOrigin(t *testing.T) {
	api := newTestAPI(t)
	signer := newTelnyxSigner(t)
	api.api.TestSetTelnyxPublicKey(signer.pubB64)
	api.api.TestSetSMSEnabled(true)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	userID := clientUserID(t, api, clientTok)
	setClientPhone(t, api, vetTok, userID, demoClientPhone)
	t.Cleanup(func() { setSMSPref(t, api, clientTok, true) })

	msgID := fmt.Sprintf("in_fail_%d", time.Now().UnixNano())
	body := fmt.Sprintf(`{"data":{"event_type":"message.received",
		"payload":{"id":%q,"direction":"inbound","text":"merci",
		"from":{"phone_number":"+32470000001"}}}}`, msgID)
	if code := signer.post(t, api, smsWebhookFailoverURL, body); code != http.StatusOK {
		t.Fatalf("failover webhook %d", code)
	}
	var viaFailover bool
	var command string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT via_failover, command FROM notifications.sms_inbound
		WHERE provider_message_id = $1`, msgID).Scan(&viaFailover, &command); err != nil {
		t.Fatalf("read inbound: %v", err)
	}
	if !viaFailover {
		t.Fatal("failover origin not recorded")
	}
	if command != "other" {
		t.Fatalf("plain reply must classify as other, got %q", command)
	}
}
