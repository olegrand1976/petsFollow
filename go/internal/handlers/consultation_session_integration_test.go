package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Walk-in consultationSession must not block (nor be blocked by) a normal booked slot.
func TestConsultationSessionExcludedFromOverlap(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	// Offset within the ±30m walk-in window, away from typical :00/:30 seed slots.
	slotTime := time.Now().UTC().Add(27 * time.Minute)
	slot := slotTime.Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "walk-in overlap probe",
	})
	if code == http.StatusBadRequest {
		// Cabinet on vacation for "now+27m" — skip rather than flake.
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("consultation create %d %#v", code, env)
	}
	consultID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+consultID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})
	if dataMap(t, env)["consultationSession"] != true {
		t.Fatalf("want consultationSession true %#v", env)
	}

	// Same slot, normal booked visit must succeed (walk-in does not occupy the slot).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"notes":           "rdv after walk-in",
	})
	if code != http.StatusCreated {
		t.Fatalf("booked create after consultation %d %#v", code, env)
	}
	bookedID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+bookedID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	// Second normal booking on same slot must conflict.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":     slot,
		"durationMinutes": 30,
		"confirmDirect":   true,
		"notes":           "should conflict",
	})
	if code != http.StatusConflict {
		t.Fatalf("second booked want 409 got %d %#v", code, env)
	}

	// Another walk-in on the same slot still allowed.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "second walk-in",
	})
	if code != http.StatusCreated {
		t.Fatalf("second consultation create %d %#v", code, env)
	}
	consult2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+consult2, vetTok, map[string]any{
			"status": "cancelled",
		})
	})
}

func TestConsultationSessionGuards(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"consultationSession": true,
		"notes":               "missing confirmDirect",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 without confirmDirect got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         "2099-06-15T14:00:00Z",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "far future",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 stale session got %d %#v", code, env)
	}
}

func TestConsultationSessionCancelBlockedAfterReport(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "cancel after report",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	// Empty Ensure draft must still allow cancel.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("ensure report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
		"status": "cancelled",
	})
	if code != http.StatusOK {
		t.Fatalf("cancel empty draft want 200 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "cancel blocked",
	})
	if code != http.StatusCreated {
		t.Fatalf("create2 %d %#v", code, env)
	}
	visit2, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		// may already be non-cancelled if guard works — still try
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2, vetTok, map[string]any{
			"status": "done",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visit2+"/report", vetTok, map[string]any{
		"bodyText": "Examen clinique OK",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visit2, vetTok, map[string]any{
		"status": "cancelled",
	})
	if code != http.StatusConflict {
		t.Fatalf("cancel after report want 409 got %d %#v", code, env)
	}
}

func TestConsultationSessionNotReschedulable(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "no reschedule",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
		"action":              "propose_reschedule",
		"proposedScheduledAt": time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
	})
	if code != http.StatusBadRequest {
		t.Fatalf("reschedule walk-in want 400 got %d %#v", code, env)
	}
}

func TestConsultationSessionFinalizeMarksDone(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "finalize auto-done",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "CR finalisable",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/visits", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list visits %d %#v", code, env)
	}
	found := false
	for _, row := range env["data"].([]any) {
		m, _ := row.(map[string]any)
		if m["id"] == visitID {
			found = true
			if m["status"] != "done" {
				t.Fatalf("after finalize status=%v want done", m["status"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("visit %s not listed", visitID)
	}

	// CTA Terminer must stay idempotent after auto-done.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
		"status": "done",
	})
	if code != http.StatusOK {
		t.Fatalf("idempotent done want 200 got %d %#v", code, env)
	}
}

func TestConsultationSessionOrphanPurge(t *testing.T) {
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
	now := time.Now().UTC().Format(time.RFC3339)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "orphan purge target",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create orphan %d %#v", code, env)
	}
	orphanID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         now,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "has CR — must not purge",
	})
	if code != http.StatusCreated {
		t.Fatalf("create with CR %d %#v", code, env)
	}
	keptID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+keptID, vetTok, map[string]any{
			"status": "done",
		})
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+keptID+"/report", vetTok, map[string]any{
		"bodyText": "contenu à conserver",
	})
	if code != http.StatusOK {
		t.Fatalf("put report kept %d %#v", code, env)
	}

	ctx := context.Background()
	if _, err := api.pool.Exec(ctx, `
		UPDATE visits.visits SET scheduled_at = NOW() - INTERVAL '7 hours' WHERE id = $1`, orphanID); err != nil {
		t.Fatalf("backdate orphan: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE visits.visits SET scheduled_at = NOW() - INTERVAL '7 hours' WHERE id = $1`, keptID); err != nil {
		t.Fatalf("backdate kept: %v", err)
	}

	st := store.New(api.pool)
	n, err := st.CancelStaleConsultationOrphans(ctx, time.Now().Add(-6*time.Hour), 50)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if n < 1 {
		t.Fatalf("purge cancelled=%d want ≥1", n)
	}

	var orphanStatus, keptStatus string
	if err := api.pool.QueryRow(ctx, `SELECT status FROM visits.visits WHERE id=$1`, orphanID).Scan(&orphanStatus); err != nil {
		t.Fatalf("orphan status: %v", err)
	}
	if orphanStatus != "cancelled" {
		t.Fatalf("orphan status=%s want cancelled", orphanStatus)
	}
	if err := api.pool.QueryRow(ctx, `SELECT status FROM visits.visits WHERE id=$1`, keptID).Scan(&keptStatus); err != nil {
		t.Fatalf("kept status: %v", err)
	}
	if keptStatus != "confirmed" {
		t.Fatalf("kept status=%s want confirmed (has CR)", keptStatus)
	}
}

func TestListVetConsultationsHasAudioFlag(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(19 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "list consultations probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("consultation create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "CR list probe",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	reportID, _ := dataMap(t, env)["id"].(string)
	if reportID == "" {
		t.Fatal("missing report id")
	}
	if dataMap(t, env)["hasAudio"] == true {
		t.Fatalf("hasAudio should be false before audio key %#v", env)
	}
	if _, ok := dataMap(t, env)["audioObjectKey"]; ok {
		t.Fatalf("audioObjectKey must not leak %#v", env)
	}

	ctx := context.Background()
	if _, err := api.pool.Exec(ctx, `
		UPDATE visits.visit_reports
		SET audio_object_key = 'visit-reports/`+visitID+`/probe.m4a', updated_at = NOW()
		WHERE id = $1`, reportID); err != nil {
		t.Fatalf("set audio key: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations?q=list+consultations+probe", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list consultations %d %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	found := false
	for _, row := range rows {
		m, _ := row.(map[string]any)
		if m["id"] != visitID {
			continue
		}
		found = true
		if m["hasReport"] != true {
			t.Fatalf("hasReport want true %#v", m)
		}
		if m["hasAudio"] != true {
			t.Fatalf("hasAudio want true %#v", m)
		}
		if _, ok := m["audioObjectKey"]; ok {
			t.Fatalf("audioObjectKey must not leak in list %#v", m)
		}
		if m["clientName"] == "" || m["petName"] == "" {
			t.Fatalf("client/pet names required %#v", m)
		}
	}
	if !found {
		t.Fatalf("visit %s not in consultations list %#v", visitID, rows)
	}
}

func TestVisitReportAudioForbiddenForSecretary(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
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

	slot := time.Now().UTC().Add(23 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"notes":               "secretary audio probe",
	})
	if code == http.StatusBadRequest {
		t.Skipf("consultation create blocked: %#v", env)
	}
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "CR secretary audio probe",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}
	reportID, _ := dataMap(t, env)["id"].(string)
	ctx := context.Background()
	if _, err := api.pool.Exec(ctx, `
		UPDATE visits.visit_reports
		SET audio_object_key = 'visit-reports/`+visitID+`/sec-probe.m4a', updated_at = NOW()
		WHERE id = $1`, reportID); err != nil {
		t.Fatalf("set audio key: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report/audio", secTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("secretary audio want 403 got %d %#v", code, env)
	}
}

