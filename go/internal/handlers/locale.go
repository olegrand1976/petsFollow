package handlers

import (
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) localeFromUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, err := authx.FromContext(r.Context()); err == nil {
			if locale, err := a.store.GetUserPreferredLocale(r.Context(), id.UserID); err == nil {
				r = r.WithContext(i18n.WithLocale(r.Context(), locale))
			}
			// After seed/redeploy, JWT practice_id can lag behind identity.users.practice_id.
			// Refresh it so /clients and other practice-scoped routes stay correct without re-login.
			if id.Role == kernel.RoleVet {
				if u, err := a.store.GetUserByID(r.Context(), id.UserID); err == nil && u.PracticeID != "" && u.PracticeID != id.PracticeID {
					id.PracticeID = u.PracticeID
					r = r.WithContext(authx.WithIdentity(r.Context(), id))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

type updateLocaleReq struct {
	Locale string `json:"locale"`
}

func (a *API) updateMeLocale(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		httpx.WriteErrorLocalized(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	var req updateLocaleReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Locale == "" {
		httpx.WriteErrorLocalized(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	locale, ok := i18n.MatchSupported(req.Locale)
	if !ok {
		httpx.WriteErrorLocalized(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.UpdateUserLocale(r.Context(), id.UserID, locale); err != nil {
		httpx.WriteErrorLocalized(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"preferredLocale": locale})
}
