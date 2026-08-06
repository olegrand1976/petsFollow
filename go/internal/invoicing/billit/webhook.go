package billit

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

// VerifyWebhookSignature checks HMAC-SHA256 hex digest of the raw body.
// Header value may be "sha256=<hex>" or bare hex. Empty secret refuses all.
func VerifyWebhookSignature(secret string, body []byte, header string) bool {
	secret = strings.TrimSpace(secret)
	header = strings.TrimSpace(header)
	if secret == "" || header == "" {
		return false
	}
	header = strings.TrimPrefix(header, "sha256=")
	want, err := hex.DecodeString(header)
	if err != nil || len(want) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	got := mac.Sum(nil)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// ParseWebhook extracts order id + mapped statuses from a Billit callback body.
// Accepts flat Order payloads and Access Point Message payloads
// (EntityDetail.OrderMessage.OrderID + EInvoiceFlowState).
func ParseWebhook(body []byte) (orderID, eventType string, status invoicing.DocStatus, peppolStatus string, err error) {
	var p WebhookPayload
	if err = json.Unmarshal(body, &p); err != nil {
		return "", "", "", "", fmt.Errorf("invalid webhook json: %w", err)
	}
	eventType = firstNonEmpty(
		p.EventType,
		p.Event,
		p.Type,
		p.UpdatedEntityType,
		flowStateFromPayload(&p),
	)
	orderID = orderIDFromPayload(&p)
	if orderID == "" {
		return "", eventType, "", "", fmt.Errorf("webhook missing order id")
	}
	raw := strings.ToLower(firstNonEmpty(
		p.DeliveryStatus,
		p.PeppolStatus,
		p.Status,
		flowStateFromPayload(&p),
	))
	// Explicit network refusal even when Success was true on an earlier ping.
	if refusedFromPayload(&p) {
		raw = "refused"
	}
	// Do not fall back to UpdatedEntityType ("Message"/"Order") as a status code.
	if raw == "" {
		raw = strings.ToLower(firstNonEmpty(p.EventType, p.Event, p.Type))
	}
	status, peppolStatus = mapWebhookStatus(raw)
	return orderID, eventType, status, peppolStatus, nil
}

func orderIDFromPayload(p *WebhookPayload) string {
	if id := firstNonEmpty(p.ExternalID, anyString(p.OrderID), anyString(p.OrderId)); id != "" {
		return id
	}
	if p.EntityDetail == nil {
		return ""
	}
	if p.EntityDetail.OrderMessage != nil {
		if id := anyString(p.EntityDetail.OrderMessage.OrderID); id != "" {
			return id
		}
	}
	return anyString(p.EntityDetail.OrderID)
}

func flowStateFromPayload(p *WebhookPayload) string {
	if p.EntityDetail == nil {
		return ""
	}
	if p.EntityDetail.AdditionalMessageInformation != nil {
		if s := strings.TrimSpace(p.EntityDetail.AdditionalMessageInformation.EInvoiceFlowState); s != "" {
			return s
		}
	}
	if p.EntityDetail.MessageAdditionalInformation != nil {
		if s := strings.TrimSpace(p.EntityDetail.MessageAdditionalInformation.EInvoiceFlowState); s != "" {
			return s
		}
	}
	return ""
}

func refusedFromPayload(p *WebhookPayload) bool {
	state := strings.ToLower(strings.TrimSpace(flowStateFromPayload(p)))
	return state == "refused" || state == "rejected"
}

// WebhookDedupeKey returns a stable idempotency key for webhook ingestion.
// Prefer Billit's EventID when present. Do NOT use generic "Id" / UpdatedEntityID
// alone (Message I stays stable across U status upgrades) — that would collapse
// sending→delivered into a false duplicate. Fallback: sha256(body).
func WebhookDedupeKey(body []byte) string {
	var p WebhookPayload
	if err := json.Unmarshal(body, &p); err == nil {
		if id := firstNonEmpty(p.EventID, p.EventId); id != "" {
			return "evt:" + id
		}
		// Message progression: same UpdatedEntityID, different flow state / update type.
		if entity := anyString(p.UpdatedEntityID); entity != "" {
			flow := flowStateFromPayload(&p)
			upd := strings.TrimSpace(p.WebhookUpdateTypeTC)
			if flow != "" || upd != "" {
				return "msg:" + entity + ":" + upd + ":" + strings.ToLower(flow)
			}
		}
	}
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func mapWebhookStatus(raw string) (invoicing.DocStatus, string) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "_", "")
	raw = strings.ReplaceAll(raw, "-", "")

	// Strict whitelist — avoid Contains("accept")/("success") false positives.
	// Billit EInvoiceFlowState: Sent → sending ; Accepted = network ack (not final delivered) ;
	// Delivered = final success ; Refused = rejected.
	switch raw {
	case "delivered", "orderdelivered", "peppoldelivered", "deliverydelivered":
		return invoicing.StatusDelivered, "delivered"
	case "rejected", "orderrejected", "peppolrejected", "failed", "fail", "error", "deliveryfailed",
		"refused":
		return invoicing.StatusRejected, "rejected"
	case "cancelled", "canceled", "ordercancelled", "ordercanceled":
		return invoicing.StatusCancelled, "cancelled"
	case "sending", "sent", "pending", "processing", "inprogress", "queued",
		"ordersent", "peppolsending", "peppolpending",
		"accepted", "orderaccepted":
		return invoicing.StatusSending, "sending"
	default:
		// Unknown codes (incl. empty): do not invent a confirmation.
		return invoicing.StatusSending, "unknown"
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func anyString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
