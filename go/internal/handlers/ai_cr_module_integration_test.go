package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAiCrModuleGateAndActivate(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetTok := loginToken(t, api.handler, "vet.onboarding@petsfollow.test", "VetDemo123!")

	// onboarding practice should not be seed-activated (incomplete profile).
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me ai-module %d %#v", code, env)
	}
	mod := dataMap(t, env)
	practiceID, _ := mod["practiceId"].(string)
	if practiceID == "" {
		t.Fatal("missing practiceId")
	}
	// Reset pollution from other AI CR tests sharing the DB.
	if mod["status"] != "none" && mod["status"] != nil {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+practiceID, adminTok, map[string]any{
			"status": "disabled",
		})
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
		if code != http.StatusOK {
			t.Fatalf("me after disable %d %#v", code, env)
		}
		mod = dataMap(t, env)
	}
	if mod["allowed"] == true {
		t.Fatalf("expected inactive module for onboarding vet, got %#v", mod)
	}

	vetDemo := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetDemo, nil)
	if code != http.StatusOK {
		t.Fatalf("demo me %d %#v", code, env)
	}
	demoMod := dataMap(t, env)
	if demoMod["allowed"] != true {
		pid, _ := demoMod["practiceId"].(string)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+pid+"/activate", adminTok, nil)
		if code != http.StatusOK {
			t.Fatalf("activate demo %d %#v", code, env)
		}
	}

	// ROI locked before J60 on freshly activated practice
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/activate", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("activate onboarding %d %#v", code, env)
	}
	act := dataMap(t, env)
	if act["status"] != "trial" || act["allowed"] != true {
		t.Fatalf("unexpected activate %#v", act)
	}
	if act["roiUnlocked"] == true {
		t.Fatalf("ROI should be locked before J60 %#v", act)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module/roi", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("roi %d %#v", code, env)
	}
	if dataMap(t, env)["unlocked"] == true {
		t.Fatalf("expected ROI locked %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me after %d %#v", code, env)
	}
	if dataMap(t, env)["allowed"] != true {
		t.Fatalf("expected allowed after activate %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/convert", adminTok, map[string]any{
		"pricePlan": "annual_390",
	})
	if code != http.StatusOK {
		t.Fatalf("convert %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "active" || dataMap(t, env)["pricePlan"] != "annual_390" {
		t.Fatalf("convert payload %#v", env)
	}
}

func TestAiCrImproveRequiresModule(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.onboarding@petsfollow.test", "VetDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	// Ensure module disabled/absent: force disable if present
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if practiceID == "" {
		t.Fatal("no practice")
	}
	if dataMap(t, env)["status"] != "none" {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+practiceID, adminTok, map[string]any{
			"status": "disabled",
		})
	}

	// Need a visit on this practice — create pet via seed may be empty.
	// Use demo client pets under VetPlus and confirm gate on vet.demo after disable is harder.
	// Instead: create visit as onboarding vet if they have pets; else skip via activate then disable path on a fresh visit from vet.demo.
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d", code)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	vetDemo := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetDemo, nil)
	demoPractice, _ := dataMap(t, env)["practiceId"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+demoPractice, adminTok, map[string]any{
		"status": "disabled",
	})
	if code != http.StatusOK && code != http.StatusNotFound {
		// activate then disable
		_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+demoPractice+"/activate", adminTok, nil)
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+demoPractice, adminTok, map[string]any{
			"status": "disabled",
		})
		if code != http.StatusOK {
			t.Fatalf("disable %d %#v", code, env)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetDemo, map[string]any{
		"scheduledAt":     "2099-12-20T10:00:00Z",
		"notes":           "ai gate",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetDemo, map[string]any{"status": "cancelled"})
		_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+demoPractice+"/activate", adminTok, nil)
	})

	_, _ = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetDemo, map[string]any{
		"bodyText": "anamnese courte pour improve",
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve", vetDemo, nil)
	if code != http.StatusPaymentRequired {
		t.Fatalf("improve without module want 402 got %d %#v", code, env)
	}
	errMap, _ := env["error"].(map[string]any)
	if errMap["msgKey"] != "ai_module_required" && errMap["messageKey"] != "ai_module_required" {
		// localized write may use messageKey field name differently
		if errMap["code"] != "payment_required" {
			t.Fatalf("unexpected error %#v", env)
		}
	}
}

func TestAiCrAdhesionDripIdempotent(t *testing.T) {
	t.Setenv("AI_MODULE_FRICTION_SECRET", "test-ai-friction-secret")
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetTok := loginToken(t, api.handler, "vet.onboarding@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if practiceID == "" {
		t.Fatal("no practice")
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+practiceID, adminTok, map[string]any{
			"status": "disabled",
		})
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM practice.ai_cr_email_sends WHERE practice_id = $1`, practiceID)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/activate", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("activate %d %#v", code, env)
	}

	var hasJ0 bool
	if err := api.pool.QueryRow(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM practice.ai_cr_email_sends WHERE practice_id=$1 AND step_key='j0_activation')`,
		practiceID).Scan(&hasJ0); err != nil {
		t.Fatalf("query j0: %v", err)
	}
	if !hasJ0 {
		t.Fatal("expected j0_activation claimed on activate")
	}

	// Backdate activation to day 3 with zero usage → j3_nudge
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE practice.ai_cr_modules
		SET activated_at = NOW() - INTERVAL '3 days',
		    trial_ends_at = NOW() + INTERVAL '87 days',
		    status = 'trial',
		    updated_at = NOW()
		WHERE practice_id = $1`, practiceID); err != nil {
		t.Fatalf("backdate: %v", err)
	}

	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/ai-module-friction/run", nil, map[string]string{
		"X-Ai-Module-Friction-Secret": "test-ai-friction-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("friction/adhesion run %d %#v", code, env)
	}
	var hasJ3 bool
	if err := api.pool.QueryRow(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM practice.ai_cr_email_sends WHERE practice_id=$1 AND step_key='j3_nudge')`,
		practiceID).Scan(&hasJ3); err != nil {
		t.Fatalf("query j3: %v", err)
	}
	if !hasJ3 {
		t.Fatal("expected j3_nudge after backdate + run")
	}

	// Second run must not duplicate (idempotent)
	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/ai-module-friction/run", nil, map[string]string{
		"X-Ai-Module-Friction-Secret": "test-ai-friction-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("second run %d %#v", code, env)
	}
	var n int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*)::int FROM practice.ai_cr_email_sends WHERE practice_id=$1 AND step_key='j3_nudge'`,
		practiceID).Scan(&n); err != nil {
		t.Fatalf("count j3: %v", err)
	}
	if n != 1 {
		t.Fatalf("j3 rows want 1 got %d", n)
	}
}

func TestAiCrCareProGateUsesVisitPractice(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/ai-module", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet me %d %#v", code, env)
	}
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if practiceID == "" {
		t.Fatal("no practice")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/care-pro/pets", farrierTok, nil)
	if code != http.StatusOK {
		t.Fatalf("care pets %d %#v", code, env)
	}
	var spiritID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] == "Spirit" {
			spiritID, _ = p["id"].(string)
			break
		}
	}
	if spiritID == "" {
		t.Skip("Spirit not granted")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+spiritID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-11-20T10:00:00Z",
		"notes":           "care_pro ai gate",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{"status": "cancelled"})
		_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/activate", adminTok, nil)
	})

	_, _ = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", farrierTok, map[string]any{
		"bodyText": "notes ferrage pour improve",
	})

	// Disable module → care_pro improve must 402
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+practiceID, adminTok, map[string]any{
		"status": "disabled",
	})
	if code != http.StatusOK && code != http.StatusNotFound {
		_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/activate", adminTok, nil)
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/ai-modules/"+practiceID, adminTok, map[string]any{
			"status": "disabled",
		})
		if code != http.StatusOK {
			t.Fatalf("disable %d %#v", code, env)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve", farrierTok, nil)
	if code != http.StatusPaymentRequired {
		t.Fatalf("care_pro improve without module want 402 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/ai-modules/"+practiceID+"/activate", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("reactivate %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/improve", farrierTok, nil)
	if code == http.StatusPaymentRequired {
		t.Fatalf("care_pro improve with module must not be 402, got %#v", env)
	}
}

func doJSONWithHeaders(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}
