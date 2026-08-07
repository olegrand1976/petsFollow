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
	petID := firstNonWalkinPetID(t, pets)

	// Far-future slots with retry — unassigned queue conflicts with seed/other tests.
	base := time.Now().UTC().Add(90 * 24 * time.Hour).Truncate(time.Hour)
	var visitAID, visitBID string
	for i := range 24 {
		slot := base.Add(time.Duration(i) * time.Hour)
		body := map[string]any{
			"confirmDirect":   true,
			"silentConfirm":   true,
			"scheduledAt":     slot.Format(time.RFC3339),
			"siteId":          primaryID,
			"durationMinutes": 30,
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
		if code == http.StatusConflict || code == http.StatusBadRequest {
			continue
		}
		if code != http.StatusCreated {
			t.Fatalf("create visit site A: %d %#v", code, env)
		}
		visitA, _ := env["data"].(map[string]any)
		visitAID, _ = visitA["id"].(string)

		// Same slot on other site must succeed (overlap is per-site).
		body["siteId"] = siteBID
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
		if code == http.StatusConflict || code == http.StatusBadRequest {
			// Site B collided independently — cancel A and try next hour.
			_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitAID, vetTok, map[string]any{"status": "cancelled"})
			visitAID = ""
			continue
		}
		if code != http.StatusCreated {
			t.Fatalf("create visit site B same slot expected 201, got %d %#v", code, env)
		}
		visitB, _ := env["data"].(map[string]any)
		visitBID, _ = visitB["id"].(string)

		// Same slot on same site must fail.
		body["siteId"] = primaryID
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, body)
		if code != http.StatusBadRequest && code != http.StatusConflict {
			t.Fatalf("same-site overlap expected 400/409, got %d %#v", code, env)
		}
		break
	}
	if visitAID == "" || visitBID == "" {
		t.Fatal("create visit site A/B: all slot retries failed (409 slot_taken)")
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

func TestBookingRequiresSiteWhenMultiBookable(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me: %d %#v", code, env)
	}
	me, _ := env["data"].(map[string]any)
	sitesRaw, _ := me["sites"].([]any)
	if len(sitesRaw) < 2 {
		t.Skip("need seed VetPlus multi-sites (primary + Antenne Liège)")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client pets: %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no client pets")
	}
	pet, _ := pets[0].(map[string]any)
	petID, _ := pet["id"].(string)

	slot := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Hour)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", clientTok, map[string]any{
		"scheduledAt": slot.Format(time.RFC3339),
	})
	if code != http.StatusBadRequest {
		t.Fatalf("booking without siteId expected 400, got %d %#v", code, env)
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj["msgKey"] != "site_required" {
		t.Fatalf("expected msgKey site_required, got %#v", env)
	}

	// Explicit siteId resolves (may still fail slot rules — not under test here).
	primary, _ := sitesRaw[0].(map[string]any)
	primaryID, _ := primary["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", clientTok, map[string]any{
		"scheduledAt": slot.Format(time.RFC3339),
		"siteId":      primaryID,
	})
	if code == http.StatusBadRequest {
		errObj, _ = env["error"].(map[string]any)
		if errObj["msgKey"] == "site_required" {
			t.Fatalf("siteId provided must not yield site_required: %#v", env)
		}
	}
	// Ce POST peut réussir : sans annulation il laisse le créneau occupé pour
	// les autres tests du paquet, qui partagent l'agenda VetPlus.
	if code == http.StatusCreated {
		if id, _ := dataMap(t, env)["id"].(string); id != "" {
			vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
			t.Cleanup(func() {
				_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+id, vetTok, map[string]any{
					"status": "cancelled",
				})
			})
		}
	}
}

func TestDeactivateBlockedByOrphanRequestedVisit(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites", vetTok, map[string]any{
		"name": "Site Orphan Request",
	})
	if code != http.StatusCreated {
		t.Fatalf("create site: %d %#v", code, env)
	}
	site, _ := env["data"].(map[string]any)
	siteID, _ := site["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets: %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID := firstNonWalkinPetID(t, pets)

	// Requested visit without scheduled slot (orphan queue).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"siteId": siteID,
		"notes":  "orphan request no slot",
	})
	if code != http.StatusCreated {
		t.Fatalf("create orphan visit: %d %#v", code, env)
	}
	visit, _ := env["data"].(map[string]any)
	visitID, _ := visit["id"].(string)
	t.Cleanup(func() {
		if visitID != "" {
			_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{"status": "cancelled"})
		}
		_, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("deactivate with orphan expected 400, got %d %#v", code, env)
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj["msgKey"] != "site_has_future_visits" {
		t.Fatalf("expected site_has_future_visits, got %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{"status": "cancelled"})
	if code != http.StatusOK {
		t.Fatalf("cancel orphan: %d %#v", code, env)
	}
	visitID = "" // cleanup already cancelled

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/deactivate", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("deactivate after cancel expected 200, got %d %#v", code, env)
	}
}
