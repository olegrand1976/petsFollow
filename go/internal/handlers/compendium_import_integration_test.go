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

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
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

	// DB dev partagée : le commit d'un run précédent laisse un « ZZZ Unmatched
	// Orphan … » (CNK 2888002) dans ref_medications que le matcher fuzzy
	// re-suggère au run suivant — l'orphelin ne serait plus « error ».
	cleanupTestRefMeds := func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM pharmacy.ref_medications
			 WHERE cnk IN ('2888001','2888002') OR name LIKE 'ZZZ Unmatched Orphan %'`)
	}
	cleanupTestRefMeds()
	t.Cleanup(cleanupTestRefMeds)

	orphanName := "ZZZ Unmatched Orphan " + uuid.NewString()[:8]
	handlers.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{CNK: "2888001", Name: "Compendium Demo Med", ATCCode: "J01CA04", IsAntibiotic: true, SourcePage: &start},
			{CNK: "", Name: orphanName, IsAntibiotic: false},
		}, nil
	})
	t.Cleanup(handlers.TestClearCompendiumExtract)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
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
	if len(detail.Rows) != 2 {
		t.Fatalf("rows=%d %#v", len(detail.Rows), detail.Rows)
	}
	var pendingID, errID string
	for _, r := range detail.Rows {
		if r.Status == "pending" && r.CNK == "2888001" {
			pendingID = r.ID
		}
		// No AFMPS hit → error (missing_cnk / cnk_unmatched) until human PATCH.
		if r.Status == "error" && (r.ErrorCode == "missing_cnk" || r.ErrorCode == "cnk_unmatched") {
			errID = r.ID
		}
	}
	if pendingID == "" || errID == "" {
		t.Fatalf("expected pending(with cnk)+unmatched orphan %#v", detail.Rows)
	}
	if detail.Job.ReviewPct != 0 {
		t.Fatalf("reviewPct want 0 after extract, got %d", detail.Job.ReviewPct)
	}

	// Human control: fix missing CNK (PATCH confirms → ready)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+errID, adminTok,
		map[string]any{"cnk": "2888002"})
	if code != http.StatusOK {
		t.Fatalf("patch %d %#v", code, env)
	}
	row, _ := env["data"].(map[string]any)
	if row["status"] != "ready" {
		t.Fatalf("expected ready after patch %#v", row)
	}

	// Bulk confirm remaining pending rows
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/confirm-ready", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}

	detail, _ = st.GetCompendiumImportDetail(context.Background(), jobID)
	if detail.Job.ReviewPct < 100 {
		t.Fatalf("reviewPct=%d ready=%d rows=%d reviewed=%d", detail.Job.ReviewPct, detail.Job.ReadyCount, detail.Job.RowCount, detail.Job.ReviewedCount)
	}
	if detail.Job.ReadyCount < 2 {
		t.Fatalf("readyCount=%d", detail.Job.ReadyCount)
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

func TestCompendiumImportCNKMatch(t *testing.T) {
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

	st := store.New(api.pool)
	if _, err := st.UpsertRefMedication(context.Background(), store.RefMedicationUpsert{
		CNK:                "2712345",
		Name:               "Vetoryl 20 mg capsules",
		PharmaceuticalForm: "gélule",
		PackSize:           "30",
		IsActive:           true,
		AFMPSMeta:          json.RawMessage(`{"manufacturer":"Dechra"}`),
	}); err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	handlers.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{
				Name:               "VETORYL 20 mg caps",
				Manufacturer:       "Dechra",
				PharmaceuticalForm: "gélule",
				PackSize:           "caps 30",
				SourcePage:         &start,
			},
		}, nil
	})
	t.Cleanup(handlers.TestClearCompendiumExtract)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/extract", adminTok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("extract %d %#v", code, env)
	}

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
		t.Fatalf("status %s err=%s", detail.Job.Status, detail.Job.ErrorMessage)
	}
	if detail.RefCatalogCount < 1 {
		t.Fatalf("refCatalogCount=%d", detail.RefCatalogCount)
	}
	if len(detail.Rows) != 1 {
		t.Fatalf("rows %#v", detail.Rows)
	}
	row := detail.Rows[0]
	if row.SuggestedCNK != "2712345" {
		t.Fatalf("suggested %#v", row)
	}
	if row.Status == "pending" && row.CNK == "" && row.ErrorCode == "cnk_unmatched" {
		code, env = doAuthJSON(t, api.handler, http.MethodPatch,
			"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+row.ID, adminTok,
			map[string]any{"cnk": row.SuggestedCNK})
		if code != http.StatusOK {
			t.Fatalf("accept suggest %d %#v", code, env)
		}
	} else if row.CNK != "2712345" || row.Status != "pending" {
		t.Fatalf("autofill expect pending+cnk %#v", row)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/confirm-ready", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("confirm %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/compendium-imports/"+jobID+"/commit", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("commit %d %#v", code, env)
	}
}

func TestCompendiumImportACL(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	code, env := doCompendiumUpload(t, api.handler, vetTok, pdf, 1, 2)
	if code != http.StatusForbidden || errCode(env) != "forbidden" {
		t.Fatalf("vet upload want 403 forbidden got %d %#v", code, env)
	}
}

func TestCompendiumImportDelete(t *testing.T) {
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

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")

	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/compendium-imports/"+jobID, vetTok, nil)
	if code != http.StatusForbidden || errCode(env) != "forbidden" {
		t.Fatalf("vet delete want 403 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/compendium-imports/"+jobID, adminTok, nil)
	if code != http.StatusNoContent {
		t.Fatalf("delete %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/compendium-imports/"+jobID, adminTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("get after delete want 404 got %d %#v", code, env)
	}

	// Conflict while extracting
	code, env = doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 2)
	if code != http.StatusCreated {
		t.Fatalf("upload2 %d %#v", code, env)
	}
	jobID2, _ := env["data"].(map[string]any)["id"].(string)
	st := store.New(api.pool)
	if err := st.MarkCompendiumExtracting(context.Background(), jobID2, 1); err != nil {
		t.Fatalf("mark extracting: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/admin/compendium-imports/"+jobID2, adminTok, nil)
	if code != http.StatusConflict || errCode(env) != "conflict" {
		t.Fatalf("delete extracting want 409 conflict got %d %#v", code, env)
	}
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
