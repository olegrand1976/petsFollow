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
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain", errors.New("orthanc_preview_404"), false},
		{"404", &orthancStatusError{Resource: "preview", Status: http.StatusNotFound}, true},
		{"400", &orthancStatusError{Resource: "preview", Status: http.StatusBadRequest}, true},
		{"500", &orthancStatusError{Resource: "preview", Status: http.StatusBadGateway}, false},
		{"other resource", &orthancStatusError{Resource: "file", Status: http.StatusNotFound}, false},
		{"wrapped 400", fmt.Errorf("wrap: %w", &orthancStatusError{Resource: "preview", Status: http.StatusBadRequest}), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := orthancPreviewOutOfRange(tc.err); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
