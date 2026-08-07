package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Les profils pprof exposent piles et symboles : ils ne doivent apparaître que
// si PPROF_SECRET est posé, et seulement avec le bon header.

func TestPprofClosedWithoutSecret(t *testing.T) {
	api := newTestAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/debug/pprof/heap", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("pprof sans PPROF_SECRET = %d, want 404 (routes non montées)", rec.Code)
	}
}

func TestPprofRequiresSecretHeader(t *testing.T) {
	api := newTestAPIWithPprof(t, "bench-pprof-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/debug/pprof/heap", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("pprof sans header = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/debug/pprof/heap", nil)
	req.Header.Set("X-Pprof-Secret", "wrong")
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("pprof avec mauvais secret = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/debug/pprof/heap", nil)
	req.Header.Set("X-Pprof-Secret", "bench-pprof-secret")
	rec = httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pprof avec le bon secret = %d, want 200", rec.Code)
	}
	// Un profil nommé, pas la page d'index : le préfixe doit être réécrit.
	if ct := rec.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("Content-Type = %q, want le profil binaire", ct)
	}
}
