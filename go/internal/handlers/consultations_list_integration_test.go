package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

// C2.17 — /vet/consultations lists walk-ins and agenda visits with a persisted CR
// (not bare agenda RDVs without report). Soft-delete: walk-in or done/cancelled+CR.
func TestListConsultationsIncludesAgendaWithReport(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	// Unique far-future slots (retry) to avoid slot_taken vs seed/other tests.
	pickSlot := func(label string, start time.Time) (id string, ok bool) {
		for i := range 24 {
			slot := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
			code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
				"scheduledAt":     slot,
				"durationMinutes": 30,
				"confirmDirect":   true,
				"silentConfirm":   true,
				"notes":           label,
			})
			if code == http.StatusCreated {
				return dataMap(t, env)["id"].(string), true
			}
			if code != http.StatusConflict && code != http.StatusBadRequest {
				t.Fatalf("create %s %d %#v", label, code, env)
			}
		}
		return "", false
	}

	base := time.Now().UTC().Add(90 * 24 * time.Hour).Truncate(time.Hour)

	// 1) Agenda RDV without CR → must NOT appear ; soft-delete refused.
	agendaNoCR, ok := pickSlot("agenda no-cr list probe", base)
	if !ok {
		t.Skip("agenda create blocked: no free slot")
	}

	// 2) Agenda RDV + CR (still confirmed) → appears ; soft-delete refused (open RDV).
	agendaWithCR, ok := pickSlot("agenda with-cr list probe", base.Add(36*time.Hour))
	if !ok {
		t.Skip("agenda+cr create blocked: no free slot")
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+agendaWithCR+"/report", vetTok, map[string]any{
		"bodyText": "CR agenda list probe",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}

	// 3) Walk-in without CR → must appear (scheduledAt ≈ now ; no slot lock).
	slotWalkIn := time.Now().UTC().Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slotWalkIn,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "walk-in no-cr list probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("walk-in create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create walk-in %d %#v", code, env)
	}
	walkIn, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	if consultationsContain(env, agendaNoCR) {
		t.Fatalf("agenda visit without CR must not appear in consultations list")
	}
	if !consultationsContain(env, agendaWithCR) {
		t.Fatalf("agenda visit with persisted CR must appear in consultations list")
	}
	if !consultationsContain(env, walkIn) {
		t.Fatalf("walk-in without CR must appear in consultations list")
	}

	// Open agenda+CR cannot be soft-deleted (would drop the upcoming RDV from calendar).
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+agendaWithCR, vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("delete open agenda+cr want 400 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["msgKey"] != "not_consultation_history" {
		t.Fatalf("delete open agenda+cr msgKey want not_consultation_history %#v", env)
	}

	// Bare agenda without CR must refuse soft-delete.
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+agendaNoCR, vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("delete agenda without CR want 400 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["msgKey"] != "not_consultation_history" {
		t.Fatalf("delete agenda without CR msgKey want not_consultation_history %#v", env)
	}

	// Mark agenda+CR done → soft-delete OK and disappears from list.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+agendaWithCR, vetTok, map[string]any{
		"status": "done",
	})
	if code != http.StatusOK {
		t.Fatalf("mark agenda+cr done %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+agendaWithCR, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("delete done agenda+cr want 200 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list after agenda delete %d %#v", code, env)
	}
	if consultationsContain(env, agendaWithCR) {
		t.Fatalf("soft-deleted agenda+cr must not appear in consultations list")
	}
	if !consultationsContain(env, walkIn) {
		t.Fatalf("walk-in must still appear after agenda soft-delete")
	}
}

// GET /vet/consultations/{id} — meta for CR detail page (practice-scoped).
func TestGetVetConsultationByID(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	pet := pets[0].(map[string]any)
	petID, _ := pet["id"].(string)
	petName, _ := pet["name"].(string)

	slotWalkIn := time.Now().UTC().Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slotWalkIn,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "get-by-id probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("walk-in create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create walk-in %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/"+visitID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get consultation %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["id"] != visitID {
		t.Fatalf("id want %s got %#v", visitID, got["id"])
	}
	if got["petId"] != petID {
		t.Fatalf("petId want %s got %#v", petID, got["petId"])
	}
	if petName != "" && got["petName"] != petName {
		t.Fatalf("petName want %s got %#v", petName, got["petName"])
	}
	if got["clientId"] == nil || got["clientId"] == "" {
		t.Fatalf("clientId missing %#v", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/not-a-uuid", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("invalid uuid want 404 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/00000000-0000-4000-8000-000000000001", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("unknown uuid want 404 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("soft-delete walk-in %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/"+visitID, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("soft-deleted want 404 got %d %#v", code, env)
	}
}
