package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyVamregRefSyncDryRunAndApply(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/target-species"):
			_ = json.NewEncoder(w).Encode([]pharmacy.VamregCodeLabel{
				{Code: "CAN", Fr: "Chien", En: "Dog"},
				{Code: "CAN", Fr: "dup"},
			})
		case strings.HasSuffix(r.URL.Path, "/indication"):
			_ = json.NewEncoder(w).Encode([]pharmacy.VamregCodeLabel{{Code: "INF", Fr: "Infection"}})
		case strings.HasSuffix(r.URL.Path, "/pharmaceutical-form"):
			_ = json.NewEncoder(w).Encode([]pharmacy.VamregCodeLabel{{Code: "TAB", Fr: "Comprimé"}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PHARMACY_VAMREG_REF_SYNC_SECRET", "test-vamreg-ref-sync")
	t.Setenv("VAMREG_AFMPS_API_KEY", "test-key")
	t.Setenv("VAMREG_AFMPS_BASE_URL", srv.URL)

	api := newTestAPI(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/pharmacy/vamreg-ref-sync", strings.NewReader(`{"dryRun":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-Vamreg-Ref-Sync-Secret", "test-vamreg-ref-sync")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dry-run %d %s", rec.Code, rec.Body.String())
	}
	var env map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	data, _ := env["data"].(map[string]any)
	fetched, _ := data["fetched"].(map[string]any)
	if int(fetched["target_species"].(float64)) != 2 {
		t.Fatalf("fetched %#v", fetched)
	}
	if data["dryRun"] != true {
		t.Fatalf("dryRun %#v", data)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/internal/pharmacy/vamreg-ref-sync", strings.NewReader(`{"dryRun":false}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-Vamreg-Ref-Sync-Secret", "test-vamreg-ref-sync")
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply %d %s", rec.Code, rec.Body.String())
	}

	st := store.New(api.pool)
	for _, kind := range []string{
		store.VamregRefKindTargetSpecies,
		store.VamregRefKindIndication,
		store.VamregRefKindPharmaceuticalForm,
	} {
		items, err := st.ListVamregRefCodes(context.Background(), kind, false, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 {
			t.Fatalf("%s %#v", kind, items)
		}
	}
	species, _ := st.ListVamregRefCodes(context.Background(), store.VamregRefKindTargetSpecies, false, 10)
	if species[0].Code != "CAN" {
		t.Fatalf("%#v", species)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, listEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/vamreg-refs?kind=target_species", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, listEnv)
	}
	got := dataMap(t, listEnv)
	arr, _ := got["items"].([]any)
	if len(arr) != 1 {
		t.Fatalf("%#v", got)
	}

	code, bad := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/vamreg-refs?kind=nope", tok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("bad kind %d %#v", code, bad)
	}
}

func TestPharmacyVamregRefSyncRejectsEmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]pharmacy.VamregCodeLabel{})
	}))
	t.Cleanup(srv.Close)

	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PHARMACY_VAMREG_REF_SYNC_SECRET", "test-vamreg-ref-sync")
	t.Setenv("VAMREG_AFMPS_API_KEY", "test-key")
	t.Setenv("VAMREG_AFMPS_BASE_URL", srv.URL)

	api := newTestAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/pharmacy/vamreg-ref-sync", strings.NewReader(`{"dryRun":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pharmacy-Vamreg-Ref-Sync-Secret", "test-vamreg-ref-sync")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestPharmacyVamregRefSyncUnauthorized(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("PHARMACY_VAMREG_REF_SYNC_SECRET", "secret")
	api := newTestAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/pharmacy/vamreg-ref-sync", strings.NewReader(`{}`))
	req.Header.Set("X-Pharmacy-Vamreg-Ref-Sync-Secret", "wrong")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("%d", rec.Code)
	}
}
