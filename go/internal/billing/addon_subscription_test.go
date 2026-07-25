package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestAddonDurationIsLifetime(t *testing.T) {
	addon, err := billing.GetAddon(billing.AddonFamily)
	if err != nil {
		t.Fatal(err)
	}
	if addon.DurationDays != 0 {
		t.Fatalf("addon duration want 0 (lifetime), got %d", addon.DurationDays)
	}
}

func TestLegacyAddonValidUntilIsOneYear(t *testing.T) {
	addon, err := billing.GetAddon(billing.AddonFamily)
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := billing.AddonValidUntil(from, addon)
	if until.Sub(from) != 365*24*time.Hour {
		t.Fatalf("legacy renew want 365d, got %v", until.Sub(from))
	}
}

func TestMockGatewayCancelSubscriptionNoop(t *testing.T) {
	gw := billing.NewMockGateway("whsec_test", "http://localhost:8291")
	if err := gw.CancelSubscription(context.Background(), "sub_mock_x"); err != nil {
		t.Fatal(err)
	}
}

func TestMockAddonCheckoutPayloadIsOneTime(t *testing.T) {
	body, header, err := billing.BuildTestWebhookPayload("whsec_test", "checkout.session.completed", map[string]any{
		"id":             "cs_addon",
		"subscription":   nil,
		"payment_intent": "pi_mock_addon_1",
		"metadata": map[string]any{
			"kind":     "addon",
			"addon_id": "a1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	gw := billing.NewMockGateway("whsec_test", "http://localhost:8291")
	ev, err := gw.VerifyWebhook(body, header)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "checkout.session.completed" {
		t.Fatalf("type = %s", ev.Type)
	}
}
