package handlers_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// Isolated vet + client so parallel suites never mutate vet.demo durations.
func provisionHRPractice(t *testing.T, h http.Handler) (vetTok, clientTok string) {
	t.Helper()
	vetEmail := uniqueEmail("hr-dur-vet")
	clientEmail := uniqueEmail("hr-dur-client")
	const password = "TestPass123!"

	code, env := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": vetEmail, "password": password, "fullName": "Dr HR Dur", "practiceName": "Cabinet HR Dur",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register vet %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, h, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm vet %d %#v", code, env)
	}
	vetTok, _ = dataMap(t, env)["accessToken"].(string)
	if vetTok == "" {
		t.Fatal("missing vet accessToken")
	}

	code, env = doAuthJSON(t, h, http.MethodPut, "/api/v1/vet/profile", vetTok, map[string]any{
		"vetFullName":           "Dr HR Dur",
		"practiceName":          "Cabinet HR Dur",
		"contactEmail":          vetEmail,
		"phone":                 "+32123456789",
		"addressLine1":          "Rue Test 1",
		"city":                  "Bruxelles",
		"postalCode":            "1000",
		"heartrateDurationsSec": []int{15, 30},
	})
	if code != http.StatusOK {
		t.Fatalf("put profile %d %#v", code, env)
	}

	code, env = doAuthJSON(t, h, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": clientEmail, "password": password, "fullName": "Client HR Dur",
	})
	if code != http.StatusCreated {
		t.Fatalf("create client %d %#v", code, env)
	}

	clientTok = loginToken(t, h, clientEmail, password)
	return vetTok, clientTok
}

func TestListPetsIncludesPracticeHeartRateDurationsAndSessionBPM(t *testing.T) {
	api := newTestAPI(t)
	_, clientTok := provisionHRPractice(t, api.handler)

	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, ok := env["data"].([]any)
	if !ok || len(pets) == 0 {
		t.Fatalf("expected pets list, got %#v", env["data"])
	}
	var found bool
	for _, row := range pets {
		m, _ := row.(map[string]any)
		if m["id"] != petID {
			continue
		}
		found = true
		got := durationInts(t, m["heartrateDurationsSec"])
		if !reflect.DeepEqual(got, []int{15, 30}) {
			t.Fatalf("pet %s heartrateDurationsSec=%#v want [15 30]", petID, got)
		}
	}
	if !found {
		t.Fatalf("created pet %s missing from GET /pets", petID)
	}

	const taps = 15
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, map[string]any{
		"durationSec": 15,
	})
	if code != http.StatusCreated {
		t.Fatalf("start 15s %d %#v", code, env)
	}
	sess := dataMap(t, env)
	sessID, _ := sess["id"].(string)
	if sessID == "" {
		t.Fatalf("missing session id: %#v", sess)
	}
	if dur, _ := sess["durationSec"].(float64); int(dur) != 15 {
		t.Fatalf("session durationSec=%v want 15", sess["durationSec"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/heartrate/sessions/"+sessID, clientTok, map[string]any{
		"tapCount": taps,
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	completed := dataMap(t, env)
	if bpm, _ := completed["bpm"].(float64); int(bpm) != 60 {
		t.Fatalf("bpm=%v want 60 (taps=%d duration=15)", completed["bpm"], taps)
	}
	if tapCount, _ := completed["tapCount"].(float64); int(tapCount) != taps {
		t.Fatalf("tapCount=%v want %d", completed["tapCount"], taps)
	}
}
