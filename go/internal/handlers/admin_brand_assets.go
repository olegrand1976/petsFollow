package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerBrandAssetAdminRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/admin/brand-assets", a.adminListBrandAssets)
		pr.Post("/admin/brand-assets/{key}", a.adminUploadBrandAsset)
		pr.Patch("/admin/brand-assets/{key}", a.adminPatchBrandAsset)
	})
}

func (a *API) adminListBrandAssets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	items, err := a.store.ListBrandAssets(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func brandAssetKeyOK(key string) bool {
	switch key {
	case store.BrandAssetQRAndroid, store.BrandAssetQRIOS:
		return true
	default:
		return false
	}
}

func (a *API) adminUploadBrandAsset(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	key := chi.URLParam(r, "key")
	if !brandAssetKeyOK(key) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_brand_asset_key")
		return
	}
	url, err := a.uploadImage(r, "brand", key)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	storeURL := strings.TrimSpace(r.FormValue("storeUrl"))
	asset, err := a.store.UpsertBrandAsset(r.Context(), key, "brand/"+key, url, storeURL)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, asset)
}

type patchBrandAssetReq struct {
	StoreURL string `json:"storeUrl"`
}

func (a *API) adminPatchBrandAsset(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	key := chi.URLParam(r, "key")
	if !brandAssetKeyOK(key) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_brand_asset_key")
		return
	}
	var req patchBrandAssetReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	asset, err := a.store.PatchBrandAssetStoreURL(r.Context(), key, strings.TrimSpace(req.StoreURL))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			asset, err = a.store.UpsertBrandAsset(r.Context(), key, "", "", strings.TrimSpace(req.StoreURL))
		}
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	httpx.WriteData(w, http.StatusOK, asset)
}
