package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestPracticeAvailabilityIncludesPhoneWhenDisabled(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/vets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me/vets %d %#v (make seed?)", code, env)
	}
	vets, ok := env["data"].([]any)
	if !ok || len(vets) == 0 {
		t.Fatalf("expected linked vets, got %#v", env["data"])
	}
	first, _ := vets[0].(map[string]any)
	practiceID, _ := first["practiceId"].(string)
	if practiceID == "" {
		t.Fatalf("missing practiceId: %#v", first)
	}

	ctx := context.Background()
	var oldEnabled bool
	hadRow := true
	err := api.pool.QueryRow(ctx, `
		SELECT client_booking_enabled FROM practice.vet_schedule WHERE practice_id = $1`, practiceID,
	).Scan(&oldEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		hadRow = false
	} else if err != nil {
		t.Fatalf("read schedule: %v", err)
	}

	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.vet_schedule (practice_id, client_booking_enabled, slot_duration_minutes, timezone, updated_at)
		VALUES ($1, false, 30, 'Europe/Brussels', NOW())
		ON CONFLICT (practice_id) DO UPDATE SET client_booking_enabled = false, updated_at = NOW()`, practiceID)
	if err != nil {
		t.Fatalf("disable booking: %v", err)
	}
	t.Cleanup(func() {
		if !hadRow {
			_, _ = api.pool.Exec(ctx, `DELETE FROM practice.vet_schedule WHERE practice_id = $1`, practiceID)
			return
		}
		_, _ = api.pool.Exec(ctx, `
			UPDATE practice.vet_schedule SET client_booking_enabled = $2, updated_at = NOW()
			WHERE practice_id = $1`, practiceID, oldEnabled)
	})

	from := time.Now().UTC().Format(time.RFC3339)
	to := time.Now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339)
	path := fmt.Sprintf("/api/v1/practices/%s/availability?from=%s&to=%s", practiceID, from, to)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, path, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("availability %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["enabled"] != false {
		t.Fatalf("expected enabled=false, got %#v", data["enabled"])
	}
	phone, _ := data["practicePhone"].(string)
	if phone == "" {
		t.Fatalf("expected practicePhone, got %#v", data)
	}
	name, _ := data["practiceName"].(string)
	if name == "" {
		t.Fatalf("expected practiceName, got %#v", data)
	}
}

func TestPracticeAvailabilityNotLinkedForbidden(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.vide@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/practices/00000000-0000-0000-0000-000000000099/availability", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d %#v", code, env)
	}
}
