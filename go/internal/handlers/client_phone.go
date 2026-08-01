package handlers

import (
	"errors"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

const contactPhoneMaxRunes = 40
const clientAddressMaxRunes = 500
const nationalRegistryMaxRunes = 20

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

func normalizeClientAddress(raw string) (address string, errCode string) {
	address = strings.TrimSpace(raw)
	if utf8.RuneCountInString(address) > clientAddressMaxRunes {
		return "", "address_too_long"
	}
	return address, ""
}

func normalizeNationalRegistry(raw string) (niss string, errCode string) {
	niss = strings.TrimSpace(raw)
	if niss == "" {
		return "", ""
	}
	if utf8.RuneCountInString(niss) > nationalRegistryMaxRunes {
		return "", "national_registry_too_long"
	}
	for _, r := range niss {
		if unicode.IsDigit(r) || r == '.' || r == '-' || r == ' ' {
			continue
		}
		return "", "national_registry_invalid"
	}
	return niss, ""
}

type patchClientReq struct {
	ContactPhone           *string `json:"contactPhone"`
	FirstName              *string `json:"firstName"`
	LastName               *string `json:"lastName"`
	Address                *string `json:"address"`
	NationalRegistryNumber *string `json:"nationalRegistryNumber"`
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
	if req.ContactPhone == nil && req.FirstName == nil && req.LastName == nil &&
		req.Address == nil && req.NationalRegistryNumber == nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "nothing_to_update")
		return
	}
	patch := store.ClientProfilePatch{}
	if req.ContactPhone != nil {
		phone, code := normalizeContactPhone(*req.ContactPhone, false)
		if code != "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		patch.ContactPhone = &phone
	}
	if req.FirstName != nil {
		v := strings.TrimSpace(*req.FirstName)
		patch.FirstName = &v
	}
	if req.LastName != nil {
		v := strings.TrimSpace(*req.LastName)
		patch.LastName = &v
	}
	if req.Address != nil {
		addr, code := normalizeClientAddress(*req.Address)
		if code != "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		patch.Address = &addr
	}
	if req.NationalRegistryNumber != nil {
		niss, code := normalizeNationalRegistry(*req.NationalRegistryNumber)
		if code != "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		patch.NationalRegistryNumber = &niss
	}
	clientID := chi.URLParam(r, "clientID")
	client, err := a.store.UpdateClientProfileByPractice(r.Context(), id.PracticeID, clientID, patch)
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
