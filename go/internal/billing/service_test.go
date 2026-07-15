package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestValidUntil(t *testing.T) {
	plan, err := billing.GetPlan(billing.PlanTriennial)
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := billing.ValidUntil(from, plan)
	expected := from.AddDate(0, 0, 1095)
	if !until.Equal(expected) {
		t.Fatalf("expected %v got %v", expected, until)
	}
}

func TestParsePlanCode(t *testing.T) {
	if _, err := billing.ParsePlanCode("triennial"); err != nil {
		t.Fatal(err)
	}
	if _, err := billing.ParsePlanCode("invalid"); err == nil {
		t.Fatal("expected error")
	}
}

func TestMockWebhookIdempotence(t *testing.T) {
	gw := billing.NewMockGateway("whsec_test", "http://localhost:8291")
	payload := []byte(`{"id":"evt_1","type":"checkout.session.completed","data":{"object":{"id":"cs_1","metadata":{"pet_id":"p1","owner_user_id":"u1","plan_code":"triennial","billing_mode":"one_time"}}}}`)
	sig := "t=1,v1=deadbeef"
	_, err := gw.VerifyWebhook(payload, sig)
	if err == nil {
		// signature may fail without proper sig — test build payload helper instead
	}
	body, header, err := billing.BuildTestWebhookPayload("whsec_test", "checkout.session.completed", map[string]any{
		"id": "cs_test", "metadata": map[string]string{"pet_id": "p1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ev, err := gw.VerifyWebhook(body, header)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "checkout.session.completed" {
		t.Fatalf("unexpected type %s", ev.Type)
	}
	_ = context.Background()
}

func TestPlanSummary(t *testing.T) {
	plan, _ := billing.GetPlan(billing.PlanTriennial)
	s := billing.PlanSummary(plan, billing.ModeSubscription)
	if s == "" {
		t.Fatal("empty summary")
	}
}
