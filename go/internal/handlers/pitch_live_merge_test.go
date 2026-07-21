package handlers

import "testing"

func TestMergeTranscriptChunk(t *testing.T) {
	cases := []struct {
		cur, in, want string
	}{
		{"", "Allo ?", "Allo ?"},
		{"Bon", "Bonjour", "Bonjour"},
		{"Bonjour", "Bon", "Bonjour"},
		{"Bon", "jour", "Bon jour"},
		{"Bon ", "jour", "Bon jour"},
		{"Cabinet", "Cabinet vétérinaire", "Cabinet vétérinaire"},
		{"hello world", "world!", "hello world!"},
	}
	for _, c := range cases {
		got := mergeTranscriptChunk(c.cur, c.in)
		if got != c.want {
			t.Errorf("merge(%q, %q) = %q, want %q", c.cur, c.in, got, c.want)
		}
	}
}
