package handlers

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const contactPhoneMaxRunes = 40

// normalizeContactPhone trims and validates length.
// requireNonEmpty: empty after trim → "contact_phone_required".
// Too long → "contact_phone_too_long". Empty errCode means OK.
func normalizeContactPhone(raw string, requireNonEmpty bool) (phone string, errCode string) {
	phone = strings.TrimSpace(raw)
	if phone == "" {
		if requireNonEmpty {
			return "", "contact_phone_required"
		}
		return "", ""
	}
	if utf8.RuneCountInString(phone) > contactPhoneMaxRunes {
		return "", "contact_phone_too_long"
	}
	return phone, ""
}

type patchClientReq struct {
	ContactPhone *string `json:"contactPhone"`
}

func (a *API) patchClient(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	var req patchClientReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.ContactPhone == nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "nothing_to_update")
		return
	}
	phone, code := normalizeContactPhone(*req.ContactPhone, false)
	if code != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", code)
		return
	}
	clientID := chi.URLParam(r, "clientID")
	client, err := a.store.UpdateClientContactPhoneByPractice(r.Context(), id.PracticeID, clientID, phone)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, client)
}
