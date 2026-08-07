package handlers

import (
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Le nom de fichier vient de l'upload, donc de l'utilisateur, et atterrit dans
// un en-tête Content-Disposition. Un guillemet ou un CRLF qui passerait
// permettrait d'injecter des en-têtes dans la réponse.
func TestDocumentDownloadNameIsHeaderSafe(t *testing.T) {
	hostile := []string{
		`"; filename="evil.html`,
		"report\r\nSet-Cookie: pf_token=stolen",
		"../../../etc/passwd",
		"rapport;charset=utf-8",
	}
	for _, name := range hostile {
		got := documentDownloadName(store.PetDocument{FileName: name, ContentType: "application/pdf"})
		if strings.ContainsAny(got, "\"\r\n;\\/") {
			t.Fatalf("%q → %q : le nom rendu doit rester inoffensif en en-tête", name, got)
		}
		if !strings.HasSuffix(got, ".pdf") {
			t.Fatalf("%q → %q : l'extension doit venir du content type stocké", name, got)
		}
	}
}

// L'extension ne doit jamais être reprise du nom d'origine : sinon un PDF
// renommé en .html se ferait servir comme du HTML par le navigateur.
func TestDocumentDownloadNameExtensionComesFromContentType(t *testing.T) {
	got := documentDownloadName(store.PetDocument{FileName: "piege.html", ContentType: "application/pdf"})
	if got != "piege.pdf" {
		t.Fatalf("got %q, want piege.pdf", got)
	}

	got = documentDownloadName(store.PetDocument{FileName: "photo.pdf", ContentType: "image/png"})
	if got != "photo.png" {
		t.Fatalf("got %q, want photo.png", got)
	}
}

// Un nom vidé par l'assainissement (non-latin, ponctuation seule) ne doit pas
// produire un en-tête tronqué : il reste toujours un nom exploitable.
func TestDocumentDownloadNameAlwaysYieldsAName(t *testing.T) {
	for _, name := range []string{"", "   ", ".pdf", "日本語.pdf", "!!!.pdf"} {
		got := documentDownloadName(store.PetDocument{FileName: name, ContentType: "application/pdf"})
		if strings.TrimSuffix(got, ".pdf") == "" {
			t.Fatalf("%q → %q : le nom ne doit jamais être vide", name, got)
		}
	}
}
