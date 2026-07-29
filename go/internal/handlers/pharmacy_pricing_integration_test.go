package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyPricingAndReorder(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999022", Name: "Price Demo Med", IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/prices/"+medID, tok, map[string]any{
		"purchasePriceCents": 1000, "sellPriceCents": 2500, "vatPercent": 21,
	})
	if code != http.StatusOK {
		t.Fatalf("put price %d %#v", code, env)
	}
	p := dataMap(t, env)
	if int(p["sellPriceCents"].(float64)) != 2500 {
		t.Fatalf("sell=%v", p["sellPriceCents"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/reorder-thresholds", tok, map[string]any{
		"medicationId": medID, "minQty": 3,
	})
	if code != http.StatusOK {
		t.Fatalf("threshold %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/reorder-alerts", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("alerts %d %#v", code, env)
	}
	rows, _ := env["data"].([]any)
	found := false
	for _, row := range rows {
		m, _ := row.(map[string]any)
		if m["medicationId"] == medID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reorder alert for empty stock, got %#v", env)
	}
}
