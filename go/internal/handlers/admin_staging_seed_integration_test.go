package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestAdminStagingSeedDisabled(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAdminStagingSeedEnabled(false)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/staging/seed", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("status %d %#v", code, env)
	}
	data := dataMap(t, env)
	if enabled, _ := data["enabled"].(bool); enabled {
		t.Fatal("expected enabled=false")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusForbidden || errCode(env) != "forbidden" {
		t.Fatalf("disabled: got %d %#v", code, env)
	}
}

func TestAdminStagingSeedConfirmAndACL(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAdminStagingSeedEnabled(true)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": "wrong",
	})
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("bad confirm: got %d %#v", code, env)
	}

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", vetTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusForbidden {
		t.Fatalf("vet: got %d %#v", code, env)
	}
}

func TestAdminStagingSeedOK(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAdminStagingSeedEnabled(true)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusOK {
		t.Fatalf("seed %d %#v", code, env)
	}
	data := dataMap(t, env)
	if ok, _ := data["ok"].(bool); !ok {
		t.Fatalf("expected ok=true %#v", data)
	}
	if _, ok := data["notified"].(float64); !ok {
		t.Fatalf("expected notified number %#v", data)
	}

	// Demo admin still usable after re-seed.
	_ = loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
}

func TestAdminStagingSeedConflictWhenLocked(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAdminStagingSeedEnabled(true)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	st := store.New(api.pool)
	held := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = st.TryWithAdvisoryLock(context.Background(), store.StagingSeedLockKey, func(ctx context.Context) error {
			close(held)
			select {
			case <-done:
			case <-time.After(30 * time.Second):
			}
			return nil
		})
	}()
	select {
	case <-held:
	case <-time.After(5 * time.Second):
		t.Fatal("advisory lock not acquired")
	}
	defer close(done)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusConflict || errCode(env) != "conflict" {
		t.Fatalf("expected 409 conflict, got %d %#v", code, env)
	}
}
