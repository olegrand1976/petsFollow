package sms

import (
	"crypto/ed25519"
	"encoding/base64"
	"strconv"
	"testing"
	"time"
)

func signPayload(t *testing.T, priv ed25519.PrivateKey, ts string, body []byte) string {
	t.Helper()
	signed := append([]byte(ts+"|"), body...)
	return base64.StdEncoding.EncodeToString(ed25519.Sign(priv, signed))
}

func TestVerifyWebhookSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pub)
	now := time.Unix(1_800_000_000, 0)
	ts := strconv.FormatInt(now.Unix(), 10)
	body := []byte(`{"data":{"event_type":"message.finalized"}}`)
	sig := signPayload(t, priv, ts, body)

	if err := VerifyWebhookSignature(pubB64, sig, ts, body, now); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}

	// Corps altéré → rejet.
	if err := VerifyWebhookSignature(pubB64, sig, ts, []byte(`{"data":{}}`), now); err == nil {
		t.Fatal("tampered body accepted")
	}
	// Horodatage hors tolérance → rejet (anti-rejeu).
	stale := now.Add(WebhookTolerance + time.Minute)
	if err := VerifyWebhookSignature(pubB64, sig, ts, body, stale); err == nil {
		t.Fatal("stale timestamp accepted")
	}
	// Clé publique d'un autre couple → rejet.
	otherPub, _, _ := ed25519.GenerateKey(nil)
	if err := VerifyWebhookSignature(base64.StdEncoding.EncodeToString(otherPub), sig, ts, body, now); err == nil {
		t.Fatal("wrong public key accepted")
	}
	// Clé absente / signature absente → rejet explicite.
	if err := VerifyWebhookSignature("", sig, ts, body, now); err == nil {
		t.Fatal("missing public key accepted")
	}
	if err := VerifyWebhookSignature(pubB64, "", ts, body, now); err == nil {
		t.Fatal("missing signature accepted")
	}
}

func TestParseWebhookDeliveryReport(t *testing.T) {
	body := []byte(`{"data":{"event_type":"message.finalized","occurred_at":"2026-08-04T10:30:00Z",
		"payload":{"id":"msg_1","direction":"outbound","to":[{"phone_number":"+32470123456","status":"delivered"}]}}}`)
	ev, err := ParseWebhook(body)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if ev.EventType != EventMessageFinalized || ev.MessageID != "msg_1" ||
		ev.Status != "delivered" || ev.ToPhone != "+32470123456" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestParseWebhookInboundWithErrors(t *testing.T) {
	body := []byte(`{"data":{"event_type":"message.received",
		"payload":{"id":"msg_2","direction":"inbound","text":"STOP",
		"from":{"phone_number":"+32470123456"},
		"to":[{"phone_number":"+32460000000","status":"delivered"}],
		"errors":[{"code":"40001","detail":"carrier issue"}]}}}`)
	ev, err := ParseWebhook(body)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if ev.Direction != "inbound" || ev.FromPhone != "+32470123456" || ev.Text != "STOP" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if ev.ErrorCode == "" {
		t.Fatal("expected aggregated error code")
	}
}

func TestParseWebhookRejectsGarbage(t *testing.T) {
	if _, err := ParseWebhook([]byte("not json")); err == nil {
		t.Fatal("expected error on invalid JSON")
	}
	if _, err := ParseWebhook([]byte(`{"data":{}}`)); err == nil {
		t.Fatal("expected error on missing event_type")
	}
}

func TestInboundCommand(t *testing.T) {
	stop := []string{"STOP", "stop", " Stop. ", "ARRET", "arrêt", "unsubscribe", "стоп", "peatu"}
	for _, s := range stop {
		if got := InboundCommand(s); got != "stop" {
			t.Errorf("InboundCommand(%q) = %q, want stop", s, got)
		}
	}
	for _, s := range []string{"START", "unstop", "старт"} {
		if got := InboundCommand(s); got != "start" {
			t.Errorf("InboundCommand(%q) = %q, want start", s, got)
		}
	}
	for _, s := range []string{"merci !", "je serai en retard", ""} {
		if got := InboundCommand(s); got != "other" {
			t.Errorf("InboundCommand(%q) = %q, want other", s, got)
		}
	}
}

func TestDeliveryFailed(t *testing.T) {
	for _, s := range []string{"delivery_failed", "sending_failed", "EXPIRED", "rejected"} {
		if !DeliveryFailed(s) {
			t.Errorf("DeliveryFailed(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"delivered", "sent", "queued", ""} {
		if DeliveryFailed(s) {
			t.Errorf("DeliveryFailed(%q) = true, want false", s)
		}
	}
}
