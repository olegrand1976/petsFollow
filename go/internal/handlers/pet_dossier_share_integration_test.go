package handlers_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

func TestPetDossierShareCreateMetaDownloadAndExpiry(t *testing.T) {
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
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": "pro.externe@petsfollow.test",
	})
	if code != http.StatusCreated {
		t.Fatalf("create share %d %#v", code, env)
	}
	data := dataMap(t, env)
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatalf("expected DEV token in response %#v", data)
	}

	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/"+token, nil)
	if code != http.StatusOK {
		t.Fatalf("meta %d %#v", code, env)
	}
	meta := dataMap(t, env)
	if meta["petName"] == "" {
		t.Fatalf("expected petName in meta %#v", meta)
	}
	if phone, _ := meta["commercialPhone"].(string); phone == "" {
		t.Fatalf("expected commercialPhone (seed or fallback) %#v", meta)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/pet-dossier/"+token+"/download", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download %d %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "zip") {
		t.Fatalf("content-type=%q", ct)
	}
	zipBytes := rec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip: %v", err)
	}
	foundPDF := false
	for _, f := range zr.File {
		if f.Name == "dossier.pdf" {
			foundPDF = true
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			head, _ := io.ReadAll(io.LimitReader(rc, 5))
			_ = rc.Close()
			if string(head) != "%PDF-" {
				t.Fatalf("dossier.pdf magic %q", head)
			}
		}
	}
	if !foundPDF {
		t.Fatal("dossier.pdf missing from zip")
	}

	// Expire token and expect 410.
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pets.dossier_share_tokens SET expires_at = $1 WHERE token = $2`,
		time.Now().UTC().Add(-time.Minute), token); err != nil {
		t.Fatalf("expire: %v", err)
	}
	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/"+token, nil)
	if code != http.StatusGone {
		t.Fatalf("expired meta want 410 got %d %#v", code, env)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/public/pet-dossier/"+token+"/download", nil)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusGone {
		t.Fatalf("expired download want 410 got %d", rec.Code)
	}
}

func TestPetDossierShareOwnerOnly(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	otherTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", otherTok, map[string]any{
		"email": "pro@example.com",
	})
	if code != http.StatusForbidden {
		t.Fatalf("non-owner want 403 got %d %#v", code, env)
	}
}

func TestUpdateMeContactPhoneCommercial(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me", tok, map[string]any{
		"contactPhone": "0499 00 11 22",
	})
	if code != http.StatusOK {
		t.Fatalf("patch phone %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["contactPhone"] != "0499 00 11 22" {
		t.Fatalf("contactPhone=%v", me["contactPhone"])
	}

	// Restore seed phone for other tests / demos.
	_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me", tok, map[string]any{
		"contactPhone": "0470 12 34 56",
	})
}
