package rag

import (
	"strings"
	"unicode/utf8"
)

const (
	DefaultChunkSize    = 1200
	DefaultChunkOverlap = 150
	// MaxChunks caps embedding fan-out (1 Gemini call per chunk).
	MaxChunks = 80
)

// Chunk splits text into overlapping rune-aware windows for embedding.
func Chunk(text string, size, overlap int) []string {
	if size <= 0 {
		size = DefaultChunkSize
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 4
	}
	cleaned := strings.TrimSpace(normalizeNewlines(text))
	if cleaned == "" {
		return nil
	}
	runes := []rune(cleaned)
	if len(runes) <= size {
		return []string{cleaned}
	}
	var out []string
	step := size - overlap
	for start := 0; start < len(runes); start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		part := strings.TrimSpace(string(runes[start:end]))
		if part != "" {
			out = append(out, part)
		}
		if end >= len(runes) {
			break
		}
	}
	return out
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// IsAllowedMime reports whether the MIME is accepted for RAG upload.
func IsAllowedMime(mime string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0])) {
	case "application/pdf", "text/plain", "text/markdown":
		return true
	default:
		return false
	}
}

// DetectMimeFromFilename falls back when Content-Type is missing/octet-stream.
func DetectMimeFromFilename(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return "text/markdown"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	default:
		return ""
	}
}

// LooksLikePDF mirrors pharmacy magic-byte check without importing pharmacy.
func LooksLikePDF(data []byte) bool {
	return len(data) >= 5 && string(data[:5]) == "%PDF-"
}

// ValidUTF8Text returns true if data is mostly valid UTF-8 text (for .txt/.md).
func ValidUTF8Text(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if !utf8.Valid(data) {
		return false
	}
	// Reject if too many NUL bytes (binary).
	nuls := 0
	for _, b := range data {
		if b == 0 {
			nuls++
		}
	}
	return nuls == 0
}
