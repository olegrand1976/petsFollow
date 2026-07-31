package pharmacy

import (
	"bytes"
	"fmt"
	"strings"
)

// CompendiumPagesPerChunk is retained for tests / future real page-trim.
// V1 extract sends the full PDF once (see handlers.runCompendiumExtract).
const CompendiumPagesPerChunk = 2

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
		to := p + size - 1
		if to > end {
			to = end
		}
		out = append(out, [2]int{p, to})
	}
	return out
}

// LooksLikePDF soft-checks magic bytes.
func LooksLikePDF(b []byte) bool {
	return len(b) >= 4 && strings.HasPrefix(string(b[:4]), "%PDF")
}
