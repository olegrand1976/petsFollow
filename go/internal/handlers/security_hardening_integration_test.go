package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Un code d'invitation fait ~40 bits : sans limite de débit, GET /public/app-invite/{code}
// est un oracle d'énumération (200 si le code existe, 404 sinon).
func TestPublicAppInviteIsRateLimited(t *testing.T) {
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "3")
	api := newTestAPI(t)

	throttled := false
	for i := 0; i < 8; i++ {
		code, env := doJSON(t, api.handler, http.MethodGet, "/api/v1/public/app-invite/ZZZZZZZZ", nil)
		if code == http.StatusTooManyRequests {
			throttled = true
			break
		}
		if code != http.StatusNotFound {
			t.Fatalf("attempt %d: got %d %#v", i, code, env)
		}
	}
	if !throttled {
		t.Fatal("public invite lookup must be rate limited (code enumeration)")
	}
}

// Un commercial peut vérifier qu'un prospect est déjà pris (anti-doublon) sans
// récupérer la fiche de contact d'une autre équipe.
func TestLookupProspectHidesContactDetailsAcrossTeams(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()

	mgrA := insertVerifiedUser(t, api, "commercial_manager", uniqueEmail("lk-mgr-a"), "CommercialDemo123!", "Manager A", nil)
	mgrB := insertVerifiedUser(t, api, "commercial_manager", uniqueEmail("lk-mgr-b"), "CommercialDemo123!", "Manager B", nil)

	ownerEmail := uniqueEmail("lk-owner")
	owner := insertVerifiedUser(t, api, "commercial", ownerEmail, "CommercialDemo123!", "Owner A",
		map[string]any{"manager_user_id": mgrA})
	teammateEmail := uniqueEmail("lk-mate")
	insertVerifiedUser(t, api, "commercial", teammateEmail, "CommercialDemo123!", "Mate A",
		map[string]any{"manager_user_id": mgrA})
	outsiderEmail := uniqueEmail("lk-outsider")
	insertVerifiedUser(t, api, "commercial", outsiderEmail, "CommercialDemo123!", "Outsider B",
		map[string]any{"manager_user_id": mgrB})

	marker := "Cabinet Lookup " + uuid.NewString()[:8]
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, status
		) VALUES ($1, $2, $3, 'Dr Contact', 'contact@cabinet.test', '+32470000000',
			'Namur', 'Notes internes confidentielles', 'commercial', 'new')`,
		uuid.NewString(), owner, marker); err != nil {
		t.Fatalf("insert prospect: %v", err)
	}

	lookup := func(email string) map[string]any {
		t.Helper()
		tok := loginToken(t, api.handler, email, "CommercialDemo123!")
		code, env := doAuthJSON(t, api.handler, http.MethodGet,
			"/api/v1/commercial/prospects/lookup?q="+url.QueryEscape(marker), tok, nil)
		if code != http.StatusOK {
			t.Fatalf("lookup as %s: %d %#v", email, code, env)
		}
		data := dataMap(t, env)
		if data["status"] != "owned" {
			t.Fatalf("lookup as %s: expected owned, got %#v", email, data)
		}
		p, ok := data["prospect"].(map[string]any)
		if !ok {
			t.Fatalf("lookup as %s: missing prospect %#v", email, data)
		}
		return p
	}

	for _, email := range []string{ownerEmail, teammateEmail} {
		p := lookup(email)
		if p["contactEmail"] != "contact@cabinet.test" || p["notes"] != "Notes internes confidentielles" {
			t.Fatalf("%s is on the owning team and must see full details: %#v", email, p)
		}
	}

	p := lookup(outsiderEmail)
	if p["practiceName"] != marker {
		t.Fatalf("anti-duplicate check needs the practice name: %#v", p)
	}
	for _, field := range []string{"contactEmail", "contactPhone", "contactName", "notes"} {
		if v, _ := p[field].(string); v != "" {
			t.Fatalf("cross-team lookup leaked %s=%q", field, v)
		}
	}
}

// Un prospect converti est déjà commissionné : le réattribuer réécrirait l'attribution.
func TestReassignProspectRefusesConverted(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	mgr := insertVerifiedUser(t, api, "commercial_manager", uniqueEmail("rea-mgr"), "CommercialDemo123!", "Reassign Mgr", nil)
	from := insertVerifiedUser(t, api, "commercial", uniqueEmail("rea-from"), "CommercialDemo123!", "Rep From",
		map[string]any{"manager_user_id": mgr})
	to := insertVerifiedUser(t, api, "commercial", uniqueEmail("rea-to"), "CommercialDemo123!", "Rep To",
		map[string]any{"manager_user_id": mgr})

	prospectID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, status
		) VALUES ($1, $2, $3, '', '', '', 'Liege', '', 'commercial', 'converted')`,
		prospectID, from, "Cabinet Converti "+uuid.NewString()[:8]); err != nil {
		t.Fatalf("insert prospect: %v", err)
	}

	err := st.ReassignProspectCommercial(ctx, prospectID, to, mgr)
	if !errors.Is(err, store.ErrValidation) {
		t.Fatalf("expected ErrValidation for a converted prospect, got %v", err)
	}

	var current string
	if err := api.pool.QueryRow(ctx,
		`SELECT commercial_user_id::text FROM sales.prospects WHERE id=$1`, prospectID).Scan(&current); err != nil {
		t.Fatal(err)
	}
	if current != from {
		t.Fatalf("converted prospect was moved: %s → %s", from, current)
	}
}

// assigned_commercial_id n'a pas de contrainte de rôle en base : le store doit
// refuser une cible qui n'est pas commercial / commercial_manager.
func TestAssignVetRefusesNonCommercialTarget(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	vetID := insertVerifiedUser(t, api, "vet", uniqueEmail("asg-vet"), "VetDemo123!", "Vet Target", nil)
	clientID := insertVerifiedUser(t, api, "client", uniqueEmail("asg-client"), "ClientDemo123!", "Not A Rep", nil)
	commID := insertVerifiedUser(t, api, "commercial", uniqueEmail("asg-comm"), "CommercialDemo123!", "Real Rep", nil)

	if err := st.AssignVetToCommercial(ctx, vetID, clientID, clientID); !errors.Is(err, store.ErrValidation) {
		t.Fatalf("expected ErrValidation when assigning a vet to a client, got %v", err)
	}
	var assigned *string
	if err := api.pool.QueryRow(ctx,
		`SELECT assigned_commercial_id::text FROM identity.users WHERE id=$1`, vetID).Scan(&assigned); err != nil {
		t.Fatal(err)
	}
	if assigned != nil {
		t.Fatalf("vet must stay unassigned, got %v", *assigned)
	}

	if err := st.AssignVetToCommercial(ctx, vetID, commID, commID); err != nil {
		t.Fatalf("assigning to a real commercial must work: %v", err)
	}
	if err := api.pool.QueryRow(ctx,
		`SELECT assigned_commercial_id::text FROM identity.users WHERE id=$1`, vetID).Scan(&assigned); err != nil {
		t.Fatal(err)
	}
	if assigned == nil || *assigned != commID {
		t.Fatalf("expected assignment to %s, got %v", commID, fmt.Sprint(assigned))
	}
}
