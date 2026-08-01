package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestResearchGroupsAndDataRoom(t *testing.T) {
	t.Setenv("RESEARCH_ENABLED", "true")
	t.Setenv("RESEARCH_ETL_SECRET", "test-research-etl")
	t.Setenv("RESEARCH_ANON_SALT", "test-research-salt")
	t.Setenv("APP_ENV", "test")
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	ensureResearchDemoUser(t, api, st)

	researchTok := loginToken(t, api.handler, "research.demo@petsfollow.test", "ResearchDemo123!")

	// Seed may enable a demo Data room group — isolate gate assertions, then restore.
	rows, err := api.pool.Query(ctx, `SELECT id::text, dataroom_enabled FROM research.groups`)
	if err != nil {
		t.Fatalf("list group flags: %v", err)
	}
	prevFlags := map[string]bool{}
	for rows.Next() {
		var id string
		var enabled bool
		if err := rows.Scan(&id, &enabled); err != nil {
			rows.Close()
			t.Fatalf("scan group flags: %v", err)
		}
		prevFlags[id] = enabled
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("group flags rows: %v", err)
	}
	t.Cleanup(func() {
		for id, enabled := range prevFlags {
			if _, err := api.pool.Exec(context.Background(),
				`UPDATE research.groups SET dataroom_enabled = $2 WHERE id = $1::uuid`, id, enabled); err != nil {
				t.Errorf("restore dataroom flag %s: %v", id, err)
			}
		}
	})
	if _, err := api.pool.Exec(ctx, `UPDATE research.groups SET dataroom_enabled = false`); err != nil {
		t.Fatalf("reset dataroom flags: %v", err)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/dataroom/events", researchTok, nil)
	if code != http.StatusForbidden || errCode(env) != "research_dataroom_forbidden" {
		t.Fatalf("dataroom without access expected 403 research_dataroom_forbidden, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/research/groups", researchTok, map[string]any{
		"name":        "Cohorte test",
		"description": "intégration",
	})
	if code != http.StatusCreated {
		t.Fatalf("create group %d %#v", code, env)
	}
	groupID, _ := dataMap(t, env)["id"].(string)
	if groupID == "" {
		t.Fatalf("missing group id: %#v", dataMap(t, env))
	}
	if dataMap(t, env)["dataroomEnabled"] == true {
		t.Fatalf("new group must not unlock dataroom by default: %#v", dataMap(t, env))
	}

	// Solo group still insufficient without admin dataroom flag.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/dataroom/events", researchTok, nil)
	if code != http.StatusForbidden || errCode(env) != "research_dataroom_forbidden" {
		t.Fatalf("dataroom after solo group expected 403, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/groups/not-a-uuid", researchTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid group id expected 400, got %d %#v", code, env)
	}

	var adminID string
	if err := api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email='admin.demo@petsfollow.test'`).Scan(&adminID); err != nil {
		t.Fatalf("admin id: %v", err)
	}
	if _, err := st.EnsureRoleProfile(ctx, adminID, kernel.RoleResearch, "", ""); err != nil {
		t.Fatalf("admin research profile: %v", err)
	}

	// Anti-enumeration: unknown email still 200 ok.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/research/groups/"+groupID+"/members", researchTok, map[string]any{
		"email": "nobody-research@petsfollow.test",
	})
	if code != http.StatusOK || dataMap(t, env)["ok"] != true {
		t.Fatalf("add unknown member expected uniform 200 ok, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/research/groups/"+groupID+"/members", researchTok, map[string]any{
		"email": "admin.demo@petsfollow.test",
	})
	if code != http.StatusOK || dataMap(t, env)["ok"] != true {
		t.Fatalf("add member %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/research/groups/"+groupID+"/dataroom", adminTok, map[string]any{
		"enabled": true,
	})
	if code != http.StatusOK {
		t.Fatalf("admin enable dataroom %d %#v", code, env)
	}
	if dataMap(t, env)["dataroomEnabled"] != true {
		t.Fatalf("expected dataroomEnabled true: %#v", dataMap(t, env))
	}

	week := time.Now().UTC().Truncate(24 * time.Hour)
	for week.Weekday() != time.Monday {
		week = week.AddDate(0, 0, -1)
	}
	for i := 0; i < store.ResearchKAnonymity; i++ {
		_, err := api.pool.Exec(ctx, `
			INSERT INTO research.anon_events (
				id, event_week, postal_code, city, country_code, species, age_band,
				signal_type, payload, source_hash, practice_id_hash
			) VALUES ($1::uuid, $2::date, '1000', 'Bruxelles', 'BE', 'dog', '1-7',
				'visit_volume', '{}'::jsonb, $3, $4)
			ON CONFLICT (source_hash) DO NOTHING`,
			uuid.NewString(), week, "test:dataroom:"+uuid.NewString(), "test-hash-"+uuid.NewString())
		if err != nil {
			t.Fatalf("insert density: %v", err)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/dataroom/events?species=dog&signal=visit_volume", researchTok, nil)
	if code != http.StatusOK {
		t.Fatalf("dataroom %d %#v", code, env)
	}
	dm := dataMap(t, env)
	items, _ := dm["items"].([]any)
	if len(items) < store.ResearchKAnonymity {
		t.Fatalf("expected >= %d events, got %d %#v", store.ResearchKAnonymity, len(items), dm)
	}
	first, _ := items[0].(map[string]any)
	for _, forbidden := range []string{
		"practiceId", "practice_id_hash", "sourceHash", "petId", "userId", "email", "city", "payload",
	} {
		if _, ok := first[forbidden]; ok {
			t.Fatalf("PII/leak key %q in dataroom event: %#v", forbidden, first)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/research/groups/"+groupID, researchTok, nil)
	if code != http.StatusOK {
		t.Fatalf("delete group %d %#v", code, env)
	}
}
