package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/phpdave11/gofpdf"
)

func TestCompendiumImportFlow(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://localhost:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	api.api.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{CNK: "2888001", Name: "Compendium Demo Med", ATCCode: "J01CA04", IsAntibiotic: true, SourcePage: &start},
			{CNK: "2888001", Name: "Dup CNK ignored", IsAntibiotic: true},
			{CNK: "", Name: "Sans CNK", IsAntibiotic: false},
		}, nil
	})

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := mustCompendiumTestPDF(t, 2)
	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	job, _ := env["data"].(map[string]any)
	jobID, _ := job["id"].(string)
	if jobID == "" {
		t.Fatalf("no job id %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/extract", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("extract %d %#v", code, env)
	}

	st := store.New(api.pool)
	deadline := time.Now().Add(5 * time.Second)
	var detail store.CompendiumImportDetail
	for time.Now().Before(deadline) {
		detail, err = st.GetCompendiumImportDetail(context.Background(), jobID)
		if err != nil {
			t.Fatalf("detail: %v", err)
		}
		if detail.Job.Status == "extracted" || detail.Job.Status == "failed" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if detail.Job.Status != "extracted" {
		t.Fatalf("expected extracted, got %s err=%s", detail.Job.Status, detail.Job.ErrorMessage)
	}
	if detail.Job.ExtractPct != 100 {
		t.Fatalf("extractPct=%d", detail.Job.ExtractPct)
	}
	if detail.Job.ReviewPct != 0 {
		t.Fatalf("reviewPct after AI should be 0, got %d", detail.Job.ReviewPct)
	}
	// Dedup dropped duplicate CNK → 2 rows (ready + missing_cnk)
	if len(detail.Rows) != 2 {
		t.Fatalf("rows=%d %#v", len(detail.Rows), detail.Rows)
	}
	var readyID, errID string
	for _, r := range detail.Rows {
		if r.Status == "ready" {
			readyID = r.ID
		}
		if r.Status == "error" && r.ErrorCode == "missing_cnk" {
			errID = r.ID
		}
	}
	if readyID == "" || errID == "" {
		t.Fatalf("expected ready+missing_cnk %#v", detail.Rows)
	}

	// Human control: fix missing CNK
	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+errID, adminTok,
		map[string]any{"cnk": "2888002", "atcCode": "J01AA02", "isAntibiotic": true})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	row, _ := env["data"].(map[string]any)
	if row["status"] != "ready" {
		t.Fatalf("expected ready after patch %#v", row)
	}
	if row["humanReviewed"] != true {
		t.Fatalf("expected humanReviewed %#v", row)
	}

	detail, _ = st.GetCompendiumImportDetail(context.Background(), jobID)
	if detail.Job.ReviewPct != 50 {
		t.Fatalf("reviewPct=%d want 50 after 1/2 human patches", detail.Job.ReviewPct)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/commit", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("commit %d %#v", code, env)
	}

	hits, err := st.SearchRefMedications(context.Background(), "2888001", 10)
	if err != nil || len(hits) < 1 {
		t.Fatalf("search after commit: %v %#v", err, hits)
	}
	if hits[0].CNK != "2888001" {
		t.Fatalf("unexpected %#v", hits[0])
	}
}

func TestCompendiumCommitRecoversAfterBatchFail(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://localhost:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	api.api.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{CNK: "2888101", Name: "Recover Batch Med", SourcePage: &start},
		}, nil
	})

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := mustCompendiumTestPDF(t, 2)
	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["id"].(string)

	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/extract", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("extract %d", code)
	}

	st := store.New(api.pool)
	deadline := time.Now().Add(5 * time.Second)
	var detail store.CompendiumImportDetail
	for time.Now().Before(deadline) {
		detail, err = st.GetCompendiumImportDetail(context.Background(), jobID)
		if err != nil {
			t.Fatalf("detail: %v", err)
		}
		if detail.Job.Status == "extracted" || detail.Job.Status == "failed" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if detail.Job.Status != "extracted" {
		t.Fatalf("expected extracted, got %s err=%s", detail.Job.Status, detail.Job.ErrorMessage)
	}

	st.TestFailNextCompendiumCommitBatch()
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/commit", adminTok, nil)
	if code != http.StatusInternalServerError {
		t.Fatalf("forced fail commit want 500 got %d %#v", code, env)
	}
	job, err := st.GetCompendiumImportJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("job: %v", err)
	}
	if job.Status != "extracted" {
		t.Fatalf("after batch fail want extracted (retryable), got %s", job.Status)
	}
	if job.ErrorMessage == "" {
		t.Fatal("expected error_message after batch fail")
	}

	// Stale committing recovery (crash mid-commit): reclaim after 5m
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'committing', updated_at = now() - interval '10 minutes'
		WHERE id = $1`, jobID); err != nil {
		t.Fatalf("force stale committing: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/commit", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("retry commit from stale committing %d %#v", code, env)
	}
	job, _ = st.GetCompendiumImportJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Fatalf("want completed got %s", job.Status)
	}
}

func TestCompendiumCommitRejectsFreshCommitting(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://localhost:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)
	api.api.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{CNK: "2888102", Name: "Fresh Commit Guard", SourcePage: &start},
		}, nil
	})

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := mustCompendiumTestPDF(t, 2)
	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["id"].(string)

	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/extract", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("extract %d", code)
	}
	st := store.New(api.pool)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := st.GetCompendiumImportJob(context.Background(), jobID)
		if err != nil {
			t.Fatalf("job: %v", err)
		}
		if job.Status == "extracted" || job.Status == "failed" {
			if job.Status != "extracted" {
				t.Fatalf("expected extracted got %s", job.Status)
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'committing', updated_at = now()
		WHERE id = $1`, jobID); err != nil {
		t.Fatalf("force fresh committing: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/commit", adminTok, nil)
	if code != http.StatusConflict {
		t.Fatalf("fresh committing want 409 got %d %#v", code, env)
	}
}

func TestCompendiumImportACL(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	pdf := mustCompendiumTestPDF(t, 2)
	code, env := doCompendiumUpload(t, api.handler, vetTok, pdf, 1, 2)
	if code != http.StatusForbidden || errCode(env) != "forbidden" {
		t.Fatalf("vet upload want 403 forbidden got %d %#v", code, env)
	}
}

func mustCompendiumTestPDF(t *testing.T, pages int) []byte {
	t.Helper()
	pdf := gofpdf.New("P", "mm", "A4", "")
	for i := 0; i < pages; i++ {
		pdf.AddPage()
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, "page")
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func doCompendiumUpload(t *testing.T, h http.Handler, token string, pdf []byte, pageStart, pageEnd int) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("pageStart", strconv.Itoa(pageStart))
	_ = w.WriteField("pageEnd", strconv.Itoa(pageEnd))
	part, err := w.CreateFormFile("file", "demo.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pdf); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/compendium-imports", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var env map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	return rec.Code, env
}
