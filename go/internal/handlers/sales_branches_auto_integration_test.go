package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSalesBranchesAutoCreateAndSkipPeerSponsored(t *testing.T) {
	secret := "test-sales-branches-auto-secret"
	t.Setenv("SALES_BRANCHES_AUTO_SECRET", secret)
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000)
	lastName := "Zbranch" + suffix // unique code base ZBRANCH{suffix}C
	indepName := "Camille " + lastName
	wantCodePrefix := "ZBRANCH" + suffix + "C"

	indepEmail := uniqueEmail("branch-indep")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": indepEmail, "password": "CommercialDemo123!", "fullName": indepName,
	})
	if code != http.StatusCreated {
		t.Fatalf("create indep commercial %d %#v", code, env)
	}
	var indepID string
	_ = api.pool.QueryRow(context.Background(),
		`SELECT id::text FROM identity.users WHERE email=$1`, indepEmail).Scan(&indepID)
	if indepID == "" {
		t.Fatal("missing indep commercial id")
	}
	_, _ = api.pool.Exec(context.Background(), `
		UPDATE identity.users SET branch_id=NULL, sponsor_user_id=NULL WHERE id=$1`, indepID)

	sponsorEmail := uniqueEmail("branch-sponsor")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": sponsorEmail, "password": "CommercialDemo123!", "fullName": "Alex Sponsor" + suffix,
	})
	if code != http.StatusCreated {
		t.Fatalf("create sponsor %d %#v", code, env)
	}
	var sponsorID string
	_ = api.pool.QueryRow(context.Background(),
		`SELECT id::text FROM identity.users WHERE email=$1`, sponsorEmail).Scan(&sponsorID)
	_, _ = api.pool.Exec(context.Background(), `
		UPDATE identity.users SET branch_id=NULL, sponsor_user_id=NULL WHERE id=$1`, sponsorID)

	peerEmail := uniqueEmail("branch-peer")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": peerEmail, "password": "CommercialDemo123!", "fullName": "Peer Under" + suffix,
	})
	if code != http.StatusCreated {
		t.Fatalf("create peer %d %#v", code, env)
	}
	var peerID string
	_ = api.pool.QueryRow(context.Background(),
		`SELECT id::text FROM identity.users WHERE email=$1`, peerEmail).Scan(&peerID)
	_, _ = api.pool.Exec(context.Background(), `
		UPDATE identity.users SET branch_id=NULL, sponsor_user_id=$2 WHERE id=$1`, peerID, sponsorID)

	// Occupy base code → auto-run must use suffix 2.
	_, err := api.pool.Exec(context.Background(), `
		INSERT INTO sales.branches (id, name, code)
		VALUES ($1, 'Occupied', $2)`, uuid.NewString(), wantCodePrefix)
	if err != nil {
		t.Fatal(err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/sales-branches/auto-run", adminTok, map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("admin auto-run %d %#v", code, env)
	}
	data := dataMap(t, env)
	created := int(data["created"].(float64))
	if created < 1 {
		t.Fatalf("expected at least 1 auto branch, got %#v", data)
	}
	if _, ok := data["skipped"]; !ok {
		t.Fatalf("expected skipped count in response, got %#v", data)
	}

	var indepBranchCode, indepBranchID string
	err = api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(b.code,''), COALESCE(u.branch_id::text,'')
		FROM identity.users u
		LEFT JOIN sales.branches b ON b.id=u.branch_id
		WHERE u.id=$1`, indepID).Scan(&indepBranchCode, &indepBranchID)
	if err != nil {
		t.Fatal(err)
	}
	if indepBranchID == "" {
		t.Fatal("independent commercial should have a branch")
	}
	if indepBranchCode != wantCodePrefix+"2" {
		t.Fatalf("expected collision suffix code %s2, got %q", wantCodePrefix, indepBranchCode)
	}

	var peerBranchID string
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(branch_id::text,'') FROM identity.users WHERE id=$1`, peerID).Scan(&peerBranchID)
	if peerBranchID != "" {
		t.Fatalf("peer-sponsored commercial should stay branchless, got %s", peerBranchID)
	}

	// Detach works via empty branchId.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/commercials/"+indepID+"/branch", adminTok, map[string]any{
		"branchId": "",
	})
	if code != http.StatusOK {
		t.Fatalf("detach %d %#v", code, env)
	}
	_ = api.pool.QueryRow(context.Background(), `
		SELECT COALESCE(branch_id::text,'') FROM identity.users WHERE id=$1`, indepID).Scan(&indepBranchID)
	if indepBranchID != "" {
		t.Fatalf("expected detached commercial, still on %s", indepBranchID)
	}
	// Re-attach for subsequent internal job idempotency checks.
	_, _ = api.pool.Exec(context.Background(), `
		UPDATE identity.users SET branch_id=(SELECT id FROM sales.branches WHERE code=$1) WHERE id=$2`,
		indepBranchCode, indepID)

	// Internal without secret → 401
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/sales-branches-auto/run", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("internal without secret expected 401, got %d", rec.Code)
	}

	// Internal with secret → 200 (idempotent: 0 new)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/internal/sales-branches-auto/run", nil)
	req.Header.Set("X-Sales-Branches-Auto-Secret", secret)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("internal with secret expected 200, got %d %s", rec.Code, rec.Body.String())
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/sales-branches", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list branches %d %#v", code, env)
	}
	listData := dataMap(t, env)
	if _, ok := listData["branches"]; !ok {
		t.Fatalf("expected branches key, got %#v", listData)
	}
	if _, ok := listData["pendingAuto"]; !ok {
		t.Fatalf("expected pendingAuto key, got %#v", listData)
	}
}
