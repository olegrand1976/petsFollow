package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) uploadMyAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	url, err := a.uploadImage(r, "avatars", id.UserID)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	if err := a.store.UpdateUserAvatarURL(r.Context(), id.UserID, url); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	data, err := a.store.GetUserMe(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, data)
}

func (a *API) uploadPetPhoto(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if !a.canEditPetPhoto(id, pet) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	url, err := a.uploadImage(r, "pets", pet.ID)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	if err := a.store.UpdatePetPhotoURL(r.Context(), pet.ID, url); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	updated, err := a.store.GetPet(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) canEditPetPhoto(id authx.Identity, pet store.Pet) bool {
	switch id.Role {
	case kernel.RoleClient:
		return pet.OwnerUserID == id.UserID
	case kernel.RoleVet:
		return pet.PracticeID == id.PracticeID
	default:
		return false
	}
}

func (a *API) uploadImage(r *http.Request, kind, entityID string) (string, error) {
	if a.media == nil {
		return "", media.ErrNotConfigured
	}
	if err := r.ParseMultipartForm(media.MaxUploadBytes + (1 << 20)); err != nil {
		return "", media.ErrTooLarge
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		return "", media.ErrEmptyFile
	}
	defer file.Close()

	ct, err := media.NormalizeContentType(hdr.Header.Get("Content-Type"), hdr.Filename)
	if err != nil {
		return "", err
	}
	ext, err := media.ExtForContentType(ct)
	if err != nil {
		return "", err
	}
	size := hdr.Size
	if size <= 0 {
		// some clients omit size; stream with limit
		limited := io.LimitReader(file, media.MaxUploadBytes+1)
		buf, err := io.ReadAll(limited)
		if err != nil {
			return "", err
		}
		if err := media.ValidateSize(int64(len(buf))); err != nil {
			return "", err
		}
		key := media.ObjectKey(kind, entityID, ext)
		return a.media.Upload(r.Context(), key, bytes.NewReader(buf), int64(len(buf)), ct)
	}
	if err := media.ValidateSize(size); err != nil {
		return "", err
	}
	key := media.ObjectKey(kind, entityID, ext)
	return a.media.Upload(r.Context(), key, file, size, ct)
}

func (a *API) writeUploadErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, media.ErrTooLarge):
		writeErr(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "image_too_large")
	case errors.Is(err, media.ErrInvalidType):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_image_type")
	case errors.Is(err, media.ErrEmptyFile):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
	case errors.Is(err, media.ErrNotConfigured):
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "media_unavailable")
	default:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	}
}
