package handlers_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestCommercialCRMFicheActivitiesAgenda(t *testing.T) {
	api := newTestAPI(t)

	adminEmail := uniqueEmail("crm-admin")
	insertVerifiedUser(t, api, string(kernel.RoleAdmin), adminEmail, "AdminDemo123!", "Admin CRM", nil)
	adminTok := loginToken(t, api.handler, adminEmail, "AdminDemo123!")

	mgrEmail := uniqueEmail("crm-mgr")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": mgrEmail, "password": "CommercialDemo123!", "fullName": "Manager CRM", "role": "commercial_manager",
	})
	if code != http.StatusCreated {
		t.Fatalf("create manager %d %#v", code, env)
	}
	mgrID := dataMap(t, env)["userId"].(string)
	mgrTok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")

	commEmail := uniqueEmail("crm-comm")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": commEmail, "password": "CommercialDemo123!", "fullName": "Rep CRM", "managerUserId": mgrID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create commercial %d %#v", code, env)
	}
	commID := dataMap(t, env)["userId"].(string)
	commTok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", commTok, map[string]any{
		"practiceName": "Cabinet CRM " + uuid.NewString()[:8],
		"contactName":  "Dr CRM",
		"contactEmail": uniqueEmail("prospect-crm"),
		"contactPhone": "0471" + fmt.Sprintf("%06d", time.Now().UnixNano()%1_000_000),
		"city":         "Liège",
		"status":       "new",
	})
	if code != http.StatusCreated {
		t.Fatalf("create prospect: %d %#v", code, env)
	}
	prospectID := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/commercial/prospects/"+prospectID, commTok, map[string]any{
		"status": "contacted",
	})
	if code != http.StatusOK {
		t.Fatalf("patch status: %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/events", commTok, map[string]any{
		"kind": "call",
		"body": "Appel découverte CRM",
	})
	if code != http.StatusCreated {
		t.Fatalf("create event: %d %#v", code, env)
	}

	due := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+prospectID+"/activities", commTok, map[string]any{
		"kind":  "follow_up",
		"title": "Relance J+1",
		"dueAt": due,
	})
	if code != http.StatusCreated {
		t.Fatalf("create activity: %d %#v", code, env)
	}
	actID := dataMap(t, env)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects/"+prospectID, commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get detail: %d %#v", code, env)
	}
	detail := dataMap(t, env)
	events, _ := detail["events"].([]any)
	if len(events) < 2 {
		t.Fatalf("expected timeline events, got %#v", detail["events"])
	}
	acts, _ := detail["activities"].([]any)
	if len(acts) < 1 {
		t.Fatalf("expected activities, got %#v", detail["activities"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/commercial/activities/"+actID, commTok, map[string]any{
		"status": "done",
	})
	if code != http.StatusOK {
		t.Fatalf("patch activity: %d %#v", code, env)
	}

	overdueDue := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial-manager/activities", mgrTok, map[string]any{
		"prospectId":     prospectID,
		"title":          "Relance manager",
		"kind":           "follow_up",
		"assigneeUserId": commID,
		"dueAt":          overdueDue,
	})
	if code != http.StatusCreated {
		t.Fatalf("manager assign: %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/followups", mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("followups: %d %#v", code, env)
	}
	fu := dataMap(t, env)
	overdue, _ := fu["overdueActivities"].([]any)
	if len(overdue) < 1 {
		t.Fatalf("expected overdue activities, got %#v", fu)
	}

	from := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	to := time.Now().UTC().AddDate(0, 0, 8).Format("2006-01-02")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/agenda?from="+from+"&to="+to, commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("agenda: %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial-manager/agenda?from="+from+"&to="+to, mgrTok, nil)
	if code != http.StatusOK {
		t.Fatalf("manager agenda: %d %#v", code, env)
	}

	// Auto status_change from TouchProspectContacted on note against a fresh "new" prospect.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects", commTok, map[string]any{
		"practiceName": "Cabinet AutoStatus " + uuid.NewString()[:8],
		"contactName":  "Dr Auto",
		"contactEmail": uniqueEmail("prospect-autostatus"),
		"city":         "Namur",
		"status":       "new",
	})
	if code != http.StatusCreated {
		t.Fatalf("create new prospect: %d %#v", code, env)
	}
	newPID := dataMap(t, env)["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/prospects/"+newPID+"/events", commTok, map[string]any{
		"kind": "note", "body": "Premier contact",
	})
	if code != http.StatusCreated {
		t.Fatalf("note on new: %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects/"+newPID, commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("detail after note: %d %#v", code, env)
	}
	detail2 := dataMap(t, env)
	p2, _ := detail2["prospect"].(map[string]any)
	if st, _ := p2["status"].(string); st != "contacted" {
		t.Fatalf("expected auto contacted, got %#v", p2["status"])
	}
	foundStatus := false
	for _, raw := range detail2["events"].([]any) {
		ev, _ := raw.(map[string]any)
		if ev["kind"] == "status_change" {
			foundStatus = true
			break
		}
	}
	if !foundStatus {
		t.Fatalf("expected status_change event after TouchProspectContacted, got %#v", detail2["events"])
	}

	// ACL: peer commercial cannot read teammate prospect.
	peerEmail := uniqueEmail("crm-peer")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": peerEmail, "password": "CommercialDemo123!", "fullName": "Peer CRM", "managerUserId": mgrID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create peer %d %#v", code, env)
	}
	peerTok := loginToken(t, api.handler, peerEmail, "CommercialDemo123!")
	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/prospects/"+prospectID, peerTok, nil)
	if code != http.StatusNotFound && code != http.StatusForbidden {
		t.Fatalf("peer get detail expected 403/404, got %d", code)
	}

	// ACL: foreign manager cannot assign on this prospect.
	otherMgrEmail := uniqueEmail("crm-othermgr")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/commercials", adminTok, map[string]any{
		"email": otherMgrEmail, "password": "CommercialDemo123!", "fullName": "Other Mgr", "role": "commercial_manager",
	})
	if code != http.StatusCreated {
		t.Fatalf("create other mgr %d %#v", code, env)
	}
	otherMgrTok := loginToken(t, api.handler, otherMgrEmail, "CommercialDemo123!")
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial-manager/activities", otherMgrTok, map[string]any{
		"prospectId": prospectID, "title": "Hack", "kind": "follow_up",
	})
	if code != http.StatusNotFound && code != http.StatusForbidden {
		t.Fatalf("foreign manager assign expected 403/404, got %d", code)
	}
}
