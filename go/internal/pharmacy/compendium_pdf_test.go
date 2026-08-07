package pharmacy

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/phpdave11/gofpdf"
)

func TestChunkPageRanges(t *testing.T) {
	got := ChunkPageRanges(1, 5, 2)
	if len(got) != 3 || got[0] != [2]int{1, 2} || got[1] != [2]int{3, 4} || got[2] != [2]int{5, 5} {
		t.Fatalf("got %#v", got)
	}
	if ChunkPageRanges(363, 521, 6)[0] != [2]int{363, 368} {
		t.Fatalf("first chunk %#v", ChunkPageRanges(363, 521, 6)[0])
	}
	n := len(ChunkPageRanges(363, 521, 6))
	if n != 27 { // 159 pages / 6
		t.Fatalf("chunks=%d want 27", n)
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

func TestExtractPDFPages(t *testing.T) {
	pdf := buildMultipagePDF(t, 5)
	trimmed, err := ExtractPDFPages(pdf, 2, 3)
	if err != nil {
		t.Fatalf("trim: %v", err)
	}
	if !LooksLikePDF(trimmed) {
		t.Fatal("trimmed not pdf")
	}
	// Heuristic PageCount under-counts optimized PDFs; use pdfcpu.
	n, err := api.PageCount(bytes.NewReader(trimmed), nil)
	if err != nil {
		t.Fatalf("pagecount: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 pages got %d", n)
	}
	if _, err := ExtractPDFPages(pdf, 0, 1); err == nil {
		t.Fatal("expected invalid range")
	}
}

func TestRemapSourcePages(t *testing.T) {
	rel := 2
	absAlready := 365
	meds := []ExtractedMedication{
		{Name: "A", SourcePage: &rel},
		{Name: "B", SourcePage: &absAlready},
		{Name: "C"},
	}
	RemapSourcePages(meds, 363, 368)
	if meds[0].SourcePage == nil || *meds[0].SourcePage != 364 {
		t.Fatalf("relative remap %#v", meds[0].SourcePage)
	}
	if meds[1].SourcePage == nil || *meds[1].SourcePage != 365 {
		t.Fatalf("absolute keep %#v", meds[1].SourcePage)
	}
	if meds[2].SourcePage == nil || *meds[2].SourcePage != 363 {
		t.Fatalf("default %#v", meds[2].SourcePage)
	}
}

func buildMultipagePDF(t *testing.T, pages int) []byte {
	t.Helper()
	pdf := gofpdf.New("P", "mm", "A4", "")
	for i := range pages {
		pdf.AddPage()
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Page %d", i+1))
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("pdf: %v", err)
	}
	return buf.Bytes()
}
