package handlers_test

import (
	"net/http"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestFiliationList_ScopedVisibility(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-list-a", "List Comm A")
	commB, _, tokB := createCommercial(t, api, adminTok, "fil-list-b", "List Comm B")

	vetEmail := uniqueEmail("fil-list-vet")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr List",
		"practiceName": "Cabinet List A",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode %d %#v", code, env)
	}

	soloEmail := uniqueEmail("fil-list-solo")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", tokB, map[string]any{
		"email": soloEmail, "password": "ClientDemo123!", "fullName": "Solo List B",
	})
	if code != http.StatusCreated {
		t.Fatalf("standalone %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokA, nil)
	if code != http.StatusOK {
		t.Fatalf("A filiation %d %#v", code, env)
	}
	rowsA, _ := env["data"].([]any)
	if !filiationHasCommercial(rowsA, commA) {
		t.Fatalf("A should see self rows %#v", rowsA)
	}
	if filiationHasCommercial(rowsA, commB) {
		t.Fatalf("A must not see B rows %#v", rowsA)
	}
	if !filiationRowMatch(rowsA, func(m map[string]any) bool {
		return m["commercialUserId"] == commA &&
			m["effectiveSource"] == store.FiliationSourceVetAssignment &&
			m["effectiveCommercialId"] == commA
	}) {
		t.Fatalf("A vet row effectiveCommercialId/source %#v", rowsA)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokB, nil)
	if code != http.StatusOK {
		t.Fatalf("B filiation %d %#v", code, env)
	}
	rowsB, _ := env["data"].([]any)
	if !filiationHasCommercial(rowsB, commB) || filiationHasCommercial(rowsB, commA) {
		t.Fatalf("B isolation %#v", rowsB)
	}
	if !filiationRowMatch(rowsB, func(m map[string]any) bool {
		return m["commercialUserId"] == commB &&
			m["effectiveSource"] == store.FiliationSourceClientReferral &&
			m["effectiveCommercialId"] == commB
	}) {
		t.Fatalf("B referral effective %#v", rowsB)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin filiation %d %#v", code, env)
	}
	rowsAdmin, _ := env["data"].([]any)
	if !filiationHasCommercial(rowsAdmin, commA) || !filiationHasCommercial(rowsAdmin, commB) {
		t.Fatalf("admin should see A and B %#v", rowsAdmin)
	}

	mgrEmail := uniqueEmail("fil-list-mgr")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": mgrEmail, "password": "CommercialDemo123!", "fullName": "List Manager",
		"role": "commercial_manager",
	})
	if code != http.StatusCreated {
		t.Fatalf("create manager %d %#v", code, env)
	}
	mgrID, _ := dataMap(t, env)["userId"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/commercials/"+commA+"/manager", adminTok, map[string]any{
		"managerUserId": mgrID,
	})
	if code != http.StatusOK {
		t.Fatalf("assign manager %d %#v", code, env)
	}
	mgrTok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/filiation", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("manager filiation %d %#v", code, env)
	}
	rowsMgr, _ := env["data"].([]any)
	if !filiationHasCommercial(rowsMgr, commA) || filiationHasCommercial(rowsMgr, commB) {
		t.Fatalf("manager scope %#v", rowsMgr)
	}
}

// F9 list: client referred to A, joins vet assigned to C → A sees referral row with effective=C;
// C sees assign row with effective=C.
func TestFiliationList_AssignPriorityEffectiveOnReferralRow(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-list-f9-a", "F9 List A")
	commC, _, tokC := createCommercial(t, api, adminTok, "fil-list-f9-c", "F9 List C")

	clientID, _, _ := referredClientClaimsAssignedVet(t, api, commA, tokC, "fil-list-f9-cli", "fil-list-f9-vet")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokA, nil)
	if code != http.StatusOK {
		t.Fatalf("A filiation %d %#v", code, env)
	}
	rowsA, _ := env["data"].([]any)
	if !filiationRowMatch(rowsA, func(m map[string]any) bool {
		return m["clientUserId"] == clientID &&
			m["commercialUserId"] == commA &&
			m["referralCommercialId"] == commA &&
			m["effectiveCommercialId"] == commC &&
			m["effectiveSource"] == store.FiliationSourceVetAssignment
	}) {
		t.Fatalf("A should list referral client with effective=C %#v", rowsA)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokC, nil)
	if code != http.StatusOK {
		t.Fatalf("C filiation %d %#v", code, env)
	}
	rowsC, _ := env["data"].([]any)
	if !filiationRowMatch(rowsC, func(m map[string]any) bool {
		return m["clientUserId"] == clientID &&
			m["commercialUserId"] == commC &&
			m["effectiveCommercialId"] == commC &&
			m["effectiveSource"] == store.FiliationSourceVetAssignment
	}) {
		t.Fatalf("C should list assigned vet client with effective=C %#v", rowsC)
	}
}

func TestFiliationList_AdminInvalidBranchUUID(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation?branchId=not-a-uuid", adminTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid branchId want 400, got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation?commercialId=abc", adminTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("invalid commercialId want 400, got %d %#v", code, env)
	}
}

func filiationHasCommercial(rows []any, commercialID string) bool {
	return filiationRowMatch(rows, func(m map[string]any) bool {
		return m["commercialUserId"] == commercialID
	})
}

func filiationRowMatch(rows []any, pred func(map[string]any) bool) bool {
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		if pred(m) {
			return true
		}
	}
	return false
}
