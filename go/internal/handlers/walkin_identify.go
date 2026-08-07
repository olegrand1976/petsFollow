package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type identifyVisitClientReq struct {
	Mode         string `json:"mode"`
	ClientUserID string `json:"clientUserId"`
	PetID        string `json:"petId"`
	NewPet       *struct {
		Name    string `json:"name"`
		Species string `json:"species"`
		Breed   string `json:"breed"`
	} `json:"newPet"`
	Client *struct {
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		ContactPhone string `json:"contactPhone"`
		Email        string `json:"email"`
		Locale       string `json:"locale"`
	} `json:"client"`
}

func (a *API) identifyVisitClient(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	visitID := chi.URLParam(r, "visitID")
	if strings.TrimSpace(visitID) == "" {
		visitID = chi.URLParam(r, "id")
	}
	var req identifyVisitClientReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	in := store.IdentifyVisitClientInput{
		PracticeID:   id.PracticeID,
		VisitID:      visitID,
		ActorUserID:  id.UserID,
		Mode:         store.IdentifyVisitMode(strings.TrimSpace(req.Mode)),
		ClientUserID: strings.TrimSpace(req.ClientUserID),
		PetID:        strings.TrimSpace(req.PetID),
	}
	if req.NewPet != nil {
		in.NewPet = &store.IdentifyVisitNewPet{
			Name:    req.NewPet.Name,
			Species: req.NewPet.Species,
			Breed:   req.NewPet.Breed,
		}
	}
	if req.Client != nil {
		in.FirstName = req.Client.FirstName
		in.LastName = req.Client.LastName
		in.ContactPhone = req.Client.ContactPhone
		in.Email = req.Client.Email
		in.Locale = req.Client.Locale
	}
	out, err := a.store.IdentifyVisitClient(r.Context(), in)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "email_taken")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			code := "validation"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "visit_already_identified"):
				code = "visit_already_identified"
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			case strings.Contains(msg, "contact_phone_required"):
				code = "contact_phone_required"
			case strings.Contains(msg, "pet_required"):
				code = "pet_required"
			case strings.Contains(msg, "name_species_required"):
				code = "name_species_required"
			case strings.Contains(msg, "invalid_mode"):
				code = "invalid_mode"
			case strings.Contains(msg, "cannot_identify_as_walkin"):
				code = "cannot_identify_as_walkin"
			case strings.Contains(msg, "invalid_email"):
				code = "invalid_email"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, out)
}

func (a *API) writeWalkinErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if errors.Is(err, store.ErrWalkinImmutable) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "walkin_immutable")
		return true
	}
	if errors.Is(err, store.ErrWalkinIdentifyRequired) {
		writeErr(w, r, http.StatusConflict, "conflict", "walkin_identify_required")
		return true
	}
	return false
}
