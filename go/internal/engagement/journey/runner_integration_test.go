package journey_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/engagement/journey"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type mockMailer struct {
	mu    sync.Mutex
	byTo  map[string][]string // email → step keys
}

func (m *mockMailer) SendJourneyStep(to, _, _, stepKey, _, _ string, _ map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byTo == nil {
		m.byTo = map[string][]string{}
	}
	m.byTo[to] = append(m.byTo[to], stepKey)
	return nil
}

func (m *mockMailer) countTo(to, step string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, s := range m.byTo[to] {
		if s == step {
			n++
		}
	}
	return n
}

func loadDotEnv() {
	dir, _ := os.Getwd()
	for i := 0; i < 8; i++ {
		envPath := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				k, v := parts[0], parts[1]
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func TestRunnerEnrollSendD0Idempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadDotEnv()
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	st := store.New(pool)
	suffix := time.Now().UnixNano()
	vetID, err := st.CreateVetAsAdmin(ctx, store.EncodeVetInput{
		Email:        fmt.Sprintf("journey-vet+%d@petsfollow.test", suffix),
		Password:     "VetDemo123!",
		FullName:     "Journey Vet",
		PracticeName: "Journey Clinic",
		Phone:        "0100000000",
		City:         "Liège",
		PostalCode:   "4000",
		AddressLine1: "1 rue Test",
	}, "")
	if err != nil {
		t.Fatalf("create vet: %v", err)
	}
	clientEmail := fmt.Sprintf("journey-client+%d@petsfollow.test", suffix)
	clientID, err := st.CreateClientForVet(ctx, vetID, store.CreateClientInput{
		Email:    clientEmail,
		Password: "ClientDemo123!",
		FullName: "Journey Client",
		Locale:   "fr",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id = $1 OR id = $2`, clientID, vetID)
	})

	// Ensure d0 is due (anchor slightly in the past).
	if _, err := pool.Exec(ctx, `
		UPDATE discovery.email_journey SET anchor_at = NOW() - INTERVAL '1 hour' WHERE user_id = $1`, clientID); err != nil {
		t.Fatalf("shift anchor: %v", err)
	}

	mailer := &mockMailer{}
	tokens := authx.NewTokenIssuer(cfg.JWTSigningKey, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	r := journey.NewRunner(st, mailer, tokens, journey.Config{
		AppDownloadURL: "https://example.com/app",
		APIPublicURL:   "http://localhost:8291",
		Enabled:        true,
		Interval:       time.Hour,
	})

	r.RunOnce(ctx)
	if mailer.countTo(clientEmail, "d0_welcome") != 1 {
		t.Fatalf("expected 1 d0_welcome for %s, got %d (%v)", clientEmail, mailer.countTo(clientEmail, "d0_welcome"), mailer.byTo[clientEmail])
	}
	var status string
	if err := pool.QueryRow(ctx, `
		SELECT status FROM discovery.email_sends WHERE user_id = $1 AND step_key = 'd0_welcome'`, clientID,
	).Scan(&status); err != nil {
		t.Fatalf("email_sends row: %v", err)
	}
	if status != "sent" {
		t.Fatalf("expected status sent, got %q", status)
	}

	r.RunOnce(ctx)
	if mailer.countTo(clientEmail, "d0_welcome") != 1 {
		t.Fatalf("expected idempotent single send, got %d", mailer.countTo(clientEmail, "d0_welcome"))
	}
}
