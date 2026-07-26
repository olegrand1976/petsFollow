package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestSMTPConfirmFailCreatesSystemAlertTicket(t *testing.T) {
	api := newTestAPI(t)
	// Non-dev SMTP (user set + non-localhost host) so confirm mail errors propagate.
	api.api.TestReplaceNotifier(email.NewNotifierAuth(
		"smtp.invalid.petsfollow", 587,
		"petsFollow <noreply@petsfollow.app>",
		"noreply@petsfollow.app", "bad-pass",
		"http://localhost:3002", "https://ll-it-sc.be",
	))
	api.api.TestSetOpsNotifyEmail("ops@petsfollow.test")

	emailAddr := uniqueEmail("alert-smtp")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": emailAddr, "password": "ClientPass123!", "fullName": "Alert Client",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}

	st := store.New(api.pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tickets, total, err := st.ListSupportTickets(ctx, store.SupportStatusOpen, "ALERT/URGENT", 20, 0)
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if total == 0 {
		t.Fatal("expected system support ticket for SMTP confirm fail")
	}
	found := false
	for _, tk := range tickets {
		if tk.Source == store.SupportSourceSystem {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected source=system ticket, got %#v", tickets)
	}
}
