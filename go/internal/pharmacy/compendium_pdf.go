package pharmacy

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// CompendiumPagesPerChunk is the max PDF pages sent to Gemini per extract call.
// Large ranges (e.g. 159 pages) time out as a single request; chunking + Trim keeps each call small.
var CompendiumPagesPerChunk = 6

// PageCount estimates PDF page count (heuristic on page objects).
// Good enough to reject absurd pageEnd; not a full PDF parser.
func PageCount(pdf []byte) (int, error) {
	if len(pdf) == 0 {
		return 0, fmt.Errorf("empty_pdf")
	}
	if !LooksLikePDF(pdf) {
		return 0, fmt.Errorf("invalid_pdf")
	}
	// Count leaf page dicts; subtract /Type /Pages trees.
	n := bytes.Count(pdf, []byte("/Type /Page")) + bytes.Count(pdf, []byte("/Type/Page"))
	n -= bytes.Count(pdf, []byte("/Type /Pages")) + bytes.Count(pdf, []byte("/Type/Pages"))
	if n < 1 {
		// Some writers omit spaces differently — fall back to at least 1.
		n = 1
	}
	return n, nil
}

// ChunkPageRanges splits [start, end] into inclusive ranges of at most size pages.
func ChunkPageRanges(start, end, size int) [][2]int {
	if size < 1 {
		size = CompendiumPagesPerChunk
	}
	if start < 1 || end < start {
		return nil
	}
	var out [][2]int
	for p := start; p <= end; p += size {
		to := min(p+size-1, end)
		out = append(out, [2]int{p, to})
	}
	return out
}

// ExtractPDFPages returns a new PDF containing only absolute pages [start, end] (1-based inclusive).
func ExtractPDFPages(pdf []byte, start, end int) ([]byte, error) {
	if len(pdf) == 0 {
		return nil, fmt.Errorf("empty_pdf")
	}
	if !LooksLikePDF(pdf) {
		return nil, fmt.Errorf("invalid_pdf")
	}
	if start < 1 || end < start {
		return nil, fmt.Errorf("invalid_page_range")
	}
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	var out bytes.Buffer
	sel := fmt.Sprintf("%d-%d", start, end)
	if err := api.Trim(bytes.NewReader(pdf), &out, []string{sel}, conf); err != nil {
		return nil, fmt.Errorf("pdf_trim: %w", err)
	}
	if out.Len() == 0 || !LooksLikePDF(out.Bytes()) {
		return nil, fmt.Errorf("pdf_trim_empty")
	}
	return out.Bytes(), nil
}

// RemapSourcePages normalizes Gemini sourcePage to absolute PDF page numbers for a chunk.
// Accepts either absolute (absStart–absEnd) or relative (1–chunkLen) values.
func RemapSourcePages(meds []ExtractedMedication, absStart, absEnd int) {
	if absStart < 1 || absEnd < absStart {
		return
	}
	chunkLen := absEnd - absStart + 1
	for i := range meds {
		if meds[i].SourcePage == nil {
			p := absStart
			meds[i].SourcePage = &p
			continue
		}
		n := *meds[i].SourcePage
		if n >= absStart && n <= absEnd {
			continue
		}
		if n >= 1 && n <= chunkLen {
			abs := absStart + n - 1
			meds[i].SourcePage = &abs
		}
	}
}

// LooksLikePDF soft-checks magic bytes.
func LooksLikePDF(b []byte) bool {
	return len(b) >= 4 && strings.HasPrefix(string(b[:4]), "%PDF")
}
