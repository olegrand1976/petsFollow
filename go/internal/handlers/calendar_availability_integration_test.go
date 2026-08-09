package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
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
	// Un cabinet multi-sites a une ligne d'agenda par site : on mémorise ceux qui
	// prennent des RDV en ligne pour ne réactiver qu'eux. Relire une seule ligne
	// puis l'appliquer à toutes les remettait toutes à false — la réservation en
	// ligne du cabinet de démo restait cassée pour les suites suivantes.
	rows, err := api.pool.Query(ctx, `
		SELECT site_id::text FROM practice.vet_schedule
		WHERE practice_id = $1 AND client_booking_enabled`, practiceID)
	if err != nil {
		t.Fatalf("read schedule: %v", err)
	}
	var bookableSiteIDs []string
	for rows.Next() {
		var siteID string
		if err := rows.Scan(&siteID); err != nil {
			rows.Close()
			t.Fatalf("scan schedule: %v", err)
		}
		bookableSiteIDs = append(bookableSiteIDs, siteID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("read schedule: %v", err)
	}

	if _, err := api.pool.Exec(ctx, `
		UPDATE practice.vet_schedule SET client_booking_enabled = false, updated_at = NOW()
		WHERE practice_id = $1`, practiceID); err != nil {
		t.Fatalf("disable booking: %v", err)
	}
	t.Cleanup(func() {
		if len(bookableSiteIDs) == 0 {
			return
		}
		_, _ = api.pool.Exec(ctx, `
			UPDATE practice.vet_schedule SET client_booking_enabled = true, updated_at = NOW()
			WHERE practice_id = $1 AND site_id::text = ANY($2)`, practiceID, bookableSiteIDs)
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
