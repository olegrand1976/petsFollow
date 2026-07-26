package billing_test

import (
	"context"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestMockCustomerID(t *testing.T) {
	got := billing.MockCustomerID("user-abc")
	if got != "cus_mock_user-abc" {
		t.Fatalf("got %q", got)
	}
}

func TestMockGatewayCreatePortalSession(t *testing.T) {
	g := billing.NewMockGateway("whsec_test", "http://localhost:8291/")
	sess, err := g.CreatePortalSession(context.Background(), "cus_mock_x", "petsfollow://home")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sess.URL, "/api/v1/billing/dev/mock-portal?") {
		t.Fatalf("url=%q", sess.URL)
	}
	if !strings.Contains(sess.URL, "customer=cus_mock_x") {
		t.Fatalf("missing customer: %q", sess.URL)
	}
	if !strings.Contains(sess.URL, "return=petsfollow") {
		t.Fatalf("missing return: %q", sess.URL)
	}
}
