package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestValidAccessPermission(t *testing.T) {
	cases := []struct {
		p    string
		want bool
	}{
		{"read", true},
		{"write_notes", true},
		{"full", true},
		{"", false},
		{"admin", false},
		{"WRITE_NOTES", false},
	}
	for _, tc := range cases {
		if got := store.ValidAccessPermission(tc.p); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.p, got, tc.want)
		}
	}
}
