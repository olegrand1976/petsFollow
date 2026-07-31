package handlers_test

import (
	"net/http"
	"testing"
)

func TestMePracticePermissionsSeedRoles(t *testing.T) {
	api := newTestAPI(t)

	cases := []struct {
		name, email, password string
		wantCommissions       bool
		wantWriteClinical     bool
		wantSharesManage      bool
	}{
		{"reference_vet", "vet.demo@petsfollow.test", "VetDemo123!", true, true, true},
		{"colleague_vet", "vet.colleague@petsfollow.test", "VetDemo123!", false, true, true},
		{"assistant", "vet.assist@petsfollow.test", "VetDemo123!", false, true, false},
		{"secretary", "secretary.demo@petsfollow.test", "VetDemo123!", false, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok := loginToken(t, api.handler, tc.email, tc.password)
			code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
			if code != http.StatusOK {
				t.Fatalf("me status %d: %#v", code, env)
			}
			me := dataMap(t, env)
			raw, ok := me["practicePermissions"].(map[string]any)
			if !ok || len(raw) == 0 {
				t.Fatalf("expected practicePermissions map, got %#v", me["practicePermissions"])
			}
			if got, _ := raw["commissions.view"].(bool); got != tc.wantCommissions {
				t.Fatalf("commissions.view: want %v got %v (%#v)", tc.wantCommissions, got, raw)
			}
			if got, _ := raw["pets.write_clinical"].(bool); got != tc.wantWriteClinical {
				t.Fatalf("pets.write_clinical: want %v got %v (%#v)", tc.wantWriteClinical, got, raw)
			}
			if got, _ := raw["clients.read"].(bool); !got {
				t.Fatalf("clients.read expected true: %#v", raw)
			}
			if got, _ := raw["shares.read"].(bool); !got {
				t.Fatalf("shares.read expected true: %#v", raw)
			}
			if got, _ := raw["shares.manage"].(bool); got != tc.wantSharesManage {
				t.Fatalf("shares.manage: want %v got %v (%#v)", tc.wantSharesManage, got, raw)
			}
			if got, _ := raw["pharmacy.read"].(bool); !got {
				t.Fatalf("pharmacy.read expected true: %#v", raw)
			}
			if got, _ := raw["pharmacy.write"].(bool); got != tc.wantWriteClinical {
				t.Fatalf("pharmacy.write: want %v got %v (%#v)", tc.wantWriteClinical, got, raw)
			}
		})
	}
}

func TestMePracticePermissionsAbsentForNonStaff(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me status %d: %#v", code, env)
	}
	me := dataMap(t, env)
	if _, ok := me["practicePermissions"]; ok {
		t.Fatalf("commercial must not expose practicePermissions: %#v", me["practicePermissions"])
	}
}
