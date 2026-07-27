package handlers_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
)

func solidPNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 80, B: 160, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestPetIdentificationAndHealthBookPDF(t *testing.T) {
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
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get pet %d %#v", code, env)
	}
	before := dataMap(t, env)
	prevChip, _ := before["microchipNumber"].(string)
	prevBook, _ := before["healthBookNumber"].(string)

	t.Cleanup(func() {
		_, _ = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/pets/"+petID, clientTok, map[string]any{
			"name":             before["name"],
			"species":          before["species"],
			"breed":            before["breed"],
			"microchipNumber":  prevChip,
			"healthBookNumber": prevBook,
		})
		_, _ = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/"+petID+"/health-book", clientTok, nil)
	})

	chip := "250269609998877"
	book := "CARNET-E2E-1"
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/pets/"+petID, clientTok, map[string]any{
		"name":             before["name"],
		"species":          before["species"],
		"breed":            before["breed"],
		"microchipNumber":  chip,
		"healthBookNumber": book,
	})
	if code != http.StatusOK {
		t.Fatalf("update pet %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get pet after update %d %#v", code, env)
	}
	pet := dataMap(t, env)
	if pet["microchipNumber"] != chip {
		t.Fatalf("microchip=%v want %s", pet["microchipNumber"], chip)
	}
	if pet["healthBookNumber"] != book {
		t.Fatalf("healthBook=%v want %s", pet["healthBookNumber"], book)
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for _, raw := range [][]byte{solidPNGBytes(t, 120, 80), solidPNGBytes(t, 90, 140)} {
		part, err := w.CreateFormFile("files", "page.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets/"+petID+"/health-book", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+clientTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if rec.Code != http.StatusOK {
		t.Fatalf("health-book upload %d %#v", rec.Code, envelope)
	}
	uploaded := dataMap(t, envelope)
	if attached, _ := uploaded["healthBookPdfAttached"].(bool); !attached {
		t.Fatalf("expected healthBookPdfAttached=true: %#v", uploaded)
	}
	if u, _ := uploaded["healthBookPdfUrl"].(string); u != "" {
		t.Fatalf("PHI pdf must not expose public URL, got %q", u)
	}

	// Authenticated stream (owner + vet with access).
	for _, tok := range []string{clientTok, vetTok} {
		req = httptest.NewRequest(http.MethodGet, "/api/v1/pets/"+petID+"/health-book", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec = httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("stream health-book %d body=%s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "pdf") {
			t.Fatalf("content-type=%q", ct)
		}
		pdf, _ := io.ReadAll(rec.Body)
		if !bytes.HasPrefix(pdf, []byte("%PDF")) {
			t.Fatalf("expected PDF body")
		}
	}

	// Unauthenticated must fail.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/pets/"+petID+"/health-book", nil)
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("unauthenticated stream must not succeed")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodDelete, "/api/v1/pets/"+petID+"/health-book", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("delete health-book %d %#v", code, env)
	}
	cleared := dataMap(t, env)
	if attached, _ := cleared["healthBookPdfAttached"].(bool); attached {
		t.Fatalf("attached should be false after delete: %#v", cleared)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/pets/"+petID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get after clear %d %#v", code, env)
	}
	pet = dataMap(t, env)
	if pet["microchipNumber"] != chip || pet["healthBookNumber"] != book {
		t.Fatalf("numbers lost after pdf clear: %#v", pet)
	}
}
