package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/seed"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestAdminProfileSwitchToSecretaryAndCommercial(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	if err := seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st); err != nil {
		t.Fatalf("ensure multi-switch: %v", err)
	}
	t.Cleanup(func() {
		_ = seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st)
	})

	tok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	secretaryID := profileIDByRole(t, api.handler, tok, "secretary")
	commercialID := profileIDByRole(t, api.handler, tok, "commercial")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", tok, map[string]any{
		"profileId": secretaryID,
	})
	if code != http.StatusOK {
		t.Fatalf("admin→secretary %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK || dataMap(t, env)["role"] != "secretary" {
		t.Fatalf("me after secretary switch: %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", access, map[string]any{
		"profileId": commercialID,
	})
	if code != http.StatusOK {
		t.Fatalf("admin→commercial %d %#v", code, env)
	}
	access, _ = dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK || dataMap(t, env)["role"] != "commercial" {
		t.Fatalf("me after commercial switch: %d %#v", code, env)
	}

	if err := seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st); err != nil {
		t.Fatalf("restore: %v", err)
	}
}

func TestCommercialProfileSwitchForbiddenRoles(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	if err := seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st); err != nil {
		t.Fatalf("ensure multi-switch: %v", err)
	}

	var userID string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users WHERE email='commercial.demo@petsfollow.test'`).Scan(&userID); err != nil {
		t.Fatalf("commercial id: %v", err)
	}
	adminPID, err := st.EnsureRoleProfile(ctx, userID, kernel.RoleAdmin, "", "")
	if err != nil {
		t.Fatalf("plant admin: %v", err)
	}
	mgrPID, err := st.EnsureRoleProfile(ctx, userID, kernel.RoleCommercialManager, "", "")
	if err != nil {
		t.Fatalf("plant manager: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.profiles WHERE id=$1 OR id=$2`, adminPID, mgrPID)
		_ = seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st)
	})

	tok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	for _, pid := range []string{adminPID, mgrPID} {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", tok, map[string]any{
			"profileId": pid,
		})
		if code != http.StatusForbidden {
			t.Fatalf("commercial forbidden switch expected 403, got %d %#v", code, env)
		}
	}
}

func TestCommercialManagerProfileSwitchMatrix(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	if err := seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st); err != nil {
		t.Fatalf("ensure multi-switch: %v", err)
	}
	t.Cleanup(func() {
		_ = seed.EnsureDemoMultiSwitchProfiles(ctx, api.pool, st)
	})

	var userID string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text FROM identity.users WHERE email='commercial.manager@petsfollow.test'`).Scan(&userID); err != nil {
		t.Fatalf("manager id: %v", err)
	}
	adminPID, err := st.EnsureRoleProfile(ctx, userID, kernel.RoleAdmin, "", "")
	if err != nil {
		t.Fatalf("plant admin: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.profiles WHERE id=$1`, adminPID)
	})

	tok := loginToken(t, api.handler, "commercial.manager@petsfollow.test", "CommercialDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", tok, map[string]any{
		"profileId": adminPID,
	})
	if code != http.StatusForbidden {
		t.Fatalf("manager→admin expected 403, got %d %#v", code, env)
	}

	commercialID := profileIDByRole(t, api.handler, tok, "commercial")
	managerID := profileIDByRole(t, api.handler, tok, "commercial_manager")

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", tok, map[string]any{
		"profileId": commercialID,
	})
	if code != http.StatusOK {
		t.Fatalf("manager→commercial %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/profiles/switch", access, map[string]any{
		"profileId": managerID,
	})
	if code != http.StatusOK {
		t.Fatalf("commercial→manager return %d %#v", code, env)
	}
	access, _ = dataMap(t, env)["accessToken"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", access, nil)
	if code != http.StatusOK || dataMap(t, env)["role"] != "commercial_manager" {
		t.Fatalf("me after return to manager: %d %#v", code, env)
	}
}

func profileIDByRole(t *testing.T, h http.Handler, tok, role string) string {
	t.Helper()
	code, env := doAuthJSON(t, h, http.MethodGet, "/api/v1/me/profiles", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("list profiles %d %#v", code, env)
	}
	for _, it := range profilesList(t, env) {
		m, _ := it.(map[string]any)
		if m["role"] == role {
			id, _ := m["id"].(string)
			if id != "" {
				return id
			}
		}
	}
	t.Fatalf("missing profile role=%s in %#v", role, env)
	return ""
}
