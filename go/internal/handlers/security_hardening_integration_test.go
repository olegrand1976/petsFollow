package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Les documents animaux (analyses, radios) sont du PHI. Ils ont été servis via
// une URL publique GCS : plus aucune URL ne doit sortir de l'API, et le flux
// authentifié doit refuser un véto étranger au dossier.
func TestPetDocumentIsPrivateAndAuthenticated(t *testing.T) {
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	strangerTok := loginToken(t, api.handler, "vet.lyon@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	pdf := []byte("%PDF-1.4\n% analyse sanguine\n")
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "analyse.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pdf); err != nil {
		t.Fatal(err)
	}
	_ = w.WriteField("title", "Analyse sécurité")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets/"+petID+"/documents", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+vetTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("upload document %d %#v", rec.Code, envelope)
	}
	doc := dataMap(t, envelope)
	docID, _ := doc["id"].(string)
	if docID == "" {
		t.Fatalf("missing document id: %#v", doc)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/documents/"+docID, vetTok, nil)
	})
	if _, ok := doc["fileUrl"]; ok {
		t.Fatalf("upload response must not carry a storage URL: %#v", doc)
	}

	// La colonne file_url reste vide : rien à fuiter même en cas de dump.
	var fileURL string
	if err := api.pool.QueryRow(context.Background(),
		`SELECT COALESCE(file_url,'') FROM pets.documents WHERE id=$1`, docID).Scan(&fileURL); err != nil {
		t.Fatal(err)
	}
	if fileURL != "" {
		t.Fatalf("file_url must stay empty, got %q", fileURL)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/documents", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list documents %d %#v", code, env)
	}
	if raw, _ := json.Marshal(env); strings.Contains(string(raw), "fileUrl") {
		t.Fatalf("listing must not expose fileUrl: %s", raw)
	}

	dl := "/api/v1/pets/" + petID + "/documents/" + docID + "/download"
	for _, tok := range []string{clientTok, vetTok} {
		req = httptest.NewRequest(http.MethodGet, dl, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec = httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("download %d body=%s", rec.Code, rec.Body.String())
		}
		got, _ := io.ReadAll(rec.Body)
		if !bytes.Equal(got, pdf) {
			t.Fatalf("unexpected document body (%d bytes)", len(got))
		}
	}

	for _, tc := range []struct{ name, token string }{
		{"anonymous", ""},
		{"vet of another practice", strangerTok},
	} {
		req = httptest.NewRequest(http.MethodGet, dl, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec = httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Fatalf("%s must not download PHI (got 200)", tc.name)
		}
	}
}

// Un commercial ne doit pas pouvoir se greffer un profil pro sur un compte hors
// de son portefeuille : ce serait une escalade vers le dossier patients d'un
// cabinet arbitraire.
func TestCommercialAttachProfileRefusesOutsidePortfolio(t *testing.T) {
	api := newTestAPI(t)

	commEmail := uniqueEmail("atc-comm")
	commID := insertVerifiedUser(t, api, "commercial", commEmail, "CommercialDemo123!", "Attach Rep", nil)
	outsider := insertVerifiedUser(t, api, "client", uniqueEmail("atc-out"), "ClientDemo123!", "Out Of Reach", nil)
	owned := insertVerifiedUser(t, api, "client", uniqueEmail("atc-own"), "ClientDemo123!", "In Portfolio",
		map[string]any{"assigned_commercial_id": commID})

	tok := loginToken(t, api.handler, commEmail, "CommercialDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/commercial/users/"+outsider+"/profiles", tok,
		map[string]any{"role": "care_pro", "specialty": "farrier"})
	if code != http.StatusForbidden && code != http.StatusNotFound {
		t.Fatalf("attaching outside the portfolio must be refused, got %d %#v", code, env)
	}

	// Même dans le portefeuille, les rôles cabinet restent réservés à l'admin.
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/commercial/users/"+owned+"/profiles", tok,
		map[string]any{"role": "vet", "practiceId": uuid.NewString()})
	if code != http.StatusForbidden {
		t.Fatalf("a commercial must not attach a vet profile, got %d %#v", code, env)
	}
}

// Le portefeuille d'un manager, ce sont les contacts de son équipe : restreindre
// l'attach au seul commercial assigné rendrait la route morte pour lui. Le
// périmètre s'arrête à son équipe — pas au reste de la force de vente.
func TestCommercialManagerAttachCoversItsTeamOnly(t *testing.T) {
	api := newTestAPI(t)

	mgrEmail := uniqueEmail("mgr-attach")
	mgrID := insertVerifiedUser(t, api, "commercial_manager", mgrEmail, "CommercialDemo123!", "Team Lead", nil)
	repID := insertVerifiedUser(t, api, "commercial", uniqueEmail("mgr-rep"), "CommercialDemo123!", "Own Rep",
		map[string]any{"manager_user_id": mgrID})
	otherRepID := insertVerifiedUser(t, api, "commercial", uniqueEmail("mgr-other"), "CommercialDemo123!", "Other Rep", nil)

	teamContact := insertVerifiedUser(t, api, "client", uniqueEmail("mgr-own"), "ClientDemo123!", "Team Contact",
		map[string]any{"assigned_commercial_id": repID})
	foreignContact := insertVerifiedUser(t, api, "client", uniqueEmail("mgr-foreign"), "ClientDemo123!", "Foreign Contact",
		map[string]any{"assigned_commercial_id": otherRepID})

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.profiles WHERE user_id = $1::uuid`, teamContact)
	})

	tok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/commercial/users/"+teamContact+"/profiles", tok,
		map[string]any{"role": "care_pro", "specialty": "farrier"})
	if code != http.StatusCreated {
		t.Fatalf("a manager must cover its team's portfolio, got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/commercial/users/"+foreignContact+"/profiles", tok,
		map[string]any{"role": "care_pro", "specialty": "farrier"})
	if code != http.StatusForbidden && code != http.StatusNotFound {
		t.Fatalf("another team's contact must stay out of reach, got %d %#v", code, env)
	}
}

// Les callbacks mock de facturation sont atteints par redirection navigateur,
// donc sans bearer : la signature HMAC est le seul contrôle.
func TestBillingMockCompleteRequiresSignature(t *testing.T) {
	api := newTestAPI(t)

	code, env := doJSON(t, api.handler, http.MethodGet,
		"/api/v1/billing/dev/mock-complete?pet_id="+uuid.NewString()+
			"&owner_user_id="+uuid.NewString()+"&plan_code=triennial&billing_mode=subscription", nil)
	if code != http.StatusForbidden {
		t.Fatalf("forged mock checkout must be rejected, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodGet,
		"/api/v1/billing/dev/mock-portal?customer=cus_mock_"+uuid.NewString(), nil)
	if code != http.StatusForbidden {
		t.Fatalf("forged mock portal must be rejected, got %d %#v", code, env)
	}
}

// Contrepartie du test précédent : une URL réellement émise par l'API doit
// continuer à passer. Sans ce cas, un mock-complete qui refuserait tout le
// monde satisferait quand même le test « forgé → 403 ».
func TestBillingMockCompleteAcceptsSignedURL(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets", clientTok, map[string]any{
		"name": "Signature Probe", "species": "dog", "breed": "test",
		"plan": "triennial", "billingMode": "subscription",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("create pet %d %#v", code, env)
	}
	data := dataMap(t, env)
	pet, _ := data["pet"].(map[string]any)
	if pet == nil {
		pet = data
	}
	petID, _ := pet["id"].(string)
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/"+petID, clientTok, nil)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodGet, mockCompletePath(t, data), clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("signed mock checkout must be honoured, got %d %#v", code, env)
	}
}

// Garde-fou inverse du durcissement commercial : la voie admin doit continuer à
// pouvoir attacher un profil cabinet, sinon on a fermé la fonctionnalité et pas
// seulement l'escalade.
func TestAdminCanStillAttachPracticeProfile(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if practiceID == "" {
		t.Skip("seeded vet has no practice")
	}

	target := insertVerifiedUser(t, api, "client", uniqueEmail("adm-attach"), "ClientDemo123!", "Admin Attach", nil)
	// L'attachement inscrit la cible à l'équipe du cabinet seedé : sans purge,
	// le test grossit l'effectif VetPlus pour tous les suivants.
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = api.pool.Exec(ctx, `DELETE FROM practice.team_members WHERE user_id = $1::uuid`, target)
		_, _ = api.pool.Exec(ctx, `DELETE FROM identity.profiles WHERE user_id = $1::uuid`, target)
	})

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/users/"+target+"/profiles", adminTok,
		map[string]any{"role": "vet", "practiceId": practiceID})
	if code != http.StatusCreated {
		t.Fatalf("admin must still attach a vet profile, got %d %#v", code, env)
	}
}

// Garde-fou inverse de la révocation : token_version ne doit bouger qu'aux
// moments prévus. Un bump à chaque refresh déconnecterait tout le monde en
// boucle, et le test « ancien token révoqué » passerait quand même.
func TestConsecutiveRefreshesKeepTheSessionAlive(t *testing.T) {
	api := newTestAPI(t)

	email := uniqueEmail("keep-session")
	insertVerifiedUser(t, api, "client", email, "ClientDemo123!", "Keep Session", nil)

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login",
		map[string]any{"email": email, "password": "ClientDemo123!"})
	if code != http.StatusOK {
		t.Fatalf("login %d %#v", code, env)
	}
	refresh, _ := dataMap(t, env)["refreshToken"].(string)

	for i := 0; i < 3; i++ {
		code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
			map[string]any{"refreshToken": refresh})
		if code != http.StatusOK {
			t.Fatalf("refresh %d must succeed, got %d %#v", i, code, env)
		}
		refresh, _ = dataMap(t, env)["refreshToken"].(string)
		if refresh == "" {
			t.Fatalf("refresh %d returned no new refresh token: %#v", i, env)
		}
	}
}

// Un reset de mot de passe doit invalider les tokens déjà émis : sans ça, un
// refresh token volé reste utilisable 30 jours après la reprise en main.
func TestPasswordResetRevokesExistingRefreshToken(t *testing.T) {
	api := newTestAPI(t)

	email := uniqueEmail("rev-user")
	insertVerifiedUser(t, api, "client", email, "ClientDemo123!", "Revoke Me", nil)

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login",
		map[string]any{"email": email, "password": "ClientDemo123!"})
	if code != http.StatusOK {
		t.Fatalf("login %d %#v", code, env)
	}
	stolen, _ := dataMap(t, env)["refreshToken"].(string)
	if stolen == "" {
		t.Fatalf("missing refresh token: %#v", env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
		map[string]any{"refreshToken": stolen})
	if code != http.StatusOK {
		t.Fatalf("refresh before reset should work: %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/forgot-password",
		map[string]any{"email": email})
	if code != http.StatusOK {
		t.Fatalf("forgot-password %d %#v", code, env)
	}
	resetToken := resetTokenFor(t, api, email)
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/reset-password",
		map[string]any{"token": resetToken, "password": "ClientDemo456!"})
	if code != http.StatusOK {
		t.Fatalf("reset-password %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
		map[string]any{"refreshToken": stolen})
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh token issued before the reset must be revoked, got %d %#v", code, env)
	}
}

// POST /auth/logout révoque côté serveur : purger le cookie httpOnly ne suffit
// pas si le refresh token a déjà fuité.
func TestLogoutRevokesRefreshToken(t *testing.T) {
	api := newTestAPI(t)

	email := uniqueEmail("lgo-user")
	insertVerifiedUser(t, api, "client", email, "ClientDemo123!", "Logout Me", nil)

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login",
		map[string]any{"email": email, "password": "ClientDemo123!"})
	if code != http.StatusOK {
		t.Fatalf("login %d %#v", code, env)
	}
	data := dataMap(t, env)
	access, _ := data["accessToken"].(string)
	refresh, _ := data["refreshToken"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/auth/logout", access, nil)
	if code != http.StatusOK {
		t.Fatalf("logout %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
		map[string]any{"refreshToken": refresh})
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout must fail, got %d %#v", code, env)
	}
}

// Changer son mot de passe doit fermer les autres appareils — c'est le geste
// qu'on fait quand on soupçonne une compromission. Mais l'appareil qui le
// demande doit survivre : le handler lui réémet une paire, sinon changer son
// mot de passe revient à se déconnecter soi-même.
func TestChangePasswordRevokesOtherDevicesButNotTheCaller(t *testing.T) {
	api := newTestAPI(t)

	email := uniqueEmail("chg-pwd")
	insertVerifiedUser(t, api, "client", email, "ClientDemo123!", "Change Pwd", nil)

	// Appareil A : la session qui restera ouverte ailleurs.
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login",
		map[string]any{"email": email, "password": "ClientDemo123!"})
	if code != http.StatusOK {
		t.Fatalf("login A %d %#v", code, env)
	}
	otherRefresh, _ := dataMap(t, env)["refreshToken"].(string)

	// Appareil B : celui qui change le mot de passe.
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login",
		map[string]any{"email": email, "password": "ClientDemo123!"})
	if code != http.StatusOK {
		t.Fatalf("login B %d %#v", code, env)
	}
	callerAccess, _ := dataMap(t, env)["accessToken"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me/password", callerAccess,
		map[string]any{"currentPassword": "ClientDemo123!", "newPassword": "ClientDemo456!"})
	if code != http.StatusOK {
		t.Fatalf("change password %d %#v", code, env)
	}
	reissued, _ := dataMap(t, env)["refreshToken"].(string)
	if reissued == "" {
		t.Fatalf("change password must re-issue tokens for the caller: %#v", env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
		map[string]any{"refreshToken": otherRefresh})
	if code != http.StatusUnauthorized {
		t.Fatalf("the other device must be revoked, got %d %#v", code, env)
	}

	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
		map[string]any{"refreshToken": reissued})
	if code != http.StatusOK {
		t.Fatalf("the caller's re-issued token must still refresh, got %d %#v", code, env)
	}
}

// /auth/refresh accepte un token volé en clair : sans limite de débit, c'est un
// oracle de validation utilisable en rafale.
func TestAuthRefreshIsRateLimited(t *testing.T) {
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "3")
	api := newTestAPI(t)

	for i := 0; i < 10; i++ {
		code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/refresh",
			map[string]any{"refreshToken": "forged." + uuid.NewString()})
		if code == http.StatusTooManyRequests {
			return
		}
		if code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: got %d %#v", i, code, env)
		}
	}
	t.Fatal("POST /auth/refresh must be rate limited like the other /auth/* routes")
}

func resetTokenFor(t *testing.T, api *testAPI, email string) string {
	t.Helper()
	var token string
	if err := api.pool.QueryRow(context.Background(), `
		SELECT t.token
		FROM identity.password_reset_tokens t
		JOIN identity.users u ON u.id = t.user_id
		WHERE u.email = $1 AND t.used_at IS NULL
		ORDER BY t.created_at DESC LIMIT 1`, email).Scan(&token); err != nil {
		t.Fatalf("reset token for %s: %v", email, err)
	}
	return token
}

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
