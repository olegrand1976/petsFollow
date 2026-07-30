package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestDevRoleSupportAndBillingGate(t *testing.T) {
	api := newTestAPI(t)
	devTok := loginToken(t, api.handler, "dev.demo@petsfollow.test", "AdminDemo123!")
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/tickets?limit=5", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev list tickets %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/support/stats", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev support stats %d %#v", code, env)
	}
	stats := dataMap(t, env)
	if _, ok := stats["byStatus"]; !ok {
		t.Fatalf("expected byStatus in stats: %#v", stats)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/runtime-flags", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev runtime flags %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/users", devTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dev list users %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/payments", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev payments expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/staging/seed", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev staging seed expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/metrics/overview", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev metrics expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/commercials", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev list commercials expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", devTok, map[string]any{
		"email": "dev-should-fail@petsfollow.test", "fullName": "Nope", "password": "VetDemo123!",
	})
	if code != http.StatusForbidden {
		t.Fatalf("dev create commercial expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/brand-assets", devTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("dev brand-assets expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/payments", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin payments %d %#v", code, env)
	}

	// Create + patch ticket as DEV.
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/support/tickets", vetTok, map[string]any{
		"source":  "nuxt_pro",
		"subject": "DEV patch test",
		"message": "ticket for DEV status change",
	})
	if code != http.StatusCreated {
		t.Fatalf("create ticket %d %#v", code, env)
	}
	ticketID, _ := dataMap(t, env)["id"].(string)
	if ticketID == "" {
		t.Fatalf("missing ticket id")
	}
	t.Cleanup(func() {
		if ticketID == "" {
			return
		}
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/support/tickets/"+ticketID, adminTok, nil)
	})
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/support/tickets/"+ticketID, devTok, map[string]any{
		"status": "in_progress",
	})
	if code != http.StatusOK {
		t.Fatalf("dev patch ticket %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "in_progress" {
		t.Fatalf("expected in_progress, got %#v", dataMap(t, env)["status"])
	}
	// DEV can also hard-delete (ops cleanup).
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/support/tickets/"+ticketID, devTok, nil)
	if code != http.StatusNoContent {
		t.Fatalf("dev delete ticket %d %#v", code, env)
	}
	ticketID = ""
}

func TestDevProfileSwitchToVet(t *testing.T) {
	assertOpsCanSwitchToVet(t, "dev.demo@petsfollow.test", string(kernel.RoleDev))
}

func TestAdminProfileSwitchToVet(t *testing.T) {
	assertOpsCanSwitchToVet(t, "admin.demo@petsfollow.test", string(kernel.RoleAdmin))
}

// TestAdminEnsureUserProfilesNoAutoClient locks the IsProRole(admin) exclusion:
// a disposable admin must not get a personal client profile from EnsureUserProfiles alone.
func TestAdminEnsureUserProfilesNoAutoClient(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	email := uniqueEmail("admin-no-client")
	userID := insertVerifiedUser(t, api, string(kernel.RoleAdmin), email, "AdminDemo123!", "Admin Jetable", nil)
	if err := st.EnsureUserProfiles(ctx, userID); err != nil {
		t.Fatalf("EnsureUserProfiles: %v", err)
	}
	profiles, err := st.ListProfiles(ctx, userID)
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	roles := map[kernel.Role]bool{}
	for _, p := range profiles {
		roles[p.Role] = true
	}
	if !roles[kernel.RoleAdmin] {
		t.Fatalf("expected admin profile, got %#v", profiles)
	}
	if roles[kernel.RoleClient] {
		t.Fatalf("admin must not auto-get client via IsProRole, got %#v", profiles)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected single admin profile, got %#v", profiles)
	}
}

func assertOpsCanSwitchToVet(t *testing.T, email, opsRole string) {
	t.Helper()
	api := newTestAPI(t)
	st := store.New(api.pool)
	if err := seed.EnsureDemoOpsVetProfiles(context.Background(), api.pool, st); err != nil {
		t.Fatalf("ensure ops vet profiles: %v", err)
	}
	_ = seed.EnsureDemoMultiSwitchProfiles(context.Background(), api.pool, st)
	t.Cleanup(func() {
		if err := seed.EnsureDemoOpsVetProfiles(context.Background(), api.pool, st); err != nil {
			t.Errorf("cleanup restore ops profiles: %v", err)
		}
		_ = seed.EnsureDemoMultiSwitchProfiles(context.Background(), api.pool, st)
	})

	tok := loginToken(t, api.handler, email, "AdminDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/profiles", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list profiles %d %#v", code, env)
	}
	list := profilesList(t, env)
	if len(list) < 2 {
		t.Fatalf("expected multi profiles for %s, got %#v", opsRole, env)
	}
	var vetProfileID string
	roles := map[string]bool{}
	for _, it := range list {
		m, _ := it.(map[string]any)
		role, _ := m["role"].(string)
		roles[role] = true
		if role == "vet" {
			vetProfileID, _ = m["id"].(string)
		}
	}
	if !roles[opsRole] || !roles["vet"] {
		t.Fatalf("%s seed expected %s+vet profiles, got %#v", opsRole, opsRole, list)
	}
	if !roles["client"] {
		t.Fatalf("%s seed expected client profile, got %#v", opsRole, list)
	}
	if vetProfileID == "" {
		t.Fatalf("%s missing vet profile: %#v", opsRole, list)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", tok, map[string]any{
		"profileId": vetProfileID,
	})
	if code != http.StatusOK {
		t.Fatalf("switch to vet %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		t.Fatalf("expected access token after switch: %#v", env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("me after switch %d %#v", code, env)
	}
	if dataMap(t, env)["role"] != "vet" {
		t.Fatalf("expected role vet after switch, got %#v", dataMap(t, env)["role"])
	}
	// Restore immediately (narrow poison window for shared seed accounts).
	if err := seed.EnsureDemoOpsVetProfiles(context.Background(), api.pool, st); err != nil {
		t.Fatalf("restore ops profiles after switch: %v", err)
	}
	_ = seed.EnsureDemoMultiSwitchProfiles(context.Background(), api.pool, st)
}

func profilesList(t *testing.T, env map[string]any) []any {
	t.Helper()
	raw := env["data"]
	if list, ok := raw.([]any); ok {
		return list
	}
	if m, ok := raw.(map[string]any); ok {
		if list, ok := m["profiles"].([]any); ok {
			return list
		}
	}
	return nil
}
