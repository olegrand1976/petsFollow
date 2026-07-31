package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

func TestVisitTypesCRUDAndCreateVisit(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/visit-types", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list empty %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/visit-types", vetTok, map[string]any{
		"items": []map[string]any{
			{"name": "Consultation", "durationMinutes": 30, "color": "#2a9d8f", "sortOrder": 0},
			{"name": "Vaccination", "durationMinutes": 15, "color": "#E9C46A", "sortOrder": 1},
			{"name": "Contrôle", "durationMinutes": 45, "color": "#E76F51", "sortOrder": 2},
		},
	})
	if code != http.StatusOK {
		t.Fatalf("put types %d %#v", code, env)
	}
	items, ok := env["data"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("expected 3 types, got %#v", env["data"])
	}
	first, _ := items[0].(map[string]any)
	typeID, _ := first["id"].(string)
	if typeID == "" || first["durationMinutes"] != float64(30) {
		t.Fatalf("bad first type %#v", first)
	}
	if first["color"] != "#2A9D8F" {
		t.Fatalf("color not normalized %#v", first["color"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/visit-types", vetTok, map[string]any{
		"items": []map[string]any{
			{"name": "Bad", "durationMinutes": 3, "color": "#112233"},
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expected invalid_duration, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	at := time.Now().UTC().Add(72 * time.Hour)
	at = time.Date(at.Year(), at.Month(), at.Day(), 18, 0, 0, 0, time.UTC)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     at.Format(time.RFC3339),
		"confirmDirect":   true,
		"visitTypeId":     typeID,
		"durationMinutes": 99, // ignored when type is set
		"notes":           "typed rdv",
	})
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visit := dataMap(t, env)
	if visit["visitTypeId"] != typeID {
		t.Fatalf("expected visitTypeId=%s got %#v", typeID, visit["visitTypeId"])
	}
	if visit["durationMinutes"] != float64(30) {
		t.Fatalf("expected duration from type 30, got %#v", visit["durationMinutes"])
	}
	visitID, _ := visit["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	from := at.Add(-time.Hour).Format(time.RFC3339)
	to := at.Add(24 * time.Hour).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/calendar?from="+from+"&to="+to, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("calendar %d %#v", code, env)
	}
	cal := dataMap(t, env)
	visits, _ := cal["visits"].([]any)
	found := false
	for _, raw := range visits {
		v, _ := raw.(map[string]any)
		if v["id"] == visitID {
			found = true
			if v["visitTypeName"] != "Consultation" {
				t.Fatalf("calendar type name %#v", v["visitTypeName"])
			}
			if v["visitTypeColor"] != "#2A9D8F" {
				t.Fatalf("calendar type color %#v", v["visitTypeColor"])
			}
		}
	}
	if !found {
		t.Fatalf("created visit not in calendar %#v", cal)
	}

	at2 := at.Add(2 * time.Hour)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     at2.Format(time.RFC3339),
		"confirmDirect":   true,
		"durationMinutes": 20,
		"notes":           "manual duration",
	})
	if code != http.StatusCreated {
		t.Fatalf("create manual %d %#v", code, env)
	}
	visit2 := dataMap(t, env)
	if visit2["durationMinutes"] != float64(20) {
		t.Fatalf("expected manual 20, got %#v", visit2["durationMinutes"])
	}
	visit2ID, _ := visit2["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2ID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})
}
