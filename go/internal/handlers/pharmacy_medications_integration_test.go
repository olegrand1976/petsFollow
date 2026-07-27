package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyMedicationSearch(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()

	st := store.New(api.pool)
	if _, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK:                "2712345",
		Name:               "Amoxicilline Vet Demo",
		ATCCode:            "J01CA04",
		IsAntibiotic:       true,
		IsActive:           true,
		PharmaceuticalForm: "cp",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if _, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK:          "2899999",
		Name:         "Vaccin Rage Demo",
		IsAntibiotic: false,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("upsert2: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/search?q=amoxi", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("search %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected hits %#v", env)
	}
	first, _ := items[0].(map[string]any)
	if first["cnk"] != "2712345" {
		t.Fatalf("unexpected first %#v", first)
	}
	if first["isAntibiotic"] != true {
		t.Fatalf("expected antibiotic badge field %#v", first)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/search?q=271", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("cnk search %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	items, _ = data["items"].([]any)
	found := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["cnk"] == "2712345" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cnk 2712345 in results %#v", env)
	}

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/search?q=amoxi", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("client should be forbidden, got %d", code)
	}
}

func TestPharmacyMedicationSearchDisabled(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "false")
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/search?q=amoxi", tok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 when disabled, got %d %#v", code, env)
	}
}
