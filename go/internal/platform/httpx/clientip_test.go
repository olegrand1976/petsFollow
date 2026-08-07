package httpx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func echoClientIP(opts httpx.RouterOptions) http.Handler {
	r := httpx.NewBaseRouter(opts)
	r.Get("/probe", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(httpx.ClientIP(req)))
	})
	return r
}

func TestClientIPIgnoresCallerSuppliedHeaders(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name:    "no forwarding header falls back to RemoteAddr",
			headers: nil,
			want:    "192.0.2.1",
		},
		{
			name:    "True-Client-IP is ignored",
			headers: map[string]string{"True-Client-IP": "203.0.113.7"},
			want:    "192.0.2.1",
		},
		{
			name:    "X-Real-IP is ignored",
			headers: map[string]string{"X-Real-IP": "203.0.113.7"},
			want:    "192.0.2.1",
		},
		{
			name:    "leftmost XFF entry is ignored, rightmost wins",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.7, 198.51.100.4"},
			want:    "198.51.100.4",
		},
		{
			name:    "single XFF entry is the one our hop appended",
			headers: map[string]string{"X-Forwarded-For": "198.51.100.4"},
			want:    "198.51.100.4",
		},
		{
			name: "forged BFF headers without secret are ignored",
			headers: map[string]string{
				"X-Forwarded-For":         "198.51.100.4",
				httpx.HeaderPFClientIP:    "203.0.113.50",
				httpx.HeaderPFProxySecret: "guess",
			},
			want: "198.51.100.4",
		},
	}

	h := echoClientIP(httpx.RouterOptions{TrustedProxyHops: 1})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if got := rec.Body.String(); got != tc.want {
				t.Fatalf("client IP = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClientIPWithoutProxyKeepsRemoteAddr(t *testing.T) {
	h := echoClientIP(httpx.RouterOptions{TrustedProxyHops: 0})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Body.String(); got != "192.0.2.1" {
		t.Fatalf("client IP = %q, want RemoteAddr host 192.0.2.1", got)
	}
}

// Browser → Nuxt → API : GFE append l'IP Nuxt en dernier. Sans secret BFF,
// hops=1 prendrait Nuxt et tout le Pro partagerait une clé. Avec secret, on
// retient l'IP que la BFF a déjà résolue côté edge Nuxt.
func TestClientIPFromAuthenticatedBFF(t *testing.T) {
	const secret = "bff-shared-secret"
	h := echoClientIP(httpx.RouterOptions{TrustedProxyHops: 1, BFFProxySecret: secret})

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.9") // rightmost = egress Nuxt
	req.Header.Set(httpx.HeaderPFClientIP, "203.0.113.50")
	req.Header.Set(httpx.HeaderPFProxySecret, secret)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Body.String(); got != "203.0.113.50" {
		t.Fatalf("client IP = %q, want BFF-reported 203.0.113.50", got)
	}
}

func TestClientIPRejectsWrongBFFSecret(t *testing.T) {
	h := echoClientIP(httpx.RouterOptions{TrustedProxyHops: 1, BFFProxySecret: "bff-shared-secret"})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.4")
	req.Header.Set(httpx.HeaderPFClientIP, "203.0.113.50")
	req.Header.Set(httpx.HeaderPFProxySecret, "wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Body.String(); got != "198.51.100.4" {
		t.Fatalf("client IP = %q, want edge XFF 198.51.100.4", got)
	}
}

func TestClientIPRejectsNonIPBFFHeader(t *testing.T) {
	const secret = "bff-shared-secret"
	h := echoClientIP(httpx.RouterOptions{TrustedProxyHops: 1, BFFProxySecret: secret})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.4")
	req.Header.Set(httpx.HeaderPFClientIP, "not-an-ip")
	req.Header.Set(httpx.HeaderPFProxySecret, secret)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Body.String(); got != "198.51.100.4" {
		t.Fatalf("client IP = %q, want edge fallback", got)
	}
}

func TestRateLimitNotBypassableByHeaderRotation(t *testing.T) {
	const limit = 3
	spoofs := map[string]func(*http.Request, int){
		"X-Real-IP":      func(req *http.Request, i int) { req.Header.Set("X-Real-IP", fmt.Sprintf("203.0.113.%d", i)) },
		"True-Client-IP": func(req *http.Request, i int) { req.Header.Set("True-Client-IP", fmt.Sprintf("203.0.113.%d", i)) },
		"leftmost XFF":   func(req *http.Request, i int) { req.Header.Set("X-Forwarded-For", fmt.Sprintf("203.0.113.%d", i)) },
		"forged BFF IP": func(req *http.Request, i int) {
			req.Header.Set(httpx.HeaderPFClientIP, fmt.Sprintf("203.0.113.%d", i))
			req.Header.Set(httpx.HeaderPFProxySecret, "nope")
		},
		"duplicated XFF": func(req *http.Request, i int) {
			req.Header.Add("X-Forwarded-For", fmt.Sprintf("203.0.113.%d", i))
			req.Header.Add("X-Forwarded-For", fmt.Sprintf("203.0.113.%d", i+100))
		},
	}

	for name, spoof := range spoofs {
		t.Run(name, func(t *testing.T) {
			rl := httpx.NewRateLimiter(limit, time.Minute)
			r := httpx.NewBaseRouter(httpx.RouterOptions{TrustedProxyHops: 1, BFFProxySecret: "real-secret"})
			r.With(rl.Middleware).Post("/auth/login", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			blocked := false
			for i := range limit + 2 {
				req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
				spoof(req, i)
				req.Header.Add("X-Forwarded-For", "198.51.100.4")
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				if rec.Code == http.StatusTooManyRequests {
					blocked = true
					break
				}
			}
			if !blocked {
				t.Fatal("rotating a caller-supplied IP header bypassed the rate limit")
			}
		})
	}
}

func TestRateLimitHonoursAuthenticatedBFFClientIP(t *testing.T) {
	const secret = "bff-shared-secret"
	const limit = 2
	rl := httpx.NewRateLimiter(limit, time.Minute)
	r := httpx.NewBaseRouter(httpx.RouterOptions{TrustedProxyHops: 1, BFFProxySecret: secret})
	r.With(rl.Middleware).Post("/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Deux clients distincts derrière le même egress Nuxt doivent avoir des quotas séparés.
	for _, clientIP := range []string{"203.0.113.10", "203.0.113.20"} {
		for i := range limit {
			req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
			req.Header.Set("X-Forwarded-For", "198.51.100.9") // Nuxt egress
			req.Header.Set(httpx.HeaderPFClientIP, clientIP)
			req.Header.Set(httpx.HeaderPFProxySecret, secret)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("client %s req %d = %d, want 200", clientIP, i, rec.Code)
			}
		}
	}
}
