package handlers_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestVetProfileHeartRateDurationsPersistAndOmit(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("e2e-hr-profile")
	password := "TestPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": password, "fullName": "Dr HR", "practiceName": "Cabinet HR",
	})
	if code != http.StatusCreated {
		t.Fatalf("register status %d: %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm status %d: %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		t.Fatal("expected accessToken after confirm")
	}

	profileBody := map[string]any{
		"vetFullName":            "Dr HR",
		"practiceName":           "Cabinet HR",
		"contactEmail":           email,
		"phone":                  "+32123456789",
		"addressLine1":           "Rue Test 1",
		"city":                   "Bruxelles",
		"postalCode":             "1000",
		"heartrateDurationsSec":  []int{15, 30},
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", access, profileBody)
	if code != http.StatusOK {
		t.Fatalf("put profile status %d: %#v", code, env)
	}
	got := durationInts(t, dataMap(t, env)["heartrateDurationsSec"])
	if !reflect.DeepEqual(got, []int{15, 30}) {
		t.Fatalf("after put with durations: got %#v", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/profile", access, nil)
	if code != http.StatusOK {
		t.Fatalf("get profile status %d: %#v", code, env)
	}
	got = durationInts(t, dataMap(t, env)["heartrateDurationsSec"])
	if !reflect.DeepEqual(got, []int{15, 30}) {
		t.Fatalf("after get: got %#v", got)
	}

	omitBody := map[string]any{
		"vetFullName":  "Dr HR Updated",
		"practiceName": "Cabinet HR",
		"contactEmail": email,
		"phone":        "+32123456789",
		"addressLine1": "Rue Test 1",
		"city":         "Bruxelles",
		"postalCode":   "1000",
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/profile", access, omitBody)
	if code != http.StatusOK {
		t.Fatalf("put profile omit durations status %d: %#v", code, env)
	}
	data := dataMap(t, env)
	if data["vetFullName"] != "Dr HR Updated" {
		t.Fatalf("expected name update, got %#v", data["vetFullName"])
	}
	got = durationInts(t, data["heartrateDurationsSec"])
	if !reflect.DeepEqual(got, []int{15, 30}) {
		t.Fatalf("omit must preserve durations, got %#v", got)
	}
}

func durationInts(t *testing.T, raw any) []int {
	t.Helper()
	arr, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected duration array, got %#v", raw)
	}
	out := make([]int, 0, len(arr))
	for _, v := range arr {
		f, ok := v.(float64)
		if !ok {
			t.Fatalf("expected number in durations, got %#v", v)
		}
		out = append(out, int(f))
	}
	return out
}
