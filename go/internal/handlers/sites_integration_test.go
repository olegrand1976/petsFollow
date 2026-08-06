package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

func TestPracticeSitesCRUDAndCalendarFilter(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/sites", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list sites: %d %#v", code, env)
	}
	items, _ := env["data"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected at least primary site, got %#v", env)
	}
	primary, _ := items[0].(map[string]any)
	primaryID, _ := primary["id"].(string)
	if primaryID == "" || primary["isPrimary"] != true {
		t.Fatalf("expected primary site, got %#v", primary)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites", vetTok, map[string]any{
		"name": "Antenne Test",
		"city": "Liège",
	})
	if code != http.StatusCreated {
		t.Fatalf("create site: %d %#v", code, env)
	}
	created, _ := env["data"].(map[string]any)
	siteID, _ := created["id"].(string)
	if siteID == "" {
		t.Fatalf("missing site id %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/schedule", vetTok, map[string]any{
		"siteId":               siteID,
		"clientBookingEnabled": true,
		"slotDurationMinutes":  30,
		"slots": []map[string]any{
			{"weekday": int(time.Now().Weekday()), "startTime": "09:00", "endTime": "12:00"},
		},
	})
	if code != http.StatusOK {
		t.Fatalf("put schedule site: %d %#v", code, env)
	}

	from := time.Now().UTC().Format("2006-01-02")
	to := time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02")
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/vet/calendar?from="+from+"&to="+to+"&siteId="+siteID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("calendar filtered: %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("deactivate site: %d %#v", code, env)
	}
	deact, _ := env["data"].(map[string]any)
	if deact["active"] != false {
		t.Fatalf("expected inactive site %#v", deact)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/sites/"+siteID, vetTok, map[string]any{
		"active": true,
	})
	if code != http.StatusOK {
		t.Fatalf("reactivate site: %d %#v", code, env)
	}
	reactivated, _ := env["data"].(map[string]any)
	if reactivated["active"] != true {
		t.Fatalf("expected active site after PATCH %#v", reactivated)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("deactivate before promote: %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/sites/"+siteID, vetTok, map[string]any{
		"isPrimary": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("promote inactive expected 400, got %d %#v", code, env)
	}

	// Cleanup: leave inactive so shared DB stays tidy.
	_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
}

func TestVisitOverlapIsPerSite(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me: %d %#v", code, env)
	}
	me, _ := env["data"].(map[string]any)
	sitesRaw, _ := me["sites"].([]any)
	if len(sitesRaw) < 1 {
		t.Fatalf("me.sites missing %#v", me)
	}
	primary, _ := sitesRaw[0].(map[string]any)
	primaryID, _ := primary["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites", vetTok, map[string]any{
		"name": "Site Overlap B",
	})
	if code != http.StatusCreated {
		t.Fatalf("create site B: %d %#v", code, env)
	}
	siteB, _ := env["data"].(map[string]any)
	siteBID, _ := siteB["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets: %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	pet, _ := pets[0].(map[string]any)
	petID, _ := pet["id"].(string)

	slot := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Hour)
	body := map[string]any{
		"confirmDirect":   true,
		"scheduledAt":     slot.Format(time.RFC3339),
		"siteId":          primaryID,
		"durationMinutes": 30,
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
	if code != http.StatusCreated {
		t.Fatalf("create visit site A: %d %#v", code, env)
	}
	visitA, _ := env["data"].(map[string]any)
	visitAID, _ := visitA["id"].(string)

	// Same slot on other site must succeed (overlap is per-site).
	body["siteId"] = siteBID
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
	if code != http.StatusCreated {
		t.Fatalf("create visit site B same slot expected 201, got %d %#v", code, env)
	}
	visitB, _ := env["data"].(map[string]any)
	visitBID, _ := visitB["id"].(string)

	// Same slot on same site must fail.
	body["siteId"] = primaryID
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
	if code != http.StatusBadRequest && code != http.StatusConflict {
		t.Fatalf("same-site overlap expected 400/409, got %d %#v", code, env)
	}

	// Cleanup: cancel visits then deactivate site B (shared DB).
	if visitAID != "" {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitAID, vetTok, map[string]any{"status": "cancelled"})
	}
	if visitBID != "" {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitBID, vetTok, map[string]any{"status": "cancelled"})
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteBID+"/deactivate", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("cleanup deactivate site B: %d %#v", code, env)
	}
}
