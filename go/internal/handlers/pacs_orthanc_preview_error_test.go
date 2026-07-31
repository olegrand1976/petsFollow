package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestOrthancPreviewOutOfRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		err   error
		frame int
		want  bool
	}{
		{"nil", nil, 1, false},
		{"plain", errors.New("orthanc_preview_404"), 1, false},
		{"404 frame0", &orthancStatusError{Resource: "preview", Status: http.StatusNotFound}, 0, false},
		{"404 frame1", &orthancStatusError{Resource: "preview", Status: http.StatusNotFound}, 1, true},
		{"400 frame2", &orthancStatusError{Resource: "preview", Status: http.StatusBadRequest}, 2, true},
		{"500 frame1 not eof", &orthancStatusError{Resource: "preview", Status: http.StatusInternalServerError}, 1, false},
		{"wrapped 400", fmt.Errorf("wrap: %w", &orthancStatusError{Resource: "preview", Status: http.StatusBadRequest}), 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := orthancPreviewOutOfRange(tc.err, tc.frame); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestOrthancBlobUnavailable(t *testing.T) {
	t.Parallel()
	if !orthancBlobUnavailable(&orthancStatusError{Resource: "file", Status: 500}) {
		t.Fatal("file 500")
	}
	if !orthancBlobUnavailable(&orthancStatusError{Resource: "preview", Status: 404}) {
		t.Fatal("preview 404")
	}
	if orthancBlobUnavailable(&orthancStatusError{Resource: "preview", Status: 400}) {
		t.Fatal("preview 400 is OOR not blob")
	}
	if orthancBlobUnavailable(fmt.Errorf("nope")) {
		t.Fatal("plain")
	}
}
