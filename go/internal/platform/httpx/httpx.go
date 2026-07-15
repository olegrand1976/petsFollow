package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

func NewBaseRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	return r
}

func WriteData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: &APIError{Code: code, Message: message}})
}

func WriteErrorLocalized(w http.ResponseWriter, r *http.Request, status int, code, msgKey string) {
	if msgKey == "" {
		msgKey = code
	}
	locale := i18n.FromContext(r.Context())
	msg := i18n.T(locale, "errors."+msgKey, nil)
	WriteError(w, status, code, msg)
}

func LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := i18n.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		next.ServeHTTP(w, r.WithContext(i18n.WithLocale(r.Context(), locale)))
	})
}

func AuthMiddleware(issuer *authx.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				WriteErrorLocalized(w, r, http.StatusUnauthorized, "unauthorized", "missing_bearer_token")
				return
			}
			id, err := issuer.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				WriteErrorLocalized(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
				return
			}
			next.ServeHTTP(w, r.WithContext(authx.WithIdentity(r.Context(), id)))
		})
	}
}

func MountHealth(r chi.Router, ready func(context.Context) error) {
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := ready(r.Context()); err != nil {
			WriteError(w, http.StatusServiceUnavailable, "not_ready", err.Error())
			return
		}
		WriteData(w, http.StatusOK, map[string]string{"status": "ready"})
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
