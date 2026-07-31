package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/handlers"
)

// uniqueOrthancID returns a 40-hex Orthanc-style id (unique per call — avoids DB residue across runs).
func uniqueOrthancID() string {
	return strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", "")[:40]
}

// minimalDicomPayload is a tiny buffer with DICOM preamble + "DICM" magic at offset 128.
func minimalDicomPayload() []byte {
	buf := make([]byte, 132)
	copy(buf[128:], []byte("DICM"))
	return buf
}

func doAuthPacsUpload(t *testing.T, h http.Handler, petID, token, filename string, content []byte) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets/"+petID+"/pacs/studies", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept-Language", "fr")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}

func TestPacsDisabledReturns404(t *testing.T) {
	t.Setenv("PACS_ENABLED", "false")
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/status", vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 pacs_disabled, got %d %#v", code, env)
	}
	if errCode(env) != "pacs_disabled" {
		t.Fatalf("err code %#v", env)
	}
}

func TestPacsStatusWakeGatesAndAdmin(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/system":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, _ := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/status", clientTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("client status expected 403 got %d", code)
	}
	code, _ = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pacs/wake", clientTok, map[string]any{})
	if code != http.StatusForbidden {
		t.Fatalf("client wake expected 403 got %d", code)
	}

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/status", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("status %d %#v", code, env)
	}
	if dataMap(t, env)["state"] != "ready" {
		t.Fatalf("expected ready %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pacs/wake", vetTok, map[string]any{})
	if code != http.StatusAccepted {
		t.Fatalf("wake %d %#v", code, env)
	}

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/pacs/logs", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin logs %d %#v", code, env)
	}
	if logs, ok := env["data"].([]any); !ok || len(logs) == 0 {
		t.Fatalf("expected logs %#v", env)
	}

	code, _ = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/pacs/logs", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet admin logs expected 403 got %d", code)
	}
}

func TestPacsStudyProxyTenantGate(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	unlinked := uniqueOrthancID()
	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.URL.Path == "/studies/"+unlinked:
			_, _ = w.Write([]byte(`{"ID":"` + unlinked + `"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	// Unknown / unlinked Orthanc study → 404 (no IDOR)
	code, env := doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/"+unlinked, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("unlinked study expected 404 got %d %#v", code, env)
	}

	// Path traversal / junk IDs: chi may 404 before handler; handler returns 400 when matched.
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/not-a-valid-orthanc-id!!!", vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("bad id expected 400 got %d %#v", code, env)
	}

	// Too short / only hyphens → 400 (regex resserrée).
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/--------", vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("hyphens-only expected 400 got %d %#v", code, env)
	}

	// Orthanc dashed SHA-1 (5×8) — format valide, étude non liée → 404 (pas 400).
	dashed := uniqueOrthancID()[:8] + "-" + uniqueOrthancID()[:8] + "-" + uniqueOrthancID()[:8] + "-" + uniqueOrthancID()[:8] + "-" + uniqueOrthancID()[:8]
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/"+dashed, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("dashed unlinked expected 404 got %d %#v", code, env)
	}
}

func TestPacsUploadMagicAndHappyPath(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()
	var deletedMu sync.Mutex
	var deletedStudy string

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			if len(body) < 132 || string(body[128:132]) != "DICM" {
				http.Error(w, "bad dicom", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{"StudyInstanceUID":"1.2.3.4","StudyDescription":"RX thorax","ModalitiesInStudy":"DX"}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/studies/"):
			deletedMu.Lock()
			deletedStudy = strings.TrimPrefix(r.URL.Path, "/studies/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	// Reject short / non-DICOM payload (magic DICM missing).
	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "fake.dcm", []byte("not-a-dicom"))
	if code != http.StatusBadRequest {
		t.Fatalf("bad magic expected 400 got %d %#v", code, env)
	}
	if errCode(env) != "validation_error" {
		t.Fatalf("err code %#v", env)
	}

	// Reject wrong extension even with magic present.
	code, env = doAuthPacsUpload(t, api.handler, petID, vetTok, "study.bin", minimalDicomPayload())
	if code != http.StatusBadRequest || errCode(env) != "validation_error" {
		t.Fatalf("bad extension expected 400 validation_error got %d %#v", code, env)
	}

	code, env = doAuthPacsUpload(t, api.handler, petID, vetTok, "study.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload happy path %d %#v", code, env)
	}
	row := dataMap(t, env)
	if row["orthancStudyId"] != studyID {
		t.Fatalf("orthancStudyId=%v want %s %#v", row["orthancStudyId"], studyID, row)
	}
	if row["petId"] != petID {
		t.Fatalf("petId=%v want %s", row["petId"], petID)
	}
	if row["modality"] != "DX" {
		t.Fatalf("modality=%v want DX", row["modality"])
	}

	// List studies for pet includes the upload.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID+"/pacs/studies", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list studies %d %#v", code, env)
	}
	list, ok := env["data"].([]any)
	if !ok || len(list) == 0 {
		t.Fatalf("expected at least one study %#v", env)
	}

	// Proxy now allowed for linked study.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/studies/"+studyID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("proxy linked study %d %#v", code, env)
	}

	deletedMu.Lock()
	defer deletedMu.Unlock()
	if deletedStudy != "" {
		t.Fatalf("unexpected compensatory delete of study %q", deletedStudy)
	}
}

func TestPacsUploadCompensatesOrthancOnDBFail(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()
	var deletedMu sync.Mutex
	var deletedInstance string
	var deletedStudy string

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/instances/"):
			deletedMu.Lock()
			deletedInstance = strings.TrimPrefix(r.URL.Path, "/instances/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/studies/"):
			deletedMu.Lock()
			deletedStudy = strings.TrimPrefix(r.URL.Path, "/studies/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	handlers.TestArmFailNextPetStudyInsert(api.api)
	t.Cleanup(func() { handlers.TestClearFailNextPetStudyInsert(api.api) })

	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "study.dcm", minimalDicomPayload())
	if code != http.StatusInternalServerError {
		t.Fatalf("forced db fail expected 500 got %d %#v", code, env)
	}

	deletedMu.Lock()
	gotInst, gotStudy := deletedInstance, deletedStudy
	deletedMu.Unlock()
	if gotInst != instID {
		t.Fatalf("expected compensatory delete of instance %s got %q (study=%q)", instID, gotInst, gotStudy)
	}
	if gotStudy != "" {
		t.Fatalf("must not delete ParentStudy when instance id is known, got study %q", gotStudy)
	}
}

func TestPacsSeriesAuthzViaParentStudy(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesLinked := uniqueOrthancID()
	seriesSibling := uniqueOrthancID()
	instID := uniqueOrthancID()

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesLinked + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{"StudyInstanceUID":"1.2.3"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/series/"+seriesSibling:
			_, _ = w.Write([]byte(`{"ID":"` + seriesSibling + `","ParentStudy":"` + studyID + `"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/series/"+seriesLinked:
			_, _ = w.Write([]byte(`{"ID":"` + seriesLinked + `","ParentStudy":"` + studyID + `"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "a.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}

	// Sibling series (not stored as orthanc_series_id) must still pass via ParentStudy.
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/series/"+seriesSibling, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("sibling series expected 200 got %d %#v", code, env)
	}
}

func TestPacsCreatePetStudyRejectsOtherPetConflict(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()
	var deletedMu sync.Mutex
	var deletedInstance string

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/instances/"):
			deletedMu.Lock()
			deletedInstance = strings.TrimPrefix(r.URL.Path, "/instances/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petA := activeDemoPetID(t, api.handler, clientTok)
	petB := activeDemoPetID(t, api.handler, clientTok)
	code, env := doAuthPacsUpload(t, api.handler, petA, vetTok, "a.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("first upload %d %#v", code, env)
	}
	code, env = doAuthPacsUpload(t, api.handler, petB, vetTok, "b.dcm", minimalDicomPayload())
	if code != http.StatusConflict {
		t.Fatalf("second pet same study expected 409 got %d %#v", code, env)
	}
	if errMsgKey(env) != "pacs_study_other_pet" {
		t.Fatalf("msgKey want pacs_study_other_pet got %q %#v", errMsgKey(env), env)
	}
	deletedMu.Lock()
	got := deletedInstance
	deletedMu.Unlock()
	if got != instID {
		t.Fatalf("conflict must compensate instance delete, got %q", got)
	}
}

// doAuthBytes hits an authenticated endpoint and returns raw body + headers (preview/file).
func doAuthBytes(t *testing.T, h http.Handler, method, path, token string) (int, []byte, http.Header) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Accept-Language", "fr")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes(), rec.Result().Header
}

// pacsOrthancFixture is a realistic Orthanc mock: GET /instances omits ParentStudy
// (production Orthanc 1.12 behaviour) and only exposes ParentSeries.
type pacsOrthancFixture struct {
	studyID, seriesID, instID string
	previewPNG, dicomFile     []byte
}

func newPacsOrthancFixture() pacsOrthancFixture {
	return pacsOrthancFixture{
		studyID:    uniqueOrthancID(),
		seriesID:   uniqueOrthancID(),
		instID:     uniqueOrthancID(),
		previewPNG: []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 'f', 'a', 'k', 'e'},
		dicomFile:  minimalDicomPayload(),
	}
}

func (f pacsOrthancFixture) handler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		switch {
		case r.URL.Path == "/system":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ID":"` + f.instID + `","ParentStudy":"` + f.studyID + `","ParentSeries":"` + f.seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+f.studyID:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ID":"` + f.studyID + `","Series":["` + f.seriesID + `"],"MainDicomTags":{"StudyInstanceUID":"1.2.826.demo","StudyDescription":"CT_Chest","ModalitiesInStudy":"CT"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/series/"+f.seriesID:
			w.Header().Set("Content-Type", "application/json")
			// Series keeps ParentStudy (always present on real Orthanc).
			_, _ = w.Write([]byte(`{"ID":"` + f.seriesID + `","ParentStudy":"` + f.studyID + `","Instances":["` + f.instID + `"]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+f.instID:
			w.Header().Set("Content-Type", "application/json")
			// No ParentStudy — reproduces staging black-screen root cause.
			_, _ = w.Write([]byte(`{"ID":"` + f.instID + `","ParentSeries":"` + f.seriesID + `","Type":"Instance"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+f.instID+"/frames/0/preview":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(f.previewPNG)
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+f.instID+"/file":
			w.Header().Set("Content-Type", "application/dicom")
			_, _ = w.Write(f.dicomFile)
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+f.instID+"/simplified-tags":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Modality":"CT","PixelSpacing":"0.5\\0.5","WindowCenter":"40","WindowWidth":"400","NumberOfFrames":"1","Rows":"64","Columns":"64"}`))
		default:
			http.NotFound(w, r)
		}
	})
}

func TestPacsInstancePreviewFileFullFlow(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)
	fx := newPacsOrthancFixture()
	orth := httptest.NewServer(fx.handler(t))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "chest.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}

	// Study + series proxy (viewer open path).
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/studies/"+fx.studyID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("study proxy %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pacs/series/"+fx.seriesID, vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("series proxy %d %#v", code, env)
	}
	series := dataMap(t, env)
	insts, _ := series["Instances"].([]any)
	if len(insts) != 1 || insts[0] != fx.instID {
		t.Fatalf("series Instances %#v want [%s]", series["Instances"], fx.instID)
	}

	// Preview — requires ParentSeries → ParentStudy fallback (no ParentStudy on instance).
	code, body, hdr := doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+fx.instID+"/frames/0/preview", vetTok)
	if code != http.StatusOK {
		t.Fatalf("preview expected 200 got %d body=%s", code, string(body))
	}
	if !bytes.Equal(body, fx.previewPNG) {
		t.Fatalf("preview bytes mismatch len=%d", len(body))
	}
	if ct := hdr.Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("preview content-type %q", ct)
	}
	if cc := hdr.Get("Cache-Control"); !strings.Contains(cc, "no-store") || !strings.Contains(cc, "private") {
		t.Fatalf("preview Cache-Control must be private,no-store got %q", cc)
	}

	// DICOM file download (same authz path).
	code, body, hdr = doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+fx.instID+"/file", vetTok)
	if code != http.StatusOK {
		t.Fatalf("file expected 200 got %d body=%s", code, string(body))
	}
	if !bytes.Equal(body, fx.dicomFile) {
		t.Fatalf("file bytes mismatch")
	}
	if cc := hdr.Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Fatalf("file Cache-Control %q", cc)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+fx.instID+"/metadata", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("metadata %d %#v", code, env)
	}
	md := dataMap(t, env)
	if md["modality"] != "CT" {
		t.Fatalf("modality %#v", md)
	}
	sp, _ := md["pixelSpacingMm"].([]any)
	if len(sp) != 2 {
		t.Fatalf("pixelSpacingMm %#v", md["pixelSpacingMm"])
	}

	// Client must never reach PHI binary endpoints.
	code, _, _ = doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+fx.instID+"/frames/0/preview", clientTok)
	if code != http.StatusForbidden {
		t.Fatalf("client preview expected 403 got %d", code)
	}
	code, _, _ = doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+fx.instID+"/file", clientTok)
	if code != http.StatusForbidden {
		t.Fatalf("client file expected 403 got %d", code)
	}
}

func TestPacsInstancePreviewOutOfRangeIs404(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")

	for _, orthStatus := range []int{http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError} {
		orthStatus := orthStatus
		t.Run(fmt.Sprintf("orthanc_%d", orthStatus), func(t *testing.T) {
			api := newTestAPI(t)
			fx := newPacsOrthancFixture()
			orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/instances/"+fx.instID+"/frames/9/preview" {
					http.Error(w, "out of range", orthStatus)
					return
				}
				fx.handler(t).ServeHTTP(w, r)
			}))
			t.Cleanup(orth.Close)
			handlers.TestSetOrthanc(api.api, orth.URL)

			clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
			vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
			petID := activeDemoPetID(t, api.handler, clientTok)
			code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "chest.dcm", minimalDicomPayload())
			if code != http.StatusCreated {
				t.Fatalf("upload %d %#v", code, env)
			}

			code, _, _ = doAuthBytes(t, api.handler, http.MethodGet,
				"/api/v1/pacs/instances/"+fx.instID+"/frames/9/preview", vetTok)
			if code != http.StatusNotFound {
				t.Fatalf("out-of-range preview want 404 got %d (orthanc %d)", code, orthStatus)
			}
		})
	}
}

func TestPacsInstanceTenantIsolationNoLeak(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)
	fx := newPacsOrthancFixture()
	orth := httptest.NewServer(fx.handler(t))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetPlusTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	vetParcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthPacsUpload(t, api.handler, petID, vetPlusTok, "chest.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}

	paths := []string{
		"/api/v1/pacs/studies/" + fx.studyID,
		"/api/v1/pacs/series/" + fx.seriesID,
		"/api/v1/pacs/instances/" + fx.instID + "/frames/0/preview",
		"/api/v1/pacs/instances/" + fx.instID + "/file",
	}
	for _, p := range paths {
		code, body, _ := doAuthBytes(t, api.handler, http.MethodGet, p, vetParcTok)
		if code != http.StatusNotFound {
			t.Fatalf("cross-practice %s expected 404 got %d body=%s", p, code, string(body))
		}
		// Opaque not_found — no Orthanc payload leak in error body.
		var envelope map[string]any
		_ = json.Unmarshal(body, &envelope)
		if errObj, _ := envelope["error"].(map[string]any); errObj != nil {
			if errObj["code"] != "not_found" {
				t.Fatalf("cross-practice %s err code %#v", p, errObj)
			}
		}
	}

	// Unlinked Orthanc instance (exists in mock but no imaging.pet_studies row) → 404.
	orphanInst := uniqueOrthancID()
	orphanSeries := uniqueOrthancID()
	orphanStudy := uniqueOrthancID()
	orth2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/instances/"+orphanInst:
			_, _ = w.Write([]byte(`{"ID":"` + orphanInst + `","ParentSeries":"` + orphanSeries + `"}`))
		case r.URL.Path == "/series/"+orphanSeries:
			_, _ = w.Write([]byte(`{"ID":"` + orphanSeries + `","ParentStudy":"` + orphanStudy + `"}`))
		case strings.HasPrefix(r.URL.Path, "/instances/"+orphanInst+"/"):
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("should-not-leak"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth2.Close)
	handlers.TestSetOrthanc(api.api, orth2.URL)

	code, body, _ := doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+orphanInst+"/frames/0/preview", vetPlusTok)
	if code != http.StatusNotFound {
		t.Fatalf("orphan instance expected 404 got %d body=%s", code, string(body))
	}
	if bytes.Contains(body, []byte("should-not-leak")) {
		t.Fatal("orphan preview must not stream Orthanc bytes")
	}
}

func TestPacsInstanceAuthzRequiresParentChain(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()
	brokenInst := uniqueOrthancID()

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+brokenInst:
			// Neither ParentStudy nor ParentSeries → authz must fail closed.
			_, _ = w.Write([]byte(`{"ID":"` + brokenInst + `","Type":"Instance"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/instances/"+instID:
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `"}`))
		case strings.HasPrefix(r.URL.Path, "/instances/"+instID+"/frames/"):
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "a.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload %d %#v", code, env)
	}

	// Direct ParentStudy on instance still works (compat older Orthanc / mocks).
	code, body, _ := doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+instID+"/frames/0/preview", vetTok)
	if code != http.StatusOK {
		t.Fatalf("direct ParentStudy preview %d %s", code, string(body))
	}

	code, _, _ = doAuthBytes(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/"+brokenInst+"/frames/0/preview", vetTok)
	if code != http.StatusNotFound {
		t.Fatalf("broken parent chain expected 404 got %d", code)
	}

	// Junk instance id → 400 (no Orthanc round-trip leak vector).
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/instances/not-valid!!!/frames/0/preview", vetTok, nil)
	if code != http.StatusBadRequest {
		t.Fatalf("bad instance id expected 400 got %d %#v", code, env)
	}
}

func TestPacsUploadResolvesParentViaSeries(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			// No ParentStudy on upload (Orthanc 1.12 asymmetry).
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/series/"+seriesID:
			_, _ = w.Write([]byte(`{"ID":"` + seriesID + `","ParentStudy":"` + studyID + `"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{"StudyDescription":"via-series","ModalitiesInStudy":"CT"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthPacsUpload(t, api.handler, petID, vetTok, "chest.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("upload via ParentSeries %d %#v", code, env)
	}
	row := dataMap(t, env)
	if row["orthancStudyId"] != studyID {
		t.Fatalf("orthancStudyId=%v want %s", row["orthancStudyId"], studyID)
	}
	if row["orthancSeriesId"] != seriesID {
		t.Fatalf("orthancSeriesId=%v want %s", row["orthancSeriesId"], seriesID)
	}
}

func TestPacsUploadRejectsOtherPractice(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	studyID := uniqueOrthancID()
	seriesID := uniqueOrthancID()
	instID := uniqueOrthancID()
	var deletedMu sync.Mutex
	var deletedInstance string

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/instances":
			_, _ = w.Write([]byte(`{"ID":"` + instID + `","ParentStudy":"` + studyID + `","ParentSeries":"` + seriesID + `","Status":"Success"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/studies/"+studyID:
			_, _ = w.Write([]byte(`{"ID":"` + studyID + `","MainDicomTags":{}}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/instances/"):
			deletedMu.Lock()
			deletedInstance = strings.TrimPrefix(r.URL.Path, "/instances/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	vetPlusTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	clientDemoTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petPlus := activeDemoPetID(t, api.handler, clientDemoTok)

	code, env := doAuthPacsUpload(t, api.handler, petPlus, vetPlusTok, "a.dcm", minimalDicomPayload())
	if code != http.StatusCreated {
		t.Fatalf("vetplus upload %d %#v", code, env)
	}

	vetParcTok := loginToken(t, api.handler, "vet.parc@petsfollow.test", "VetDemo123!")
	marieTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	petParc := activeDemoPetID(t, api.handler, marieTok)

	deletedMu.Lock()
	deletedInstance = ""
	deletedMu.Unlock()

	code, env = doAuthPacsUpload(t, api.handler, petParc, vetParcTok, "b.dcm", minimalDicomPayload())
	if code != http.StatusConflict {
		t.Fatalf("cross-practice expected 409 got %d %#v", code, env)
	}
	if errMsgKey(env) != "pacs_study_other_practice" {
		t.Fatalf("msgKey want pacs_study_other_practice got %q %#v", errMsgKey(env), env)
	}
	deletedMu.Lock()
	got := deletedInstance
	deletedMu.Unlock()
	if got != instID {
		t.Fatalf("must compensate instance delete, got %q", got)
	}
}

func TestPacsPurgeOrthancStudiesDeletesMapped(t *testing.T) {
	// Purge Orthanc must work even when the product flag is off (RGPD).
	t.Setenv("PACS_ENABLED", "false")
	api := newTestAPI(t)

	orthancStudy := uniqueOrthancID()
	var deletedMu sync.Mutex
	var deletedStudy string
	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/studies/") {
			deletedMu.Lock()
			deletedStudy = strings.TrimPrefix(r.URL.Path, "/studies/")
			deletedMu.Unlock()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	handlers.TestPurgeOrthancStudies(api.api, []string{orthancStudy})

	deletedMu.Lock()
	got := deletedStudy
	deletedMu.Unlock()
	if got != orthancStudy {
		t.Fatalf("Orthanc DELETE study want %q got %q", orthancStudy, got)
	}
}

func TestPacsAdminWake(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)
	handlers.TestSetOrthanc(api.api, "http://127.0.0.1:1")
	tok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/pacs/wake", tok, nil)
	if code != http.StatusAccepted {
		t.Fatalf("wake status=%d %#v", code, env)
	}
	row := dataMap(t, env)
	if row["state"] != "starting" {
		t.Fatalf("wake state want starting got %#v", row)
	}
}
