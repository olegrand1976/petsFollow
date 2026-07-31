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
func ParseWebhook(body []byte) (orderID, eventType string, status invoicing.DocStatus, peppolStatus string, err error) {
	var p WebhookPayload
	if err = json.Unmarshal(body, &p); err != nil {
		return "", "", "", "", fmt.Errorf("invalid webhook json: %w", err)
	}
	eventType = firstNonEmpty(p.EventType, p.Event, p.Type)
	orderID = firstNonEmpty(p.ExternalID, anyString(p.OrderID), anyString(p.OrderId))
	if orderID == "" {
		return "", eventType, "", "", fmt.Errorf("webhook missing order id")
	}
	raw := strings.ToLower(firstNonEmpty(p.DeliveryStatus, p.PeppolStatus, p.Status, eventType))
	status, peppolStatus = mapWebhookStatus(raw)
	return orderID, eventType, status, peppolStatus, nil
}

// WebhookDedupeKey returns a stable idempotency key for webhook ingestion.
// Prefer Billit's EventID when present. Do NOT use generic "Id" (often the order id) —
// that would collapse sending→delivered into a false duplicate. Fallback: sha256(body).
func WebhookDedupeKey(body []byte) string {
	var p WebhookPayload
	if err := json.Unmarshal(body, &p); err == nil {
		if id := firstNonEmpty(p.EventID, p.EventId); id != "" {
			return "evt:" + id
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
	switch raw {
	case "delivered", "orderdelivered", "peppoldelivered", "deliverydelivered":
		return invoicing.StatusDelivered, "delivered"
	case "rejected", "orderrejected", "peppolrejected", "failed", "fail", "error", "deliveryfailed":
		return invoicing.StatusRejected, "rejected"
	case "cancelled", "canceled", "ordercancelled", "ordercanceled":
		return invoicing.StatusCancelled, "cancelled"
	case "sending", "sent", "pending", "processing", "inprogress", "queued",
		"ordersent", "peppolsending", "peppolpending":
		return invoicing.StatusSending, "sending"
	default:
		// Unknown codes: do not invent a confirmation — keep awaiting network status.
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
