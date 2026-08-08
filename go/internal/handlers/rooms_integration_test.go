package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

func TestRoomsCRUDAndVisitResources(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/sites", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list sites: %d %#v", code, env)
	}
	sites, _ := env["data"].([]any)
	if len(sites) < 1 {
		t.Fatal("expected at least one site")
	}
	primary := sites[0].(map[string]any)
	siteID, _ := primary["id"].(string)
	for _, s := range sites {
		m := s.(map[string]any)
		if m["isPrimary"] == true {
			siteID, _ = m["id"].(string)
			break
		}
	}

	suffix := time.Now().UTC().Format("150405.000")
	roomNameA := "Salle test A " + suffix
	roomNameB := "Salle test B " + suffix

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/rooms", vetTok, map[string]any{
		"name": roomNameA,
	})
	if code != http.StatusCreated {
		t.Fatalf("create room: %d %#v", code, env)
	}
	roomA := env["data"].(map[string]any)
	roomAID, _ := roomA["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/rooms", vetTok, map[string]any{
		"name": roomNameB,
	})
	if code != http.StatusCreated {
		t.Fatalf("create room B: %d %#v", code, env)
	}
	roomB := env["data"].(map[string]any)
	roomBID, _ := roomB["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/sites/"+siteID+"/rooms", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list rooms: %d %#v", code, env)
	}

	// Duplicate name → 400
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/rooms", vetTok, map[string]any{
		"name": roomNameA,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("dup room expected 400, got %d %#v", code, env)
	}

	// Team + pet for booking
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/team", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("team: %d %#v", code, env)
	}
	teamData := env["data"].(map[string]any)
	members, _ := teamData["members"].([]any)
	if len(members) < 1 {
		t.Fatal("no team members")
	}
	assigneeID, _ := members[0].(map[string]any)["userId"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets: %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) < 1 {
		t.Fatal("no pets")
	}
	petID := firstNonWalkinPetID(t, pets)

	// Créneaux lointains + cleanup : l'agenda VetPlus est partagé avec le seed et
	// les autres tests du paquet. Une fenêtre à J+2 donne un slot_taken selon
	// l'ordre d'exécution (CI staging : unassigned should coexist → 409).
	cancelOnCleanup := func(id string) {
		t.Helper()
		if id == "" {
			return
		}
		t.Cleanup(func() {
			_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+id, vetTok, map[string]any{
				"status": "cancelled",
			})
		})
	}

	bookAssigned := func(label string, start time.Time, roomID string) (map[string]any, time.Time) {
		t.Helper()
		for i := range 36 {
			at := start.Add(time.Duration(i) * time.Hour)
			code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
				"scheduledAt":     at.Format(time.RFC3339),
				"siteId":          siteID,
				"confirmDirect":   true,
				"durationMinutes": 30,
				"assigneeUserId":  assigneeID,
				"roomId":          roomID,
				"notes":           label,
			})
			if code == http.StatusOK || code == http.StatusCreated {
				v := env["data"].(map[string]any)
				cancelOnCleanup(v["id"].(string))
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
	for base.Weekday() == time.Saturday || base.Weekday() == time.Sunday {
		base = base.Add(24 * time.Hour)
	}
	base = time.Date(base.Year(), base.Month(), base.Day(), 9, 0, 0, 0, time.UTC)

	visit, start := bookAssigned("resource test", base, roomAID)
	visitID, _ := visit["id"].(string)
	if visit["assigneeUserId"] != assigneeID {
		t.Fatalf("assignee not set: %#v", visit["assigneeUserId"])
	}
	if visit["roomId"] != roomAID {
		t.Fatalf("room not set: %#v", visit["roomId"])
	}

	// Same assignee overlapping → assignee_busy
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     start.Format(time.RFC3339),
		"siteId":          siteID,
		"confirmDirect":   true,
		"durationMinutes": 30,
		"assigneeUserId":  assigneeID,
		"roomId":          roomBID,
	})
	if code != http.StatusConflict {
		t.Fatalf("expected assignee_busy conflict, got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "assignee_busy" && errObj["msgKey"] != "assignee_busy" {
		t.Fatalf("expected assignee_busy code, got %#v", errObj)
	}

	// Same room overlapping (different assignee if available) → room_busy
	otherAssignee := ""
	for _, m := range members {
		uid, _ := m.(map[string]any)["userId"].(string)
		if uid != "" && uid != assigneeID {
			otherAssignee = uid
			break
		}
	}
	if otherAssignee != "" {
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":     start.Format(time.RFC3339),
			"siteId":          siteID,
			"confirmDirect":   true,
			"durationMinutes": 30,
			"assigneeUserId":  otherAssignee,
			"roomId":          roomAID,
		})
		if code != http.StatusConflict {
			t.Fatalf("expected room_busy conflict, got %d %#v", code, env)
		}
		if errObj, _ := env["error"].(map[string]any); errObj["code"] != "room_busy" && errObj["msgKey"] != "room_busy" {
			t.Fatalf("expected room_busy code, got %#v", errObj)
		}
	}

	// Unassigned can share the same wall-clock slot as an assigned visit (parallel).
	// Retry the *pair* together: a free assigned slot may still have an orphan
	// unassigned from seed/other tests → slot_taken on the second create.
	var assignedParallelID, unassignedID string
	var unassignedStart time.Time
	pairBase := start.Add(48 * time.Hour)
	okPair := false
	for i := range 36 {
		at := pairBase.Add(time.Duration(i) * time.Hour)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":     at.Format(time.RFC3339),
			"siteId":          siteID,
			"confirmDirect":   true,
			"durationMinutes": 30,
			"assigneeUserId":  assigneeID,
			"roomId":          roomAID,
			"notes":           "assigned parallel",
		})
		if code != http.StatusOK && code != http.StatusCreated {
			continue
		}
		assigned := env["data"].(map[string]any)
		aid, _ := assigned["id"].(string)
		cancelOnCleanup(aid)

		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":     at.Format(time.RFC3339),
			"siteId":          siteID,
			"confirmDirect":   true,
			"durationMinutes": 30,
			"notes":           "unassigned parallel OK",
		})
		if code == http.StatusOK || code == http.StatusCreated {
			uid, _ := env["data"].(map[string]any)["id"].(string)
			cancelOnCleanup(uid)
			assignedParallelID = aid
			unassignedID = uid
			unassignedStart = at
			okPair = true
			break
		}
		// Slot not suitable for coexistence — cancel the assigned we just created
		// (cleanup still runs too) and try the next hour.
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+aid, vetTok, map[string]any{
			"status": "cancelled",
		})
	}
	if !okPair {
		t.Fatalf("unassigned should coexist with assigned: no free parallel slot")
	}
	_ = unassignedStart

	// Clearing resources into a busy unassigned queue → slot_taken
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+assignedParallelID, vetTok, map[string]any{
		"action":         "set_resources",
		"assigneeUserId": "",
		"roomId":         "",
	})
	if code != http.StatusConflict {
		t.Fatalf("clear into unassigned queue expected slot_taken, got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "slot_taken" && errObj["msgKey"] != "slot_taken" {
		t.Fatalf("expected slot_taken code, got %#v", errObj)
	}
	_ = unassignedID

	// Room on wrong site → invalid_room
	if len(sites) > 1 {
		otherSite := ""
		for _, s := range sites {
			m := s.(map[string]any)
			id, _ := m["id"].(string)
			if id != siteID && m["active"] != false {
				otherSite = id
				break
			}
		}
		if otherSite != "" {
			code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+otherSite+"/rooms", vetTok, map[string]any{
				"name": "Foreign room " + suffix,
			})
			if code == http.StatusCreated {
				foreign := env["data"].(map[string]any)
				fid, _ := foreign["id"].(string)
				code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
					"action": "set_resources",
					"roomId": fid,
				})
				if code != http.StatusBadRequest {
					t.Fatalf("expected invalid_room, got %d %#v", code, env)
				}
			}
		}
	}

	// Deactivate + reactivate room
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/sites/"+siteID+"/rooms/"+roomBID+"/deactivate", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("deactivate room: %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/sites/"+siteID+"/rooms/"+roomBID, vetTok, map[string]any{
		"active": true,
	})
	if code != http.StatusOK {
		t.Fatalf("reactivate room: %d %#v", code, env)
	}
	if room, _ := env["data"].(map[string]any); room["active"] != true {
		t.Fatalf("room not active after reactivate: %#v", room)
	}

	// Assigned visit may reschedule onto a wall-clock slot held by an unassigned visit.
	var parallelSlot time.Time
	okReschedule := false
	rescheduleBase := start.Add(96 * time.Hour)
	for i := range 36 {
		at := rescheduleBase.Add(time.Duration(i) * time.Hour)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":     at.Format(time.RFC3339),
			"siteId":          siteID,
			"confirmDirect":   true,
			"durationMinutes": 30,
			"notes":           "unassigned occupies slot",
		})
		if code != http.StatusOK && code != http.StatusCreated {
			continue
		}
		uid, _ := env["data"].(map[string]any)["id"].(string)
		cancelOnCleanup(uid)

		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"action":              "reschedule_direct",
			"proposedScheduledAt": at.Format(time.RFC3339),
		})
		if code == http.StatusOK {
			parallelSlot = at
			okReschedule = true
			break
		}
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+uid, vetTok, map[string]any{
			"status": "cancelled",
		})
	}
	if !okReschedule {
		t.Fatalf("assigned reschedule onto unassigned slot should succeed")
	}
	_ = parallelSlot

	// Patch team defaultSiteId
	memberID, _ := members[0].(map[string]any)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/team/"+memberID, vetTok, map[string]any{
		"defaultSiteId": siteID,
	})
	// reference_vet may be first member — still allowed for defaultSiteId only
	if code != http.StatusOK && code != http.StatusForbidden {
		t.Fatalf("patch team defaultSite: %d %#v", code, env)
	}
	if code == http.StatusForbidden {
		// pick non-reference
		for _, m := range members {
			mm := m.(map[string]any)
			if mm["teamRole"] != "reference_vet" {
				memberID, _ = mm["id"].(string)
				code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/team/"+memberID, vetTok, map[string]any{
					"defaultSiteId": siteID,
				})
				break
			}
		}
		if code != http.StatusOK {
			t.Fatalf("patch non-ref defaultSite: %d %#v", code, env)
		}
	}
}
