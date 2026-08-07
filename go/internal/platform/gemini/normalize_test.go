package gemini

import "testing"

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

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for range n {
		b = append(b, s...)
	}
	return string(b)
}
