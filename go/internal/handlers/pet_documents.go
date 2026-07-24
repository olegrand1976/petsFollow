package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) listPetDocuments(w http.ResponseWriter, r *http.Request) {
	pet, _, ok := a.petAccessForDocuments(w, r)
	if !ok {
		return
	}
	docs, err := a.store.ListPetDocuments(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, docs)
}

func (a *API) uploadPetDocument(w http.ResponseWriter, r *http.Request) {
	pet, id, ok := a.petAccessForDocuments(w, r)
	if !ok {
		return
	}
	url, ct, size, fileName, objectKey, err := a.uploadDocumentFile(r, "documents", pet.ID)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	doc, err := a.store.CreatePetDocument(r.Context(), store.CreatePetDocumentInput{
		PetID:            pet.ID,
		UploadedByUserID: id.UserID,
		Title:            title,
		FileName:         fileName,
		ContentType:      ct,
		FileURL:          url,
		ObjectKey:        objectKey,
		SizeBytes:        size,
	})
	if err != nil {
		if a.media != nil && objectKey != "" {
			_ = a.media.Delete(r.Context(), objectKey)
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, doc)
}

func (a *API) deletePetDocument(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	docID := chi.URLParam(r, "documentID")
	doc, err := a.store.GetPetDocument(r.Context(), docID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "document_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pet, err := a.store.GetPet(r.Context(), doc.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if !a.canAccessPetDocuments(r.Context(), id, pet, store.PermWriteNotes) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	practiceVet := id.Role == kernel.RoleVet && id.PracticeID != "" && pet.PracticeID == id.PracticeID
	ownerClient := id.Role == kernel.RoleClient && pet.OwnerUserID == id.UserID
	if ownerClient && doc.UploadedByUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if !practiceVet && !ownerClient {
		// Shared access (care_pro / external vet / co-owner): deletion requires full.
		if !a.canAccessPetDocuments(r.Context(), id, pet, store.PermFull) {
			writeErr(w, r, http.StatusForbidden, "forbidden", "full_permission_required")
			return
		}
	}
	deleted, err := a.store.DeletePetDocument(r.Context(), doc.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "document_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if a.media != nil && deleted.ObjectKey != "" {
		_ = a.media.Delete(r.Context(), deleted.ObjectKey)
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *API) petAccessForDocuments(w http.ResponseWriter, r *http.Request) (store.Pet, authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return store.Pet{}, authx.Identity{}, false
	}
	need := store.PermRead
	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		need = store.PermWriteNotes
	}
	pet, ok := a.requirePetAccess(w, r, chi.URLParam(r, "petID"), id, need)
	if !ok {
		return store.Pet{}, authx.Identity{}, false
	}
	return pet, id, true
}

func (a *API) canAccessPetDocuments(ctx context.Context, id authx.Identity, pet store.Pet, need store.AccessPermission) bool {
	ok, err := a.store.CanAccessPet(ctx, store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, need)
	return err == nil && ok
}

func (a *API) uploadDocumentFile(r *http.Request, kind, entityID string) (url, contentType string, size int64, fileName, objectKey string, err error) {
	if a.media == nil {
		return "", "", 0, "", "", media.ErrNotConfigured
	}
	if err := r.ParseMultipartForm(media.MaxDocumentBytes + (1 << 20)); err != nil {
		return "", "", 0, "", "", media.ErrTooLarge
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		return "", "", 0, "", "", media.ErrEmptyFile
	}
	defer file.Close()

	ct, err := media.NormalizeDocumentType(hdr.Header.Get("Content-Type"), hdr.Filename)
	if err != nil {
		return "", "", 0, "", "", err
	}
	ext, err := media.ExtForDocument(ct)
	if err != nil {
		return "", "", 0, "", "", err
	}
	fileName = path.Base(strings.TrimSpace(hdr.Filename))
	if fileName == "" || fileName == "." {
		fileName = "document" + ext
	}
	size = hdr.Size
	objectKey = media.ObjectKey(kind, entityID, ext)

	if size <= 0 {
		limited := io.LimitReader(file, media.MaxDocumentBytes+1)
		buf, readErr := io.ReadAll(limited)
		if readErr != nil {
			return "", "", 0, "", "", readErr
		}
		if err := media.ValidateSizeLimit(int64(len(buf)), media.MaxDocumentBytes); err != nil {
			return "", "", 0, "", "", err
		}
		size = int64(len(buf))
		url, err = a.media.Upload(r.Context(), objectKey, bytes.NewReader(buf), size, ct)
		return url, ct, size, fileName, objectKey, err
	}
	if err := media.ValidateSizeLimit(size, media.MaxDocumentBytes); err != nil {
		return "", "", 0, "", "", err
	}
	url, err = a.media.Upload(r.Context(), objectKey, file, size, ct)
	return url, ct, size, fileName, objectKey, err
}
