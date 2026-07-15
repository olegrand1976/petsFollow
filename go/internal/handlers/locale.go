package handlers

import (
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

func (a *API) localeFromUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, err := authx.FromContext(r.Context()); err == nil {
			if locale, err := a.store.GetUserPreferredLocale(r.Context(), id.UserID); err == nil {
				r = r.WithContext(i18n.WithLocale(r.Context(), locale))
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
	locale := i18n.NormalizeLocale(req.Locale)
	supported := false
	for _, loc := range i18n.Supported {
		if req.Locale == loc {
			supported = true
			locale = loc
			break
		}
	}
	if !supported {
		httpx.WriteErrorLocalized(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.UpdateUserLocale(r.Context(), id.UserID, locale); err != nil {
		httpx.WriteErrorLocalized(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"preferredLocale": locale})
}
