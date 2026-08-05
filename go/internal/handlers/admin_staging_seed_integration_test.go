package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

// TestAdminStagingSeedOK couvre le contrat de l'endpoint (ACL admin + confirm + verrou
// consultatif tenu pendant le seed + notify) avec un runner stubbé : lancer le vrai
// seed.Run ici tronquait la base partagée en cours de suite (identifiants staff et
// graphe démo régénérés → 401 / pet_not_found aléatoires sur les tests suivants,
// invisible en CI où la base venait d'être seedée). Le seed lui-même est couvert au
// niveau le plus bas dans internal/seed (TestSeedPreserve*, seed_mass) ; le parcours
// complet via HTTP reste dispo en opt-in (TestAdminStagingSeedRealRun).
func TestAdminStagingSeedOK(t *testing.T) {
	api := newTestAPI(t)
	if !handlers.StagingSeedRunnerIsDefault(api.api) {
		t.Fatal("NewAPI doit câbler /admin/staging/seed sur seed.Run")
	}
	api.api.TestSetAdminStagingSeedEnabled(true)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	st := store.New(api.pool)
	var calls int
	var lockHeld bool
	api.api.TestSetStagingSeedRunner(func(ctx context.Context, pool *pgxpool.Pool) error {
		calls++
		if pool == nil {
			t.Error("seed runner: pool nil")
		}
		// Le seed doit tourner sous le verrou consultatif (pas de reset concurrent).
		err := st.TryWithAdvisoryLock(context.Background(), store.StagingSeedLockKey, func(context.Context) error {
			return nil
		})
		lockHeld = errors.Is(err, store.ErrAdvisoryLockBusy)
		return nil
	})
	t.Cleanup(func() { api.api.TestSetStagingSeedRunner(nil) })

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusOK {
		t.Fatalf("seed %d %#v", code, env)
	}
	if calls != 1 {
		t.Fatalf("seed runner appelé %d fois, attendu 1", calls)
	}
	if !lockHeld {
		t.Fatal("seed doit tourner sous StagingSeedLockKey")
	}
	data := dataMap(t, env)
	if ok, _ := data["ok"].(bool); !ok {
		t.Fatalf("expected ok=true %#v", data)
	}
	if _, ok := data["notified"].(float64); !ok {
		t.Fatalf("expected notified number %#v", data)
	}

	// Demo admin still usable after the endpoint ran.
	_ = loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
}

// TestAdminStagingSeedRealRun exécute le vrai seed.Run via HTTP — destructif pour la base
// locale (tout le graphe démo est régénéré), donc opt-in : PF_TEST_REAL_STAGING_SEED=1
// go test ./internal/handlers/ -run TestAdminStagingSeedRealRun. Ne pas activer dans
// `make test-go` ni en CI : les tests suivants de la suite tourneraient sur une base
// re-seedée à mi-parcours.
func TestAdminStagingSeedRealRun(t *testing.T) {
	if os.Getenv("PF_TEST_REAL_STAGING_SEED") != "1" {
		t.Skip("destructif : poser PF_TEST_REAL_STAGING_SEED=1 pour lancer le vrai seed")
	}
	api := newTestAPI(t)
	api.api.TestSetAdminStagingSeedEnabled(true)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/staging/seed", adminTok, map[string]any{
		"confirm": handlers.StagingSeedConfirmPhrase,
	})
	if code != http.StatusOK {
		t.Fatalf("seed %d %#v", code, env)
	}
	if ok, _ := dataMap(t, env)["ok"].(bool); !ok {
		t.Fatalf("expected ok=true %#v", env)
	}
	// Les comptes staff démo doivent rester utilisables après re-seed.
	_ = loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	_ = loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
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
