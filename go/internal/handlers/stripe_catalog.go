package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerStripeCatalogRoutes(r chi.Router) {
	r.Get("/admin/stripe-catalog", a.adminGetStripeCatalog)
	r.Post("/admin/stripe-catalog/import", a.adminImportStripeCatalog)
}

func (a *API) adminGetStripeCatalog(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	cat, err := a.store.ListStripeCatalog(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, cat)
}

func (a *API) adminImportStripeCatalog(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_multipart")
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.FormValue("kind")))
	if kind == "" {
		kind = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	}
	if kind != "products" && kind != "prices" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "kind_required")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()

	var result store.StripeCatalogImportResult
	switch kind {
	case "products":
		products, perr := store.ParseStripeProductsCSV(file)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_csv")
			return
		}
		if len(products) == 0 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "empty_csv")
			return
		}
		result, err = a.store.UpsertStripeProducts(r.Context(), products)
	case "prices":
		prices, perr := store.ParseStripePricesCSV(file)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_csv")
			return
		}
		if len(prices) == 0 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "empty_csv")
			return
		}
		result, err = a.store.UpsertStripePrices(r.Context(), prices)
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, result)
}
