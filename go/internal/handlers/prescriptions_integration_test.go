package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		"careAdvice":  "Repos et eau a volonte",
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
	if doc["careAdvice"] != "Repos et eau a volonte" {
		t.Fatalf("careAdvice %#v", doc)
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
		"petId": petID, "medications": []any{}, "careAdvice": "repos 48h",
	})
	if code != http.StatusCreated {
		t.Fatalf("empty meds draft want 201 got %d %#v", code, env)
	}
	emptyDoc := dataMap(t, env)
	if emptyDoc["status"] != "draft" {
		t.Fatalf("empty meds status %#v", emptyDoc)
	}
	medsRaw, _ := emptyDoc["medications"].([]any)
	if len(medsRaw) != 0 {
		t.Fatalf("empty meds want [] got %#v", emptyDoc["medications"])
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

func TestPrescriptionsCareAdviceAndVisitLink(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	visitID := createVisitWithOptionalReport(t, api, vetTok, petID,
		"## Medication proposee\nMetacam 1mg SID\n## Plan / suivi\nRepos 48h, eau a volonte",
		"consignes-link", -2*time.Hour, true, true)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId":       petID,
		"visitId":     visitID,
		"careAdvice":  "Repos 48h",
		"medications": []map[string]any{{"name": "Metacam", "posology": "SID"}},
	})
	if code != http.StatusCreated {
		t.Fatalf("create with visit %d %#v", code, env)
	}
	doc := dataMap(t, env)
	if doc["visitId"] != visitID {
		t.Fatalf("visitId %#v", doc)
	}
	if doc["careAdvice"] != "Repos 48h" {
		t.Fatalf("careAdvice %#v", doc)
	}
	rxID, _ := doc["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"careAdvice": "Repos 72h + controle",
	})
	if code != http.StatusOK {
		t.Fatalf("patch careAdvice %d %#v", code, env)
	}
	if dataMap(t, env)["careAdvice"] != "Repos 72h + controle" {
		t.Fatalf("patched %#v", env)
	}

	// mismatch visit/pet
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId":       petID,
		"visitId":     "00000000-0000-4000-8000-000000000099",
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusBadRequest || errCode(env) != "visit_mismatch" {
		t.Fatalf("bad visit want visit_mismatch got %d %q %#v", code, errCode(env), env)
	}
}

func TestPrescriptionsSuggestFromVisitNoReportAndNoGemini(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	// Visit with empty report body → no_visit_report
	visitEmpty := createVisitWithOptionalReport(t, api, vetTok, petID, "", "empty-cr", -5*time.Hour, false, false)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions/suggest-from-visit", vetTok, map[string]any{
		"visitId": visitEmpty,
	})
	if code != http.StatusBadRequest || errCode(env) != "no_visit_report" {
		t.Fatalf("want no_visit_report got %d %q %#v", code, errCode(env), env)
	}

	visitID2 := createVisitWithOptionalReport(t, api, vetTok, petID,
		"## Plan / suivi\nRepos", "suggest-cr", -6*time.Hour, false, false)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions/suggest-from-visit", vetTok, map[string]any{
		"visitId": visitID2,
	})
	// Sans clé / mock : 503 not_configured ; avec clé invalide/réseau : 502 gemini_error ; live OK : 200.
	switch code {
	case http.StatusServiceUnavailable:
		if errCode(env) != "not_configured" {
			t.Fatalf("503 want not_configured got %q %#v", errCode(env), env)
		}
	case http.StatusBadGateway:
		if errCode(env) != "gemini_error" {
			t.Fatalf("502 want gemini_error got %q %#v", errCode(env), env)
		}
	case http.StatusOK:
		data := dataMap(t, env)
		if _, ok := data["careAdvice"]; !ok {
			t.Fatalf("200 missing careAdvice %#v", data)
		}
		if _, ok := data["medications"]; !ok {
			t.Fatalf("200 missing medications %#v", data)
		}
	default:
		t.Fatalf("want 503/502/200 got %d %q %#v", code, errCode(env), env)
	}
}

func demoOtherPetSamePractice(t *testing.T, api *testAPI, excludePetID string) string {
	t.Helper()
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	for _, row := range pets {
		p, _ := row.(map[string]any)
		id, _ := p["id"].(string)
		if id == "" || id == excludePetID {
			continue
		}
		if p["practiceId"] != nil && p["practiceId"] != "" {
			return id
		}
	}
	t.Fatal("need a second demo pet with practice")
	return ""
}

func TestPrescriptionsVisitACLAndClear(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petA := demoPetIDWithPractice(t, api)
	petB := demoOtherPetSamePractice(t, api, petA)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	visitB := createVisitWithOptionalReport(t, api, vetTok, petB,
		"## Plan / suivi\nother pet", "cross-pet", -7*time.Hour, false, false)

	// visit of pet B attached to prescription for pet A → mismatch
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petA, "visitId": visitB,
		"medications": []map[string]any{{"name": "X"}},
	})
	if code != http.StatusBadRequest || errCode(env) != "visit_mismatch" {
		t.Fatalf("cross-pet visit want visit_mismatch got %d %q %#v", code, errCode(env), env)
	}

	visitA := createVisitWithOptionalReport(t, api, vetTok, petA,
		"## Plan / suivi\nok", "clear-visit", -8*time.Hour, false, false)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions", vetTok, map[string]any{
		"petId": petA, "visitId": visitA, "careAdvice": "a",
		"medications": []map[string]any{{"name": "Y"}},
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	rxID, _ := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"visitId": "",
	})
	if code != http.StatusOK {
		t.Fatalf("clear visitId %d %#v", code, env)
	}
	if v := dataMap(t, env)["visitId"]; v != nil && v != "" {
		t.Fatalf("expected cleared visitId %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/prescriptions/"+rxID, vetTok, map[string]any{
		"visitId": visitB,
	})
	if code != http.StatusBadRequest || errCode(env) != "visit_mismatch" {
		t.Fatalf("patch cross-pet want visit_mismatch got %d %q %#v", code, errCode(env), env)
	}
}

func TestPrescriptionsSuggestACL(t *testing.T) {
	t.Setenv("PRESCRIPTIONS_ENABLED", "true")
	api := newTestAPI(t)
	petID := demoPetIDWithPractice(t, api)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	visitID := createVisitWithOptionalReport(t, api, vetTok, petID,
		"## Plan / suivi\nsuggest acl", "suggest-acl", -9*time.Hour, false, false)

	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions/suggest-from-visit", parcTok, map[string]any{
		"visitId": visitID,
	})
	if code != http.StatusForbidden && code != http.StatusNotFound {
		t.Fatalf("cross-cabinet suggest want 403/404 got %d %#v", code, env)
	}

	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/prescriptions/suggest-from-visit", secTok, map[string]any{
		"visitId": visitID,
	})
	if code != http.StatusForbidden {
		t.Fatalf("secretary suggest want 403 got %d %#v", code, env)
	}
}
