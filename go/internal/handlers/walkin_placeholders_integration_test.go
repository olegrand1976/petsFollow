package handlers_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestWalkinPlaceholdersEnsureAndImmutability(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list clients %d %#v", code, env)
	}
	list, _ := env["data"].([]any)
	var walkin map[string]any
	for _, raw := range list {
		c, _ := raw.(map[string]any)
		if c != nil && c["isWalkinPlaceholder"] == true {
			walkin = c
			break
		}
	}
	if walkin == nil {
		t.Fatal("expected walk-in Nouveau client in practice clients list")
	}
	clientID, _ := walkin["userId"].(string)
	if clientID == "" {
		t.Fatal("walkin userId empty")
	}
	if name, _ := walkin["fullName"].(string); !strings.Contains(strings.ToLower(name), "nouveau") {
		t.Fatalf("unexpected walkin name %#v", walkin["fullName"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/clients/"+clientID, vetTok, map[string]any{
		"firstName": "Hacked",
	})
	if code != http.StatusForbidden {
		t.Fatalf("patch walkin want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+clientID+"/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("walkin pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("expected walkin pet")
	}
	pet, _ := pets[0].(map[string]any)
	if pet["isWalkinPlaceholder"] != true {
		t.Fatalf("expected walkin pet flag %#v", pet)
	}
	petID, _ := pet["id"].(string)

	// Login with synthetic email must fail.
	email, _ := walkin["email"].(string)
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": "anything",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("walkin login want 401 got %d %#v", code, env)
	}

	_ = petID
}

func TestWalkinIdentifyCreateAndExisting(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("clients %d %#v", code, env)
	}
	var walkinClientID, walkinPetID string
	for _, raw := range env["data"].([]any) {
		c := raw.(map[string]any)
		if c["isWalkinPlaceholder"] == true {
			walkinClientID, _ = c["userId"].(string)
			break
		}
	}
	if walkinClientID == "" {
		t.Fatal("no walkin client")
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/clients/"+walkinClientID+"/pets", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v", code, env)
	}
	for _, raw := range env["data"].([]any) {
		p := raw.(map[string]any)
		if p["isWalkinPlaceholder"] == true {
			walkinPetID, _ = p["id"].(string)
			break
		}
	}
	if walkinPetID == "" {
		t.Fatal("no walkin pet")
	}

	slot := time.Now().UTC().Add(12 * time.Minute).Truncate(time.Second).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+walkinPetID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"confirmDirect":       true,
		"consultationSession": true,
		"notes":               "walkin identify test",
		"callbackPhone":       "0470 99 88 77",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create walkin visit %d %#v", code, env)
	}
	visit := env["data"].(map[string]any)
	visitID, _ := visit["id"].(string)
	if visitID == "" {
		t.Fatal("visit id empty")
	}
	if visit["callbackPhone"] != "0470 99 88 77" {
		t.Fatalf("callbackPhone want set got %#v", visit["callbackPhone"])
	}

	// Missing phone on walk-in pet must 400.
	slotBad := time.Now().UTC().Add(8 * time.Minute).Truncate(time.Second).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+walkinPetID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slotBad,
		"confirmDirect":       true,
		"consultationSession": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("walkin without phone want 400 got %d %#v", code, env)
	}

	// Finalize CR must be blocked while still on walk-in pet.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("finalize before identify want 409 got %d %#v", code, env)
	}

	uniq := strings.ReplaceAll(time.Now().UTC().Format("150405.000"), ".", "")
	email := "walkin.identified." + uniq + "@petsfollow.test"
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/identify-client", vetTok, map[string]any{
		"mode": "create",
		"client": map[string]any{
			"firstName":    "Alice",
			"lastName":     "Walkin",
			"contactPhone": "0470 11 22 33",
			"email":        email,
		},
		"newPet": map[string]any{
			"name":    "Rex",
			"species": "dog",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("identify create %d %#v", code, env)
	}
	out := env["data"].(map[string]any)
	client := out["client"].(map[string]any)
	pet := out["pet"].(map[string]any)
	newPetID, _ := pet["id"].(string)
	if newPetID == "" || newPetID == walkinPetID {
		t.Fatalf("expected new pet id, got %#v", pet)
	}
	if client["isWalkinPlaceholder"] == true {
		t.Fatal("identified client must not be walkin")
	}
	reVisit := out["visit"].(map[string]any)
	if reVisit["petId"] != newPetID {
		t.Fatalf("visit petId want %s got %#v", newPetID, reVisit["petId"])
	}

	// Second visit on same walk-in pet, identify to existing client.
	slot2 := time.Now().UTC().Add(18 * time.Minute).Truncate(time.Second).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+walkinPetID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot2,
		"confirmDirect":       true,
		"consultationSession": true,
		"callbackPhone":       "0471 00 11 22",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("second walkin visit %d %#v", code, env)
	}
	visit2ID, _ := env["data"].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visit2ID+"/identify-client", vetTok, map[string]any{
		"mode":         "existing",
		"clientUserId": client["userId"],
		"newPet": map[string]any{
			"name":    "Luna",
			"species": "cat",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("identify existing %d %#v", code, env)
	}
	pet2 := env["data"].(map[string]any)["pet"].(map[string]any)
	if pet2["id"] == walkinPetID || pet2["id"] == newPetID {
		t.Fatalf("expected another new pet, got %#v", pet2)
	}
}

func TestWalkinPlaceholdersExcludedFromInactiveRetentionList(t *testing.T) {
	api := newTestAPI(t)
	ctx := t.Context()
	st := store.New(api.pool)

	var walkinID string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users
		WHERE is_walkin_placeholder AND role = 'client'
		LIMIT 1`).Scan(&walkinID); err != nil || walkinID == "" {
		t.Fatalf("need seeded walk-in: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users
		SET last_login_at = NOW() - INTERVAL '4 years', created_at = NOW() - INTERVAL '4 years'
		WHERE id = $1`, walkinID); err != nil {
		t.Fatal(err)
	}

	cutoff := time.Now().Add(-3 * 365 * 24 * time.Hour)
	accounts, err := st.ListInactiveAccounts(ctx, cutoff, 500)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accounts {
		if a.ID == walkinID {
			t.Fatalf("walk-in placeholder %s must not appear in inactive retention list", walkinID)
		}
	}
}
