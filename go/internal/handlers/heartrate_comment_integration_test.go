package handlers_test

import (
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func activeDemoPetID(t *testing.T, h http.Handler, clientTok string) string {
	t.Helper()
	// Always provision a fresh entitled pet so tests don't depend on seed entitlement freshness.
	code, env := doAuthJSON(t, h, http.MethodPost, "/api/v1/pets", clientTok, map[string]any{
		"name": "HRTestPet", "species": "dog", "breed": "test",
		"plan": "triennial", "billingMode": "subscription",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create pet %d %#v (make seed?)", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	if pet == nil {
		pet = data
	}
	petID, _ := pet["id"].(string)
	ownerID, _ := pet["ownerUserId"].(string)
	if ownerID == "" {
		meCode, meEnv := doAuthJSON(t, h, http.MethodGet, "/api/v1/me", clientTok, nil)
		if meCode != http.StatusOK {
			t.Fatalf("me %d %#v", meCode, meEnv)
		}
		ownerID, _ = dataMap(t, meEnv)["userId"].(string)
	}
	if petID == "" || ownerID == "" {
		t.Fatalf("missing pet/owner: %#v", env)
	}
	code, env = doAuthJSON(t, h, http.MethodGet,
		"/api/v1/billing/dev/mock-complete?pet_id="+petID+"&owner_user_id="+ownerID+"&plan_code=triennial&billing_mode=subscription",
		clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("mock-complete %d %#v", code, env)
	}
	return petID
}

func startCompleteSession(t *testing.T, h http.Handler, clientTok, petID string, taps int) string {
	t.Helper()
	code, env := doAuthJSON(t, h, http.MethodPost, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusCreated {
		t.Fatalf("start %d %#v", code, env)
	}
	sessID, _ := dataMap(t, env)["id"].(string)
	if sessID == "" {
		t.Fatalf("missing session id: %#v", env)
	}
	code, env = doAuthJSON(t, h, http.MethodPatch, "/api/v1/heartrate/sessions/"+sessID, clientTok, map[string]any{
		"tapCount": taps,
	})
	if code != http.StatusOK {
		t.Fatalf("complete %d %#v", code, env)
	}
	return sessID
}

func TestHeartRateValidateCommentPersisted(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	sessID := startCompleteSession(t, api.handler, clientTok, petID, 60)

	const comment = "agité ce matin"
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/heartrate/sessions/"+sessID+"/validate", clientTok, map[string]any{
		"comment": comment,
	})
	if code != http.StatusOK {
		t.Fatalf("validate %d %#v", code, env)
	}
	got, _ := dataMap(t, env)["comment"].(string)
	if got != comment {
		t.Fatalf("validate comment=%q want %q", got, comment)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/heartrate/sessions", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list client %d %#v", code, env)
	}
	if !sessionHasComment(env["data"], sessID, comment) {
		t.Fatalf("client list missing comment: %#v", env["data"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/heartrate/sessions", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list vet %d %#v", code, env)
	}
	if !sessionHasComment(env["data"], sessID, comment) {
		t.Fatalf("vet list missing comment: %#v", env["data"])
	}
}

func TestHeartRateValidateBlankComment(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	sessID := startCompleteSession(t, api.handler, clientTok, petID, 45)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/heartrate/sessions/"+sessID+"/validate", clientTok, map[string]any{
		"comment": "   ",
	})
	if code != http.StatusOK {
		t.Fatalf("validate blank %d %#v", code, env)
	}
	if _, has := dataMap(t, env)["comment"]; has {
		t.Fatalf("expected omitted/nil comment, got %#v", dataMap(t, env)["comment"])
	}
}

func TestHeartRateValidateCommentTruncated(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	sessID := startCompleteSession(t, api.handler, clientTok, petID, 50)

	long := strings.Repeat("é", store.MaxHeartRateCommentLen+25)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/heartrate/sessions/"+sessID+"/validate", clientTok, map[string]any{
		"comment": long,
	})
	if code != http.StatusOK {
		t.Fatalf("validate long %d %#v", code, env)
	}
	got, _ := dataMap(t, env)["comment"].(string)
	if n := utf8.RuneCountInString(got); n != store.MaxHeartRateCommentLen {
		t.Fatalf("rune count=%d want %d", n, store.MaxHeartRateCommentLen)
	}
}

func sessionHasComment(data any, sessID, want string) bool {
	rows, ok := data.([]any)
	if !ok {
		return false
	}
	for _, row := range rows {
		m, _ := row.(map[string]any)
		if m["id"] == sessID {
			c, _ := m["comment"].(string)
			return c == want
		}
	}
	return false
}
