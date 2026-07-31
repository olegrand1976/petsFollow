package store

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"vet.demo@petsfollow.test", "v***@petsfollow.test"},
		{"A@b.co", "a***@b.co"},
		{"", "***"},
		{"nodomain", "***"},
		{"@x.com", "***"},
	}
	for _, tc := range cases {
		if got := MaskEmail(tc.in); got != tc.want {
			t.Errorf("MaskEmail(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestLikeContainsPatternEscapesWildcards(t *testing.T) {
	got := likeContainsPattern(`a%_b\c`)
	want := `%a\%\_b\\c%`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
