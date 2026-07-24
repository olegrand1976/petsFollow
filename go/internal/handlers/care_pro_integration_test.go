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
