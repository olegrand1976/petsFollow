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
}
