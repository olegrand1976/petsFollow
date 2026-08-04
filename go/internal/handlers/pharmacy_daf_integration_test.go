package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyDAFFinalizeCancelPDF(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://localhost:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999001", Name: "DAF Demo Med", IsActive: true, IsAntibiotic: false,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	abID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999002", Name: "DAF AB Med", IsActive: true, IsAntibiotic: true,
	})
	if err != nil {
		t.Fatalf("ab: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "DAF-LOT-1", "expiresOn": exp, "qty": 10,
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"items": []map[string]any{{
			"medicationId": abID, "qty": 1, "ammNumber": "BE-V1",
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft ab %d %#v", code, env)
	}
	abIDDoc, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+abIDDoc+"/finalize", tok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("antibio finalize want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"notes": "demo",
		"items": []map[string]any{{
			"medicationId": medID, "qty": 2, "ammNumber": "BE-V-DEMO-1", "unit": "box",
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/preview-fefo", tok, map[string]any{
		"items": []map[string]any{{"medicationId": medID, "qty": 2}},
	})
	if code != http.StatusOK {
		t.Fatalf("preview %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	fin := dataMap(t, env)
	if d, ok := fin["daf"].(map[string]any); ok {
		fin = d
	}
	if fin["status"] != "finalized" {
		t.Fatalf("status %#v", env)
	}
	if fin["displayNumber"] == nil || fin["displayNumber"] == "" {
		t.Fatalf("missing displayNumber %#v", fin)
	}
	num1 := fin["dafNumber"]

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/movements?dafId="+dafID+"&limit=20", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("movements %d %#v", code, env)
	}
	movData, _ := env["data"].(map[string]any)
	movItems, _ := movData["items"].([]any)
	if len(movItems) < 1 {
		t.Fatalf("expected daf-linked movements %#v", env)
	}
	mov0, _ := movItems[0].(map[string]any)
	if mov0["dafId"] != dafID {
		t.Fatalf("movement dafId %#v", mov0)
	}
	if mov0["dafItemId"] == nil || mov0["dafItemId"] == "" {
		t.Fatalf("missing dafItemId %#v", mov0)
	}
	if mov0["lotNumber"] != "DAF-LOT-1" {
		t.Fatalf("lot on movement %#v", mov0)
	}
	if mov0["reason"] != "daf" {
		t.Fatalf("reason %#v", mov0)
	}
	if mov0["delta"].(float64) != -2 {
		t.Fatalf("expected delta -2 %#v", mov0)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"items": []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-V-DEMO-2"}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft2 %d %#v", code, env)
	}
	d2 := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+d2+"/finalize", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize2 %d %#v", code, env)
	}
	fin2 := dataMap(t, env)
	if d, ok := fin2["daf"].(map[string]any); ok {
		fin2 = d
	}
	if fin2["dafNumber"] == num1 {
		t.Fatalf("expected gapless increment %#v vs %#v", fin2["dafNumber"], num1)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vet/pharmacy/daf/"+dafID+"/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pdf %d %s", rec.Code, rec.Body.String())
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) {
		t.Fatalf("not a pdf")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/cancel", tok, map[string]any{
		"reason": "error", "restock": true,
	})
	if code != http.StatusOK {
		t.Fatalf("cancel %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "cancelled" {
		t.Fatalf("cancel status %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/movements?dafId="+dafID+"&limit=50", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("movements after cancel %d %#v", code, env)
	}
	movData, _ = env["data"].(map[string]any)
	movItems, _ = movData["items"].([]any)
	var sawCancel bool
	for _, raw := range movItems {
		m, _ := raw.(map[string]any)
		if m["reason"] == "daf_cancel" && m["dafId"] == dafID && m["dafItemId"] != "" {
			sawCancel = true
			break
		}
	}
	if !sawCancel {
		t.Fatalf("expected daf_cancel movement linked to daf %#v", env)
	}

	// Annulation sans restock → mouvement delta=0 toujours tracé.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "DAF-LOT-NR", "expiresOn": exp, "qty": 3,
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt nr %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"items": []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-V-NR"}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft nr %d %#v", code, env)
	}
	dafNR := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafNR+"/finalize", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize nr %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafNR+"/cancel", tok, map[string]any{
		"reason": "demo", "restock": false,
	})
	if code != http.StatusOK {
		t.Fatalf("cancel nr %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/movements?dafId="+dafNR+"&limit=20", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("movements nr %d %#v", code, env)
	}
	movData, _ = env["data"].(map[string]any)
	movItems, _ = movData["items"].([]any)
	var sawZeroCancel bool
	for _, raw := range movItems {
		m, _ := raw.(map[string]any)
		if m["reason"] == "daf_cancel" && m["dafId"] == dafNR {
			if m["delta"].(float64) != 0 {
				t.Fatalf("no_restock cancel want delta 0 %#v", m)
			}
			sawZeroCancel = true
		}
	}
	if !sawZeroCancel {
		t.Fatalf("expected delta=0 daf_cancel %#v", env)
	}
}

func TestPharmacyDAFWithVisitID(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999010", Name: "DAF Visit Link Med", IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	var petID, ownerID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		if p["name"] == "Rex" || p["name"] == "Bella" || p["name"] == "Spirit" {
			petID, _ = p["id"].(string)
			ownerID, _ = p["ownerUserId"].(string)
			if petID != "" {
				break
			}
		}
	}
	if petID == "" {
		// fallback first pet
		if rows, ok := env["data"].([]any); ok && len(rows) > 0 {
			p, _ := rows[0].(map[string]any)
			petID, _ = p["id"].(string)
			ownerID, _ = p["ownerUserId"].(string)
		}
	}
	if petID == "" {
		t.Skip("no pets seeded")
	}
	makePetDAFEligible(t, api, petID)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", tok, map[string]any{
		"scheduledAt":     "2099-04-21T11:00:00Z",
		"notes":           "daf visit link",
		"durationMinutes": 30,
		"confirmDirect":   true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, tok, map[string]any{"status": "cancelled"})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"clientUserId": ownerID,
		"petId":        petID,
		"visitId":      visitID,
		"notes":        "from consultation",
		"items": []map[string]any{{
			"medicationId": medID, "qty": 1, "ammNumber": "BE-VISIT-1",
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft with visit %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if doc["visitId"] != visitID {
		t.Fatalf("visitId=%v want %s", doc["visitId"], visitID)
	}
	if ownerID != "" && doc["clientUserId"] != ownerID {
		t.Fatalf("clientUserId=%v want %s", doc["clientUserId"], ownerID)
	}
	if doc["petId"] != petID {
		t.Fatalf("petId=%v want %s", doc["petId"], petID)
	}

	// Foreign visit UUID must be rejected (ownership).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"clientUserId": ownerID,
		"petId":        petID,
		"visitId":      "00000000-0000-4000-8000-000000000099",
		"items": []map[string]any{{
			"medicationId": medID, "qty": 1, "ammNumber": "BE-VISIT-BAD",
		}},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("foreign visit want 400 got %d %#v", code, env)
	}
}

func TestPharmacyDAFVAMRegDryRun(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("VAMREG_DRY_RUN", "true")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://localhost:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	ctx := context.Background()
	st := store.New(api.pool)
	abID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999011", Name: "VAMReg AB Med", IsActive: true, IsAntibiotic: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": abID, "lotNumber": "VAM-LOT-1", "expiresOn": exp, "qty": 5,
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"items": []map[string]any{{
			"medicationId": abID, "qty": 1, "ammNumber": "BE-VAM-1",
			"vamregPayload": map[string]any{"species": "dog", "indication": "infection", "durationDays": 7},
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if doc["vamregStatus"] != "sent" {
		t.Fatalf("vamregStatus=%v want sent (dry-run sync)", doc["vamregStatus"])
	}
	if doc["hasAntibiotic"] != true {
		t.Fatalf("hasAntibiotic=%v", doc["hasAntibiotic"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/daf/"+dafID+"/vamreg/audits", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("audits %d %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	if len(rows) < 1 {
		t.Fatalf("want audit rows, got %#v", env)
	}

	// Cancel blocked once VAMReg sent.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/cancel", tok, map[string]any{
		"reason": "e2e", "restock": true,
	})
	if code != http.StatusConflict {
		t.Fatalf("cancel after sent want 409 got %d %#v", code, env)
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj["code"] != "daf_vamreg_already_sent" {
		t.Fatalf("want daf_vamreg_already_sent got %#v", env)
	}
}

func TestPharmacyDAFCancelBlockedWhileVAMRegPending(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("VAMREG_DRY_RUN", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	abID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999012", Name: "VAMReg Pending Med", IsActive: true, IsAntibiotic: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": abID, "lotNumber": "VAM-PEND-1", "expiresOn": exp, "qty": 3,
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
		"items": []map[string]any{{
			"medicationId": abID, "qty": 1, "ammNumber": "BE-PEND-1",
			"vamregPayload": map[string]any{"species": "dog", "indication": "infection", "durationDays": 5},
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if nested, ok := doc["daf"].(map[string]any); ok {
		doc = nested
	}
	practiceID, _ := doc["practiceId"].(string)
	if practiceID == "" {
		t.Fatalf("missing practiceId %#v", doc)
	}
	// Dry-run marks sent; force pending via SQL to exercise in-flight cancel guard.
	if _, err := api.pool.Exec(ctx, `
		UPDATE pharmacy.daf_documents SET vamreg_status = 'pending'
		WHERE practice_id = $1 AND id = $2`, practiceID, dafID); err != nil {
		t.Fatalf("force pending: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/cancel", tok, map[string]any{
		"reason": "race", "restock": true,
	})
	if code != http.StatusConflict {
		t.Fatalf("cancel pending want 409 got %d %#v", code, env)
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj["code"] != "daf_vamreg_in_flight" {
		t.Fatalf("want daf_vamreg_in_flight got %#v", env)
	}
}

func TestPharmacyDAFUpsertForVisit(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999020", Name: "Consult DAF Med", IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	var petID string
	for _, row := range env["data"].([]any) {
		p, _ := row.(map[string]any)
		petID, _ = p["id"].(string)
		if petID != "" {
			break
		}
	}
	if petID == "" {
		t.Skip("no pets seeded")
	}
	makePetDAFEligible(t, api, petID)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", tok, map[string]any{
		"scheduledAt":         time.Now().UTC().Format(time.RFC3339),
		"durationMinutes":     30,
		"confirmDirect":       true,
		"consultationSession": true,
		"silentConfirm":       true,
	})
	if code != http.StatusCreated {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := dataMap(t, env)["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, tok, map[string]any{"status": "cancelled"})
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/daf/for-visit?visitId="+visitID, tok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("get empty want 404 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/daf/for-visit", tok, map[string]any{
		"visitId": visitID,
		"items": []map[string]any{{
			"medicationId": medID, "qty": 1, "ammNumber": "BE-CONSULT-1",
		}},
	})
	if code != http.StatusOK {
		t.Fatalf("upsert create %d %#v", code, env)
	}
	doc1 := dataMap(t, env)
	dafID, _ := doc1["id"].(string)
	if doc1["visitId"] != visitID || doc1["status"] != "draft" {
		t.Fatalf("unexpected upsert %#v", doc1)
	}
	if n := len(doc1["items"].([]any)); n != 1 {
		t.Fatalf("items=%d %#v", n, doc1)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/daf/for-visit", tok, map[string]any{
		"visitId": visitID,
		"items": []map[string]any{{
			"medicationId": medID, "qty": 3, "ammNumber": "BE-CONSULT-2",
		}},
	})
	if code != http.StatusOK {
		t.Fatalf("upsert replace %d %#v", code, env)
	}
	doc2 := dataMap(t, env)
	if doc2["id"] != dafID {
		t.Fatalf("expected same draft id %s got %#v", dafID, doc2)
	}
	item0, _ := doc2["items"].([]any)[0].(map[string]any)
	if item0["qty"] != float64(3) {
		t.Fatalf("qty not replaced %#v", item0)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/daf/for-visit?visitId="+visitID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get draft %d %#v", code, env)
	}
	if dataMap(t, env)["id"] != dafID {
		t.Fatalf("get mismatch %#v", env)
	}
}

func TestPharmacyDAFFromPrescription(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999021", Name: "Rx Bridge Med", IsActive: true,
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
	var petID string
	for _, row := range pets {
		p, _ := row.(map[string]any)
		if p["practiceId"] != nil && p["practiceId"] != "" {
			petID, _ = p["id"].(string)
			break
		}
	}
	if petID == "" {
		t.Fatal("no pet with practice")
	}
	makePetDAFEligible(t, api, petID)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID,
		"medications": []map[string]any{{
			"name": "Rx Bridge Med", "dosage": "1", "form": "tab",
			"quantity": "2", "posology": "SID",
			"cnk": "2999021", "ref_medication_id": medID,
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("rx create %d %#v", code, env)
	}
	rxID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/from-prescription", vetTok, map[string]any{
		"prescriptionId": rxID,
	})
	if code != http.StatusCreated {
		t.Fatalf("from-prescription %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if doc["status"] != "draft" {
		t.Fatalf("status %#v", doc)
	}
	if doc["petId"] != petID {
		t.Fatalf("petId %#v", doc)
	}
	items, _ := doc["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items %#v", doc)
	}
	it0, _ := items[0].(map[string]any)
	if it0["medicationId"] != medID {
		t.Fatalf("medicationId %#v", it0)
	}
	if it0["qty"] != float64(2) {
		t.Fatalf("qty from rx quantity %#v", it0)
	}

	// Without catalog link → daf_empty.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID,
		"medications": []map[string]any{{
			"name": "Free text only", "quantity": "1",
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("rx free %d %#v", code, env)
	}
	rxFree, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/from-prescription", vetTok, map[string]any{
		"prescriptionId": rxFree,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("want daf_empty 400 got %d %#v", code, env)
	}
}
