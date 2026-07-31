package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrescriptionsDraftCRUDAndPDF(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no demo pets")
	}
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

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId":       petID,
		"paperFormat": "A5",
		"notes":       "demo note",
		"validUntil":  "2030-12-31",
		"medications": []map[string]any{{
			"name": "Amoxicilline", "dosage": "50mg", "form": "tablet",
			"quantity": "1 box", "posology": "1x/day 7d", "withdrawal_period": "",
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	doc := dataMap(t, env)
	rxID, _ := doc["id"].(string)
	if rxID == "" || doc["status"] != "draft" {
		t.Fatalf("unexpected create %#v", doc)
	}
	if doc["paperFormat"] != "A5" {
		t.Fatalf("format %#v", doc)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/prescriptions?status=draft&petId="+petID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	items, _ := dataMap(t, env)["items"].([]any)
	found := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["id"] == rxID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("draft not listed %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"notes": "updated",
		"medications": []map[string]any{{
			"name": "Metacam", "dosage": "1mg", "form": "liquid",
			"quantity": "1", "posology": "SID",
		}},
	})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	if dataMap(t, env)["notes"] != "updated" {
		t.Fatalf("notes %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"status": "signed",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("signed status want 400 got %d %#v", code, env)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vet/prescriptions/"+rxID+"/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pdf %d %s", rec.Code, rec.Body.String())
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) {
		t.Fatalf("not a pdf")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/vet/prescriptions/"+rxID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("delete %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/prescriptions/"+rxID, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("after delete want 404 got %d %#v", code, env)
	}
}

func TestPrescriptionsDisabled(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "false")
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/prescriptions", tok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("want 404 got %d %#v", code, env)
	}
	if errCode(env) != "prescriptions_disabled" {
		t.Fatalf("want prescriptions_disabled got %q %#v", errCode(env), env)
	}
}

func demoPetIDWithPractice(t *testing.T, api *testAPI) string {
	t.Helper()
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	for _, row := range pets {
		p, _ := row.(map[string]any)
		if p["practiceId"] != nil && p["practiceId"] != "" {
			id, _ := p["id"].(string)
			if id != "" {
				return id
			}
		}
	}
	t.Fatal("no pet with practice")
	return ""
}

func TestPrescriptionsValidation(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID, "medications": []any{},
	})
	if code != http.StatusBadRequest || errCode(env) != "invalid_medications" {
		t.Fatalf("empty meds want 400 invalid_medications got %d %q %#v", code, errCode(env), env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID, "paperFormat": "A3",
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusBadRequest || errCode(env) != "invalid_format" {
		t.Fatalf("bad format want 400 invalid_format got %d %q %#v", code, errCode(env), env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID, "validUntil": "2031-06-01",
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	rxID, _ := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"validUntil": "",
	})
	if code != http.StatusOK {
		t.Fatalf("clear validUntil %d %#v", code, env)
	}
	if dataMap(t, env)["validUntil"] != nil {
		t.Fatalf("expected cleared validUntil %#v", env)
	}
}

func TestPrescriptionsCrossCabinetForbidden(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api) // VetPlus / client.demo
	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", parcTok, map[string]any{
		"petId":       petID,
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusForbidden && code != http.StatusNotFound {
		// forbidden (CanAccessPet) or not found depending on ACL path
		t.Fatalf("cross-cabinet want 403/404 got %d %#v", code, env)
	}
}

func TestPrescriptionsSecretaryForbidden(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api)
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", secTok, map[string]any{
		"petId":       petID,
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusForbidden {
		t.Fatalf("secretary want 403 got %d %#v", code, env)
	}
}

func TestPrescriptionsInvalidUUIDAndPayload(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/prescriptions/not-a-uuid", vetTok, nil)
	if code != http.StatusBadRequest || errCode(env) != "validation_error" {
		t.Fatalf("bad id want 400 validation_error got %d %q %#v", code, errCode(env), env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": "not-a-uuid", "medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusBadRequest || errCode(env) != "validation_error" {
		t.Fatalf("bad petId want 400 validation_error got %d %q %#v", code, errCode(env), env)
	}

	petID := demoPetIDWithPractice(t, api)
	meds := make([]map[string]any, 51)
	for i := range meds {
		meds[i] = map[string]any{"name": "X"}
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petID, "medications": meds,
	})
	if code != http.StatusBadRequest || errCode(env) != "payload_too_large" {
		t.Fatalf("too many meds want 400 payload_too_large got %d %q %#v", code, errCode(env), env)
	}
}
