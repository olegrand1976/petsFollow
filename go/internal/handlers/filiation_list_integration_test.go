package handlers_test

import (
	"fmt"
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
	rowsA := filiationItems(t, env)
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
	rowsB := filiationItems(t, env)
	if !filiationHasCommercial(rowsB, commB) || filiationHasCommercial(rowsB, commA) {
		t.Fatalf("B isolation %#v", rowsB)
	}
	// Orphan referral (no practice_clients): Effectif none — aligned with Resolve/Accrue.
	if !filiationRowMatch(rowsB, func(m map[string]any) bool {
		return m["commercialUserId"] == commB &&
			m["clientUserId"] != "" &&
			m["effectiveSource"] == store.FiliationSourceNone &&
			(m["effectiveCommercialId"] == nil || m["effectiveCommercialId"] == "")
	}) {
		t.Fatalf("B orphan referral Effectif none %#v", rowsB)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/filiation", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin filiation %d %#v", code, env)
	}
	rowsAdmin := filiationItems(t, env)
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
	rowsMgr := filiationItems(t, env)
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
	rowsA := filiationItems(t, env)
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
	rowsC := filiationItems(t, env)
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

// After unassign (F10-like): list Effectif falls back to client_referral.
func TestFiliationList_UnassignFallsBackToReferralEffective(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-list-f10-a", "F10 List A")

	clientID, vetID, _ := referredClientClaimsAssignedVet(t, api, commA, tokA, "fil-list-f10-cli", "fil-list-f10-vet")

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/vets/"+vetID+"/unassign", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("unassign %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation", tokA, nil)
	if code != http.StatusOK {
		t.Fatalf("filiation %d %#v", code, env)
	}
	rows := filiationItems(t, env)
	if !filiationRowMatch(rows, func(m map[string]any) bool {
		return m["clientUserId"] == clientID &&
			m["commercialUserId"] == commA &&
			m["effectiveCommercialId"] == commA &&
			m["effectiveSource"] == store.FiliationSourceClientReferral
	}) {
		t.Fatalf("after unassign want effectiveSource=client_referral %#v", rows)
	}
}

func TestFiliationList_LimitTruncates(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	_, _, tok := createCommercial(t, api, adminTok, "fil-list-lim", "Limit Comm")

	for i := 0; i < 3; i++ {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/clients", tok, map[string]any{
			"email": uniqueEmail(fmt.Sprintf("fil-lim-%d", i)), "password": "ClientDemo123!", "fullName": fmt.Sprintf("Lim %d", i),
		})
		if code != http.StatusCreated {
			t.Fatalf("client %d %d %#v", i, code, env)
		}
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation?limit=1", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("filiation limit %d %#v", code, env)
	}
	page := dataMap(t, env)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("limit=1 want 1 item, got %d %#v", len(items), page)
	}
	if truncated, _ := page["truncated"].(bool); !truncated {
		t.Fatalf("expected truncated=true %#v", page)
	}
	if lim, _ := page["limit"].(float64); int(lim) != 1 {
		t.Fatalf("limit field %#v", page["limit"])
	}

	// format=items → legacy flat array (non-breaking for old clients).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/filiation?format=items&limit=2", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("format=items %d %#v", code, env)
	}
	flat, ok := env["data"].([]any)
	if !ok {
		t.Fatalf("format=items want array, got %#v", env["data"])
	}
	if len(flat) == 0 || len(flat) > 2 {
		t.Fatalf("format=items unexpected len %d", len(flat))
	}
}

func TestFiliationList_ManagerEventsScopedToTeam(t *testing.T) {
	api := newTestAPI(t)
	adminTok := ensureAdminToken(t, api)
	commA, _, tokA := createCommercial(t, api, adminTok, "fil-evt-a", "Evt Comm A")
	commB, _, tokB := createCommercial(t, api, adminTok, "fil-evt-b", "Evt Comm B")

	// Encode under A → event for A.
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokA, map[string]any{
		"email": uniqueEmail("fil-evt-va"), "password": "VetDemo123!", "fullName": "Dr EvtA",
		"practiceName": "Cab Evt A",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode A %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", tokB, map[string]any{
		"email": uniqueEmail("fil-evt-vb"), "password": "VetDemo123!", "fullName": "Dr EvtB",
		"practiceName": "Cab Evt B",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode B %d %#v", code, env)
	}

	mgrEmail := uniqueEmail("fil-evt-mgr")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": mgrEmail, "password": "CommercialDemo123!", "fullName": "Evt Manager",
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

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/filiation/events", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("manager events %d %#v", code, env)
	}
	events := filiationEventItems(t, env)
	if !filiationEventMatch(events, func(m map[string]any) bool {
		return m["eventType"] == store.FiliationEventVetAssigned && m["commercialUserId"] == commA
	}) {
		t.Fatalf("manager should see A events %#v", events)
	}
	if filiationEventMatch(events, func(m map[string]any) bool {
		return m["commercialUserId"] == commB
	}) {
		t.Fatalf("manager must not see B events %#v", events)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/filiation?commercialId="+commA, mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("manager filiation filter A %d %#v", code, env)
	}
	rows := filiationItems(t, env)
	if !filiationHasCommercial(rows, commA) {
		t.Fatalf("filter A missing rows %#v", rows)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/filiation?commercialId="+commB, mgrTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("filter B outside team want 404, got %d %#v", code, env)
	}
}

func filiationItems(t *testing.T, env map[string]any) []any {
	t.Helper()
	data := env["data"]
	if m, ok := data.(map[string]any); ok {
		items, _ := m["items"].([]any)
		if items == nil {
			return []any{}
		}
		return items
	}
	// Backward-compat if a test hits an old array shape.
	if arr, ok := data.([]any); ok {
		return arr
	}
	t.Fatalf("filiation data shape: %#v", data)
	return nil
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
