package rag

import "testing"

func TestChunk_Overlap(t *testing.T) {
	text := stringsRepeat("abcdefghij", 200) // 2000 runes
	parts := Chunk(text, 500, 100)
	if len(parts) < 4 {
		t.Fatalf("want several chunks got %d", len(parts))
	}
	if parts[0] == "" {
		t.Fatal("empty first chunk")
	}
}

func TestChunk_Short(t *testing.T) {
	parts := Chunk("hello world", 1200, 150)
	if len(parts) != 1 || parts[0] != "hello world" {
		t.Fatalf("got %#v", parts)
	}
}

func TestIsAllowedMime(t *testing.T) {
	if !IsAllowedMime("application/pdf") || !IsAllowedMime("text/plain; charset=utf-8") {
		t.Fatal("expected allowed")
	}
	if IsAllowedMime("image/png") {
		t.Fatal("png not allowed")
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
