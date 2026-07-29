package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPetTimelineIncludesVisitReportForVetNotClient(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	probe := fmt.Sprintf("timeline-cr-probe-%d", time.Now().UnixNano())
	visitID := createDoneVisitWithReport(t, api, vetTok, petID, probe, "", 4*time.Hour)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet timeline %d %#v", code, env)
	}
	vetItem := findTimelineVisit(t, env, visitID)
	meta, _ := vetItem["meta"].(map[string]any)
	if meta == nil {
		t.Fatalf("vet visit meta missing %#v", vetItem)
	}
	if meta["hasReport"] != true {
		t.Fatalf("vet hasReport want true got %#v", meta["hasReport"])
	}
	if meta["reportStatus"] != "draft" {
		t.Fatalf("vet reportStatus want draft got %#v", meta["reportStatus"])
	}
	body, _ := vetItem["body"].(string)
	if !strings.Contains(body, probe) {
		t.Fatalf("vet body %q should contain CR probe %q", body, probe)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client timeline %d %#v", code, env)
	}
	clientItem := findTimelineVisit(t, env, visitID)
	clientMeta, _ := clientItem["meta"].(map[string]any)
	// Draft CR: client must not get hasReport (final-only).
	if clientMeta == nil {
		t.Fatalf("client visit meta missing %#v", clientItem)
	}
	if clientMeta["hasReport"] == true {
		t.Fatalf("client must not see hasReport for draft %#v", clientMeta)
	}
	clientBody, _ := clientItem["body"].(string)
	if strings.Contains(clientBody, probe) {
		t.Fatalf("client body must not leak CR: %q", clientBody)
	}
}

func TestPetTimelineClientHasReportOnFinalCR(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	probe := fmt.Sprintf("timeline-final-cr-%d", time.Now().UnixNano())
	visitID := createDoneVisitWithReport(t, api, vetTok, petID, probe, "", 9*time.Hour)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("client timeline %d %#v", code, env)
	}
	clientItem := findTimelineVisit(t, env, visitID)
	clientMeta, _ := clientItem["meta"].(map[string]any)
	if clientMeta["hasReport"] != true {
		t.Fatalf("client hasReport want true got %#v", clientMeta)
	}
	if clientMeta["reportStatus"] != "final" {
		t.Fatalf("client reportStatus want final got %#v", clientMeta["reportStatus"])
	}
	clientBody, _ := clientItem["body"].(string)
	if strings.Contains(clientBody, probe) {
		t.Fatalf("client body must not leak CR text: %q", clientBody)
	}
}

func TestPetTimelineNonOwnerStripsHasReport(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	ownerTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	granteeTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, ownerTok)

	probe := fmt.Sprintf("timeline-share-cr-%d", time.Now().UnixNano())
	visitID := createDoneVisitWithReport(t, api, vetTok, petID, probe, "", 10*time.Hour)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/visits/"+visitID+"/report/finalize", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/shares", ownerTok, map[string]any{
		"email":      "client.marie@petsfollow.test",
		"permission": "read",
	})
	if code != http.StatusCreated {
		t.Fatalf("share create %d %#v", code, env)
	}
	granteeID, _ := dataMap(t, env)["granteeUserId"].(string)
	t.Cleanup(func() {
		if granteeID != "" {
			_, _ = doAuthJSON(t, api.handler, http.MethodDelete,
				"/api/v1/pets/"+petID+"/shares/"+granteeID, ownerTok, nil)
		}
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", granteeTok, nil)
	if code != http.StatusOK {
		t.Fatalf("grantee timeline %d %#v", code, env)
	}
	item := findTimelineVisit(t, env, visitID)
	meta, _ := item["meta"].(map[string]any)
	if meta != nil && meta["hasReport"] == true {
		t.Fatalf("non-owner must not see hasReport %#v", meta)
	}
}

func TestPetTimelineConfirmedVisitWithReportVisible(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	probe := fmt.Sprintf("timeline-confirmed-cr-%d", time.Now().UnixNano())
	visitID := createConfirmedVisitWithReport(t, api, vetTok, petID, probe, 8*time.Hour)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet timeline %d %#v", code, env)
	}
	item := findTimelineVisit(t, env, visitID)
	meta, _ := item["meta"].(map[string]any)
	if meta["hasReport"] != true {
		t.Fatalf("confirmed+CR hasReport want true got %#v", meta)
	}
	if meta["status"] != "confirmed" {
		t.Fatalf("status want confirmed got %#v", meta["status"])
	}
	body, _ := item["body"].(string)
	if !strings.Contains(body, probe) {
		t.Fatalf("body %q should contain CR probe %q", body, probe)
	}
}

func TestPetTimelineEmptyDraftReportNotHasReport(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	visitID := createDoneVisitWithReport(t, api, vetTok, petID, "", "", 5*time.Hour)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet timeline %d %#v", code, env)
	}
	item := findTimelineVisit(t, env, visitID)
	meta, _ := item["meta"].(map[string]any)
	if meta != nil && meta["hasReport"] == true {
		t.Fatalf("empty draft must not set hasReport %#v", meta)
	}
}

func TestPetTimelinePrefersReportBodyOverVisitNotes(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	crProbe := fmt.Sprintf("timeline-cr-prefer-%d", time.Now().UnixNano())
	notesProbe := fmt.Sprintf("timeline-notes-%d", time.Now().UnixNano())
	visitID := createDoneVisitWithReport(t, api, vetTok, petID, crProbe, notesProbe, 6*time.Hour)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet timeline %d %#v", code, env)
	}
	item := findTimelineVisit(t, env, visitID)
	body, _ := item["body"].(string)
	if !strings.Contains(body, crProbe) {
		t.Fatalf("body %q should prefer CR %q", body, crProbe)
	}
	if strings.Contains(body, notesProbe) {
		t.Fatalf("body %q should not use visit notes when CR exists", body)
	}
}

func TestPetTimelineSecretaryHasReportWithoutBodyExcerpt(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	probe := fmt.Sprintf("timeline-cr-secretary-%d", time.Now().UnixNano())
	visitID := createDoneVisitWithReport(t, api, vetTok, petID, probe, "", 7*time.Hour)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", secTok, nil)
	if code != http.StatusOK {
		t.Fatalf("secretary timeline %d %#v", code, env)
	}
	item := findTimelineVisit(t, env, visitID)
	meta, _ := item["meta"].(map[string]any)
	if meta["hasReport"] != true {
		t.Fatalf("secretary hasReport want true got %#v", meta)
	}
	body, _ := item["body"].(string)
	if strings.Contains(body, probe) {
		t.Fatalf("secretary must not see CR excerpt in body: %q", body)
	}
}

func TestPetTimelineStaffWithoutMessagingOmitsMessages(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	secTok := loginToken(t, api.handler, "secretary.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := demoClientPetID(t, api, clientTok)

	_, meEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", clientTok, nil)
	clientID, _ := dataMap(t, meEnv)["userId"].(string)
	if clientID == "" {
		t.Fatalf("missing client userId %#v", meEnv)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads", vetTok, map[string]any{
		"clientUserId": clientID,
		"petId":        petID,
	})
	if code != http.StatusOK {
		t.Fatalf("ensure thread %d %#v", code, env)
	}
	threadID, _ := dataMap(t, env)["id"].(string)
	if threadID == "" {
		t.Fatalf("missing thread id %#v", env)
	}

	probe := fmt.Sprintf("timeline-msg-gate-%d", time.Now().UnixNano())
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/messaging/threads/"+threadID+"/messages", vetTok, map[string]any{
		"body": probe,
	})
	if code != http.StatusCreated {
		t.Fatalf("send message %d %#v", code, env)
	}

	var prevPerms []byte
	err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(permissions, '{}'::jsonb)
		FROM practice.team_members tm
		JOIN identity.users u ON u.id = tm.user_id
		WHERE u.email = 'secretary.demo@petsfollow.test' AND tm.status = 'active'
		LIMIT 1`).Scan(&prevPerms)
	if err != nil {
		t.Fatalf("load secretary permissions: %v", err)
	}
	_, err = api.pool.Exec(ctx, `
		UPDATE practice.team_members tm
		SET permissions = COALESCE(tm.permissions, '{}'::jsonb) || '{"messaging": false}'::jsonb
		FROM identity.users u
		WHERE u.id = tm.user_id AND u.email = 'secretary.demo@petsfollow.test' AND tm.status = 'active'`)
	if err != nil {
		t.Fatalf("disable secretary messaging: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(ctx, `
			UPDATE practice.team_members tm
			SET permissions = $1::jsonb
			FROM identity.users u
			WHERE u.id = tm.user_id AND u.email = 'secretary.demo@petsfollow.test' AND tm.status = 'active'`,
			prevPerms)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", secTok, nil)
	if code != http.StatusOK {
		t.Fatalf("secretary timeline %d %#v", code, env)
	}
	for _, raw := range env["data"].([]any) {
		item, _ := raw.(map[string]any)
		if item["type"] == "message" {
			body, _ := item["body"].(string)
			if strings.Contains(body, probe) {
				t.Fatalf("secretary without messaging must not see message in timeline: %#v", item)
			}
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/timeline", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet timeline %d %#v", code, env)
	}
	found := false
	for _, raw := range env["data"].([]any) {
		item, _ := raw.(map[string]any)
		if item["type"] != "message" {
			continue
		}
		body, _ := item["body"].(string)
		if strings.Contains(body, probe) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("vet timeline should include message probe %q", probe)
	}
}

func demoClientPetID(t *testing.T, api *testAPI, clientTok string) string {
	t.Helper()
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets %d %#v (make seed?)", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)
	if petID == "" {
		t.Fatalf("missing pet id %#v", pets[0])
	}
	return petID
}

func createDoneVisitWithReport(t *testing.T, api *testAPI, vetTok, petID, bodyText, notes string, offset time.Duration) string {
	t.Helper()
	visitID := createVisitWithOptionalReport(t, api, vetTok, petID, bodyText, notes, offset, true)
	return visitID
}

func createConfirmedVisitWithReport(t *testing.T, api *testAPI, vetTok, petID, bodyText string, offset time.Duration) string {
	t.Helper()
	return createVisitWithOptionalReport(t, api, vetTok, petID, bodyText, "", offset, false)
}

func createVisitWithOptionalReport(t *testing.T, api *testAPI, vetTok, petID, bodyText, notes string, offset time.Duration, markDone bool) string {
	t.Helper()
	var visitID string
	for attempt := 0; attempt < 4; attempt++ {
		when := time.Now().UTC().Add(offset + time.Duration(attempt)*time.Hour).Truncate(time.Minute).Format(time.RFC3339)
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
			"scheduledAt":     when,
			"notes":           notes,
			"durationMinutes": 30,
			"confirmDirect":   true,
		})
		if code == http.StatusCreated || code == http.StatusOK {
			visitID, _ = dataMap(t, env)["id"].(string)
			break
		}
		if code != http.StatusConflict {
			t.Fatalf("create visit %d %#v", code, env)
		}
	}
	if visitID == "" {
		t.Fatal("create visit: all slot retries failed (409)")
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "cancelled",
		})
	})

	code, env := doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": bodyText,
	})
	if code != http.StatusOK {
		t.Fatalf("put report %d %#v", code, env)
	}

	if markDone {
		code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/visits/"+visitID, vetTok, map[string]any{
			"status": "done",
		})
		if code != http.StatusOK {
			t.Fatalf("mark done %d %#v", code, env)
		}
	}
	return visitID
}

func findTimelineVisit(t *testing.T, env map[string]any, visitID string) map[string]any {
	t.Helper()
	items, _ := env["data"].([]any)
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if item == nil {
			continue
		}
		if item["type"] != "visit" {
			continue
		}
		if item["id"] == visitID {
			return item
		}
		meta, _ := item["meta"].(map[string]any)
		if meta != nil && meta["visitId"] == visitID {
			return item
		}
	}
	t.Fatalf("visit %s not found in timeline (%d items)", visitID, len(items))
	return nil
}
