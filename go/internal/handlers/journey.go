package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func (a *API) registerJourneyPublicRoutes(r chi.Router) {
	r.Get("/public/journey/unsubscribe", a.journeyUnsubscribe)
	r.Post("/public/journey/unsubscribe", a.journeyUnsubscribe)
}

func (a *API) journeyUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" && r.Body != nil {
		var body struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		token = strings.TrimSpace(body.Token)
	}
	if token == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	id, err := a.tokens.ParseJourneyUnsubscribe(token)
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
		return
	}
	// Opt-out discovery emails only — billing events (past_due / pending_payment) keep using pref billing.
	if err := a.store.SetClientDiscoveryPrefOnly(r.Context(), id.UserID, false); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"status":  "unsubscribed",
		"message": t(r, "success.journey_unsubscribed", nil),
	})
}
