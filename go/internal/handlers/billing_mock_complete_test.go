package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrefersMockCompleteHTML(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		accept string
		query  string
		want   bool
	}{
		{name: "browser", accept: "text/html,application/xhtml+xml", want: true},
		{name: "empty accept", accept: "", want: true},
		{name: "json accept", accept: "application/json", want: false},
		{name: "curl star", accept: "*/*", want: false},
		{name: "format json overrides", accept: "text/html", query: "format=json", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/dev/mock-complete?"+tc.query, nil)
			req.Header.Set("Accept", tc.accept)
			if got := prefersMockCompleteHTML(req); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestAllowedMockReturnURL(t *testing.T) {
	t.Parallel()
	if !allowedMockReturnURL("petsfollow://payment/success") {
		t.Fatal("deep link should be allowed")
	}
	if allowedMockReturnURL("https://app.example/ok") {
		t.Fatal("https must be rejected (open redirect)")
	}
	if allowedMockReturnURL("http://localhost/ok") {
		t.Fatal("http must be rejected")
	}
	if allowedMockReturnURL("javascript:alert(1)") {
		t.Fatal("javascript should be rejected")
	}
	if allowedMockReturnURL(`petsfollow://x"><script>`) {
		t.Fatal("markup in deep link should be rejected")
	}
}

func TestWriteMockCompleteResultHTML(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/x?success_url=petsfollow://payment/success", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	ok := writeMockCompleteResult(rec, req, "petsfollow://payment/success")
	if !ok {
		t.Fatal("expected HTML response")
	}
	body := rec.Body.String()
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type=%q", ct)
	}
	if !strings.Contains(body, "Payment completed") {
		t.Fatalf("body=%q", body)
	}
	if !strings.Contains(body, "petsfollow://payment/success") {
		t.Fatalf("missing deep link: %q", body)
	}
	if strings.Contains(body, "<script") {
		t.Fatal("must not embed inline script")
	}
}

func TestWriteMockCompleteResultRejectsHTTPSRedirect(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	if writeMockCompleteResult(rec, req, "https://evil.example/phish") {
		t.Fatal("https success_url must not write HTML redirect")
	}
}

func TestWriteMockCompleteResultKeepsJSONPath(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept", "*/*")
	rec := httptest.NewRecorder()
	if writeMockCompleteResult(rec, req, "") {
		t.Fatal("empty success_url must not write HTML")
	}
}
