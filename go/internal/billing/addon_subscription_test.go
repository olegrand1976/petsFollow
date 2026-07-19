package billing_test

import (
	"context"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestAddonValidUntilIsOneYear(t *testing.T) {
	addon, err := billing.GetAddon(billing.AddonFamily)
	if err != nil {
		t.Fatal(err)
	}
	if addon.DurationDays != 365 {
		t.Fatalf("addon duration want 365, got %d", addon.DurationDays)
	}
}

func TestMockGatewayCancelSubscriptionNoop(t *testing.T) {
	gw := billing.NewMockGateway("whsec_test", "http://localhost:8291")
	if err := gw.CancelSubscription(context.Background(), "sub_mock_x"); err != nil {
		t.Fatal(err)
	}
}

func TestMockAddonCheckoutPayloadHasSubscription(t *testing.T) {
	body, header, err := billing.BuildTestWebhookPayload("whsec_test", "checkout.session.completed", map[string]any{
		"id":           "cs_addon",
		"subscription": "sub_mock_addon_1",
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
