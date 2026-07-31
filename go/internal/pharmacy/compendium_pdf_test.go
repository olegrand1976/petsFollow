package pharmacy

import "testing"

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
