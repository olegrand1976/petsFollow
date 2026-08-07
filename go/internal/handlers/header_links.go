package handlers

import (
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/headerlinks"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// getVetHeaderLinks returns resolved quick links for the Pro topbar + catalog for settings.
func (a *API) getVetHeaderLinks(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.PracticeID == "" || !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	prefs, country, err := a.store.GetPracticeHeaderLinks(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	locale := i18n.FromContext(r.Context())
	if q := r.URL.Query().Get("locale"); q != "" {
		locale = i18n.NormalizeLocale(q)
	}
	labelFn := headerlinks.LabelFR
	items := headerlinks.Resolve(country, locale, prefs, labelFn)
	catalog := headerlinks.CatalogForSettings(country, locale, prefs, labelFn)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"countryCode": country,
		"items":       items,
		"catalog":     catalog,
		"prefs":       prefs,
		"maxCustom":   headerlinks.MaxCustomLinks,
	})
}
