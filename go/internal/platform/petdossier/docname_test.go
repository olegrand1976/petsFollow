package petdossier

import "testing"

func TestDocumentZipPath(t *testing.T) {
	cases := []struct {
		title, file, key, want string
	}{
		{"Radio thorax", "scan.pdf", "", "documents/Radio_thorax.pdf"},
		{"Analyse", "lab.xlsx", "documents/x/lab.xlsx", "documents/Analyse.xlsx"},
		{"", "report.pdf", "", "documents/report.pdf"},
		{"Note", "", "pets/docs/note.pdf", "documents/Note.pdf"},
		{"../evil", "ok.pdf", "", "documents/document.pdf"},
		{"Radio.pdf", "scan.pdf", "", "documents/Radio.pdf"},
	}
	for _, tc := range cases {
		got := DocumentZipPath(tc.title, tc.file, tc.key)
		if got != tc.want {
			t.Fatalf("DocumentZipPath(%q,%q,%q)=%q want %q", tc.title, tc.file, tc.key, got, tc.want)
		}
	}
}
