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

	// Créneaux lointains + retry : l'agenda VetPlus est partagé avec le seed et
	// avec les autres tests du paquet, qui tournent en parallèle. Une heure fixe
	// donne un slot_taken selon l'ordre d'exécution.
	book := func(label string, start time.Time, body map[string]any) (map[string]any, time.Time) {
		t.Helper()
		for i := 0; i < 24; i++ {
			at := start.Add(time.Duration(i) * time.Hour)
			body["scheduledAt"] = at.Format(time.RFC3339)
			body["notes"] = label
			code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
			if code == http.StatusCreated {
				v := dataMap(t, env)
				id, _ := v["id"].(string)
				t.Cleanup(func() {
					_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+id, vetTok, map[string]any{
						"status": "cancelled",
					})
				})
				return v, at
			}
			if code != http.StatusConflict && code != http.StatusBadRequest {
				t.Fatalf("create %s %d %#v", label, code, env)
			}
		}
		t.Fatalf("no free slot for %s", label)
		return nil, time.Time{}
	}

	base := time.Now().UTC().Add(120 * 24 * time.Hour).Truncate(time.Hour)

	visit, at := book("typed rdv", base, map[string]any{
		"confirmDirect":   true,
		"visitTypeId":     typeID,
		"durationMinutes": 99, // ignored when type is set
	})
	if visit["visitTypeId"] != typeID {
		t.Fatalf("expected visitTypeId=%s got %#v", typeID, visit["visitTypeId"])
	}
	if visit["durationMinutes"] != float64(30) {
		t.Fatalf("expected duration from type 30, got %#v", visit["durationMinutes"])
	}
	visitID, _ := visit["id"].(string)

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

	visit2, _ := book("manual duration", base.Add(36*time.Hour), map[string]any{
		"confirmDirect":   true,
		"durationMinutes": 20,
	})
	if visit2["durationMinutes"] != float64(20) {
		t.Fatalf("expected manual 20, got %#v", visit2["durationMinutes"])
	}
}
