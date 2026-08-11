package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	if detail.Job.ExtractTotal != 1 {
		t.Fatalf("extractTotal=%d want 1 for 2-page range", detail.Job.ExtractTotal)
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

func TestCompendiumImportMultiChunk(t *testing.T) {
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

	prev := pharmacy.CompendiumPagesPerChunk
	pharmacy.CompendiumPagesPerChunk = 3
	t.Cleanup(func() { pharmacy.CompendiumPagesPerChunk = prev })

	var calls [][2]int
	handlers.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		calls = append(calls, [2]int{start, end})
		p := start
		return []pharmacy.ExtractedMedication{
			{CNK: fmt.Sprintf("2889%03d", start), Name: fmt.Sprintf("Chunk Med %d", start), SourcePage: &p},
		}, nil
	})
	t.Cleanup(handlers.TestClearCompendiumExtract)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	// pages 1–8 with chunk=3 → 3 chunks: 1-3, 4-6, 7-8
	code, env := doCompendiumUpload(t, api.handler, adminTok, pdf, 1, 8)
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}
	jobID, _ := env["data"].(map[string]any)["id"].(string)

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
		t.Fatalf("status %s err=%s calls=%v", detail.Job.Status, detail.Job.ErrorMessage, calls)
	}
	if detail.Job.ExtractTotal != 3 || detail.Job.ExtractDone != 3 {
		t.Fatalf("progress done=%d total=%d", detail.Job.ExtractDone, detail.Job.ExtractTotal)
	}
	if len(calls) != 3 || calls[0] != [2]int{1, 3} || calls[1] != [2]int{4, 6} || calls[2] != [2]int{7, 8} {
		t.Fatalf("calls %#v", calls)
	}
	if len(detail.Rows) != 3 {
		t.Fatalf("rows=%d %#v", len(detail.Rows), detail.Rows)
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

func TestCompendiumLookupCNK(t *testing.T) {
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
	cnk := "2734567"
	cleanup := func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM pharmacy.ref_medications WHERE cnk = $1 OR name LIKE 'ZZZ Lookup Med %'`, cnk)
	}
	cleanup()
	t.Cleanup(cleanup)

	if _, err := st.UpsertRefMedication(context.Background(), store.RefMedicationUpsert{
		CNK:                cnk,
		Name:               "ZZZ Lookup Med Capsules 10 mg",
		PharmaceuticalForm: "gélule",
		IsActive:           true,
	}); err != nil {
		t.Fatalf("seed ref: %v", err)
	}

	handlers.TestSetCompendiumExtract(func(ctx context.Context, pdf []byte, start, end int) ([]pharmacy.ExtractedMedication, error) {
		return []pharmacy.ExtractedMedication{
			{Name: "ZZZ Lookup Med Caps 10mg", PharmaceuticalForm: "gélule", SourcePage: &start},
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
	if detail.Job.Status != "extracted" || len(detail.Rows) != 1 {
		t.Fatalf("status=%s rows=%d err=%s", detail.Job.Status, len(detail.Rows), detail.Job.ErrorMessage)
	}
	rowID := detail.Rows[0].ID

	// Name rematch (no CNK in body)
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID+"/lookup-cnk", adminTok, map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("lookup name %d %#v", code, env)
	}
	data, _ := env["data"].(map[string]any)
	rowOut, _ := data["row"].(map[string]any)
	if sug, _ := rowOut["suggestedCnk"].(string); sug != cnk {
		t.Fatalf("suggested after name lookup %#v", rowOut)
	}
	if data["cnkFound"] != nil {
		t.Fatalf("cnkFound should be omitted on name-only lookup %#v", data)
	}

	// Exact CNK verify (found) — row may already have autofilled CNK from rematch
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID+"/lookup-cnk", adminTok,
		map[string]any{"cnk": cnk})
	if code != http.StatusOK {
		t.Fatalf("lookup exact %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	if data["cnkFound"] != true {
		t.Fatalf("cnkFound want true %#v", data)
	}
	rowOut, _ = data["row"].(map[string]any)
	if rowOut["suggestedCnk"] != cnk {
		t.Fatalf("suggestedCnk want %s %#v", cnk, rowOut)
	}
	if st, _ := rowOut["status"].(string); st != "pending" && st != "error" {
		t.Fatalf("status after lookup want pending/error got %#v", rowOut)
	}

	// Exact CNK verify (missing) → still name suggestions; do not wipe existing CNK
	before, err := st.GetCompendiumImportRow(context.Background(), jobID, rowID)
	if err != nil {
		t.Fatalf("get before miss: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID+"/lookup-cnk", adminTok,
		map[string]any{"cnk": "2999999"})
	if code != http.StatusOK {
		t.Fatalf("lookup miss %d %#v", code, env)
	}
	data, _ = env["data"].(map[string]any)
	if data["cnkFound"] != false {
		t.Fatalf("cnkFound want false %#v", data)
	}
	rowOut, _ = data["row"].(map[string]any)
	if before.CNK != "" && rowOut["cnk"] != before.CNK {
		t.Fatalf("miss must not overwrite existing cnk before=%q after=%#v", before.CNK, rowOut)
	}

	// Promote to ready then lookup → 409 (no demote)
	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID, adminTok,
		map[string]any{"cnk": cnk, "name": "ZZZ Lookup Med Caps 10mg"})
	if code != http.StatusOK {
		t.Fatalf("patch ready %d %#v", code, env)
	}
	if got, _ := env["data"].(map[string]any)["status"].(string); got != "ready" {
		t.Fatalf("want ready got %#v", env["data"])
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID+"/lookup-cnk", adminTok,
		map[string]any{"cnk": cnk})
	if code != http.StatusConflict || errCode(env) != "conflict" {
		t.Fatalf("lookup ready want 409 conflict got %d %#v", code, env)
	}
	detail, _ = st.GetCompendiumImportDetail(context.Background(), jobID)
	if detail.Rows[0].Status != "ready" {
		t.Fatalf("ready must stay ready after rejected lookup %#v", detail.Rows[0])
	}

	// Excluded → 409
	code, env = doAuthJSON(t, api.handler, http.MethodPatch,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID, adminTok,
		map[string]any{"excluded": true})
	if code != http.StatusOK {
		t.Fatalf("exclude %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/admin/compendium-imports/"+jobID+"/rows/"+rowID+"/lookup-cnk", adminTok, map[string]any{})
	if code != http.StatusConflict || errCode(env) != "conflict" {
		t.Fatalf("lookup excluded want 409 got %d %#v", code, env)
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
