package media

import (
	"errors"
	"os"
	"testing"

	"cloud.google.com/go/storage"
)

func TestIsNotExist(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"os", os.ErrNotExist, true},
		{"gcs", storage.ErrObjectNotExist, true},
		{"wrapped gcs", errors.Join(errors.New("open"), storage.ErrObjectNotExist), true},
		{"other", errors.New("permission denied"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNotExist(tc.err); got != tc.want {
				t.Fatalf("got %v want %v for %v", got, tc.want, tc.err)
			}
		})
	}
}
