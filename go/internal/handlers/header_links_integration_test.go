package handlers_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestHeaderLinks_DefaultsAndCustom(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("header-links")
	password := "HeaderLinks1!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Links", "practiceName": "Cabinet Links",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		t.Fatal("expected accessToken")
	}

	body := map[string]any{
		"practiceName": "Cabinet Links",
		"phone":        "0470000000",
		"contactEmail": email,
		"addressLine1": "1 rue Test",
		"city":         "Bruxelles",
		"postalCode":   "1000",
		"countryCode":  "BE",
		"vetFullName":  "Dr Links",
		"website":      "",
		"headerLinks": map[string]any{
			"configured": true,
			"enabled":    []string{"be_vetcompendium", "be_ordre"},
			"order":      []string{"be_vetcompendium", "custom_lab1", "be_ordre"},
			"custom": []map[string]any{
				{"id": "custom_lab1", "label": "Lab XYZ", "url": "https://lab.example/"},
			},
		},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", access, body)
	if code != http.StatusOK {
		t.Fatalf("put profile %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/header-links", access, nil)
	if code != http.StatusOK {
		t.Fatalf("header-links %d %#v", code, env)
	}
	data := dataMap(t, env)
	items, _ := data["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %#v", items)
	}
	first, _ := items[0].(map[string]any)
	if first["id"] != "be_vetcompendium" {
		t.Fatalf("first %#v", first)
	}
	second, _ := items[1].(map[string]any)
	if second["kind"] != "custom" || second["label"] != "Lab XYZ" {
		t.Fatalf("second %#v", second)
	}

	body["headerLinks"] = map[string]any{
		"configured": true,
		"enabled":    []string{"be_ordre"},
		"custom": []map[string]any{
			{"label": "Bad", "url": "http://insecure.example/"},
		},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", access, body)
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 for http url, got %d %#v", code, env)
	}

	body["headerLinks"] = map[string]any{
		"configured": true,
		"enabled":    []string{"fr_ordre"},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", access, body)
	if code != http.StatusBadRequest {
		t.Fatalf("want 400 for foreign catalog id, got %d %#v", code, env)
	}
}

func TestHeaderLinks_DefaultsBE(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("header-defaults")
	password := "HeaderLinks1!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr Def", "practiceName": "Cabinet Def",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		t.Fatal("expected accessToken")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/header-links", access, nil)
	if code != http.StatusOK {
		t.Fatalf("header-links %d %#v", code, env)
	}
	data := dataMap(t, env)
	items, _ := data["items"].([]any)
	if len(items) == 0 {
		t.Fatal("expected default BE catalog items")
	}
	catalog, _ := data["catalog"].([]any)
	if len(catalog) < 5 {
		t.Fatalf("expected BE catalog, got %#v", catalog)
	}
}
