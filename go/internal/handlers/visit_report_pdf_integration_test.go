package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestVisitReportPDF(t *testing.T) {
	api := newTestAPI(t)
	api.api.TestSetAiCrAdvancedEnabled(true)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets", clientTok, nil)
	if code != 200 {
		t.Fatalf("pets %d %#v", code, env)
	}
	pets, _ := env["data"].([]any)
	if len(pets) == 0 {
		t.Fatal("no pets")
	}
	petID, _ := pets[0].(map[string]any)["id"].(string)

	slot := time.Now().UTC().Add(13 * time.Minute).Format(time.RFC3339)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/visits", vetTok, map[string]any{
		"scheduledAt":         slot,
		"notes":               "report pdf export",
		"durationMinutes":     30,
		"confirmDirect":       true,
		"silentConfirm":       true,
		"consultationSession": true,
		"callbackPhone":       "0470000001",
	})
	if code != 201 && code != 200 {
		t.Fatalf("create visit %d %#v", code, env)
	}
	visitID, _ := asMap(env["data"])["id"].(string)
	if visitID == "" {
		t.Fatalf("visit %#v", env)
	}
	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/visits/"+visitID, vetTok, nil)
	})

	path := "/api/v1/visits/" + visitID + "/report/pdf"

	// Aucun CR encore écrit : refus distinct du CR vide, sinon le message affiché
	// au véto (« Compte-rendu introuvable » vs « vide ») ne veut plus rien dire.
	code, body, hdr := doAuthBytes(t, api.handler, http.MethodGet, path, vetTok)
	if code != http.StatusNotFound || !strings.Contains(string(body), "report_not_found") {
		t.Fatalf("no report yet: want 404 report_not_found, got %d %s", code, body)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "   ",
	})
	if code != http.StatusOK {
		t.Fatalf("put empty report %d %#v", code, env)
	}
	code, body, _ = doAuthBytes(t, api.handler, http.MethodGet, path, vetTok)
	if code != http.StatusBadRequest || !strings.Contains(string(body), "report_empty") {
		t.Fatalf("empty report: want 400 report_empty, got %d %s", code, body)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": "## Anamnèse\n\nToux depuis 3 jours.",
	})
	if code != 200 {
		t.Fatalf("put report %d %#v", code, env)
	}
	reportID, _ := asMap(env["data"])["id"].(string)
	if reportID == "" {
		t.Fatalf("report id %#v", env)
	}

	code, body, hdr = doAuthBytes(t, api.handler, http.MethodGet, path, vetTok)
	if code != 200 {
		t.Fatalf("pdf %d %s", code, body)
	}
	ct := hdr.Get("Content-Type")
	if !strings.Contains(ct, "application/pdf") {
		t.Fatalf("content-type %q", ct)
	}
	cd := hdr.Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") || !strings.Contains(cd, ".pdf") {
		t.Fatalf("content-disposition %q", cd)
	}
	if len(body) < 100 || !strings.HasPrefix(string(body), "%PDF") {
		t.Fatalf("not a pdf (%d bytes, prefix %q)", len(body), string(body[:min(8, len(body))]))
	}

	code, body, _ = doAuthBytes(t, api.handler, http.MethodGet, path, clientTok)
	if code != 403 && code != 401 {
		t.Fatalf("client pdf want forbidden got %d %s", code, body)
	}

	// Citations from latest completed improve run appear on GET report.
	ctx := context.Background()
	st := store.New(api.pool)
	var practiceID, vetID string
	if err := api.pool.QueryRow(ctx, `
		SELECT u.id::text, u.practice_id::text
		FROM identity.users u WHERE u.email = 'vet.demo@petsfollow.test'`).Scan(&vetID, &practiceID); err != nil {
		t.Fatalf("vet ids: %v", err)
	}
	run, err := st.CreateImproveRun(ctx, visitID, reportID, practiceID, vetID)
	if err != nil {
		t.Fatalf("CreateImproveRun: %v", err)
	}
	if err := st.CompleteImproveRun(ctx, run.ID, store.ImproveRunCompleted, []string{"Guide BSAVA: cough"}, map[string]any{"ragHitCount": 1}, "", 12); err != nil {
		t.Fatalf("CompleteImproveRun: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/visits/"+visitID+"/report", vetTok, nil)
	if code != 200 {
		t.Fatalf("get report %d %#v", code, env)
	}
	cites, _ := asMap(env["data"])["lastImproveCitations"].([]any)
	if len(cites) == 0 {
		t.Fatalf("citations absentes: %#v", asMap(env["data"])["lastImproveCitations"])
	}
	if first, _ := cites[0].(string); !strings.Contains(first, "BSAVA") {
		t.Fatalf("citations %#v", cites)
	}

	// Export patient : stripCitations retire la section références RAG du PDF.
	withRefs := "## Anamnèse\n\nToux.\n\n## Références RAG\n\n- Guide BSAVA: cough\n"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/visits/"+visitID+"/report", vetTok, map[string]any{
		"bodyText": withRefs,
	})
	if code != 200 {
		t.Fatalf("put report refs %d %#v", code, env)
	}
	code, body, _ = doAuthBytes(t, api.handler, http.MethodGet, path+"?stripCitations=1", vetTok)
	if code != 200 || !strings.HasPrefix(string(body), "%PDF") {
		t.Fatalf("pdf stripCitations %d %s", code, body)
	}
	// PDF binaire : on vérifie juste que l'export ne plante pas ; strip unitaire = consultationpdf.TestStripCitations.
}
