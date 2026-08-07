package httpx

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"net/netip"

	"github.com/go-chi/chi/v5/middleware"
)

// Headers BFF → API. Sans secret partagé, X-PF-Client-IP est ignoré (sinon
// n'importe quel client direct forgerait sa clé de rate limit).
const (
	HeaderPFClientIP    = "X-PF-Client-IP"
	HeaderPFProxySecret = "X-PF-Proxy-Secret"
)

// DefaultTrustedProxyHops — Cloud Run interpose exactement un hop devant le conteneur.
const DefaultTrustedProxyHops = 1

type ctxKey int

const bffClientIPCtxKey ctxKey = 1

// RouterOptions configure le routeur de base (IP client + reverse proxies).
type RouterOptions struct {
	TrustedProxyHops int
	// BFFProxySecret — partagé avec la BFF Nuxt. Vide = n'accepte pas X-PF-Client-IP.
	BFFProxySecret string
}

// clientIPMiddleware résout l'IP vue par notre edge (entrée XFF la plus à droite).
//
// chi a déprécié middleware.RealIP en v5.3.0 (GHSA-3fxj-6jh8-hvhx) : il réécrivait
// RemoteAddr depuis des en-têtes fournis par l'appelant. Derrière Cloud Run,
// seule l'entrée ajoutée par GFE est digne de confiance pour un client direct
// (Flutter). Le trafic Pro passe par la BFF : voir bffClientIPMiddleware.
func clientIPMiddleware(hops int) func(http.Handler) http.Handler {
	if hops <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return middleware.ClientIPFromXFFTrustedProxies(hops)
}

// bffClientIPMiddleware lit l'IP client pré-résolue par Nuxt, uniquement si le
// secret partagé matche en temps constant. Sans ça, hops=1 derrière
// Browser→Nuxt→API prendrait l'IP d'egress Nuxt et tout le Pro partagerait une
// seule clé de quota /auth/*.
func bffClientIPMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret != "" &&
				subtle.ConstantTimeCompare([]byte(r.Header.Get(HeaderPFProxySecret)), []byte(secret)) == 1 {
				if ip := parseClientIP(r.Header.Get(HeaderPFClientIP)); ip != "" {
					r = r.WithContext(context.WithValue(r.Context(), bffClientIPCtxKey, ip))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseClientIP(raw string) string {
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return ""
	}
	return addr.Unmap().WithZone("").String()
}

// ClientIP retourne l'IP appelante de confiance pour le rate limit :
// 1. IP BFF authentifiée (Pro web),
// 2. IP edge XFF (Flutter / appels directs derrière Cloud Run),
// 3. hôte de RemoteAddr (fail-closed).
func ClientIP(r *http.Request) string {
	if ip, _ := r.Context().Value(bffClientIPCtxKey).(string); ip != "" {
		return ip
	}
	if ip := middleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
