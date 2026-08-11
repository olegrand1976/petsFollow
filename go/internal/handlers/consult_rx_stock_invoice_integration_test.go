package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// TestConsultConsignesStockInvoiceFlow enchaîne consultation → consignes → DAF/stock → prefill facture
// sur le même visitId (pont Rx→DAF via héritage visitId + FEFO + lignes acte+médicament).
func TestConsultConsignesStockInvoiceFlow(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	suffix := uuid.NewString()[:8]
	cnk := "2779" + suffix[:4]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: cnk, Name: "Flow Bridge Med " + suffix, IsActive: true, IsAntibiotic: false,
		AMMNumber: "BE-FLOW-" + suffix,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	var petID, practiceID string
	for _, row := range pets {
		p, _ := row.(map[string]any)
		if p["practiceId"] != nil && p["practiceId"] != "" {
			petID, _ = p["id"].(string)
			practiceID, _ = p["practiceId"].(string)
			break
		}
	}
	if petID == "" || practiceID == "" {
		t.Fatal("no pet with practice (make seed?)")
	}
	makePetDAFEligible(t, api, petID)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	vtID := uuid.NewString()
	vtName := uniqueLabel("Flow Consult")
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO practice.visit_types (
			id, practice_id, name, duration_minutes, color, is_active, sort_order,
			price_excl_cents, vat_percent
		) VALUES ($1::uuid, $2::uuid, $3, 30, '#2A9D8F', true, 99, 5200, 21)`,
		vtID, practiceID, vtName); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM practice.visit_types WHERE id = $1::uuid`, vtID)
	})

	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pharmacy.medication_prices (practice_id, medication_id, sell_price_cents, vat_percent)
		VALUES ($1::uuid, $2::uuid, 1800, 21)
		ON CONFLICT (practice_id, medication_id) DO UPDATE
		SET sell_price_cents = EXCLUDED.sell_price_cents, vat_percent = EXCLUDED.vat_percent`,
		practiceID, medID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM pharmacy.medication_prices WHERE practice_id = $1::uuid AND medication_id = $2::uuid`,
			practiceID, medID)
	})

	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	lot := "FLOW-LOT-" + suffix
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", vetTok, map[string]any{
		"medicationId": medID, "lotNumber": lot, "expiresOn": exp, "qty": 10, "unit": "box",
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}

	var visitID string
	for attempt := range 4 {
		when := time.Now().UTC().Add(time.Duration(attempt) * time.Hour).Truncate(time.Minute).Format(time.RFC3339)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":         when,
			"durationMinutes":     30,
			"confirmDirect":       true,
			"silentConfirm":       true,
			"consultationSession": true,
			"visitTypeId":         vtID,
			"notes":               "flow consult-rx-stock-invoice",
		})
		if code == http.StatusCreated || code == http.StatusOK {
			visitID, _ = dataMap(t, env)["id"].(string)
			break
		}
		if code != http.StatusConflict {
			t.Fatalf("create visit %d %#v", code, env)
		}
	}
	if visitID == "" {
		t.Fatal("create visit: all slot retries failed (409)")
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "## Plan\nRepos + Flow Bridge Med 2 boites\n",
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId":      petID,
		"visitId":    visitID,
		"careAdvice": "Repos 48h — flux e2e",
		"medications": []map[string]any{{
			"name": "Flow Bridge Med " + suffix, "dosage": "1", "form": "box",
			"quantity": "2", "posology": "SID",
			"cnk": cnk, "ref_medication_id": medID,
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("rx create %d %#v", code, env)
	}
	rx := dataMap(t, env)
	rxID, _ := rx["id"].(string)
	if rx["visitId"] != visitID {
		t.Fatalf("rx visitId %#v want %s", rx["visitId"], visitID)
	}

	// Sans visitId dans le body : héritage depuis la consigne liée à la visite.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/from-prescription", vetTok, map[string]any{
		"prescriptionId": rxID,
	})
	if code != http.StatusCreated {
		t.Fatalf("from-prescription %d %#v", code, env)
	}
	daf := dataMap(t, env)
	dafID, _ := daf["id"].(string)
	if daf["status"] != "draft" {
		t.Fatalf("daf status %#v", daf)
	}
	if daf["visitId"] != visitID {
		t.Fatalf("daf visitId %#v want %s (héritage consignes)", daf["visitId"], visitID)
	}
	if daf["petId"] != petID {
		t.Fatalf("daf petId %#v", daf)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	fin := dataMap(t, env)
	if d, ok := fin["daf"].(map[string]any); ok {
		fin = d
	}
	if fin["status"] != "finalized" {
		t.Fatalf("finalize status %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/movements?dafId="+dafID+"&limit=20", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("movements %d %#v", code, env)
	}
	movData, _ := env["data"].(map[string]any)
	movItems, _ := movData["items"].([]any)
	if len(movItems) < 1 {
		t.Fatalf("expected stock movements %#v", env)
	}
	var deltaSum float64
	linked := 0
	for _, raw := range movItems {
		m, _ := raw.(map[string]any)
		if m["reason"] != "daf" || m["dafId"] != dafID {
			continue
		}
		linked++
		switch d := m["delta"].(type) {
		case float64:
			deltaSum += d
		case int:
			deltaSum += float64(d)
		default:
			t.Fatalf("delta type %#v", m["delta"])
		}
		if m["lotNumber"] != lot {
			t.Fatalf("FEFO lot=%v want %s (médicament unique au run)", m["lotNumber"], lot)
		}
		if m["dafItemId"] == nil || m["dafItemId"] == "" {
			t.Fatalf("missing dafItemId %#v", m)
		}
	}
	if linked < 1 {
		t.Fatalf("no daf-linked movements %#v", env)
	}
	if deltaSum != -2 {
		t.Fatalf("expected total delta -2 got %v %#v", deltaSum, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/practices/me/invoicing/prefill?visitId="+visitID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("prefill %d %#v", code, env)
	}
	lines := prefillLines(t, env)
	if len(lines) != 2 {
		t.Fatalf("want act + 1 daf line, got %d %#v", len(lines), env)
	}
	if lines[0]["description"] != vtName {
		t.Fatalf("act description=%v want %s", lines[0]["description"], vtName)
	}
	if lines[0]["unitPriceExclCents"] != float64(5200) {
		t.Fatalf("act price %#v", lines[0])
	}
	if lines[1]["unitPriceExclCents"] != float64(1800) || lines[1]["quantity"] != float64(2) {
		t.Fatalf("daf line %#v", lines[1])
	}
	if dataMap(t, env)["dafId"] != dafID {
		t.Fatalf("prefill dafId=%v want %s", dataMap(t, env)["dafId"], dafID)
	}
}

// TestPharmacyDAFFromPrescriptionVisitMismatch refuse un body.visitId ≠ consignes.visitId.
func TestPharmacyDAFFromPrescriptionVisitMismatch(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2780" + suffix[:4], Name: "Mismatch Med " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	petID := demoPetIDWithPractice(t, api)
	makePetDAFEligible(t, api, petID)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	visitA := createVisitWithOptionalReport(t, api, vetTok, petID, "cr A", "mismatch-a", -3*time.Hour, false, false)
	visitB := createVisitWithOptionalReport(t, api, vetTok, petID, "cr B", "mismatch-b", -4*time.Hour, false, false)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId":   petID,
		"visitId": visitA,
		"medications": []map[string]any{{
			"name": "Mismatch Med", "quantity": "1", "ref_medication_id": medID,
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("rx %d %#v", code, env)
	}
	rxID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/from-prescription", vetTok, map[string]any{
		"prescriptionId": rxID,
		"visitId":        visitB,
	})
	if code != http.StatusBadRequest || errCode(env) != "visit_mismatch" {
		t.Fatalf("want visit_mismatch got %d %q %#v", code, errCode(env), env)
	}
}
