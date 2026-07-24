package httpx

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter — compteur à fenêtre fixe par IP, en mémoire.
// Suffisant pour freiner credential stuffing / spam d'inscriptions ;
// multi-instance : chaque instance applique sa propre limite.
type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string]*rateBucket
	limit  int
	window time.Duration
	lastGC time.Time
}

type rateBucket struct {
	count int
	reset time.Time
}

// NewRateLimiter crée un limiteur `limit` requêtes / `window`. limit <= 0 désactive le limiteur.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:   make(map[string]*rateBucket),
		limit:  limit,
		window: window,
		lastGC: time.Now(),
	}
}

func (rl *RateLimiter) allow(key string) bool {
	if rl.limit <= 0 {
		return true
	}
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if now.Sub(rl.lastGC) > 10*rl.window {
		for k, b := range rl.hits {
			if now.After(b.reset) {
				delete(rl.hits, k)
			}
		}
		rl.lastGC = now
	}
	b, ok := rl.hits[key]
	if !ok || now.After(b.reset) {
		rl.hits[key] = &rateBucket{count: 1, reset: now.Add(rl.window)}
		return true
	}
	b.count++
	return b.count <= rl.limit
}

// Middleware limite par IP (middleware.RealIP est appliqué en amont sur le routeur de base).
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(r.RemoteAddr) {
			WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
