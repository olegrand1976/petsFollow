package store

import "testing"

func TestIsValidSupportStatusTransition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		from, to string
		ok       bool
	}{
		{"open", "in_progress", true},
		{"open", "closed", true},
		{"open", "done", false},
		{"open", "to_test", false},
		{"in_progress", "to_test", true},
		{"in_progress", "done", false},
		{"to_test", "done", true},
		{"to_test", "in_progress", true},
		{"done", "closed", true},
		{"done", "open", false},
		{"closed", "open", true},
		{"closed", "done", false},
		{"open", "open", true},
		{"", "open", false},
		{"open", "nope", false},
	}
	for _, tc := range cases {
		got := IsValidSupportStatusTransition(tc.from, tc.to)
		if got != tc.ok {
			t.Fatalf("%s→%s: got %v want %v", tc.from, tc.to, got, tc.ok)
		}
	}
	next := AllowedSupportStatusTransitions(SupportStatusOpen)
	if len(next) != 2 {
		t.Fatalf("open next %#v", next)
	}
}
