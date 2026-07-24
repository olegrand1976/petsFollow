package handlers_test

import (
	"net/http"
	"testing"
)

func TestCareProLoginAndListPets(t *testing.T) {
	api := newTestAPI(t)

	for _, tc := range []struct {
		email, pass, specialty string
	}{
		{"farrier.demo@petsfollow.test", "CareProDemo123!", "farrier"},
		{"vetlight.demo@petsfollow.test", "CareProDemo123!", "vet_light"},
	} {
		t.Run(tc.specialty, func(t *testing.T) {
			tok := loginToken(t, api.handler, tc.email, tc.pass)
			code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
			if code != http.StatusOK {
				t.Fatalf("me %d %#v (make seed?)", code, env)
			}
			me := dataMap(t, env)
			if me["role"] != "care_pro" {
				t.Fatalf("role=%v", me["role"])
			}
			if me["professionalSpecialty"] != tc.specialty {
				t.Fatalf("specialty=%v want %s", me["professionalSpecialty"], tc.specialty)
			}

			code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/care-pro/pets", tok, nil)
			if code != http.StatusOK {
				t.Fatalf("pets %d %#v", code, env)
			}
			pets, ok := env["data"].([]any)
			if !ok || len(pets) == 0 {
				t.Fatalf("expected Spirit grant, got %#v", env["data"])
			}
			found := false
			for _, row := range pets {
				p, _ := row.(map[string]any)
				name, _ := p["name"].(string)
				perm, _ := p["permission"].(string)
				if name == "Spirit" && (perm == "write_notes" || perm == "full") {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Spirit with write_notes missing: %#v", pets)
			}

			code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/care-pro/visits", tok, nil)
			if code != http.StatusOK {
				t.Fatalf("visits %d %#v", code, env)
			}
		})
	}
}

func TestCareProWriteNotesACLForbidden(t *testing.T) {
	api := newTestAPI(t)
	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("owner pets %d %#v", code, env)
	}
	var otherPetID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] != "Spirit" {
			otherPetID, _ = p["id"].(string)
			if otherPetID != "" {
				break
			}
		}
	}
	if otherPetID == "" {
		t.Skip("no non-Spirit pet for ACL check (seed incomplete)")
	}

	// Farrier has no grant on other pets → create visit must 403.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+otherPetID+"/visits", farrierTok, map[string]any{
		"scheduledAt":     "2099-01-15T10:00:00Z",
		"notes":           "ACL probe",
		"durationMinutes": 30,
	})
	if code != http.StatusForbidden {
		t.Fatalf("expected forbidden create visit, got %d %#v", code, env)
	}

	// Read-only grant: list OK-ish path via care-pro pets after temporary read share.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+otherPetID+"/shares", ownerTok, map[string]any{
		"email":      "farrier.demo@petsfollow.test",
		"permission": "read",
	})
	if code != http.StatusCreated {
		t.Fatalf("share read %d %#v", code, env)
	}
	grant := dataMap(t, env)
	granteeID, _ := grant["granteeUserId"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete,
			"/api/v1/pets/"+otherPetID+"/shares/"+granteeID, ownerTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+otherPetID+"/visits", farrierTok, map[string]any{
		"scheduledAt":     "2099-02-15T10:00:00Z",
		"notes":           "ACL probe read-only",
		"durationMinutes": 30,
	})
	if code != http.StatusForbidden {
		t.Fatalf("read-only should forbid visit create, got %d %#v", code, env)
	}
}

func TestCareProCannotAccessAdmin(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/stripe-catalog", tok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("care_pro admin access: got %d %#v", code, env)
	}
}

func TestCareProClearCoordsAndMarkDone(t *testing.T) {
	api := newTestAPI(t)
	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/care-pro/pets", farrierTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
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
		t.Skip("Spirit not granted to farrier")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+spiritID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-07-01T10:00:00Z",
		"notes":           "care_pro clearCoords/done probe",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("vet create visit %d %#v", code, env)
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

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/location", farrierTok, map[string]any{
		"addressText": "Écurie test clearCoords",
		"lat":         50.85,
		"lng":         4.35,
	})
	if code != http.StatusOK {
		t.Fatalf("set location %d %#v", code, env)
	}
	loc := dataMap(t, env)
	if loc["lat"] == nil || loc["lng"] == nil {
		t.Fatalf("expected lat/lng after set, got %#v", loc)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/location", farrierTok, map[string]any{
		"addressText": "Écurie test sans GPS",
		"clearCoords": true,
	})
	if code != http.StatusOK {
		t.Fatalf("clearCoords %d %#v", code, env)
	}
	cleared := dataMap(t, env)
	if cleared["lat"] != nil || cleared["lng"] != nil {
		t.Fatalf("expected null lat/lng after clearCoords, got %#v", cleared)
	}
	if addr, _ := cleared["addressText"].(string); addr != "Écurie test sans GPS" {
		t.Fatalf("addressText=%v", cleared["addressText"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, farrierTok, map[string]any{
		"status": "done",
	})
	if code != http.StatusOK {
		t.Fatalf("mark done %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "done" {
		t.Fatalf("status=%v want done", dataMap(t, env)["status"])
	}
}

func TestCareProReadOnlyCannotClearCoords(t *testing.T) {
	api := newTestAPI(t)
	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("owner pets %d %#v", code, env)
	}
	var petID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] != "Spirit" {
			petID, _ = p["id"].(string)
			if petID != "" {
				break
			}
		}
	}
	if petID == "" {
		t.Skip("no non-Spirit pet")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     "2099-06-01T10:00:00Z",
		"notes":           "read-only location probe",
		"durationMinutes": 30,
		"confirmDirect":   true,
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

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/shares", ownerTok, map[string]any{
		"email":      "farrier.demo@petsfollow.test",
		"permission": "read",
	})
	if code != http.StatusCreated && code != http.StatusOK && code != http.StatusConflict {
		t.Fatalf("share read %d %#v", code, env)
	}
	if code == http.StatusCreated || code == http.StatusOK {
		grant := dataMap(t, env)
		granteeID, _ := grant["granteeUserId"].(string)
		if granteeID != "" {
			t.Cleanup(func() {
				_, _ = doAuthJSON(t, api.handler, http.MethodDelete,
					"/api/v1/pets/"+petID+"/shares/"+granteeID, ownerTok, nil)
			})
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID+"/location", farrierTok, map[string]any{
		"addressText": "should fail",
		"clearCoords": true,
	})
	if code != http.StatusForbidden {
		t.Fatalf("read-only clearCoords: got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, farrierTok, map[string]any{
		"status": "done",
	})
	if code != http.StatusForbidden {
		t.Fatalf("read-only mark done: got %d %#v", code, env)
	}
}
