package handlers_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestVetLookupInviteSuggestAndPracticeStamp(t *testing.T) {
	_ = os.Setenv("OPS_NOTIFY_EMAIL", "ops@petsfollow.test")
	api := newTestAPI(t)

	email := uniqueEmail("e2e-vet-lookup")
	password := "ClientPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "Lookup Client",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := strings.TrimPrefix(confirmPath, "/confirm-email?token=")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{
		"token": token,
	})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	clientTok, _ := dataMap(t, env)["accessToken"].(string)
	if clientTok == "" {
		clientTok = loginToken(t, api.handler, email, password)
	}
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM practice.vet_leads WHERE client_user_id = (
				SELECT id FROM identity.users WHERE email = $1)`, email)
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM practice.client_vet_link_requests WHERE client_user_id = (
				SELECT id FROM identity.users WHERE email = $1)`, email)
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM practice.practice_clients WHERE client_user_id = (
				SELECT id FROM identity.users WHERE email = $1)`, email)
		_, _ = api.pool.Exec(context.Background(), `
			DELETE FROM identity.users WHERE email = $1`, email)
	})

	// Lookup by practice fragment.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/vets/lookup?q=VetPlus", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("lookup %d %#v", code, env)
	}
	hits, ok := env["data"].([]any)
	if !ok || len(hits) == 0 {
		t.Fatalf("expected lookup hits, got %#v", env["data"])
	}
	var vetUserID string
	for _, raw := range hits {
		m, _ := raw.(map[string]any)
		if em, _ := m["vetEmail"].(string); em == "vet.demo@petsfollow.test" {
			vetUserID, _ = m["vetUserId"].(string)
			break
		}
	}
	if vetUserID == "" {
		t.Fatalf("vet.demo not in lookup hits: %#v", hits)
	}

	// Invite by vetUserId.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/invite", clientTok, map[string]any{
		"vetUserId": vetUserID,
	})
	if code != http.StatusOK {
		t.Fatalf("invite %d %#v", code, env)
	}
	if dataMap(t, env)["found"] != true {
		t.Fatalf("expected found invite: %#v", env)
	}

	// List pending for vet and accept → stamps client practice_id.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/link-requests", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("link-requests %d %#v", code, env)
	}
	reqs, _ := env["data"].([]any)
	var reqID string
	for _, raw := range reqs {
		m, _ := raw.(map[string]any)
		if ce, _ := m["clientEmail"].(string); ce == email {
			reqID, _ = m["id"].(string)
			break
		}
	}
	if reqID == "" {
		t.Fatalf("pending request for %s not found: %#v", email, reqs)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/link-requests/"+reqID+"/accept", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("accept %d %#v", code, env)
	}

	var practiceID string
	err := api.pool.QueryRow(context.Background(), `
		SELECT practice_id::text FROM identity.users WHERE email = $1`, email).Scan(&practiceID)
	if err != nil || practiceID == "" {
		t.Fatalf("expected stamped practice_id after accept: %v %q", err, practiceID)
	}

	// Suggest unknown vet → lead persisted.
	suggestEmail := uniqueEmail("suggested-vet")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/suggest", clientTok, map[string]any{
		"email": suggestEmail, "phone": "+32470000000", "fullName": "Dr Suggest",
	})
	if code != http.StatusCreated {
		t.Fatalf("suggest %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "suggested" {
		t.Fatalf("expected suggested: %#v", env)
	}
	var leadCount int
	err = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*)::int FROM practice.vet_leads WHERE lower(email) = lower($1)`, suggestEmail,
	).Scan(&leadCount)
	if err != nil || leadCount != 1 {
		t.Fatalf("expected 1 vet_lead, got %d (%v)", leadCount, err)
	}

	// Suggest of an existing vet email upgrades to invite.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/suggest", clientTok, map[string]any{
		"email": "vet.parc@petsfollow.test", "phone": "+32471111111",
	})
	if code != http.StatusOK {
		t.Fatalf("suggest existing %d %#v", code, env)
	}
	if dataMap(t, env)["found"] != true {
		t.Fatalf("expected invite for existing vet: %#v", env)
	}
}
