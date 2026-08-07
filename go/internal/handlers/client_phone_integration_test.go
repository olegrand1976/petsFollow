package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestClientContactPhoneCreateListGetPatch(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.phone.%s@petsfollow.test", uuid.NewString()[:8])
	phone := "0470 11 22 33"

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Phone Test", "contactPhone": phone,
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["contactPhone"] != phone {
		t.Fatalf("get contactPhone=%v want %q", got["contactPhone"], phone)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	found := false
	for _, raw := range list {
		row, _ := raw.(map[string]any)
		if row["userId"] == clientID {
			found = true
			if row["contactPhone"] != phone {
				t.Fatalf("list contactPhone=%v want %q", row["contactPhone"], phone)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created client not in list")
	}

	updated := "0499 00 11 22"
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": updated,
	})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	if dataMap(t, env)["contactPhone"] != updated {
		t.Fatalf("patch contactPhone=%v want %q", dataMap(t, env)["contactPhone"], updated)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": "",
	})
	if code != http.StatusOK {
		t.Fatalf("clear %d %#v", code, env)
	}
	cleared := dataMap(t, env)["contactPhone"]
	if cleared != nil && cleared != "" {
		t.Fatalf("clear contactPhone=%v want empty", cleared)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/clients/00000000-0000-0000-0000-000000000099", vetTok, map[string]any{
			"contactPhone": "0470 00 00 99",
		})
	if code != http.StatusNotFound {
		t.Fatalf("unknown client want 404 got %d %#v", code, env)
	}

	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, parcTok, map[string]any{
		"contactPhone": "0470 99 99 99",
	})
	if code != http.StatusNotFound {
		t.Fatalf("other practice want 404 got %d %#v", code, env)
	}

	tooLong := strings.Repeat("1", 41)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": tooLong,
	})
	if code != http.StatusBadRequest || errorMsgKey(env) != "contact_phone_too_long" {
		t.Fatalf("too long want 400 contact_phone_too_long got %d %#v", code, env)
	}
}

func TestClientContactPhoneCreateOptionalAndTooLong(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	emailNoPhone := fmt.Sprintf("client.nophone.%s@petsfollow.test", uuid.NewString()[:8])
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": emailNoPhone, "password": "TempPass12!", "fullName": "No Phone",
	})
	if code != http.StatusCreated {
		t.Fatalf("create without phone %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	if phone := dataMap(t, env)["contactPhone"]; phone != nil && phone != "" {
		t.Fatalf("expected empty contactPhone, got %v", phone)
	}

	emailLong := fmt.Sprintf("client.longphone.%s@petsfollow.test", uuid.NewString()[:8])
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": emailLong, "password": "TempPass12!", "fullName": "Long Phone",
		"contactPhone": strings.Repeat("1", 41),
	})
	if code != http.StatusBadRequest || errorMsgKey(env) != "contact_phone_too_long" {
		t.Fatalf("create too long want 400 contact_phone_too_long got %d %#v", code, env)
	}
}

func TestClientContactPhonePatchACL(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.phone.acl.%s@petsfollow.test", uuid.NewString()[:8])
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Phone ACL", "contactPhone": "0470 00 00 10",
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, map[string]any{
		"contactPhone": "0470 00 00 11",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("unauth want 401 got %d %#v", code, env)
	}

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, clientTok, map[string]any{
		"contactPhone": "0470 00 00 12",
	})
	if code != http.StatusForbidden {
		t.Fatalf("client role want 403 got %d %#v", code, env)
	}

	commTok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, commTok, map[string]any{
		"contactPhone": "0470 00 00 13",
	})
	if code != http.StatusForbidden {
		t.Fatalf("commercial want 403 got %d %#v", code, env)
	}
}

// TestClientContactPhoneAccountGlobalLastWriteWins documents the product choice:
// contact_phone lives on identity.users (account-global). Any linked practice with
// clients.write can update it; other linked practices see the same value (last write wins).
// Unlinked practices still get 404 (see TestClientContactPhoneCreateListGetPatch).
func TestClientContactPhoneAccountGlobalLastWriteWins(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	parcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")

	email := fmt.Sprintf("client.phone.shared.%s@petsfollow.test", uuid.NewString()[:8])
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Shared Phone", "contactPhone": "0470 11 11 11",
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)

	var parcID, parcPracticeID string
	if err := api.pool.QueryRow(ctx, `
		SELECT u.id::text, u.practice_id::text FROM identity.users u WHERE u.email='vet.parc@petsfollow.test'`).Scan(&parcID, &parcPracticeID); err != nil {
		t.Fatalf("parc: %v", err)
	}

	// Link same client to Clinique du Parc (multi-cabinet household).
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), parcPracticeID, clientID, parcID); err != nil {
		t.Fatalf("link parc: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, parcTok, map[string]any{
		"contactPhone": "0470 22 22 22",
	})
	if code != http.StatusOK {
		t.Fatalf("parc patch %d %#v", code, env)
	}
	if dataMap(t, env)["contactPhone"] != "0470 22 22 22" {
		t.Fatalf("parc patch phone=%v", dataMap(t, env)["contactPhone"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vetplus get %d %#v", code, env)
	}
	if dataMap(t, env)["contactPhone"] != "0470 22 22 22" {
		t.Fatalf("account-global: VetPlus should see Parc write, got %v", dataMap(t, env)["contactPhone"])
	}
}

// TestClientContactPhonePatchIgnoresActiveRole: multi-profile user linked as client can be
// patched even if identity.users.role is temporarily not 'client'.
func TestClientContactPhonePatchIgnoresActiveRole(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.phone.role.%s@petsfollow.test", uuid.NewString()[:8])
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email": email, "password": "TempPass12!", "fullName": "Role Flip", "contactPhone": "0470 00 00 20",
	})
	if code != http.StatusCreated {
		t.Fatalf("create %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)

	// Simulate active profile switch: users.role no longer 'client' while practice_clients remains.
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE identity.users SET role = 'vet' WHERE id = $1`, clientID); err != nil {
		t.Fatalf("flip role: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"contactPhone": "0470 00 00 21",
	})
	if code != http.StatusOK {
		t.Fatalf("patch with flipped role want 200 got %d %#v", code, env)
	}
	if dataMap(t, env)["contactPhone"] != "0470 00 00 21" {
		t.Fatalf("contactPhone=%v", dataMap(t, env)["contactPhone"])
	}
}

func TestClientIdentityCreateWithoutPasswordAndPatch(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.identity.%s@petsfollow.test", uuid.NewString()[:8])

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email":                  email,
		"firstName":              "Alice",
		"lastName":               "Dupont",
		"contactPhone":           "0470 55 66 77",
		"address":                "12 rue des Lilas, 1000 Bruxelles",
		"nationalRegistryNumber": "85.07.30-123.45",
	})
	if code != http.StatusCreated {
		t.Fatalf("create without password %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["firstName"] != "Alice" || got["lastName"] != "Dupont" {
		t.Fatalf("names=%v %v", got["firstName"], got["lastName"])
	}
	if got["fullName"] != "Alice Dupont" {
		t.Fatalf("fullName=%v", got["fullName"])
	}
	if got["address"] != "12 rue des Lilas, 1000 Bruxelles" {
		t.Fatalf("address=%v", got["address"])
	}
	if got["nationalRegistryNumber"] != "85.07.30-123.45" {
		t.Fatalf("niss=%v", got["nationalRegistryNumber"])
	}
	if got["contactPhone"] != "0470 55 66 77" {
		t.Fatalf("phone=%v", got["contactPhone"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"firstName":              "Alicia",
		"lastName":               "Martin",
		"address":                "1 av. Louise",
		"nationalRegistryNumber": "90010112345",
		"contactPhone":           "0499 11 22 33",
	})
	if code != http.StatusOK {
		t.Fatalf("patch identity %d %#v", code, env)
	}
	got = dataMap(t, env)
	if got["fullName"] != "Alicia Martin" || got["firstName"] != "Alicia" || got["lastName"] != "Martin" {
		t.Fatalf("patched names %#v", got)
	}
	if got["address"] != "1 av. Louise" || got["nationalRegistryNumber"] != "90010112345" {
		t.Fatalf("patched address/niss %#v", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email":                  fmt.Sprintf("client.badniss.%s@petsfollow.test", uuid.NewString()[:8]),
		"firstName":              "Bad",
		"lastName":               "Niss",
		"nationalRegistryNumber": "abc",
	})
	if code != http.StatusBadRequest || errorMsgKey(env) != "national_registry_invalid" {
		t.Fatalf("bad niss want 400 national_registry_invalid got %d %#v", code, env)
	}
}

func TestClientBillingPatchAndExport(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	email := fmt.Sprintf("client.billing.%s@petsfollow.test", uuid.NewString()[:8])
	password := "ClientDemo123!"

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/clients", vetTok, map[string]any{
		"email":                email,
		"password":             password,
		"firstName":            "Billie",
		"lastName":             "Ng",
		"billingVatNumber":     "BE1000000021",
		"billingCompanyNumber": "1000000021",
		"billingStreet":        "1 rue Peppol",
		"billingCity":          "Liège",
		"billingPostal":        "4000",
		"billingCountry":       "be",
	})
	if code != http.StatusCreated {
		t.Fatalf("create with billing %d %#v", code, env)
	}
	clientID, _ := dataMap(t, env)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing userId: %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["billingVatNumber"] != "BE1000000021" || got["billingCountry"] != "BE" {
		t.Fatalf("create billing %#v", got)
	}
	if got["billingStreet"] != "1 rue Peppol" || got["billingCity"] != "Liège" || got["billingPostal"] != "4000" {
		t.Fatalf("create address billing %#v", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"billingStreet": "2 av. Louise",
		"billingCity":   "Bruxelles",
		"billingPostal": "1050",
	})
	if code != http.StatusOK {
		t.Fatalf("patch billing %d %#v", code, env)
	}
	got = dataMap(t, env)
	if got["billingStreet"] != "2 av. Louise" || got["billingCity"] != "Bruxelles" || got["billingPostal"] != "1050" {
		t.Fatalf("patched billing %#v", got)
	}
	if got["billingVatNumber"] != "BE1000000021" {
		t.Fatalf("vat should persist %#v", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"billingCountry": "B",
	})
	if code != http.StatusBadRequest || errorMsgKey(env) != "billing_country_invalid" {
		t.Fatalf("short country want 400 billing_country_invalid got %d %#v", code, env)
	}

	clientTok := loginToken(t, api.handler, email, password)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	profile, _ := dataMap(t, env)["profile"].(map[string]any)
	if profile == nil {
		t.Fatalf("export missing profile %#v", env)
	}
	if profile["billing_vat_number"] != "BE1000000021" {
		t.Fatalf("export billing_vat_number=%v", profile["billing_vat_number"])
	}
	if profile["billing_street"] != "2 av. Louise" {
		t.Fatalf("export billing_street=%v", profile["billing_street"])
	}
}
