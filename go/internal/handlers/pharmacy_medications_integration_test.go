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

func TestPharmacyMedicationListByLetter(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	meta := []byte(`{"source":"afmps-pack-csv","manufacturer":"Demo Labs","active_substance":"amoxicillin"}`)
	idA, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK:                "2711111",
		Name:               "Amoxicilline Letter Demo",
		ATCCode:            "J01CA04",
		IsAntibiotic:       true,
		IsActive:           true,
		PharmaceuticalForm: "cp",
		AFMPSMeta:          meta,
	})
	if err != nil {
		t.Fatalf("upsert A: %v", err)
	}
	if _, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK:      "2799990",
		Name:     "9Numeric Med Demo",
		IsActive: true,
	}); err != nil {
		t.Fatalf("upsert #: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications?letter=A&limit=50&offset=0", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list A %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	items, _ := data["items"].([]any)
	found := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["cnk"] == "2711111" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Amoxicilline under letter A %#v", env)
	}
	if data["letter"] != "A" {
		t.Fatalf("letter field %#v", data["letter"])
	}
	counts, _ := data["letterCounts"].(map[string]any)
	if counts == nil {
		t.Fatalf("missing letterCounts %#v", data)
	}
	if toFloat(data["catalogTotal"]) < 1 {
		t.Fatalf("catalogTotal %#v", data["catalogTotal"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications?letter=%23", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list # %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	items, _ = data["items"].([]any)
	foundHash := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["cnk"] == "2799990" {
			foundHash = true
			break
		}
	}
	if !foundHash {
		t.Fatalf("expected numeric name under # %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications?letter=A&limit=1&offset=0", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("page %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	items, _ = data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("limit=1 got %d items %#v", len(items), env)
	}
	if toFloat(data["total"]) < 1 {
		t.Fatalf("total %#v", data["total"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications?letter=ZZ", tok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid letter want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/"+idA, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	detail, _ := env["data"].(map[string]any)
	if detail["afmpsSource"] != "afmps-pack-csv" {
		t.Fatalf("afmpsSource %#v", detail)
	}
	metaObj, _ := detail["afmpsMeta"].(map[string]any)
	if metaObj["manufacturer"] != "Demo Labs" {
		t.Fatalf("afmpsMeta %#v", detail["afmpsMeta"])
	}

	// List rows expose afmpsSource; pagination can skip letterCounts.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications?letter=A&includeCounts=0", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list no counts %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	if _, ok := data["letterCounts"]; ok {
		t.Fatalf("letterCounts should be omitted when includeCounts=0 %#v", data)
	}
	items, _ = data["items"].([]any)
	foundSrc := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["cnk"] == "2711111" {
			if m["afmpsSource"] != "afmps-pack-csv" {
				t.Fatalf("list afmpsSource %#v", m)
			}
			foundSrc = true
			break
		}
	}
	if !foundSrc {
		t.Fatalf("expected list row with afmpsSource %#v", env)
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}
