package handlers

import (
	"net/http"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// requireTermsAcceptedMiddleware — clients sans terms_accepted_at : allowlist
// activation (me / accept-terms / password / export / delete / locale).
// Les rôles Pro ne sont pas bloqués ici (pas d'UI accept-terms Nuxt ; team invite).
func (a *API) requireTermsAcceptedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := authx.FromContext(r.Context())
		if err != nil || id.Role != kernel.RoleClient {
			next.ServeHTTP(w, r)
			return
		}
		if termsAllowlisted(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		ok, err := a.store.UserHasAcceptedTerms(r.Context(), id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !ok {
			writeErr(w, r, http.StatusForbidden, "consent_required", "consent_required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func termsAllowlisted(method, path string) bool {
	p := strings.TrimPrefix(path, "/api/v1")
	if p == "" {
		p = path
	}
	switch {
	case method == http.MethodGet && p == "/me":
		return true
	case method == http.MethodPost && p == "/me/accept-terms":
		return true
	case method == http.MethodPatch && p == "/me/password":
		return true
	case method == http.MethodPatch && p == "/me/locale":
		return true
	case method == http.MethodGet && p == "/me/export":
		return true
	case method == http.MethodDelete && p == "/me":
		return true
	default:
		return false
	}
}
