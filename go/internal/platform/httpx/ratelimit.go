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

// Allow reports whether key may proceed (same window as Middleware).
func (rl *RateLimiter) Allow(key string) bool {
	return rl.allow(key)
}

// Middleware limite par IP appelante de confiance (cf. ClientIP) : la clé ne doit
// pas dépendre d'un en-tête que l'appelant choisit, sinon le quota se contourne
// en faisant tourner X-Forwarded-For à chaque requête.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(ClientIP(r)) {
			WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
