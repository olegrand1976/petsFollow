package handlers_test

import (
	"net/http"
	"testing"
)

func completeHRSession(t *testing.T, h http.Handler, clientTok, sessID string, taps int) map[string]any {
	t.Helper()
	code, env := doAuthJSON(t, h, http.MethodPatch, "/api/v1/heartrate/sessions/"+sessID, clientTok, map[string]any{
		"tapCount": taps,
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	return dataMap(t, env)
}

func validateHRSession(t *testing.T, h http.Handler, clientTok, sessID string) {
	t.Helper()
	code, env := doAuthJSON(t, h, http.MethodPost, "/api/v1/heartrate/sessions/"+sessID+"/validate", clientTok, map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("validate %d %#v", code, env)
	}
}

func TestHeartRateDeltaAlertVsPreviousValidated(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	// First reading: 60 taps / 60s = 60 bpm — no previous → no alert.
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("start1 %d %#v", code, env)
	}
	sess1, _ := dataMap(t, env)["id"].(string)
	m1 := completeHRSession(t, api.handler, clientTok, sess1, 60)
	if alert, _ := m1["isAlert"].(bool); alert {
		t.Fatalf("first reading should not alert: %#v", m1)
	}
	validateHRSession(t, api.handler, clientTok, sess1)

	// Rise +29 → no alert (delta seed = 30).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("start2 %d %#v", code, env)
	}
	sess2, _ := dataMap(t, env)["id"].(string)
	m2 := completeHRSession(t, api.handler, clientTok, sess2, 89) // 89 bpm
	if alert, _ := m2["isAlert"].(bool); alert {
		t.Fatalf("+29 should not alert: %#v", m2)
	}
	validateHRSession(t, api.handler, clientTok, sess2)

	// Rise +30 from last validated (89) → 119 bpm alerts.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("start3 %d %#v", code, env)
	}
	sess3, _ := dataMap(t, env)["id"].(string)
	m3 := completeHRSession(t, api.handler, clientTok, sess3, 119)
	if alert, _ := m3["isAlert"].(bool); !alert {
		t.Fatalf("+30 from 89 should alert: %#v", m3)
	}
}

func TestHeartRateStartRejectsOtherSpecies(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", clientTok, map[string]any{
		"name": "OtherPet", "species": "other", "breed": "misc",
		"plan": "triennial", "billingMode": "subscription",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create pet %d %#v", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	if pet == nil {
		pet = data
	}
	petID, _ := pet["id"].(string)
	ownerID, _ := pet["ownerUserId"].(string)
	if ownerID == "" {
		meCode, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", clientTok, nil)
		if meCode != http.StatusOK {
			t.Fatalf("me %d %#v", meCode, meEnv)
		}
		ownerID, _ = dataMap(t, meEnv)["userId"].(string)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/billing/dev/mock-complete?pet_id="+petID+"&owner_user_id="+ownerID+"&plan_code=triennial&billing_mode=subscription",
		clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("mock-complete %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("start other want 403 got %d %#v", code, env)
	}
	if msg := errorMsgKey(env); msg != "heartrate_not_supported" {
		t.Fatalf("msgKey=%q want heartrate_not_supported %#v", msg, env)
	}
}
