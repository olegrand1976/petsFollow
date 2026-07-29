package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// Export + DELETE /me sur un client jetable (pas de seed demo).
func TestRGPDExportAndDeleteDisposableClient(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("rgpd-client")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "RGPD Disposable",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	const prefix = "/confirm-email?token="
	if len(confirmPath) > len(prefix) {
		token = confirmPath[len(prefix):]
	}
	if token == "" {
		t.Fatalf("missing confirm token: %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", access, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	data := dataMap(t, env)
	profile, _ := data["profile"].(map[string]any)
	if profile == nil {
		t.Fatalf("export missing profile: %#v", data)
	}
	if _, hasHash := profile["password_hash"]; hasHash {
		t.Fatal("export must not include password_hash")
	}
	if _, ok := data["pets"]; !ok {
		t.Fatalf("client export missing pets key: %#v", data)
	}
	for _, key := range []string{"petDocuments", "deviceTokens"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("client export missing %s: %#v", key, data)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("delete me %d %#v", code, env)
	}
	if ok, _ := dataMap(t, env)["ok"].(bool); !ok {
		t.Fatalf("expected ok true: %#v", env)
	}

	// Login après purge doit échouer.
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code == http.StatusOK {
		t.Fatalf("login after delete should fail, got 200 %#v", env)
	}
}

// DELETE /me assistant → tombstone (pas 404).
func TestRGPDDeleteAssistantTombstone(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	email := uniqueEmail("rgpd-assist")
	password := "VetDemo123!"
	id := insertVerifiedUser(t, api, string(kernel.RoleVetAssistant), email, password, "Assist RGPD", nil)

	tok := loginToken(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me assistant %d %#v", code, env)
	}
	var emailOut string
	if err := api.pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1`, id).Scan(&emailOut); err != nil {
		t.Fatal(err)
	}
	if !store.IsTombstoneEmail(emailOut) {
		t.Fatalf("want tombstone email, got %q", emailOut)
	}
}

// Dual care_pro + pet client : DELETE /me purge le pet puis tombstone.
func TestRGPDDeleteCareProPurgesOwnedPets(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	email := uniqueEmail("rgpd-care")
	password := "CareProDemo123!"
	userID := insertVerifiedUser(t, api, string(kernel.RoleCarePro), email, password, "Care Dual", map[string]any{
		"professional_specialty": "farrier",
	})
	if err := st.EnsureUserProfiles(ctx, userID); err != nil {
		t.Fatal(err)
	}
	petID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
		VALUES ($1, NULL, $2, 'DualPet', 'horse', 'mix', 'pending_payment')`, petID, userID); err != nil {
		t.Fatal(err)
	}

	tok := loginToken(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("export care_pro %d %#v", code, env)
	}
	exportData := dataMap(t, env)
	if _, ok := exportData["pets"]; !ok {
		t.Fatalf("dual care_pro export missing pets: %#v", exportData)
	}
	pets, _ := exportData["pets"].([]any)
	if len(pets) == 0 {
		t.Fatalf("dual care_pro export pets empty: %#v", exportData["pets"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me care_pro %d %#v", code, env)
	}
	var petN int
	if err := api.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pets.pets WHERE id=$1`, petID).Scan(&petN); err != nil {
		t.Fatal(err)
	}
	if petN != 0 {
		t.Fatalf("owned pet must be purged, count=%d", petN)
	}
	var emailOut string
	if err := api.pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1`, userID).Scan(&emailOut); err != nil {
		t.Fatal(err)
	}
	if !store.IsTombstoneEmail(emailOut) {
		t.Fatalf("want tombstone email, got %q", emailOut)
	}
}

// Clients provisionnés : termsAcceptedAt null jusqu'à POST /me/accept-terms.
func TestRGPDAcceptTermsProvisionedClient(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	email := uniqueEmail("rgpd-prov")
	password := "ClientDemo123!"
	userID := insertVerifiedUser(t, api, "client", email, password, "Prov Client", nil)
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users SET terms_accepted_at = NULL WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}

	tok := loginToken(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["termsAcceptedAt"] != nil {
		t.Fatalf("want termsAcceptedAt null, got %#v", me["termsAcceptedAt"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": false,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("accept-terms without consent want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": true,
	})
	if code != http.StatusOK {
		t.Fatalf("accept-terms %d %#v", code, env)
	}
	me = dataMap(t, env)
	if me["termsAcceptedAt"] == nil || me["termsAcceptedAt"] == "" {
		t.Fatalf("want termsAcceptedAt set, got %#v", me)
	}

	// API gated : pets hors allowlist tant que consent manquant (re-null puis 403).
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users SET terms_accepted_at = NULL WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", tok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("pets without terms want 403 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj != nil {
		if c, _ := errObj["code"].(string); c != "consent_required" {
			t.Fatalf("want error.code consent_required got %#v", errObj)
		}
	} else {
		t.Fatalf("want error object with consent_required, got %#v", env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/app-invite", tok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("app-invite without terms want 403 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": true,
	})
	if code != http.StatusOK {
		t.Fatalf("re-accept %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets after accept want 200 got %d %#v", code, env)
	}
}
