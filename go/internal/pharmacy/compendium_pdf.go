package pharmacy

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const CompendiumPagesPerChunk = 2

// PageCount returns the PDF page count via pdfcpu.
func PageCount(pdf []byte) (int, error) {
	if len(pdf) == 0 {
		return 0, fmt.Errorf("empty_pdf")
	}
	if !LooksLikePDF(pdf) {
		return 0, fmt.Errorf("invalid_pdf")
	}
	n, err := api.PageCount(bytes.NewReader(pdf), nil)
	if err != nil {
		return 0, fmt.Errorf("pdf_page_count: %w", err)
	}
	if n < 1 {
		return 0, fmt.Errorf("pdf_no_pages")
	}
	return n, nil
}

// ExtractPageRange returns a new PDF containing only pages [start, end] (1-based inclusive).
func ExtractPageRange(pdf []byte, start, end int) ([]byte, error) {
	if len(pdf) == 0 {
		return nil, fmt.Errorf("empty_pdf")
	}
	if start < 1 || end < start {
		return nil, fmt.Errorf("invalid_page_range")
	}
	total, err := PageCount(pdf)
	if err != nil {
		return nil, err
	}
	if end > total {
		return nil, fmt.Errorf("page_end_out_of_range:%d>%d", end, total)
	}
	if start == 1 && end == total {
		// Full document — avoid a no-op rewrite when possible.
		return pdf, nil
	}
	sel := fmt.Sprintf("%d-%d", start, end)
	conf := model.NewDefaultConfiguration()
	var out bytes.Buffer
	if err := api.Trim(bytes.NewReader(pdf), &out, []string{sel}, conf); err != nil {
		return nil, fmt.Errorf("pdf_trim: %w", err)
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("pdf_trim_empty")
	}
	return out.Bytes(), nil
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
		to := p + size - 1
		if to > end {
			to = end
		}
		out = append(out, [2]int{p, to})
	}
	return out
}

// ReadAllLimited reads at most max+1 bytes (caller checks size).
func ReadAllLimited(r io.Reader, max int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, max+1))
}

// LooksLikePDF soft-checks magic bytes.
func LooksLikePDF(b []byte) bool {
	return len(b) >= 4 && strings.HasPrefix(string(b[:4]), "%PDF")
}
