package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyStockReceiptAndWaste(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PHARMACY_EXPIRY_SECRET", "test-pharmacy-secret")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2888002", Name: "Stock API Demo", IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/deposits", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("deposits %d %#v", code, env)
	}

	exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID,
		"lotNumber":    "LOT-API-1",
		"expiresOn":    exp,
		"qty":          10,
		"unit":         "box",
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	batch, _ := data["batch"].(map[string]any)
	batchID, _ := batch["id"].(string)
	if batchID == "" {
		t.Fatalf("no batch id %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID,
		"lotNumber":    "LOT-OLD",
		"expiresOn":    time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
		"qty":          1,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("expired receipt want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/expiry/summary", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("summary %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches/"+batchID+"/waste", tok, map[string]any{
		"reason": "expired",
		"qty":    2,
	})
	if code != http.StatusOK {
		t.Fatalf("waste %d %#v", code, env)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/pharmacy/expiry-run", bytes.NewBufferString(`{"forceDigest":false}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-Expiry-Secret", "test-pharmacy-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expiry-run %d %s", rec.Code, rec.Body.String())
	}
}

func TestPharmacyStockDisabled(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "false")
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, _ := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/batches", tok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", code)
	}
}
