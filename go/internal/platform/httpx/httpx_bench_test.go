package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

// Chemin traversé par chaque réponse de l'API : mesurer les allocations ici
// permet de repérer une régression avant qu'elle ne coûte du GC en production.
// Style Go 1.26 : b.Loop ne bloque plus l'inlining du corps.

type petRow struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Species string   `json:"species"`
	Breed   string   `json:"breed"`
	Tags    []string `json:"tags"`
}

func petRows(n int) []petRow {
	out := make([]petRow, n)
	for i := range out {
		out[i] = petRow{
			ID:      "11111111-2222-3333-4444-555555555555",
			Name:    "Spirit",
			Species: "horse",
			Breed:   "Selle Français",
			Tags:    []string{"care", "kennel"},
		}
	}
	return out
}

func BenchmarkWriteDataSmall(b *testing.B) {
	payload := map[string]string{"status": "ok"}
	b.ReportAllocs()
	for b.Loop() {
		rec := httptest.NewRecorder()
		httpx.WriteData(rec, http.StatusOK, payload)
	}
}

func BenchmarkWriteDataList(b *testing.B) {
	payload := petRows(200)
	b.ReportAllocs()
	for b.Loop() {
		rec := httptest.NewRecorder()
		httpx.WriteData(rec, http.StatusOK, payload)
	}
}

func BenchmarkClientIP(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.4")
	b.ReportAllocs()
	for b.Loop() {
		_ = httpx.ClientIP(req)
	}
}

func BenchmarkRateLimiterAllow(b *testing.B) {
	rl := httpx.NewRateLimiter(1_000_000_000, time.Minute)
	b.ReportAllocs()
	for b.Loop() {
		_ = rl.Allow("198.51.100.4")
	}
}
