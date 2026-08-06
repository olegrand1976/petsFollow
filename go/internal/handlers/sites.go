package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) listVetSites(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	includeInactive := r.URL.Query().Get("includeInactive") == "1" || r.URL.Query().Get("includeInactive") == "true"
	items, err := a.store.ListSites(r.Context(), id.PracticeID, includeInactive)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type createSiteReq struct {
	Name                   string `json:"name"`
	Phone                  string `json:"phone"`
	AddressLine1           string `json:"addressLine1"`
	AddressLine2           string `json:"addressLine2"`
	City                   string `json:"city"`
	PostalCode             string `json:"postalCode"`
	CountryCode            string `json:"countryCode"`
	Timezone               string `json:"timezone"`
	CopyScheduleFromSiteID string `json:"copyScheduleFromSiteId"`
}

func (a *API) createVetSite(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req createSiteReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	site, err := a.store.CreateSite(r.Context(), id.PracticeID, store.CreateSiteInput{
		Name: req.Name, Phone: req.Phone, AddressLine1: req.AddressLine1, AddressLine2: req.AddressLine2,
		City: req.City, PostalCode: req.PostalCode, CountryCode: req.CountryCode, Timezone: req.Timezone,
		CopyScheduleFromSiteID: req.CopyScheduleFromSiteID,
	})
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_site"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			case strings.Contains(msg, "invalid_timezone"):
				code = "invalid_timezone"
			case strings.Contains(msg, "invalid_template_site"):
				code = "invalid_template_site"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, site)
}

type patchSiteReq struct {
	Name         *string `json:"name"`
	Phone        *string `json:"phone"`
	AddressLine1 *string `json:"addressLine1"`
	AddressLine2 *string `json:"addressLine2"`
	City         *string `json:"city"`
	PostalCode   *string `json:"postalCode"`
	CountryCode  *string `json:"countryCode"`
	Timezone     *string `json:"timezone"`
	IsPrimary    *bool   `json:"isPrimary"`
	Active       *bool   `json:"active"`
	SortOrder    *int    `json:"sortOrder"`
}

func (a *API) patchVetSite(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req patchSiteReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	site, err := a.store.PatchSite(r.Context(), id.PracticeID, chi.URLParam(r, "id"), store.PatchSiteInput{
		Name: req.Name, Phone: req.Phone, AddressLine1: req.AddressLine1, AddressLine2: req.AddressLine2,
		City: req.City, PostalCode: req.PostalCode, CountryCode: req.CountryCode, Timezone: req.Timezone,
		IsPrimary: req.IsPrimary, Active: req.Active, SortOrder: req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "site_not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_site"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "invalid_timezone"):
				code = "invalid_timezone"
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			case strings.Contains(msg, "cannot_deactivate_primary"):
				code = "cannot_deactivate_primary"
			case strings.Contains(msg, "cannot_promote_inactive"):
				code = "cannot_promote_inactive"
			case strings.Contains(msg, "site_has_future_visits"):
				code = "site_has_future_visits"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, site)
}

func (a *API) deactivateVetSite(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	site, err := a.store.DeactivateSite(r.Context(), id.PracticeID, chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "site_not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_site"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "cannot_deactivate_primary"):
				code = "cannot_deactivate_primary"
			case strings.Contains(msg, "site_has_future_visits"):
				code = "site_has_future_visits"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, site)
}

// siteIDFromRequest reads siteId from query, falling back to body field via optional getter.
func siteIDQuery(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("siteId"))
}
