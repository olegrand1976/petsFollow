package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestPublicPreconsultTokenXSSAndFlow(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":       "2099-11-15T10:00:00Z",
		"notes":             "preconsult public probe",
		"durationMinutes":   30,
		"confirmDirect":     true,
		"requestPreconsult": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	if visitID == "" {
		t.Fatalf("no visit id %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	var token string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		err := api.pool.QueryRow(context.Background(), `
			SELECT token FROM visits.preconsult_tokens WHERE visit_id = $1 ORDER BY created_at DESC LIMIT 1`, visitID,
		).Scan(&token)
		if err == nil && token != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if token == "" {
		t.Fatal("expected preconsult token after opt-in confirm")
	}

	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/preconsult/"+token, nil)
	if code != http.StatusOK {
		t.Fatalf("public get %d %#v", code, env)
	}
	pub := dataMap(t, env)
	if pub["status"] != "pending" {
		t.Fatalf("status=%v", pub["status"])
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/public/preconsult/"+token, map[string]any{
		"answers": map[string]any{
			"chiefComplaint": "<script>x</script>",
			"duration":       "today",
			"behavior":       "normal",
			"appetite":       "normal",
			"thirst":         "normal",
			"elimination":    "normal",
			"urgency":        "low",
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("xss reject want 400 got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/public/preconsult/"+token, map[string]any{
		"answers": map[string]any{
			"chiefComplaint": "Vomissements depuis hier",
			"duration":       "few_days",
			"behavior":       "lethargic",
			"appetite":       "decreased",
			"thirst":         "normal",
			"elimination":    "normal",
			"urgency":        "medium",
			"comment":        "plain text only",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("submit %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/preconsult", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet read %d %#v", code, env)
	}
	in := dataMap(t, env)
	if in["status"] != "submitted" {
		t.Fatalf("intake status=%v", in["status"])
	}

	code, _ = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/preconsult/not-a-real-token", nil)
	if code != http.StatusNotFound {
		t.Fatalf("invalid token want 404 got %d", code)
	}
}

func TestPublicPreconsultHighUrgencyNotifiesVet(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":       "2099-11-17T10:00:00Z",
		"notes":             "preconsult urgent probe",
		"durationMinutes":   30,
		"confirmDirect":     true,
		"requestPreconsult": true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	if visitID == "" {
		t.Fatalf("no visit id %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	var token string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		err := api.pool.QueryRow(context.Background(), `
			SELECT token FROM visits.preconsult_tokens WHERE visit_id = $1 ORDER BY created_at DESC LIMIT 1`, visitID,
		).Scan(&token)
		if err == nil && token != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if token == "" {
		t.Fatal("expected preconsult token")
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/public/preconsult/"+token, map[string]any{
		"answers": map[string]any{
			"chiefComplaint": "Détresse respiratoire",
			"duration":       "today",
			"behavior":       "lethargic",
			"appetite":       "decreased",
			"thirst":         "normal",
			"elimination":    "normal",
			"urgency":        "high",
			"comment":        "urgent declared",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("submit %d %#v", code, env)
	}

	var logCount int
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		_ = api.pool.QueryRow(context.Background(), `
			SELECT COUNT(*)::int FROM notifications.notification_log
			WHERE kind = 'preconsult_urgent' AND payload->>'visitId' = $1`, visitID,
		).Scan(&logCount)
		if logCount > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if logCount == 0 {
		t.Fatal("expected preconsult_urgent notification_log")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/calendar?from=2099-11-01&to=2099-11-30", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("calendar %d %#v", code, env)
	}
	cal := dataMap(t, env)
	visits, _ := cal["visits"].([]any)
	foundAlert := false
	for _, raw := range visits {
		v, _ := raw.(map[string]any)
		if v["id"] == visitID {
			if v["preconsultAlert"] == "urgent" {
				foundAlert = true
			}
			break
		}
	}
	if !foundAlert {
		t.Fatalf("expected preconsultAlert=urgent on calendar visit, got %#v", cal)
	}
}

func TestConfirmWithoutPreconsultDoesNotIssueToken(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	petID, _ := env["data"].([]any)[0].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":       "2099-11-16T11:00:00Z",
		"notes":             "no preconsult",
		"durationMinutes":   30,
		"confirmDirect":     true,
		"requestPreconsult": false,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	time.Sleep(200 * time.Millisecond)
	var n int
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM visits.preconsult_tokens WHERE visit_id = $1`, visitID).Scan(&n)
	if n != 0 {
		t.Fatalf("expected no token without opt-in, got %d", n)
	}
}
