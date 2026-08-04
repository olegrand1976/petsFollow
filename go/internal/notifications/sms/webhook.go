package sms

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Headers de signature Telnyx (HTTP : insensibles à la casse).
const (
	SignatureHeader = "telnyx-signature-ed25519"
	TimestampHeader = "telnyx-timestamp"
)

// WebhookTolerance borne l'écart d'horloge accepté : au-delà, l'événement est
// rejeté (protection anti-rejeu).
const WebhookTolerance = 5 * time.Minute

// VerifyWebhookSignature valide la signature Ed25519 d'un webhook Telnyx.
// Telnyx signe la chaîne "<timestamp>|<body>" et transmet la signature et la clé
// publique en base64 (clé publique : portail Telnyx, section Auth).
func VerifyWebhookSignature(publicKeyB64, signatureB64, timestamp string, body []byte, now time.Time) error {
	if publicKeyB64 == "" {
		return fmt.Errorf("telnyx_public_key_missing")
	}
	if signatureB64 == "" || timestamp == "" {
		return fmt.Errorf("telnyx_signature_missing")
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(timestamp), 10, 64)
	if err != nil {
		return fmt.Errorf("telnyx_timestamp_invalid")
	}
	delta := now.Sub(time.Unix(ts, 0))
	if delta < 0 {
		delta = -delta
	}
	if delta > WebhookTolerance {
		return fmt.Errorf("telnyx_timestamp_stale")
	}
	pub, err := base64.StdEncoding.DecodeString(strings.TrimSpace(publicKeyB64))
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("telnyx_public_key_invalid")
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(signatureB64))
	if err != nil || len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("telnyx_signature_invalid")
	}
	signed := append([]byte(strings.TrimSpace(timestamp)+"|"), body...)
	if !ed25519.Verify(ed25519.PublicKey(pub), signed, sig) {
		return fmt.Errorf("telnyx_signature_mismatch")
	}
	return nil
}

// Types d'événements Telnyx Messaging v2 exploités.
const (
	EventMessageSent      = "message.sent"
	EventMessageFinalized = "message.finalized"
	EventMessageReceived  = "message.received"
)

// WebhookEvent est la projection minimale du payload v2 dont on a besoin.
type WebhookEvent struct {
	EventType string
	MessageID string
	Direction string
	// Status : statut du destinataire (queued|sent|delivered|sending_failed|delivery_failed…).
	Status string
	// ErrorCode agrège les erreurs opérateur éventuelles.
	ErrorCode  string
	FromPhone  string
	ToPhone    string
	Text       string
	OccurredAt time.Time
}

type telnyxWebhookEnvelope struct {
	Data struct {
		EventType  string    `json:"event_type"`
		OccurredAt time.Time `json:"occurred_at"`
		Payload    struct {
			ID        string `json:"id"`
			Direction string `json:"direction"`
			Text      string `json:"text"`
			From      struct {
				PhoneNumber string `json:"phone_number"`
			} `json:"from"`
			To []struct {
				PhoneNumber string `json:"phone_number"`
				Status      string `json:"status"`
			} `json:"to"`
			Errors []struct {
				Code   string `json:"code"`
				Detail string `json:"detail"`
			} `json:"errors"`
		} `json:"payload"`
	} `json:"data"`
}

// ParseWebhook extrait l'événement utile du payload v2.
func ParseWebhook(body []byte) (WebhookEvent, error) {
	var env telnyxWebhookEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return WebhookEvent{}, fmt.Errorf("telnyx_webhook_invalid_json: %w", err)
	}
	d := env.Data
	if d.EventType == "" {
		return WebhookEvent{}, fmt.Errorf("telnyx_webhook_missing_event_type")
	}
	ev := WebhookEvent{
		EventType:  d.EventType,
		MessageID:  d.Payload.ID,
		Direction:  d.Payload.Direction,
		Text:       d.Payload.Text,
		FromPhone:  d.Payload.From.PhoneNumber,
		OccurredAt: d.OccurredAt,
	}
	if len(d.Payload.To) > 0 {
		ev.ToPhone = d.Payload.To[0].PhoneNumber
		ev.Status = d.Payload.To[0].Status
	}
	if len(d.Payload.Errors) > 0 {
		parts := make([]string, 0, len(d.Payload.Errors))
		for _, e := range d.Payload.Errors {
			parts = append(parts, strings.TrimSpace(e.Code+" "+e.Detail))
		}
		ev.ErrorCode = strings.Join(parts, "; ")
	}
	return ev, nil
}

// InboundCommand classe un SMS entrant. Les mots-clés couvrent les 8 locales
// servies par l'API ; STOP est le seul universellement attendu par les opérateurs.
func InboundCommand(text string) string {
	t := strings.ToLower(strings.TrimSpace(text))
	t = strings.Trim(t, ".!? \t\n")
	switch t {
	case "stop", "stopp", "arret", "arrêt", "unsub", "unsubscribe", "desabonner",
		"désabonner", "stoppen", "parar", "baja", "peatu", "lopeta", "basta",
		"стоп", "стійп", "отписка", "відписка":
		return "stop"
	case "start", "unstop", "oui", "ja", "yes", "si", "sí", "jah", "start sms",
		"старт", "почати":
		return "start"
	default:
		return "other"
	}
}

// DeliveryFailed indique un statut DLR terminal en échec.
func DeliveryFailed(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "sending_failed", "delivery_failed", "expired", "rejected":
		return true
	}
	return false
}
