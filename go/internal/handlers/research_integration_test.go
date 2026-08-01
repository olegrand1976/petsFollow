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

func TestResearchDisabled(t *testing.T) {
	t.Setenv("RESEARCH_ENABLED", "false")
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/overview", tok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 research_disabled, got %d %#v", code, env)
	}
	if errCode(env) != "research_disabled" {
		t.Fatalf("expected research_disabled, got %#v", env)
	}
}

func TestResearchOptInETLAndOverview(t *testing.T) {
	t.Setenv("RESEARCH_ENABLED", "true")
	t.Setenv("RESEARCH_ETL_SECRET", "test-research-etl")
	t.Setenv("RESEARCH_ANON_SALT", "test-research-salt")
	t.Setenv("APP_ENV", "test")
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()

	ensureResearchDemoUser(t, api, st)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/practice/research-opt-in", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("opt-in %d %#v", code, env)
	}
	if dataMap(t, env)["optedIn"] != true {
		t.Fatalf("expected optedIn true: %#v", dataMap(t, env))
	}

	var practiceID, petID string
	err := api.pool.QueryRow(ctx, `
		SELECT u.practice_id::text, p.id::text
		FROM identity.users u
		JOIN pets.pets p ON p.practice_id = u.practice_id
		WHERE u.email = 'vet.demo@petsfollow.test'
		LIMIT 1`).Scan(&practiceID, &petID)
	if err != nil {
		t.Fatalf("lookup practice/pet: %v", err)
	}
	visitID := uuid.NewString()
	_, err = api.pool.Exec(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, scheduled_at, status, notes, source, created_at)
		VALUES ($1, $2, $3, NOW(), 'confirmed', 'research etl test', 'vet', NOW())`,
		visitID, petID, practiceID)
	if err != nil {
		t.Fatalf("insert visit: %v", err)
	}

	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/research-etl/run", nil, map[string]string{})
	if code != http.StatusUnauthorized {
		t.Fatalf("etl without secret expected 401, got %d %#v", code, env)
	}
	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/research-etl/run", nil, map[string]string{
		"X-Research-Etl-Secret": "test-research-etl",
	})
	if code != http.StatusOK {
		t.Fatalf("etl run %d %#v", code, env)
	}

	researchTok := loginToken(t, api.handler, "research.demo@petsfollow.test", "ResearchDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/overview", researchTok, nil)
	if code != http.StatusOK {
		t.Fatalf("overview %d %#v", code, env)
	}
	ov := dataMap(t, env)
	for _, forbidden := range []string{"email", "fullName", "petId", "userId", "practiceId", "practice_id_hash"} {
		if _, ok := ov[forbidden]; ok {
			t.Fatalf("PII/leak key %q in overview: %#v", forbidden, ov)
		}
	}

	hash := store.HashPracticeID("test-research-salt", practiceID)
	var n int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM research.anon_events WHERE practice_id_hash = $1`, hash).Scan(&n); err != nil {
		t.Fatalf("count anon events: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected anon events after ETL, got %d", n)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/heatmap?from=not-a-date", researchTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("heatmap bad from expected 400, got %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/research/opt-ins", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin opt-ins %d %#v", code, env)
	}
	items, _ := dataMap(t, env)["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected at least one opt-in practice, got %#v", dataMap(t, env))
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/vet/practice/research-opt-in", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("opt-out %d %#v", code, env)
	}
	if dataMap(t, env)["optedIn"] != false {
		t.Fatalf("expected optedIn false: %#v", dataMap(t, env))
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM research.anon_events WHERE practice_id_hash = $1`, hash).Scan(&n); err != nil {
		t.Fatalf("count after purge: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected purge of anon_events, got %d", n)
	}
}

func TestResearchMultiProfileAttachAdminOnly(t *testing.T) {
	t.Setenv("RESEARCH_ENABLED", "true")
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()

	email := uniqueEmail("vet-research")
	userID := insertVerifiedUser(t, api, string(kernel.RoleVet), email, "VetDemo123!", "Vet Research", nil)
	if err := st.EnsureUserProfiles(ctx, userID); err != nil {
		t.Fatalf("profiles: %v", err)
	}

	commTok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/users/"+userID+"/profiles", commTok, map[string]any{
		"role": "research",
	})
	if code != http.StatusForbidden {
		t.Fatalf("commercial attach research expected 403, got %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/users/"+userID+"/profiles", adminTok, map[string]any{
		"role": "research",
	})
	if code != http.StatusCreated {
		t.Fatalf("attach research %d %#v", code, env)
	}
	profiles, err := st.ListProfiles(ctx, userID)
	if err != nil {
		t.Fatalf("list profiles: %v", err)
	}
	found := false
	for _, p := range profiles {
		if p.Role == kernel.RoleResearch {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected research profile on user, got %#v", profiles)
	}
}

func TestResearchHeatmapDistinctPractices(t *testing.T) {
	t.Setenv("RESEARCH_ENABLED", "true")
	t.Setenv("RESEARCH_ETL_SECRET", "test-research-etl")
	t.Setenv("RESEARCH_ANON_SALT", "test-research-salt")
	t.Setenv("APP_ENV", "test")
	api := newTestAPI(t)
	st := store.New(api.pool)
	ctx := context.Background()
	ensureResearchDemoUser(t, api, st)

	week := store.ResearchIsoWeekMonday(time.Now())
	postal := "99119"
	// Many events from a single practice must not unlock the cell.
	for i := 0; i < 10; i++ {
		_, err := api.pool.Exec(ctx, `
			INSERT INTO research.anon_events (
				id, event_week, postal_code, city, country_code, species, age_band,
				signal_type, payload, source_hash, practice_id_hash
			) VALUES ($1::uuid, $2::date, $3, 'Test', 'BE', 'dog', '1-7',
				'visit_volume', '{}'::jsonb, $4, 'solo-practice-hash')
			ON CONFLICT (source_hash) DO NOTHING`,
			uuid.NewString(), week, postal, "test:heatmap-solo:"+uuid.NewString())
		if err != nil {
			t.Fatalf("insert solo: %v", err)
		}
	}

	researchTok := loginToken(t, api.handler, "research.demo@petsfollow.test", "ResearchDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/heatmap?signal=visit_volume", researchTok, nil)
	if code != http.StatusOK {
		t.Fatalf("heatmap %d %#v", code, env)
	}
	items, _ := dataMap(t, env)["items"].([]any)
	for _, raw := range items {
		cell, _ := raw.(map[string]any)
		if cell["postalCode"] == postal {
			t.Fatalf("solo-practice cell must be hidden: %#v", cell)
		}
	}

	for i := 0; i < store.ResearchKAnonymity; i++ {
		_, err := api.pool.Exec(ctx, `
			INSERT INTO research.anon_events (
				id, event_week, postal_code, city, country_code, species, age_band,
				signal_type, payload, source_hash, practice_id_hash
			) VALUES ($1::uuid, $2::date, $3, 'Test', 'BE', 'dog', '1-7',
				'visit_volume', '{}'::jsonb, $4, $5)
			ON CONFLICT (source_hash) DO NOTHING`,
			uuid.NewString(), week, postal, "test:heatmap-k:"+uuid.NewString(), "multi-hash-"+uuid.NewString())
		if err != nil {
			t.Fatalf("insert multi: %v", err)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/research/heatmap?signal=visit_volume", researchTok, nil)
	if code != http.StatusOK {
		t.Fatalf("heatmap after density %d %#v", code, env)
	}
	found := false
	items, _ = dataMap(t, env)["items"].([]any)
	for _, raw := range items {
		cell, _ := raw.(map[string]any)
		if cell["postalCode"] == postal {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heatmap cell for postal %s after ≥%d practices", postal, store.ResearchKAnonymity)
	}
}

func ensureResearchDemoUser(t *testing.T, api *testAPI, st *store.Store) {
	t.Helper()
	ctx := context.Background()
	var n int
	_ = api.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM identity.users WHERE email = 'research.demo@petsfollow.test'`).Scan(&n)
	if n > 0 {
		return
	}
	userID := insertVerifiedUser(t, api, string(kernel.RoleResearch), "research.demo@petsfollow.test", "ResearchDemo123!", "Nora Research", nil)
	if err := st.EnsureUserProfiles(ctx, userID); err != nil {
		t.Fatalf("ensure research profiles: %v", err)
	}
	if _, err := st.EnsureRoleProfile(ctx, userID, kernel.RoleResearch, "", ""); err != nil {
		t.Fatalf("ensure research role: %v", err)
	}
}
