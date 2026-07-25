package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerCommercialDiscoveryRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	r.With(rateLimit).Get("/commercials/nearby", a.listNearbyCommercials)
}

func (a *API) listNearbyCommercials(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	radius, _ := strconv.ParseFloat(q.Get("radiusKm"), 64)
	postal := strings.TrimSpace(q.Get("postalCode"))

	var latPtr, lngPtr *float64
	if latStr := strings.TrimSpace(q.Get("lat")); latStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_lat")
			return
		}
		latPtr = &lat
	}
	if lngStr := strings.TrimSpace(q.Get("lng")); lngStr != "" {
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_lng")
			return
		}
		lngPtr = &lng
	}
	if (latPtr == nil) != (lngPtr == nil) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "lat_lng_required")
		return
	}
	if latPtr == nil && postal == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "location_required")
		return
	}
	// Reject postal-only queries with no digits (e.g. "abc") — would match nothing useful.
	if latPtr == nil {
		digits := 0
		for _, r := range postal {
			if r >= '0' && r <= '9' {
				digits++
			}
		}
		if digits == 0 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_postal_code")
			return
		}
	}

	rows, err := a.store.ListNearbyCommercials(r.Context(), store.NearbyCommercialsQuery{
		Lat:        latPtr,
		Lng:        lngPtr,
		PostalCode: postal,
		Limit:      limit,
		RadiusKm:   radius,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.NearbyCommercial{}
	}
	httpx.WriteData(w, http.StatusOK, rows)
}
