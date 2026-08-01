package handlers_test

import (
	"net/http"
	"testing"
)

func TestVetPetLifecycleDates(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get pet %d %#v", code, env)
	}
	before := dataMap(t, env)

	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/lifecycle", vetTok, map[string]any{
			"adoptedAt":  "",
			"soldAt":     "",
			"deceasedAt": "",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/lifecycle", vetTok, map[string]any{
		"adoptedAt":  "2020-03-15",
		"soldAt":     "",
		"deceasedAt": "2025-01-10",
	})
	if code != http.StatusOK {
		t.Fatalf("lifecycle patch %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["adoptedAt"] != "2020-03-15" {
		t.Fatalf("adoptedAt %#v", got["adoptedAt"])
	}
	if got["soldAt"] != nil {
		t.Fatalf("soldAt want nil got %#v", got["soldAt"])
	}
	if got["deceasedAt"] != "2025-01-10" {
		t.Fatalf("deceasedAt %#v", got["deceasedAt"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get after patch %d %#v", code, env)
	}
	pet := dataMap(t, env)
	adopted := stringifyDate(pet["adoptedAt"])
	deceased := stringifyDate(pet["deceasedAt"])
	if adopted != "2020-03-15" {
		t.Fatalf("pet adoptedAt=%q want 2020-03-15 (before name=%v)", adopted, before["name"])
	}
	if deceased != "2025-01-10" {
		t.Fatalf("pet deceasedAt=%q want 2025-01-10", deceased)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/lifecycle", vetTok, map[string]any{
		"adoptedAt": "not-a-date",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid date want 400 got %d %#v", code, env)
	}
}

func stringifyDate(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
