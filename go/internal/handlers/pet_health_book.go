package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/healthbookpdf"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	maxMicrochipLen        = 64
	maxHealthBookNumberLen = 64
	maxDomicileLocationLen = 500
)

func clipPetIDField(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes])
}

// parseOptionalPetDate accepts YYYY-MM-DD or empty (clears). Invalid format → error.
func parseOptionalPetDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func formatPetDateJSON(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}

// PATCH /vet/pets/{petID}/lifecycle — adoptedAt / soldAt / deceasedAt (DATE, clear with "").
func (a *API) patchVetPetLifecycle(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "pets.write_clinical")
	if !ok {
		return
	}
	var body struct {
		AdoptedAt  *string `json:"adoptedAt"`
		SoldAt     *string `json:"soldAt"`
		DeceasedAt *string `json:"deceasedAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid_json", "invalid_json")
		return
	}
	if body.AdoptedAt == nil && body.SoldAt == nil && body.DeceasedAt == nil {
		writeErr(w, r, http.StatusBadRequest, "validation_error", "validation_error")
		return
	}
	petID := chi.URLParam(r, "petID")
	existing, err := a.store.GetPet(r.Context(), petID)
	if err != nil || existing.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	adoptedAt, soldAt, deceasedAt := existing.AdoptedAt, existing.SoldAt, existing.DeceasedAt
	if body.AdoptedAt != nil {
		t, perr := parseOptionalPetDate(*body.AdoptedAt)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_adopted_at")
			return
		}
		adoptedAt = t
	}
	if body.SoldAt != nil {
		t, perr := parseOptionalPetDate(*body.SoldAt)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_sold_at")
			return
		}
		soldAt = t
	}
	if body.DeceasedAt != nil {
		t, perr := parseOptionalPetDate(*body.DeceasedAt)
		if perr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_deceased_at")
			return
		}
		deceasedAt = t
	}
	if err := a.store.SetPetLifecycleDates(r.Context(), id.PracticeID, petID, adoptedAt, soldAt, deceasedAt); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"id":         petID,
		"adoptedAt":  formatPetDateJSON(adoptedAt),
		"soldAt":     formatPetDateJSON(soldAt),
		"deceasedAt": formatPetDateJSON(deceasedAt),
	})
}

func (a *API) getPetHealthBook(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	pet, ok := a.requirePetAccess(w, r, chi.URLParam(r, "petID"), id, store.PermRead)
	if !ok {
		return
	}
	key := strings.TrimSpace(pet.HealthBookPDFObjectKey)
	if key == "" || a.media == nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "health_book_not_found")
		return
	}
	rc, ct, err := a.media.Open(r.Context(), key)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "health_book_not_found")
		return
	}
	defer rc.Close()
	if ct == "" || ct == "application/octet-stream" {
		ct = "application/pdf"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", `inline; filename="health-book.pdf"`)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = io.Copy(w, rc)
}

func (a *API) uploadPetHealthBook(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "media_unavailable")
		return
	}

	maxForm := int64(healthbookpdf.MaxImages)*healthbookpdf.MaxImageBytes + (2 << 20)
	if err := r.ParseMultipartForm(maxForm); err != nil {
		writeErr(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "image_too_large")
		return
	}

	jpegs, err := collectHealthBookJPEGPages(r)
	if err != nil {
		a.writeHealthBookErr(w, r, err)
		return
	}

	pdfBytes, err := healthbookpdf.BuildPDFFromJPEGs(jpegs)
	if err != nil {
		a.writeHealthBookErr(w, r, err)
		return
	}

	oldKey := strings.TrimSpace(pet.HealthBookPDFObjectKey)
	key := media.ObjectKey("health-books", pet.ID, ".pdf")
	url, err := a.media.Upload(r.Context(), key, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf")
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Sensitive upload returns empty public URL — persist object key only.
	_ = url
	if err := a.store.UpdatePetHealthBookPDF(r.Context(), pet.ID, "", key); err != nil {
		_ = a.media.Delete(r.Context(), key)
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if oldKey != "" && oldKey != key {
		_ = a.media.Delete(r.Context(), oldKey)
	}

	updated, err := a.store.GetPet(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) deletePetHealthBook(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	oldKey, err := a.store.ClearPetHealthBookPDF(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if a.media != nil && strings.TrimSpace(oldKey) != "" {
		_ = a.media.Delete(r.Context(), oldKey)
	}
	updated, err := a.store.GetPet(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func collectHealthBookJPEGPages(r *http.Request) ([][]byte, error) {
	if r.MultipartForm == nil {
		return nil, healthbookpdf.ErrNoImages
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		if f, ok := r.MultipartForm.File["file"]; ok {
			files = f
		}
	}
	if len(files) == 0 {
		return nil, healthbookpdf.ErrNoImages
	}
	if len(files) > healthbookpdf.MaxImages {
		return nil, healthbookpdf.ErrTooManyPages
	}

	out := make([][]byte, 0, len(files))
	for _, hdr := range files {
		raw, err := readHealthBookPart(hdr)
		if err != nil {
			return nil, err
		}
		page, err := healthbookpdf.ToJPEGPage(raw)
		raw = nil // allow GC before assembling PDF
		if err != nil {
			return nil, err
		}
		out = append(out, page)
	}
	return out, nil
}

func readHealthBookPart(hdr *multipart.FileHeader) ([]byte, error) {
	if hdr.Size > healthbookpdf.MaxImageBytes {
		return nil, healthbookpdf.ErrImageTooLarge
	}
	f, err := hdr.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return healthbookpdf.ReadLimited(io.Reader(f), healthbookpdf.MaxImageBytes)
}

func (a *API) writeHealthBookErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, healthbookpdf.ErrNoImages):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
	case errors.Is(err, healthbookpdf.ErrTooManyPages):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "too_many_images")
	case errors.Is(err, healthbookpdf.ErrImageTooLarge):
		writeErr(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "image_too_large")
	case errors.Is(err, healthbookpdf.ErrInvalidImage), errors.Is(err, media.ErrInvalidType):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_image_type")
	default:
		a.writeUploadErr(w, r, err)
	}
}
