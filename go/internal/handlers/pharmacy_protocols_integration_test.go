package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyClinicalProtocolsList(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999033", Name: "Protocol Demo Med", IsActive: true, AMMNumber: "BE-PROTO-1",
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	var practiceID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		practiceID, _ = p["practiceId"].(string)
		if practiceID != "" {
			break
		}
	}
	if practiceID == "" {
		t.Skip("no practiceId on pets")
	}
	_, err = st.UpsertClinicalProtocol(ctx, practiceID, "E2E Protocol", "test", []store.ClinicalProtocolLine{
		{MedicationID: medID, Qty: 1, AMMNumber: "BE-PROTO-1", Unit: "unit"},
	}, 99)
	if err != nil {
		t.Fatalf("upsert protocol: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/protocols", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list protocols %d %#v", code, env)
	}
	payload := dataMap(t, env)
	items, _ := payload["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected protocols %#v", env)
	}
}

func TestPharmacyDAFPetDispensesAndDrafts(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999034", Name: "Dispense Demo Med", IsActive: true, AMMNumber: "BE-DISP-1",
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d", code)
	}
	var petID, practiceID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		petID, _ = p["id"].(string)
		practiceID, _ = p["practiceId"].(string)
		if petID != "" {
			break
		}
	}
	if petID == "" {
		t.Skip("no pets")
	}
	makePetDAFEligible(t, api, petID)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/daf-dispenses", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("client dispenses want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/daf-drafts?visitIds=not-a-uuid,also-bad", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("garbage visitIds want 200 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", tok, map[string]any{
		"scheduledAt": time.Now().UTC().Format(time.RFC3339), "durationMinutes": 30,
		"confirmDirect": true, "consultationSession": true, "silentConfirm": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, tok, map[string]any{"status": "cancelled"})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/daf/for-visit", tok, map[string]any{
		"visitId": visitID,
		"items":   []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-DISP-1"}},
	})
	if code != http.StatusOK {
		t.Fatalf("upsert %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)

	// Fresh draft (<1h) should not appear in stale map.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/daf-drafts?visitIds="+visitID+",not-uuid", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("drafts %d %#v", code, env)
	}
	drafts, _ := dataMap(t, env)["drafts"].(map[string]any)
	if drafts != nil && drafts[visitID] != nil {
		t.Fatalf("fresh draft should not be stale %#v", drafts)
	}

	// Backdate updated_at → draft appears as stale (>1h) for consultation badges.
	if _, err := api.pool.Exec(ctx, `
		UPDATE pharmacy.daf_documents SET updated_at = NOW() - INTERVAL '2 hours' WHERE id = $1::uuid`, dafID); err != nil {
		t.Fatalf("backdate draft: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/consultations/daf-drafts?visitIds="+visitID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("stale drafts %d %#v", code, env)
	}
	drafts, _ = dataMap(t, env)["drafts"].(map[string]any)
	if drafts == nil || drafts[visitID] != dafID {
		t.Fatalf("want stale draft map[%s]=%s got %#v", visitID, dafID, drafts)
	}

	// Preview without stock → stock error with medicationId detail.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/preview-fefo", tok, map[string]any{
		"items": []map[string]any{{"medicationId": medID, "qty": 1}},
	})
	if code != http.StatusConflict {
		// may succeed if seed lots exist — then skip detail assert
		if code != http.StatusOK {
			t.Fatalf("preview %d %#v", code, env)
		}
	} else {
		errObj, _ := env["error"].(map[string]any)
		details, _ := errObj["details"].(map[string]any)
		if details == nil || details["medicationId"] != medID {
			t.Fatalf("want medicationId detail %#v", env)
		}
	}

	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "DISP-LOT-1", "expiresOn": exp, "qty": 5,
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("receipt %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/daf-dispenses", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("dispenses %d %#v", code, env)
	}
	items, _ := dataMap(t, env)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected dispenses after finalize %#v", env)
	}

	// Timeline non-PHI event for owner (localized title/body).
	var et, content string
	err = st.Pool().QueryRow(ctx, `
		SELECT event_type, content FROM pets.dossier_events
		WHERE pet_id = $1::uuid
		ORDER BY created_at DESC LIMIT 1`, petID).Scan(&et, &content)
	if err != nil {
		t.Fatalf("dossier query: %v", err)
	}
	if et == "" || content == "" {
		t.Fatalf("expected dossier timeline event after finalize got type=%q body=%q", et, content)
	}
	_ = practiceID
}
