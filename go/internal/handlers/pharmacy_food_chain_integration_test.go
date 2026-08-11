package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyFoodChainAndWithdrawal(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	meat, milk, eggs := 28, 7, 0
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2666" + suffix[:4], Name: "FoodChain Med " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
		"withdrawalMeatDays": meat, "withdrawalMilkDays": milk, "withdrawalEggsDays": eggs,
		"foodChainBanned": false,
	})
	if code != http.StatusOK {
		t.Fatalf("withdrawal %d %#v", code, env)
	}
	// Overlay must not mutate national ref_medications.
	base, err := st.GetRefMedication(ctx, medID)
	if err != nil {
		t.Fatalf("base med: %v", err)
	}
	if base.WithdrawalMeatDays != nil || base.FoodChainBanned {
		t.Fatalf("national ref mutated by practice PATCH: %#v", base)
	}
	meCode, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if meCode != http.StatusOK {
		t.Fatalf("me %d %#v", meCode, meEnv)
	}
	practiceID, _ := dataMap(t, meEnv)["practiceId"].(string)
	merged, err := st.GetRefMedicationForPractice(ctx, practiceID, medID)
	if err != nil {
		t.Fatalf("merged: %v", err)
	}
	if merged.WithdrawalMeatDays == nil || *merged.WithdrawalMeatDays != meat {
		t.Fatalf("practice overlay missing meat days %#v", merged)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/medications/"+medID, tok, nil)
	if code != http.StatusOK {
		t.Fatalf("get medication %d %#v", code, env)
	}
	gotMed := dataMap(t, env)
	if gotMed["cnk"] == nil || gotMed["id"] != medID {
		t.Fatalf("get medication payload %#v", gotMed)
	}
	if gotMed["withdrawalMeatDays"] == nil || int(gotMed["withdrawalMeatDays"].(float64)) != meat {
		t.Fatalf("GET medication missing overlay %#v", gotMed)
	}

	// Spirit (seed horse) — set food_producing
	petsCode, petsEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if petsCode != http.StatusOK {
		t.Fatalf("pets %d %#v", petsCode, petsEnv)
	}
	var petID, clientID string
	items, _ := petsEnv["data"].([]any)
	for _, raw := range items {
		p, _ := raw.(map[string]any)
		if p["name"] == "Spirit" || p["species"] == "horse" {
			petID, _ = p["id"].(string)
			clientID, _ = p["ownerUserId"].(string)
			break
		}
	}
	if petID == "" {
		t.Skip("no horse pet in seed for food-chain test")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/food-chain", tok, map[string]any{
		"foodChainStatus":  "food_producing",
		"domicileLocation": "Écurie Test DAF — Bruxelles",
	})
	if code != http.StatusOK {
		t.Fatalf("food-chain %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["foodChainStatus"] != "food_producing" {
		t.Fatalf("status %#v", got)
	}
	if got["domicileLocation"] != "Écurie Test DAF — Bruxelles" {
		t.Fatalf("domicile %#v", got)
	}
	getCode, getEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, tok, nil)
	if getCode != http.StatusOK {
		t.Fatalf("get pet %d %#v", getCode, getEnv)
	}
	petData := dataMap(t, getEnv)
	if petData["foodChainStatus"] != "food_producing" || petData["domicileLocation"] != "Écurie Test DAF — Bruxelles" {
		t.Fatalf("pet payload missing regulatory fields %#v", petData)
	}

	exp := time.Now().AddDate(0, 0, 200).Format("2006-01-02")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID, "lotNumber": "LOT-FC-" + suffix, "expiresOn": exp, "qty": 5,
	})
	if code != http.StatusCreated {
		t.Fatalf("receive %d %#v", code, env)
	}

	// Clear withdrawal on med → draft still snapshots at insert; create draft AFTER clearing would fail finalize
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
		"withdrawalMeatDays": nil, "withdrawalMilkDays": nil, "withdrawalEggsDays": nil,
		"foodChainBanned": false,
	})
	if code != http.StatusOK {
		t.Fatalf("clear withdrawal %d %#v", code, env)
	}

	body := map[string]any{
		"petId": petID, "clientUserId": clientID, "notes": "fc-test",
		"items": []map[string]any{{
			"medicationId": medID, "qty": 1, "ammNumber": "BE-TEST-AMM",
		}},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
	if code != http.StatusCreated {
		// clientUserId may be required differently
		t.Fatalf("daf draft %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("finalize without withdrawal want 400 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "food_chain_withdrawal_required" {
		t.Fatalf("code %#v", env)
	}

	// Restore withdrawal + new draft
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
		"withdrawalMeatDays": meat, "withdrawalMilkDays": milk, "withdrawalEggsDays": eggs,
	})
	if code != http.StatusOK {
		t.Fatalf("restore withdrawal %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
	if code != http.StatusCreated {
		t.Fatalf("daf draft2 %d %#v", code, env)
	}
	dafID, _ = dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("finalize ok %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if nested, ok := doc["daf"].(map[string]any); ok {
		doc = nested
	}
	items2, _ := doc["items"].([]any)
	if len(items2) == 0 {
		// reload
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/daf/"+dafID, tok, nil)
		if code != http.StatusOK {
			t.Fatalf("get daf %d %#v", code, env)
		}
		doc = dataMap(t, env)
		items2, _ = doc["items"].([]any)
	}
	if len(items2) == 0 {
		t.Fatalf("no items %#v", doc)
	}
	it, _ := items2[0].(map[string]any)
	if it["withdrawalMeatDays"] == nil {
		t.Fatalf("expected withdrawal snapshot %#v", it)
	}

	// Partial PATCH: meat only must preserve ban + milk/eggs.
	banned := true
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
		"withdrawalMeatDays": meat, "withdrawalMilkDays": milk, "withdrawalEggsDays": eggs,
		"foodChainBanned": banned,
	})
	if code != http.StatusOK {
		t.Fatalf("set ban+days %d %#v", code, env)
	}
	newMeat := 40
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
		"withdrawalMeatDays": newMeat,
	})
	if code != http.StatusOK {
		t.Fatalf("partial meat patch %d %#v", code, env)
	}
	partial := dataMap(t, env)
	if int(partial["withdrawalMeatDays"].(float64)) != newMeat {
		t.Fatalf("meat not updated %#v", partial)
	}
	if int(partial["withdrawalMilkDays"].(float64)) != milk {
		t.Fatalf("milk wiped on partial patch %#v", partial)
	}
	if partial["foodChainBanned"] != true {
		t.Fatalf("ban wiped on partial patch %#v", partial)
	}

	// Reset pet to companion for other tests
	doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/food-chain", tok, map[string]any{
		"foodChainStatus": "companion",
	})
}
