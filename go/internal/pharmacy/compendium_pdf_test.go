package pharmacy

import (
	"bytes"
	"testing"

	"github.com/phpdave11/gofpdf"
)

func TestChunkPageRanges(t *testing.T) {
	got := ChunkPageRanges(1, 5, 2)
	if len(got) != 3 || got[0] != [2]int{1, 2} || got[1] != [2]int{3, 4} || got[2] != [2]int{5, 5} {
		t.Fatalf("got %#v", got)
	}
}

func TestLooksLikePDF(t *testing.T) {
	if !LooksLikePDF([]byte("%PDF-1.4")) {
		t.Fatal("expected pdf")
	}
	if LooksLikePDF([]byte("not")) {
		t.Fatal("expected false")
	}
}

func TestExtractPageRangeTrim(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	for i := 1; i <= 4; i++ {
		pdf.AddPage()
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, "page")
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	n, err := PageCount(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("pages=%d", n)
	}
	slice, err := ExtractPageRange(raw, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	sn, err := PageCount(slice)
	if err != nil {
		t.Fatal(err)
	}
	if sn != 2 {
		t.Fatalf("slice pages=%d", sn)
	}
	if _, err := ExtractPageRange(raw, 2, 99); err == nil {
		t.Fatal("expected out of range")
	}
}
