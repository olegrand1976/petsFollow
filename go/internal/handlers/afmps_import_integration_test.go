package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestAFMPSImportTripleGate(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	csv := "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Code ATC;Usage Humain/Vétérinaire\n" +
		"AFMPS Gate Med;Gélule;10;2888111;Lab;BE-V1;Oui;QJ01CA04;Usage vétérinaire\n"

	code, env := doAFMPSUpload(t, api.handler, adminTok, []byte(csv), "afmps-test.csv")
	if code != http.StatusCreated {
		t.Fatalf("upload/validate %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	job, _ := data["job"].(map[string]any)
	jobID, _ := job["id"].(string)
	if jobID == "" {
		t.Fatalf("no job %#v", env)
	}
	if job["status"] != "validated" {
		t.Fatalf("status=%v", job["status"])
	}

	// Gate 3 blocked before review
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/commit", adminTok, map[string]any{
		"confirm": pharmacy.AFMPSConfirmPhrase,
	})
	if code != http.StatusConflict {
		t.Fatalf("commit before review want 409 got %d %#v", code, env)
	}

	// Exclude then ensure KPIs refresh (ready/insert drop)
	detailCode, detailEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/afmps-imports/"+jobID+"?limit=50&offset=0", adminTok, nil)
	if detailCode != http.StatusOK {
		t.Fatalf("detail %d %#v", detailCode, detailEnv)
	}
	detailData, _ := detailEnv["data"].(map[string]any)
	detailRows, _ := detailData["rows"].([]any)
	if len(detailRows) == 0 {
		t.Fatalf("expected staged rows %#v", detailEnv)
	}
	row0, _ := detailRows[0].(map[string]any)
	rowID, _ := row0["id"].(string)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/afmps-imports/"+jobID+"/rows/"+rowID, adminTok, map[string]any{
		"exclude": true,
	})
	if code != http.StatusOK {
		t.Fatalf("exclude %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/afmps-imports/"+jobID, adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("detail after exclude %d %#v", code, env)
	}
	jobAfter, _ := env["data"].(map[string]any)["job"].(map[string]any)
	if intFromAny(jobAfter["readyCount"]) != 0 {
		t.Fatalf("ready after exclude want 0 got %#v", jobAfter["readyCount"])
	}
	// Re-include for commit path
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/afmps-imports/"+jobID+"/rows/"+rowID, adminTok, map[string]any{
		"exclude": false,
	})
	if code != http.StatusOK {
		t.Fatalf("include %d %#v", code, env)
	}

	// Gate 2
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/mark-reviewed", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("mark-reviewed %d %#v", code, env)
	}

	// Bad confirm
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/commit", adminTok, map[string]any{
		"confirm": "yes",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("bad confirm %d %#v", code, env)
	}

	// Gate 3
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/commit", adminTok, map[string]any{
		"confirm": "IMPORT_AFMPS",
	})
	if code != http.StatusOK {
		t.Fatalf("commit %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	result, _ := data["result"].(map[string]any)
	if intFromAny(result["upserted"]) != 1 {
		t.Fatalf("upserted %#v", result)
	}

	st := store.New(api.pool)
	found, err := st.SearchRefMedications(context.Background(), "2888111", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 || found[0].CNK != "2888111" {
		t.Fatalf("ref missing %#v", found)
	}
	if !found[0].IsAntibiotic {
		t.Fatal("expected antibiotic from ATC")
	}
}

func TestAFMPSImportMetaMerge(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	st := store.New(api.pool)
	cnk := "2888222"
	_, err := st.UpsertRefMedication(context.Background(), store.RefMedicationUpsert{
		CNK: cnk, Name: "Prior Comp Med", IsActive: true,
		AFMPSMeta: json.RawMessage(`{"source":"compendium-pdf","keep":"yes"}`),
	})
	if err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	csv := "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Commercialisé;Code ATC;Usage Humain/Vétérinaire\n" +
		"AFMPS Merge Med;Gélule;10;" + cnk + ";Lab;Oui;QJ01CA04;Usage vétérinaire\n"
	code, env := doAFMPSUpload(t, api.handler, adminTok, []byte(csv), "merge.csv")
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["job"].(map[string]any)["id"].(string)
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/mark-reviewed", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("review %d", code)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/commit", adminTok, map[string]any{
		"confirm": "IMPORT_AFMPS",
	})
	if code != http.StatusOK {
		t.Fatalf("commit %d %#v", code, env)
	}

	var meta []byte
	err = api.pool.QueryRow(context.Background(), `
		SELECT afmps_meta FROM pharmacy.ref_medications WHERE cnk = $1`, cnk).Scan(&meta)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		t.Fatal(err)
	}
	if m["keep"] != "yes" {
		t.Fatalf("expected prior meta keep=yes, got %#v", m)
	}
	if m["source"] != "afmps-pack-csv" && m["source"] != "compendium-pdf" {
		// jsonb || : right-hand keys win for conflicts → source should be afmps-pack-csv
		t.Fatalf("meta source %#v", m["source"])
	}
	if m["source"] != "afmps-pack-csv" {
		t.Fatalf("expected afmps source override, got %#v", m)
	}
	if m["manufacturer"] != "Lab" {
		t.Fatalf("expected manufacturer from AFMPS, got %#v", m)
	}
}

func TestAFMPSImportGate1BlockedNoReady(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	csv := "cnk;name\n;Empty CNK\nabc;Bad\n"
	code, env := doAFMPSUpload(t, api.handler, adminTok, []byte(csv), "bad.csv")
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422 got %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	job, _ := data["job"].(map[string]any)
	if job["status"] != "blocked" {
		t.Fatalf("%#v", job)
	}
	jobID, _ := job["id"].(string)
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/mark-reviewed", adminTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("mark-reviewed on blocked want 409 got %d", code)
	}
}

func TestInternalAfmpsImportRunGate1(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("AFMPS_IMPORT_SECRET", "test-afmps-import-secret")
	t.Setenv("AFMPS_IMPORT_OBJECT_KEY", "afmps-imports/latest.csv")
	t.Setenv("OPS_NOTIFY_EMAIL", "ops@petsfollow.test")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{MediaLocalDir: t.TempDir(), APIPublicURL: "http://localhost:8291"})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	csv := "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Code ATC;Usage Humain/Vétérinaire\n" +
		"AFMPS Cron Med;Gélule;10;2888444;Lab;BE-V2;Oui;QJ01CA04;Usage vétérinaire\n"
	if _, err := bundle.Store.Upload(context.Background(), "afmps-imports/latest.csv", bytes.NewReader([]byte(csv)), int64(len(csv)), "text/csv"); err != nil {
		t.Fatalf("upload csv: %v", err)
	}

	clearOpenAFMPSJobsForCronTest(t, api)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Afmps-Import-Secret", "test-afmps-import-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("run %d %s", rec.Code, rec.Body.String())
	}
	var env map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	data, _ := env["data"].(map[string]any)
	if data["skipped"] == true {
		t.Fatalf("want not skipped %#v", data)
	}
	job, _ := data["job"].(map[string]any)
	if job["status"] != "validated" {
		t.Fatalf("status=%v", job["status"])
	}
	jobID, _ := job["id"].(string)
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pharmacy.afmps_import_jobs WHERE id = $1`, jobID)
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pharmacy.ref_medications WHERE cnk = '2888444'`)
	})

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Afmps-Import-Secret", "test-afmps-import-secret")
	api.handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("pending run %d %s", rec2.Code, rec2.Body.String())
	}
	var env2 map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &env2)
	data2, _ := env2["data"].(map[string]any)
	if data2["skipped"] != true || data2["reason"] != "pending_job" {
		t.Fatalf("want pending_job skip %#v", data2)
	}
}

func TestInternalAfmpsImportRunObjectMissing(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("AFMPS_IMPORT_SECRET", "test-afmps-import-secret")
	t.Setenv("AFMPS_IMPORT_OBJECT_KEY", "afmps-imports/missing-test.csv")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{MediaLocalDir: t.TempDir(), APIPublicURL: "http://localhost:8291"})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)
	clearOpenAFMPSJobsForCronTest(t, api)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Afmps-Import-Secret", "test-afmps-import-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("run %d %s", rec.Code, rec.Body.String())
	}
	var env map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	data, _ := env["data"].(map[string]any)
	if data["skipped"] != true || data["reason"] != "object_missing" {
		t.Fatalf("want object_missing %#v", data)
	}
}

func TestInternalAfmpsImportRunUnauthorized(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("AFMPS_IMPORT_SECRET", "test-afmps-import-secret")
	api := newTestAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Afmps-Import-Secret", "wrong")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rec.Code)
	}
}

func TestInternalAfmpsImportRunDisabled(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "false")
	t.Setenv("AFMPS_IMPORT_SECRET", "test-afmps-import-secret")
	api := newTestAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Afmps-Import-Secret", "test-afmps-import-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", rec.Code)
	}
}

func TestInternalAfmpsImportRunUnchangedChecksum(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("AFMPS_IMPORT_SECRET", "test-afmps-import-secret")
	t.Setenv("AFMPS_IMPORT_OBJECT_KEY", "afmps-imports/latest.csv")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{MediaLocalDir: t.TempDir(), APIPublicURL: "http://localhost:8291"})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	csv := "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Code ATC;Usage Humain/Vétérinaire\n" +
		"AFMPS Unchanged Med;Gélule;10;2888555;Lab;BE-V3;Oui;QJ01CA04;Usage vétérinaire\n"
	clearOpenAFMPSJobsForCronTest(t, api)
	_, _ = api.pool.Exec(context.Background(), `DELETE FROM pharmacy.ref_medications WHERE cnk = '2888555'`)

	code, env := doAFMPSUpload(t, api.handler, adminTok, []byte(csv), "afmps-unchanged.csv")
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["job"].(map[string]any)["id"].(string)
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/mark-reviewed", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("review %d", code)
	}
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/afmps-imports/"+jobID+"/commit", adminTok, map[string]any{
		"confirm": "IMPORT_AFMPS",
	})
	if code != http.StatusOK {
		t.Fatalf("commit %d", code)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pharmacy.afmps_import_jobs WHERE id = $1`, jobID)
		_, _ = api.pool.Exec(context.Background(), `DELETE FROM pharmacy.ref_medications WHERE cnk = '2888555'`)
	})

	if _, err := bundle.Store.Upload(context.Background(), "afmps-imports/latest.csv", bytes.NewReader([]byte(csv)), int64(len(csv)), "text/csv"); err != nil {
		t.Fatalf("upload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/afmps-import/run", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Afmps-Import-Secret", "test-afmps-import-secret")
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("run %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	data, _ := out["data"].(map[string]any)
	if data["skipped"] != true || data["reason"] != "unchanged_checksum" {
		t.Fatalf("want unchanged_checksum %#v", data)
	}
}

func doAFMPSUpload(t *testing.T, h http.Handler, token string, csv []byte, filename string) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(csv); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/afmps-imports", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var env map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &env)
	return rr.Code, env
}

// clearOpenAFMPSJobsForCronTest removes open jobs that would block the internal
// cron tests: tiny fixtures, or system/CLI jobs (created_by_admin_id NULL).
// Admin mid-review uploads (with admin id + typically larger review) are kept.
func clearOpenAFMPSJobsForCronTest(t *testing.T, api *testAPI) {
	t.Helper()
	_, err := api.pool.Exec(context.Background(), `
		DELETE FROM pharmacy.afmps_import_jobs
		WHERE status IN ('validated', 'reviewed', 'blocked')
		  AND (file_bytes < 100000 OR created_by_admin_id IS NULL)`)
	if err != nil {
		t.Fatalf("clear open afmps jobs for cron test: %v", err)
	}
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}
