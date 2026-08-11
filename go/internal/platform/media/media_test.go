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
	h := DenySensitivePrefixes(okHandler)
	for _, path := range []string{
		"/media/visit-reports/v1/a.m4a",
		"/media/visit-reports",
		"/media/visit-reports/",
		"/media/./visit-reports/v1/a.m4a",
		"/media/foo/../visit-reports/v1/a.m4a",
		"/media/Visit-Reports/v1/a.m4a",
		"/media/documents/p1/a.pdf",
		"/media/consultation-shares-v2/t1/report.pdf",
		"/media/",
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
		"health-books",
		"health-books/",
		"health-books/p1/a.pdf",
		"Health-Books/x",
		"dossier-shares",
		"dossier-shares/",
		"dossier-shares/t1/pack.zip",
		"Dossier-Shares/x",
		"consultation-shares",
		"consultation-shares/",
		"consultation-shares/t1/report.pdf",
		"daf",
		"daf/",
		"daf/p1/d1.pdf",
		"prescriptions",
		"prescriptions/",
		"prescriptions/p1/r1.pdf",
		"compendium-imports",
		"compendium-imports/",
		"compendium-imports/job1.pdf",
		"Compendium-Imports/x",
		"afmps-imports",
		"afmps-imports/",
		"afmps-imports/latest.csv",
		"Afmps-Imports/x",
		// Base documentaire RAG (guides) — privée, stream auth uniquement.
		"rag-docs",
		"rag-docs/",
		"rag-docs/platform/d1.pdf",
		"rag-docs/practice-id/d1.txt",
		"Rag-Docs/x",
		// Documents du dossier animal (PDF d'analyses, radios) — PHI.
		"documents/p1/a.pdf",
		// Le suffixe -v2 ne doit pas rouvrir le partage de consultation.
		"consultation-shares-v2/t1/report.pdf",
		"pitch-sims/s1/rec.webm",
		// Fail-closed : un namespace inconnu ou vide reste privé.
		"unknown-kind/x.pdf",
		"visit-report/x",
		"",
		"/",
		".",
	}
	for _, k := range yes {
		if !IsSensitiveObjectKey(k) {
			t.Fatalf("expected sensitive: %q", k)
		}
	}
	no := []string{
		"avatars/u1.png", "pets/p1.jpg", "messages/t1/clip.mp4",
		"brand/qr_android", "Avatars/U1.PNG",
	}
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
	pdf := []byte("%PDF-1.4\n%%EOF\n")
	for _, key := range []string{
		"compendium-imports/job.pdf",
		// Namespaces refermés par le durcissement : ils avaient une URL publique.
		"documents/p1/analyse.pdf",
		"pitch-sims/s1/call.webm",
		"consultation-shares-v2/t1/report.pdf",
		"rag-docs/platform/guide.pdf",
	} {
		url, err = st.Upload(nil, key, bytes.NewReader(pdf), int64(len(pdf)), "application/pdf")
		if err != nil {
			t.Fatal(err)
		}
		if url != "" {
			t.Fatalf("expected empty public URL for %s, got %q", key, url)
		}
	}
}

// publicPrefixes est toute la frontière PHI : un objet dont la clé y correspond
// reçoit une URL `storage.googleapis.com` lisible par n'importe qui (le bucket
// porte un binding `allUsers`). Élargir cette liste expose des données sans
// autre garde-fou, donc le changement doit être explicite et relu — d'où
// l'épinglage plutôt qu'une simple vérification de présence.
func TestPublicPrefixesArePinned(t *testing.T) {
	want := []string{"avatars/", "pets/", "messages/", "brand/"}
	if len(publicPrefixes) != len(want) {
		t.Fatalf("public namespaces changed: %v (want %v) — see documentation/36-RGPD.md", publicPrefixes, want)
	}
	for i, p := range want {
		if publicPrefixes[i] != p {
			t.Fatalf("public namespaces changed: %v (want %v) — see documentation/36-RGPD.md", publicPrefixes, want)
		}
	}
}
