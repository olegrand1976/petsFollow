package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Regression: do() must not cancel the request context before the caller reads
// the body — otherwise larger Orthanc preview/file payloads fail with
// "context canceled" while Orthanc itself already returned 200.
func TestOrthancClientDoKeepsContextAliveUntilBodyClose(t *testing.T) {
	t.Parallel()
	payload := strings.Repeat("x", 256*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		const chunkSize = 4096
		for off := 0; off < len(payload); off += chunkSize {
			end := off + chunkSize
			if end > len(payload) {
				end = len(payload)
			}
			if _, err := w.Write([]byte(payload[off:end])); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
			time.Sleep(2 * time.Millisecond)
		}
	}))
	t.Cleanup(srv.Close)

	client := newOrthancClient(srv.URL, "", "", false)
	// Force Transport without client-level Timeout so only the per-request ctx matters.
	client.http = &http.Client{}

	resp, err := client.do(context.Background(), http.MethodGet, "/big", nil, "", 5*time.Second)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll (context must stay alive until Close): %v", err)
	}
	if len(body) != len(payload) {
		t.Fatalf("body len=%d want %d", len(body), len(payload))
	}
}
