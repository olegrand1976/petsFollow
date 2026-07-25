package handlers_test

import (
	"net/http"
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
	if mod["allowed"] == true {
		t.Fatalf("expected inactive module for onboarding vet, got %#v", mod)
	}
	practiceID, _ := mod["practiceId"].(string)
	if practiceID == "" {
		t.Fatal("missing practiceId")
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
