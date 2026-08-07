package handlers_test

import (
	"net/http"
	"testing"
)

func TestTeamIncludeInCalendar(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/team", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list team: %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	members, _ := data["members"].([]any)
	if len(members) < 1 {
		t.Fatal("expected at least one team member")
	}

	memberID := ""
	userID := ""
	for _, m := range members {
		mm := m.(map[string]any)
		if mm["includeInCalendar"] != true {
			t.Fatalf("default includeInCalendar should be true, got %#v for %v", mm["includeInCalendar"], mm["id"])
		}
		if memberID == "" {
			memberID, _ = mm["id"].(string)
			userID, _ = mm["userId"].(string)
		}
		// Prefer a non-reference member when available (reference also allowed for this field).
		if mm["teamRole"] != "reference_vet" {
			memberID, _ = mm["id"].(string)
			userID, _ = mm["userId"].(string)
			break
		}
	}
	if memberID == "" || userID == "" {
		t.Fatal("no member id")
	}

	restore := true
	t.Cleanup(func() {
		if !restore {
			return
		}
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/team/"+memberID, vetTok, map[string]any{
			"includeInCalendar": true,
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/team/"+memberID, vetTok, map[string]any{
		"includeInCalendar": false,
	})
	if code != http.StatusOK {
		t.Fatalf("patch includeInCalendar false: %d %#v", code, env)
	}
	patched, _ := env["data"].(map[string]any)
	if patched["includeInCalendar"] != false {
		t.Fatalf("expected includeInCalendar=false, got %#v", patched["includeInCalendar"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/team", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list team after patch: %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	members, _ = data["members"].([]any)
	found := false
	for _, m := range members {
		mm := m.(map[string]any)
		if mm["id"] == memberID {
			found = true
			if mm["includeInCalendar"] != false {
				t.Fatalf("list: expected includeInCalendar=false for %s, got %#v", memberID, mm["includeInCalendar"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("member %s not found after includeInCalendar patch", memberID)
	}

	// Hidden member cannot be newly assigned via API.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets: %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	petID := ""
	for _, p := range pets {
		pm, _ := p.(map[string]any)
		if pm["isWalkinPlaceholder"] == true {
			continue
		}
		petID, _ = pm["id"].(string)
		if petID != "" {
			break
		}
	}
	if petID == "" {
		t.Fatal("no non-walkin pet for assignee check")
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2030-06-15T10:00:00+02:00",
		"confirmDirect":   true,
		"durationMinutes": 30,
		"assigneeUserId":  userID,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("create visit with hidden assignee: want 400, got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["msgKey"] != "invalid_assignee" {
		t.Fatalf("expected invalid_assignee, got %#v", env["error"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/team/"+memberID, vetTok, map[string]any{
		"includeInCalendar": true,
	})
	if code != http.StatusOK {
		t.Fatalf("patch includeInCalendar true: %d %#v", code, env)
	}
	restored, _ := env["data"].(map[string]any)
	if restored["includeInCalendar"] != true {
		t.Fatalf("expected includeInCalendar=true, got %#v", restored["includeInCalendar"])
	}
	restore = false
}
