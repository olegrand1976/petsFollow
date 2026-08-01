package handlers_test

import (
	"net/http"
	"testing"
)

func TestLabPanelCRUDAndTimeline(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/lab-panels", clientTok, map[string]any{
		"labName": "LabDemo",
		"results": []map[string]any{{"analyteCode": "crea", "valueNum": 1.2, "unit": "mg/dL"}},
	})
	if code != http.StatusForbidden {
		t.Fatalf("client create want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/lab-panels", vetTok, map[string]any{
		"labName": "BioVet",
		"notes":   "contrôle rénal",
		"results": []map[string]any{
			{"analyteCode": "crea", "valueNum": 2.5, "unit": "mg/dL", "refLow": 0.5, "refHigh": 1.5},
			{"analyteCode": "alat", "valueNum": 40, "unit": "U/L", "refLow": 10, "refHigh": 100},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create panel %d %#v", code, env)
	}
	created := dataMap(t, env)
	panelID, _ := created["id"].(string)
	if panelID == "" {
		t.Fatalf("missing panel id %#v", created)
	}
	if ab, _ := created["abnormalCount"].(float64); ab != 1 {
		t.Fatalf("abnormalCount=%v want 1", created["abnormalCount"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/lab-panels", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list client %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	if len(list) == 0 {
		t.Fatalf("client should list panels")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/lab-panels/"+panelID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get panel %d %#v", code, env)
	}
	detail := dataMap(t, env)
	results, _ := detail["results"].([]any)
	if len(results) != 2 {
		t.Fatalf("results=%d want 2", len(results))
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/lab-analytes/crea/trend", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("trend %d %#v", code, env)
	}
	trend, _ := env["data"].([]any)
	if len(trend) != 1 {
		t.Fatalf("trend points=%d want 1", len(trend))
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/pets/"+petID+"/lab-panels/"+panelID, vetTok, map[string]any{
		"results": []map[string]any{
			{"analyteCode": "crea", "valueNum": 1.0, "unit": "mg/dL", "refLow": 0.5, "refHigh": 1.5},
		},
	})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	if ab, _ := dataMap(t, env)["abnormalCount"].(float64); ab != 0 {
		t.Fatalf("abnormal after patch=%v", dataMap(t, env)["abnormalCount"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("timeline %d %#v", code, env)
	}
	found := false
	for _, raw := range env["data"].([]any) {
		item, _ := raw.(map[string]any)
		if item["type"] == "lab_panel" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("timeline missing lab_panel")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/pets/"+petID+"/lab-panels/"+panelID, vetTok, map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("patch without results want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/lab-panels", vetTok, map[string]any{
		"labName": "Empty",
		"results": []map[string]any{},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("empty panel want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/"+petID+"/lab-panels/"+panelID, "", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("delete unauth want 401 got %d %#v", code, env)
	}

	farrierTok := loginToken(t, api.handler, "farrier.demo@petsfollow.test", "CareProDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/lab-panels", farrierTok, map[string]any{
		"labName": "CarePro",
		"results": []map[string]any{{"analyteCode": "crea", "valueNum": 1.0, "unit": "mg/dL"}},
	})
	if code != http.StatusForbidden {
		t.Fatalf("care_pro create want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/"+petID+"/lab-panels/"+panelID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("delete %d %#v", code, env)
	}
}
