package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Filiation chain anti-regression (F1–F14).
//
// Invariants locked here:
//  1. First-wins — a second commercial/client claim never overwrites the first promoter;
//     practice_clients first-wins per (practice_id, client) — same cabinet keeps first vet_user_id.
//  2. Invite > nearby — commercial invite_code persists (or backfills); covered mainly in P4–P6.
//  3. Vet assign durable — second encode of an already-assigned vet email → 409; assign unchanged.
//  4. Chain isolation — client referred to A + claim vet QR keeps commercial_referrals = A.
//  5. Chain Comm→Vet→Client — ResolveVetCommercial returns the vet's assigned commercial.
//  6. Unassign — Resolve falls back to commercial_referrals when assigned_commercial_id is cleared.
//  7. Soft-fail invite — invalid code must not attach nearby (P9).
//  8. Multi-cabinet — Resolve(practiceID) is scoped; Resolve("") = latest global link (F11).
//  9. RGPD — export includes commercialReferrals / clientReferrals; DELETE /me purges referrals (F12).
// 10. Accrue — multi-cabinet pays the commercial of the pet's practice (F13).
// 11. Audit — filiation_events on assign / referral / unassign (F14).
//
// Commission priority (documented, not a filiation loss):
//   assigned_commercial_id on the linked vet always wins over commercial_referrals (F9).
// Linked commercialCreateClient does not write commercial_referrals (commission via vet assign).

func ensureAdminToken(t *testing.T, api *testAPI) string {
	t.Helper()
	return loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
}

func createCommercial(t *testing.T, api *testAPI, adminTok, prefix, fullName string) (commID, email, tok string) {
	t.Helper()
	email = uniqueEmail(prefix)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": email, "password": "CommercialDemo123!", "fullName": fullName,
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID, _ = dataMap(t, env)["userId"].(string)
	if commID == "" {
		t.Fatalf("missing commercial userId %#v", env)
	}
	tok = loginToken(t, api.handler, email, "CommercialDemo123!")
	return commID, email, tok
}

func assignedCommercialID(t *testing.T, api *testAPI, vetEmailOrID string, byEmail bool) string {
	t.Helper()
	var assigned string
	var err error
	if byEmail {
		err = api.pool.QueryRow(context.Background(), `
			SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`,
			vetEmailOrID).Scan(&assigned)
	} else {
		err = api.pool.QueryRow(context.Background(), `
			SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`,
			vetEmailOrID).Scan(&assigned)
	}
	if err != nil {
		t.Fatal(err)
	}
	return assigned
}

func commercialReferralOf(t *testing.T, api *testAPI, clientUserID string) (commercialID, inviteCode string) {
	t.Helper()
	err := api.pool.QueryRow(context.Background(), `
		SELECT commercial_user_id::text, COALESCE(invite_code,'')
		FROM practice.commercial_referrals WHERE client_user_id=$1`, clientUserID).
		Scan(&commercialID, &inviteCode)
	if err != nil {
		t.Fatalf("commercial_referrals for %s: %v", clientUserID, err)
	}
	return commercialID, inviteCode
}

func commercialReferralCount(t *testing.T, api *testAPI, clientUserID string) int {
	t.Helper()
	var n int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientUserID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func userIDByEmail(t *testing.T, api *testAPI, email string) string {
	t.Helper()
	var id string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT id::text FROM identity.users WHERE email=$1`, email).Scan(&id); err != nil {
		t.Fatalf("user %s: %v", email, err)
	}
	return id
}

func verifyEmail(t *testing.T, api *testAPI, userID string) {
	t.Helper()
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE identity.users SET email_verified_at=NOW() WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
}

func insertVetWithPractice(t *testing.T, api *testAPI, st *store.Store, prefix, practiceName string) (vetID, practiceID, inviteCode string) {
	t.Helper()
	ctx := context.Background()
	vetID = insertVerifiedUser(t, api, "vet", uniqueEmail(prefix), "VetDemo123!", practiceName, nil)
	practiceID = uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO practice.practices (id, name, profile_completed_at)
		VALUES ($1::uuid, $2, NOW())`, practiceID, practiceName); err != nil {
		t.Fatal(err)
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users SET practice_id=$2::uuid WHERE id=$1`, vetID, practiceID); err != nil {
		t.Fatal(err)
	}
	inv, err := st.EnsureAppInviteCode(ctx, vetID)
	if err != nil {
		t.Fatal(err)
	}
	return vetID, practiceID, inv.Code
}

func practiceClientVet(t *testing.T, api *testAPI, clientID, practiceID string) string {
	t.Helper()
	var vetID string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT vet_user_id::text FROM practice.practice_clients
		WHERE client_user_id=$1 AND practice_id=$2`, clientID, practiceID).Scan(&vetID); err != nil {
		t.Fatalf("practice_clients (%s,%s): %v", clientID, practiceID, err)
	}
	return vetID
}

func clientPrimaryPractice(t *testing.T, api *testAPI, clientID string) string {
	t.Helper()
	var practiceID string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(practice_id::text,'') FROM identity.users WHERE id=$1`, clientID).Scan(&practiceID); err != nil {
		t.Fatal(err)
	}
	return practiceID
}

// referredClientClaimsAssignedVet: commercial referrerInvite owns commercial_referrals;
// encodeVetCommercial owns the vet (may equal referrer). Returns client, vet, practice IDs.
func referredClientClaimsAssignedVet(
	t *testing.T, api *testAPI,
	referrerCommID, encodeVetCommTok, clientPrefix, vetPrefix string,
) (clientID, vetID, practiceID string) {
	t.Helper()
	ctx := context.Background()
	st := store.New(api.pool)

	invRef, err := st.EnsureAppInviteCode(ctx, referrerCommID)
	if err != nil {
		t.Fatal(err)
	}

	vetEmail := uniqueEmail(vetPrefix)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", encodeVetCommTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr Fil",
		"practiceName": "Cabinet Fil " + vetPrefix,
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ = dataMap(t, env)["userId"].(string)
	invVet, err := st.EnsureAppInviteCode(ctx, vetID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail(clientPrefix)
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Client Fil",
		"consent": true, "inviteCode": invRef.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	clientID = userIDByEmail(t, api, clientEmail)
	ref, savedCode := commercialReferralOf(t, api, clientID)
	if ref != referrerCommID || savedCode != invRef.Code {
		t.Fatalf("referral=%s code=%q want %s / %s", ref, savedCode, referrerCommID, invRef.Code)
	}

	verifyEmail(t, api, clientID)
	tok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": invVet.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim vet %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["status"] != "linked" || data["kind"] != "vet" {
		t.Fatalf("claim status/kind=%v/%v want linked/vet", data["status"], data["kind"])
	}

	if err := api.pool.QueryRow(ctx, `
		SELECT practice_id::text FROM identity.users WHERE id=$1`, vetID).Scan(&practiceID); err != nil {
		t.Fatal(err)
	}
	return clientID, vetID, practiceID
}

// F1 — Encode POST /commercial/vets persists assigned_commercial_id.
func TestFiliationChain_EncodeVetAssignsCommercial(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, _, commTok := createCommercial(t, api, adminTok, "fil-f1-comm", "F1 Comm")

	vetEmail := uniqueEmail("fil-f1-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F1",
		"practiceName": "Cabinet F1", "city": "Bruxelles",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)
	if vetID == "" {
		t.Fatalf("missing vet userId %#v", env)
	}
	if got := assignedCommercialID(t, api, vetID, false); got != commID {
		t.Fatalf("assigned_commercial_id=%s want %s (filiation commercial→véto lost)", got, commID)
	}
}

// F2 — Second commercial cannot steal an already-assigned vet (409 already_assigned).
func TestFiliationChain_EncodeConflictKeepsFirstAssign(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-f2-a", "F2 Comm A")
	_, _, tokB := createCommercial(t, api, adminTok, "fil-f2-b", "F2 Comm B")

	vetEmail := uniqueEmail("fil-f2-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F2",
		"practiceName": "Cabinet F2",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode A %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokB, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F2 Steal",
		"practiceName": "Cabinet Steal",
	})
	if code != http.StatusConflict {
		t.Fatalf("second encode want 409, got %d %#v", code, env)
	}
	if errorMsgKey(env) != "already_assigned" {
		t.Fatalf("msgKey=%q want already_assigned %#v", errorMsgKey(env), env)
	}
	if got := assignedCommercialID(t, api, vetEmail, true); got != commA {
		t.Fatalf("assign overwritten to %s want A %s", got, commA)
	}
}

// F3 — Admin unassign clears assign and returns vet to pool.
func TestFiliationChain_AdminUnassignReturnsToPool(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, _, commTok := createCommercial(t, api, adminTok, "fil-f3-comm", "F3 Comm")

	vetEmail := uniqueEmail("fil-f3-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F3",
		"practiceName": "Cabinet F3",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/vets/"+vetID+"/unassign", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassign %d %#v", code, env)
	}
	if got := assignedCommercialID(t, api, vetID, false); got != "" {
		t.Fatalf("after unassign assigned=%s want empty", got)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/vets/unassigned", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("pool list %d %#v", code, env)
	}
	found := false
	rows, _ := env["data"].([]any)
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		if m["email"] == vetEmail || m["userId"] == vetID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("unassigned vet not in pool %#v (comm was %s)", env["data"], commID)
	}
}

// F4 — Existing client without cabinet + claim-invite vet QR → practice_clients linked.
func TestFiliationChain_ClaimVetInviteLinksExistingClient(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	vetID, practiceID, invCode := insertVetWithPractice(t, api, st, "fil-f4-vet", "Cabinet F4")

	clientEmail := uniqueEmail("fil-f4-cli")
	clientID := insertVerifiedUser(t, api, "client", clientEmail, "ClientDemo123!", "Client F4", nil)
	_ = st.EnsureUserProfiles(ctx, clientID)
	tok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": invCode,
	})
	if code != http.StatusOK {
		t.Fatalf("claim-invite %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["status"] != "linked" {
		t.Fatalf("status=%v want linked", data["status"])
	}
	if data["kind"] != "vet" {
		t.Fatalf("kind=%v", data["kind"])
	}

	if linkedVet := practiceClientVet(t, api, clientID, practiceID); linkedVet != vetID {
		t.Fatalf("vet_user_id=%s want %s", linkedVet, vetID)
	}
	if primary := clientPrimaryPractice(t, api, clientID); primary != practiceID {
		t.Fatalf("client practice_id=%s want %s", primary, practiceID)
	}
}

// F5 — Reclaim same practice → already_linked; multi-cabinet keeps A; same-cabinet colleague cannot steal vet_user_id.
func TestFiliationChain_ClaimVetInviteNoStealFirstLink(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	vetA, practiceA, codeA := insertVetWithPractice(t, api, st, "fil-f5-a", "Cabinet F5A")
	vetB, practiceB, codeB := insertVetWithPractice(t, api, st, "fil-f5-b", "Cabinet F5B")

	// Colleague C shares practice A (first-wins on practice_clients.vet_user_id).
	vetC := insertVerifiedUser(t, api, "vet", uniqueEmail("fil-f5-c"), "VetDemo123!", "Dr F5C", map[string]any{
		"practice_id": practiceA,
	})
	invC, err := st.EnsureAppInviteCode(ctx, vetC)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("fil-f5-cli")
	clientID := insertVerifiedUser(t, api, "client", clientEmail, "ClientDemo123!", "Client F5", nil)
	_ = st.EnsureUserProfiles(ctx, clientID)
	tok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": codeA,
	})
	if code != http.StatusOK || dataMap(t, env)["status"] != "linked" {
		t.Fatalf("first claim %d %#v", code, env)
	}

	// Same practice / same invite again → already_linked, A unchanged.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": codeA,
	})
	if code != http.StatusOK {
		t.Fatalf("reclaim A %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "already_linked" {
		t.Fatalf("reclaim status=%v want already_linked", dataMap(t, env)["status"])
	}
	if vetOnA := practiceClientVet(t, api, clientID, practiceA); vetOnA != vetA {
		t.Fatalf("practice A vet overwritten: %s want %s", vetOnA, vetA)
	}

	// Different practice B: multi-cabinet OK, but A preserved + primary practice stays A.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": codeB,
	})
	if code != http.StatusOK {
		t.Fatalf("claim B %d %#v", code, env)
	}
	dataB := dataMap(t, env)
	if dataB["status"] != "linked" || dataB["kind"] != "vet" {
		t.Fatalf("claim B status/kind=%v/%v want linked/vet", dataB["status"], dataB["kind"])
	}
	if vetOnA := practiceClientVet(t, api, clientID, practiceA); vetOnA != vetA {
		t.Fatalf("after claim B, A stolen: got %s want %s", vetOnA, vetA)
	}
	if vetOnB := practiceClientVet(t, api, clientID, practiceB); vetOnB != vetB {
		t.Fatalf("practice B link missing/wrong: %s want %s", vetOnB, vetB)
	}
	if primary := clientPrimaryPractice(t, api, clientID); primary != practiceA {
		t.Fatalf("primary practice_id=%s want A %s (must not steal primary)", primary, practiceA)
	}

	// Same practice, colleague C invite: already_linked + vet_user_id stays A (ON CONFLICT DO NOTHING).
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": invC.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim colleague C %d %#v", code, env)
	}
	dataC := dataMap(t, env)
	if dataC["status"] != "already_linked" {
		t.Fatalf("colleague claim status=%v want already_linked", dataC["status"])
	}
	if vetOnA := practiceClientVet(t, api, clientID, practiceA); vetOnA != vetA {
		t.Fatalf("colleague stole practice A vet: %s want %s (C=%s)", vetOnA, vetA, vetC)
	}
}

// F6 — standalone → commercial_referrals; linked → practice_clients + Resolve, no referral row.
func TestFiliationChain_CommercialCreateClientReferral(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, _, commTok := createCommercial(t, api, adminTok, "fil-f6-comm", "F6 Comm")

	vetEmail := uniqueEmail("fil-f6-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F6",
		"practiceName": "Cabinet F6",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)

	linkedEmail := uniqueEmail("fil-f6-linked")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", commTok, map[string]any{
		"email": linkedEmail, "password": "ClientDemo123!", "fullName": "Linked F6", "vetUserId": vetID,
	})
	if code != http.StatusCreated {
		t.Fatalf("linked client %d %#v", code, env)
	}
	linkedID := userIDByEmail(t, api, linkedEmail)
	var n int
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.practice_clients WHERE client_user_id=$1 AND vet_user_id=$2`,
		linkedID, vetID).Scan(&n)
	if n != 1 {
		t.Fatalf("practice_clients rows=%d want 1", n)
	}
	if commercialReferralCount(t, api, linkedID) != 0 {
		t.Fatalf("linked client must not get commercial_referrals (commission via vet assign)")
	}
	st := store.New(api.pool)
	var practiceID string
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(practice_id::text,'') FROM identity.users WHERE id=$1`, vetID).Scan(&practiceID)
	_, effComm, err := st.ResolveVetCommercial(context.Background(), linkedID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if effComm != commID {
		t.Fatalf("Resolve linked=%s want %s", effComm, commID)
	}

	soloEmail := uniqueEmail("fil-f6-solo")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", commTok, map[string]any{
		"email": soloEmail, "password": "ClientDemo123!", "fullName": "Solo F6",
	})
	if code != http.StatusCreated {
		t.Fatalf("standalone client %d %#v", code, env)
	}
	soloID := userIDByEmail(t, api, soloEmail)
	ref, _ := commercialReferralOf(t, api, soloID)
	if ref != commID {
		t.Fatalf("commercial_referrals=%s want %s (standalone filiation lost)", ref, commID)
	}
}

// F7 — Comm invite → register vet → QR vet → register-client → Resolve = comm.
func TestFiliationChain_CommVetClientEndToEnd(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commID, _, _ := createCommercial(t, api, adminTok, "fil-f7-comm", "F7 Comm")
	invComm, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}

	vetEmail := uniqueEmail("fil-f7-vet")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F7",
		"practiceName": "Cabinet F7", "consent": true, "inviteCode": invComm.Code,
	})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("vet register %d %#v", code, env)
	}
	if got := assignedCommercialID(t, api, vetEmail, true); got != commID {
		t.Fatalf("vet assign=%s want %s", got, commID)
	}
	vetID := userIDByEmail(t, api, vetEmail)
	verifyEmail(t, api, vetID)

	invVet, err := st.EnsureAppInviteCode(ctx, vetID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("fil-f7-cli")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Client F7",
		"consent": true, "inviteCode": invVet.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	clientID := userIDByEmail(t, api, clientEmail)

	var practiceID string
	_ = api.pool.QueryRow(ctx, `
		SELECT practice_id::text FROM identity.users WHERE id=$1`, vetID).Scan(&practiceID)
	if linkedVet := practiceClientVet(t, api, clientID, practiceID); linkedVet != vetID {
		t.Fatalf("linked vet=%s want %s", linkedVet, vetID)
	}

	_, effComm, err := st.ResolveVetCommercial(ctx, clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if effComm != commID {
		t.Fatalf("ResolveVetCommercial=%s want chain commercial %s", effComm, commID)
	}
}

// F8 — Client referred A + claim QR of vet assigned to A → referral stays A, Resolve = A.
func TestFiliationChain_ReferralPreservedWhenJoiningAssignedVet(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-f8-a", "F8 Comm A")

	clientID, _, practiceID := referredClientClaimsAssignedVet(t, api, commA, tokA, "fil-f8-cli", "fil-f8-vet")

	ref, _ := commercialReferralOf(t, api, clientID)
	if ref != commA {
		t.Fatalf("referral overwritten to %s want A %s", ref, commA)
	}
	_, eff, err := st.ResolveVetCommercial(context.Background(), clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if eff != commA {
		t.Fatalf("Resolve=%s want A %s", eff, commA)
	}
}

// F9 — Client referred A + claim QR of vet assigned to C → referral row stays A; Resolve = C (assign wins).
func TestFiliationChain_AssignPriorityOverReferralRow(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commA, _, _ := createCommercial(t, api, adminTok, "fil-f9-a", "F9 Comm A")
	commC, _, tokC := createCommercial(t, api, adminTok, "fil-f9-c", "F9 Comm C")

	clientID, _, practiceID := referredClientClaimsAssignedVet(t, api, commA, tokC, "fil-f9-cli", "fil-f9-vet")

	ref, _ := commercialReferralOf(t, api, clientID)
	if ref != commA {
		t.Fatalf("referral row lost/stolen: %s want A %s (assign priority must not erase row)", ref, commA)
	}
	_, eff, err := st.ResolveVetCommercial(context.Background(), clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if eff != commC {
		t.Fatalf("Resolve=%s want C %s (vet assign priority)", eff, commC)
	}
}

// F10 — Unassign after F8-like setup → Resolve falls back to commercial_referrals.
func TestFiliationChain_UnassignFallsBackToReferral(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-f10-a", "F10 Comm A")

	clientID, vetID, practiceID := referredClientClaimsAssignedVet(t, api, commA, tokA, "fil-f10-cli", "fil-f10-vet")

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/vets/"+vetID+"/unassign", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassign %d %#v", code, env)
	}
	if got := assignedCommercialID(t, api, vetID, false); got != "" {
		t.Fatalf("assign still set: %s", got)
	}

	ref, _ := commercialReferralOf(t, api, clientID)
	if ref != commA {
		t.Fatalf("referral lost after unassign: %s", ref)
	}
	_, eff, err := st.ResolveVetCommercial(context.Background(), clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if eff != commA {
		t.Fatalf("Resolve after unassign=%s want fallback referral A %s", eff, commA)
	}
}

// F10b — After F9 (referral A, assign C), unassign → Resolve falls back to A (not silent 0).
func TestFiliationChain_UnassignAfterAssignPriorityFallsBackToReferral(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commA, _, _ := createCommercial(t, api, adminTok, "fil-f10b-a", "F10b Comm A")
	_, _, tokC := createCommercial(t, api, adminTok, "fil-f10b-c", "F10b Comm C")

	clientID, vetID, practiceID := referredClientClaimsAssignedVet(t, api, commA, tokC, "fil-f10b-cli", "fil-f10b-vet")

	_, effBefore, err := st.ResolveVetCommercial(context.Background(), clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if effBefore == "" || effBefore == commA {
		t.Fatalf("precondition: Resolve should be C before unassign, got %s", effBefore)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/vets/"+vetID+"/unassign", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassign %d %#v", code, env)
	}
	_, eff, err := st.ResolveVetCommercial(context.Background(), clientID, practiceID)
	if err != nil {
		t.Fatal(err)
	}
	if eff != commA {
		t.Fatalf("Resolve after unassign (was C)=%s want fallback A %s", eff, commA)
	}
}

// F11 — Multi-cabinet: Resolve is practice-scoped; empty practiceID = latest global link.
func TestFiliationChain_MultiCabinetResolveScopedByPractice(t *testing.T) {
	api := newTestAPI(t)
	st := store.New(api.pool)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-f11-a", "F11 Comm A")
	commB, _, tokB := createCommercial(t, api, adminTok, "fil-f11-b", "F11 Comm B")

	vetEmailA := uniqueEmail("fil-f11-va")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": vetEmailA, "password": "VetDemo123!", "fullName": "Dr F11A",
		"practiceName": "Cabinet F11A",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode A %d %#v", code, env)
	}
	vetA, _ := dataMap(t, env)["userId"].(string)

	vetEmailB := uniqueEmail("fil-f11-vb")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokB, map[string]any{
		"email": vetEmailB, "password": "VetDemo123!", "fullName": "Dr F11B",
		"practiceName": "Cabinet F11B",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode B %d %#v", code, env)
	}
	vetB, _ := dataMap(t, env)["userId"].(string)

	ctx := context.Background()
	invA, err := st.EnsureAppInviteCode(ctx, vetA)
	if err != nil {
		t.Fatal(err)
	}
	invB, err := st.EnsureAppInviteCode(ctx, vetB)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("fil-f11-cli")
	clientID := insertVerifiedUser(t, api, "client", clientEmail, "ClientDemo123!", "Client F11", nil)
	_ = st.EnsureUserProfiles(ctx, clientID)
	tok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": invA.Code,
	})
	if code != http.StatusOK || dataMap(t, env)["status"] != "linked" {
		t.Fatalf("claim A %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": invB.Code,
	})
	if code != http.StatusOK || dataMap(t, env)["status"] != "linked" {
		t.Fatalf("claim B %d %#v", code, env)
	}

	var practiceA, practiceB string
	_ = api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE id=$1`, vetA).Scan(&practiceA)
	_ = api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE id=$1`, vetB).Scan(&practiceB)

	_, effA, err := st.ResolveVetCommercial(ctx, clientID, practiceA)
	if err != nil {
		t.Fatal(err)
	}
	if effA != commA {
		t.Fatalf("Resolve(P_A)=%s want A %s", effA, commA)
	}
	_, effB, err := st.ResolveVetCommercial(ctx, clientID, practiceB)
	if err != nil {
		t.Fatal(err)
	}
	if effB != commB {
		t.Fatalf("Resolve(P_B)=%s want B %s", effB, commB)
	}
	_, effGlobal, err := st.ResolveVetCommercial(ctx, clientID, "")
	if err != nil {
		t.Fatal(err)
	}
	if effGlobal != commB {
		t.Fatalf("Resolve(\"\")=%s want latest B %s", effGlobal, commB)
	}

	// List Effectif on each commercial's part_a row matches Resolve(practice).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokA, nil)
	if code != http.StatusOK {
		t.Fatalf("A filiation %d %#v", code, env)
	}
	rowsA := filiationItems(t, env)
	if !filiationRowMatch(rowsA, func(m map[string]any) bool {
		return m["clientUserId"] == clientID &&
			m["practiceId"] == practiceA &&
			m["effectiveCommercialId"] == commA &&
			m["effectiveSource"] == store.FiliationSourceVetAssignment
	}) {
		t.Fatalf("A list Effectif vs Resolve(P_A) %#v", rowsA)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokB, nil)
	if code != http.StatusOK {
		t.Fatalf("B filiation %d %#v", code, env)
	}
	rowsB := filiationItems(t, env)
	if !filiationRowMatch(rowsB, func(m map[string]any) bool {
		return m["clientUserId"] == clientID &&
			m["practiceId"] == practiceB &&
			m["effectiveCommercialId"] == commB &&
			m["effectiveSource"] == store.FiliationSourceVetAssignment
	}) {
		t.Fatalf("B list Effectif vs Resolve(P_B) %#v", rowsB)
	}
}

// F12 — RGPD export includes commercialReferrals (+ clientReferrals key) for referred clients.
func TestFiliationChain_ExportIncludesReferrals(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, _, tok := createCommercial(t, api, adminTok, "fil-f12-comm", "F12 Comm")

	clientEmail := uniqueEmail("fil-f12-cli")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", tok, map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Client F12",
	})
	if code != http.StatusCreated {
		t.Fatalf("standalone client %d %#v", code, env)
	}
	clientID := userIDByEmail(t, api, clientEmail)
	ref, _ := commercialReferralOf(t, api, clientID)
	if ref != commID {
		t.Fatalf("referral=%s want %s", ref, commID)
	}

	access := loginToken(t, api.handler, clientEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", access, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	data := dataMap(t, env)
	cr, ok := data["commercialReferrals"].([]any)
	if !ok || len(cr) == 0 {
		t.Fatalf("export missing commercialReferrals: %#v", data["commercialReferrals"])
	}
	found := false
	for _, raw := range cr {
		m, _ := raw.(map[string]any)
		if m == nil {
			continue
		}
		if fmt.Sprint(m["commercial_user_id"]) == commID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("commercialReferrals does not include %s: %#v", commID, cr)
	}
	if _, ok := data["clientReferrals"]; !ok {
		t.Fatalf("export missing clientReferrals key: %#v", data)
	}
	if _, ok := data["filiationEvents"]; !ok {
		t.Fatalf("export missing filiationEvents key: %#v", data)
	}
	fe, _ := data["filiationEvents"].([]any)
	if len(fe) == 0 {
		t.Fatalf("export filiationEvents empty, want client_referral event: %#v", data["filiationEvents"])
	}
	foundEvent := false
	for _, raw := range fe {
		m, _ := raw.(map[string]any)
		if m == nil {
			continue
		}
		if fmt.Sprint(m["event_type"]) == store.FiliationEventClientReferral ||
			fmt.Sprint(m["eventType"]) == store.FiliationEventClientReferral {
			foundEvent = true
			break
		}
	}
	if !foundEvent {
		t.Fatalf("export filiationEvents missing client_referral: %#v", fe)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me %d %#v", code, env)
	}
	if n := commercialReferralCount(t, api, clientID); n != 0 {
		t.Fatalf("commercial_referrals after purge count=%d want 0", n)
	}
	var eventN int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.filiation_events
		WHERE client_user_id=$1 OR actor_user_id=$1`, clientID).Scan(&eventN); err != nil {
		t.Fatal(err)
	}
	if eventN != 0 {
		t.Fatalf("filiation_events after purge count=%d want 0", eventN)
	}
}

// F13 — Multi-cabinet Accrue: pet at practice A → ledger CommA; pet at B → CommB.
func TestFiliationChain_AccruePaysPracticeScopedCommercial(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	if err := st.EnsureDefaultCommissionTiers(ctx); err != nil {
		t.Fatal(err)
	}
	_ = st.EnsureCommissionSettings(ctx)
	_ = st.SetCommercialRateBps(ctx, store.DefaultCommercialCommissionRateBps)

	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-f13-a", "F13 Comm A")
	commB, _, tokB := createCommercial(t, api, adminTok, "fil-f13-b", "F13 Comm B")

	vetEmailA := uniqueEmail("fil-f13-va")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": vetEmailA, "password": "VetDemo123!", "fullName": "Dr F13A",
		"practiceName": "Cabinet F13A",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode A %d %#v", code, env)
	}
	vetA, _ := dataMap(t, env)["userId"].(string)

	vetEmailB := uniqueEmail("fil-f13-vb")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokB, map[string]any{
		"email": vetEmailB, "password": "VetDemo123!", "fullName": "Dr F13B",
		"practiceName": "Cabinet F13B",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode B %d %#v", code, env)
	}
	vetB, _ := dataMap(t, env)["userId"].(string)

	invA, err := st.EnsureAppInviteCode(ctx, vetA)
	if err != nil {
		t.Fatal(err)
	}
	invB, err := st.EnsureAppInviteCode(ctx, vetB)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("fil-f13-cli")
	clientID := insertVerifiedUser(t, api, "client", clientEmail, "ClientDemo123!", "Client F13", nil)
	_ = st.EnsureUserProfiles(ctx, clientID)
	tok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{"code": invA.Code})
	if code != http.StatusOK {
		t.Fatalf("claim A %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{"code": invB.Code})
	if code != http.StatusOK {
		t.Fatalf("claim B %d %#v", code, env)
	}

	var practiceA, practiceB string
	_ = api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE id=$1`, vetA).Scan(&practiceA)
	_ = api.pool.QueryRow(ctx, `SELECT practice_id::text FROM identity.users WHERE id=$1`, vetB).Scan(&practiceB)

	activateAndAccrue := func(practiceID, petName string) (entID string) {
		t.Helper()
		petID := uuid.NewString()
		if _, err := api.pool.Exec(ctx, `
			INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
			VALUES ($1, $2, $3, $4, 'dog', 'lab', 'pending_payment')`,
			petID, practiceID, clientID, petName); err != nil {
			t.Fatal(err)
		}
		baseCents := 3500
		ent, err := st.CreateEntitlement(ctx, petID, clientID, "annual", "subscription", baseCents)
		if err != nil {
			t.Fatal(err)
		}
		now := time.Now().UTC()
		if err := st.ActivateEntitlement(ctx, store.ActivateEntitlementParams{
			PetID: petID, Status: "active", ValidFrom: now, ValidUntil: now.Add(365 * 24 * time.Hour),
		}); err != nil {
			t.Fatal(err)
		}
		if err := st.AccrueCommissionForPetActivation(ctx, petID); err != nil {
			t.Fatal(err)
		}
		return ent.ID
	}

	entA := activateAndAccrue(practiceA, "Pet F13A")
	entB := activateAndAccrue(practiceB, "Pet F13B")

	ledgerComm := func(entID string) string {
		t.Helper()
		var commercialID string
		if err := api.pool.QueryRow(ctx, `
			SELECT commercial_user_id::text
			FROM billing.commercial_commission_ledger
			WHERE source_id=$1 AND source_type='subscription_pct'`, entID).Scan(&commercialID); err != nil {
			t.Fatalf("ledger for %s: %v", entID, err)
		}
		return commercialID
	}
	if got := ledgerComm(entA); got != commA {
		t.Fatalf("Accrue practice A commercial=%s want %s", got, commA)
	}
	if got := ledgerComm(entB); got != commB {
		t.Fatalf("Accrue practice B commercial=%s want %s", got, commB)
	}
}

// F14 — Audit trail: encode → vet_assigned; standalone client → client_referral; unassign → vet_unassigned.
func TestFiliationChain_AuditTrailEvents(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, _, tok := createCommercial(t, api, adminTok, "fil-f14-comm", "F14 Comm")

	vetEmail := uniqueEmail("fil-f14-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr F14",
		"practiceName": "Cabinet F14",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation/events", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("events %d %#v", code, env)
	}
	events := filiationEventItems(t, env)
	if !filiationEventHas(events, store.FiliationEventVetAssigned, vetID) {
		t.Fatalf("missing vet_assigned for %s %#v", vetID, events)
	}

	cliEmail := uniqueEmail("fil-f14-cli")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", tok, map[string]any{
		"email": cliEmail, "password": "ClientDemo123!", "fullName": "Client F14",
	})
	if code != http.StatusCreated {
		t.Fatalf("standalone %d %#v", code, env)
	}
	clientID := userIDByEmail(t, api, cliEmail)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation/events", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("events2 %d %#v", code, env)
	}
	events = filiationEventItems(t, env)
	if !filiationEventMatch(events, func(m map[string]any) bool {
		return m["eventType"] == store.FiliationEventClientReferral && m["clientUserId"] == clientID
	}) {
		t.Fatalf("missing client_referral %#v", events)
	}

	// practice_client_linked via claim QR of assigned vet.
	st := store.New(api.pool)
	inv, err := st.EnsureAppInviteCode(context.Background(), vetID)
	if err != nil {
		t.Fatal(err)
	}
	// Re-assign after we will unassign later — encode already assigned; claim uses that vet.
	// First re-assign for claim (still assigned from encode).
	verifyEmail(t, api, clientID)
	cliTok := loginToken(t, api.handler, cliEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", cliTok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation/events", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("events3 %d %#v", code, env)
	}
	events = filiationEventItems(t, env)
	if !filiationEventMatch(events, func(m map[string]any) bool {
		return m["eventType"] == store.FiliationEventPracticeClientLinked && m["clientUserId"] == clientID
	}) {
		t.Fatalf("missing practice_client_linked %#v", events)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/vets/"+vetID+"/unassign", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassign %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation/events?commercialId="+commID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin events %d %#v", code, env)
	}
	events = filiationEventItems(t, env)
	if !filiationEventHas(events, store.FiliationEventVetUnassigned, vetID) {
		t.Fatalf("missing vet_unassigned %#v", events)
	}
}

// Pro DELETE /me (tombstone) purges filiation_events and clears live attribution
// (assigned_commercial_id + commercial_referrals) so Accrue cannot pay a dead commercial.
func TestFiliationChain_ProDeletePurgesEvents(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commID, email, tok := createCommercial(t, api, adminTok, "fil-pro-del", "Pro Del Comm")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tok, map[string]any{
		"email": uniqueEmail("fil-pro-del-v"), "password": "VetDemo123!", "fullName": "Dr ProDel",
		"practiceName": "Cab ProDel",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)
	var nBefore int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.filiation_events WHERE commercial_user_id=$1`, commID).Scan(&nBefore); err != nil {
		t.Fatal(err)
	}
	if nBefore == 0 {
		t.Fatal("expected events before pro delete")
	}

	// Standalone referral also cleared on commercial tombstone.
	cliEmail := uniqueEmail("fil-pro-del-cli")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", tok, map[string]any{
		"email": cliEmail, "password": "ClientDemo123!", "fullName": "Cli ProDel",
	})
	if code != http.StatusCreated {
		t.Fatalf("create client %d %#v", code, env)
	}
	var refCount int
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.commercial_referrals WHERE commercial_user_id=$1`, commID).Scan(&refCount)
	if refCount == 0 {
		t.Fatal("expected commercial_referrals before delete")
	}

	access := loginToken(t, api.handler, email, "CommercialDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me pro %d %#v", code, env)
	}
	var nAfter int
	if err := api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.filiation_events
		WHERE commercial_user_id=$1 OR actor_user_id=$1`, commID).Scan(&nAfter); err != nil {
		t.Fatal(err)
	}
	if nAfter != 0 {
		t.Fatalf("filiation_events after pro tombstone count=%d want 0", nAfter)
	}
	var assignedLeft string
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`, vetID).Scan(&assignedLeft)
	if assignedLeft != "" {
		t.Fatalf("assigned_commercial_id still set after commercial tombstone: %s", assignedLeft)
	}
	var refAfter int
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM practice.commercial_referrals WHERE commercial_user_id=$1`, commID).Scan(&refAfter)
	if refAfter != 0 {
		t.Fatalf("commercial_referrals after tombstone count=%d want 0", refAfter)
	}
}

// Admin re-assign records vet_unassigned for previous commercial then vet_assigned for new.
func TestFiliationChain_ReassignEmitsUnassign(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-reass-a", "Reass A")
	commB, _, _ := createCommercial(t, api, adminTok, "fil-reass-b", "Reass B")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": uniqueEmail("fil-reass-v"), "password": "VetDemo123!", "fullName": "Dr Reass",
		"practiceName": "Cab Reass",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/commercials/"+commB+"/assign", adminTok, map[string]any{
		"vetUserId": vetID,
	})
	if code != http.StatusOK {
		t.Fatalf("reassign %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation/events?commercialId="+commA, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("events A %d %#v", code, env)
	}
	eventsA := filiationEventItems(t, env)
	if !filiationEventHas(eventsA, store.FiliationEventVetUnassigned, vetID) {
		t.Fatalf("missing vet_unassigned for A %#v", eventsA)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation/events?commercialId="+commB, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("events B %d %#v", code, env)
	}
	eventsB := filiationEventItems(t, env)
	if !filiationEventHas(eventsB, store.FiliationEventVetAssigned, vetID) {
		t.Fatalf("missing vet_assigned for B %#v", eventsB)
	}
}

func filiationEventItems(t *testing.T, env map[string]any) []any {
	t.Helper()
	data := env["data"]
	if m, ok := data.(map[string]any); ok {
		items, _ := m["items"].([]any)
		if items == nil {
			return []any{}
		}
		return items
	}
	if arr, ok := data.([]any); ok {
		return arr
	}
	t.Fatalf("filiation events data shape: %#v", data)
	return nil
}

func filiationEventHas(events []any, eventType, vetID string) bool {
	return filiationEventMatch(events, func(m map[string]any) bool {
		return m["eventType"] == eventType && m["vetUserId"] == vetID
	})
}

func filiationEventMatch(events []any, pred func(map[string]any) bool) bool {
	for _, raw := range events {
		m, _ := raw.(map[string]any)
		if pred(m) {
			return true
		}
	}
	return false
}
