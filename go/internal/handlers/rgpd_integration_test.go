package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// Export + DELETE /me sur un client jetable (pas de seed demo).
func TestRGPDExportAndDeleteDisposableClient(t *testing.T) {
	api := newTestAPI(t)
	email := uniqueEmail("rgpd-client")
	password := "ClientPass123!"

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "RGPD Disposable",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register-client %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	const prefix = "/confirm-email?token="
	if len(confirmPath) > len(prefix) {
		token = confirmPath[len(prefix):]
	}
	if token == "" {
		t.Fatalf("missing confirm token: %#v", env)
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", access, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	data := dataMap(t, env)
	profile, _ := data["profile"].(map[string]any)
	if profile == nil {
		t.Fatalf("export missing profile: %#v", data)
	}
	if _, hasHash := profile["password_hash"]; hasHash {
		t.Fatal("export must not include password_hash")
	}
	if _, ok := data["pets"]; !ok {
		t.Fatalf("client export missing pets key: %#v", data)
	}
	for _, key := range []string{"petDocuments", "deviceTokens"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("client export missing %s: %#v", key, data)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("delete me %d %#v", code, env)
	}
	if ok, _ := dataMap(t, env)["ok"].(bool); !ok {
		t.Fatalf("expected ok true: %#v", env)
	}

	// Login après purge doit échouer.
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	})
	if code == http.StatusOK {
		t.Fatalf("login after delete should fail, got 200 %#v", env)
	}
}

// DELETE /me assistant → tombstone (pas 404).
func TestRGPDDeleteAssistantTombstone(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	email := uniqueEmail("rgpd-assist")
	password := "VetDemo123!"
	id := insertVerifiedUser(t, api, string(kernel.RoleVetAssistant), email, password, "Assist RGPD", nil)

	tok := loginToken(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me assistant %d %#v", code, env)
	}
	var emailOut string
	if err := api.pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1`, id).Scan(&emailOut); err != nil {
		t.Fatal(err)
	}
	if !store.IsTombstoneEmail(emailOut) {
		t.Fatalf("want tombstone email, got %q", emailOut)
	}
}

// Dual care_pro + pet client : DELETE /me purge le pet puis tombstone.
func TestRGPDDeleteCareProPurgesOwnedPets(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	email := uniqueEmail("rgpd-care")
	password := "CareProDemo123!"
	userID := insertVerifiedUser(t, api, string(kernel.RoleCarePro), email, password, "Care Dual", map[string]any{
		"professional_specialty": "farrier",
	})
	if err := st.EnsureUserProfiles(ctx, userID); err != nil {
		t.Fatal(err)
	}
	petID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
		VALUES ($1, NULL, $2, 'DualPet', 'horse', 'mix', 'pending_payment')`, petID, userID); err != nil {
		t.Fatal(err)
	}

	tok := loginToken(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("export care_pro %d %#v", code, env)
	}
	exportData := dataMap(t, env)
	if _, ok := exportData["pets"]; !ok {
		t.Fatalf("dual care_pro export missing pets: %#v", exportData)
	}
	pets, _ := exportData["pets"].([]any)
	if len(pets) == 0 {
		t.Fatalf("dual care_pro export pets empty: %#v", exportData["pets"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me care_pro %d %#v", code, env)
	}
	var petN int
	if err := api.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pets.pets WHERE id=$1`, petID).Scan(&petN); err != nil {
		t.Fatal(err)
	}
	if petN != 0 {
		t.Fatalf("owned pet must be purged, count=%d", petN)
	}
	var emailOut string
	if err := api.pool.QueryRow(ctx, `SELECT email FROM identity.users WHERE id=$1`, userID).Scan(&emailOut); err != nil {
		t.Fatal(err)
	}
	if !store.IsTombstoneEmail(emailOut) {
		t.Fatalf("want tombstone email, got %q", emailOut)
	}
}

// Clients provisionnés : termsAcceptedAt null jusqu'à POST /me/accept-terms.
func TestRGPDAcceptTermsProvisionedClient(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	email := uniqueEmail("rgpd-prov")
	password := "ClientDemo123!"
	userID := insertVerifiedUser(t, api, "client", email, password, "Prov Client", nil)
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users SET terms_accepted_at = NULL WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}

	tok := loginTokenRaw(t, api.handler, email, password)
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("me %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["termsAcceptedAt"] != nil {
		t.Fatalf("want termsAcceptedAt null, got %#v", me["termsAcceptedAt"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": false,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("accept-terms without consent want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": true,
	})
	if code != http.StatusOK {
		t.Fatalf("accept-terms %d %#v", code, env)
	}
	me = dataMap(t, env)
	if me["termsAcceptedAt"] == nil || me["termsAcceptedAt"] == "" {
		t.Fatalf("want termsAcceptedAt set, got %#v", me)
	}

	// API gated : pets hors allowlist tant que consent manquant (re-null puis 403).
	if _, err := api.pool.Exec(ctx, `
		UPDATE identity.users SET terms_accepted_at = NULL WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", tok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("pets without terms want 403 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj != nil {
		if c, _ := errObj["code"].(string); c != "consent_required" {
			t.Fatalf("want error.code consent_required got %#v", errObj)
		}
	} else {
		t.Fatalf("want error object with consent_required, got %#v", env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/app-invite", tok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("app-invite without terms want 403 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/me/accept-terms", tok, map[string]any{
		"consent": true,
	})
	if code != http.StatusOK {
		t.Fatalf("re-accept %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("pets after accept want 200 got %d %#v", code, env)
	}
}

// DELETE /me client redacts pharmacy.job_audit payloads linked via DAF.client_user_id.
func TestRGPDDeleteClientRedactsPharmacyJobAudit(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()

	email := uniqueEmail("rgpd-pharm-client")
	password := "ClientPass123!"
	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/register-client", map[string]any{
		"email": email, "password": password, "fullName": "RGPD Pharm",
		"consent": true,
	})
	if code != http.StatusCreated {
		t.Fatalf("register %d %#v", code, env)
	}
	confirmPath, _ := dataMap(t, env)["confirmPath"].(string)
	token := ""
	const prefix = "/confirm-email?token="
	if len(confirmPath) > len(prefix) {
		token = confirmPath[len(prefix):]
	}
	code, env = doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/confirm-email", map[string]any{"token": token})
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	access, _ := dataMap(t, env)["accessToken"].(string)
	if access == "" {
		access = loginToken(t, api.handler, email, password)
	}
	var clientID string
	if err := api.pool.QueryRow(ctx, `SELECT id::text FROM identity.users WHERE email=$1`, email).Scan(&clientID); err != nil {
		t.Fatal(err)
	}

	var vetID, practiceID string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text, practice_id::text FROM identity.users WHERE email='vet.demo@petsfollow.test'`).
		Scan(&vetID, &practiceID); err != nil {
		t.Fatal(err)
	}

	dafID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pharmacy.daf_documents (
			id, practice_id, daf_year, status, client_user_id, prescriber_user_id,
			has_antibiotic, vamreg_status
		) VALUES ($1,$2,EXTRACT(YEAR FROM now())::int,'draft',$3,$4,true,'pending')`,
		dafID, practiceID, clientID, vetID); err != nil {
		t.Fatalf("insert daf: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pharmacy.job_audit (
			id, practice_id, job_type, entity_id, attempt, status, request_json, error
		) VALUES ($1,$2,'vamreg',$3,1,'failed',$4::jsonb,'boom')`,
		uuid.NewString(), practiceID, dafID,
		`{"species":"dog","ownerEmail":"secret@petsfollow.test"}`); err != nil {
		t.Fatalf("insert audit: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", access, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me %d %#v", code, env)
	}

	var reqJSON, errMsg string
	var clientLeft *string
	if err := api.pool.QueryRow(ctx, `
		SELECT ja.request_json::text, COALESCE(ja.error,''), d.client_user_id::text
		FROM pharmacy.job_audit ja
		JOIN pharmacy.daf_documents d ON d.id = ja.entity_id
		WHERE ja.entity_id = $1 AND ja.job_type = 'vamreg'
		LIMIT 1`, dafID).Scan(&reqJSON, &errMsg, &clientLeft); err != nil {
		t.Fatal(err)
	}
	if reqJSON != "{}" {
		t.Fatalf("want redacted request_json {}, got %s", reqJSON)
	}
	if errMsg != "redacted" {
		t.Fatalf("want error redacted, got %q", errMsg)
	}
	if clientLeft != nil {
		t.Fatalf("client_user_id must be cleared, got %v", clientLeft)
	}
}

// Export includes ragImproveRuns; Pro tombstone purges them (CR visit_reports stay).
func TestRGPDImproveRunsExportAndProPurge(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	adminTok := ensureAdminToken(t, api)
	_, _, commTok := createCommercial(t, api, adminTok, "rgpd-ir", "RGPD IR Comm")

	vetEmail := uniqueEmail("rgpd-ir-v")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr RGPD IR",
		"practiceName": "Cab RGPD IR",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)
	if vetID == "" {
		t.Fatalf("missing vet userId %#v", env)
	}
	vetTok := loginToken(t, api.handler, vetEmail, "VetDemo123!")

	var practiceID string
	if err := api.pool.QueryRow(ctx, `
		SELECT practice_id::text FROM identity.users WHERE id = $1`, vetID).Scan(&practiceID); err != nil || practiceID == "" {
		t.Fatalf("practice for vet: %v %q", err, practiceID)
	}

	clientID := insertVerifiedUser(t, api, string(kernel.RoleClient), uniqueEmail("rgpd-ir-c"), "ClientDemo123!", "Cli IR", nil)
	petID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO pets.pets (id, practice_id, owner_user_id, name, species, breed, payment_status)
		VALUES ($1, $2, $3, 'IRPet', 'dog', 'mix', 'pending_payment')`, petID, practiceID, clientID); err != nil {
		t.Fatalf("insert pet: %v", err)
	}
	visitID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO visits.visits (id, pet_id, practice_id, site_id, scheduled_at, status, notes, source, created_at)
		VALUES ($1, $2, $3,
			(SELECT id FROM practice.sites WHERE practice_id = $3::uuid AND is_primary LIMIT 1),
			NOW(), 'confirmed', 'rgpd improve run', 'vet', NOW())`,
		visitID, petID, practiceID); err != nil {
		t.Fatalf("insert visit: %v", err)
	}

	st := store.New(api.pool)
	rep, err := st.EnsureVisitReport(ctx, visitID, vetID)
	if err != nil {
		t.Fatalf("EnsureVisitReport: %v", err)
	}
	run, err := st.CreateImproveRun(ctx, visitID, rep.ID, practiceID, vetID)
	if err != nil {
		t.Fatalf("CreateImproveRun: %v", err)
	}
	if err := st.AppendImproveRunStep(ctx, run.ID, map[string]any{"agent": "tri", "label": "phi-step"}); err != nil {
		t.Fatalf("AppendImproveRunStep: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("export %d %#v", code, env)
	}
	data := dataMap(t, env)
	runs, ok := data["ragImproveRuns"].([]any)
	if !ok || len(runs) == 0 {
		t.Fatalf("export missing ragImproveRuns: %#v", data["ragImproveRuns"])
	}
	found := false
	for _, raw := range runs {
		m, _ := raw.(map[string]any)
		if str(m["id"]) == run.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("export ragImproveRuns missing run %s: %#v", run.ID, runs)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me vet %d %#v", code, env)
	}
	var n int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM rag.improve_runs WHERE user_id = $1`, vetID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("improve_runs after pro tombstone count=%d want 0", n)
	}
	var reportLeft int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.visit_reports WHERE id = $1`, rep.ID).Scan(&reportLeft); err != nil {
		t.Fatal(err)
	}
	if reportLeft != 1 {
		t.Fatalf("visit_reports should remain after pro tombstone, count=%d", reportLeft)
	}
}

func TestRGPDRAGDocumentsUnlinkedOnProTombstone(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	adminTok := ensureAdminToken(t, api)
	_, _, commTok := createCommercial(t, api, adminTok, "rgpd-rag", "RGPD RAG Comm")

	vetEmail := uniqueEmail("rgpd-rag-v")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/commercial/vets", commTok, map[string]any{
		"email": vetEmail, "password": "VetDemo123!", "fullName": "Dr RGPD RAG",
		"practiceName": "Cab RGPD RAG",
	})
	if code != http.StatusCreated {
		t.Fatalf("encode vet %d %#v", code, env)
	}
	vetID, _ := dataMap(t, env)["userId"].(string)
	if vetID == "" {
		t.Fatalf("missing vet userId %#v", env)
	}
	vetTok := loginToken(t, api.handler, vetEmail, "VetDemo123!")

	var practiceID string
	if err := api.pool.QueryRow(ctx, `
		SELECT practice_id::text FROM identity.users WHERE id = $1`, vetID).Scan(&practiceID); err != nil || practiceID == "" {
		t.Fatalf("practice for vet: %v %q", err, practiceID)
	}

	pendingID := uuid.NewString()
	readyID := uuid.NewString()
	indexingID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO rag.documents (
			id, scope, practice_id, title, filename, mime_type, content_sha256,
			source_object_key, byte_size, status, uploaded_by
		) VALUES
		($1::uuid, 'practice', $4::uuid, 'Pending', 'p.pdf', 'application/pdf', 'sha-p',
		 $6, 10, 'pending', $5::uuid),
		($2::uuid, 'practice', $4::uuid, 'Ready', 'r.pdf', 'application/pdf', 'sha-r',
		 $7, 10, 'ready', $5::uuid),
		($3::uuid, 'practice', $4::uuid, 'Indexing', 'i.pdf', 'application/pdf', 'sha-i',
		 $8, 10, 'indexing', $5::uuid)`,
		pendingID, readyID, indexingID, practiceID, vetID,
		"rag-docs/"+pendingID+".pdf", "rag-docs/"+readyID+".pdf", "rag-docs/"+indexingID+".pdf"); err != nil {
		t.Fatalf("insert rag docs: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me vet %d %#v", code, env)
	}

	var pendingLeft, indexingLeft int
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM rag.documents WHERE id = $1`, pendingID).Scan(&pendingLeft); err != nil {
		t.Fatal(err)
	}
	if pendingLeft != 0 {
		t.Fatalf("pending rag doc should be deleted, count=%d", pendingLeft)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM rag.documents WHERE id = $1`, indexingID).Scan(&indexingLeft); err != nil {
		t.Fatal(err)
	}
	if indexingLeft != 0 {
		t.Fatalf("indexing rag doc should be deleted, count=%d", indexingLeft)
	}
	var uploader any
	if err := api.pool.QueryRow(ctx, `
		SELECT uploaded_by FROM rag.documents WHERE id = $1`, readyID).Scan(&uploader); err != nil {
		t.Fatal(err)
	}
	if uploader != nil {
		t.Fatalf("ready rag uploaded_by must be null after tombstone, got %#v", uploader)
	}
}

func TestRGPDCommercialCRMRedactOnTombstone(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	adminTok := ensureAdminToken(t, api)
	commID, _, commTok := createCommercial(t, api, adminTok, "rgpd-crm", "RGPD CRM Comm")

	prospectID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, status, source
		) VALUES ($1, $2, 'Cab CRM', 'Contact', 'contact@example.test', 'new', 'commercial')`,
		prospectID, commID); err != nil {
		t.Fatalf("insert prospect: %v", err)
	}
	sendID := uuid.NewString()
	clickID := uuid.NewString()
	actID := uuid.NewString()
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.email_sends (
			id, prospect_id, commercial_user_id, to_email, subject, body_html_rendered,
			status, open_token, sent_at
		) VALUES ($1, $2, $3, 'prospect@example.test', 'Hello PHI', '<p>secret</p>',
			'sent', $4, NOW())`,
		sendID, prospectID, commID, "open-"+sendID); err != nil {
		t.Fatalf("insert send: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.email_clicks (id, send_id, click_token, target_url)
		VALUES ($1, $2, $3, 'https://example.com/track-me')`,
		clickID, sendID, "click-"+clickID); err != nil {
		t.Fatalf("insert click: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.prospect_events (id, prospect_id, actor_user_id, kind, body)
		VALUES ($1, $2, $3, 'note', 'private note')`,
		uuid.NewString(), prospectID, commID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if _, err := api.pool.Exec(ctx, `
		INSERT INTO sales.activities (
			id, prospect_id, assignee_user_id, created_by, kind, title, status
		) VALUES ($1, $2, $3, $3, 'follow_up', 'Call prospect', 'open')`,
		actID, prospectID, commID); err != nil {
		t.Fatalf("insert activity: %v", err)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/me", commTok, nil)
	if code != http.StatusOK {
		t.Fatalf("DELETE /me commercial %d %#v", code, env)
	}

	var toEmail, subject, body, target, title, eventBody string
	var assignee any
	var actor any
	if err := api.pool.QueryRow(ctx, `
		SELECT to_email, subject, body_html_rendered FROM sales.email_sends WHERE id = $1`, sendID).
		Scan(&toEmail, &subject, &body); err != nil {
		t.Fatal(err)
	}
	if toEmail != "[redacted]" || subject != "[redacted]" || body != "" {
		t.Fatalf("send not redacted: %q %q %q", toEmail, subject, body)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT target_url FROM sales.email_clicks WHERE id = $1`, clickID).Scan(&target); err != nil {
		t.Fatal(err)
	}
	if target != "[redacted]" {
		t.Fatalf("click target_url want [redacted] got %q", target)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT actor_user_id, body FROM sales.prospect_events WHERE prospect_id = $1`, prospectID).
		Scan(&actor, &eventBody); err != nil {
		t.Fatal(err)
	}
	if actor != nil {
		t.Fatalf("event actor must be null, got %#v", actor)
	}
	if eventBody != "[redacted]" {
		t.Fatalf("event body want [redacted] got %q", eventBody)
	}
	// COALESCE scan path (nullable assignee after 000177) must not break list/get.
	var assigneeText string
	if err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(assignee_user_id::text,''), title FROM sales.activities WHERE id = $1`, actID).
		Scan(&assigneeText, &title); err != nil {
		t.Fatalf("activity coalesce scan: %v", err)
	}
	if assigneeText != "" {
		t.Fatalf("activity assignee must be empty, got %q", assigneeText)
	}
	if title != "[redacted]" {
		t.Fatalf("activity title want [redacted] got %q", title)
	}
	if err := api.pool.QueryRow(ctx, `
		SELECT assignee_user_id FROM sales.activities WHERE id = $1`, actID).Scan(&assignee); err != nil {
		t.Fatal(err)
	}
	if assignee != nil {
		t.Fatalf("activity assignee must be null, got %#v", assignee)
	}

	st := store.New(api.pool)
	got, err := st.GetActivity(ctx, actID)
	if err != nil {
		t.Fatalf("GetActivity after null assignee: %v", err)
	}
	if got.AssigneeUserID != "" || got.Title != "[redacted]" {
		t.Fatalf("GetActivity %#v", got)
	}
	list, err := st.ListActivitiesByProspect(ctx, prospectID, false)
	if err != nil {
		t.Fatalf("ListActivitiesByProspect: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected activity still listed on prospect")
	}
}
