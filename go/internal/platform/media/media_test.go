package media

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalUploadAndServe(t *testing.T) {
	root := t.TempDir()
	st, handler, err := newLocal(root, "http://localhost:8291")
	if err != nil {
		t.Fatal(err)
	}
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	}
	url, err := st.Upload(nil, ObjectKey("avatars", "user-1", ".png"), bytes.NewReader(png), int64(len(png)), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "http://localhost:8291/media/avatars/user-1/") {
		t.Fatalf("unexpected url: %s", url)
	}
	key := strings.TrimPrefix(url, "http://localhost:8291/media/")
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(key))); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/media/"+key, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("serve status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !bytes.Equal(body, png) {
		t.Fatal("body mismatch")
	}
	if err := st.Delete(nil, key); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(key))); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, got %v", err)
	}
}

func TestNormalizeContentType(t *testing.T) {
	ct, err := NormalizeContentType("application/octet-stream", "photo.JPG")
	if err != nil || ct != "image/jpeg" {
		t.Fatalf("got %q %v", ct, err)
	}
	if _, err := NormalizeContentType("text/plain", "a.txt"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeMessageMediaType(t *testing.T) {
	ct, err := NormalizeMessageMediaType("application/octet-stream", "clip.mp4")
	if err != nil || ct != "video/mp4" {
		t.Fatalf("got %q %v", ct, err)
	}
	if MediaKind(ct) != "video" {
		t.Fatalf("expected video kind")
	}
	if _, err := NormalizeMessageMediaType("text/plain", "a.txt"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDenySensitivePrefixes(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := DenySensitivePrefixes(okHandler, "visit-reports/")
	for _, path := range []string{
		"/media/visit-reports/v1/a.m4a",
		"/media/visit-reports",
		"/media/visit-reports/",
		"/media/./visit-reports/v1/a.m4a",
		"/media/foo/../visit-reports/v1/a.m4a",
		"/media/Visit-Reports/v1/a.m4a",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s: want 403 got %d", path, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/media/avatars/u1/x.png", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("avatar should pass, got %d", rec.Code)
	}
}

func TestIsSensitiveObjectKey(t *testing.T) {
	yes := []string{
		"visit-reports",
		"visit-reports/",
		"visit-reports/v1/a.m4a",
		"Visit-Reports/x",
		"/visit-reports/x",
		"./visit-reports/x",
	}
	for _, k := range yes {
		if !IsSensitiveObjectKey(k) {
			t.Fatalf("expected sensitive: %q", k)
		}
	}
	no := []string{"avatars/u1.png", "pets/p1.jpg", "visit-report/x", ""}
	for _, k := range no {
		if IsSensitiveObjectKey(k) {
			t.Fatalf("expected not sensitive: %q", k)
		}
	}
}

func TestSensitiveUploadNoPublicURL(t *testing.T) {
	root := t.TempDir()
	st, _, err := newLocal(root, "http://localhost:8291")
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("audio-bytes")
	url, err := st.Upload(nil, "visit-reports/v1/clip.m4a", bytes.NewReader(data), int64(len(data)), "audio/mp4")
	if err != nil {
		t.Fatal(err)
	}
	if url != "" {
		t.Fatalf("expected empty public URL for PHI, got %q", url)
	}
	if !IsSensitiveObjectKey("visit-reports/v1/clip.m4a") {
		t.Fatal("expected sensitive")
	}
}
