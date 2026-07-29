package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func cancelOpenInventorySessions(t *testing.T, api *testAPI, tok string) {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/inventory/sessions", tok, nil)
	if code != http.StatusOK {
		return
	}
	rows, _ := env["data"].([]any)
	for _, raw := range rows {
		s, _ := raw.(map[string]any)
		if s["status"] == "open" {
			sid, _ := s["id"].(string)
			doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/inventory/sessions/"+sid+"/cancel", tok, map[string]any{})
		}
	}
}

func inventoryCountsForClose(lines []any, targetLineID string, targetCounted float64) []map[string]any {
	counts := make([]map[string]any, 0, len(lines))
	for _, raw := range lines {
		ln, _ := raw.(map[string]any)
		id, _ := ln["id"].(string)
		sys, _ := ln["systemQty"].(float64)
		qty := sys
		if id == targetLineID {
			qty = targetCounted
		}
		counts = append(counts, map[string]any{"lineId": id, "countedQty": qty})
	}
	return counts
}

func TestPharmacyInventorySession(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2888" + suffix[:4], Name: "Inv Med " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	cancelOpenInventorySessions(t, api, tok)

	exp := time.Now().AddDate(0, 0, 200).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "LOT-INV-" + suffix, "expiresOn": exp, "qty": 10,
	})
	if code != http.StatusCreated {
		t.Fatalf("receive %d %#v", code, env)
	}
	batchObj, _ := dataMap(t, env)["batch"].(map[string]any)
	batchID, _ := batchObj["id"].(string)
	if batchID == "" {
		t.Fatalf("no batch id %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/inventory/sessions", tok, map[string]any{})
	if code != http.StatusCreated {
		t.Fatalf("start inv %d %#v", code, env)
	}
	sess := dataMap(t, env)
	sessionID, _ := sess["id"].(string)
	lines, _ := sess["lines"].([]any)
	if sessionID == "" || len(lines) == 0 {
		t.Fatal("missing session/lines")
	}

	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/inventory/sessions", tok, map[string]any{})
	if code != http.StatusConflict {
		t.Fatalf("second open want 409 got %d", code)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/vet/pharmacy/inventory/sessions/"+sessionID+"/close", tok, map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("incomplete want 400 got %d %#v", code, env)
	}

	var lineID string
	for _, raw := range lines {
		ln, _ := raw.(map[string]any)
		if ln["lotNumber"] == "LOT-INV-"+suffix || ln["batchId"] == batchID {
			lineID, _ = ln["id"].(string)
			batchID, _ = ln["batchId"].(string)
			break
		}
	}
	if lineID == "" {
		t.Fatal("target line not found")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches/"+batchID+"/adjust", tok, map[string]any{
		"delta": -2, "detail": "drift-during-inv",
	})
	if code != http.StatusOK {
		t.Fatalf("drift adjust %d %#v", code, env)
	}

	counts := inventoryCountsForClose(lines, lineID, 7)
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/vet/pharmacy/inventory/sessions/"+sessionID+"/close", tok, map[string]any{"counts": counts})
	if code != http.StatusOK {
		t.Fatalf("close %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "closed" {
		t.Fatalf("status=%v", dataMap(t, env)["status"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/batches", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("batches %d %#v", code, env)
	}
	found := false
	for _, raw := range dataMap(t, env)["items"].([]any) {
		b, _ := raw.(map[string]any)
		if b["id"] == batchID {
			found = true
			if b["qtyOnHand"].(float64) != 7 {
				t.Fatalf("qty after inv=%v want 7 (drift-safe)", b["qtyOnHand"])
			}
		}
	}
	if !found {
		t.Fatal("batch missing after inventory")
	}
}

func TestPharmacyInventoryQuarantineAdjust(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2777" + suffix[:4], Name: "Inv Q " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	cancelOpenInventorySessions(t, api, tok)

	exp := time.Now().AddDate(0, 0, 200).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "LOT-Q-" + suffix, "expiresOn": exp, "qty": 5,
	})
	if code != http.StatusCreated {
		t.Fatalf("receive %d %#v", code, env)
	}
	recv := dataMap(t, env)
	batchObj, _ := recv["batch"].(map[string]any)
	batchID, _ := batchObj["id"].(string)
	practiceID, _ := batchObj["practiceId"].(string)
	if batchID == "" || practiceID == "" {
		t.Fatalf("missing ids %#v", env)
	}

	if _, err := st.QuarantineBatch(ctx, practiceID, batchID, "", "inventory-test"); err != nil {
		t.Fatalf("quarantine store: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/inventory/sessions", tok, map[string]any{})
	if code != http.StatusCreated {
		t.Fatalf("start %d %#v", code, env)
	}
	sess := dataMap(t, env)
	sessionID, _ := sess["id"].(string)
	lines, _ := sess["lines"].([]any)

	var lineID string
	for _, raw := range lines {
		ln, _ := raw.(map[string]any)
		if ln["batchId"] == batchID {
			lineID, _ = ln["id"].(string)
			break
		}
	}
	if lineID == "" {
		t.Fatal("quarantine batch not in snapshot")
	}

	counts := inventoryCountsForClose(lines, lineID, 3)
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/vet/pharmacy/inventory/sessions/"+sessionID+"/close", tok, map[string]any{"counts": counts})
	if code != http.StatusOK {
		t.Fatalf("close quarantine inv %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/batches", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	for _, raw := range dataMap(t, env)["items"].([]any) {
		row, _ := raw.(map[string]any)
		if row["id"] == batchID {
			if row["qtyOnHand"].(float64) != 3 {
				t.Fatalf("q qty=%v want 3", row["qtyOnHand"])
			}
			return
		}
	}
	t.Fatal("batch not found")
}
