package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/handlers"
)

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

	orth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/system":
			_, _ = w.Write([]byte(`{"Version":"1.12.7"}`))
		case r.URL.Path == "/studies/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa":
			_, _ = w.Write([]byte(`{"ID":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(orth.Close)
	handlers.TestSetOrthanc(api.api, orth.URL)

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	// Unknown / unlinked Orthanc study → 404 (no IDOR)
	code, env := doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", vetTok, nil)
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
	// ID volontairement hors fixture démo (évite collision seed pacs-demo-seed).
	dashed := "aaaaaaaa-bbbbbbbb-cccccccc-dddddddd-eeeeeeee"
	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/pacs/studies/"+dashed, vetTok, nil)
	if code != http.StatusNotFound {
		t.Fatalf("dashed unlinked expected 404 got %d %#v", code, env)
	}
}

func TestPacsUploadMagicAndHappyPath(t *testing.T) {
	t.Setenv("PACS_ENABLED", "true")
	api := newTestAPI(t)

	const (
		studyID  = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		seriesID = "cccccccccccccccccccccccccccccccccccccccc"
		instID   = "dddddddddddddddddddddddddddddddddddddddd"
	)
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

	const (
		studyID  = "1111111111111111111111111111111111111111"
		seriesID = "2222222222222222222222222222222222222222"
		instID   = "3333333333333333333333333333333333333333"
	)
	var deletedMu sync.Mutex
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
	got := deletedStudy
	deletedMu.Unlock()
	if got != studyID {
		t.Fatalf("expected compensatory delete of %s got %q", studyID, got)
	}
}
