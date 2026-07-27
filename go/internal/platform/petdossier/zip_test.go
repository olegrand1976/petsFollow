package petdossier

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// Le nom de fichier vient d'un upload client et finit dans une archive ouverte par
// un pro externe : aucune entrée ne doit pouvoir s'extraire hors du dossier cible.
func TestBuildZipDropsPathTraversalEntries(t *testing.T) {
	data, err := BuildZip([]ZipFile{
		{Name: "dossier.pdf", Data: []byte("%PDF-1.4")},
		{Name: "documents/../../../etc/cron.d/pwn", Data: []byte("payload")},
		{Name: "documents/..\\..\\windows\\system32\\evil.dll", Data: []byte("payload")},
		{Name: "/absolu/passwd", Data: []byte("payload")},
	})
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "/") || f.Name == ".." || strings.HasPrefix(f.Name, "../") {
			t.Fatalf("entrée d'archive hors du dossier cible: %q", f.Name)
		}
		names = append(names, f.Name)
	}
	if len(names) != 2 {
		t.Fatalf("attendu dossier.pdf + l'entrée absolue nettoyée, obtenu %q", names)
	}
	if names[0] != "dossier.pdf" {
		t.Fatalf("dossier.pdf manquant: %q", names)
	}
}

func TestSanitizeZipName(t *testing.T) {
	dropped := []string{"..", "../x", "documents/../../x", ".", "", "   "}
	for _, in := range dropped {
		if got := sanitizeZipName(in); got != "" {
			t.Fatalf("sanitizeZipName(%q) = %q, attendu vide", in, got)
		}
	}
	kept := map[string]string{
		"dossier.pdf":            "dossier.pdf",
		"/carnet.pdf":            "carnet.pdf",
		"documents/mon scan.pdf": "documents/mon_scan.pdf",
		"documents/a/../b.pdf":   "documents/b.pdf",
	}
	for in, want := range kept {
		if got := sanitizeZipName(in); got != want {
			t.Fatalf("sanitizeZipName(%q) = %q, attendu %q", in, got, want)
		}
	}
}
