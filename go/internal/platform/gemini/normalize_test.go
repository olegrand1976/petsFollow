package gemini

import (
	"strings"
	"testing"
)

func TestNormalizeVisitReportTextStripsNULAndControls(t *testing.T) {
	in := "hello\x00world\x01\nnext"
	got := NormalizeVisitReportText(in)
	if got != "helloworld\nnext" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeVisitReportTextCapsLength(t *testing.T) {
	long := stringsRepeat("a", maxVisitReportRunes+50)
	got := NormalizeVisitReportText(long)
	if len([]rune(got)) != maxVisitReportRunes {
		t.Fatalf("want %d runes got %d", maxVisitReportRunes, len([]rune(got)))
	}
}

func TestNormalizeVisitReportTextEscapeCorpus(t *testing.T) {
	in := "a\x00b \"quotes\" 'apos' «fr» back\\slash &amp; emoji 🐶\nkeep"
	got := NormalizeVisitReportText(in)
	if strings.Contains(got, "\x00") {
		t.Fatalf("NUL not stripped: %q", got)
	}
	for _, want := range []string{`"quotes"`, `'apos'`, `«fr»`, `back\slash`, `&amp;`, `🐶`, "\nkeep"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
