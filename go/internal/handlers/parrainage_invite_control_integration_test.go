package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// Contrôle anti-régression : le code parrainage (QR / invite) ne doit ni se perdre
// ni rattacher le mauvais promoteur (véto / care_pro / commercial).

func insertVerifiedUser(t *testing.T, api *testAPI, role, email, password, fullName string, extraCols map[string]any) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	cols := []string{"id", "email", "password_hash", "full_name", "role", "email_verified_at", "terms_accepted_at"}
	vals := []any{id, email, string(hash), fullName, role}
	placeholders := []string{"$1", "$2", "$3", "$4", "$5", "NOW()", "NOW()"}
	i := 6
	for k, v := range extraCols {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	q := fmt.Sprintf(`INSERT INTO identity.users (%s) VALUES (%s)`,
		strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	if _, err := api.pool.Exec(context.Background(), q, vals...); err != nil {
		t.Fatalf("insert user %s: %v", email, err)
	}
	return id
}

func TestParrainageControl_CodeDurableAndIssuer(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	commEmail := uniqueEmail("parr-comm")
	commID := insertVerifiedUser(t, api, "commercial", commEmail, "CommercialDemo123!", "Parr Comm", nil)
	inv1, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}
	inv2, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}
	if inv1.Code == "" || inv1.Code != inv2.Code {
		t.Fatalf("code must be durable: %q vs %q", inv1.Code, inv2.Code)
	}
	if inv1.UserID != commID || inv1.Role != "commercial" {
		t.Fatalf("issuer mismatch: %#v", inv1)
	}

	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/app-invite", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me/app-invite %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["code"] != inv1.Code {
		t.Fatalf("API code=%v want %s", data["code"], inv1.Code)
	}
	if data["role"] != "commercial" {
		t.Fatalf("role=%v", data["role"])
	}
	if data["vetRegisterUrl"] == nil || data["vetRegisterUrl"] == "" {
		t.Fatal("commercial QR must expose vetRegisterUrl")
	}
	if !strings.Contains(fmt.Sprint(data["vetRegisterUrl"]), "invite="+inv1.Code) {
		t.Fatalf("vetRegisterUrl must embed code: %v", data["vetRegisterUrl"])
	}
	if !strings.Contains(fmt.Sprint(data["inviteUrl"]), "/invite/"+inv1.Code) {
		t.Fatalf("inviteUrl must point to landing: %v", data["inviteUrl"])
	}

	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/app-invite/"+inv1.Code, nil)
	if code != http.StatusOK {
		t.Fatalf("public invite %d %#v", code, env)
	}
	pub := dataMap(t, env)
	if pub["code"] != inv1.Code || pub["role"] != "commercial" {
		t.Fatalf("public payload %#v", pub)
	}
}

func TestParrainageControl_VetInviteLinksClientToPromoter(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	// Seeded demo vet has practice — use it for a durable invite.
	var vetID, practiceID string
	err := api.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text FROM identity.users
		WHERE email='vet.demo@petsfollow.test'`).Scan(&vetID, &practiceID)
	if err != nil || practiceID == "" {
		t.Skip("vet.demo seed required")
	}
	inv, err := st.EnsureAppInviteCode(ctx, vetID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("parr-vet-cli")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Client Vet Invite",
		"consent": true, "inviteCode": strings.ToLower(inv.Code), // normalisation casse
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	if dataMap(t, env)["inviteStatus"] != "linked" {
		t.Fatalf("inviteStatus=%v want linked", dataMap(t, env)["inviteStatus"])
	}

	var clientID, linkedPractice, linkedVet string
	if err := api.pool.QueryRow(ctx, `SELECT id::text, COALESCE(practice_id::text,'') FROM identity.users WHERE email=$1`, clientEmail).
		Scan(&clientID, &linkedPractice); err != nil {
		t.Fatal(err)
	}
	if linkedPractice != practiceID {
		t.Fatalf("practice_id=%s want %s", linkedPractice, practiceID)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM practice.practice_clients
		WHERE practice_id=$1 AND client_user_id=$2`, practiceID, clientID).Scan(&linkedVet); err != nil {
		t.Fatal(err)
	}
	if linkedVet != vetID {
		t.Fatalf("vet_user_id=%s want promoter %s", linkedVet, vetID)
	}
}

func TestParrainageControl_CareProInviteGrantsAccess(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	careEmail := uniqueEmail("parr-care")
	careID := insertVerifiedUser(t, api, "care_pro", careEmail, "CareProDemo123!", "Parr Care", map[string]any{
		"professional_specialty": "farrier",
	})
	_ = st.EnsureUserProfiles(ctx, careID)
	inv, err := st.EnsureAppInviteCode(ctx, careID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("parr-care-cli")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Client Care",
		"consent": true, "inviteCode": inv.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	if dataMap(t, env)["inviteStatus"] != "granted" {
		t.Fatalf("inviteStatus=%v want granted", dataMap(t, env)["inviteStatus"])
	}

	var clientID string
	_ = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	var n int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.client_access
		WHERE client_user_id=$1 AND grantee_user_id=$2
		  AND (expires_at IS NULL OR expires_at > NOW())`,
		clientID, careID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("care_pro access rows=%d want 1", n)
	}
}

func TestParrainageControl_CommercialFirstWinsAndInviteBeatsNearby(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	aID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-a"), "CommercialDemo123!", "Comm A", nil)
	bID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-b"), "CommercialDemo123!", "Comm B", nil)
	invA, err := st.EnsureAppInviteCode(ctx, aID)
	if err != nil {
		t.Fatal(err)
	}
	invB, err := st.EnsureAppInviteCode(ctx, bID)
	if err != nil {
		t.Fatal(err)
	}

	// Invite A wins even if nearby B is also sent.
	clientEmail := uniqueEmail("parr-first")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "First Wins",
		"consent": true, "inviteCode": invA.Code, "commercialUserId": bID,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	if dataMap(t, env)["inviteStatus"] != "referred" {
		t.Fatalf("status=%v", dataMap(t, env)["inviteStatus"])
	}
	var clientID, referral, savedCode string
	_ = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	if err := api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text, COALESCE(invite_code,'')
		FROM practice.commercial_referrals WHERE client_user_id=$1`, clientID).
		Scan(&referral, &savedCode); err != nil {
		t.Fatal(err)
	}
	if referral != aID {
		t.Fatalf("referral=%s want A %s (invite must beat nearby)", referral, aID)
	}
	if savedCode != invA.Code {
		t.Fatalf("invite_code lost: got %q want %q", savedCode, invA.Code)
	}

	// Second commercial code must not steal attribution.
	_, _ = api.pool.Exec(ctx, `UPDATE identity.users SET email_verified_at=NOW() WHERE id=$1`, clientID)
	clientTok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", clientTok, map[string]any{
		"code": invB.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("second claim %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["status"] != "already_linked" {
		t.Fatalf("second claim status=%v want already_linked", data["status"])
	}
	if data["inviterId"] != aID {
		t.Fatalf("inviterId=%v want original promoter A %s (must not report B)", data["inviterId"], aID)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&referral); err != nil {
		t.Fatal(err)
	}
	if referral != aID {
		t.Fatalf("attribution overwritten to %s", referral)
	}
}

func TestParrainageControl_NearbyThenSameInviteBackfillsCode(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-near"), "CommercialDemo123!", "Near Comm", nil)
	inv, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}

	clientEmail := uniqueEmail("parr-near-cli")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Nearby First",
		"consent": true, "commercialUserId": commID,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	var clientID, savedCode string
	_ = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	_ = api.pool.QueryRow(ctx, `
		SELECT COALESCE(invite_code,'') FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&savedCode)
	if savedCode != "" {
		t.Fatalf("nearby link should start without code, got %q", savedCode)
	}

	_, _ = api.pool.Exec(ctx, `UPDATE identity.users SET email_verified_at=NOW() WHERE id=$1`, clientID)
	clientTok := loginToken(t, api.handler, clientEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", clientTok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim %d %#v", code, env)
	}
	_ = api.pool.QueryRow(ctx, `
		SELECT COALESCE(invite_code,'') FROM practice.commercial_referrals WHERE client_user_id=$1`,
		clientID).Scan(&savedCode)
	if savedCode != inv.Code {
		t.Fatalf("invite_code not backfilled: got %q want %q", savedCode, inv.Code)
	}
}

func TestParrainageControl_VetRegisterOnlyAcceptsSalesCode(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-reg-c"), "CommercialDemo123!", "Reg Comm", nil)
	invComm, err := st.EnsureAppInviteCode(ctx, commID)
	if err != nil {
		t.Fatal(err)
	}

	var vetID string
	err = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email='vet.demo@petsfollow.test'`).Scan(&vetID)
	if err != nil {
		t.Skip("vet.demo required")
	}
	invVet, err := st.EnsureAppInviteCode(ctx, vetID)
	if err != nil {
		t.Fatal(err)
	}

	okEmail := uniqueEmail("parr-vet-ok")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": okEmail, "password": "VetDemo123!", "fullName": "Véto OK",
		"practiceName": "Cabinet Parrain OK", "consent": true, "inviteCode": invComm.Code,
	})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("register with sales code %d %#v", code, env)
	}
	var assigned string
	_ = api.pool.QueryRow(ctx, `SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`, okEmail).Scan(&assigned)
	if assigned != commID {
		t.Fatalf("assigned_commercial_id=%s want %s", assigned, commID)
	}

	badEmail := uniqueEmail("parr-vet-bad")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": badEmail, "password": "VetDemo123!", "fullName": "Véto Bad",
		"practiceName": "Cabinet Bad", "consent": true, "inviteCode": invVet.Code,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("vet code on register want 400, got %d %#v", code, env)
	}

	poolEmail := uniqueEmail("parr-vet-pool")
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": poolEmail, "password": "VetDemo123!", "fullName": "Véto Pool",
		"practiceName": "Cabinet Pool2", "consent": true,
		"assignedCommercialId": commID, // must be ignored without inviteCode
	})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("register without code %d %#v", code, env)
	}
	_ = api.pool.QueryRow(ctx, `SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE email=$1`, poolEmail).Scan(&assigned)
	if assigned != "" {
		t.Fatalf("nearby/assignedCommercialId must not assign without invite code, got %s", assigned)
	}
}

func TestParrainageControl_ClientInviteIssuanceAndSelfReject(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	sponsorEmail := uniqueEmail("parr-cli-sp")
	sponsorID := insertVerifiedUser(t, api, "client", sponsorEmail, "ClientDemo123!", "Sponsor", nil)
	_ = st.EnsureUserProfiles(ctx, sponsorID)

	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Role != "client" || inv.Code == "" {
		t.Fatalf("invite %#v", inv)
	}

	tok := loginToken(t, api.handler, sponsorEmail, "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/app-invite", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("client app-invite %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["code"] != inv.Code || data["role"] != "client" {
		t.Fatalf("payload %#v", data)
	}
	if data["vetRegisterUrl"] != nil && fmt.Sprint(data["vetRegisterUrl"]) != "" {
		t.Fatalf("client must not expose vetRegisterUrl: %#v", data["vetRegisterUrl"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("self claim want 400, got %d %#v", code, env)
	}
	msg := fmt.Sprint(env)
	if !strings.Contains(msg, "self_referral") {
		t.Fatalf("want self_referral, got %#v", env)
	}
}

func TestParrainageControl_ClientInviteReferralAndCabinet(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	var vetID, practiceID string
	var prevAssigned *string
	err := api.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text, assigned_commercial_id::text
		FROM identity.users u WHERE email='vet.demo@petsfollow.test'`).Scan(&vetID, &practiceID, &prevAssigned)
	if err != nil || practiceID == "" {
		t.Skip("vet.demo seed required")
	}
	t.Cleanup(func() {
		if prevAssigned == nil || *prevAssigned == "" {
			_, _ = api.pool.Exec(context.Background(), `
				UPDATE identity.users SET assigned_commercial_id=NULL WHERE id=$1`, vetID)
		} else {
			_, _ = api.pool.Exec(context.Background(), `
				UPDATE identity.users SET assigned_commercial_id=$2::uuid WHERE id=$1`, vetID, *prevAssigned)
		}
	})

	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-cli-c"), "CommercialDemo123!", "Cli Comm", nil)
	_, _ = api.pool.Exec(ctx, `UPDATE identity.users SET assigned_commercial_id=$2 WHERE id=$1`, vetID, commID)

	sponsorEmail := uniqueEmail("parr-cli-sp2")
	sponsorID := insertVerifiedUser(t, api, "client", sponsorEmail, "ClientDemo123!", "Sponsor Cab", map[string]any{
		"practice_id": practiceID,
	})
	_ = st.EnsureUserProfiles(ctx, sponsorID)
	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceID, sponsorID, vetID)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}

	filleulEmail := uniqueEmail("parr-cli-fil")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": filleulEmail, "password": "ClientDemo123!", "fullName": "Filleul",
		"consent": true, "inviteCode": inv.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	status := dataMap(t, env)["inviteStatus"]
	if status != "linked" && status != "referred" {
		t.Fatalf("inviteStatus=%v", status)
	}

	var filleulID, sponsorLinked, commercialLinked, linkedVet string
	_ = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, filleulEmail).Scan(&filleulID)
	err = api.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		filleulID).Scan(&sponsorLinked)
	if err != nil || sponsorLinked != sponsorID {
		t.Fatalf("client_referrals sponsor=%s err=%v want %s", sponsorLinked, err, sponsorID)
	}
	_ = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		filleulID).Scan(&commercialLinked)
	if commercialLinked != commID {
		t.Fatalf("inherited commercial=%s want %s", commercialLinked, commID)
	}
	err = api.pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM practice.practice_clients WHERE client_user_id=$1 AND practice_id=$2`,
		filleulID, practiceID).Scan(&linkedVet)
	if err != nil {
		t.Fatalf("expected cabinet membership: %v", err)
	}
	if linkedVet != vetID {
		t.Fatalf("vet_user_id=%s want %s (not sponsor)", linkedVet, vetID)
	}
	if linkedVet == sponsorID {
		t.Fatal("vet must never be client sponsor id")
	}

	// First-wins parrain: second sponsor code must not overwrite.
	sponsorB := insertVerifiedUser(t, api, "client", uniqueEmail("parr-cli-spb"), "ClientDemo123!", "Sponsor B", nil)
	_ = st.EnsureUserProfiles(ctx, sponsorB)
	invB, err := st.EnsureAppInviteCode(ctx, sponsorB)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = api.pool.Exec(ctx, `UPDATE identity.users SET email_verified_at=NOW() WHERE id=$1`, filleulID)
	filleulTok := loginToken(t, api.handler, filleulEmail, "ClientDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", filleulTok, map[string]any{
		"code": invB.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("second claim %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "already_linked" {
		t.Fatalf("status=%v", dataMap(t, env)["status"])
	}
	_ = api.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		filleulID).Scan(&sponsorLinked)
	if sponsorLinked != sponsorID {
		t.Fatalf("sponsor overwritten to %s", sponsorLinked)
	}
}

func TestParrainageControl_ClientInviteKeepsExistingCommercial(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	commA := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-keep-a"), "CommercialDemo123!", "Keep A", nil)
	commB := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-keep-b"), "CommercialDemo123!", "Keep B", nil)

	sponsorID := insertVerifiedUser(t, api, "client", uniqueEmail("parr-keep-sp"), "ClientDemo123!", "Keep Sp", nil)
	_ = st.EnsureUserProfiles(ctx, sponsorID)
	_, err := api.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, 'KEEPSP01', NOW())`, sponsorID, commB)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}

	filleulEmail := uniqueEmail("parr-keep-fil")
	filleulID := insertVerifiedUser(t, api, "client", filleulEmail, "ClientDemo123!", "Keep Fil", nil)
	_ = st.EnsureUserProfiles(ctx, filleulID)
	_, err = api.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, 'KEEPA001', NOW())`, filleulID, commA)
	if err != nil {
		t.Fatal(err)
	}

	tok := loginToken(t, api.handler, filleulEmail, "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim %d %#v", code, env)
	}
	var sponsorLinked, gotComm string
	err = api.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		filleulID).Scan(&sponsorLinked)
	if err != nil || sponsorLinked != sponsorID {
		t.Fatalf("sponsor=%s err=%v", sponsorLinked, err)
	}
	err = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		filleulID).Scan(&gotComm)
	if err != nil || gotComm != commA {
		t.Fatalf("commercial must stay A=%s, got %s err=%v", commA, gotComm, err)
	}
}

func TestParrainageControl_ClientInviteNoCabinetSteal(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	var vetA, practiceA string
	err := api.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text FROM identity.users WHERE email='vet.demo@petsfollow.test'`).Scan(&vetA, &practiceA)
	if err != nil || practiceA == "" {
		t.Skip("vet.demo seed required")
	}
	var vetB, practiceB string
	err = api.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text FROM identity.users WHERE email='vet.parc@petsfollow.test'`).Scan(&vetB, &practiceB)
	if err != nil || practiceB == "" || practiceB == practiceA {
		t.Skip("vet.parc seed required")
	}

	sponsorID := insertVerifiedUser(t, api, "client", uniqueEmail("parr-steal-sp"), "ClientDemo123!", "Sp Steal", map[string]any{
		"practice_id": practiceA,
	})
	_ = st.EnsureUserProfiles(ctx, sponsorID)
	_, _ = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceA, sponsorID, vetA)
	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}

	filleulEmail := uniqueEmail("parr-steal-fil")
	filleulID := insertVerifiedUser(t, api, "client", filleulEmail, "ClientDemo123!", "Fil Steal", map[string]any{
		"practice_id": practiceB,
	})
	_ = st.EnsureUserProfiles(ctx, filleulID)
	_, _ = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceB, filleulID, vetB)

	tok := loginToken(t, api.handler, filleulEmail, "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
		"code": inv.Code,
	})
	if code != http.StatusOK {
		t.Fatalf("claim %d %#v", code, env)
	}
	var n int
	_ = api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.practice_clients WHERE client_user_id=$1 AND practice_id=$2`,
		filleulID, practiceA).Scan(&n)
	if n != 0 {
		t.Fatalf("must not steal/add cabinet via client QR when already attached, got %d", n)
	}
	var sponsorLinked string
	_ = api.pool.QueryRow(ctx, `
		SELECT sponsor_client_user_id::text FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		filleulID).Scan(&sponsorLinked)
	if sponsorLinked != sponsorID {
		t.Fatalf("sponsor link missing: %s", sponsorLinked)
	}
}

func TestParrainageControl_ClientInviteInheritsCommercialReferral(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-inh-c"), "CommercialDemo123!", "Inherit Comm", nil)
	sponsorID := insertVerifiedUser(t, api, "client", uniqueEmail("parr-inh-sp"), "ClientDemo123!", "Inherit Sp", nil)
	_ = st.EnsureUserProfiles(ctx, sponsorID)
	_, err := api.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, 'SPONSOR1', NOW())`, sponsorID, commID)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}

	filleulEmail := uniqueEmail("parr-inh-fil")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": filleulEmail, "password": "ClientDemo123!", "fullName": "Inherit Fil",
		"consent": true, "inviteCode": inv.Code,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	var filleulID, gotComm string
	_ = api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, filleulEmail).Scan(&filleulID)
	err = api.pool.QueryRow(ctx, `
		SELECT commercial_user_id::text FROM practice.commercial_referrals WHERE client_user_id=$1`,
		filleulID).Scan(&gotComm)
	if err != nil || gotComm != commID {
		t.Fatalf("commercial inherit got=%s err=%v want %s", gotComm, err, commID)
	}
}

func TestParrainageControl_InvalidCodeDoesNotSilentlyAttachNearbyWhenCodePresent(t *testing.T) {
	api := newTestAPI(t)
	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("parr-inv-ign"), "CommercialDemo123!", "Ign Comm", nil)

	clientEmail := uniqueEmail("parr-inv-ign-cli")
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": clientEmail, "password": "ClientDemo123!", "fullName": "Invalid Code",
		"consent": true, "inviteCode": "NOTREAL1", "commercialUserId": commID,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	if dataMap(t, env)["inviteStatus"] != "ignored" {
		t.Fatalf("inviteStatus=%v want ignored", dataMap(t, env)["inviteStatus"])
	}
	var clientID string
	_ = api.pool.QueryRow(context.Background(), `SELECT id::text FROM identity.users WHERE email=$1`, clientEmail).Scan(&clientID)
	var n int
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COUNT(*)::int FROM practice.commercial_referrals WHERE client_user_id=$1`, clientID).Scan(&n)
	if n != 0 {
		t.Fatalf("invalid inviteCode must block nearby fallback (avoid wrong promoter), got %d referrals", n)
	}
}

func TestParrainageControl_SelfReferralSoftIgnored(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	sponsorEmail := uniqueEmail("parr-soft-self")
	sponsorID := insertVerifiedUser(t, api, "client", sponsorEmail, "ClientDemo123!", "Soft Self", nil)
	_ = st.EnsureUserProfiles(ctx, sponsorID)
	inv, err := st.EnsureAppInviteCode(ctx, sponsorID)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register-client", nil)
	status := handlers.TryClaimInvite(api.api, req, sponsorID, inv.Code)
	if status != "ignored" {
		t.Fatalf("soft self-referral want ignored, got %q", status)
	}
	var n int
	_ = api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		sponsorID).Scan(&n)
	if n != 0 {
		t.Fatalf("self soft claim must not write client_referrals, got %d", n)
	}
}

func TestParrainageControl_ClientInviteConcurrentNoMultiCabinet(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	vetA, practiceA, _ := insertVetWithPractice(t, api, st, "parr-race-va", "Race Cab A")
	vetB, practiceB, _ := insertVetWithPractice(t, api, st, "parr-race-vb", "Race Cab B")
	if practiceA == practiceB {
		t.Fatal("need two distinct practices")
	}

	sponsorA := insertVerifiedUser(t, api, "client", uniqueEmail("parr-race-sa"), "ClientDemo123!", "Race Sp A", map[string]any{
		"practice_id": practiceA,
	})
	_ = st.EnsureUserProfiles(ctx, sponsorA)
	_, _ = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceA, sponsorA, vetA)
	invA, err := st.EnsureAppInviteCode(ctx, sponsorA)
	if err != nil {
		t.Fatal(err)
	}

	sponsorB := insertVerifiedUser(t, api, "client", uniqueEmail("parr-race-sb"), "ClientDemo123!", "Race Sp B", map[string]any{
		"practice_id": practiceB,
	})
	_ = st.EnsureUserProfiles(ctx, sponsorB)
	_, _ = api.pool.Exec(ctx, `
		INSERT INTO practice.practice_clients (id, practice_id, client_user_id, vet_user_id)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), practiceB, sponsorB, vetB)
	invB, err := st.EnsureAppInviteCode(ctx, sponsorB)
	if err != nil {
		t.Fatal(err)
	}

	filleulEmail := uniqueEmail("parr-race-fil")
	filleulID := insertVerifiedUser(t, api, "client", filleulEmail, "ClientDemo123!", "Race Fil", nil)
	_ = st.EnsureUserProfiles(ctx, filleulID)
	tok := loginToken(t, api.handler, filleulEmail, "ClientDemo123!")

	type claimRes struct {
		code int
		env  map[string]any
	}
	ch := make(chan claimRes, 2)
	go func() {
		c, e := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
			"code": invA.Code,
		})
		ch <- claimRes{c, e}
	}()
	go func() {
		c, e := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/vets/claim-invite", tok, map[string]any{
			"code": invB.Code,
		})
		ch <- claimRes{c, e}
	}()
	r1, r2 := <-ch, <-ch
	if r1.code != http.StatusOK || r2.code != http.StatusOK {
		t.Fatalf("concurrent claims want 200: %d %#v / %d %#v", r1.code, r1.env, r2.code, r2.env)
	}

	var n int
	_ = api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.practice_clients WHERE client_user_id=$1`,
		filleulID).Scan(&n)
	if n != 1 {
		t.Fatalf("concurrent client QR must leave exactly 1 cabinet membership, got %d", n)
	}
	var sponsors int
	_ = api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.client_referrals WHERE referred_client_user_id=$1`,
		filleulID).Scan(&sponsors)
	if sponsors != 1 {
		t.Fatalf("first-wins sponsor want 1 row, got %d", sponsors)
	}
}
