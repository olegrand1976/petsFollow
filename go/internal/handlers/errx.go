package handlers

import (
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

func writeErr(w http.ResponseWriter, r *http.Request, status int, code, msgKey string) {
	httpx.WriteErrorLocalized(w, r, status, code, msgKey)
}

func writeInternal(w http.ResponseWriter, r *http.Request) {
	httpx.WriteErrorLocalized(w, r, http.StatusInternalServerError, "internal", "internal")
}

func localeOf(r *http.Request) string {
	return i18n.FromContext(r.Context())
}

func t(r *http.Request, key string, vars map[string]string) string {
	return i18n.T(localeOf(r), key, vars)
}
